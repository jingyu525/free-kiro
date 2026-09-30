package cli

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// GitHubIssueURL holds the parsed components of a github.com/<owner>/<repo>/issues/<n>
// URL. Path components only — query / fragment are ignored.
type GitHubIssueURL struct {
	Owner string
	Repo  string
	Number int
}

// ParseGitHubIssueURL accepts common GitHub issue URLs and returns the
// parsed components. Accepted shapes:
//
//	https://github.com/<owner>/<repo>/issues/<n>
//	https://github.com/<owner>/<repo>/issues/<n>#issuecomment-...
//	git@github.com:<owner>/<repo>/issues/<n>.git  (rare; just in case)
//	<owner>/<repo>#<n>                            (shorthand)
//
// Returns an error (UsageError) when the input doesn't look like a
// GitHub issue reference.
func ParseGitHubIssueURL(s string) (GitHubIssueURL, error) {
	s = strings.TrimSpace(s)
	// shorthand: owner/repo#123
	if !strings.Contains(s, "://") && !strings.Contains(s, "@") && strings.Contains(s, "#") {
		parts := strings.SplitN(s, "#", 2)
		if len(parts) == 2 {
			ownerRepo := strings.SplitN(parts[0], "/", 2)
			if len(ownerRepo) == 2 {
				n, err := atoiSafe(parts[1])
				if err != nil {
					return GitHubIssueURL{}, ferrors.New("spec.from-issue",
						"shorthand needs <owner>/<repo>#<number>; got "+s)
				}
				return GitHubIssueURL{Owner: ownerRepo[0], Repo: ownerRepo[1], Number: n}, nil
			}
		}
		return GitHubIssueURL{}, ferrors.New("spec.from-issue",
			"shorthand needs <owner>/<repo>#<number>; got "+s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return GitHubIssueURL{}, ferrors.Wrap("spec.from-issue", err, "parse URL")
	}
	host := strings.ToLower(u.Host)
	if host != "github.com" && host != "www.github.com" {
		return GitHubIssueURL{}, ferrors.New("spec.from-issue",
			"only github.com URLs are supported; got "+host)
	}
	segs := strings.Split(strings.Trim(u.Path, "/"), "/")
	// Expect: <owner>/<repo>/issues/<n>[...]
	if len(segs) < 4 || segs[2] != "issues" {
		return GitHubIssueURL{}, ferrors.New("spec.from-issue",
			"URL must end with /issues/<number>; got "+u.Path)
	}
	n, err := atoiSafe(segs[3])
	if err != nil {
		return GitHubIssueURL{}, ferrors.New("spec.from-issue",
			"issue number must be a positive integer; got "+segs[3])
	}
	return GitHubIssueURL{Owner: segs[0], Repo: segs[1], Number: n}, nil
}

// FetchIssueTitleAndBody shells out to `gh issue view` to fetch the
// title + body of the named issue. Requires `gh` to be authenticated
// (the user already has it per the project bootstrap). Returns a
// ready-to-use string suitable for the spec prompt.
func FetchIssueTitleAndBody(ref GitHubIssueURL) (string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return "", ferrors.New("spec.from-issue",
			"`gh` CLI not found in PATH — install it from https://cli.github.com")
	}
	number := fmt.Sprintf("%d", ref.Number)
	repo := ref.Owner + "/" + ref.Repo
	// gh issue view <number> --repo <owner/repo> --json title,body -q '.title + "\n\n" + .body'
	out, err := exec.Command("gh", "issue", "view", number,
		"--repo", repo, "--json", "title,body", "-q", ".title + \"\\n\\n\" + .body").Output()
	if err != nil {
		// Fall back to plain text if --json fails (older gh versions
		// may not support it).
		out, err = exec.Command("gh", "issue", "view", number, "--repo", repo).Output()
		if err != nil {
			return "", ferrors.Wrap("spec.from-issue", err,
				"gh issue view failed — is gh authenticated for "+repo+"?")
		}
	}
	body := strings.TrimSpace(string(out))
	if body == "" {
		return "", ferrors.New("spec.from-issue",
			"gh returned empty body for "+repo+" issue "+number)
	}
	return body, nil
}

// atoiSafe parses a non-negative integer. Returns -1 + UsageError on
// failure so `free-kiro spec new --from-issue <bad>` exits with code 3.
func atoiSafe(s string) (int, error) {
	n := 0
	if s == "" {
		return -1, ferrors.NewUsage("spec.from-issue.atoi", "empty number")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return -1, ferrors.NewUsage("spec.from-issue.atoi", fmt.Sprintf("not a number: %q", s))
		}
		n = n*10 + int(r-'0')
	}
	if n <= 0 {
		return -1, ferrors.NewUsage("spec.from-issue.atoi", "issue number must be positive")
	}
	return n, nil
}

// slugFromText derives a short spec name (kebab-case, ≤40 chars) from
// arbitrary issue title text. Falls back to "spec-from-issue" when the
// input has no usable alphanumerics.
func slugFromText(s string) string {
	out := make([]rune, 0, len(s))
	lastDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' || r == ':':
			if !lastDash && len(out) > 0 {
				out = append(out, '-')
				lastDash = true
			}
		}
		// drop other chars
	}
	slug := strings.TrimRight(string(out), "-")
	if len(slug) > 40 {
		slug = slug[:40]
		slug = strings.TrimRight(slug, "-")
	}
	if slug == "" {
		return "spec-from-issue"
	}
	return slug
}