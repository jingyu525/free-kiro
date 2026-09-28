package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCanTransition_Legal(t *testing.T) {
	legal := []struct{ from, to Phase }{
		{PhaseDraft, PhaseRequirements},
		{PhaseDraft, PhaseDesign},
		{PhaseRequirements, PhaseDesign},
		{PhaseRequirements, PhaseTasks},
		{PhaseDesign, PhaseRequirements},
		{PhaseDesign, PhaseTasks},
		{PhaseTasks, PhaseApproved},
		{PhaseTasks, PhaseRequirements},
		{PhaseTasks, PhaseDesign},
		{PhaseApproved, PhaseImplementing},
		{PhaseApproved, PhaseTasks},
		{PhaseImplementing, PhaseDone},
	}
	for _, c := range legal {
		if !CanTransition(c.from, c.to) {
			t.Errorf("CanTransition(%s -> %s) should be true", c.from, c.to)
		}
		if err := AssertTransition(c.from, c.to); err != nil {
			t.Errorf("AssertTransition(%s -> %s) should not error: %v", c.from, c.to, err)
		}
	}
}

func TestCanTransition_Illegal(t *testing.T) {
	illegal := []struct{ from, to Phase }{
		{PhaseDraft, PhaseTasks},        // skip planning phases
		{PhaseDraft, PhaseApproved},     // skip planning + approval
		{PhaseDraft, PhaseImplementing}, // no direct jump
		{PhaseRequirements, PhaseApproved},
		{PhaseDesign, PhaseApproved},
		{PhaseApproved, PhaseDone},       // must go via IMPLEMENTING
		{PhaseImplementing, PhaseApproved},
		{PhaseImplementing, PhaseRequirements},
		{PhaseDone, PhaseImplementing}, // DONE is terminal
		{PhaseDone, PhaseDraft},
	}
	for _, c := range illegal {
		if CanTransition(c.from, c.to) {
			t.Errorf("CanTransition(%s -> %s) should be false", c.from, c.to)
		}
		if err := AssertTransition(c.from, c.to); err == nil {
			t.Errorf("AssertTransition(%s -> %s) should error", c.from, c.to)
		}
	}
}

func TestPhaseOrder_Monotonic(t *testing.T) {
	phases := AllPhases()
	if len(phases) < 2 {
		t.Fatal("expected at least 2 phases")
	}
	for i := 1; i < len(phases); i++ {
		if phases[i].Order() <= phases[i-1].Order() {
			t.Errorf("phase order not monotonic: %s(%d) <= %s(%d)",
				phases[i], phases[i].Order(), phases[i-1], phases[i-1].Order())
		}
	}
}

func TestPhaseDocFor_FeatureVsBugfix(t *testing.T) {
	if got := PhaseDocFor("feature", PhaseRequirements); got != "requirements.md" {
		t.Errorf("feature requirements doc should be requirements.md; got %q", got)
	}
	if got := PhaseDocFor("bugfix", PhaseRequirements); got != "bugfix.md" {
		t.Errorf("bugfix requirements doc should be bugfix.md; got %q", got)
	}
	// design and tasks are type-agnostic.
	if got := PhaseDocFor("bugfix", PhaseDesign); got != "design.md" {
		t.Errorf("bugfix design doc should be design.md; got %q", got)
	}
	if got := PhaseDocFor("bugfix", PhaseTasks); got != "tasks.md" {
		t.Errorf("bugfix tasks doc should be tasks.md; got %q", got)
	}
}

func TestPlanningOrder_RequirementsFirst(t *testing.T) {
	got := PlanningOrder(WorkflowRequirementsFirst, SpecTypeFeature)
	want := []string{"requirements.md", "design.md", "tasks.md"}
	if !equalStrings(got, want) {
		t.Errorf("requirements-first feature: got %v, want %v", got, want)
	}
}

func TestPlanningOrder_DesignFirst(t *testing.T) {
	got := PlanningOrder(WorkflowDesignFirst, SpecTypeFeature)
	want := []string{"design.md", "requirements.md", "tasks.md"}
	if !equalStrings(got, want) {
		t.Errorf("design-first feature: got %v, want %v", got, want)
	}
}

func TestPlanningOrder_BugfixFirst(t *testing.T) {
	got := PlanningOrder(WorkflowRequirementsFirst, SpecTypeBugfix)
	want := []string{"bugfix.md", "design.md", "tasks.md"}
	if !equalStrings(got, want) {
		t.Errorf("requirements-first bugfix: got %v, want %v", got, want)
	}
}

func TestNewSpecMeta_Defaults(t *testing.T) {
	m := NewSpecMeta("user-auth", "add login", "", "", false)
	if m.Name != "user-auth" {
		t.Errorf("name: got %q", m.Name)
	}
	if m.Phase != PhaseDraft {
		t.Errorf("phase: got %s, want draft", m.Phase)
	}
	if m.Workflow != WorkflowRequirementsFirst {
		t.Errorf("workflow: got %q, want %q", m.Workflow, WorkflowRequirementsFirst)
	}
	if m.SpecType != SpecTypeFeature {
		t.Errorf("spec_type: got %q, want %q", m.SpecType, SpecTypeFeature)
	}
	if m.Approved {
		t.Error("new spec should not be approved")
	}
	if m.CreatedAt == "" || m.UpdatedAt == "" {
		t.Error("timestamps should be populated")
	}
}

func TestNewSpecMeta_RespectsExplicit(t *testing.T) {
	m := NewSpecMeta("fix-login", "broken redirect", WorkflowDesignFirst, SpecTypeBugfix, true)
	if m.Workflow != WorkflowDesignFirst {
		t.Errorf("workflow: got %q", m.Workflow)
	}
	if m.SpecType != SpecTypeBugfix {
		t.Errorf("spec_type: got %q", m.SpecType)
	}
	if !m.Quick {
		t.Error("quick flag lost")
	}
}

func TestSpecMeta_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	m := NewSpecMeta("demo", "test prompt", WorkflowDesignFirst, SpecTypeBugfix, true)
	m.Baseline = map[string]int{"ac_count": 5, "task_count": 8}
	if err := m.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, MetaFileName)); err != nil {
		t.Fatalf("meta.json not written: %v", err)
	}
	got, err := LoadSpecMeta(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Name != m.Name || got.Phase != m.Phase || got.Workflow != m.Workflow {
		t.Errorf("roundtrip mismatch: %+v vs %+v", got, m)
	}
	if got.Baseline["ac_count"] != 5 || got.Baseline["task_count"] != 8 {
		t.Errorf("baseline lost: %v", got.Baseline)
	}
}

func TestLoadSpecMeta_Malformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, MetaFileName), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadSpecMeta(dir); err == nil {
		t.Fatal("malformed meta.json should produce an error")
	}
}

func TestApplyDefaults_HandlesMissingFields(t *testing.T) {
	// Simulate a meta.json written by an older / partial tool.
	raw := `{"name":"x","phase":"design"}`
	var m SpecMeta
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m.applyDefaults()
	if m.Workflow != WorkflowRequirementsFirst {
		t.Errorf("workflow default missing")
	}
	if m.SpecType != SpecTypeFeature {
		t.Errorf("spec_type default missing")
	}
	if m.Generator != GeneratorTemplate {
		t.Errorf("generator default missing")
	}
	if m.Baseline == nil {
		t.Error("baseline should be initialized to empty map, not nil")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}