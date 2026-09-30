package ide

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want ID
		err  bool
	}{
		{"", "", false},     // empty → auto-detect
		{"auto", "", false}, // explicit auto
		{"claude-code", ClaudeCode, false},
		{"Claude-Code", ClaudeCode, false}, // case-insensitive
		{"claude", ClaudeCode, false},      // alias
		{"codebuddy", CodeBuddy, false},
		{"none", "", false}, // treated as auto-detect signal (caller resolves)
		{"unknown", "", true},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if (err != nil) != c.err {
			t.Errorf("Parse(%q): err=%v, want err=%v", c.in, err, c.err)
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectAll_MarksExistingDirs(t *testing.T) {
	home := t.TempDir()
	// Create ~/.claude/ but not ~/.codebuddy/
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := DetectAll(home)
	var claude, codebuddy bool
	for _, d := range got {
		if d.ID == ClaudeCode {
			claude = d.DirExists
		}
		if d.ID == CodeBuddy {
			codebuddy = d.DirExists
		}
	}
	if !claude {
		t.Error("claude-code should be detected")
	}
	if codebuddy {
		t.Error("codebuddy should NOT be detected")
	}
}

func TestInstallHooks_CreatesFile(t *testing.T) {
	home := t.TempDir()
	path, note, err := InstallHooks(ClaudeCode, home)
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("path should be set")
	}
	if !strings.Contains(note, path) {
		t.Errorf("note should reference path; got %q", note)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("settings.json should exist: %v", err)
	}
}

func TestInstallHooks_PreservesExistingUserHooks(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	// Claude Code's actual schema: event-keyed map.
	existing := `{
  "model": "sonnet",
  "hooks": {
    "PostToolUse": [
      {"matcher": "Write|Edit", "hooks": [{"type": "command", "command": "bash ~/.claude/scripts/auto-format.sh"}]}
    ]
  }
}`
	if err := os.WriteFile(cfg, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := InstallHooks(ClaudeCode, home); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var s settingsShape
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	// user hook preserved.
	var userHook bool
	for _, e := range s.Hooks["PostToolUse"] {
		if len(e.Hooks) > 0 && e.Hooks[0].Command == "bash ~/.claude/scripts/auto-format.sh" {
			userHook = true
		}
	}
	if !userHook {
		t.Error("user hook should be preserved")
	}
	// free-kiro PreToolUse present.
	var lintGate bool
	for _, e := range s.Hooks["PreToolUse"] {
		if isFreeKiroEntry(e) {
			lintGate = true
		}
	}
	if !lintGate {
		t.Error("free-kiro lint-gate should be installed")
	}
	// model field preserved.
	if !strings.Contains(string(data), `"model": "sonnet"`) {
		t.Errorf("model field should be preserved; got %s", data)
	}
}

func TestInstallHooks_Idempotent(t *testing.T) {
	home := t.TempDir()
	for i := 0; i < 3; i++ {
		if _, _, err := InstallHooks(ClaudeCode, home); err != nil {
			t.Fatal(err)
		}
	}
	cfg := filepath.Join(home, ".claude", "settings.json")
	data, _ := os.ReadFile(cfg)
	var s settingsShape
	_ = json.Unmarshal(data, &s)
	count := 0
	for _, entries := range s.Hooks {
		for _, e := range entries {
			if isFreeKiroEntry(e) {
				count++
			}
		}
	}
	if count != 2 {
		t.Errorf("expected exactly 2 free-kiro entries; got %d", count)
	}
}

func TestInstallHooks_RefusesMalformedJSON(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte("{ this is not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := InstallHooks(ClaudeCode, home); err == nil {
		t.Fatal("expected error when settings.json is malformed")
	}
}

func TestInstallHooks_UpdatesOldFreeKiroEntries(t *testing.T) {
	// First install writes "lint-gate" command.
	home := t.TempDir()
	if _, _, err := InstallHooks(ClaudeCode, home); err != nil {
		t.Fatal(err)
	}
	// Manually replace the marker (simulating an older free-kiro version).
	cfg := filepath.Join(home, ".claude", "settings.json")
	data, _ := os.ReadFile(cfg)
	updated := strings.ReplaceAll(string(data),
		freeKiroMarker+" lint-gate: free-kiro lint || exit 2",
		"# free-kiro-managed: lint-gate: NEW COMMAND")
	if err := os.WriteFile(cfg, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	// Re-install.
	if _, _, err := InstallHooks(ClaudeCode, home); err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(cfg)
	if !strings.Contains(string(data2), "free-kiro lint || exit 2") {
		t.Errorf("re-install should restore canonical command; got %s", data2)
	}
	if strings.Contains(string(data2), "NEW COMMAND") {
		t.Errorf("old free-kiro command should be removed; got %s", data2)
	}
}

func TestWriteAgentsMD_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".kiro"), 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := WriteAgentsMD(dir, "zh", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "AGENTS.md") {
		t.Errorf("path should end with AGENTS.md; got %q", path)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "free-kiro") {
		t.Errorf("content should mention free-kiro; got %q", string(data))
	}
}

func TestWriteAgentsMD_EnglishVariant(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".kiro"), 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := WriteAgentsMD(dir, "en", false)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	body := string(data)
	if !strings.Contains(body, "AI coding agents") {
		t.Errorf("English template should contain 'AI coding agents'; got %q", body)
	}
}

func TestWriteAgentsMD_Idempotent(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".kiro"), 0o755)
	first, _ := WriteAgentsMD(dir, "zh", false)
	second, _ := WriteAgentsMD(dir, "zh", false)
	if first != second {
		t.Errorf("expected same path on repeat")
	}
}

func TestUpsertHookEntry_StripsOldOnes(t *testing.T) {
	marker := freeKiroMarker + " old"
	existing := []HookSpec{
		{Hooks: []HookIn{{Type: "command", Command: marker}}},      // free-kiro old
		{Hooks: []HookIn{{Type: "command", Command: "user hook"}}}, // user-managed
	}
	fk := FreeKiroHookSpec{Event: "PreToolUse", Matcher: "Edit|Write", Command: "new"}
	out := upsertHookEntry(existing, fk)
	var sawOld, sawUser, sawNew int
	for _, e := range out {
		if len(e.Hooks) == 0 {
			continue
		}
		switch e.Hooks[0].Command {
		case marker:
			sawOld++
		case "user hook":
			sawUser++
		case "new":
			sawNew++
		}
	}
	if sawOld != 0 || sawUser != 1 || sawNew != 1 {
		t.Errorf("sawOld=%d sawUser=%d sawNew=%d", sawOld, sawUser, sawNew)
	}
}
