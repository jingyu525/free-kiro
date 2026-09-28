package hooks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// setup creates a temp workspace with .kiro/hooks and returns a Registry.
func setup(t *testing.T) (*Registry, string) {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{".kiro/specs", ".kiro/steering", ".kiro/hooks"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".kiro/settings.json"),
		[]byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return NewRegistry(workspace.New(root)), root
}

func writeHook(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNormaliseHook_FlatShape(t *testing.T) {
	raw := rawHook{
		"id":          "lint-on-save",
		"event":       "file.save",
		"glob":        "*.go",
		"action_type": "shell",
		"action":      "gofmt -w $FILE",
		"description": "format Go files",
	}
	h, err := normaliseHook(raw)
	if err != nil {
		t.Fatal(err)
	}
	if h.ID != "lint-on-save" || h.Event != "file.save" || h.Glob != "*.go" ||
		h.ActionType != "shell" || h.Action != "gofmt -w $FILE" {
		t.Errorf("flat normalise wrong: %+v", h)
	}
	if !h.Enabled {
		t.Errorf("default should be enabled")
	}
}

func TestNormaliseHook_KiroV1Shape(t *testing.T) {
	raw := rawHook{
		"name":        "kiro-lint",
		"trigger":     "PostFileSave",
		"matcher":     `\.tsx$`,
		"action":      map[string]any{"type": "command", "command": "npx eslint --fix"},
		"description": "lint tsx",
	}
	h, err := normaliseHook(raw)
	if err != nil {
		t.Fatal(err)
	}
	if h.ID != "kiro-lint" || h.Event != "PostFileSave" || !h.IsRegex || h.Glob != `\.tsx$` {
		t.Errorf("v1 normalise wrong: %+v", h)
	}
	if h.ActionType != "shell" {
		t.Errorf("Kiro 'command' should alias to 'shell'; got %q", h.ActionType)
	}
}

func TestNormaliseHook_Agent(t *testing.T) {
	raw := rawHook{
		"id":     "review",
		"event":  "manual",
		"action": map[string]any{"type": "agent", "prompt": "review the diff"},
	}
	h, err := normaliseHook(raw)
	if err != nil {
		t.Fatal(err)
	}
	if h.ActionType != "agent" || h.Action != "review the diff" {
		t.Errorf("agent normalise wrong: %+v", h)
	}
}

func TestNormaliseHook_InvalidActionType(t *testing.T) {
	raw := rawHook{
		"id":     "x",
		"event":  "file.save",
		"action": map[string]any{"type": "wat", "command": "x"},
	}
	if _, err := normaliseHook(raw); err == nil {
		t.Fatal("invalid action type should error")
	}
}

func TestNormaliseHook_MissingID(t *testing.T) {
	raw := rawHook{"event": "file.save", "action": "x"}
	if _, err := normaliseHook(raw); err == nil {
		t.Fatal("missing id should error")
	}
}

func TestLoadAll_Flat(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/lint.json"), `{
  "id": "lint",
  "event": "file.save",
  "glob": "*.go",
  "action_type": "shell",
  "action": "true"
}`)
	hooks, err := reg.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 1 || hooks[0].ID != "lint" {
		t.Errorf("load flat: %v", hooks)
	}
}

func TestLoadAll_V1Envelope(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/two.json"), `{
  "version": "v1",
  "hooks": [
    {"name": "a", "trigger": "PostFileSave", "action": {"type": "command", "command": "echo a"}},
    {"name": "b", "trigger": "PreToolUse", "action": {"type": "agent", "prompt": "review"}}
  ]
}`)
	hooks, err := reg.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 2 {
		t.Errorf("expected 2; got %d", len(hooks))
	}
	if hooks[0].Event != "PostFileSave" || hooks[0].ActionType != "shell" {
		t.Errorf("hook[0] wrong: %+v", hooks[0])
	}
	if hooks[1].ActionType != "agent" {
		t.Errorf("hook[1] should be agent; got %q", hooks[1].ActionType)
	}
}

func TestLoadAll_TopLevelArray(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/multi.json"), `[
  {"id": "a", "event": "manual", "action_type": "shell", "action": "echo a"},
  {"id": "b", "event": "manual", "action_type": "shell", "action": "echo b"}
]`)
	hooks, err := reg.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 2 {
		t.Errorf("expected 2; got %d", len(hooks))
	}
}

func TestMatch_EventAndGlob(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/a.json"), `{"id":"go","event":"file.save","glob":"*.go","action_type":"shell","action":"true"}`)
	writeHook(t, filepath.Join(root, ".kiro/hooks/b.json"), `{"id":"md","event":"file.save","glob":"*.md","action_type":"shell","action":"true"}`)
	writeHook(t, filepath.Join(root, ".kiro/hooks/c.json"), `{"id":"all","event":"file.save","action_type":"shell","action":"true"}`)
	matched, err := reg.Match("file.save", "main.go")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, h := range matched {
		ids[h.ID] = true
	}
	if !ids["go"] || !ids["all"] || ids["md"] {
		t.Errorf("match for *.go: %v", matched)
	}
}

func TestMatch_DisabledSkipped(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/d.json"),
		`{"id":"x","event":"manual","action_type":"shell","action":"true","enabled":false}`)
	matched, _ := reg.Match("manual", "")
	if len(matched) != 0 {
		t.Errorf("disabled hook should be skipped; got %v", matched)
	}
}

func TestMatch_RegexMatcher(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/re.json"), `{
  "name": "r",
  "trigger": "PostFileSave",
  "matcher": "\\.tsx$",
  "action": {"type": "command", "command": "true"}
}`)
	matched, _ := reg.Match("PostFileSave", "components/Button.tsx")
	if len(matched) != 1 || matched[0].ID != "r" {
		t.Errorf("regex matcher: %v", matched)
	}
}

func TestAdd_WritesV1Envelope(t *testing.T) {
	reg, _ := setup(t)
	h := &models.Hook{
		ID:         "lint-on-save",
		Event:      "file.save",
		Glob:       "*.go",
		ActionType: "shell",
		Action:     "gofmt -w",
		Enabled:    true,
	}
	path, err := reg.Add(h)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"version": "v1"`) {
		t.Errorf("written file should be v1 envelope; got %s", data)
	}
	if !strings.Contains(string(data), `"trigger": "file.save"`) {
		t.Errorf("written file should have trigger; got %s", data)
	}
	if !strings.HasSuffix(path, "lint-on-save.json") {
		t.Errorf("path should end with id.json; got %s", path)
	}
	// Verify it round-trips.
	var env envelopeFile
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("written file should be valid envelope; %v", err)
	}
	if env.Version != "v1" {
		t.Errorf("version: %q", env.Version)
	}
}

