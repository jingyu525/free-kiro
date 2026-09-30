package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupDemoRoot builds a tmpdir shaped like the free-kiro repo root:
// contains `examples/todo-app/` (the bare minimum demo checks). The
// caller chdir's into dir before invoking the demo command.
func setupDemoRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "examples", "todo-app"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runDemoReset runs demoCmd.RunE with fresh flag state. We use this in
// every test instead of demoCmd.Execute() because cobra's per-command
// SetOut/SetErr/SetArgs leaks through the global cmd tree when tests run
// in sequence. Resetting flags guarantees each test starts clean.
func runDemoReset(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	demoCmd.ResetFlags()
	demoCmd.Flags().Bool("no-color", false, "禁止 ANSI 颜色输出（适合 CI / tee）")
	demoCmd.Flags().String("ide", "none", "demo 完后要展示的 IDE hook 片段")
	demoCmd.SilenceErrors = false
	demoCmd.SilenceUsage = false
	if err := demoCmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags(%v): %v", args, err)
	}
	var out, errBuf bytes.Buffer
	demoCmd.SetOut(&out)
	demoCmd.SetErr(&errBuf)
	err := demoCmd.RunE(demoCmd, nil)
	return out.String(), errBuf.String(), err
}

func TestDemo_HappyPath(t *testing.T) {
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	out, _, err := runDemoReset(t)
	if err != nil {
		t.Fatalf("demo: %v", err)
	}
	for _, want := range []string{"5 分钟端到端 demo", "examples/todo-app", "serve", "watch --preset reactive"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, out)
		}
	}

	// Marker file written + well-formed RFC3339.
	data, err := os.ReadFile(filepath.Join(dir, ".kiro", ".demostart"))
	if err != nil {
		t.Fatalf("marker should exist: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, string(data)); err != nil {
		t.Errorf("marker not RFC3339: %q (%v)", string(data), err)
	}
}

func TestDemo_WrongCwd(t *testing.T) {
	// Empty tmpdir — no examples/todo-app → UsageError (exit 3).
	dir := t.TempDir()
	t.Chdir(dir)

	_, _, err := runDemoReset(t)
	if err == nil {
		t.Fatal("expected error when examples/todo-app is missing")
	}
	// Per requirements AC, message must point the user at `cd` to repo root.
	if !strings.Contains(err.Error(), "examples/todo-app") {
		t.Errorf("error must mention examples/todo-app; got %v", err)
	}
}

func TestDemo_DedupeRecentMarker(t *testing.T) {
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	// Pre-write a marker that's 5 seconds old — within the 30-second window.
	markerPath := filepath.Join(dir, ".kiro", ".demostart")
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o755); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UTC().Add(-5 * time.Second).Format(time.RFC3339)
	if err := os.WriteFile(markerPath, []byte(stamp), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, err := runDemoReset(t)
	if err != nil {
		t.Fatalf("demo: %v", err)
	}
	// Must warn instead of overwriting.
	if !strings.Contains(out, "已有最近的 demo marker") {
		t.Errorf("expected recent-marker warning; got %q", out)
	}
	// File must NOT be rewritten with a fresh timestamp.
	got, _ := os.ReadFile(markerPath)
	if string(got) != stamp {
		t.Errorf("marker rewritten; before=%q after=%q", stamp, string(got))
	}
}

func TestDemo_DedupeOldMarkerOverwrites(t *testing.T) {
	// Marker > 30 seconds old → should be silently overwritten.
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	markerPath := filepath.Join(dir, ".kiro", ".demostart")
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339)
	if err := os.WriteFile(markerPath, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := runDemoReset(t)
	if err != nil {
		t.Fatalf("demo: %v", err)
	}
	got, _ := os.ReadFile(markerPath)
	if string(got) == stale {
		t.Error("stale marker should have been overwritten with fresh timestamp")
	}
}

func TestDemo_IDESnippet_ClaudeCode(t *testing.T) {
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	out, _, err := runDemoReset(t, "--ide", "claude-code")
	if err != nil {
		t.Fatalf("demo --ide claude-code: %v", err)
	}
	for _, want := range []string{"claude-code", "PreToolUse", "free-kiro-managed", "free-kiro lint || exit 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("snippet missing %q\n--- got ---\n%s", want, out)
		}
	}
}

func TestDemo_NoColor_ZeroANSI(t *testing.T) {
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	out, _, err := runDemoReset(t, "--no-color")
	if err != nil {
		t.Fatalf("demo --no-color: %v", err)
	}
	// ESC = 0x1B. With --no-color the output must contain zero.
	if bytes.ContainsRune([]byte(out), 0x1B) {
		t.Errorf("--no-color output should contain 0 ANSI; got: %q", out)
	}
}

func TestDemo_UnknownIDE(t *testing.T) {
	dir := setupDemoRoot(t)
	t.Chdir(dir)

	_, _, err := runDemoReset(t, "--ide", "vscode")
	if err == nil {
		t.Fatal("expected UsageError for unknown IDE")
	}
	// Parse returns UsageError mentioning supported IDs.
	if !strings.Contains(err.Error(), "claude-code") {
		t.Errorf("error must mention supported IDEs; got %v", err)
	}
}
