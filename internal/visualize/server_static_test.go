package visualize

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandleStatic_NotFound covers AC "missing file → 404 + JSON".
// Verified with a definitely-absent path; embed.FS.ReadFile returns
// fs.ErrNotExist which the handler maps to 404.
func TestHandleStatic_NotFound(t *testing.T) {
	s := NewServer(":0", stubWS{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/definitely-not-here-xyz.js", nil)
	s.handleStatic(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "asset not found") {
		t.Errorf("body = %q; want contains 'asset not found'", rec.Body.String())
	}
}

// TestHandleStatic_PathEscape covers AC "path-traversal attempt → 400".
// Both "../" and absolute-leading "/" must be rejected before reaching
// the embed.FS read so a hostile URL cannot read ../../etc/passwd out
// of the embed (which would be impossible anyway, but defence in depth).
func TestHandleStatic_PathEscape(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"dot dot", "/assets/..%2F..%2Fetc%2Fpasswd"},
		{"raw dot dot", "/assets/../../etc/passwd"},
		{"absolute leading slash", "/assets//etc/passwd"},
		{"embedded dot dot", "/assets/foo/../../bar.js"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewServer(":0", stubWS{}, nil)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			s.handleStatic(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d for %q; want 400", rec.Code, c.path)
			}
			if !strings.Contains(rec.Body.String(), "invalid path") {
				t.Errorf("body = %q; want contains 'invalid path'", rec.Body.String())
			}
		})
	}
}

// TestHandleStatic_NonGetMethod covers AC "non-GET → 405 + Allow".
// Only GET is permitted; POST/PUT/DELETE/PATCH are rejected.
func TestHandleStatic_NonGetMethod(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			s := NewServer(":0", stubWS{}, nil)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(m, "/assets/main.abc.js", nil)
			s.handleStatic(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d; want 405", rec.Code)
			}
			if got := rec.Header().Get("Allow"); got != http.MethodGet {
				t.Errorf("Allow header = %q; want 'GET'", got)
			}
		})
	}
}

// TestContentTypeForAsset covers the per-extension Content-Type table
// in handleStatic. Pulled out as a pure function so this test runs
// without any embed.FS fixture file present (dist/ may be empty
// before dashboard-frontend-foundation lands).
func TestContentTypeForAsset(t *testing.T) {
	cases := []struct {
		rel  string
		want string
	}{
		{"main.abc123.js", "application/javascript; charset=utf-8"},
		{"main.abc123.css", "text/css; charset=utf-8"},
		{"manifest.json", "application/json; charset=utf-8"},
		{"icon.svg", "image/svg+xml"},
		{"logo.png", "image/png"},
		{"favicon.ico", "application/octet-stream"},
		{"unknown", "application/octet-stream"},
		{"", "application/octet-stream"},
	}
	for _, c := range cases {
		t.Run(c.rel, func(t *testing.T) {
			if got := contentTypeForAsset(c.rel); got != c.want {
				t.Errorf("contentTypeForAsset(%q) = %q; want %q", c.rel, got, c.want)
			}
		})
	}
}

// TestHandleStatic_EmptyPath covers the boundary case where the URL
// is exactly "/assets/" with no path segment after the prefix.
// Should be rejected as 400 because the relative path is empty.
func TestHandleStatic_EmptyPath(t *testing.T) {
	s := NewServer(":0", stubWS{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/", nil)
	s.handleStatic(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want 400", rec.Code)
	}
}
