package cli

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/skill"
)

// skillInstallCmd writes the SKILL.md bundle to each target app's
// skills directory. Downloads from GitHub releases by default, or from
// --from <local-dir-or-zip-url>.
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			return runSkillInstall(ctx, cmd, appFlag, from, version, dryRun, force)
		},
	}
	addAppFlag(c, &appFlag, "all")
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
	writeOutln(out, "Installing free-kiro SKILL.md bundle...")
	writeOut(out, "  source:  %s\n", sourceLabel(from, version))
	labels := make([]string, len(targets))
	for i, a := range targets {
		labels[i] = a.Label()
	}
	writeOut(out, "  target:  %s\n", strings.Join(labels, ", "))
	if dryRun {
		writeOutln(out, "  mode:    dry-run (no writes)")
	} else {
		writeOutln(out, "  mode:    write")
	}
	writeOutln(out)

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
			writeOut(out, "  ✗ %-12s %s\n", app.Label(), res.Err)
			anyErr = true
			continue
		}
		writeOut(out, "  ✓ %-12s %-15s %s\n", app.Label(), res.Status, res.SkillsDir)
		if res.Status != "dry-run" && res.Status != "already-current" {
			writeOut(out, "               version=%s files=%d sha256=%v\n",
				res.Version, res.FilesWritten, res.SHA256OK)
		}
	}
	writeOutln(out)
	if anyErr {
		return exitWithError(ferrors.New("skill.install", "one or more apps failed; see above"))
	}
	writeOutln(out, "Done. Verify with: free-kiro skill show")
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			targets, home, err := targetsAndHome(appFlag)
			if err != nil {
				return exitWithError(err)
			}
			out := cmd.OutOrStdout()
			for _, app := range targets {
				removed, err := skill.UninstallOne(app, home, "free-kiro")
				if err != nil {
					writeOut(out, "  ✗ %-12s %s\n", app.Label(), err)
					return exitWithError(err)
				}
				if len(removed) == 0 {
					writeOut(out, "  · %-12s (not installed)\n", app.Label())
				} else {
					writeOut(out, "  ✓ %-12s removed %s\n", app.Label(), removed[0])
				}
			}
			return nil
		},
	}
	addAppFlag(c, &appFlag, "all")
	return c
}
