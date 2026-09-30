package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/upgrade"
)

// Build-time version info — populated by GoReleaser via ldflags:
//
//	-X main.version={{.Version}}  -X main.commit={{.Commit}}
//	-X main.date={{.Date}}
//
// Falls back to "dev" when built locally without the makefile target
// (see .goreleaser.yaml for the build flags).
var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

// upgradeCmd checks the GitHub release stream and (by default)
// upgrades the running binary in place. Use --check to preview only.
//
// Re-execs the new binary on success (POSIX systems) so the user
// sees the new version immediately. On Windows, Apply instructs the
// user to re-launch manually.
func upgradeCmdFactory() *cobra.Command {
	var checkOnly, force bool
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "检查并升级 free-kiro 到最新 release",
		Long: `查询 GitHub releases/latest，比较当前二进制版本，下载并替换为最新版本。

  $ free-kiro upgrade               # 检查 + 升级（如有新版本）
  $ free-kiro upgrade --check       # 只检查，不下载
  $ free-kiro upgrade --force       # 强制重装当前版本（修复用）

升级流程：
  1. 调 GitHub API 拉 latest release
  2. 匹配当前平台（darwin/linux/windows + amd64/arm64）的 tarball
  3. 下载 + SHA256 校验
  4. 解压到临时路径 + 原子 rename 到当前 binary 位置
  5. re-exec 替换当前进程（POSIX；Windows 提示手动重启）

退出码：
  0   已是最新（或升级成功）
  1   网络 / SHA256 / IO 失败
  2   当前 binary 路径无法解析`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			plan, err := upgrade.Check(ctx, buildVersion)
			if err != nil {
				return exitWithError(err)
			}
			printPlan(cmd, plan)
			if checkOnly {
				if !plan.Same {
					return exitWithError(ferrors.NewUsage("upgrade.check", "a newer version is available — re-run without --check to upgrade"))
				}
				return nil
			}
			if err := upgrade.Apply(ctx, plan, force); err != nil {
				return exitWithError(err)
			}
			// On POSIX, Apply re-execs and we never reach this line.
			// On Windows, instruct the user to restart.
			fmt.Fprintln(cmd.OutOrStdout(),
				"upgrade installed. Please re-run `free-kiro` to use the new version.")
			return nil
		},
	}
	c.Flags().BoolVar(&checkOnly, "check", false, "only check for updates; do not download or install")
	c.Flags().BoolVar(&force, "force", false, "force reinstall even when current == latest")
	return c
}

func printPlan(cmd *cobra.Command, p *upgrade.Plan) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "current: %s\n", displayVersion(p.Current))
	fmt.Fprintf(out, "latest:  v%s\n", p.Latest)
	if p.Same {
		fmt.Fprintln(out, "status:  already on the latest version")
	} else {
		fmt.Fprintln(out, "status:  update available")
	}
	fmt.Fprintf(out, "target:  %s\n", p.Target)
	if !p.Same {
		fmt.Fprintf(out, "download: %s\n", p.Download)
	}
}

func displayVersion(v string) string {
	if v == "" {
		return "(unknown — built without ldflags)"
	}
	return "v" + v
}