package cli

import (
	"strings"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
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
//
// The heavy lifting for install / uninstall / update lives in
// skill_install.go and skill_sync.go; this file owns the parent command
// tree + shared helpers (resolveAppTargets, targetsAndHome, sourceLabel,
// addAppFlag) used by every subcommand.
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

func skillPathCmd() *cobra.Command {
	var appFlag string
	c := &cobra.Command{
		Use:   "path",
		Short: "打印指定 app 的 SKILL.md bundle 安装路径",
		Long: `解析 ~/.{app}/skills/free-kiro/ 的绝对路径。

  $ free-kiro skill path --app claude-code`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := skill.ParseApp(appFlag)
			if err != nil {
				return exitWithError(err)
			}
			if app == "" {
				return exitWithError(ferrors.NewUsage("skill.path", "--app required (claude-code, opencode, codex, codebuddy)"))
			}
			home, err := skill.HomeDir()
			if err != nil {
				return exitWithError(err)
			}
			writeOutln(cmd.OutOrStdout(), skill.SkillsDir(app, home, "free-kiro"))
			return nil
		},
	}
	addAppFlag(c, &appFlag, "")
	return c
}

func skillVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "打印 bundle 与 binary 版本",
		Long: `对比已装 bundle 版本（从任意已装 app 读 skill.json）与 binary buildVersion。

  $ free-kiro skill version`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			writeOut(out, "binary:   %s\n", buildVersion)
			home, err := skill.HomeDir()
			if err != nil {
				writeOut(out, "(home: %s)\n", err)
				return nil
			}
			states := skill.ShowInstalled(home, "free-kiro")
			for _, s := range states {
				if s.Installed {
					writeOut(out, "bundle:   %s (%s)\n", s.Version, s.App.Label())
					if s.FreeKiroMin != "" && s.FreeKiroMin != buildVersion && buildVersion != "dev" {
						writeOut(out, "  warn:   bundle requires free-kiro >= %s\n", s.FreeKiroMin)
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

// addAppFlag wires the `--app` flag onto cmd. defaultVal should be "all"
// for install/uninstall/update and "" for path (where a specific app is
// required).
func addAppFlag(cmd *cobra.Command, target *string, defaultVal string) {
	cmd.Flags().StringVar(target, "app", defaultVal,
		"目标 app: all | claude-code | opencode | codex | codebuddy")
}

// keep the strings import referenced when the subcommand helpers move
// to other files in this package; future additions can drop this.
var _ = strings.Join
