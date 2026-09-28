package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/liujingyu/free-kiro/internal/lint"
	"github.com/liujingyu/free-kiro/internal/models"
	"github.com/liujingyu/free-kiro/internal/workspace"
)

func newTestEngine(t *testing.T) (*Engine, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".kiro", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".kiro", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".kiro", "settings.json"),
		[]byte(`{"generator":"template"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ws := workspace.New(dir)
	return New(ws), dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestNewSpec_BasicAndDefaults(t *testing.T) {
	eng, _ := newTestEngine(t)
	meta, err := eng.NewSpec("demo", "add login", "", "", false)
	if err != nil {
		t.Fatalf("NewSpec: %v", err)
	}
	if meta.Name != "demo" || meta.Phase != models.PhaseDraft {
		t.Errorf("unexpected meta: %+v", meta)
	}
	if meta.Workflow != models.WorkflowRequirementsFirst {
		t.Errorf("default workflow wrong: %q", meta.Workflow)
	}
	if meta.SpecType != models.SpecTypeFeature {
		t.Errorf("default spec_type wrong: %q", meta.SpecType)
	}
}

func TestNewSpec_RejectsCollision(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("dup", "first", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.NewSpec("dup", "second", "", "", false); err == nil {
		t.Fatal("expected collision error")
	}
}

func TestNewSpec_RejectsInvalidWorkflow(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("x", "p", "nope", "", false); err == nil {
		t.Fatal("expected workflow validation error")
	}
}

func TestGenerate_AdvancesPhase(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.Generate("demo", models.PhaseRequirements, false); err != nil {
		t.Fatalf("Generate requirements: %v", err)
	}
	meta, err := eng.loadMeta("demo")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Phase != models.PhaseRequirements {
		t.Errorf("phase should advance to requirements; got %s", meta.Phase)
	}
	if !fileExists(filepath.Join(eng.WS().SpecDir("demo"), "requirements.md")) {
		t.Error("requirements.md should exist")
	}
}

func TestGenerate_RejectsIllegalTransition(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	// Draft cannot jump directly to tasks.
	if _, err := eng.Generate("demo", models.PhaseTasks, false); err == nil {
		t.Fatal("expected illegal transition error")
	}
}

func TestGenerateAll_RequirementsFirst(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", models.WorkflowRequirementsFirst, "", false); err != nil {
		t.Fatal(err)
	}
	paths, err := eng.GenerateAll("demo", false)
	if err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}
	if len(paths) != 3 {
		t.Errorf("expected 3 docs; got %d", len(paths))
	}
	want := []string{"requirements.md", "design.md", "tasks.md"}
	for i, p := range paths {
		base := filepath.Base(p)
		if base != want[i] {
			t.Errorf("doc %d: got %s, want %s", i, base, want[i])
		}
	}
	meta, _ := eng.loadMeta("demo")
	if meta.Phase != models.PhaseTasks {
		t.Errorf("phase should be tasks after generate all; got %s", meta.Phase)
	}
}

func TestGenerateAll_DesignFirst(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", models.WorkflowDesignFirst, "", false); err != nil {
		t.Fatal(err)
	}
	paths, err := eng.GenerateAll("demo", false)
	if err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}
	want := []string{"design.md", "requirements.md", "tasks.md"}
	for i, p := range paths {
		if filepath.Base(p) != want[i] {
			t.Errorf("doc %d: got %s, want %s", i, filepath.Base(p), want[i])
		}
	}
}

func TestGenerateAll_BugfixSpec(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("fix", "broken login", "", models.SpecTypeBugfix, false); err != nil {
		t.Fatal(err)
	}
	paths, err := eng.GenerateAll("fix", false)
	if err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}
	if filepath.Base(paths[0]) != "bugfix.md" {
		t.Errorf("bugfix spec first doc should be bugfix.md; got %s", filepath.Base(paths[0]))
	}
}

func TestApprove_LintGateBlocks(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	// Templates have TODO placeholders; no real EARS → no-ears ERROR → gate blocks.
	if _, err := eng.Approve("demo"); err == nil {
		t.Fatal("expected lint gate to block approval of template-only docs")
	}
}

func TestApprove_HappyPath(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	// Overwrite the templates with valid content.
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"), `# demo

## User Stories
As a user I want it.

## Acceptance Criteria
WHEN foo THE SYSTEM SHALL bar.
`)
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 First task\n")
	meta, err := eng.Approve("demo")
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if !meta.Approved || meta.Phase != models.PhaseApproved {
		t.Errorf("approve should set phase=approved and approved=true; got %+v", meta)
	}
	if meta.Baseline["ac_count"] != 1 {
		t.Errorf("baseline AC count: got %d, want 1", meta.Baseline["ac_count"])
	}
	if meta.Baseline["task_count"] != 1 {
		t.Errorf("baseline task count: got %d, want 1", meta.Baseline["task_count"])
	}
}

func TestApprove_QuickWaivesNothing(t *testing.T) {
	// Quick spec is approval-free at the GENERATION step, but the
	// approve gate still runs once the user invokes `spec approve`.
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", true); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.Approve("demo"); err == nil {
		t.Fatal("expected lint gate to block quick spec approval too")
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"), `# demo

## Acceptance Criteria
WHEN foo THE SYSTEM SHALL bar.
`)
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 a\n")
	if _, err := eng.Approve("demo"); err != nil {
		t.Fatalf("approve after fix: %v", err)
	}
}

func TestStartAndComplete(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"), `# demo

## Acceptance Criteria
WHEN foo THE SYSTEM SHALL bar.
`)
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 a\n")
	if _, err := eng.Approve("demo"); err != nil {
		t.Fatal(err)
	}
	meta, err := eng.Start("demo")
	if err != nil || meta.Phase != models.PhaseImplementing {
		t.Fatalf("Start: %v phase=%s", err, meta.Phase)
	}
	meta, err = eng.Complete("demo")
	if err != nil || meta.Phase != models.PhaseDone {
		t.Fatalf("Complete: %v phase=%s", err, meta.Phase)
	}
}

func TestStatus_JSON(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"), `# demo

## Acceptance Criteria
WHEN foo THE SYSTEM SHALL bar.
WHEN baz THE SYSTEM SHALL qux.
`)
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 a\n- [ ] #2 b [deps: #1]\n")
	if _, err := eng.Approve("demo"); err != nil {
		t.Fatal(err)
	}
	status, err := eng.Status("demo")
	if err != nil {
		t.Fatal(err)
	}
	if status["phase"] != "approved" {
		t.Errorf("phase: %v", status["phase"])
	}
	if status["tasks"].(map[string]any)["total"].(int) != 2 {
		t.Errorf("task total: %v", status["tasks"])
	}
	// Adding an AC should produce a drift signal.
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\nWHEN baz THE SYSTEM SHALL qux.\nWHEN new THE SYSTEM SHALL drift.\n")
	status2, _ := eng.Status("demo")
	drift := status2["drift"].([]DriftSignal)
	if len(drift) == 0 {
		t.Error("expected drift signal after adding AC")
	}
	// JSON-marshalable sanity check.
	if _, err := json.Marshal(status); err != nil {
		t.Errorf("status not JSON-marshalable: %v", err)
	}
}

func TestSync_RefreshesBaseline(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.GenerateAll("demo", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"), "# demo\nWHEN foo THE SYSTEM SHALL bar.\n")
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 a\n")
	if _, err := eng.Approve("demo"); err != nil {
		t.Fatal(err)
	}
	// Edit AC (legitimate change).
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\nWHEN foo THE SYSTEM SHALL bar.\nWHEN baz THE SYSTEM SHALL qux.\n")
	// Pre-sync drift should be non-empty.
	pre, _ := eng.Status("demo")
	if len(pre["drift"].([]DriftSignal)) == 0 {
		t.Fatal("expected drift before sync")
	}
	// Sync — drift should clear.
	if _, err := eng.Sync("demo"); err != nil {
		t.Fatal(err)
	}
	post, _ := eng.Status("demo")
	if len(post["drift"].([]DriftSignal)) != 0 {
		t.Errorf("expected no drift after sync; got %v", post["drift"])
	}
}

func TestNextAction_Phases(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	// Draft → suggest generate.
	na, _ := eng.NextAction("demo")
	if na["phase"] != "draft" {
		t.Errorf("phase: %v", na["phase"])
	}
	if na["command"] == "" {
		t.Errorf("expected non-empty command suggestion")
	}
}

func TestShow_NotGenerated(t *testing.T) {
	eng, _ := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	out, err := eng.Show("demo", models.PhaseRequirements)
	if err != nil {
		t.Fatal(err)
	}
	if out != "(not generated)" {
		t.Errorf("Show on empty spec: %q", out)
	}
}

func TestListSpecs(t *testing.T) {
	eng, _ := newTestEngine(t)
	eng.NewSpec("a", "a", "", "", false)
	eng.NewSpec("b", "b", "", "", false)
	metas, err := eng.ListSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 specs; got %d", len(metas))
	}
}

func TestAnalyze_TrivialSpec(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\n\n## User Stories\nAs a user I want it.\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\n")
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "- [ ] #1 a\n")
	findings := eng.Analyze("demo")
	for _, f := range findings {
		if f.Code == "tasks-without-requirements" {
			t.Errorf("tasks-without-requirements should not fire when AC is present")
		}
	}
}

func TestAnalyze_VagueLanguage(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL maybe do something user-friendly.\n")
	findings := eng.Analyze("demo")
	var found bool
	for _, f := range findings {
		if f.Code == "vague-language" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected vague-language finding; got %v", findings)
	}
}

func TestAnalyze_DuplicateACs(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\nWHEN foo THE SYSTEM SHALL bar.\n")
	findings := eng.Analyze("demo")
	var found bool
	for _, f := range findings {
		if f.Code == "duplicate-acceptance-criteria" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate-acceptance-criteria finding; got %v", findings)
	}
}

func TestAnalyze_UncoveredACs(t *testing.T) {
	eng, dir := newTestEngine(t)
	if _, err := eng.NewSpec("demo", "test", "", "", false); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/requirements.md"),
		"# demo\n\n## Acceptance Criteria\nWHEN foo THE SYSTEM SHALL bar.\nWHEN baz THE SYSTEM SHALL qux.\n")
	writeFile(t, filepath.Join(dir, ".kiro/specs/demo/tasks.md"), "# tasks\n\nNo items.\n")
	findings := eng.Analyze("demo")
	var found bool
	for _, f := range findings {
		if f.Code == "uncovered-acceptance-criteria" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected uncovered-acceptance-criteria finding; got %v", findings)
	}
}

// Use lint to avoid an unused-import warning on the multi-purpose linter.
var _ = lint.SeverityError

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}