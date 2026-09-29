package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/skill"
)

// skillCmdFactory assembles the `free-kiro skill {install,uninstall,
// update,show,path,version}` subcommand tree. Subcommands:
//   - install   write SKILL.md bundle to ~/.{app}/skills/<name>/
//   - uninstall remove the installed bundle
//   - update    re-fetch latest and re-install
//   - show      print installed state per app
//   - path      print resolved skills dir for one app
//   - version   print bundle version + binary build version
func skillCmdFactory() *cobra.Command {
	c := &cobra.Command{
		Use:   "skill",
		Short: "管理 free-kiro 的可安装 SKILL.md bundle",
		Long: `SKILL.md bundle 把 free-kiro 的能力暴露给 Claude Code / OpenCode /
Codex CLI / CodeBuddy。装到对应 AI 助手的 ~/.{app}/skills/<name>/
后，AI 在用户谈到 "spec / EARS / lint / PRD / wave" 时会自动调用。

四种安装入口：
  free-kiro skill install              # 推荐：binary 内部安装
  curl -fsSL .../contrib/skills/install.sh | bash   # curl|bash
  npx @jingyu525/free-kiro-skill                   # npm 包装（薄）
  手动从 GitHub Releases 解压                      # 完全离线

退出码：
  0   成功（或已是最新 / 已是当前版本）
  1   sha256 校验失败 / 安装异常
  2   engine error（网络 / GitHub API 失败）
  3   参数错`,
	}
	c.AddCommand(skillInstallCmd(), skillUninstallCmd(), skillUpdateCmd(),
		skillShowCmd(), skillPathCmd(), skillVersionCmd())
	return c
}

func skillInstallCmd() *cobra.Command {
	var (
		appFlag string
		from    string
		version string
		dryRun  bool
		force   bool
	)
	c := &cobra.Command{
		Use:   "install",
		Short: "把 SKILL.md bundle 安装到目标 app 的 skills 目录",
		Long: `下载 free-kiro-skill_<version>.zip（默认最新 release），校验 sha256，
写入 ~/.{app}/skills/<name>/。

  $ free-kiro skill install                          # 装到全部 detected apps
  $ free-kiro skill install --app claude-code        # 只装到 Claude Code
  $ free-kiro skill install --app all                # 全部（包括未 detected）
  $ free-kiro skill install --dry-run                # 只打印，不写盘
  $ free-kiro skill install --version 0.7.0          # 指定版本
  $ free-kiro skill install --from /path/to/dir      # 本地已解压目录
  $ free-kiro skill install --from https://.../x.zip # 远程 zip URL
  $ free-kiro skill install --force                  # 覆盖已安装版本

退出码：
  0   全部成功（含 already-current）
  1   部分 / 全部 app 失败
  2   engine error`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			return runSkillInstall(ctx, cmd, appFlag, from, version, dryRun, force)
		},
	}
	c.Flags().StringVar(&appFlag, "app", "all", "目标 app: all | claude-code | opencode | codex | codebuddy")
	c.Flags().StringVar(&from, "from", "", "本地目录或 zip URL（跳过 GitHub 下载）")
	c.Flags().StringVar(&version, "version", "", "bundle 版本（默认 latest）")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "只打印计划，不写盘")
	c.Flags().BoolVar(&force, "force", false, "覆盖已安装版本（默认短路返回）")
	return c
}

