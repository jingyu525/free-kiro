package cli

import (
	"github.com/spf13/cobra"

	"github.com/liujingyu/free-kiro/internal/workspace"
)

// specCmd groups the 12 spec lifecycle subcommands. The actual command
// implementations live in spec_*.go files; this file owns the parent
// command and the shared workspace/engine factory.
//
// `free-kiro spec` without a subcommand prints the same help as
// `free-kiro spec --help` (cobra's default behaviour when no subcommand
// matches).
var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "管理 spec（new / generate / approve / start / complete …）",
	Long: `spec 子命令是 free-kiro 工作流的核心入口：

  new        新建 spec（含 workflow / type / quick 变体）
  generate   生成 planning 文档（requirements / design / tasks / bugfix）
  quick      一次性生成 + 免审批（Quick Spec 变体）
  show       打印某个 phase 的文档
  list       列出全部 specs
  approve    审批 + 捕获 baseline（必须过 lint gate）
  status     查看 spec 状态 + 漂移（JSON）
  next       预言机：返回下一步动作 + 命令（JSON）
  sync       重新基线化（合法编辑后消除漂移）
  start      标记开始实现（APPROVED → IMPLEMENTING）
  complete   标记 spec 完成（IMPLEMENTING → DONE）
  analyze    advisory 一致性分析（不阻塞）`,
	// No RunE — cobra will print help with the subcommand list when
	// `free-kiro spec` is invoked without arguments.
}

// loadEngine returns a SpecEngine rooted at the workspace containing the
// current directory (or above). Exits with KiroError if no .kiro exists.
func loadEngine() (*engineAndWS, error) {
	ws, err := workspace.Find("").Require()
	if err != nil {
		return nil, exitWithError(err)
	}
	return &engineAndWS{ws: ws}, nil
}

// engineAndWS is a tiny holder to keep the factory signature stable
// while the actual engine type lives in a different package (avoids an
// import cycle: cli → spec).
type engineAndWS struct {
	ws *workspace.Workspace
}