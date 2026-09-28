// Package models defines the single source of truth for the *shape* of a
// spec-driven workflow: the phase state machine, spec metadata, and the
// structural records (Task, SteeringDoc, Hook) the engine manipulates.
//
// This file is the phase state machine — the most critical piece. Every
// illegal phase jump is caught at the boundary here so commands cannot
// produce half-written specs.
package models

import (
	ferrors "github.com/liujingyu/free-kiro/internal/errors"
)

// Phase is the workflow position of a spec. The string values are the wire
// format — they appear in .meta.json, CLI output, and hook payloads — so
// changing them is a breaking change.
type Phase string

const (
	PhaseDraft        Phase = "draft"
	PhaseRequirements Phase = "requirements"
	PhaseDesign       Phase = "design"
	PhaseTasks        Phase = "tasks"
	PhaseApproved     Phase = "approved"
	PhaseImplementing Phase = "implementing"
	PhaseDone         Phase = "done"
)

// Order returns the linear ordering of phases for display purposes. The
// _ALLOWED table (below) governs *legal transitions*, which is a richer
// relation than this linear order — a spec can iterate between planning
// phases before approval.
func (p Phase) Order() int {
	return phaseOrder[p]
}

var phaseOrder = map[Phase]int{
	PhaseDraft:        0,
	PhaseRequirements: 1,
	PhaseDesign:       2,
	PhaseTasks:        3,
	PhaseApproved:     4,
	PhaseImplementing: 5,
	PhaseDone:         6,
}

// AllPhases returns every phase in linear order (useful for tests and
// status displays).
func AllPhases() []Phase {
	return []Phase{
		PhaseDraft, PhaseRequirements, PhaseDesign, PhaseTasks,
		PhaseApproved, PhaseImplementing, PhaseDone,
	}
}

// allowed maps each phase to the set of phases it may legally transition
// to. The three planning phases (requirements / design / tasks) form a
// bounded, fully-reconnected graph: a spec may enter from DRAFT via either
// requirements or design, and may iterate between the three planning docs
// in either order before reaching APPROVED. Once approved, only TASKS
// (re-edit) or IMPLEMENTING are reachable; once DONE, no transitions.
//
// Forward advances through this graph are additionally gated by lint
// (see internal/lint); this table only encodes *which transitions exist*,
// not whether they're clean.
var allowed = map[Phase]map[Phase]bool{
	PhaseDraft: {
		PhaseRequirements: true,
		PhaseDesign:       true,
	},
	PhaseRequirements: {
		PhaseRequirements: true, // iterate / refine
		PhaseDesign:       true,
		PhaseTasks:        true,
	},
	PhaseDesign: {
		PhaseDesign:       true, // iterate / refine
		PhaseRequirements: true,
		PhaseTasks:        true,
	},
	PhaseTasks: {
		PhaseTasks:        true, // iterate / refine
		PhaseApproved:     true,
		PhaseRequirements: true,
		PhaseDesign:       true,
	},
	PhaseApproved: {
		PhaseTasks:        true, // last-minute task tweaks before start
		PhaseImplementing: true,
	},
	PhaseImplementing: {
		PhaseImplementing: true, // in-progress markers
		PhaseDone:         true,
	},
	PhaseDone: {}, // terminal
}

// CanTransition reports whether a phase may legally move to another.
func CanTransition(from, to Phase) bool {
	next, ok := allowed[from]
	if !ok {
		return false
	}
	return next[to]
}

// AssertTransition raises TransitionError if the move is illegal.
func AssertTransition(from, to Phase) error {
	if !CanTransition(from, to) {
		legal := make([]string, 0)
		for p := range allowed[from] {
			legal = append(legal, string(p))
		}
		if len(legal) == 0 {
			return ferrors.Wrap(
				"phase.transition", nil,
				"illegal phase transition: "+string(from)+" -> "+string(to)+" (terminal)",
			)
		}
		return ferrors.Wrap(
			"phase.transition", nil,
			"illegal phase transition: "+string(from)+" -> "+string(to)+"; allowed: "+joinPhases(legal),
		)
	}
	return nil
}

// PhaseDocName returns the document filename produced for a given phase in
// a feature spec. Bug-fix specs override REQUIREMENTS → "bugfix.md"
// (see PhaseDocFor).
func PhaseDocName(p Phase) string {
	return phaseDoc[p]
}

var phaseDoc = map[Phase]string{
	PhaseRequirements: "requirements.md",
	PhaseDesign:       "design.md",
	PhaseTasks:        "tasks.md",
}

// PhaseDocFor returns the document filename produced for p in a spec of the
// given type. Bugfix specs replace "requirements.md" with "bugfix.md" for
// the requirements phase; everything else is type-agnostic.
func PhaseDocFor(specType string, p Phase) string {
	if p == PhaseRequirements && specType == "bugfix" {
		return "bugfix.md"
	}
	return phaseDoc[p]
}

// FirstPlanningDoc returns the requirements-equivalent document name for a
// spec type. Bugfix specs produce bugfix.md; everything else produces
// requirements.md.
func FirstPlanningDoc(specType string) string {
	if specType == "bugfix" {
		return "bugfix.md"
	}
	return "requirements.md"
}

// PlanningOrder returns the ordered list of planning-document filenames for
// a spec, respecting the chosen workflow (requirements-first / design-first)
// and spec type. Both workflows converge on tasks.md.
func PlanningOrder(workflow, specType string) []string {
	first := FirstPlanningDoc(specType)
	if workflow == "design-first" {
		return []string{"design.md", first, "tasks.md"}
	}
	return []string{first, "design.md", "tasks.md"}
}

func joinPhases(s []string) string {
	out := ""
	for i, p := range s {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}