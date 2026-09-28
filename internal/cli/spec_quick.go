package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func specQuickCmd() *cobra.Command {
	var specType string
	c := &cobra.Command{
		Use:   "quick <name>",
		Short: "Quick Spec：一次性生成 + 免审批",
		Long: `Quick Spec 变体：新建 spec + 一次性生成 requirements/design/tasks 三份文档，
并免除 formal ` + "`approve`" + ` 仪式——生成完即可 ` + "`spec start`" + ` 进入实现。
lint 门禁仍生效（文档错了照样拦），只是跳过人工签字那一步。

适合小改动 / 你已胸有成竹的变更；大功能仍走正式 requirements → approve 流程。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			promptFlag, _ := cmd.Flags().GetString("prompt")
			p, err := readPrompt(promptFlag)
			if err != nil {
				return exitWithError(err)
			}
			meta, err := eng.NewSpec(args[0], p, "requirements-first", specType, true)
			if err != nil {
				return exitWithError(err)
			}
			if _, err := eng.GenerateAll(args[0], false); err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"quick spec %q generated (requirements/design/tasks); approval waived\n", meta.Name)
			fmt.Fprintf(cmd.OutOrStdout(),
				"next: free-kiro spec start %s   (or review, then free-kiro spec approve %s)\n",
				meta.Name, meta.Name)
			return nil
		},
	}
	c.Flags().String("prompt", "", "spec prompt text (or pipe via stdin)")
	c.Flags().StringVar(&specType, "type", "feature", "spec type: feature or bugfix")
	return c
}