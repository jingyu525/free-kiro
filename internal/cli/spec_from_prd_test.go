package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFetchPRD_HTML(t *testing.T) {
	// Tiny HTML page with the standard noise a real PRD has: scripts,
	// styles, nav, footer. Verifies our walker drops them.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>Login PRD</title>
  <style>body { color: red; }</style>
</head>
<body>
  <nav>Home | About</nav>
  <h1>User Login</h1>
  <script>console.log('ad')</script>
  <p>We need users to log in. Goals:</p>
  <ul>
    <li>Email + password</li>
    <li>2FA optional</li>
  </ul>
  <footer>© 2026 Acme</footer>
</body>
</html>`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	title, body, err := FetchPRD(ctx, server.URL+"/prds/login.html")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Login PRD" {
		t.Errorf("title: got %q, want %q", title, "Login PRD")
	}
	// Body must contain the actual content.
	for _, want := range []string{"User Login", "Email + password", "2FA optional", "Goals"} {
		if !contains(body, want) {
			t.Errorf("body missing %q; got:\n%s", want, body)
		}
	}
	// Body must NOT contain noise.
	for _, notWant := range []string{"color: red", "console.log", "Home | About", "© 2026"} {
		if contains(body, notWant) {
			t.Errorf("body should not contain %q; got:\n%s", notWant, body)
		}
	}
}

func TestFetchPRD_Markdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/markdown")
		_, _ = w.Write([]byte("# Onboarding Flow\n\nWe need a 3-step wizard.\n\n- Step 1\n- Step 2\n- Step 3\n"))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	title, body, err := FetchPRD(ctx, server.URL+"/onboarding.md")
	if err != nil {
		t.Fatal(err)
	}
	if title == "" {
		t.Errorf("title should be derived from filename when no <title> tag")
	}
	if !contains(body, "3-step wizard") {
		t.Errorf("markdown body should be preserved; got %q", body)
	}
}

func TestFetchPRD_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := FetchPRD(ctx, server.URL)
	if err == nil {
		t.Fatal("404 should error")
	}
}

func TestFetchPRD_BadScheme(t *testing.T) {
	ctx := context.Background()
	_, _, err := FetchPRD(ctx, "ftp://example.com")
	if err == nil {
		t.Fatal("ftp:// should error")
	}
}

func TestFetchPRD_TruncatesLargeBodies(t *testing.T) {
	// Build a huge page; verify truncation kicks in.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		body := "<html><body><p>" + strings.Repeat("word ", 20000) + "</p></body></html>"
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, body, err := FetchPRD(ctx, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(body, "truncated") {
		t.Errorf("expected truncation marker in body; got %s…", body[:min(200, len(body))])
	}
}

func TestExtractTitleFromFilename(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/prds/user-login.html", "user login"},
		{"/docs/onboarding-v2.md", "onboarding v2"},
		{"/", "example.com"},
	}
	for _, c := range cases {
		if got := extractTitleFromFilename(mustParse(t, "https://example.com"+c.path)); got != c.want {
			t.Errorf("extractTitleFromFilename(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestLooksLikeHTML(t *testing.T) {
	if !looksLikeHTML([]byte("<!DOCTYPE html><html>")) {
		t.Error("doctype should match")
	}
	if !looksLikeHTML([]byte("<html><body>")) {
		t.Error("html should match")
	}
	if looksLikeHTML([]byte("# Markdown\n\nHello")) {
		t.Error("markdown should not match")
	}
}

func TestCollapseWhitespace(t *testing.T) {
	got := collapseWhitespace("  hello   world  \n\n  foo  \n  bar  ")
	want := "hello world\nfoo\nbar"
	if got != want {
		t.Errorf("collapseWhitespace: got %q, want %q", got, want)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func mustParse(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
