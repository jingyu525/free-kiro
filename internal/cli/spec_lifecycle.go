package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// specApproveCmd marks a spec approved and captures a drift baseline.
// Runs the lint gate first; a non-missing ERROR blocks the transition.
func specApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <name>",
		Short: "审批 spec（捕获 baseline）",
		Args:  cobra.ExactArgs(1),
		Long: `审批 spec + 捕获漂移 baseline。

lint gate：ERROR（no-ears / placeholder-ac / tasks 环/悬挂/自引用）会拦截审批。
missing-* 警告不算 ERROR，不拦截。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			meta, err := eng.Approve(args[0])
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"approved spec %q; baseline captured for drift detection\n", meta.Name)
			return nil
		},
	}
}

// specStartCmd transitions APPROVED → IMPLEMENTING. Pure bookkeeping.
func specStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "标记开始实现（APPROVED → IMPLEMENTING）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			meta, err := eng.Start(args[0])
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"started implementation of spec %q (phase: %s)\n", meta.Name, meta.Phase)
			fmt.Fprintf(cmd.OutOrStdout(),
				"next: free-kiro spec complete %s  (once all tasks are done)\n", meta.Name)
			return nil
		},
	}
}

// specCompleteCmd transitions IMPLEMENTING → DONE.
func specCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <name>",
		Short: "标记 spec 完成（IMPLEMENTING → DONE）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			meta, err := eng.Complete(args[0])
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"completed spec %q (phase: %s)\n", meta.Name, meta.Phase)
			return nil
		},
	}
}

// specSyncCmd re-baselines the spec to its current document state.
// Useful after the author has legitimately edited AC / task lists and
// wants to silence the drift warning emitted by `spec status`.
func specSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <name>",
		Short: "重新基线化（消除漂移）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			if _, err := eng.Sync(args[0]); err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"re-baselined spec %q (drift comparison reset to current docs)\n", args[0])
			return nil
		},
	}
}

// specStatusCmd prints a JSON snapshot of the spec's lifecycle position
// plus drift signals. Designed to be piped to `jq` by IDE hooks.
func specStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <name>",
		Short: "查看 spec 状态 + 漂移（JSON）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			status, err := eng.Status(args[0])
			if err != nil {
				return exitWithError(err)
			}
			out, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

// specNextCmd prints the oracle's recommendation for the next action.
func specNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next <name>",
		Short: "预言机：返回下一步动作 + 命令（JSON）",
		Args:  cobra.ExactArgs(1),
		Long: `返回当前 phase + lint 状态下的推荐命令。每个回合开头跑一次防断线：

  $ free-kiro spec next my-spec | jq -r .command
  free-kiro spec generate my-spec --phase all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			action, err := eng.NextAction(args[0])
			if err != nil {
				return exitWithError(err)
			}
			out, _ := json.MarshalIndent(action, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

// specAnalyzeCmd runs the advisory consistency analysis (vague language,
// duplicate ACs, requirements↔tasks traceability). Never blocks.
func specAnalyzeCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "analyze <name>",
		Short: "advisory 一致性分析（vague / 重复 AC / 可追溯性）",
		Args:  cobra.ExactArgs(1),
		Long: `advisory 一致性分析，**永不拦截状态机**。
与 lint 的区别：lint 管形状（门禁），analyze 管质感（人审助手）。

检查：
  - vague-language       模糊措辞（etc / and/or / maybe / user-friendly …）
  - duplicate-acceptance-criteria  重复 EARS 行
  - uncovered-acceptance-criteria 有 AC 但无 tasks
  - tasks-without-requirements     有 tasks 但无 AC`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			findings := eng.Analyze(args[0])
			if asJSON {
				out, _ := json.MarshalIndent(findings, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(out))
				return nil
			}
			if len(findings) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: no consistency findings\n", args[0])
				return nil
			}
			// Severity histogram.
			hist := map[string]int{}
			for _, f := range findings {
				hist[f.Severity]++
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %d finding(s)\n", args[0], len(findings))
			for _, f := range findings {
				loc := ""
				if f.Location != "" {
					loc = " [" + f.Location + "]"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  (%s) %s%s: %s\n",
					f.Severity, f.Code, loc, f.Message)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "emit findings as JSON")
	return c
}