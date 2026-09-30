package lint

import (
	"errors"
	"fmt"
	"io/fs"
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
	data, err := os.ReadFile(firstPath)
	switch {
	case err == nil:
		text := string(data)
		if specType == models.SpecTypeBugfix {
			out = append(out, Bugfix(text)...)
		} else {
			out = append(out, Requirements(text)...)
		}
	case errors.Is(err, fs.ErrNotExist):
		out = append(out, Issue{
			Severity: SeverityError,
			Code:     "missing-" + stem(firstDoc),
			Message:  firstDoc + " not found",
			Location: firstDoc,
		})
	default:
		// Non-NotExist IO error (EACCES, EISDIR, …). Surface the
		// underlying message instead of misleadingly reporting
		// "missing" (F1: lint IO error must wrap, not swallow).
		out = append(out, Issue{
			Severity: SeverityError,
			Code:     "lint-" + stem(firstDoc) + "-read-error",
			Message:  fmt.Errorf("read %s: %w", firstPath, err).Error(),
			Location: firstDoc,
		})
	}

	tasksPath := filepath.Join(specDir, "tasks.md")
	tasksData, tasksErr := os.ReadFile(tasksPath)
	switch {
	case tasksErr == nil:
		out = append(out, Tasks(string(tasksData))...)
	case errors.Is(tasksErr, fs.ErrNotExist):
		out = append(out, Issue{
			Severity: SeverityWarning,
			Code:     "missing-tasks",
			Message:  "tasks.md not found",
			Location: "tasks.md",
		})
	default:
		out = append(out, Issue{
			Severity: SeverityWarning,
			Code:     "lint-tasks-read-error",
			Message:  fmt.Errorf("read %s: %w", tasksPath, tasksErr).Error(),
			Location: "tasks.md",
		})
	}

	designPath := filepath.Join(specDir, "design.md")
	if _, statErr := os.Stat(designPath); statErr != nil {
		switch {
		case errors.Is(statErr, fs.ErrNotExist):
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "missing-design",
				Message:  "design.md not found (spec may be incomplete)",
				Location: "design.md",
			})
		default:
			out = append(out, Issue{
				Severity: SeverityWarning,
				Code:     "lint-design-stat-error",
				Message:  fmt.Errorf("stat %s: %w", designPath, statErr).Error(),
				Location: "design.md",
			})
		}
	}

	if applyBaseline {
		for i := range out {
			if base.ShouldIgnore(out[i].Code) {
				out[i].Baseline = true
				// Prepend the legacy "[baseline] " prefix to Message so
				// `free-kiro lint` output stays human-readable; Gate()
				// does NOT rely on this prefix (it reads Issue.Baseline).
				out[i].Message = baselinePrefix + out[i].Message
			}
		}
	}
	return out
}

// Gate returns the ERROR findings that should block advance/approve.
// "missing-*" findings (a phase the author has not written yet) and
// issues flagged with Issue.Baseline (whitelisted by `.baseline.json`)
// are excluded from the gate — neither should block a spec the author
// is still drafting or has knowingly accepted.
//
// Implementation note: we filter on Issue.Baseline (an explicit field)
// rather than inspecting the Message prefix. The Message is still
// prefixed with `[baseline] ` for human-readable CLI output, but a
// user-authored Message that happens to start with that string cannot
// accidentally bypass the gate (the L3 fix).
func Gate(specDir string) []Issue {
	var gate []Issue
	for _, i := range Spec(specDir) {
		if i.Severity != SeverityError {
			continue
		}
		if isMissingCode(i.Code) {
			continue
		}
		if i.Baseline {
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
	if before, _, ok := strings.Cut(base, "."); ok {
		return before
	}
	return base
}
