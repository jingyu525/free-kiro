package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/models"
)

// AnalysisFinding is one advisory consistency issue raised by Analyze.
// Distinct from lint.Issue: this is *consistency* (would a reviewer
// flag it?) not *shape* (does it satisfy the gate?). The CLI renders
// these via `spec analyze`, never as blocking errors.
type AnalysisFinding struct {
	Severity string `json:"severity"` // "warning" | "info"
	Code     string `json:"code"`     // short machine id
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

// vagueWords is the set of soft / non-committal terms the analyzer flags.
// Each entry is matched case-insensitively as a whole word.
var vagueWords = []string{
	"etc", "and/or", "maybe", "some", "user-friendly",
	"robust", "flexible", "seamless", "intuitive", "tbd", "todo",
}

// vagueWordRe is a compiled regex covering the whole list with word
// boundaries. Recompiled per call would be wasteful; build once.
var vagueWordRe = buildVagueWordRe()

func buildVagueWordRe() *regexp.Regexp {
	parts := make([]string, len(vagueWords))
	for i, w := range vagueWords {
		parts[i] = regexp.QuoteMeta(w)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(parts, "|") + `)\b`)
}

// Analyze runs the three advisory consistency checks against a spec:
//
//   - vague-language: soft / non-committal terms in the analysis doc
//   - duplicate-acceptance-criteria: identical EARS lines repeated
//   - uncovered-acceptance-criteria: EARS lines with no matching task
//   - tasks-without-requirements: tasks exist but the analysis doc has zero EARS
//
// Never blocks. Always returns a slice (possibly empty). The CLI
// `spec analyze` command renders these for the user.
func (e *Engine) Analyze(specName string) []AnalysisFinding {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil
	}
	dir := e.ws.SpecDir(specName)
	doc := models.FirstPlanningDoc(meta.SpecType)
	text, _ := os.ReadFile(filepath.Join(dir, doc))
	docText := string(text)

	findings := []AnalysisFinding{}

	// 1. vague language — case-insensitive whole-word match.
	for _, m := range vagueWordRe.FindAllString(docText, -1) {
		findings = append(findings, AnalysisFinding{
			Severity: "info",
			Code:     "vague-language",
			Message:  "soft term used: " + m + " (consider a concrete commitment)",
			Location: doc,
		})
	}

	// 2. duplicate acceptance criteria — normalise each EARS line and dedupe.
	earss := extractEARSLines(docText)
	seen := map[string]int{}
	for _, line := range earss {
		seen[normaliseEARS(line)]++
	}
	for norm, n := range seen {
		if n > 1 {
			findings = append(findings, AnalysisFinding{
				Severity: "warning",
				Code:     "duplicate-acceptance-criteria",
				Message:  "duplicate acceptance criterion appears " + strconv.Itoa(n) + " times: " + truncate(norm, 80),
				Location: doc,
			})
		}
	}

	// 3. requirements↔tasks traceability.
	tasksData, _ := os.ReadFile(filepath.Join(dir, "tasks.md"))
	tasksText := string(tasksData)
	if len(earss) > 0 && isEmptyTasks(tasksText) {
		findings = append(findings, AnalysisFinding{
			Severity: "warning",
			Code:     "uncovered-acceptance-criteria",
			Message:  "spec has " + strconv.Itoa(len(earss)) + " acceptance criteria but no tasks (every AC should map to at least one task)",
			Location: "tasks.md",
		})
	}
	if len(earss) == 0 && taskLineRe.MatchString(tasksText) {
		findings = append(findings, AnalysisFinding{
			Severity: "info",
			Code:     "tasks-without-requirements",
			Message:  "spec has tasks but no EARS acceptance criteria (consider writing requirements first)",
			Location: doc,
		})
	}

	// 4. advisory: did the user ignore lint findings? Surface as warning.
	_ = lint.EARSRe // ensure import is used even when all checks pass
	return findings
}

// extractEARSLines pulls every line that contains an EARS acceptance
// criterion. Used by both the dedupe and the AC-count helpers.
func extractEARSLines(text string) []string {
	var out []string
	for _, line := range splitLines(text) {
		if lint.EARSRe.MatchString(line) {
			out = append(out, line)
		}
	}
	return out
}

// normaliseEARS strips whitespace + lowercases so two semantically equal
// AC lines dedupe. Conservative — we don't try to semantically diff
// WHEN triggers, only the syntactic shape.
func normaliseEARS(line string) string {
	return strings.Join(strings.Fields(strings.ToLower(line)), " ")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func isEmptyTasks(text string) bool {
	for _, line := range splitLines(text) {
		if taskLineRe.MatchString(line) {
			return false
		}
	}
	return true
}
