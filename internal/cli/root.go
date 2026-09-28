// Package cli assembles the cobra command tree for free-kiro.
//
// Subcommand mirrors Kiro's mental model:
//
//	free-kiro init                    bootstrap .kiro workspace + sample steering
//	free-kiro spec {new,generate,quick,show,list,approve,status,next,sync,start,complete,analyze}
//	free-kiro steering {list,show,context}
//	free-kiro task list               parallel-wave view of tasks.md
//	free-kiro hook {list,add,run}     event-driven automations
//	free-kiro lint [name]             offline spec-quality checks (exit 1 on ERROR)
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// rootCmd is the base command. Subcommands attach themselves in init() or
// via AddCommand from each command file.
var rootCmd = &cobra.Command{
	Use:   "free-kiro",
	Short: "Spec-driven development workflow engine (Kiro core clone, in Go)",
	Long: `free-kiro — 规约式开发工作流引擎

免费的 Kiro Spec 工作流克隆版，把"先想清楚 → 再动手"做成强约束门禁。
面向 Claude Code / CodeBuddy / Cursor / Continue 等 AI coding 工具，让
可能性空间收敛为有限状态机，而不是发散失控。

子命令：
  init       初始化 .kiro 工作区 + 写入示例 steering
  spec      管理 spec（new / generate / approve / start / complete …）
  steering  查看项目约定文档（自动注入生成上下文）
  task      查看 tasks.md 的并行 wave 视图
  hook      管理事件驱动 hook（写出的 JSON 兼容 Kiro v1 信封）
  lint      离线质量门禁（exit 1 = ERROR，可被 IDE PreToolUse 拦截）`,
	SilenceUsage:  true, // don't dump help on engine errors
	SilenceErrors: true, // we print errors ourselves via Execute
}

// Execute runs the root command and returns an error whose Error() string
// is the exit message. The exit code is derived from ferrors.ExitCode.
//
// Note: cobra does not propagate errors when SilenceUsage/SilenceErrors are
// set, so we run Execute() and translate cobra's int return into our own
// error type for a single, predictable contract.
func Execute() error {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	if err := rootCmd.Execute(); err != nil {
		// Cobra already printed the error to SetErr; surface it for the
		// caller to convert to a non-zero exit.
		fmt.Fprintln(os.Stderr, "error:", err)
		return err
	}
	return nil
}

// exitWithError is a helper for subcommands that want to raise an engine
// error and have cobra print it. Usage:
//
//	return exitWithError(ferrors.New("foo", "bar"))
func exitWithError(err error) error {
	if err == nil {
		return nil
	}
	// Wrap as KiroError if it isn't already typed.
	var k *ferrors.KiroError
	if errors.As(err, &k) {
		return err
	}
	return ferrors.Wrap("cli", err, err.Error())
}

// init wires the subcommands onto rootCmd. Each command file (init.go,
// spec.go, etc.) registers itself here.
func init() {
	rootCmd.AddCommand(initCmd)
	initSpecSubcommands(specCmd)
	rootCmd.AddCommand(specCmd)
	rootCmd.AddCommand(steeringCmdFactory())
	rootCmd.AddCommand(taskCmdFactory())
	rootCmd.AddCommand(hookCmdFactory())
	rootCmd.AddCommand(lintCmdFactory())
	rootCmd.AddCommand(doctorCmdFactory())
	rootCmd.AddCommand(reportCmdFactory())
	rootCmd.AddCommand(serveCmdFactory())
}