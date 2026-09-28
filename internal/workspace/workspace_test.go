package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// withTempDir creates a temp directory, chdirs into it, and returns the
// absolute path. Cleanup restores cwd and removes the temp tree.
func withTempDir(t *testing.T) string {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return dir
}

func TestFind_NoKiro(t *testing.T) {
	dir := withTempDir(t)
	ws := Find(dir)
	if ws.Root() != dir {
		t.Fatalf("Find without .kiro should fall back to start dir; got %q want %q", ws.Root(), dir)
	}
}

func TestFind_AtRoot(t *testing.T) {
	dir := withTempDir(t)
	if err := os.MkdirAll(filepath.Join(dir, KiroDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := Find(dir)
	if ws.Root() != dir {
		t.Fatalf("Find with .kiro at root should return root; got %q", ws.Root())
	}
}

func TestFind_WalksUp(t *testing.T) {
	dir := withTempDir(t)
	if err := os.MkdirAll(filepath.Join(dir, KiroDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	child := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}
	ws := Find(child)
	if ws.Root() != dir {
		t.Fatalf("Find from nested child should walk up; got %q want %q", ws.Root(), dir)
	}
}

func TestEnsureLayout_CreatesDirsAndSettings(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	for _, p := range []string{ws.KiroDir(), ws.SpecsDir(), ws.SteeringDir(), ws.HooksDir()} {
		if !isDir(p) {
			t.Errorf("expected dir %s to exist", p)
		}
	}
	if !fileExists(ws.SettingsPath()) {
		t.Fatalf("settings.json should exist")
	}
	// Verify settings.json is valid JSON with the expected defaults.
	data, err := os.ReadFile(ws.SettingsPath())
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	if got["generator"] != "template" {
		t.Errorf("default generator should be 'template'; got %v", got["generator"])
	}
}

func TestEnsureLayout_Idempotent(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("first EnsureLayout: %v", err)
	}
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("second EnsureLayout should not error: %v", err)
	}
}

func TestEnsureLayout_PreservesExistingSettings(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	custom := map[string]any{"generator": "custom", "user_set": true}
	if err := ws.SaveSettings(custom); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("re-EnsureLayout: %v", err)
	}
	got, err := ws.LoadSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got["generator"] != "custom" || got["user_set"] != true {
		t.Errorf("EnsureLayout should preserve existing settings; got %v", got)
	}
}

func TestRequire_Missing(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if _, err := ws.Require(); err == nil {
		t.Fatal("Require should error when .kiro does not exist")
	}
}

func TestRequire_Present(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if _, err := ws.Require(); err != nil {
		t.Fatalf("Require should succeed after EnsureLayout: %v", err)
	}
}

func TestLoadSettings_MissingFileReturnsEmpty(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	if err := ws.EnsureLayout(); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := os.Remove(ws.SettingsPath()); err != nil {
		t.Fatalf("remove: %v", err)
	}
	got, err := ws.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings on missing file should not error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected empty map; got %v", got)
	}
}

func TestSpecDir(t *testing.T) {
	dir := withTempDir(t)
	ws := New(dir)
	got := ws.SpecDir("user-auth")
	want := filepath.Join(ws.SpecsDir(), "user-auth")
	if got != want {
		t.Errorf("SpecDir(%q) = %q, want %q", "user-auth", got, want)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}