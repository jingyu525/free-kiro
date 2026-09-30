package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchBrowserHTML_BskMissing(t *testing.T) {
	// Point to a definitely-missing binary.
	t.Setenv("FREE_KIRO_BSK_BIN", "/nonexistent/bsk-stub-binary")
	_, _, err := FetchBrowserHTML(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error for missing bsk")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bsk") || !strings.Contains(msg, "browser-skill") {
		t.Errorf("error should mention bsk and browser-skill; got: %s", msg)
	}
	if !strings.Contains(msg, "--from-prd") {
		t.Errorf("error should suggest --from-prd as fallback; got: %s", msg)
	}
}

func TestFetchBrowserHTML_Success(t *testing.T) {
	// Use the stub bsk in testdata/.
	stubPath, err := filepath.Abs("../../testdata/bsk-stub/bsk")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FREE_KIRO_BSK_BIN", stubPath)

	title, body, err := FetchBrowserHTML(context.Background(), "https://example.com/page")
	if err != nil {
		t.Fatalf("FetchBrowserHTML error: %v", err)
	}
	if title != "Browser PRD Title" {
		t.Errorf("title=%q, want %q", title, "Browser PRD Title")
	}
	if !strings.Contains(body, "Browser Rendered Spec") {
		t.Errorf("body should contain 'Browser Rendered Spec'; got: %s", body)
	}
	if strings.Contains(body, "skip me") {
		t.Errorf("body should drop <nav> content; got: %s", body)
	}
	if strings.Contains(body, "ignored()") {
		t.Errorf("body should drop <script> content; got: %s", body)
	}
}

func TestFetchBrowserHTML_NavigateFailure(t *testing.T) {
	stubPath, err := filepath.Abs("../../testdata/bsk-stub/bsk")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FREE_KIRO_BSK_BIN", stubPath)
	// Non-https URL makes the stub's `navigate` exit 1.
	_, _, err = FetchBrowserHTML(context.Background(), "http://insecure.example.com")
	if err == nil {
		t.Fatal("expected error for failed navigate")
	}
	if !strings.Contains(err.Error(), "navigate") {
		t.Errorf("error should mention navigate; got: %s", err.Error())
	}
}

func TestResolvePromptAndNameEx_MutuallyExclusive(t *testing.T) {
	_, _, _, err := resolvePromptAndNameEx(context.Background(),
		"", "https://github.com/foo/bar/issues/1", "https://example.com", "",
		nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually exclusive error; got %v", err)
	}
	_, _, _, err = resolvePromptAndNameEx(context.Background(),
		"", "", "https://example.com", "https://example.com", nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually exclusive error for prd+browser; got %v", err)
	}
	_, _, _, err = resolvePromptAndNameEx(context.Background(),
		"", "https://github.com/foo/bar/issues/1", "", "https://example.com", nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually exclusive error for issue+browser; got %v", err)
	}
}

func TestResolvePromptAndNameEx_BrowserSource(t *testing.T) {
	// Stub bsk so FetchBrowserHTML succeeds.
	stubPath, err := filepath.Abs("../../testdata/bsk-stub/bsk")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FREE_KIRO_BSK_BIN", stubPath)
	p, name, source, err := resolvePromptAndNameEx(context.Background(),
		"", "", "", "https://example.com/page", []string{"my-spec"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if source != "from-browser" {
		t.Errorf("source=%q, want from-browser", source)
	}
	if name != "my-spec" {
		t.Errorf("name=%q, want my-spec", name)
	}
	if !strings.Contains(p, "# Source (browser-rendered):") {
		t.Errorf("prompt should have browser-rendered source prefix; got: %s", p)
	}
	if !strings.Contains(p, "Browser PRD Title") {
		t.Errorf("prompt should contain title; got: %s", p)
	}
}
