package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/liujingyu/free-kiro/internal/workspace"
)

// initCmd bootstraps a .kiro workspace in the given directory (default:
// current directory) and writes three sample steering documents:
//
//	.kiro/steering/product.md     (mode: always)
//	.kiro/steering/structure.md   (mode: always)
//	.kiro/steering/tech.md        (mode: auto)
//
// Re-running on an existing workspace is a no-op for existing files; only
// missing directories and settings.json are created.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 .kiro 工作区 + 写入示例 steering 文档",
	Long: `在指定目录初始化 free-kiro 工作区：

  .kiro/
  ├── settings.json    生成器配置
  ├── specs/           每个 spec 一个子目录
  ├── steering/        项目级约定文档（product/structure/tech）
  └── hooks/           事件驱动自动化

若目录已存在 .kiro，重复运行是幂等的（不会覆盖已有 steering / settings）。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		ws := workspace.New(path)
		if err := ws.EnsureLayout(); err != nil {
			return exitWithError(err)
		}
		// Write sample steering docs only if missing.
		for name, content := range sampleSteering() {
			dest := ws.SteeringDir() + "/" + name
			if _, err := stat(dest); err == nil {
				continue
			}
			if err := writeFile(dest, content); err != nil {
				return exitWithError(err)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已初始化 free-kiro 工作区于 %s\n", ws.KiroDir())
		fmt.Fprintln(cmd.OutOrStdout(), "已写入示例 steering：product.md、structure.md、tech.md")
		return nil
	},
}

func init() {
	initCmd.Flags().String("path", ".", "目标项目根目录（默认：当前目录）")
}

func sampleSteering() map[string]string {
	return map[string]string{
		"product.md": "---\nmode: always\ndescription: 这个产品的目标、核心能力与边界\n---\n# Product\n\n用 1-3 段说明产品的目标和必须提供的核心能力。\n本文档作为 always 上下文注入到每次生成中。\n",
		"structure.md": "---\nmode: always\ndescription: 代码组织与架构约定\n---\n# Structure\n\n- 模块保持小且单一职责。\n- IO 隔离在显式适配器后面。\n- 除非有明确理由，优先使用标准库而非新增依赖。\n",
		"tech.md": "---\nmode: auto\ndescription: Go 项目技术栈与开发规范\n---\n# Tech\n\n- 语言：Go 1.22+。\n- 不引入非必要第三方依赖。\n- 测试用标准库 testing 包，断言用 testify。\n",
	}
}

// stat / writeFile are tiny indirection so the tests can stub them later
// without rewriting the command.
var (
	stat      = func(p string) (any, error) { return osStat(p) }
	writeFile = func(p, content string) error { return osWriteFile(p, content) }
)