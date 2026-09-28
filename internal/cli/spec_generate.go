package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/liujingyu/free-kiro/internal/models"
)

func specGenerateCmd() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "generate <name>",
		Short: "生成 spec 文档（requirements / design / tasks / all）",
		Args:  cobra.ExactArgs(1),
		Long: `生成 planning 文档。默认生成全部（按 workflow 顺序）；可指定单 phase：

  --phase requirements    生成 requirements.md（或 bugfix.md）
  --phase design          生成 design.md
  --phase tasks           生成 tasks.md
  --phase all             按 workflow 顺序生成全部（默认）

  --force                  覆盖已存在的文档（默认不覆盖）

lint gate：前向进阶会被门禁拦截（draft→requirements→design→tasks→approved）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			phaseStr, _ := cmd.Flags().GetString("phase")
			if phaseStr == "all" || phaseStr == "" {
				paths, err := eng.GenerateAll(args[0], force)
				if err != nil {
					return exitWithError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(),
					"generated %d document(s) for %q\n", len(paths), args[0])
				return nil
			}
			meta, err := eng.Generate(args[0], models.Phase(phaseStr), force)
			if err != nil {
				return exitWithError(err)
			}
			doc := models.PhaseDocFor(meta.SpecType, models.Phase(phaseStr))
			fmt.Fprintf(cmd.OutOrStdout(), "generated %s for %q\n", doc, args[0])
			return nil
		},
	}
	c.Flags().String("phase", "all", "phase to generate: requirements | design | tasks | all")
	c.Flags().BoolVar(&force, "force", false, "overwrite existing documents")
	return c
}