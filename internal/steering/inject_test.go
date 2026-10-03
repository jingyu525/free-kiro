package steering

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jingyu525/free-kiro/internal/workspace"
)

// writeFile is a tiny test helper that fails the test on IO error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// setupWorkspace creates a temp workspace with:
//   - .kiro/steering/{product,structure}.md (mode: always)
//   - All 5 IDE project-root instruction files, each containing the
//     marker region with a placeholder line and hand-written content
//     around the markers.
//
// Returns the workspace root path. Each test starts from a fresh dir.
func setupWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".kiro", "steering", "product.md"), "---\nmode: always\ndescription: product fixture\n---\n\n# Product\n\nfixture body for product\n")
	writeFile(t, filepath.Join(root, ".kiro", "steering", "structure.md"), "---\nmode: always\ndescription: structure fixture\n---\n\n# Structure\n\nfixture body for structure\n")
	for _, rel := range DefaultInjectTargets {
		// Prepend a `mode: manual` frontmatter so the steering store
		// does not classify these IDE instruction files as always-mode
		// docs (the default mode is "always" when frontmatter is
		// absent; see internal/steering/store.go parseDoc).
		writeFile(t, filepath.Join(root, rel), "---\nmode: manual\n---\n# free-kiro-managed:\n\nHand-written top section.\n\n"+InjectMarkerStart+"\nplaceholder\n"+InjectMarkerEnd+"\n\nHand-written tail section.\n")
	}
	return root
}

func TestStore_InjectAll_WritesBlockIntoMarkerRegion(t *testing.T) {
	root := setupWorkspace(t)
	ws := workspace.New(root)
	store := NewStore(ws, "")

	res := store.InjectAll("")
	if res.DocCount != 2 {
		t.Fatalf("DocCount = %d, want 2", res.DocCount)
	}
	if len(res.Written) != len(DefaultInjectTargets) {
		t.Fatalf("Written count = %d, want %d (targets %v)", len(res.Written), len(DefaultInjectTargets), res.Written)
	}
	if len(res.Skipped) != 0 {
		t.Fatalf("Skipped = %v, want empty", res.Skipped)
	}

	// Verify the on-disk CLAUDE.md content.
	data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	got := string(data)

	// Hand-written top + tail must be preserved verbatim.
	if !strings.Contains(got, "Hand-written top section.") {
		t.Errorf("missing hand-written top section in:\n%s", got)
	}
	if !strings.Contains(got, "Hand-written tail section.") {
		t.Errorf("missing hand-written tail section in:\n%s", got)
	}
	// Marker lines must be preserved verbatim.
	if !strings.Contains(got, InjectMarkerStart) {
		t.Errorf("missing start marker in:\n%s", got)
	}
	if !strings.Contains(got, InjectMarkerEnd) {
		t.Errorf("missing end marker in:\n%s", got)
	}
	// The injected block must include both always docs in alphabetical order.
	productIdx := strings.Index(got, "## product.md")
	structureIdx := strings.Index(got, "## structure.md")
	if productIdx == -1 || structureIdx == -1 {
		t.Fatalf("missing doc headers in:\n%s", got)
	}
	if productIdx > structureIdx {
		t.Errorf("docs not in alphabetical order: product at %d, structure at %d", productIdx, structureIdx)
	}
	// The auto-generated header line must appear inside the marker region.
	if !strings.Contains(got, InjectBlockHeader) {
		t.Errorf("missing auto-generated header in:\n%s", got)
	}
	// The original "placeholder" line inside the marker must be replaced.
	// Find the region between start and end markers; assert placeholder
	// is NOT in it.
	startLine := strings.Index(got, InjectMarkerStart)
	endLine := strings.Index(got, InjectMarkerEnd)
	if startLine == -1 || endLine == -1 || endLine <= startLine {
		t.Fatalf("marker region malformed in:\n%s", got)
	}
	region := got[startLine:endLine]
	if strings.Contains(region, "placeholder") {
		t.Errorf("placeholder line was not replaced inside marker region:\n%s", region)
	}
}

