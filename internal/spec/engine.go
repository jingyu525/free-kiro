package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// Engine is the spec lifecycle orchestrator. All state-machine transitions
// route through it; every advance/approve calls lint.LintGate first.
type Engine struct {
	ws *workspace.Workspace
}

// New constructs an Engine for the given workspace. Use Workspace.Require
// upstream if you want a guaranteed-present .kiro.
func New(ws *workspace.Workspace) *Engine {
	return &Engine{ws: ws}
}

// WS exposes the underlying workspace (CLI layer uses this to load
// settings and find spec dirs).
func (e *Engine) WS() *workspace.Workspace { return e.ws }

// NewSpec creates a fresh spec directory and writes .meta.json. The spec
// starts in PhaseDraft; the workflow/spec_type/quick flags are persisted
// for later phases.
//
// Errors:
//   - spec directory already exists → UsageError (name collision)
//   - workspace is not bootstrapped → WorkspaceError
func (e *Engine) NewSpec(name, prompt, workflow, specType string, quick bool) (*models.SpecMeta, error) {
	if name == "" {
		return nil, ferrors.Wrap("spec.new", nil, "spec name is required")
	}
	if workflow == "" {
		workflow = models.WorkflowRequirementsFirst
	}
	if specType == "" {
		specType = models.SpecTypeFeature
	}
	if workflow != models.WorkflowRequirementsFirst && workflow != models.WorkflowDesignFirst {
		return nil, ferrors.Wrap("spec.new", nil,
			"workflow must be requirements-first or design-first; got "+workflow)
	}
	if specType != models.SpecTypeFeature && specType != models.SpecTypeBugfix {
		return nil, ferrors.Wrap("spec.new", nil,
			"spec_type must be feature or bugfix; got "+specType)
	}

	dir := e.ws.SpecDir(name)
	if _, err := os.Stat(filepath.Join(dir, models.MetaFileName)); err == nil {
		return nil, ferrors.Wrap("spec.new", nil,
			"spec "+name+" already exists at "+dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, ferrors.Wrap("spec.new", err, "mkdir "+dir)
	}

	meta := models.NewSpecMeta(name, prompt, workflow, specType, quick)
	if err := meta.Save(dir); err != nil {
		return nil, err
	}
	// Mark as the active spec so the IDE SessionStart hook prints the
	// right `next` action. Best-effort: failure here doesn't block
	// spec creation (the user can pick the spec manually).
	if err := e.ws.WriteCurrent(name); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not mark %q as active: %v\n", name, err)
	}
	return meta, nil
}

// Generate writes the document for a single phase. Safe to call multiple
// times for the same phase: existing files are left alone unless force is
// true. Advances the phase pointer to `phase` after writing.
//
// If the spec is in a phase that cannot legally reach `phase`, the
// transition is rejected (TransitionError) before any disk write — the
// gate runs first to keep state machine and file system in lockstep.
func (e *Engine) Generate(specName string, phase models.Phase, force bool) (*models.SpecMeta, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	if err := models.AssertTransition(meta.PhaseEnum(), phase); err != nil {
		return nil, ferrors.Wrap("spec.generate", err, "phase transition rejected")
	}

	dir := e.ws.SpecDir(specName)
	if _, err := generateOne(dir, meta.Name, meta.SpecType, phase, force); err != nil {
		return nil, err
	}

	meta.Phase = phase
	if err := meta.Save(dir); err != nil {
		return nil, err
	}
	return meta, nil
}

// GenerateAll writes every planning document in workflow order. Stops at
// the first failure (so the caller can retry without --force).
func (e *Engine) GenerateAll(specName string, force bool) ([]string, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	dir := e.ws.SpecDir(specName)
	var written []string
	for _, doc := range models.PlanningOrder(meta.Workflow, meta.SpecType) {
		phase := phaseFromDoc(doc)
		if err := models.AssertTransition(meta.PhaseEnum(), phase); err != nil {
			return written, ferrors.Wrap("spec.generate", err, "phase transition rejected")
		}
		path, err := generateOne(dir, meta.Name, meta.SpecType, phase, force)
		if err != nil {
			return written, err
		}
		written = append(written, path)
		meta.Phase = phase
		if err := meta.Save(dir); err != nil {
			return written, err
		}
	}
	return written, nil
}

