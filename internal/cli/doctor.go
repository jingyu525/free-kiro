package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/ide"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

const GitHubRepo = "jingyu525/free-kiro"

// doctorCmd runs a series of self-checks and prints a human-readable
// report. Exits 0 when everything is healthy, 1 when there are warnings
// that should be addressed, 2 when something is fatally broken (e.g.
// the binary can't be located). Exit codes match the lint convention
// so IDE hooks can route them consistently.
//
// By default doctor only reports on IDEs that are actually installed
// (no warnings for Claude Code / CodeBuddy directories that simply
// don't exist on this machine). Use --verbose to see every supported
// IDE regardless of installation status.
func doctorCmdFactory() *cobra.Command {
	var strict, verbose bool
	c := &cobra.Command{
		Use:   "doctor",
		Short: "一键诊断 free-kiro 安装与配置",
		Long: `诊断以下项目并打印人类可读报告：

  ✓ free-kiro binary location + version
  ✓ PATH 配置（~/.local/bin 是否在 PATH 中）
  ✓ 当前目录 .kiro/ 工作区状态
  ✓ 本机已安装的 IDE（Claude Code / CodeBuddy）
  ✓ Hook 配置是否存在
  ✓ GitHub 最新 release 版本（可选）

--strict 启用更严格检查（如有警告也返回 exit 1）。
--verbose 报告所有受支持的 IDE（即使本机未安装也显示）。
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rep := runDoctorChecks(cmd.OutOrStdout(), verbose)
			if rep.FatalCount > 0 {
				return fmt.Errorf("%d fatal issue(s) — run the fixes listed above", rep.FatalCount)
			}
			if strict && rep.WarnCount > 0 {
				return fmt.Errorf("%d warning(s) (strict mode)", rep.WarnCount)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&strict, "strict", false, "treat warnings as failures")
	c.Flags().BoolVar(&verbose, "verbose", false, "report all supported IDEs even when not installed")
	return c
}

// doctorReport aggregates the per-check results.
type doctorReport struct {
	OK         int
	WarnCount  int
	FatalCount int
	InfoCount  int
	Issues     []doctorIssue
}

type doctorIssue struct {
	Severity string // "ok" | "info" | "warn" | "fatal"
	Title    string
	Detail   string
	Fix      string // optional one-line fix hint
}

func (i doctorIssue) render() string {
	prefix := "✓"
	switch i.Severity {
	case "info":
		prefix = "ℹ"
	case "warn":
		prefix = "⚠"
	case "fatal":
		prefix = "✗"
	}
	out := fmt.Sprintf("%s %s", prefix, i.Title)
	if i.Detail != "" {
		out += "\n    " + i.Detail
	}
	if i.Fix != "" {
		out += "\n    fix: " + i.Fix
	}
	return out
}

// runDoctorChecks executes every check and writes a formatted report.
// Exported so tests (and the install.sh script, in spirit) can reuse it.
func runDoctorChecks(w io.Writer, verbose bool) doctorReport {
	rep := doctorReport{}

	add := func(issue doctorIssue) {
		rep.Issues = append(rep.Issues, issue)
		switch issue.Severity {
		case "ok":
			rep.OK++
		case "info":
			rep.InfoCount++
		case "warn":
			rep.WarnCount++
		case "fatal":
			rep.FatalCount++
		}
	}

	// 1. Binary location + version (we always pass — we ARE the binary).
	add(doctorIssue{
		Severity: "ok",
		Title:    "free-kiro binary",
		Detail:   "running from " + selfPath(),
	})

	// 2. PATH configuration.
	pathIssue := checkPath()
	if pathIssue != nil {
		add(*pathIssue)
	} else {
		add(doctorIssue{Severity: "ok", Title: "PATH configuration", Detail: "$HOME/.local/bin is on PATH"})
	}

	// 3. Workspace.
	if cwd, err := os.Getwd(); err == nil {
		ws := workspace.Find(cwd)
		if !ws.Exists() {
			add(doctorIssue{
				Severity: "warn",
				Title:    "no .kiro/ workspace in this directory",
				Detail:   "current dir: " + cwd,
				Fix:      "run `free-kiro init` to create one",
			})
		} else {
			add(doctorIssue{
				Severity: "ok",
				Title:    "workspace",
				Detail:   ws.KiroDir(),
			})
		}
	}

	// 4. Installed IDEs — summarise. Detailed per-IDE hook state
	// appears in section 5.
	ides := ide.DetectAll("")
	installed := 0
	for _, d := range ides {
		if d.DirExists {
			installed++
		}
	}
	if installed == 0 {
		add(doctorIssue{
			Severity: "warn",
			Title:    "no AI coding IDE detected",
			Detail:   "checked ~/.claude/ and ~/.codebuddy/",
			Fix:      "install Claude Code (https://claude.ai/code) or CodeBuddy, then re-run `free-kiro init --ide auto`",
		})
	} else {
		var names []string
		for _, d := range ides {
			if d.DirExists {
				names = append(names, string(d.ID))
			}
		}
		add(doctorIssue{
			Severity: "ok",
			Title:    "IDE detected",
			Detail:   strings.Join(names, ", "),
		})
	}

	// 5. Hook configuration — only for INSTALLED IDEs by default.
	// Uninstalled IDEs are reported as info (or skipped entirely
	// without --verbose) so users don't get noise they can't act on.
	for _, d := range ides {
		if !d.DirExists {
			if verbose {
				add(doctorIssue{
					Severity: "info",
					Title:    string(d.ID) + " not installed",
					Detail:   "skipping hook check for " + d.ConfigPath,
				})
			}
			continue
		}
		ok, hookCount, err := countFreeKiroHooks(d.ConfigPath)
		if err != nil {
			add(doctorIssue{
				Severity: "warn",
				Title:    string(d.ID) + " hooks unreadable",
				Detail:   err.Error(),
				Fix:      "settings.json may be malformed; back it up and re-run `free-kiro init --ide " + string(d.ID) + "`",
			})
			continue
		}
		if !ok || hookCount == 0 {
			add(doctorIssue{
				Severity: "warn",
				Title:    string(d.ID) + " hooks not installed",
				Detail:   d.ConfigPath,
				Fix:      "run `free-kiro init --ide " + string(d.ID) + "`",
			})
		} else {
			add(doctorIssue{
				Severity: "ok",
				Title:    string(d.ID) + " hooks",
				Detail:   fmt.Sprintf("%d free-kiro hook(s) installed", hookCount),
			})
		}
	}

	// 6. Latest release (best-effort, 5s timeout).
	if latest := fetchLatestVersion(); latest != "" {
		add(doctorIssue{
			Severity: "ok",
			Title:    "GitHub releases reachable",
			Detail:   "latest: v" + latest,
		})
	} else {
		add(doctorIssue{
			Severity: "warn",
			Title:    "could not reach GitHub",
			Detail:   "skipping version check (network blocked?)",
		})
	}

	// Render.
	w.Write([]byte("free-kiro doctor\n"))
	w.Write([]byte("================\n\n"))
	for _, issue := range rep.Issues {
		w.Write([]byte(issue.render()))
		w.Write([]byte("\n\n"))
	}
	fmt.Fprintf(w, "summary: %d ok, %d info, %d warn, %d fatal\n",
		rep.OK, rep.InfoCount, rep.WarnCount, rep.FatalCount)

	return rep
}

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
	defer resp.Body.Close()
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

// keep exec import referenced for potential future checks.
var _ = exec.Command