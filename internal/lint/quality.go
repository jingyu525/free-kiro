package lint

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/jingyu525/free-kiro/internal/text"
)

// Quality rules — add semantic checks on top of the shape gate
// (requirements.go / bugfix.go). Each Check* is a pure function over the
// document text; it returns issues but does not print or mutate. The
// orchestrator in Requirements() (see requirements.go) is responsible
// for stitching them together.
//
// Naming convention:
//
//   - Check*    : per-rule pure checker; returns []Issue.
//   - extract*  : helper that pulls a sub-string out of a matched AC
//                 line (trigger phrase, response phrase, …).
//
// Severity conventions:
//
//   - ERROR     : blocks `spec approve` / `spec advance` (CI gate).
//   - WARNING   : advisory; surfaces in `free-kiro lint` output but does
//                 not block. Used for heuristic rules that risk false
//                 positives during the rollout period.

// acIDPrefixRe matches the `[AC-N]` identifier prefix at the start of an
// AC line. Used to identify the FIRST line of a (possibly multi-line)
// AC — continuation lines of the same AC are not matched. Accepts an
// optional `- ` bullet prefix to mirror acIDRe's tolerance: specs that
// prefer bullet-style ACs (`- [AC-1] WHEN ...`) must be recognized as
// the AC marker line, otherwise extractEARSLines would mistreat the
// following template-keyword continuation as a fresh AC.
var acIDPrefixRe = regexp.MustCompile(`(?i)^\s*(?:-\s+)?\[AC-\d+\]`)

// templateStartRe matches the START of a legacy single-line AC: one of
// the EARS template keywords at line start (after optional whitespace
// and an optional `-` bullet). Together with EARSRe this identifies
// legacy specs that pre-date the `[AC-N]` ID convention.
var templateStartRe = regexp.MustCompile(`(?i)^\s*(?:-\s+)?(?:WHEN\b|WHILE\b|WHERE\b|UNLESS\b|IF\b|THE\s+SYSTEM\s+SHALL\b)`)

// ACLine pairs the matched AC line text with its 1-based file line
// number so the lint output's Location points at the actual line in the
// requirements document (not the index in the filtered list).
type ACLine struct {
	LineNum int    // 1-based line number in the source text
	Text    string // the line content (without trailing newline)
}

// extractEARSLines returns every AC-shaped line in text with its file
// line number. A line qualifies when EITHER (a) it starts with `[AC-N]`
// — the canonical AC marker, regardless of whether the line also matches
// EARSRe (the AC may span multiple lines with `THE SYSTEM SHALL` on a
// continuation), OR (b) it starts with an EARS template keyword AND
// matches EARSRe on the same line — the legacy single-line AC format.
// Continuation lines of multi-line ACs do NOT start with any of these
// markers and are filtered out, so each AC contributes exactly one line
// to the output.
//
// Implementation note: we also skip template-keyword lines that appear
// immediately after an `[AC-N]` line — those are continuations of the
// just-started multi-line AC, not standalone ACs.
func extractEARSLines(doc string) []ACLine {
	var out []ACLine
	prevWasACID := false
	for i, line := range text.RangeLines(doc) {
		if acIDPrefixRe.MatchString(line) {
			out = append(out, ACLine{LineNum: i + 1, Text: line})
			prevWasACID = true
			continue
		}
		if prevWasACID && templateStartRe.MatchString(line) {
			// Continuation of a multi-line AC; skip.
			continue
		}
		if templateStartRe.MatchString(line) && EARSRe.MatchString(line) {
			out = append(out, ACLine{LineNum: i + 1, Text: line})
		}
		prevWasACID = false
	}
	return out
}

// extractTrigger returns the phrase that follows the trigger keyword
// (WHEN / WHILE / etc.) and precedes " THE SYSTEM SHALL". For example:
//
//	"WHEN rate > 100/min THE SYSTEM SHALL throttle" → "rate > 100/min"
//	"WHILE session is active THE SYSTEM SHALL refresh" → "session is active"
//
// Returns "" if the line does not match the WHEN/WHILE template or the
// SHALLOCCUR pattern is not found.
func extractTrigger(line string) string {
	// Try WHEN first, then WHILE. Other templates (WHERE/UNLESS/IF-THEN)
	// don't have a notion of trigger observability in the same way, so
	// they're skipped by the trigger-observability check.
	if m := whenTriggerRe.FindStringSubmatch(line); m != nil {
		return strings.TrimSpace(m[1])
	}
	if m := whileTriggerRe.FindStringSubmatch(line); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// whenTriggerRe matches `WHEN <trigger> THE SYSTEM SHALL` and captures
// the trigger phrase in group 1.
var whenTriggerRe = regexp.MustCompile(`(?i)^.*?\bWHEN\s+(.+?)\s+THE\s+SYSTEM\s+SHALL\b`)

// whileTriggerRe matches `WHILE <state> THE SYSTEM SHALL` and captures
// the state phrase in group 1.
var whileTriggerRe = regexp.MustCompile(`(?i)^.*?\bWHILE\s+(.+?)\s+THE\s+SYSTEM\s+SHALL\b`)

// extractResponse returns the phrase after "THE SYSTEM SHALL" up to the
// end of line / sentence. Returns "" if the line does not match the
// EARS pattern.
func extractResponse(line string) string {
	idx := SHALLRe.FindStringIndex(line)
	if idx == nil {
		return ""
	}
	return strings.TrimSpace(line[idx[1]:])
}

// etcListRe matches `\b(etc|and/or)\b` inside EARS AC lines. The pattern
// is anchored at a word boundary so words containing `etc` as a substring
// (e.g. `etcd`, `kubernetes`) are not flagged. `and/or` is anchored on
// both sides because Go's `\b` treats `/` as a non-word boundary.
var etcListRe = regexp.MustCompile(`(?i)\b(etc|and/or)\b`)

// CheckEtcList — AC-1 (ERROR): an EARS acceptance criterion containing
// `etc` or `and/or` is a sign the author punted on enumerating concrete
// cases. Force them to write them out.
func CheckEtcList(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		if etcListRe.MatchString(ac.Text) {
			out = append(out, Issue{
				Severity: SeverityError,
				Code:     "ears-etc-list",
				Message:  "AC contains 'etc' or 'and/or' — enumerate the concrete cases instead",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
				Hint:     "see docs/EARS.md#vague-words for the blacklist policy",
			})
		}
	}
	return out
}

