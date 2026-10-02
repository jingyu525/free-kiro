package visualize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// newTestServer builds a Server with a real workspace + engine rooted
// at the given tmpdir. Used by handler tests so we exercise the same
// Engine.TaskList / Status code paths production uses (Engine is a
// concrete type, not an interface — no fake-engine trick available).
func newTestServer(t *testing.T, root string) *Server {
	t.Helper()
	ws := workspace.New(root)
	eng := spec.New(ws)
	return NewServer(":0", &realWS{root: root}, eng)
}

// TestHandleSpecTasks_HappyPath covers /api/spec/<name>/tasks with a
// realistic tasks.md. Verifies wave grouping, done counts, and the
// "wave" field on each task item.
func TestHandleSpecTasks_HappyPath(t *testing.T) {
	root := setupSpecWithTasks(t, "demo",
		"- [ ] #1 alpha\n- [x] #2 beta [deps: #1]\n- [ ] #3 gamma [deps: #1]\n")
	s := newTestServer(t, root)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/spec/demo/tasks", nil)
	s.handleSpec(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	var resp tasksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Spec != "demo" {
		t.Errorf("spec = %q; want 'demo'", resp.Spec)
	}
	if resp.Summary.Total != 3 || resp.Summary.Done != 1 {
		t.Errorf("summary = %+v; want total=3 done=1", resp.Summary)
	}
	if len(resp.Waves) != 2 {
		t.Fatalf("len(waves) = %d; want 2", len(resp.Waves))
	}
	if resp.Waves[0].Index != 1 || len(resp.Waves[0].Tasks) != 1 {
		t.Errorf("wave 1 = %+v; want index=1 len=1", resp.Waves[0])
	}
	if resp.Waves[1].Index != 2 || len(resp.Waves[1].Tasks) != 2 {
		t.Errorf("wave 2 = %+v; want index=2 len=2", resp.Waves[1])
	}
	for _, wv := range resp.Waves {
		for _, ti := range wv.Tasks {
			if ti.Wave != wv.Index {
				t.Errorf("task #%d wave=%d but in wave index %d", ti.ID, ti.Wave, wv.Index)
			}
		}
	}
}

// TestHandleSpecTasks_NotFound covers the 404 path for a missing spec.
func TestHandleSpecTasks_NotFound(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/spec/ghost/tasks", nil)
	s.handleSpec(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "spec not found") {
		t.Errorf("body = %q; want contains 'spec not found'", rec.Body.String())
	}
}

// TestHandleSpecTasks_PathEscape covers the 400 path for traversal
// attempts. decodeSpecName rejects ".." anywhere in the name.
// Note: %2F in URL is decoded to "/" before reaching Go's http.ServeMux,
// so the test uses a literal ".." segment that's hard to escape but
// still trips the path-traversal guard.
func TestHandleSpecTasks_PathEscape(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	// "..foo" decodes to "..foo" which contains ".." → 400.
	req := httptest.NewRequest(http.MethodGet, "/api/spec/..foo/tasks", nil)
	s.handleSpec(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", rec.Code)
	}
}

// TestHandleSpecDrift_NotFound covers the 404 path. The empty-signals
// branch (unapproved spec) is hit when the spec exists but no
// baseline — covered by integration smoke (`free-kiro status --json`).
func TestHandleSpecDrift_NotFound(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/spec/ghost/drift", nil)
	s.handleSpec(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rec.Code)
	}
}

// TestHandleSpecTimeline_NotFound covers the 404 path.
func TestHandleSpecTimeline_NotFound(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/spec/ghost/timeline", nil)
	s.handleSpec(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rec.Code)
	}
}

