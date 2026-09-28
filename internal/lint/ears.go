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

	"github.com/jingyu525/free-kiro/internal/models"
)

// LintIssue is a single finding from the lint gate.
type LintIssue struct {
	Severity string // "error" | "warning"
	Code     string // short machine-readable id (e.g. "no-ears")
	Message  string // human-readable explanation (English, for tooling)
	Location string // optional filename:section hint
}

// String renders the issue for CLI output (severity-prefixed, location-suffixed).
func (i LintIssue) String() string {
	s := i.Severity + ":" + i.Code + " " + i.Message
	if i.Location != "" {
		s += " [" + i.Location + "]"
	}
	return s
}

// Severity constants — pass through to LintIssue.Severity.
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

func joinAndTrim(lines []string) string {
	// Trim leading/trailing blank lines.
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

// normaliseHeading strips a trailing `(...)` qualifier, trims whitespace,
// and lowercases. Used to make lint section lookups case- and qualifier-
// insensitive.
func normaliseHeading(h string) string {
	out := ""
	for _, r := range h {
		if r == '(' {
			break
		}
		out += string(r)
	}
	// trim leading + trailing whitespace manually.
	for len(out) > 0 && (out[0] == ' ' || out[0] == '\t' || out[0] == '\n' || out[0] == '\r') {
		out = out[1:]
	}
	for len(out) > 0 && (out[len(out)-1] == ' ' || out[len(out)-1] == '\t' || out[len(out)-1] == '\n' || out[len(out)-1] == '\r') {
		out = out[:len(out)-1]
	}
	// lower-case via a tiny ASCII loop (no strings.ToLower allocation).
	lower := ""
	for _, r := range out {
		if r >= 'A' && r <= 'Z' {
			lower += string(r + 32)
		} else {
			lower += string(r)
		}
	}
	return lower
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