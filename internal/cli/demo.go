package cli

// demoCmd drives the "5 分钟端到端 demo" onboarding path.
//
// The command assumes it is run from the free-kiro repository root (so
// `examples/todo-app/` is reachable). When that is not the case it exits
// 3 with a one-line Chinese stderr message — the user can self-correct
// without reading docs.
//
// Companion spec: .kiro/specs/top1-demo-onboarding/.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/ide"
)

// demoDedupeWindow is how recent a `.kiro/.demostart` marker must be for
// the demo command to consider itself "already running". Older markers
// are silently overwritten.
const demoDedupeWindow = 30 * time.Second

// demoMarkerRelPath is the path (relative to cwd) where the run marker
// is written. Lives inside the demo workspace, not next to source.
const demoMarkerRelPath = ".kiro/.demostart"

// demoExampleRelPath is the canonical location of the example spec. The
// demo command requires this path to exist (it's our repo-root check).
const demoExampleRelPath = "examples/todo-app"

// demoCmd is the cobra command exposed at the root of the CLI.
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "5 分钟端到端 demo — 打印 onboarding 摘要并启动样例",
	Long: `demo 是 free-kiro 的一键 onboarding 入口。

它会:
  1. 校验当前目录必须是 free-kiro 仓库根（含 examples/todo-app/）；
  2. 写一个时间戳 marker 到 .kiro/.demostart（30 秒内重复执行会去重）；
  3. 打印 5 行中文步骤摘要，告诉用户下一步该跑哪些命令；
  4. 在 --ide <name> 不为 none 时，额外打印对应 IDE 的 hook 配置片段。

退出码：
  0  成功
  3  cwd 不在仓库根 / --ide 名称非法（UsageError）

不修改用户项目，不写任何 .kiro 业务文件之外的内容。
`,
	RunE: runDemo,
}

func init() {
	demoCmd.Flags().Bool("no-color", false, "禁止 ANSI 颜色输出（适合 CI / tee）")
	demoCmd.Flags().String("ide", "none", "demo 完后要展示的 IDE hook 片段：claude-code | codebuddy | cursor | continue | opencode | none")
	// Demo prints one-line guidance on UsageError (wrong cwd / unknown IDE);
	// we want that guidance visible in stderr for both human users and CI logs,
	// so unset SilenceErrors/SilenceUsage that demoCmd inherits from rootCmd.
	demoCmd.SilenceErrors = false
	demoCmd.SilenceUsage = false
}

func runDemo(cmd *cobra.Command, _ []string) error {
	noColor, _ := cmd.Flags().GetBool("no-color")
	ideName, _ := cmd.Flags().GetString("ide")

	// 1) cwd 校验：必须在仓库根（examples/todo-app 可达）。
	cwd, err := os.Getwd()
	if err != nil {
		return ferrors.Wrap("demo.cwd", err, "resolve cwd")
	}
	if _, err := os.Stat(filepath.Join(cwd, demoExampleRelPath)); err != nil {
		return ferrors.NewUsage("demo",
			"未找到 "+demoExampleRelPath+"/；请先 `cd` 到 free-kiro 仓库根目录再跑 `free-kiro demo`")
	}

	// 2) 解析 --ide（Parse 把空 / "auto" / "none" 映射为 ID=""）。
	resolvedIDE, err := ide.Parse(ideName)
	if err != nil {
		return err
	}

	// 3) 写 marker；30 秒内重复执行视为已运行，跳过覆盖并打 warning。
	markerPath := filepath.Join(cwd, demoMarkerRelPath)
	recent, prevStamp := demoMarkerIsRecent(markerPath)
	if recent {
		writeOutln(cmd.OutOrStdout(),
			fmt.Sprintf("⚠️  已有最近的 demo marker（%s）；30 秒内重复执行将被忽略", prevStamp))
	} else if err := writeDemoMarker(markerPath); err != nil {
		return err
	}

	// 4) 打印 5 步 onboarding 摘要。
	printOnboarding(cmd.OutOrStdout(), cwd, noColor)

	// 5) 可选打印 hook 片段。
	if resolvedIDE != "" {
		writeOutln(cmd.OutOrStdout(), "")
		writeOutln(cmd.OutOrStdout(), "— — — — — 适用于 "+resolvedIDE.String()+" 的 hook 片段 — — — — —")
		printIDESnippet(cmd.OutOrStdout(), resolvedIDE)
	}

	// 提示 next:
	writeOutln(cmd.OutOrStdout(), "")
	writeOutln(cmd.OutOrStdout(), "next: 在另一终端跑 `free-kiro serve` 启动看板，或 `free-kiro watch --preset reactive` 监听变化")
	return nil
}