func runSkillInstall(ctx context.Context, cmd *cobra.Command, appFlag, from, version string, dryRun, force bool) error {
	targets, home, err := targetsAndHome(appFlag)
	if err != nil {
		return exitWithError(err)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Installing free-kiro SKILL.md bundle...")
	fmt.Fprintf(out, "  source:  %s\n", sourceLabel(from, version))
	labels := make([]string, len(targets))
	for i, a := range targets {
		labels[i] = a.Label()
	}
	fmt.Fprintf(out, "  target:  %s\n", strings.Join(labels, ", "))
	if dryRun {
		fmt.Fprintln(out, "  mode:    dry-run (no writes)")
	} else {
		fmt.Fprintln(out, "  mode:    write")
	}
	fmt.Fprintln(out)

	anyErr := false
	for _, app := range targets {
		res := skill.InstallOne(ctx, skill.InstallOptions{
			App:     app,
			Home:    home,
			Subdir:  "free-kiro",
			Source:  from,
			Version: version,
			DryRun:  dryRun,
			Force:   force,
		})
		if res.Err != nil {
			fmt.Fprintf(out, "  ✗ %-12s %s\n", app.Label(), res.Err)
			anyErr = true
			continue
		}
		fmt.Fprintf(out, "  ✓ %-12s %-15s %s\n", app.Label(), res.Status, res.SkillsDir)
		if res.Status != "dry-run" && res.Status != "already-current" {
			fmt.Fprintf(out, "               version=%s files=%d sha256=%v\n",
				res.Version, res.FilesWritten, res.SHA256OK)
		}
	}
	fmt.Fprintln(out)
	if anyErr {
		return exitWithError(fmt.Errorf("one or more apps failed; see above"))
	}
	fmt.Fprintln(out, "Done. Verify with: free-kiro skill show")
	return nil
}

func skillUninstallCmd() *cobra.Command {
	var appFlag string
	c := &cobra.Command{
		Use:   "uninstall",
		Short: "移除已安装的 SKILL.md bundle",
		Long: `删除 ~/.{app}/skills/free-kiro/。未装则 no-op（幂等）。

  $ free-kiro skill uninstall --app claude-code
  $ free-kiro skill uninstall --app all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, home, err := targetsAndHome(appFlag)
			if err != nil {
				return exitWithError(err)
			}
			out := cmd.OutOrStdout()
			for _, app := range targets {
				removed, err := skill.UninstallOne(app, home, "free-kiro")
				if err != nil {
					fmt.Fprintf(out, "  ✗ %-12s %s\n", app.Label(), err)
					return exitWithError(err)
				}
				if len(removed) == 0 {
					fmt.Fprintf(out, "  · %-12s (not installed)\n", app.Label())
				} else {
					fmt.Fprintf(out, "  ✓ %-12s removed %s\n", app.Label(), removed[0])
				}
			}
			return nil
		},
	}
	c.Flags().StringVar(&appFlag, "app", "all", "目标 app: all | claude-code | opencode | codex | codebuddy")
	return c
}

func skillUpdateCmd() *cobra.Command {
	var (
		appFlag string
		check   bool
		force   bool
	)
	c := &cobra.Command{
		Use:   "update",
		Short: "检查 / 应用 SKILL.md bundle 更新",
		Long: `拉 latest release，比对已装版本；不同则重装。

  $ free-kiro skill update           # 检查 + 更新
  $ free-kiro skill update --check   # 只检查（exit 0 = 最新, 1 = 有新版）`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			targets, home, err := targetsAndHome(appFlag)
			if err != nil {
				return exitWithError(err)
			}
			latest, err := skill.LatestRelease(ctx)
			if err != nil {
				return exitWithError(err)
			}
			out := cmd.OutOrStdout()
			// Snapshot once: ShowInstalled walks all 4 apps + reads manifests,
			// so calling it per target would do up to 16 stats/manifest reads.
			installed := skill.ShowInstalled(home, "free-kiro")
			installedVer := map[skill.App]string{}
			for _, s := range installed {
				if s.Installed {
					installedVer[s.App] = s.Version
				}
			}
			anyNewer := false
			anyErr := false
			for _, app := range targets {
				installedVersion := installedVer[app]
				if installedVersion == "" {
					installedVersion = "(none)"
				}
				if installedVersion == latest {
					fmt.Fprintf(out, "  = %-12s up-to-date (%s)\n", app.Label(), latest)
					continue
				}
				anyNewer = true
				if check {
					fmt.Fprintf(out, "  ↑ %-12s %s → %s\n", app.Label(), installedVersion, latest)
					continue
				}
				res := skill.InstallOne(ctx, skill.InstallOptions{
					App: app, Home: home, Subdir: "free-kiro", Version: latest, Force: force,
				})
				if res.Err != nil {
					fmt.Fprintf(out, "  ✗ %-12s %s\n", app.Label(), res.Err)
					anyErr = true
					continue
				}
				fmt.Fprintf(out, "  ✓ %-12s %s (%s → %s)\n",
					app.Label(), res.Status, installedVersion, latest)
			}
			if check && anyNewer {
				return exitWithError(fmt.Errorf("updates available — re-run without --check"))
			}
			if anyErr {
				return exitWithError(fmt.Errorf("one or more apps failed"))
			}
			fmt.Fprintln(out, "Done.")
			return nil
		},
	}
	c.Flags().StringVar(&appFlag, "app", "all", "目标 app: all | claude-code | opencode | codex | codebuddy")
	c.Flags().BoolVar(&check, "check", false, "只检查；不实际安装")
	c.Flags().BoolVar(&force, "force", false, "强制重装（即使已是最新）")
	return c
}

func skillShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "打印每个 app 的已装 SKILL.md bundle 状态",
		Long: `逐 app 列出已装的 bundle：路径 / 版本 / free-kiro 最低版本要求。

  $ free-kiro skill show`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := skill.HomeDir()
			if err != nil {
				return exitWithError(err)
			}
			states := skill.ShowInstalled(home, "free-kiro")
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "binary:  %s (%s, %s)\n", buildVersion, buildCommit, buildDate)
			fmt.Fprintln(out)
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "APP\tSTATUS\tVERSION\tPATH")
			for _, s := range states {
				status := "not installed"
				version := "-"
				if s.Installed {
					status = "installed"
					version = s.Version
				}
				exp := ""
				if s.Experimental {
					exp = " (experimental)"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s%s\n",
					s.App.Label(), status, version, s.SkillsDir, exp)
			}
			tw.Flush()
			return nil
		},
	}
}

func skillPathCmd() *cobra.Command {
	var appFlag string
	c := &cobra.Command{
		Use:   "path",
		Short: "打印指定 app 的 SKILL.md bundle 安装路径",
		Long: `解析 ~/.{app}/skills/free-kiro/ 的绝对路径。

  $ free-kiro skill path --app claude-code`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := skill.ParseApp(appFlag)
			if err != nil {
				return exitWithError(err)
			}
			if app == "" {
				return exitWithError(fmt.Errorf("--app required (claude-code, opencode, codex, codebuddy)"))
			}
			home, err := skill.HomeDir()
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), skill.SkillsDir(app, home, "free-kiro"))
			return nil
		},
	}
	c.Flags().StringVar(&appFlag, "app", "", "目标 app (必填)")
	return c
}

func skillVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "打印 bundle 与 binary 版本",
		Long: `对比已装 bundle 版本（从任意已装 app 读 skill.json）与 binary buildVersion。

  $ free-kiro skill version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "binary:   %s\n", buildVersion)
			home, err := skill.HomeDir()
			if err != nil {
				fmt.Fprintf(out, "(home: %s)\n", err)
				return nil
			}
			states := skill.ShowInstalled(home, "free-kiro")
			for _, s := range states {
				if s.Installed {
					fmt.Fprintf(out, "bundle:   %s (%s)\n", s.Version, s.App.Label())
					if s.FreeKiroMin != "" && s.FreeKiroMin != buildVersion && buildVersion != "dev" {
						fmt.Fprintf(out, "  warn:   bundle requires free-kiro >= %s\n", s.FreeKiroMin)
					}
				}
			}
			return nil
		},
	}
}

// resolveAppTargets parses --app flag. "all" returns the full list;
// any other value returns a single-element list.
func resolveAppTargets(appFlag string) ([]skill.App, error) {
	app, err := skill.ParseApp(appFlag)
	if err != nil {
		return nil, err
	}
	if app != "" {
		return []skill.App{app}, nil
	}
	return skill.AllApps(), nil
}

// targetsAndHome is the shared prelude for the install/uninstall/update
// subcommands: parse --app, resolve home directory.
func targetsAndHome(appFlag string) ([]skill.App, string, error) {
	targets, err := resolveAppTargets(appFlag)
	if err != nil {
		return nil, "", err
	}
	home, err := skill.HomeDir()
	if err != nil {
		return nil, "", err
	}
	return targets, home, nil
}

func sourceLabel(from, version string) string {
	if from != "" {
		return from
	}
	if version != "" {
		return "v" + version + " (GitHub release)"
	}
	return "latest (GitHub release)"
}