package cli

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/skill"
)

// skillUpdateCmd checks / applies SKILL.md bundle updates. With --check
// it only reports whether a newer version is available; without it, it
// fetches the latest and reinstalls into each target app.
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
				if check {
					fmt.Fprintf(out, "  ↑ %-12s %s → %s\n", app.Label(), installedVersion, latest)
					anyNewer = true
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
				return exitWithError(ferrors.NewUsage("skill.update.check", "updates available — re-run without --check"))
			}
			if anyErr {
				return exitWithError(ferrors.New("skill.update", "one or more apps failed"))
			}
			fmt.Fprintln(out, "Done.")
			return nil
		},
	}
	addAppFlag(c, &appFlag, "all")
	c.Flags().BoolVar(&check, "check", false, "只检查；不实际安装")
	c.Flags().BoolVar(&force, "force", false, "强制重装（即使已是最新）")
	return c
}

// skillShowCmd prints the installed SKILL.md bundle state for every app.
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

// ensure context is referenced when all subcommands in this file move
// to other packages — currently used by skillUpdateCmd above.
var _ context.Context