func TestDispatch_ShellReceivesStdinJSON(t *testing.T) {
	reg, root := setup(t)
	tmp := filepath.Join(root, "captured.json")
	cmd := `cat > ` + tmp
	writeHook(t, filepath.Join(root, ".kiro/hooks/cat.json"),
		`{"id":"cat","event":"file.save","action_type":"shell","action":"`+cmd+`"}`)
	results, err := reg.Dispatch(context.Background(), "file.save", "main.go", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].OK {
		t.Fatalf("dispatch failed: %+v", results)
	}
	data, _ := os.ReadFile(tmp)
	if !strings.Contains(string(data), `"event"`) || !strings.Contains(string(data), `"file"`) {
		t.Errorf("STDIN JSON missing fields; got %s", data)
	}
	if !strings.Contains(string(data), "main.go") {
		t.Errorf("STDIN should include file path; got %s", data)
	}
}

func TestDispatch_AgentNoFnReturnsPlaceholder(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/a.json"),
		`{"id":"a","event":"manual","action":{"type":"agent","prompt":"review this"}}`)
	results, _ := reg.Dispatch(context.Background(), "manual", "", nil)
	if len(results) != 1 {
		t.Fatalf("expected 1; got %d", len(results))
	}
	r := results[0]
	if !r.OK {
		t.Errorf("agent without fn should report OK (placeholder); got error %s", r.Error)
	}
	if !strings.Contains(r.Output, "[agent hook] would execute prompt") {
		t.Errorf("placeholder output wrong: %s", r.Output)
	}
}

func TestDispatch_AgentWithFn(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/a.json"),
		`{"id":"a","event":"manual","action":{"type":"agent","prompt":"review this"}}`)
	fn := func(p string) (string, error) { return "model output: " + p, nil }
	results, _ := reg.Dispatch(context.Background(), "manual", "", fn)
	if len(results) != 1 || !results[0].OK {
		t.Fatalf("dispatch with fn failed: %+v", results)
	}
	if !strings.Contains(results[0].Output, "model output: review this") {
		t.Errorf("fn output not propagated: %s", results[0].Output)
	}
}

func TestDispatch_ShellFailureCaptured(t *testing.T) {
	reg, root := setup(t)
	writeHook(t, filepath.Join(root, ".kiro/hooks/fail.json"),
		`{"id":"fail","event":"manual","action_type":"shell","action":"exit 1"}`)
	results, _ := reg.Dispatch(context.Background(), "manual", "", nil)
	if len(results) != 1 || results[0].OK {
		t.Errorf("failed hook should be OK=false; got %+v", results)
	}
}

func TestGlobMatch(t *testing.T) {
	cases := []struct{ pattern, path string; want bool }{
		{"**/*.tsx", "components/Button.tsx", true},
		{"**/*.tsx", "components/Button.ts", false},
		{"*.go", "main.go", true},
		{"*.go", "sub/main.go", false},
		{"src/**/*.go", "src/a/b/c.go", true},
	}
	for _, c := range cases {
		if got := globMatch(c.pattern, c.path); got != c.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}