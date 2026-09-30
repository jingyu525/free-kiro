package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
)

// taskCmd prints the parallel-wave view of tasks.md. Replaces the stub
// left over from Wave 1. The parent has no RunE so cobra prints help
// with the subcommand list when invoked bare.
func taskCmdFactory() *cobra.Command {
	c := &cobra.Command{
		Use:   "task",
		Short: "查看 tasks.md 的并行 wave 视图",
		Long: `解析任务的依赖图，按拓扑分层输出并行 wave。
Wave 1 内的任务没有相互依赖，可并发执行；Wave 2 依赖 Wave 1 的输出，依此类推。`,
	}
	c.AddCommand(taskListCmd())
	return c
}

func taskListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <spec>",
		Short: "按 wave 列出任务",
		Args:  cobra.ExactArgs(1),
		Long: `  $ free-kiro task list my-spec

输出示例：

  Wave 1:
    [ ] #1 Set up module layout  (deps: -)
    [ ] #2 Bootstrap dependencies (deps: -)
  Wave 2:
    [ ] #3 Implement core [deps: #1]  (deps: #1)
  Wave 3:
    [ ] #4 Polish [deps: #2,#3]  (deps: #2,#3)

  tasks: 1/4 done, 3 wave(s)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			specDir := eng.WS().SpecDir(args[0])
			tasksPath := filepath.Join(specDir, "tasks.md")
			data, err := os.ReadFile(tasksPath)
			if err != nil {
				return exitWithError(ferrors.NewUsage("task.list", fmt.Sprintf(
					"tasks.md not generated yet (run `free-kiro spec generate %s --phase tasks`)",
					args[0])))
			}
			tasks := taskgraph.ParseTasks(string(data))
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no tasks parsed)")
				return nil
			}
			waves, cycleErr := safeWaves(tasks)
			if cycleErr != nil {
				fmt.Fprintln(cmd.OutOrStdout(), cycleErr)
				return exitWithError(cycleErr)
			}
			printWaves(cmd, waves, args[0])
			return nil
		},
	}
}

// safeWaves wraps taskgraph.ExecutionWaves so the CLI can report cycles
// gracefully instead of returning nil. Cycles surface as a typed
// TaskGraphError so main.go's ExitCode maps them to exit 1 (the existing
// IDE hook contract for unparseable tasks.md).
func safeWaves(tasks []models.Task) ([][]models.Task, error) {
	if cycle := taskgraph.DetectCycle(tasks); cycle != nil {
		return nil, ferrors.NewTaskGraphError("task.list", "dependency cycle detected: "+formatCycle(cycle))
	}
	return taskgraph.ExecutionWaves(tasks), nil
}

func formatCycle(cycle []int) string {
	out := ""
	for i, id := range cycle {
		if i > 0 {
			out += " -> "
		}
		out += fmt.Sprintf("#%d", id)
	}
	return out
}

func printWaves(cmd *cobra.Command, waves [][]models.Task, specName string) {
	for i, wave := range waves {
		fmt.Fprintf(cmd.OutOrStdout(), "Wave %d:\n", i+1)
		for _, t := range wave {
			mark := " "
			if t.Done {
				mark = "x"
			}
			deps := "-"
			if len(t.Deps) > 0 {
				deps = ""
				for j, d := range t.Deps {
					if j > 0 {
						deps += ","
					}
					deps += fmt.Sprintf("#%d", d)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] #%d %s  (deps: %s)\n",
				mark, t.ID, t.Title, deps)
		}
	}
	s := taskgraph.Summary(tasksOf(waves))
	fmt.Fprintf(cmd.OutOrStdout(), "\n%s: %d/%d done, %d wave(s)\n",
		specName, s.Done, s.Total, s.Waves)
}

// tasksOf flattens a wave view back into a task list (for Summary).
func tasksOf(waves [][]models.Task) []models.Task {
	var out []models.Task
	for _, w := range waves {
		out = append(out, w...)
	}
	return out
}