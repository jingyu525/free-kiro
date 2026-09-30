package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/hooks"
	"github.com/jingyu525/free-kiro/internal/models"
)

// hookCmdFactory replaces the Wave 1 stub with three subcommands:
// list, add, run. free-kiro stays a passive planning layer — it never
// fires events itself; the IDE's hook system calls `hook run` when an
// event fires.
func hookCmdFactory() *cobra.Command {
	c := &cobra.Command{
		Use:   "hook",
		Short: "管理事件驱动 hook（list / add / run）",
		Long: `hook 文档是 JSON 文件，位于 .kiro/hooks/<id>.json。
每个 hook 声明一个事件 + 过滤 + 动作（shell 命令或 agent prompt）。
free-kiro 写出的 hook 同时兼容自家 flat shape 与 Kiro 官方 v1 信封，
可直接被官方 Kiro IDE 加载。

事件名（兼容两套命名）：
  free-kiro  : file.save / file.create / file.delete / prompt.submit / task.run / manual
  Kiro 官方  : SessionStart / UserPromptSubmit / PreToolUse / PostToolUse / PostFileSave / PostFileCreate / ...

定位：free-kiro 是被动的规划层——它从不主动触发 hook，只在被调用
` + "`hook run <event>`" + ` 时执行匹配的动作。`,
	}
	c.AddCommand(hookListCmd())
	c.AddCommand(hookAddCmd())
	c.AddCommand(hookRunCmdFlag())
	return c
}

func hookListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出全部 hook",
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := newHookRegistry()
			if err != nil {
				return err
			}
			hs, err := reg.LoadAll()
			if err != nil {
				return exitWithError(err)
			}
			if len(hs) == 0 {
				writeOutln(cmd.OutOrStdout(), "(no hooks)")
				return nil
			}
			writeOut(cmd.OutOrStdout(), "%-20s %-16s %-10s %-7s %-5s ENABLED\n",
				"ID", "EVENT", "FILTER", "TYPE", "TO")
			for _, h := range hs {
				filt := h.Glob
				if filt == "" {
					filt = "*"
				}
				if h.IsRegex {
					filt += " (re)"
				}
				to := "-"
				if h.Timeout != nil {
					if *h.Timeout == 0 {
						to = "off"
					} else {
						to = fmt.Sprintf("%ds", *h.Timeout)
					}
				}
				writeOut(cmd.OutOrStdout(), "%-20s %-16s %-10s %-7s %-5s %v\n",
					h.ID, h.Event, filt, h.ActionType, to, h.Enabled)
			}
			return nil
		},
	}
}

func hookAddCmd() *cobra.Command {
	var (
		event      string
		glob       string
		actionType string
		action     string
		desc       string
		timeout    int
		disabled   bool
	)
	c := &cobra.Command{
		Use:   "add --id <id> --event <event> --action-type <shell|agent> --action <cmd|prompt>",
		Short: "新增一个 hook（自动写出 Kiro v1 信封）",
		Args:  cobra.NoArgs,
		Long: `向 .kiro/hooks/<id>.json 写入一个新 hook，格式为 Kiro v1 信封
（可被官方 Kiro IDE 直接加载）。

必填：
  --id <id>                  hook 唯一名
  --event <event>            触发事件（file.save / PreToolUse / ...）
  --action-type <type>       shell 或 agent
  --action <cmd|prompt>      shell 命令 或 agent prompt

可选：
  --glob <pattern>           路径 glob 过滤
  --description <text>       描述
  --timeout <seconds>        shell 命令超时（0 = 禁用超时）
  --disabled                 写入但默认 enabled=false`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			idFlag, _ := cmd.Flags().GetString("id")
			if idFlag == "" {
				return exitWithError(ferrors.NewUsage("hook.add", "--id is required"))
			}
			to := timeout
			h := &models.Hook{
				ID:          idFlag,
				Event:       event,
				Glob:        glob,
				ActionType:  actionType,
				Action:      action,
				Description: desc,
				Enabled:     !disabled,
			}
			if cmd.Flags().Changed("timeout") {
				h.Timeout = &to
			}
			reg, err := newHookRegistry()
			if err != nil {
				return err
			}
			path, err := reg.Add(h)
			if err != nil {
				return exitWithError(err)
			}
			writeOut(cmd.OutOrStdout(), "added hook %q -> %s\n", h.ID, path)
			return nil
		},
	}
	c.Flags().String("id", "", "hook id (required)")
	c.Flags().StringVar(&event, "event", "", "event name (file.save / PostFileSave / ...)")
	c.Flags().StringVar(&glob, "glob", "", "path glob filter")
	c.Flags().StringVar(&actionType, "action-type", "", "shell or agent")
	c.Flags().StringVar(&action, "action", "", "shell command or agent prompt")
	c.Flags().StringVar(&desc, "description", "", "human-readable description")
	c.Flags().IntVar(&timeout, "timeout", 0, "command timeout seconds (0 = disabled)")
	c.Flags().BoolVar(&disabled, "disabled", false, "write as enabled=false")
	return c
}

func hookRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <event>",
		Short: "触发匹配 (event, --file) 的全部 hook",
		Args:  cobra.ExactArgs(1),
		Long: `执行匹配 <event>（可选 --file）的所有 hook。shell 动作通过 STDIN 收
event context JSON；agent 动作若无 agent_fn 则返回占位符。

通常由 IDE 的 hook 系统调用；agent 直接运行也可用于本地调试。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			reg, err := newHookRegistry()
			if err != nil {
				return err
			}
			results, err := reg.Dispatch(cmd.Context(), args[0], file, nil)
			if err != nil {
				return exitWithError(err)
			}
			if len(results) == 0 {
				writeOut(cmd.OutOrStdout(),
					"(no hooks matched event=%q file=%q)\n", args[0], file)
				return nil
			}
			failed := 0
			for _, r := range results {
				status := "ok"
				if !r.OK {
					status = "FAIL"
					failed++
				}
				writeOut(cmd.OutOrStdout(), "[%s] %s: %s\n",
					status, r.ID, trimTrailing(r.Output))
				if r.Error != "" {
					writeOut(cmd.OutOrStdout(), "    error: %s\n", r.Error)
				}
			}
			if failed > 0 {
				return exitWithError(ferrors.New("hook.run", fmt.Sprintf("%d hook(s) failed", failed)))
			}
			return nil
		},
	}
}

func hookRunCmdFlag() *cobra.Command {
	c := hookRunCmd()
	c.Flags().String("file", "", "path of the file being worked on (matches globs/matcher)")
	return c
}

// trimTrailing strips a trailing newline for prettier CLI output.
func trimTrailing(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func newHookRegistry() (*hooks.Registry, error) {
	holder, err := loadEngine()
	if err != nil {
		return nil, err
	}
	return hooks.NewRegistry(holder.ws), nil
}
