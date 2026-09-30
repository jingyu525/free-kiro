package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jingyu525/free-kiro/internal/ide"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// checkPath reports whether ~/.local/bin is on PATH.
func checkPath() *doctorIssue {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	target := filepath.Join(home, ".local", "bin")
	pathEnv := os.Getenv("PATH")
	for _, p := range filepath.SplitList(pathEnv) {
		if p == target {
			return nil
		}
	}
	return &doctorIssue{
		Severity: "warn",
		Title:    target + " is not on PATH",
		Detail:   "free-kiro installed there won't be found by `free-kiro …` invocations",
		Fix:      "add to PATH:  export PATH=\"$HOME/.local/bin:$PATH\"",
	}
}

// selfPath returns the path to the running binary (best-effort).
func selfPath() string {
	p, err := os.Executable()
	if err != nil {
		return "(unknown)"
	}
	return p
}

// countFreeKiroHooks reads an IDE's settings.json and counts free-kiro
// hooks (identified by the `# free-kiro-managed:` command marker).
func countFreeKiroHooks(path string) (bool, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, 0, nil
		}
		return false, 0, err
	}
	var s struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return false, 0, err
	}
	const marker = "# free-kiro-managed:"
	count := 0
	for _, entries := range s.Hooks {
		for _, e := range entries {
			if len(e.Hooks) > 0 && strings.HasPrefix(e.Hooks[0].Command, marker) {
				count++
			}
		}
	}
	return true, count, nil
}

// checkIDEInstructions verifies that every project-root instruction file
// free-kiro should have written for the given IDE id actually exists and
// carries the free-kiro-managed marker. Returns nil when every file is
// present and marked; otherwise returns a single doctorIssue describing
// the first failure (read error takes priority over missing/unmarked).
//
// `cwd` is the doctor's cwd; the actual workspace root is resolved via
// `workspace.Find` so the check matches what `init --ide` would write.
func checkIDEInstructions(cwd string, id ide.ID) *doctorIssue {
	rels := ide.InstructionFiles(id)
	if len(rels) == 0 {
		return nil
	}
	root := workspace.Find(cwd).Root()
	for _, rel := range rels {
		abs := filepath.Join(root, rel)
		isFK, readErr := ide.IsFreeKiroInstruction(abs)
		if readErr != nil {
			return &doctorIssue{
				Severity: "warn",
				Title:    string(id) + " instruction file unreadable: " + rel,
				Detail:   readErr.Error(),
				Fix:      "re-run `free-kiro init --ide " + string(id) + " --overwrite-instructions`",
			}
		}
		if !isFK {
			return &doctorIssue{
				Severity: "info",
				Title:    string(id) + " instruction file not yet written: " + rel,
				Detail:   abs,
				Fix:      "run `free-kiro init --ide " + string(id) + "`",
			}
		}
	}
	return nil
}

// fetchLatestVersion queries the GitHub API for the latest release tag
// (no auth). Returns "" on any error (network, parse, etc.).
func fetchLatestVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "https://api.github.com/repos/" + GitHubRepo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return ""
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	return strings.TrimPrefix(body.TagName, "v")
}

// getwd is a thin wrapper so tests can stub it without touching every
// doctor call-site. Currently delegates to os.Getwd.
func getwd() (string, error) {
	return os.Getwd()
}

// osGetwd is referenced from doctor_report.go to keep the import cycle
// docs explicit; replaced with the direct call when doctor_report.go is
// inlined.
var osGetwd = os.Getwd
