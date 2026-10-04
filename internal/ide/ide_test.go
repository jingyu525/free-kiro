package ide

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
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
		{"cursor", Cursor, false},
		{"Cursor", Cursor, false}, // case-insensitive
		{"continue", Continue, false},
		{"opencode", OpenCode, false},
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

func TestParse_UnknownReturnsUsageError(t *testing.T) {
	// Per .kiro/steering/coding-style.md 8.6 (zero-exemption), unknown IDE values
	// must surface a UsageError so callers can map to exit code 3.
	_, err := Parse("vscode")
	if err == nil {
		t.Fatal("expected error for unknown IDE")
	}
	var target *ferrors.UsageError
	if !errors.As(err, &target) {
		t.Fatalf("expected *UsageError; got %T (%v)", err, err)
	}
	// Message should list every supported IDE so users can self-correct.
	msg := err.Error()
	for _, name := range []string{"claude-code", "codebuddy", "cursor", "continue", "opencode"} {
		if !strings.Contains(msg, name) {
			t.Errorf("UsageError message should mention %q; got %q", name, msg)
		}
	}
}

func TestAll_ReturnsFive(t *testing.T) {
	got := All()
	if len(got) != 5 {
		t.Fatalf("All() should return 5 IDEs; got %d (%v)", len(got), got)
	}
	want := map[ID]bool{ClaudeCode: false, CodeBuddy: false, Cursor: false, Continue: false, OpenCode: false}
	for _, id := range got {
		want[id] = true
	}
	for id, seen := range want {
		if !seen {
			t.Errorf("All() missing %q", id)
		}
	}
}

func TestConfigPathFor_AllFive(t *testing.T) {
	home := "/tmp/home"
	cases := map[ID]string{
		ClaudeCode: filepath.Join(home, ".claude", "settings.json"),
		CodeBuddy:  filepath.Join(home, ".codebuddy", "settings.json"),
		Cursor:     filepath.Join(home, ".cursor", "settings.json"),
		Continue:   filepath.Join(home, ".continue", "config.json"),
		OpenCode:   filepath.Join(home, ".opencode", "settings.json"),
	}
	for id, want := range cases {
		got := configPathFor(id, home)
		if got != want {
			t.Errorf("configPathFor(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestDetectAll_MarksExistingDirs(t *testing.T) {
	home := t.TempDir()
	// Create ~/.claude/ and ~/.cursor/ but not the other three.
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := DetectAll(home)
	want := map[ID]bool{
		ClaudeCode: true, CodeBuddy: false, Cursor: true,
		Continue: false, OpenCode: false,
	}
	got2 := map[ID]bool{}
	for _, d := range got {
		got2[d.ID] = d.DirExists
	}
	for id, w := range want {
		if got2[id] != w {
			t.Errorf("DetectAll(%s) = %v, want %v", id, got2[id], w)
		}
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
	for range 3 {
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

// TestInstructionTemplate_ExplainsBothMarkers is a regression guard for
// docs-template-marker-explanation. It asserts each of the four embedded
// IDE instruction templates (instructions_zh / instructions_en / agents_zh
// / agents_en) names both the top-of-file `# free-kiro-managed:` marker
// (injected by `prependMarker`) and the tail-of-file
// `<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`
// markers (injected by `steering inject`), and credits each marker to its
// injection source. A failure here means a future template edit has
// dropped one half of the explanation and is likely to confuse readers
// into thinking the top marker is a bug.
func TestInstructionTemplate_ExplainsBothMarkers(t *testing.T) {
	templates := []string{
		"templates/instructions_zh.md",
		"templates/instructions_en.md",
		"templates/agents_zh.md",
		"templates/agents_en.md",
	}
	for _, p := range templates {
		body, err := templatesFS.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		assertBothMarkersExplained(t, p, string(body))
	}
}

// assertBothMarkersExplained performs the per-template assertions called
// out by docs-template-marker-explanation AC-4 / AC-5: presence of the
// three marker literals plus explicit attribution of each marker to its
// injection source (`prependMarker` for the top marker, `steering inject`
// for the tail block). Chinese templates may phrase the top marker
// source as `init 注入` or `运行时注入` instead of `prependMarker`;
// accept any of the three.
func assertBothMarkersExplained(t *testing.T, path, body string) {
	t.Helper()

	// AC-4: both marker string families must appear literally.
	mustContain := []string{
		"# free-kiro-managed:",
		"<!-- free-kiro-managed:start -->",
		"<!-- free-kiro-managed:end -->",
	}
	for _, s := range mustContain {
		if !strings.Contains(body, s) {
			t.Errorf("%s missing marker string %q", path, s)
		}
	}

	// AC-5: the top marker's injection source must be named explicitly.
	// Accept either the Go function name or an equivalent Chinese phrase.
	hasTopSource := strings.Contains(body, "prependMarker") ||
		strings.Contains(body, "init 注入") ||
		strings.Contains(body, "运行时注入")
	if !hasTopSource {
		t.Errorf("%s missing prependMarker / init 注入 / 运行时注入", path)
	}

	// AC-1: the tail block's injection source must be named explicitly.
	if !strings.Contains(body, "steering inject") {
		t.Errorf("%s missing steering inject", path)
	}
}

func TestInstallHooks_CursorCreatesFile(t *testing.T) {
	// Per docs/HOOKS.md, Cursor settings live at ~/.cursor/settings.json.
	// Verify free-kiro hook installer creates the directory + file with
	// the canonical event-keyed envelope (same as Claude Code / CodeBuddy).
	home := t.TempDir()
	path, note, err := InstallHooks(Cursor, home)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".cursor", "settings.json")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if !strings.Contains(note, want) {
		t.Errorf("note should reference %q; got %q", want, note)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("settings.json should exist: %v", err)
	}
	var sh settingsShape
	if err := json.Unmarshal(data, &sh); err != nil {
		t.Fatalf("settings.json not valid JSON: %v\nraw: %s", err, data)
	}
	// free-kiro-managed marker present on PreToolUse lint-gate.
	var lintGate bool
	for _, e := range sh.Hooks["PreToolUse"] {
		if isFreeKiroEntry(e) {
			lintGate = true
		}
	}
	if !lintGate {
		t.Error("free-kiro lint-gate should be installed for Cursor")
	}
}

func TestInstallHooks_FiveIDEsRoundTrip(t *testing.T) {
	// Smoke-test that InstallHooks succeeds for all 5 IDs, each writing
	// to a distinct config path. This guards against future edits that
	// accidentally only handle a subset of All().
	home := t.TempDir()
	ids := All()
	if len(ids) < 5 {
		t.Fatalf("sanity: All() returned %d IDs (< 5); test premise broken", len(ids))
	}
	paths := map[ID]string{}
	for _, id := range ids {
		path, _, err := InstallHooks(id, home)
		if err != nil {
			t.Fatalf("InstallHooks(%s): %v", id, err)
		}
		paths[id] = path
		if _, err := os.Stat(path); err != nil {
			t.Errorf("settings file for %s not written: %v", id, err)
		}
	}
	// All 5 paths must be distinct (sanity: configPathFor is exhaustive).
	seen := map[string]ID{}
	for id, p := range paths {
		if other, dup := seen[p]; dup {
			t.Errorf("path collision between %s and %s at %s", id, other, p)
		}
		seen[p] = id
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
