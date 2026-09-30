package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/steering"
)

// steeringCmd replaces the Wave 1 stub with three subcommands:
// list, show, context. The parent command has no RunE so cobra prints
// help with the subcommand list when invoked bare.
func steeringCmdFactory() *cobra.Command {
	c := &cobra.Command{
		Use:   "steering",
		Short: "查看 / 注入项目约定文档",
		Long: `steering 文档是项目的"持久上下文"——product / structure / tech / AGENTS
等 markdown 文件，每次生成都会按 mode 规则注入：

  always     每次都注入
  auto       prompt 关键词命中 description 才注入
  manual     只有显式 #name 引用才注入
  filematch  编辑的文件匹配 fileMatchPattern 才注入

两级作用域合并：workspace <root>/.kiro/steering/*.md 覆盖同名 global ~/.kiro/steering/*.md。`,
	}
	c.AddCommand(steeringListCmd())
	c.AddCommand(steeringShowCmd())
	c.AddCommand(steeringContextCmd())
	return c
}

func steeringListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出全部 steering 文档",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := newSteeringStore()
			if err != nil {
				return err
			}
			docs := store.LoadAll()
			if len(docs) == 0 {
				writeOutln(cmd.OutOrStdout(), "(no steering docs)")
				return nil
			}
			sort.Slice(docs, func(i, j int) bool { return docs[i].Name < docs[j].Name })
			writeOut(cmd.OutOrStdout(), "%-20s %-10s %-10s DESCRIPTION\n", "NAME", "SCOPE", "MODE")
			for _, d := range docs {
				writeOut(cmd.OutOrStdout(), "%-20s %-10s %-10s %s\n",
					d.Name, d.Scope, d.Mode, d.Description)
			}
			return nil
		},
	}
}

func steeringShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "打印单个 steering 文档",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := newSteeringStore()
			if err != nil {
				return err
			}
			doc := store.Get(args[0])
			if doc == nil {
				return exitWithError(ferrors.NewUsage("steering.show", fmt.Sprintf("steering doc %q not found", args[0])))
			}
			writeOut(cmd.OutOrStdout(), "# %s (mode=%s, scope=%s)\n\n%s\n",
				doc.Name, doc.Mode, doc.Scope, doc.Content)
			return nil
		},
	}
}

func steeringContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "组装 steering 上下文（agent 注入用）",
		Long: `组装当前应注入到生成上下文的 steering block。

  --file <path>     触发 fileMatch 模式（编辑该文件时拉起相关约定）
  --prompt <text>   触发 auto 模式（关键词命中 description 的文档会被拉起）

agent 的工具应在每次生成前调用此命令（按需加 --file），把输出拼到生成 prompt 前缀。`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := newSteeringStore()
			if err != nil {
				return err
			}
			promptFlag, _ := cmd.Flags().GetString("prompt")
			fileFlag, _ := cmd.Flags().GetString("file")
			ctx := store.Assemble(promptFlag, nil, fileFlag)
			if ctx == "" {
				writeOutln(cmd.OutOrStdout(), "(no steering matched)")
				return nil
			}
			writeOutln(cmd.OutOrStdout(), ctx)
			return nil
		},
	}
}

// newSteeringStore loads the workspace + constructs a SteeringStore.
// Returns an error if the workspace doesn't exist (mirrors other CLI
// commands' behaviour for a missing .kiro).
func newSteeringStore() (*steering.Store, error) {
	holder, err := loadEngine()
	if err != nil {
		return nil, err
	}
	return steering.NewStore(holder.ws, ""), nil
}

// Ensure filepath import is used (it's referenced via filepath.ToSlash
// in steering.Assemble; this is a build-only reference to keep the
// import here so future changes can use it directly without editing).
var _ = filepath.ToSlash