// printOnboarding writes the 5-line onboarding summary. We avoid any
// ANSI when `noColor` is true so CI logs / `tee` output stay clean.
func printOnboarding(w io.Writer, cwd string, noColor bool) {
	marker := "✓"
	head := func(s string) string {
		if noColor {
			return s
		}
		return "\033[1;36m" + s + "\033[0m"
	}
	writeOutln(w, head("free-kiro 5 分钟端到端 demo — 把这些跑一遍："))
	writeOutln(w, marker+" 1. cd examples/todo-app && ../../dist/free-kiro_*/free-kiro serve    "+head("→")+" 启动看板")
	writeOutln(w, marker+" 2. 浏览器打开 http://127.0.0.1:7373                                         "+head("→")+" 看 add-task-priority 状态")
	writeOutln(w, marker+" 3. cd .. && free-kiro task list examples/todo-app                            "+head("→")+" 看 4 wave 拓扑")
	writeOutln(w, marker+" 4. cd examples/todo-app && free-kiro spec status add-task-priority --human   "+head("→")+" 看 drift 状态")
	writeOutln(w, marker+" 5. git checkout .kiro/specs/add-task-priority/requirements.md                "+head("→")+" 撤回编辑，恢复 drift: none")
	writeOutln(w, "当前 cwd："+cwd)
}

// printIDESnippet renders a free-kiro hook fragment the user can paste
// into their IDE's settings.json. Mirrors the canonical hook set written
// by `free-kiro init --ide <name>` so the demo message matches what
// they'd actually get on init.
func printIDESnippet(w io.Writer, id ide.ID) {
	switch id {
	case ide.ClaudeCode:
		writeOutln(w, "  ~/.claude/settings.json:")
		writeOutln(w, `  {`)
		writeOutln(w, `    "hooks": {`)
		writeOutln(w, `      "PreToolUse": [{"matcher": "Edit|Write", "hooks": [`)
		writeOutln(w, `        {"type": "command", "command": "# free-kiro-managed: free-kiro lint || exit 2"}`)
		writeOutln(w, `      ]}]`)
		writeOutln(w, `    }`)
		writeOutln(w, `  }`)
	case ide.CodeBuddy:
		writeOutln(w, "  ~/.codebuddy/settings.json（结构同 Claude Code）：")
		writeOutln(w, `  { "hooks": { "PreToolUse": [{"matcher": "Edit|Write", "hooks": [`)
		writeOutln(w, `    {"type": "command", "command": "# free-kiro-managed: free-kiro lint || exit 2"}`)
		writeOutln(w, `  ]}]} }`)
	case ide.Cursor:
		writeOutln(w, "  ~/.cursor/settings.json（结构同 Claude Code）：")
		writeOutln(w, `  { "hooks": { "PreToolUse": [{"matcher": "Edit|Write", "hooks": [`)
		writeOutln(w, `    {"type": "command", "command": "# free-kiro-managed: free-kiro lint || exit 2"}`)
		writeOutln(w, `  ]}]} }`)
	case ide.Continue:
		writeOutln(w, "  ~/.continue/config.json（结构同 Claude Code envelope）：")
		writeOutln(w, `  { "hooks": { "PreToolUse": [{"matcher": "Edit|Write", "hooks": [`)
		writeOutln(w, `    {"type": "command", "command": "# free-kiro-managed: free-kiro lint || exit 2"}`)
		writeOutln(w, `  ]}]} }`)
	case ide.OpenCode:
		writeOutln(w, "  ~/.opencode/settings.json（结构同 Claude Code envelope）：")
		writeOutln(w, `  { "hooks": { "PreToolUse": [{"matcher": "Edit|Write", "hooks": [`)
		writeOutln(w, `    {"type": "command", "command": "# free-kiro-managed: free-kiro lint || exit 2"}`)
		writeOutln(w, `  ]}]} }`)
	}
}

// demoMarkerIsRecent reads the marker file at `path` and reports whether
// it was written within demoDedupeWindow. Returns (recent, original
// timestamp) so callers can render a "already running" hint.
//
// Missing / unreadable marker → (false, "") so the caller proceeds to
// overwrite.
func demoMarkerIsRecent(path string) (bool, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, ""
	}
	ts, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return false, ""
	}
	if time.Since(ts) < demoDedupeWindow {
		return true, ts.Format(time.RFC3339)
	}
	return false, ""
}

// writeDemoMarker writes the current RFC3339 timestamp into `path`. The
// parent dir (.kiro/) is created if missing; failures wrap a demo.write
// op so the error path stays typed (ExitCode → 2).
func writeDemoMarker(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ferrors.Wrap("demo.write", err, "mkdir "+filepath.Dir(path))
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	if err := os.WriteFile(path, []byte(ts), 0o644); err != nil {
		return ferrors.Wrap("demo.write", err, "write "+path)
	}
	return nil
}