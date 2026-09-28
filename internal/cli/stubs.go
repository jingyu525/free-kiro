package cli

import (
	"github.com/spf13/cobra"
)

// These are stub commands so root.go can wire them in before the full
// implementations land in Wave 2/3/4. Each stub prints a "not yet
// implemented" message and exits 0; they will be replaced as the
// corresponding waves complete.
var specCmd = stubCmd("spec", "管理 spec（new / generate / approve / start / complete …）")
var steeringCmd = stubCmd("steering", "查看项目约定文档")
var taskCmd = stubCmd("task", "查看 tasks.md 的并行 wave 视图")
var hookCmd = stubCmd("hook", "管理事件驱动 hook")
var lintCmd = stubCmd("lint", "离线质量门禁")

func stubCmd(name, short string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("(子命令 %q 在后续 wave 实现；当前仅为占位符)\n", name)
			return nil
		},
	}
}