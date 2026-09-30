package lint

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jingyu525/free-kiro/internal/models"
)

// baselinePrefix marks an issue's Message when the issue's Code is in
// the spec's `.baseline.json`. Gate() filters out such issues so they
// don't block advance/approve. `free-kiro lint` output still prints
// them (with the prefix) so authors see what they're carrying.
const baselinePrefix = "[baseline] "

// Spec runs every rule over one spec's documents and applies the
// per-spec `.baseline.json` whitelist: issues whose Code is in
// `ignored_issues` are still returned but their Message is prefixed
// with `[baseline]` and Gate() filters them out. See SpecStrict for
// the `--strict-baseline` variant.
//
// Missing documents are flagged as ERROR / WARNING depending on whether
// the spec was expected to have written them yet (see Gate for the
// gating subset). The branch on spec_type picks the right rule for the
// analysis doc: feature specs use requirements.md + EARS; bugfix specs
// use bugfix.md + the Current/Expected/Unchanged contract.
//
// Returns issues in a deterministic order: type-dependent doc first,
// then tasks.md, then design.md warnings.
func Spec(specDir string) []Issue {
	return specInternal(specDir, true)
}

// SpecStrict runs lint without consulting the spec's `.baseline.json`.
// Triggered by `--strict-baseline`; useful when an author wants to
// re-survey the full set of findings (e.g. before deleting an entry
// from the baseline).
func SpecStrict(specDir string) []Issue {
	return specInternal(specDir, false)
}

func specInternal(specDir string, applyBaseline bool) []Issue {
	var out []Issue
	var base Baseline
	if applyBaseline {
		b, err := LoadBaseline(specDir)
		if err != nil {
			// Surface the parse error as a single ERROR so the user
			// knows the baseline didn't apply (rather than silently
			// re-flagging every ignored issue).
			out = append(out, Issue{
				Severity: SeverityError,
				Code:     "baseline-parse-error",
				Message:  err.Error(),
				Location: BaselineFileName,
			})
		} else {
			base = b
		}
	}

	specType := SpecTypeFor(specDir)
	firstDoc := models.FirstPlanningDoc(specType)

	firstPath := filepath.Join(specDir, firstDoc)
	if data, err := os.ReadFile(firstPath); err == nil {
		text := string(data)
		if specType == models.SpecTypeBugfix {
			out = append(out, Bugfix(text)...)
		} else {
			out = append(out, Requirements(text)...)
		}
	} else {
		out = append(out, Issue{
			Severity: SeverityError,
			Code:     "missing-" + stem(firstDoc),
			Message:  firstDoc + " not found",
			Location: firstDoc,
		})
	}

	tasksPath := filepath.Join(specDir, "tasks.md")
	if data, err := os.ReadFile(tasksPath); err == nil {
		out = append(out, Tasks(string(data))...)
	} else {
		out = append(out, Issue{
			Severity: SeverityWarning,
			Code:     "missing-tasks",
			Message:  "tasks.md not found",
			Location: "tasks.md",
		})
	}

	designPath := filepath.Join(specDir, "design.md")
	if _, err := os.Stat(designPath); os.IsNotExist(err) {
		out = append(out, Issue{
			Severity: SeverityWarning,
			Code:     "missing-design",
			Message:  "design.md not found (spec may be incomplete)",
			Location: "design.md",
		})
	}

	if applyBaseline {
		for i := range out {
			if base.ShouldIgnore(out[i].Code) {
				out[i].Message = baselinePrefix + out[i].Message
			}
		}
	}
	return out
}

// Gate returns the ERROR findings that should block advance/approve.
// "missing-*" findings (a phase the author has not written yet) and
// `[baseline]`-prefixed issues (whitelisted by `.baseline.json`) are
// excluded from the gate — neither should block a spec the author is
// still drafting or has knowingly accepted.
func Gate(specDir string) []Issue {
	var gate []Issue
	for _, i := range Spec(specDir) {
		if i.Severity != SeverityError {
			continue
		}
		if isMissingCode(i.Code) {
			continue
		}
		if strings.HasPrefix(i.Message, baselinePrefix) {
			continue
		}
		gate = append(gate, i)
	}
	return gate
}

// isMissingCode reports whether a lint code marks a missing-doc finding.
// `missing-` findings are excluded from the advance/approve gate (see
// Gate) — you cannot be failed for a document that does not exist.
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
