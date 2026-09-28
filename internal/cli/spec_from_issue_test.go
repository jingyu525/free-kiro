package cli

import (
	"strings"
	"testing"
)

func TestParseGitHubIssueURL_FullURL(t *testing.T) {
	cases := []struct {
		in        string
		wantOwner string
		wantRepo  string
		wantNum   int
		wantErr   bool
	}{
		{"https://github.com/jingyu525/free-kiro/issues/42", "jingyu525", "free-kiro", 42, false},
		{"https://github.com/jingyu525/free-kiro/issues/42#issuecomment-1", "jingyu525", "free-kiro", 42, false},
		{"https://www.github.com/jingyu525/free-kiro/issues/7", "jingyu525", "free-kiro", 7, false},
		{"https://github.com/jingyu525/free-kiro/issues/1/anything", "jingyu525", "free-kiro", 1, false},
		{"https://github.com/jingyu525/free-kiro/pull/1", "", "", 0, true},      // PR is not an issue
		{"https://gitlab.com/foo/bar/issues/1", "", "", 0, true},              // non-github host
		{"not a url at all", "", "", 0, true},
		{"https://github.com/jingyu525/free-kiro/issues/abc", "", "", 0, true}, // non-numeric
	}
	for _, c := range cases {
		got, err := ParseGitHubIssueURL(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseGitHubIssueURL(%q): err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if got.Owner != c.wantOwner || got.Repo != c.wantRepo || got.Number != c.wantNum {
			t.Errorf("ParseGitHubIssueURL(%q) = %+v, want {%s/%s #%d}",
				c.in, got, c.wantOwner, c.wantRepo, c.wantNum)
		}
	}
}

func TestParseGitHubIssueURL_Shorthand(t *testing.T) {
	got, err := ParseGitHubIssueURL("jingyu525/free-kiro#42")
	if err != nil {
		t.Fatal(err)
	}
	if got.Owner != "jingyu525" || got.Repo != "free-kiro" || got.Number != 42 {
		t.Errorf("shorthand wrong: %+v", got)
	}
}

func TestSlugFromText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Add login form", "add-login-form"},
		{"Fix: broken redirect / 404 page", "fix-broken-redirect-404-page"},
		{"!!!", "spec-from-issue"},
		{"a-b_c.d:e/f", "a-b-c-d-e-f"},
		{"   trim   spaces   ", "trim-spaces"},
	}
	for _, c := range cases {
		got := slugFromText(c.in)
		if got != c.want {
			t.Errorf("slugFromText(%q) = %q, want %q", c.in, got, c.want)
		}
		// Always non-empty, kebab-case (only [a-z0-9-]), ≤ 40 chars.
		if got == "" {
			t.Errorf("slugFromText(%q) returned empty", c.in)
		}
		if len(got) > 40 {
			t.Errorf("slugFromText(%q) = %q (len=%d, > 40)", c.in, got, len(got))
		}
		for _, r := range got {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
				t.Errorf("slugFromText(%q) = %q contains %q (not kebab)", c.in, got, r)
			}
		}
	}
}

func TestSlugFromText_TooLongIsTruncated(t *testing.T) {
	long := strings.Repeat("abcdefghij", 10) // 100 chars
	got := slugFromText(long)
	if len(got) > 40 {
		t.Errorf("slug length %d > 40", len(got))
	}
}