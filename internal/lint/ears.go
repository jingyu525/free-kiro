// Package lint implements the offline spec-quality gate for free-kiro.
//
// The gate enforces document *shape* (the structural contract that lets the
// rest of the system trust a spec). It is distinct from the advisory
// `analyze` check (see internal/spec/analyze.go) which hunts consistency
// problems a reviewer would catch by eye. The lint gate blocks state-machine
// transitions; analyze never does.
//
// Every rule here is a pure function over a document's text. No model, no
// network, no side effects — easy to unit-test, easy to run in CI.
//
// Five EARS templates + the ubiquitous baseline are recognised:
//
//	event-driven    WHEN <event> THE SYSTEM SHALL <response>
//	state-driven    WHILE <state> THE SYSTEM SHALL <response>
//	optional        WHERE <feature> THE SYSTEM SHALL <response>
//	exception       UNLESS <exemption> THE SYSTEM SHALL <response>   (per-request alias for "UNLESS")
//	complex         IF <condition> THEN <event> THE SYSTEM SHALL <response>
//	ubiquitous      THE SYSTEM SHALL <requirement>
//
// Bugfix docs follow a different shape (Current / Expected / Unchanged
// Behavior) — see bugfix.go.
package lint

import (
	"regexp"
	"strings"

	"github.com/jingyu525/free-kiro/internal/models"
)

// Issue is a single finding from the lint gate.
type Issue struct {
	Severity string // "error" | "warning"
	Code     string // short machine-readable id (e.g. "no-ears")
	Message  string // human-readable explanation (English, for tooling)
	Location string // optional filename:section hint
	Hint     string // optional fix hint (URL or guidance)
}

// String renders the issue for CLI output (severity-prefixed, location-suffixed,
// hint-appended). Used by both `lint` and `spec analyze`.
func (i Issue) String() string {
	out := i.Severity + ":" + i.Code
	if i.Location != "" {
		out += " [" + i.Location + "]"
	}
	out += " " + i.Message
	if i.Hint != "" {
		out += "  →  " + i.Hint
	}
	return out
}

// Severity constants — pass through to Issue.Severity.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// EARSRe matches any of the six EARS acceptance-criterion templates. The
// whole expression is one big alternation; each branch is anchored at the
// trigger keyword and matches greedily up to "THE SYSTEM SHALL" (which is
// required for all but the ubiquitous form, which starts with "THE SYSTEM
// SHALL" itself).
//
// The case-insensitive flag mirrors kiro's Python regex so both tools
// produce the same lint verdict on the same document.
var EARSRe = regexp.MustCompile(`(?i)(?:` +
	`WHEN\s+.+?\s+THE\s+SYSTEM\s+SHALL` +
	`|WHILE\s+.+?\s+THE\s+SYSTEM\s+SHALL` +
	`|WHERE\s+.+?\s+THE\s+SYSTEM\s+SHALL` +
	`|UNLESS\s+.+?\s+THE\s+SYSTEM\s+SHALL` +
	`|IF\s+.+?\s+THEN\s+.+?\s+THE\s+SYSTEM\s+SHALL` +
	`|THE\s+SYSTEM\s+SHALL` +
	`)`)

// SHALLRe matches "THE SYSTEM SHALL" as a standalone token (used by bugfix
// lint rules).
var SHALLRe = regexp.MustCompile(`(?i)THE\s+SYSTEM\s+SHALL`)

// SHALLContinueRe matches the regression-prevention form "THE SYSTEM SHALL
// CONTINUE TO" used in bugfix Unchanged Behavior sections.
var SHALLContinueRe = regexp.MustCompile(`(?i)THE\s+SYSTEM\s+SHALL\s+CONTINUE\s+TO`)

// UserStoryRe matches a User Stories heading or "user stor…" token.
var UserStoryRe = regexp.MustCompile(`(?i)user\s+stor`)

