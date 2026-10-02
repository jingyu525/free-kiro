package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
)

// Status returns a structured snapshot of the spec's lifecycle position
// plus drift signals (compared to the baseline). JSON-serialisable so
// the CLI can pipe it to jq / IDE hooks.
func (e *Engine) Status(specName string) (map[string]any, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	dir := e.ws.SpecDir(specName)
	current := captureBaseline(dir, meta.SpecType)
	drift := computeDrift(meta.Baseline, current)

	done, total := countTasks(dir)
	waves := countWaves(dir)

	status := map[string]any{
		"name":       meta.Name,
		"phase":      string(meta.Phase),
		"workflow":   meta.Workflow,
		"spec_type":  meta.SpecType,
		"quick":      meta.Quick,
		"approved":   meta.Approved,
		"baseline":   meta.Baseline,
		"current":    current,
		"drift":      drift,
		"tasks":      map[string]any{"done": done, "total": total, "waves": waves},
		"created_at": meta.CreatedAt,
		"updated_at": meta.UpdatedAt,
	}
	return status, nil
}

// NextAction is the oracle: returns the recommended next command for
// the spec's current phase + lint state. Used by IDE SessionStart hooks
// to keep the agent oriented.
func (e *Engine) NextAction(specName string) (map[string]any, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	dir := e.ws.SpecDir(specName)
	gate := lint.Gate(dir)

	action := map[string]any{
		"spec":      meta.Name,
		"phase":     string(meta.Phase),
		"workflow":  meta.Workflow,
		"spec_type": meta.SpecType,
		"quick":     meta.Quick,
	}
	switch meta.Phase {
	case models.PhaseDraft, models.PhaseRequirements:
		action["suggested"] = "generate-requirements-or-design"
		action["command"] = "free-kiro spec generate " + meta.Name + " --phase all"
	case models.PhaseDesign, models.PhaseTasks:
		action["suggested"] = "review-and-iterate"
		action["command"] = "free-kiro spec generate " + meta.Name + " --phase " + nextPhase(meta.PhaseEnum())
	case models.PhaseApproved:
		if meta.Quick {
			// Quick spec waives the formal approve gate.
			action["suggested"] = "start-implementation"
			action["command"] = "free-kiro spec start " + meta.Name
		} else {
			action["suggested"] = "start-implementation"
			action["command"] = "free-kiro spec start " + meta.Name
		}
	case models.PhaseImplementing:
		action["suggested"] = "complete-when-done"
		action["command"] = "free-kiro spec complete " + meta.Name
	case models.PhaseDone:
		action["suggested"] = "spec-done"
		action["command"] = ""
	}
	if len(gate) > 0 && meta.Phase != models.PhaseDone {
		action["lint_errors"] = len(gate)
		action["lint_command"] = "free-kiro lint " + meta.Name
	}
	return action, nil
}

// Show prints the contents of one planning document (or "(not generated)"
// when the file does not exist).
func (e *Engine) Show(specName string, phase models.Phase) (string, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return "", err
	}
	doc := models.PhaseDocFor(meta.SpecType, phase)
	path := filepath.Join(e.ws.SpecDir(specName), doc)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "(not generated)", nil
		}
		return "", ferrors.Wrap("spec.show", err, "read "+path)
	}
	return string(data), nil
}

// ListSpecs enumerates all specs in the workspace. Returns the metadata
// for each one whose .meta.json parses correctly; skips orphans silently.
func (e *Engine) ListSpecs() ([]*models.SpecMeta, error) {
	dir := e.ws.SpecsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, ferrors.Wrap("spec.list", err, "read "+dir)
	}
	var out []*models.SpecMeta
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		meta, err := models.LoadSpecMeta(filepath.Join(dir, ent.Name()))
		if err != nil {
			continue
		}
		out = append(out, meta)
	}
	return out, nil
}

