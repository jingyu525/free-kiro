package cli

import (
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/watch"
)

// watchCmdFactory wires `free-kiro watch` — the IDE-style file watcher.
//
// Usage:
//
//	free-kiro watch                            # default: run lint on any .kiro/ change
//	free-kiro watch --command "free-kiro lint && echo ✓"
//	free-kiro watch --debounce 1s             # longer quiet period
//	free-kiro watch --root ./docs --command "make"
// presetCommands maps the --preset flag to a list of shell snippets.
// Presets are convenience groupings; --command (repeatable) overrides.
var presetCommands = map[string][]string{
	"default":  {"free-kiro lint"},
	"lint":     {"free-kiro lint"},
	"status":   {"free-kiro status --human"},
	"reactive": {"free-kiro lint", "free-kiro status --human"},
	"full": {
		"free-kiro lint",
		"free-kiro status --human",
		"free-kiro report",
	},
}

func watchCmdFactory() *cobra.Command {
	var (
		root     []string
		commands []string
		preset   string
		debounce time.Duration
		verbose  bool
	)
	c := &cobra.Command{
		Use:   "watch",
		Short: "监听 .kiro/ 变化，自动跑 lint（或自定义命令集）",
		Long: `监听 .kiro/ 目录下的文件变化，每次安静期（默认 500ms）后
执行一组 shell 命令（按顺序），每条命令独立运行。一条失败不影响其他。

预设：
  default / lint    只跑 lint
  status            只跑 status（人类可读）
  reactive          lint + status（推荐 IDE 终端用）
  full              lint + status + report

  $ free-kiro watch                          # = --preset default
  $ free-kiro watch --preset reactive        # = lint + status
  $ free-kiro watch --preset full            # = lint + status + report

自定义（可重复 --command 或忽略 --preset）：
  $ free-kiro watch --command "free-kiro lint"
  $ free-kiro watch --command "make" --command "go test"

其他：
  --debounce 1s      长安静期
  --verbose          看每个 fs event
  --root ./docs      自定义路径

退出 / 信号：SIGINT/SIGTERM 干净退出（Ctrl+C 不留 leak）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(),
				syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			// Resolve command list: --command overrides --preset (each
			// --command appends; preset is only used when no --command).
			cmds := commands
			if len(cmds) == 0 {
				if p, ok := presetCommands[preset]; ok {
					cmds = p
				} else {
					return exitWithError(fmt.Errorf(
						"unknown preset %q (valid: default, lint, status, reactive, full)",
						preset))
				}
			}
			opts := watch.Options{
				Roots:    root,
				Commands: cmds,
				Debounce: debounce,
				Verbose:  verbose,
			}
			w, err := watch.New(opts)
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"free-kiro watch: running %d command(s) on .kiro/ changes.\n", len(cmds))
			for _, c := range cmds {
				fmt.Fprintf(cmd.OutOrStdout(), "  → %s\n", c)
			}
			if err := w.Run(ctx); err != nil && ctx.Err() == nil {
				return exitWithError(err)
			}
			return nil
		},
	}
	c.Flags().StringSliceVar(&root, "root", []string{".kiro"},
		"directories to watch (repeatable; default: .kiro)")
	c.Flags().StringSliceVar(&commands, "command", nil,
		"shell snippet to run on each change (repeatable; overrides --preset)")
	c.Flags().StringVar(&preset, "preset", "default",
		"command preset when --command is not given (default|status|reactive|full)")
	c.Flags().DurationVar(&debounce, "debounce", 500*time.Millisecond,
		"quiet period after last change before triggering (default 500ms)")
	c.Flags().BoolVar(&verbose, "verbose", false, "log every fs event the watcher sees")
	return c
}