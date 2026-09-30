package steering

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jingyu525/free-kiro/internal/workspace"
)

// writeSteering writes a steering file with frontmatter + body.
func writeSteering(t *testing.T, path, fm, body string) {
	t.Helper()
	full := "---\n" + fm + "---\n" + body
	if err := os.WriteFile(path, []byte(full), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeRaw writes a raw markdown file (no frontmatter).
func writeRaw(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestParseFrontmatter_Basic(t *testing.T) {
	text := "---\nmode: always\ndescription: project overview\n---\n# Title\n\nbody text"
	meta, body := ParseFrontmatter(text)
	if meta["mode"] != "always" {
		t.Errorf("mode: %q", meta["mode"])
	}
	if meta["description"] != "project overview" {
		t.Errorf("description: %q", meta["description"])
	}
	if body == "" || body[:8] != "# Title\n" {
		t.Errorf("body: %q", body)
	}
}

func TestParseFrontmatter_StripsQuotes(t *testing.T) {
	text := "---\ndescription: \"with: colon\"\nmode: 'auto'\n---\nbody"
	meta, _ := ParseFrontmatter(text)
	if meta["description"] != "with: colon" {
		t.Errorf("description: %q", meta["description"])
	}
	if meta["mode"] != "auto" {
		t.Errorf("mode: %q", meta["mode"])
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	text := "# Just a heading\nbody"
	meta, body := ParseFrontmatter(text)
	if len(meta) != 0 {
		t.Errorf("expected empty meta; got %v", meta)
	}
	if body != text {
		t.Errorf("body should round-trip: %q", body)
	}
}

func TestParseFrontmatter_Unclosed(t *testing.T) {
	text := "---\nmode: auto\nbody without closing fence"
	meta, body := ParseFrontmatter(text)
	if len(meta) != 0 {
		t.Errorf("unclosed should yield empty meta; got %v", meta)
	}
	if body != text {
		t.Errorf("unclosed should yield full text as body")
	}
}

func TestParseFrontmatter_InclusionKey(t *testing.T) {
	text := "---\ninclusion: filematch\nfileMatchPattern: \"**/*.tsx\"\ndescription: react components\n---\nbody"
	meta, _ := ParseFrontmatter(text)
	if meta["inclusion"] != "filematch" {
		t.Errorf("inclusion: %q", meta["inclusion"])
	}
	if meta["fileMatchPattern"] != "**/*.tsx" {
		t.Errorf("fileMatchPattern: %q", meta["fileMatchPattern"])
	}
}

func TestParseFilePatterns_Single(t *testing.T) {
	got := ParseFilePatterns(`"**/*.tsx"`)
	if len(got) != 1 || got[0] != "**/*.tsx" {
		t.Errorf("single pattern: %v", got)
	}
}

func TestParseFilePatterns_List(t *testing.T) {
	got := ParseFilePatterns(`["**/*.ts", "**/*.tsx"]`)
	if len(got) != 2 || got[0] != "**/*.ts" || got[1] != "**/*.tsx" {
		t.Errorf("list patterns: %v", got)
	}
}

func TestParseFilePatterns_Empty(t *testing.T) {
	if got := ParseFilePatterns(""); got != nil {
		t.Errorf("empty should be nil; got %v", got)
	}
}

func TestGlobMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"**/*.tsx", "components/Button.tsx", true},
		{"**/*.tsx", "components/Button.ts", false},
		{"**/*.tsx", "src/a/b/c.tsx", true},
		{"src/**/*.go", "src/a/b/c.go", true},
		{"src/**/*.go", "tests/a.go", false},
		{"*.md", "README.md", true},
		{"*.md", "docs/README.md", false},
		{"docs/*", "docs/index.md", true},
		{"docs/*", "docs/sub/page.md", false},
		{"docs/**", "docs/sub/page.md", true},
		{"?single.md", "asingle.md", true},
		{"?single.md", "single.md", false},
	}
	for _, c := range cases {
		if got := GlobMatch(c.pattern, c.path); got != c.want {
			t.Errorf("GlobMatch(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

// --- store + assemble tests against real workspace ---

func TestStore_LoadAllWorkspaceAndGlobal(t *testing.T) {
	// Fake HOME so globalDir is isolated.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	globalDir := filepath.Join(tmp, ".kiro", "steering")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(globalDir, "global-only.md"),
		"mode: always\ndescription: global overview\n", "global body")

	// Workspace with a same-named doc (should override) and a unique doc.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "global-only.md"),
		"mode: always\ndescription: workspace override\n", "workspace body")
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "ws-only.md"),
		"mode: auto\ndescription: ws only\n", "ws body")

	ws := workspace.New(root)
	store := NewStore(ws, globalDir)
	docs := store.LoadAll()

	byName := map[string]string{}
	for _, d := range docs {
		byName[d.Name] = d.Content
	}
	if got := byName["global-only"]; got != "workspace body" {
		t.Errorf("workspace should override global; got %q", got)
	}
	if got := byName["ws-only"]; got != "ws body" {
		t.Errorf("ws-only should load; got %q", got)
	}
}

func TestStore_AGENTS_MD(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	globalDir := filepath.Join(tmp, ".kiro", "steering")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Workspace AGENTS.md at root.
	writeRaw(t, filepath.Join(root, "AGENTS.md"), "my agent instructions\n")

	ws := workspace.New(root)
	store := NewStore(ws, globalDir)
	docs := store.LoadAll()
	var agents *struct{ Content string }
	for _, d := range docs {
		if d.Name == "AGENTS" {
			agents = &struct{ Content string }{d.Content}
		}
	}
	if agents == nil {
		t.Fatal("AGENTS doc not loaded")
	}
	if agents.Content != "my agent instructions" {
		t.Errorf("AGENTS content: %q", agents.Content)
	}
}

func TestStore_InvalidModeSkipped(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	globalDir := filepath.Join(tmp, ".kiro", "steering")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "bad.md"),
		"mode: nonsense\n", "body")
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "good.md"),
		"mode: always\n", "good body")

	ws := workspace.New(root)
	store := NewStore(ws, globalDir)
	docs := store.LoadAll()
	names := map[string]bool{}
	for _, d := range docs {
		names[d.Name] = true
	}
	if names["bad"] {
		t.Error("malformed doc should be skipped, not loaded")
	}
	if !names["good"] {
		t.Error("good doc should be loaded")
	}
}

