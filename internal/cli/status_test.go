package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jingyu525/free-kiro/internal/visualize"
)

// setupTestKiro builds a tmpdir with a minimal .kiro/ workspace containing
// specs/<name>/.meta.json for each name, plus an optional .current
// pointer. Returns the tmpdir; caller is responsible for chdir.
func setupTestKiro(t *testing.T, specs []string, currentName string) string {
	t.Helper()
	dir := t.TempDir()
	kiro := filepath.Join(dir, ".kiro")
	if err := os.MkdirAll(filepath.Join(kiro, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range specs {
		specDir := filepath.Join(kiro, "specs", name)
		if err := os.MkdirAll(specDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Minimal but well-formed .meta.json — applyDefaults() in
		// models.LoadSpecMeta fills in any missing defaults, but we
		// pass the full canonical set to keep this fixture hermetic.
		meta := `{"name":"` + name + `","phase":"tasks","workflow":"requirements-first",` +
			`"spec_type":"feature","approved":false,"quick":false,"generator":"template",` +
			`"prompt":"test","created_at":"2026-01-01T00:00:00Z",` +
			`"updated_at":"2026-01-01T00:00:00Z","baseline":{}}`
		if err := os.WriteFile(filepath.Join(specDir, ".meta.json"), []byte(meta), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if currentName != "" {
		if err := os.WriteFile(filepath.Join(kiro, ".current"), []byte(currentName), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestStatusHumanOutput verifies the default (human-readable) rendering:
// table header, all spec names present, active marker on the .current spec.
func TestStatusHumanOutput(t *testing.T) {
	dir := setupTestKiro(t, []string{"foo", "bar"}, "foo")
	t.Chdir(dir)

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status --human: %v\nstderr: %s", err, errBuf.String())
	}
	s := out.String()
	for _, want := range []string{"NAME", "PHASE", "APPROVED", "DRIFT", "TASKS", "foo", "bar"} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, s)
		}
	}
	// Active marker: the `foo` row should have `*` in the first column
	// (tabwriter aligns columns but keeps the leading character).
	lines := strings.Split(strings.TrimSpace(s), "\n")
	foundActive := false
	for _, ln := range lines {
		trimmed := strings.TrimLeft(ln, " \t")
		if strings.HasPrefix(trimmed, "*") && strings.Contains(ln, "foo") {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Errorf("expected '*' active marker on foo row\n--- got ---\n%s", s)
	}
}

// TestStatusJSONOutput verifies --json returns a ProjectReport JSON with
// correct active field and spec count.
func TestStatusJSONOutput(t *testing.T) {
	dir := setupTestKiro(t, []string{"foo", "bar"}, "foo")
	t.Chdir(dir)

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status --json: %v\nstderr: %s", err, errBuf.String())
	}
	var rep visualize.ProjectReport
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatalf("JSON parse: %v\n--- got ---\n%s", err, out.String())
	}
	if rep.Active != "foo" {
		t.Errorf("active = %q, want foo", rep.Active)
	}
	if len(rep.Specs) != 2 {
		t.Errorf("specs = %d, want 2", len(rep.Specs))
	}
	// Schema compatibility sanity: generated_at + per-spec meta.
	if rep.GeneratedAt.IsZero() {
		t.Error("generated_at should be non-zero")
	}
	for _, s := range rep.Specs {
		if s.Meta == nil || s.Meta.Name == "" {
			t.Errorf("spec missing meta.name: %+v", s)
		}
	}
}

// TestStatusNoWorkspace verifies the .kiro/ missing case yields a
// user-facing error that hints at `free-kiro init`.
func TestStatusNoWorkspace(t *testing.T) {
	dir := t.TempDir() // empty — no .kiro/
	t.Chdir(dir)

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing workspace")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error should hint at init: %v", err)
	}
}

// TestStatusNoSpecs verifies the empty-specs case prints "No specs found"
// and exits 0.
func TestStatusNoSpecs(t *testing.T) {
	dir := setupTestKiro(t, nil, "")
	t.Chdir(dir)

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(out.String(), "No specs found") {
		t.Errorf("output should contain 'No specs found', got: %q", out.String())
	}
}

// TestStatusPointerMissing verifies that a .current pointing to a
// non-existent spec surfaces as a footer note, not an error.
func TestStatusPointerMissing(t *testing.T) {
	dir := setupTestKiro(t, []string{"foo"}, "ghost")
	t.Chdir(dir)

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(out.String(), "(current=ghost not found)") {
		t.Errorf("output should flag dangling .current pointer, got: %q", out.String())
	}
}

// TestStatusIsReadOnly verifies that running `status` does not modify any
// spec metadata file. We compare .meta.json mtime before/after with a
// short sleep to expose any unintentional write.
func TestStatusIsReadOnly(t *testing.T) {
	dir := setupTestKiro(t, []string{"foo"}, "foo")
	t.Chdir(dir)

	metaPath := filepath.Join(dir, ".kiro", "specs", "foo", ".meta.json")
	infoBefore, err := os.Stat(metaPath)
	if err != nil {
		t.Fatal(err)
	}

	cmd := statusCmdFactory()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{})
	if execErr := cmd.Execute(); execErr != nil {
		t.Fatalf("status: %v", execErr)
	}

	// Sleep so any spurious write would shift it past equal().
	time.Sleep(10 * time.Millisecond)
	infoAfter, err := os.Stat(metaPath)
	if err != nil {
		t.Fatal(err)
	}
	if !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Errorf("status modified .meta.json: mtime before=%v after=%v",
			infoBefore.ModTime(), infoAfter.ModTime())
	}
}