// TestHandleSpecTimeline_SortByMtime covers the descending-mtime
// invariant: the most recently modified file must be first.
func TestHandleSpecTimeline_SortByMtime(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, ".kiro", "specs", "t")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(specDir, "old.md")
	newPath := filepath.Join(specDir, "new.md")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(oldPath, past, past); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/spec/t/timeline", nil)
	s.handleSpec(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	var resp timelineResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Files) != 2 {
		t.Fatalf("len(files) = %d; want 2", len(resp.Files))
	}
	if resp.Files[0].Path != "new.md" {
		t.Errorf("files[0].Path = %q; want 'new.md' (newest first)", resp.Files[0].Path)
	}
	if resp.Files[1].Path != "old.md" {
		t.Errorf("files[1].Path = %q; want 'old.md'", resp.Files[1].Path)
	}
}

// TestHandleHealth_Fields covers the response shape and that uptime
// is a non-negative integer.
func TestHandleHealth_Fields(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	s.handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	var resp healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q; want 'ok'", resp.Status)
	}
	if resp.UptimeSeconds < 0 {
		t.Errorf("uptime = %d; want >= 0", resp.UptimeSeconds)
	}
	if resp.Version == "" {
		t.Errorf("version = empty; want 'dev' or build version")
	}
	if resp.Subscribers != 0 {
		t.Errorf("subscribers = %d; want 0 (no SSE clients)", resp.Subscribers)
	}
	if resp.LastRefreshAt != "" {
		t.Errorf("last_refresh_at = %q; want empty before first refresh", resp.LastRefreshAt)
	}
}

// TestHandleHealth_AfterBroadcast covers that broadcast() updates
// lastRefreshAt so /api/health surfaces "last successful refresh".
func TestHandleHealth_AfterBroadcast(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	s.broadcast()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	s.handleHealth(rec, req)
	var resp healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.LastRefreshAt == "" {
		t.Errorf("last_refresh_at empty after broadcast; want RFC3339 timestamp")
	}
}

// TestHandleHooks_Disabled covers the registry_disabled fallback:
// listHooks returns nil → handler returns hooks:[] with
// registry_disabled=true (NOT a 503).
func TestHandleHooks_Disabled(t *testing.T) {
	root := t.TempDir()
	s := newTestServer(t, root)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/hooks", nil)
	s.handleHooks(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200 (graceful degrade)", rec.Code)
	}
	var resp hooksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.RegistryDisabled {
		t.Errorf("registry_disabled = false; want true (listHooks returns nil)")
	}
	if resp.Hooks == nil {
		t.Errorf("hooks = nil; want empty slice (so JSON is '[]' not 'null')")
	}
}

// TestHandleSpec_RouterDispatch covers that handleSpec correctly
// routes the three sub-paths and rejects unknown sub-paths with 404.
func TestHandleSpec_RouterDispatch(t *testing.T) {
	root := setupSpecWithTasks(t, "r", "- [ ] #1 only task\n")
	s := newTestServer(t, root)
	cases := []struct {
		sub  string
		want int
	}{
		{"tasks", http.StatusOK},
		{"drift", http.StatusNotFound}, // ghost spec, returns 404
		{"timeline", http.StatusOK},
		{"unknown", http.StatusNotFound},
	}
	for _, c := range cases {
		t.Run(c.sub, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/spec/r/"+c.sub, nil)
			s.handleSpec(rec, req)
			if rec.Code != c.want {
				t.Errorf("status = %d; want %d", rec.Code, c.want)
			}
		})
	}
}

// setupSpecWithTasks creates a workspace + spec with the given tasks
// content and returns the workspace root path.
func setupSpecWithTasks(t *testing.T, name, content string) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".kiro", "specs", name)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// realWS satisfies WorkspacePaths against a real workspace.Workspace
// rooted at `root`. Used by newTestServer; kept here (vs reusing
// stubWS in server_sse_test.go) so this file stands alone.
type realWS struct {
	root string
}

func (r *realWS) Root() string            { return r.root }
func (r *realWS) KiroDir() string         { return filepath.Join(r.root, ".kiro") }
func (r *realWS) SpecDir(n string) string { return filepath.Join(r.root, ".kiro", "specs", n) }
func (r *realWS) ReadCurrent() string     { return "" }
func (r *realWS) KiroDirExists() bool     { return true }
