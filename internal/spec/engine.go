package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// Engine is the spec lifecycle orchestrator. All state-machine transitions
// route through it; every advance/approve calls lint.LintGate first.
//
// File layout (Wave 5 refactor):
//
//	engine.go        — Engine struct + lifecycle (NewSpec / Generate /
//	                   Approve / Start / Complete / Sync)
//	engine_state.go  — read-only queries (Status / NextAction / Show /
//	                   ListSpecs / StatusForList) + phase helpers
//	engine_io.go     — IO + counting helpers (loadMeta / countTasks /
//	                   countWaves / taskLineRe / splitLines)
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
		return nil, &ferrors.LintGateError{KiroError: ferrors.Wrap(
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

// Ensure json import stays referenced after splitting files.
var _ = json.Marshal
