package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/visualize"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// reportCmd generates a complete markdown project report at
// .kiro/REPORT.md (or stdout with --stdout). Designed to be committed
// alongside code so reviewers can see the spec state without leaving
// the PR.
func reportCmdFactory() *cobra.Command {
	var stdout bool
	c := &cobra.Command{
		Use:   "report",
		Short: "生成项目级 markdown 报告（贴 PR 即可见）",
		Long: `聚合所有 spec 的当前状态、drift、wave 进度到一个 markdown 文件：

  $ free-kiro report                     # 写入 .kiro/REPORT.md
  $ free-kiro report --stdout            # 输出到 stdout（适合管道）
  $ free-kiro report --output ./REPORT.md # 自定义路径

报告内容：
  - Summary 表（spec / phase / drift / tasks）
  - Drift alerts 表 + 修复指引
  - 项目级 Mermaid 图（所有 spec 概览）
  - 每个 spec 的 phase / workflow / wave Mermaid 图

适合贴到 GitHub PR description、Notion、Confluence。`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			holder, err := loadEngine()
			if err != nil {
				return err
			}
			ws := workspace.New(holder.ws.Root())
			path, _ := cmd.Flags().GetString("output")
			if stdout {
				path = "-"
			}
			if path == "" {
				path = filepath.Join(ws.KiroDir(), "REPORT.md")
			}
			r, err := visualize.RenderAndWriteReport(path, ws, holder.engine())
			if err != nil {
				return exitWithError(err)
			}
			if stdout {
				// file already went to stdout; don't double-print.
				return nil
			}
			writeOut(cmd.OutOrStdout(),
				"wrote report (%d specs, %d drift) to %s\n",
				len(r.Specs), driftCount(r.Specs), path)
			return nil
		},
	}
	c.Flags().String("output", "", "output file path (default: .kiro/REPORT.md)")
	c.Flags().BoolVar(&stdout, "stdout", false, "write to stdout instead of a file")
	return c
}

func driftCount(specs []*visualize.SpecReport) int {
	n := 0
	for _, s := range specs {
		n += len(s.Drift)
	}
	return n
}