// Approve marks a spec approved and captures a drift baseline (AC count +
// task count). Runs the lint gate first; returns LintGateError if any
// non-missing ERROR is present in the planning documents.
//
// After Approve, phase == APPROVED.
func (e *Engine) Approve(specName string) (*models.SpecMeta, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	// Gate: refuse to approve a spec with structural defects.
	if gate := lint.LintGate(e.ws.SpecDir(specName)); len(gate) > 0 {
		return nil, &ferrors.LintGateError{ferrors.Wrap(
			"spec.approve",
			nil,
			"lint gate blocked approval: "+formatGate(gate),
		)}
	}
	if err := models.AssertTransition(meta.PhaseEnum(), models.PhaseApproved); err != nil {
		return nil, ferrors.Wrap("spec.approve", err, "phase transition rejected")
	}

	// Capture baseline so future drift checks can compare.
	meta.Baseline = captureBaseline(e.ws.SpecDir(specName), meta.SpecType)
	meta.Phase = models.PhaseApproved
	meta.Approved = true
	if err := meta.Save(e.ws.SpecDir(specName)); err != nil {
		return nil, err
	}
	return meta, nil
}

// Start marks the spec as actively being implemented. Pure bookkeeping —
// the engine does not run code. Phase moves APPROVED → IMPLEMENTING.
func (e *Engine) Start(specName string) (*models.SpecMeta, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	if err := models.AssertTransition(meta.PhaseEnum(), models.PhaseImplementing); err != nil {
		return nil, ferrors.Wrap("spec.start", err, "phase transition rejected")
	}
	meta.Phase = models.PhaseImplementing
	if err := meta.Save(e.ws.SpecDir(specName)); err != nil {
		return nil, err
	}
	return meta, nil
}

// Complete marks the spec done. Phase moves IMPLEMENTING → DONE.
// Clears .current so the SessionStart hook doesn't keep recommending
// actions for a finished spec.
func (e *Engine) Complete(specName string) (*models.SpecMeta, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	if err := models.AssertTransition(meta.PhaseEnum(), models.PhaseDone); err != nil {
		return nil, ferrors.Wrap("spec.complete", err, "phase transition rejected")
	}
	meta.Phase = models.PhaseDone
	if err := meta.Save(e.ws.SpecDir(specName)); err != nil {
		return nil, err
	}
	if e.ws.ReadCurrent() == specName {
		_ = e.ws.ClearCurrent()
	}
	return meta, nil
}

// Sync re-baselines the spec to the current state of its documents.
// Useful after the author has legitimately edited AC / task lists and
// wants to silence the drift warning.
func (e *Engine) Sync(specName string) (*models.SpecMeta, error) {
	meta, err := e.loadMeta(specName)
	if err != nil {
		return nil, err
	}
	meta.Baseline = captureBaseline(e.ws.SpecDir(specName), meta.SpecType)
	if err := meta.Save(e.ws.SpecDir(specName)); err != nil {
		return nil, err
	}
	return meta, nil
}

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
	gate := lint.LintGate(dir)

	action := map[string]any{
		"spec":     meta.Name,
		"phase":    string(meta.Phase),
		"workflow": meta.Workflow,
		"spec_type": meta.SpecType,
		"quick":    meta.Quick,
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

// loadMeta reads .meta.json from the spec's directory.
func (e *Engine) loadMeta(specName string) (*models.SpecMeta, error) {
	return models.LoadSpecMeta(e.ws.SpecDir(specName))
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
func formatGate(issues []lint.LintIssue) string {
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

// countTasks / countWaves parse tasks.md just enough to report progress.
// The CLI's `task list` command uses taskgraph directly for the full
// dependency-aware wave view; this is the lightweight summary used in
// status / analyze output.
func countTasks(specDir string) (done, total int) {
	data, err := os.ReadFile(filepath.Join(specDir, "tasks.md"))
	if err != nil {
		return 0, 0
	}
	for _, line := range splitLines(string(data)) {
		m := taskLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		total++
		if m[1] == "x" || m[1] == "X" {
			done++
		}
	}
	return done, total
}

// countWaves returns the number of topological levels in tasks.md. Runs
// the dependency-aware topological leveller via taskgraph.
func countWaves(specDir string) int {
	data, err := os.ReadFile(filepath.Join(specDir, "tasks.md"))
	if err != nil {
		return 0
	}
	tasks := taskgraph.ParseTasks(string(data))
	return len(taskgraph.ExecutionWaves(tasks))
}

var (
	taskLineRe = regexp.MustCompile(`^\s*-\s*\[( |x|X)\]\s*#(\d+)\s+`)
)

func splitLines(text string) []string {
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

// Ensure json import is referenced (build-only).
var _ = json.Marshal