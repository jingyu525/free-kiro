package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/ide"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// initCmd bootstraps a .kiro workspace in the given directory (default:
// current directory) and writes three sample steering documents:
//
//	.kiro/steering/product.md     (mode: always)
//	.kiro/steering/structure.md   (mode: always)
//	.kiro/steering/tech.md        (mode: auto)
//
// v0.2.0+: --ide auto|claude-code|codebuddy|none also writes the IDE's
// hook configuration and creates .kiro/AGENTS.md so the agent is
// immediately aware of free-kiro.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 .kiro 工作区 + 可选配置 IDE hook",
	Long: `在指定目录初始化 free-kiro 工作区：

  .kiro/
  ├── settings.json    生成器配置
  ├── specs/           每个 spec 一个子目录
  ├── steering/        项目级约定文档（product/structure/tech）
  ├── hooks/           事件驱动自动化
  └── AGENTS.md        写入给 agent 的总体指令（v0.2.0+）

--ide <target>  配置 AI coding 工具的 hook（默认: auto）：
  auto            检测本机已装的 IDE（Claude Code / CodeBuddy）
  claude-code     写入 ~/.claude/settings.json
  codebuddy       写入 ~/.codebuddy/settings.json
  none            跳过 IDE 配置

若目录已存在 .kiro，重复运行是幂等的（不会覆盖已有 steering / settings）。
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		ideTarget, _ := cmd.Flags().GetString("ide")
		lang, _ := cmd.Flags().GetString("lang")
		overwrite, _ := cmd.Flags().GetBool("overwrite-agents")

		ws := workspace.New(path)
		if err := ws.EnsureLayout(); err != nil {
			return exitWithError(err)
		}
		// Write sample steering docs only if missing.
		for name, content := range sampleSteering() {
			dest := ws.SteeringDir() + "/" + name
			if _, err := stat(dest); err == nil {
				continue
			}
			if err := writeFile(dest, content); err != nil {
				return exitWithError(err)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "已初始化 free-kiro 工作区于 %s\n", ws.KiroDir())
		fmt.Fprintln(cmd.OutOrStdout(), "已写入示例 steering：product.md、structure.md、tech.md")

		// IDE integration (v0.2.0+).
		if err := runIdeInit(cmd, ideTarget, ws.Root(), lang, overwrite); err != nil {
			// IDE failure is non-fatal — the workspace is still usable.
			fmt.Fprintf(cmd.OutOrStdout(),
				"⚠️  IDE 配置失败：%v\n   workspace 已就绪，可后续运行 `free-kiro doctor` 排查\n",
				err)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().String("path", ".", "目标项目根目录（默认：当前目录）")
	initCmd.Flags().String("ide", "auto", "IDE 集成目标: auto | claude-code | codebuddy | none")
	initCmd.Flags().String("lang", "zh", "AGENTS.md 语言: zh | en（默认探测 LANG）")
	initCmd.Flags().Bool("overwrite-agents", false, "覆盖已存在的 .kiro/AGENTS.md")
}

// runIdeInit resolves the --ide flag and writes hooks + AGENTS.md.
// All errors bubble up so the caller can decide whether to make them fatal.
func runIdeInit(cmd *cobra.Command, target, workspaceRoot, lang string, overwrite bool) error {
	// Write AGENTS.md FIRST — always, even when --ide none skips hook
	// config. The agent's onboarding depends on this file regardless
	// of which IDE is in use.
	agentsPath, err := ide.WriteAgentsMD(workspaceRoot, lang, overwrite)
	if err != nil {
		return ferrors.Wrap("init.writeAgents", err, "write AGENTS.md")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "已写入 %s\n", agentsPath)

	// Resolve "auto" / "none" / explicit.
	resolved, err := resolveIdeTarget(target)
	if err != nil {
		return err
	}
	if resolved == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "已跳过 IDE 配置（--ide none）")
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ferrors.Wrap("init.resolveHome", err, "resolve HOME")
	}

	// Then hooks for each resolved IDE.
	for _, target := range resolved {
		path, note, err := ide.InstallHooks(target, home)
		if err != nil {
			return ferrors.Wrap("init.installHooks", err, fmt.Sprintf("install hooks for %s", target))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ %s\n", note)
		_ = path
	}
	return nil
}

// resolveIdeTarget maps the --ide flag value to a list of concrete IDEs
// to configure. "none" → nil; "auto" → detected IDEs (filtered to those
// with existing config dirs); explicit ID → that single IDE.
func resolveIdeTarget(target string) ([]ide.ID, error) {
	parsed, err := ide.Parse(target)
	if err != nil {
		return nil, err
	}
	if parsed == "" && target != "auto" {
		// Parse("none") returns ("", nil); surface as "skip" explicitly.
		return nil, nil
	}
	if parsed != "" {
		return []ide.ID{parsed}, nil
	}
	// auto: detect installed IDEs.
	detected := ide.DetectAll("")
	var present []ide.ID
	for _, d := range detected {
		if d.DirExists {
			present = append(present, d.ID)
		}
	}
	if len(present) == 0 {
		fmt.Fprintf(os.Stderr,
			"⚠️  未检测到 Claude Code 或 CodeBuddy 目录；跳过 IDE 配置\n"+
				"   安装其中一个并重跑 `free-kiro init --ide auto`，或显式指定 --ide <name>\n")
		return nil, nil
	}
	return present, nil
}

func sampleSteering() map[string]string {
	return map[string]string{
		"product.md": "---\nmode: always\ndescription: 这个产品的目标、核心能力与边界\n---\n# Product\n\n用 1-3 段说明产品的目标和必须提供的核心能力。\n本文档作为 always 上下文注入到每次生成中。\n",
		"structure.md": "---\nmode: always\ndescription: 代码组织与架构约定\n---\n# Structure\n\n- 模块保持小且单一职责。\n- IO 隔离在显式适配器后面。\n- 除非有明确理由，优先使用标准库而非新增依赖。\n",
		"tech.md": "---\nmode: auto\ndescription: Go 项目技术栈与开发规范\n---\n# Tech\n\n- 语言：Go 1.22+。\n- 不引入非必要第三方依赖。\n- 测试用标准库 testing 包，断言用 testify。\n",
	}
}

// stat / writeFile are tiny indirection so the tests can stub them later
// without rewriting the command.
var (
	stat      = func(p string) (any, error) { return osStat(p) }
	writeFile = func(p, content string) error { return osWriteFile(p, content) }
)

// ensure filepath is referenced (used by AGENTS.md resolution later).
var _ = filepath.Join