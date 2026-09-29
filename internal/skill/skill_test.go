package skill

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseApp(t *testing.T) {
	cases := []struct {
		in   string
		want App
		err  bool
	}{
		{"all", "", false},
		{"", "", false},
		{"claude-code", AppClaudeCode, false},
		{"Claude-Code", AppClaudeCode, false},
		{"claude", AppClaudeCode, false},
		{"opencode", AppOpenCode, false},
		{"codex", AppCodex, false},
		{"codebuddy", AppCodeBuddy, false},
		{"unknown", "", true},
	}
	for _, c := range cases {
		got, err := ParseApp(c.in)
		if (err != nil) != c.err {
			t.Errorf("ParseApp(%q): err=%v, want err=%v", c.in, err, c.err)
		}
		if got != c.want {
			t.Errorf("ParseApp(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSkillsDir(t *testing.T) {
	home := "/home/test"
	cases := []struct {
		app App
		sub string
		want string
	}{
		{AppClaudeCode, "free-kiro", "/home/test/.claude/skills/free-kiro"},
		{AppOpenCode, "free-kiro", "/home/test/.opencode/skills/free-kiro"},
		{AppCodex, "free-kiro", "/home/test/.codex/skills/free-kiro"},
		{AppCodeBuddy, "free-kiro", "/home/test/.codebuddy/skills/free-kiro"},
		{AppClaudeCode, "", "/home/test/.claude/skills/claude-code"}, // empty subdir → app name
	}
	for _, c := range cases {
		got := SkillsDir(c.app, home, c.sub)
		if got != c.want {
			t.Errorf("SkillsDir(%q,%q,%q) = %q, want %q", c.app, home, c.sub, got, c.want)
		}
	}
}

func TestDetectInstalledApps(t *testing.T) {
	home := t.TempDir()
	// Create only ~/.claude
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := DetectInstalledApps(home)
	if len(got) != 1 || got[0] != AppClaudeCode {
		t.Errorf("DetectInstalledApps = %v, want [claude-code]", got)
	}
}

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()
	manifest := `{
		"name": "free-kiro",
		"version": "0.7.0",
		"free_kiro_min_version": "0.7.0",
		"apps": [{"id": "claude-code", "skills_dir": "~/.claude/skills"}],
		"files": [{"path": "SKILL.md", "required": true}],
		"sha256": {"SKILL.md": "abc123"}
	}`
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "free-kiro" || m.Version != "0.7.0" {
		t.Errorf("unexpected manifest: %+v", m)
	}
}

func TestLoadManifest_MissingFields(t *testing.T) {
	dir := t.TempDir()
	bad := `{"name": "free-kiro"}` // missing version
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(dir); err == nil {
		t.Error("expected error for missing version")
	}
}

func TestVerifyBundle_EmptySHA(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"name":"x","version":"0.1.0","sha256":{}}`
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := LoadManifest(dir)
	failures, err := m.VerifyBundle(dir)
	if err != nil {
		t.Fatalf("VerifyBundle errored: %v", err)
	}
	if failures != nil {
		t.Errorf("empty sha should pass; got %v", failures)
	}
}

func TestVerifyBundle_Mismatch(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"name":"x","version":"0.1.0","sha256":{"SKILL.md":"deadbeef"}}`
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := LoadManifest(dir)
	failures, err := m.VerifyBundle(dir)
	if err != nil {
		t.Fatalf("VerifyBundle errored: %v", err)
	}
	if len(failures) != 1 || !strings.Contains(failures[0], "SKILL.md") {
		t.Errorf("expected SKILL.md failure; got %v", failures)
	}
}

func TestVerifyBundle_OK(t *testing.T) {
	dir := t.TempDir()
	content := []byte("hello world")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	want := hashString(content)
	manifest := `{"name":"x","version":"0.1.0","sha256":{"SKILL.md":"` + want + `"}}`
	if err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := LoadManifest(dir)
	failures, _ := m.VerifyBundle(dir)
	if len(failures) != 0 {
		t.Errorf("expected no failures; got %v", failures)
	}
}

func TestInstallOne_LocalDir(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "skill.json"),
		[]byte(`{"name":"x","version":"1.2.3","sha256":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res := InstallOne(context.Background(), InstallOptions{
		App:    AppClaudeCode,
		Home:   home,
		Source: src,
	})
	if res.Err != nil {
		t.Fatalf("InstallOne error: %v", res.Err)
	}
	if res.Status != "installed" {
		t.Errorf("status=%q, want installed", res.Status)
	}
	// Verify the file landed.
	target := filepath.Join(home, ".claude", "skills", "free-kiro", "SKILL.md")
	if _, err := os.Stat(target); err != nil {
		t.Errorf("expected SKILL.md at %s; got %v", target, err)
	}
}

func TestInstallOne_AlreadyCurrent(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "skill.json"),
		[]byte(`{"name":"x","version":"1.2.3"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := InstallOptions{App: AppClaudeCode, Home: home, Source: src, Version: "1.2.3"}
	res := InstallOne(context.Background(), opts)
	if res.Err != nil {
		t.Fatalf("first install error: %v", res.Err)
	}
	// Re-run same version → already-current
	res = InstallOne(context.Background(), opts)
	if res.Status != "already-current" {
		t.Errorf("status=%q, want already-current", res.Status)
	}
}

func TestUninstallOne_Idempotent(t *testing.T) {
	home := t.TempDir()
	// uninstall when not installed → no-op, no error
	removed, err := UninstallOne(AppClaudeCode, home, "free-kiro")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Errorf("expected no files; got %v", removed)
	}
	// Now install then uninstall
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "skill.json"),
		[]byte(`{"name":"x","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills", "free-kiro"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Use a manual write to simulate prior install (since InstallOne requires same source format)
	target := filepath.Join(home, ".claude", "skills", "free-kiro")
	if err := os.WriteFile(filepath.Join(target, "skill.json"),
		[]byte(`{"name":"x","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	removed, err = UninstallOne(AppClaudeCode, home, "free-kiro")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 {
		t.Errorf("expected 1 dir removed; got %d", len(removed))
	}
}

func TestShowInstalled(t *testing.T) {
	home := t.TempDir()
	// pre-install one
	dir := filepath.Join(home, ".claude", "skills", "free-kiro")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skill.json"),
		[]byte(`{"name":"x","version":"2.0.0","free_kiro_min_version":"2.0.0"}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	states := ShowInstalled(home, "free-kiro")
	var claudeState *InstalledState
	for i, s := range states {
		if s.App == AppClaudeCode {
			claudeState = &states[i]
			break
		}
	}
	if claudeState == nil {
		t.Fatal("claude-code not in states")
	}
	if !claudeState.Installed {
		t.Error("claude-code should be installed")
	}
	if claudeState.Version != "2.0.0" {
		t.Errorf("version=%q, want 2.0.0", claudeState.Version)
	}
	if claudeState.Experimental {
		t.Error("claude-code should NOT be experimental")
	}
	// check codebuddy experimental flag
	for _, s := range states {
		if s.App == AppCodeBuddy && !s.Experimental {
			t.Error("codebuddy should be experimental")
		}
	}
}

func TestAssetURL(t *testing.T) {
	got := AssetURL("0.7.0")
	want := "https://github.com/jingyu525/free-kiro/releases/download/v0.7.0/free-kiro-skill_0.7.0.zip"
	if got != want {
		t.Errorf("AssetURL = %q, want %q", got, want)
	}
}

func TestSHA256SUMSURL(t *testing.T) {
	got := SHA256SUMSURL("v0.7.0")
	want := "https://github.com/jingyu525/free-kiro/releases/download/v0.7.0/free-kiro_0.7.0_SHA256SUMS"
	if got != want {
		t.Errorf("SHA256SUMSURL = %q, want %q", got, want)
	}
}

// hashString returns the hex sha256 of data. Replaces the previous
// temp-file-roundtrip helper — direct sha256.Sum256 is cleaner.
func hashString(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}