package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/ide"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// GitHubRepo is the GitHub owner/repo slug for this project. Used by
// `upgrade` and `doctor` to look up releases / report issues.
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
//
// Heavy lifting is split across two files:
//
//	doctor_report.go — doctorReport / doctorIssue types + formatter
//	doctor_checks.go — individual IO checks (PATH, hooks, latest
//	                   release) used by runDoctorChecks below
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			rep := runDoctorChecks(cmd.OutOrStdout(), verbose)
			if rep.FatalCount > 0 {
				// Fatal issues map to engine error (exit 2); matches the
				// original doctor comment "Exits 0 when healthy, 1 when
				// warnings, 2 when fatal".
				return ferrors.New("doctor.run", fmt.Sprintf("%d fatal issue(s) — run the fixes listed above", rep.FatalCount))
			}
			if strict && rep.WarnCount > 0 {
				// Strict-mode warnings also surface as engine error (exit 2)
				// after the ExitCode unification — the user-visible signal
				// (non-zero exit + warning detail) is preserved.
				return ferrors.New("doctor.run.strict", fmt.Sprintf("%d warning(s) (strict mode)", rep.WarnCount))
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
	if cwd, err := getwd(); err == nil {
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
			Detail:   joinStrings(names, ", "),
		})
	}

	// 5. Hook configuration — only for INSTALLED IDEs by default.
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
	_, _ = w.Write([]byte("free-kiro doctor\n"))
	_, _ = w.Write([]byte("================\n\n"))
	for _, issue := range rep.Issues {
		_, _ = w.Write([]byte(issue.render()))
		_, _ = w.Write([]byte("\n\n"))
	}
	writeOut(w, "summary: %d ok, %d info, %d warn, %d fatal\n",
		rep.OK, rep.InfoCount, rep.WarnCount, rep.FatalCount)

	return rep
}