// CheckSingleSHALLLine — AC-2 (WARNING): a single line containing two or
// more `THE SYSTEM SHALL` clauses is hard to trace back to tasks and
// tends to hide conditionals. Force authors to split into separate ACs.
func CheckSingleSHALLLine(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		n := len(SHALLRe.FindAllString(ac.Text, -1))
		if n >= 2 {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-multi-shall-line",
				Message:  "line contains " + strconv.Itoa(n) + " SHALL keywords — split into separate AC lines",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
			})
		}
	}
	return out
}

// CheckFewAC — AC-3 (WARNING): a requirements.md with fewer than
// models.MinAcceptanceCriteria EARS ACs leaves the spec under-specified.
// WARNING rather than ERROR so tiny docs (one AC for a small bugfix-style
// feature) can still ship.
func CheckFewAC(doc string, minAC int) []Issue {
	count := len(extractEARSLines(doc))
	if count >= minAC {
		return nil
	}
	return []Issue{{
		Severity: SeverityWarning,
		Code:     "ears-few-ac",
		Message: "only " + strconv.Itoa(count) + " EARS AC(s) found — write at least " +
			strconv.Itoa(minAC) + " (configurable via models.MinAcceptanceCriteria)",
		Location: "requirements.md",
		Hint:     "see docs/EARS.md#semantic-quality-gates",
	}}
}

// CheckTemplateDiversity — AC-4 (WARNING): a doc that only uses one EARS
// template (e.g. 8 ACs all starting with WHEN) likely missed boundary
// cases. Encourage authors to cover state/exception/optional branches.
//
// We count "ubiquitous" (a line with bare `THE SYSTEM SHALL` and no
// trigger keyword) only when such a line actually exists, so that two
// WHEN lines don't artificially count as 2 templates just because each
// contains "THE SYSTEM SHALL".
func CheckTemplateDiversity(doc string) []Issue {
	used := 0
	if WHENRe.MatchString(doc) {
		used++
	}
	if WHILERe.MatchString(doc) {
		used++
	}
	if WHERERe.MatchString(doc) {
		used++
	}
	if UNLESSRe.MatchString(doc) {
		used++
	}
	if UsesIFTHEN(doc) {
		used++
	}
	for _, ac := range extractEARSLines(doc) {
		if !WHENRe.MatchString(ac.Text) &&
			!WHILERe.MatchString(ac.Text) &&
			!WHERERe.MatchString(ac.Text) &&
			!UNLESSRe.MatchString(ac.Text) &&
			!UsesIFTHEN(ac.Text) &&
			SHALLRe.MatchString(ac.Text) {
			used++
			break
		}
	}
	if used >= 2 {
		return nil
	}
	return []Issue{{
		Severity: SeverityWarning,
		Code:     "ears-low-template-diversity",
		Message:  "only 1 EARS template used — cover at least 2 of WHEN/WHILE/WHERE/UNLESS/IF-THEN/ubiquitous",
		Location: "requirements.md",
		Hint:     "see docs/EARS.md#semantic-quality-gates",
	}}
}

// acIDRe matches the `[AC-N]` prefix an AC line must carry. Both
// bulleted (`- [AC-1] …`) and bare (`[AC-1] …`) forms are accepted so
// specs that prefer paragraph-style ACs don't have to invent bullets.
// Anchored at line start; trailing whitespace tolerance via `\s+`.
var acIDRe = regexp.MustCompile(`(?i)^\s*(?:-\s+)?\[AC-\d+\]\s+`)

// CheckACMissingID — AC-5 (WARNING): every EARS AC must carry an `[AC-N]`
// prefix so the bidirectional trace to tasks.md is mechanical instead of
// heuristic. WARNING because the feature is opt-in during rollout.
func CheckACMissingID(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		if !acIDRe.MatchString(ac.Text) {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-ac-missing-id",
				Message:  "AC line is missing [AC-N] prefix — add e.g. '- [AC-1]' at the start",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
			})
		}
	}
	return out
}