func TestStore_InjectAll_SkipsMissingMarker(t *testing.T) {
	root := setupWorkspace(t)
	// Overwrite AGENTS.md with a version that intentionally lacks the marker
	// (frontmatter `mode: manual` keeps the steering store from treating it
	// as an always-mode doc, isolating this test's intent).
	writeFile(t, filepath.Join(root, "AGENTS.md"), "---\nmode: manual\n---\n# free-kiro-managed:\n\nNo marker region here.\n")

	ws := workspace.New(root)
	store := NewStore(ws, "")

	res := store.InjectAll("")
	if res.DocCount != 2 {
		t.Fatalf("DocCount = %d, want 2", res.DocCount)
	}
	if len(res.Skipped) != 1 || res.Skipped[0].Path != "AGENTS.md" {
		t.Fatalf("Skipped = %+v, want one entry for AGENTS.md", res.Skipped)
	}
	if !strings.Contains(res.Skipped[0].Reason, "missing inject marker") {
		t.Errorf("Skip reason = %q, want substring 'missing inject marker'", res.Skipped[0].Reason)
	}
	// The other 4 targets must have been written.
	if len(res.Written) != len(DefaultInjectTargets)-1 {
		t.Errorf("Written count = %d, want %d", len(res.Written), len(DefaultInjectTargets)-1)
	}
	// AGENTS.md must remain byte-identical.
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(data), "No marker region here.") {
		t.Errorf("AGENTS.md was modified despite missing marker:\n%s", string(data))
	}
}

func TestStore_InjectAll_OnlyGlob(t *testing.T) {
	root := setupWorkspace(t)
	// Set up two targets that both have markers.
	writeFile(t, filepath.Join(root, "AGENTS.md"), "# free-kiro-managed:\n\n"+InjectMarkerStart+"\nold\n"+InjectMarkerEnd+"\n")

	ws := workspace.New(root)
	store := NewStore(ws, "")

	res := store.InjectAll("CLAUDE.md")
	if len(res.Written) != 1 || res.Written[0] != "CLAUDE.md" {
		t.Errorf("Written = %v, want [CLAUDE.md]", res.Written)
	}
	if len(res.Skipped) != 0 {
		t.Errorf("Skipped = %v, want empty", res.Skipped)
	}
	// AGENTS.md must NOT have been touched.
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(data), "\nold\n") {
		t.Errorf("AGENTS.md content unexpectedly changed:\n%s", string(data))
	}
}

func TestStore_InjectAll_NoAlwaysDocs(t *testing.T) {
	root := t.TempDir()
	// Only auto-mode docs — InjectAll must report DocCount=0.
	writeFile(t, filepath.Join(root, ".kiro", "steering", "tech.md"), "---\nmode: auto\ndescription: tech\n---\n\ntech body\n")

	ws := workspace.New(root)
	store := NewStore(ws, "")
	res := store.InjectAll("")
	if res.DocCount != 0 {
		t.Errorf("DocCount = %d, want 0", res.DocCount)
	}
	if res.Block != "" {
		t.Errorf("Block = %q, want empty", res.Block)
	}
}