// StatusForList returns one row per spec for `spec list`. Kept separate
// from Status() to avoid mixing the JSON-shape needs of the two callers.
// Adds an "active" flag so the CLI can highlight the spec the user is
// currently working on (read from .kiro/.current).
func (e *Engine) StatusForList() ([]map[string]any, error) {
	specs, err := e.ListSpecs()
	if err != nil {
		return nil, err
	}
	current := e.ws.ReadCurrent()
	var out []map[string]any
	for _, m := range specs {
		out = append(out, map[string]any{
			"name":     m.Name,
			"phase":    string(m.Phase),
			"approved": m.Approved,
			"active":   m.Name == current,
		})
	}
	return out, nil
}

// phaseFromDoc maps a planning-doc filename back to its phase.
func phaseFromDoc(doc string) models.Phase {
	switch doc {
	case "requirements.md", "bugfix.md":
		return models.PhaseRequirements
	case "design.md":
		return models.PhaseDesign
	case "tasks.md":
		return models.PhaseTasks
	}
	return models.PhaseDraft
}

// nextPhase returns the next-plausible phase for an in-progress spec.
func nextPhase(p models.Phase) string {
	switch p {
	case models.PhaseRequirements:
		return string(models.PhaseDesign)
	case models.PhaseDesign:
		return string(models.PhaseTasks)
	case models.PhaseTasks:
		return string(models.PhaseApproved)
	}
	return string(p)
}

// formatGate renders a one-line summary of lint-gate failures for the
// "approval blocked" message. Pure string-builder, no JSON.
func formatGate(issues []lint.Issue) string {
	if len(issues) == 0 {
		return "no errors"
	}
	var parts []string
	for _, i := range issues {
		parts = append(parts, i.Code)
		if len(parts) >= 5 {
			parts = append(parts, fmt.Sprintf("…(%d more)", len(issues)-5))
			break
		}
	}
	return strings.Join(parts, ", ")
}

// ErrSpecNotFound is returned by Engine methods when the named spec
// does not exist on disk. Callers map this to a 404 in the HTTP layer.
// Defined here (not in errors.go) because only spec-engine code raises
// it; centralising in errors.go would invite unrelated packages to
// fabricate it.
var ErrSpecNotFound = errors.New("spec not found")

// Wave groups tasks that share the same topological level and may
// run concurrently. Index is 1-based; Done / Total are pre-computed
// so handlers don't have to re-count on every render.
type Wave struct {
	Index int
	Tasks []models.Task
	Done  int
	Total int
}

// TaskList returns the full task structure for one spec, parsed
// from tasks.md and grouped into execution waves. Reuses
// taskgraph.ParseTasks + ExecutionWaves (no re-implementation).
//
// Spec missing → ErrSpecNotFound (handler maps to 404).
// tasks.md missing or empty → ([]Wave{}, nil) — a spec still in
// planning has no tasks yet, this is not an error.
// tasks.md unparseable → wrapped error so callers can surface a 500.
func (e *Engine) TaskList(specName string) ([]Wave, error) {
	specDir := e.ws.SpecDir(specName)
	if _, err := os.Stat(specDir); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSpecNotFound
		}
		return nil, ferrors.Wrap("spec.tasklist", err, "stat "+specDir)
	}
	tasksPath := filepath.Join(specDir, "tasks.md")
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Wave{}, nil
		}
		return nil, ferrors.Wrap("spec.tasklist", err, "read "+tasksPath)
	}
	tasks := taskgraph.ParseTasks(string(data))
	if len(tasks) == 0 {
		return []Wave{}, nil
	}
	groups := taskgraph.ExecutionWaves(tasks)
	if groups == nil {
		// Cycle detected — taskgraph.ExecutionWaves returns nil on cycle.
		// Lint will have surfaced it; here we return empty so the dashboard
		// renders a blank task list rather than crashing.
		return []Wave{}, nil
	}
	out := make([]Wave, 0, len(groups))
	for i, g := range groups {
		done := 0
		for _, t := range g {
			if t.Done {
				done++
			}
		}
		out = append(out, Wave{
			Index: i + 1,
			Tasks: g,
			Done:  done,
			Total: len(g),
		})
	}
	return out, nil
}