// Per-template EARS regexes — split out from EARSRe so callers can ask
// "did this doc actually use WHEN?" or "what trigger phrase follows
// WHILE?" without re-deriving the alternation. Used by quality.go and
// (read-only) by spec analyze for per-template coverage stats.
//
// Each regex matches the same EARS template as the corresponding branch
// in EARSRe, anchored at the trigger keyword and requiring the
// mandatory `THE SYSTEM SHALL` tail. Case-insensitive to mirror EARSRe.
//
// The IF-THEN template is exposed as a function (UsesIFTHEN) rather
// than a regex because RE2 cannot backtrack across two lazy `.+?\s+`
// segments — the IF prefix and SHALL suffix are matched independently
// and combined. (The IF branch inside EARSRe itself relies on the
// alternation falling through to the ubiquitous `THE SYSTEM SHALL`
// branch.)
var (
	WHENRe    = regexp.MustCompile(`(?i)WHEN\s+.+?\s+THE\s+SYSTEM\s+SHALL`)
	WHILERe   = regexp.MustCompile(`(?i)WHILE\s+.+?\s+THE\s+SYSTEM\s+SHALL`)
	WHERERe   = regexp.MustCompile(`(?i)WHERE\s+.+?\s+THE\s+SYSTEM\s+SHALL`)
	UNLESSRe  = regexp.MustCompile(`(?i)UNLESS\s+.+?\s+THE\s+SYSTEM\s+SHALL`)

	ifThenPrefixRe = regexp.MustCompile(`(?i)\bIF\s+.+?\s+THEN\b`)
	ifThenSuffixRe = regexp.MustCompile(`(?i)\bTHEN\s+(?:.+?\s+)?THE\s+SYSTEM\s+SHALL`)
)

// UsesIFTHEN reports whether text uses the IF-THEN EARS template. Both
// halves must match for the doc to count as IF-THEN, because the full
// `IF ... THEN ... THE SYSTEM SHALL` regex is not expressible as a
// single RE2 expression (two `.+?\s+` lazy segments cannot backtrack
// across each other).
func UsesIFTHEN(text string) bool {
	return ifThenPrefixRe.MatchString(text) && ifThenSuffixRe.MatchString(text)
}

// sectionRe splits a markdown doc on `## <heading>` lines.
var sectionRe = regexp.MustCompile(`^##\s+(.*)$`)

// splitSections returns {lowercased heading: body} for each `## …` section
// in text. Trailing parenthetical qualifiers are stripped so
// `## Current Behavior (Defect)` normalises to `current behavior`.
//
// Body is the text between this heading and the next `## …` (or EOF),
// trimmed of surrounding whitespace. The leading `#` / `# Title` H1 is
// ignored; only H2 (`##`) boundaries participate.
func splitSections(text string) map[string]string {
	sections := map[string]string{}
	var current string
	var buf []string
	for _, line := range rangeLines(text) {
		if m := sectionRe.FindStringSubmatch(line); m != nil {
			if current != "" {
				sections[current] = joinAndTrim(buf)
			}
			current = normaliseHeading(m[1])
			buf = nil
			continue
		}
		if current != "" {
			buf = append(buf, line)
		}
	}
	if current != "" {
		sections[current] = joinAndTrim(buf)
	}
	return sections
}

// rangeLines is a tiny iterator that returns each line of text with the
// trailing newline stripped. We avoid strings.Split + range to keep the
// hot path allocation-free for large docs.
func rangeLines(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			out = append(out, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		out = append(out, text[start:])
	}
	return out
}

// joinAndTrim joins lines with newlines and trims surrounding blank lines.
func joinAndTrim(lines []string) string {
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// normaliseHeading strips a trailing `(...)` qualifier, trims whitespace,
// and lowercases. Used to make lint section lookups case- and qualifier-
// insensitive.
func normaliseHeading(h string) string {
	if i := strings.Index(h, "("); i >= 0 {
		h = h[:i]
	}
	return strings.ToLower(strings.TrimSpace(h))
}

// SpecTypeFor loads .meta.json and returns spec_type ("feature" by default).
// Errors are silently ignored — the linter only needs the type to pick the
// right rule set; an unreadable meta falls back to feature.
func SpecTypeFor(specDir string) string {
	m, err := models.LoadSpecMeta(specDir)
	if err != nil || m == nil {
		return models.SpecTypeFeature
	}
	return m.SpecType
}