func TestAssembleBlock_AlphabeticalOrder(t *testing.T) {
	// Pass docs in reverse order; assembleBlock must still emit
	// alphabetically (the caller pre-sorts via alwaysDocsSorted, but
	// assembleBlock is the source of truth for layout).
	docs := []struct {
		name, body string
	}{
		{"tech", "t-body"},
		{"product", "p-body"},
		{"structure", "s-body"},
	}
	// Build a slice matching models.SteeringDoc but without importing the
	// struct directly to keep this test focused.
	type stub struct {
		name, body string
	}
	_ = docs
	// Just exercise via the public API.
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".kiro", "steering", "tech.md"), "---\nmode: always\ndescription: t\n---\n\nt-body\n")
	writeFile(t, filepath.Join(root, ".kiro", "steering", "product.md"), "---\nmode: always\ndescription: p\n---\n\np-body\n")
	writeFile(t, filepath.Join(root, ".kiro", "steering", "structure.md"), "---\nmode: always\ndescription: s\n---\n\ns-body\n")

	ws := workspace.New(root)
	store := NewStore(ws, "")
	res := store.InjectAll("")
	if !strings.Contains(res.Block, "## product.md") ||
		!strings.Contains(res.Block, "## structure.md") ||
		!strings.Contains(res.Block, "## tech.md") {
		t.Fatalf("block missing one or more doc headers:\n%s", res.Block)
	}
	pi := strings.Index(res.Block, "## product.md")
	si := strings.Index(res.Block, "## structure.md")
	ti := strings.Index(res.Block, "## tech.md")
	if !(pi < si && si < ti) {
		t.Errorf("docs not alphabetical: product=%d structure=%d tech=%d", pi, si, ti)
	}
	// Sanity: stable ordering across two calls.
	res2 := store.InjectAll("")
	if res.Block != res2.Block {
		t.Errorf("assembleBlock is not deterministic")
	}
}

func TestFilterTargets(t *testing.T) {
	targets := []string{"CLAUDE.md", "AGENTS.md", ".cursorrules", ".cursor/rules/free-kiro.md", ".continue/rules/free-kiro.md"}
	cases := []struct {
		name   string
		glob   string
		wantN  int
	}{
		{"empty glob keeps all", "", 5},
		{"glob matches one", "CLAUDE.md", 1},
		{"glob matches none", "*.doesnotexist", 0},
		{"glob with ? matches .cursorrules", ".cur?orru*", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := filterTargets(targets, c.glob)
			if len(got) != c.wantN {
				t.Errorf("filterTargets(%q) returned %d items, want %d", c.glob, len(got), c.wantN)
			}
			// Sorted-input check: result must preserve input order.
			sort.SliceStable(got, func(i, j int) bool { return got[i] < got[j] })
			for i := 0; i < len(got); i++ {
				if got[i] != targets[i] && c.glob == "" {
					t.Errorf("empty glob must preserve order, got %v want %v", got, targets)
				}
			}
		})
	}
}

func TestIndexLineContaining(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		substr string
		want   int
	}{
		{"empty", "", "x", -1},
		{"first line", "alpha\nbeta\ngamma", "alpha", 0},
		{"middle line", "alpha\nbeta\ngamma", "beta", 1},
		{"last line", "alpha\nbeta\ngamma", "gamma", 2},
		{"not found", "alpha\nbeta", "gamma", -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := indexLineContaining(c.text, c.substr); got != c.want {
				t.Errorf("indexLineContaining = %d, want %d", got, c.want)
			}
		})
	}
}