// measurableRe detects at least one concrete measurable promise in a
// response phrase. We accept any of:
//   - a digit (e.g. "3 retries", "30 minutes")
//   - a time unit token: ms, s, sec, second, min, minute, h, hour, day
//   - an HTTP-style status code: 2xx / 4xx / 5xx
//   - a percentage: % or percent
//   - the words `within` or `at most`
//   - a comparison operator (`<`/`>`/`≤`/`≥`) followed by a digit
//
// Hits are enough — we don't validate the magnitude, just that the
// author committed to a concrete number/limit.
var measurableRe = regexp.MustCompile(`(?i)` +
	`\d|` +
	`\bms\b|\bsec(?:ond)?s?\b|\bmin(?:ute)?s?\b|\bh(?:our)?s?\b|\bdays?\b|` +
	`\b[245]\d\d\b|` + // HTTP-ish status codes 200-599 (allows any 2xx/4xx/5xx)
	`%|percent|` +
	`\bwithin\b|` +
	`\bat\s+most\b|` +
	`[<>≤≥]\s*\d`,
)

// CheckMeasurableResponse — AC-6 (WARNING): a SHALLOW response without a
// concrete measurable commitment ("respond fast", "be robust") cannot be
// objectively verified. Force authors to commit to a number, time unit,
// status code, or limit.
func CheckMeasurableResponse(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		resp := extractResponse(ac.Text)
		if resp == "" {
			continue
		}
		if !measurableRe.MatchString(resp) {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-response-immeasurable",
				Message:  "SHALL response has no measurable commitment (digit / time unit / status code / within / at most)",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
				Hint:     "see docs/EARS.md#semantic-quality-gates",
			})
		}
	}
	return out
}

// triggerVagueWords are words that look like triggers but cannot be
// objectively observed. Case-insensitive. Kept conservative — the bar is
// "an automated test could observe this trigger without human judgment".
var triggerVagueWords = regexp.MustCompile(`(?i)\b(busy|slow|normal|large|small|many|recently|soon|fast|robust|flexible|seamless|intuitive)\b`)

// CheckTriggerObservable — AC-7 (WARNING): a WHEN/WHILE trigger phrase
// built on subjective language ("when system is busy", "while recently
// logged in") cannot gate automated tests. Force authors to commit to a
// concrete signal (rate, count, status, threshold).
func CheckTriggerObservable(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		trigger := extractTrigger(ac.Text)
		if trigger == "" {
			continue
		}
		if hit := triggerVagueWords.FindString(trigger); hit != "" {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-trigger-unobservable",
				Message:  "trigger phrase contains vague word " + strconv.Quote(hit) + " — use an observable signal (count, threshold, status)",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
			})
		}
	}
	return out
}

// pastTenseVerbs are past-tense verbs in active voice that signal an
// event, not a state. The check is conservative (WARNING): we accept
// false positives on participle usage ("WHILE user is logged in") in
// exchange for catching the common mistake of writing `WHILE user
// logged in ...` instead of `WHEN user logs in ...`.
var pastTenseVerbs = regexp.MustCompile(`(?i)\b(logged|submitted|clicked|executed|started|finished|failed|expired|received|sent|pressed|opened|closed)\b`)

// CheckKeywordMisuse — AC-8 (WARNING): a WHILE line whose state phrase
// ends in a past-tense verb is mis-classifying an event as a state.
// Suggest switching to WHEN. Conservative: we only flag the strong
// pattern (state ends exactly in one of the listed verbs) to avoid noise.
func CheckKeywordMisuse(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		if !WHILERe.MatchString(ac.Text) {
			continue
		}
		m := whileTriggerRe.FindStringSubmatch(ac.Text)
		if m == nil {
			continue
		}
		state := strings.TrimSpace(m[1])
		if pastTenseVerbs.MatchString(state) {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-keyword-misuse-while-as-when",
				Message:  "WHILE state phrase ends in past-tense verb — this is an event, not a state; consider WHEN",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
			})
		}
	}
	return out
}

// passiveResponseRe matches a SHALLOW response phrase that starts with a
// pure passive verb (be / is / are / been). Such responses have no
// concrete subject action — "the system SHALL be fast" doesn't say who
// does what.
var passiveResponseRe = regexp.MustCompile(`(?i)^\s*(be|is|are|been)\b`)

// CheckPassiveResponse — AC-9 (WARNING): "SHALL be X" / "SHALL be
// considered done" are passive promises without a concrete subject
// action. Encourage authors to commit to a measurable response.
func CheckPassiveResponse(doc string) []Issue {
	var out []Issue
	for _, ac := range extractEARSLines(doc) {
		resp := extractResponse(ac.Text)
		if resp == "" {
			continue
		}
		if passiveResponseRe.MatchString(resp) {
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "ears-passive-response",
				Message:  "SHALL response is pure passive (starts with be/is/are/been) — describe a concrete action",
				Location: "requirements.md:" + strconv.Itoa(ac.LineNum),
			})
		}
	}
	return out
}
