package spec

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/models"
)

// captureBaseline reads the current state of a spec's planning documents
// and returns a structured snapshot suitable for drift detection. Keys:
//
//	ac_count      — number of EARS acceptance-criterion matches in the analysis doc
//	task_count    — number of parseable task rows in tasks.md
//	design_sections — count of `## <heading>` lines in design.md
//
// Captured at Approve time and refreshed by Sync. Status() compares the
// current snapshot to the baseline and reports any deltas.
func captureBaseline(specDir, specType string) map[string]int {
	return map[string]int{
		"ac_count":        countAC(specDir, specType),
		"task_count":      countParseableTasks(specDir),
		"design_sections": countDesignSections(specDir),
	}
}

// countAC returns the number of EARS acceptance criteria in the spec's
// analysis doc (requirements.md for feature specs, bugfix.md for bugfix).
func countAC(specDir, specType string) int {
	doc := models.FirstPlanningDoc(specType)
	data, err := os.ReadFile(filepath.Join(specDir, doc))
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range splitLines(string(data)) {
		if lint.EARSRe.MatchString(line) {
			count++
		}
	}
	return count
}

// countParseableTasks returns the number of task rows matching the
// canonical `- [ ] #N Title` format.
func countParseableTasks(specDir string) int {
	data, err := os.ReadFile(filepath.Join(specDir, "tasks.md"))
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range splitLines(string(data)) {
		if taskLineRe.MatchString(line) {
			count++
		}
	}
	return count
}

// countDesignSections counts `## <heading>` lines in design.md (a rough
// progress signal for the design phase).
func countDesignSections(specDir string) int {
	data, err := os.ReadFile(filepath.Join(specDir, "design.md"))
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range splitLines(string(data)) {
		if h2Re.MatchString(line) {
			count++
		}
	}
	return count
}

// h2Re matches `## <heading>` lines (used only by countDesignSections).
var h2Re = regexp.MustCompile(`^##\s+\S`)

// DriftSignal is one delta between baseline and current state.
type DriftSignal struct {
	Key      string `json:"key"`
	Baseline int    `json:"baseline"`
	Current  int    `json:"current"`
	Delta    int    `json:"delta"`
}

// computeDrift diffs two baseline snapshots. Returns nil when they match.
func computeDrift(baseline, current map[string]int) []DriftSignal {
	if len(baseline) == 0 {
		return nil
	}
	var out []DriftSignal
	for k, b := range baseline {
		c := current[k]
		if c != b {
			out = append(out, DriftSignal{Key: k, Baseline: b, Current: c, Delta: c - b})
		}
	}
	return out
}