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
func watchCmdFactory() *cobra.Command {
	var (
		root     []string
		command  string
		debounce time.Duration
		verbose  bool
	)
	c := &cobra.Command{
		Use:   "watch",
		Short: "监听 .kiro/ 变化，自动跑 lint（或自定义命令）",
		Long: `监听 .kiro/ 目录下的文件变化，每次安静期（默认 500ms）后
执行指定的 shell 命令。内置默认命令是 free-kiro lint。

典型用法：
  $ free-kiro watch                          # 默认：每次保存自动 lint
  $ free-kiro watch --command "make check"  # 跑自定义脚本
  $ free-kiro watch --debounce 1s             # 1s 安静期
  $ free-kiro watch --verbose                 # 看每个 fs event（调试）

退出 / 信号：SIGINT/SIGTERM 干净退出（Ctrl+C 不留 leak）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(),
				syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			opts := watch.Options{
				Roots:    root,
				Command:  command,
				Debounce: debounce,
				Verbose:  verbose,
			}
			w, err := watch.New(opts)
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(),
				"free-kiro watch: running. Edit a file under .kiro/ to trigger.")
			if err := w.Run(ctx); err != nil && ctx.Err() == nil {
				return exitWithError(err)
			}
			return nil
		},
	}
	c.Flags().StringSliceVar(&root, "root", []string{".kiro"},
		"directories to watch (repeatable; default: .kiro)")
	c.Flags().StringVar(&command, "command", "free-kiro lint",
		"shell snippet to execute after each quiet period; $FILE = changed path")
	c.Flags().DurationVar(&debounce, "debounce", 500*time.Millisecond,
		"quiet period after last change before triggering (default 500ms)")
	c.Flags().BoolVar(&verbose, "verbose", false, "log every fs event the watcher sees")
	return c
}