package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
	"github.com/jingyu525/free-kiro/internal/visualize"
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

// specStatusCmd prints a snapshot of the spec's lifecycle position plus
// drift signals. Default output is a flat, human-readable block; pass
// --json to get the original nested JSON for piping into jq / hooks.
// --graph emits a Mermaid graph LR block instead.
func specStatusCmd() *cobra.Command {
	var asJSON bool
	var asGraph bool
	c := &cobra.Command{
		Use:   "status <name>",
		Short: "查看 spec 状态 + 漂移（默认人类可读，--json/--graph 可选）",
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
			switch {
			case asJSON:
				out, err := json.MarshalIndent(status, "", "  ")
				if err != nil {
					return exitWithError(err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(out))
			case asGraph:
				renderMermaidStatus(cmd.OutOrStdout(), eng, args[0], status)
			default:
				renderHumanStatus(cmd.OutOrStdout(), args[0], status)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "emit nested JSON (for piping into jq / IDE hooks)")
	c.Flags().BoolVar(&asGraph, "graph", false, "emit Mermaid graph LR block (paste into GitHub PR)")
	return c
}

// renderHumanStatus flattens the spec status into a top-level view.
// Drift signals are surfaced prominently at the top — they're the
// reason most users run `status`.
func renderHumanStatus(w io.Writer, name string, s map[string]any) {
	fmt.Fprintf(w, "%s\n", name)
	fmt.Fprintf(w, "  phase:      %s\n", s["phase"])
	fmt.Fprintf(w, "  workflow:   %s\n", s["workflow"])
	fmt.Fprintf(w, "  spec_type:  %s\n", s["spec_type"])
	if approved, _ := s["approved"].(bool); approved {
		fmt.Fprintln(w, "  approved:   yes")
	} else {
		fmt.Fprintln(w, "  approved:   no")
	}

	// Tasks line (handy quick view).
	if t, ok := s["tasks"].(map[string]any); ok {
		done, _ := t["done"].(int)
		total, _ := t["total"].(int)
		waves, _ := t["waves"].(int)
		fmt.Fprintf(w, "  tasks:      %d/%d done, %d wave(s)\n", done, total, waves)
	}

	// Drift — promoted to the top because it's the action item.
	if drift, ok := s["drift"].([]any); ok && len(drift) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  DRIFT (baseline → current):")
		for _, item := range drift {
			d, ok := item.(map[string]any)
			if !ok {
				continue
			}
			key, _ := d["key"].(string)
			base, _ := d["baseline"].(int)
			cur, _ := d["current"].(int)
			delta, _ := d["delta"].(int)
			sign := " "
			if delta > 0 {
				sign = "+"
			} else if delta < 0 {
				sign = "-"
			}
			fmt.Fprintf(w, "    %s: %d → %d  (%s%d)\n",
				key, base, cur, sign, absDelta(delta))
		}
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  fix: either revert the change, or run `free-kiro spec sync <name>` to accept it as the new baseline")
	} else {
		fmt.Fprintln(w, "  drift:      none")
	}

	// Baseline / current snapshot at the bottom (for context).
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  baseline:")
	if b, ok := s["baseline"].(map[string]int); ok {
		for k, v := range b {
			fmt.Fprintf(w, "    %s: %d\n", k, v)
		}
	}
	fmt.Fprintln(w, "  current:")
	if c, ok := s["current"].(map[string]int); ok {
		for k, v := range c {
			fmt.Fprintf(w, "    %s: %d\n", k, v)
		}
	}
}

func absDelta(d int) int {
	if d < 0 {
		return -d
	}
	return d
}

// renderMermaidStatus emits a Mermaid graph LR block for one spec.
// Falls back to a no-op graph (with a comment line) when the spec
// has no tasks.md yet.
func renderMermaidStatus(w io.Writer, eng *spec.Engine, name string, status map[string]any) {
	phase := models.Phase("")
	if p, ok := status["phase"].(string); ok {
		phase = models.Phase(p)
	}
	tasks, waves := loadTasksForMermaid(eng, name)
	if len(tasks) == 0 {
		fmt.Fprintln(w, "graph LR")
		fmt.Fprintf(w, "  spec_%s[\"%s<br/>phase: %s<br/>no tasks yet\"]\n",
			name, name, phase)
		return
	}
	visualize.RenderMermaidSpec(w, name, phase, tasks, waves)
}

// loadTasksForMermaid reads tasks.md and returns the wave grouping.
func loadTasksForMermaid(eng *spec.Engine, name string) ([]models.Task, [][]models.Task) {
	data, err := os.ReadFile(filepath.Join(eng.WS().SpecDir(name), "tasks.md"))
	if err != nil {
		return nil, nil
	}
	tasks := taskgraph.ParseTasks(string(data))
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks, taskgraph.ExecutionWaves(tasks)
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