package lint

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jingyu525/free-kiro/internal/models"
)

// LintSpec runs every rule over one spec's documents. Missing documents
// are flagged as ERROR / WARNING depending on whether the spec was
// expected to have written them yet (see LintGate for the gating subset).
//
// The branch on spec_type picks the right rule for the analysis doc:
// feature specs use requirements.md + EARS; bugfix specs use bugfix.md
// + the Current/Expected/Unchanged contract.
//
// Returns issues in a deterministic order: type-dependent doc first, then
// tasks.md, then design.md warnings.
func LintSpec(specDir string) []LintIssue {
	var out []LintIssue
	specType := SpecTypeFor(specDir)
	firstDoc := models.FirstPlanningDoc(specType)

	firstPath := filepath.Join(specDir, firstDoc)
	if data, err := os.ReadFile(firstPath); err == nil {
		text := string(data)
		if specType == models.SpecTypeBugfix {
			out = append(out, LintBugfix(text)...)
		} else {
			out = append(out, LintRequirements(text)...)
		}
	} else {
		// Missing-doc findings use code `missing-<stem>` so the gate can
		// filter them out (see LintGate).
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "missing-" + stem(firstDoc),
			Message:  firstDoc + " not found",
			Location: firstDoc,
		})
	}

	tasksPath := filepath.Join(specDir, "tasks.md")
	if data, err := os.ReadFile(tasksPath); err == nil {
		out = append(out, LintTasks(string(data))...)
	} else {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "missing-tasks",
			Message:  "tasks.md not found",
			Location: "tasks.md",
		})
	}

	designPath := filepath.Join(specDir, "design.md")
	if _, err := os.Stat(designPath); os.IsNotExist(err) {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "missing-design",
			Message:  "design.md not found (spec may be incomplete)",
			Location: "design.md",
		})
	}

	return out
}

// LintGate returns the ERROR findings that should block advance/approve.
// "missing-*" findings (a phase the author has not written yet) are
// excluded from the gate — you cannot be failed for a document that does
// not exist. This lets the very first `generate` (nothing written yet)
// through while still blocking an advance off a malformed document.
func LintGate(specDir string) []LintIssue {
	var gate []LintIssue
	for _, i := range LintSpec(specDir) {
		if i.Severity != SeverityError {
			continue
		}
		if isMissingCode(i.Code) {
			continue
		}
		gate = append(gate, i)
	}
	return gate
}

// isMissingCode reports whether a lint code marks a missing-doc finding.
// `missing-` findings are excluded from the advance/approve gate (see
// LintGate) — you cannot be failed for a document that does not exist.
func isMissingCode(code string) bool {
	return strings.HasPrefix(code, "missing-")
}

func stem(p string) string {
	base := filepath.Base(p)
	if dot := strings.IndexByte(base, '.'); dot >= 0 {
		return base[:dot]
	}
	return base
}