func TestStore_Get(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "product.md"),
		"mode: always\n", "product body")

	ws := workspace.New(root)
	store := NewStore(ws, filepath.Join(tmp, ".kiro", "steering"))
	d := store.Get("product")
	if d == nil {
		t.Fatal("Get returned nil")
	}
	if d.Content != "product body" {
		t.Errorf("content: %q", d.Content)
	}
	if d.Scope != "workspace" {
		t.Errorf("scope: %q", d.Scope)
	}
}

func TestAssemble_AlwaysIncluded(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "always.md"),
		"mode: always\n", "always body")

	ws := workspace.New(root)
	store := NewStore(ws, filepath.Join(tmp, ".kiro", "steering"))
	ctx := store.Assemble("anything", nil, "")
	if ctx == "" {
		t.Fatal("always doc should be in context")
	}
	if !contains(ctx, "always body") {
		t.Errorf("context should contain body; got %q", ctx)
	}
}

func TestAssemble_AutoKeywordMatch(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "api.md"),
		"mode: auto\ndescription: REST API design patterns\n", "api body")

	ws := workspace.New(root)
	store := NewStore(ws, filepath.Join(tmp, ".kiro", "steering"))
	// Prompt with overlapping keyword ("api").
	if ctx := store.Assemble("design api endpoint", nil, ""); !contains(ctx, "api body") {
		t.Errorf("matching prompt should pull auto doc; got %q", ctx)
	}
	// Prompt without overlap.
	if ctx := store.Assemble("unrelated topic", nil, ""); contains(ctx, "api body") {
		t.Errorf("non-matching prompt should not pull auto doc; got %q", ctx)
	}
}

func TestAssemble_ManualOnly(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "manual.md"),
		"mode: manual\n", "manual body")

	ws := workspace.New(root)
	store := NewStore(ws, filepath.Join(tmp, ".kiro", "steering"))
	if ctx := store.Assemble("anything", nil, ""); contains(ctx, "manual body") {
		t.Errorf("manual should not auto-include; got %q", ctx)
	}
	if ctx := store.Assemble("anything", []string{"manual"}, ""); !contains(ctx, "manual body") {
		t.Errorf("manual should appear when named; got %q", ctx)
	}
}

func TestAssemble_FileMatch(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".kiro", "steering"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSteering(t, filepath.Join(root, ".kiro", "steering", "react.md"),
		"mode: filematch\nfileMatchPattern: \"**/*.tsx\"\ndescription: react rules\n", "react body")

	ws := workspace.New(root)
	store := NewStore(ws, filepath.Join(tmp, ".kiro", "steering"))
	if ctx := store.Assemble("", nil, "components/Button.tsx"); !contains(ctx, "react body") {
		t.Errorf("matching file should pull doc; got %q", ctx)
	}
	if ctx := store.Assemble("", nil, "main.go"); contains(ctx, "react body") {
		t.Errorf("non-matching file should not pull doc; got %q", ctx)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