// TestInject_IgnoresMarkerInProse is the regression guard for spec
// `.kiro/specs/fix-inject-marker-match/`. The init template's
// explanation paragraph quotes the marker literals inline (e.g.
// `本文件末尾由 <!-- free-kiro-managed:start --> / ... <!-- free-kiro-managed:end --> marker 包裹的 markdown 块`).
// A naive substring match on these lines would treat them as the
// real marker region and write the always-mode block between them
// — corrupting the prose. After the fix, only whole-line matches
// (after strings.TrimSpace) are recognized as the marker region.
func TestInject_IgnoresMarkerInProse(t *testing.T) {
	root := t.TempDir()
	// Always-mode steering docs (product + structure), no tech — keeps
	// the assertion focused on "did the block land in the right slot"
	// rather than on per-doc content.
	writeFile(t, filepath.Join(root, ".kiro", "steering", "product.md"),
		"---\nmode: always\ndescription: product fixture\n---\n\n# Product\n\nPROD_BODY\n")
	writeFile(t, filepath.Join(root, ".kiro", "steering", "structure.md"),
		"---\nmode: always\ndescription: structure fixture\n---\n\n# Structure\n\nSTRUCT_BODY\n")

	// Target file mimics init's CLAUDE.md layout: hand-written
	// preamble, an explanation paragraph that quotes the marker
	// literals inline, a coding-standards section, then the real
	// standalone marker block at the end.
	const prose1 = "本文件末尾由 `<!-- free-kiro-managed:start -->` /"
	const prose2 = "`<!-- free-kiro-managed:end -->` marker 包裹的 markdown 块，由"
	target := strings.Join([]string{
		"# free-kiro-managed:",
		"",
		"Hand-written preamble line 1.",
		"Hand-written preamble line 2.",
		"",
		"## 项目上下文（自动注入的 steering）",
		"",
		prose1,
		prose2 + " `free-kiro steering inject` 在每次 `init` / `inject` 运行时自动生成，",
		"内容来自 `.kiro/steering/*.md` 中 `mode: always` 的文档。",
		"",
		"## 编码规范（SessionStart 必须先 Read）",
		"",
		"CODING_STANDARDS_SECTION",
		"",
		InjectMarkerStart,
		"", // empty marker block — inject fills this
		InjectMarkerEnd,
		"",
	}, "\n")
	writeFile(t, filepath.Join(root, "CLAUDE.md"), target)

	ws := workspace.New(root)
	store := NewStore(ws, "")
	res := store.InjectAll("")
	if len(res.Written) != 1 {
		t.Fatalf("expected 1 written file, got %d (skipped: %+v)", len(res.Written), res.Skipped)
	}

	got, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	gotStr := string(got)

	// AC-6: prose paragraph MUST remain verbatim (not overwritten by
	// the always-mode block). If the bug regresses, "PROD_BODY" /
	// "STRUCT_BODY" would appear in this paragraph.
	if !strings.Contains(gotStr, prose1) {
		t.Errorf("prose1 marker-quote line was lost — inject wrote into the explanation paragraph")
	}
	if !strings.Contains(gotStr, prose2) {
		t.Errorf("prose2 marker-quote line was lost")
	}
	if strings.Contains(gotStr, "PROD_BODY") && !strings.Contains(gotStr, "## product.md") {
		t.Errorf("PROD_BODY leaked into prose without a heading — inject wrote into prose")
	}

	// AC-4: real marker block MUST have been filled with the always
	// docs in alphabetical order (product → structure).
	if !strings.Contains(gotStr, "## product.md") {
		t.Errorf("real marker block missing ## product.md heading")
	}
	if !strings.Contains(gotStr, "## structure.md") {
		t.Errorf("real marker block missing ## structure.md heading")
	}
	if !strings.Contains(gotStr, "PROD_BODY") {
		t.Errorf("real marker block missing PROD_BODY content")
	}
	if !strings.Contains(gotStr, "STRUCT_BODY") {
		t.Errorf("real marker block missing STRUCT_BODY content")
	}

	// Pre-amble + coding-standards section also preserved verbatim.
	if !strings.Contains(gotStr, "Hand-written preamble line 1.") {
		t.Errorf("preamble was modified")
	}
	if !strings.Contains(gotStr, "CODING_STANDARDS_SECTION") {
		t.Errorf("coding-standards section was modified")
	}

	// The standalone marker lines themselves are preserved (inject
	// replaces the block BETWEEN them, not the lines themselves).
	startCount := strings.Count(gotStr, InjectMarkerStart)
	endCount := strings.Count(gotStr, InjectMarkerEnd)
	if startCount != 2 {
		t.Errorf("expected 2 occurrences of start marker (1 prose quote + 1 standalone), got %d", startCount)
	}
	if endCount != 2 {
		t.Errorf("expected 2 occurrences of end marker (1 prose quote + 1 standalone), got %d", endCount)
	}
}
