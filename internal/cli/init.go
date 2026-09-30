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
//
// v0.8.0+: each configured IDE additionally gets the project-level file it
// actually loads for agent instructions
// (Claude Code → CLAUDE.md, Cursor → .cursorrules + .cursor/rules/free-kiro.md,
// Continue → .continuerules + .continue/rules/free-kiro.md, OpenCode /
// CodeBuddy → AGENTS.md). `.kiro/AGENTS.md` is still written as the
// workspace-level steering doc.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 .kiro 工作区 + 可选配置 IDE hook",
	Long: `在指定目录初始化 free-kiro 工作区：

  .kiro/
  ├── settings.json    生成器配置
  ├── specs/           每个 spec 一个子目录
  ├── steering/        项目级约定文档（product/structure/tech）
  ├── hooks/           事件驱动自动化
  └── AGENTS.md        workspace 级 agent 指令（steering store 加载）

项目根额外生成 IDE-specific 指令文件（取决于 --ide）：
  CLAUDE.md                  Claude Code 加载
  .cursorrules               Cursor 加载（legacy 兼容）
  .cursor/rules/free-kiro.md Cursor 模块化规则
  .continuerules             Continue 加载（legacy 兼容）
  .continue/rules/free-kiro.md Continue 模块化规则
  AGENTS.md                  OpenCode / CodeBuddy 加载

--ide <target>  配置 AI coding 工具的 hook（默认: auto）：
  auto            检测本机已装的 IDE（Claude Code / CodeBuddy / Cursor / Continue / OpenCode）
  claude-code     写入 ~/.claude/settings.json
  codebuddy       写入 ~/.codebuddy/settings.json
  cursor          写入 ~/.cursor/settings.json
  continue        写入 ~/.continue/config.json
  opencode        写入 ~/.opencode/settings.json
  none            跳过 IDE 配置

--overwrite-instructions   覆盖已存在的所有 IDE 指令文件（默认: false）
--overwrite-agents         DEPRECATED: 重命名为 --overwrite-instructions，
                          仍接受同一行为，传一次后会 stderr 警告。

若目录已存在 .kiro，重复运行是幂等的（不会覆盖已有 steering / settings）。
`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")
		ideTarget, _ := cmd.Flags().GetString("ide")
		lang, _ := cmd.Flags().GetString("lang")

		// Resolve overwrite flag with deprecation forwarding.
		overwriteInstructions, _ := cmd.Flags().GetBool("overwrite-instructions")
		overwriteAgentsLegacy, _ := cmd.Flags().GetBool("overwrite-agents")
		if cmd.Flags().Changed("overwrite-agents") {
			writeOut(os.Stderr,
				"⚠️  --overwrite-agents 已废弃，请改用 --overwrite-instructions（同一行为，下个 minor 移除）\n")
			overwriteInstructions = overwriteInstructions || overwriteAgentsLegacy
		}

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
		writeOut(cmd.OutOrStdout(), "已初始化 free-kiro 工作区于 %s\n", ws.KiroDir())
		writeOutln(cmd.OutOrStdout(), "已写入示例 steering：product.md、structure.md、tech.md")

		// IDE integration (v0.2.0+).
		if err := runIdeInit(cmd, ideTarget, ws.Root(), lang, overwriteInstructions); err != nil {
			// IDE failure is non-fatal — the workspace is still usable.
			writeOut(cmd.OutOrStdout(),
				"⚠️  IDE 配置失败：%v\n   workspace 已就绪，可后续运行 `free-kiro doctor` 排查\n",
				err)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().String("path", ".", "目标项目根目录（默认：当前目录）")
	initCmd.Flags().String("ide", "auto", "IDE 集成目标: auto | claude-code | codebuddy | cursor | continue | opencode | none")
	initCmd.Flags().String("lang", "zh", "AGENTS.md 语言: zh | en（默认探测 LANG）")
	initCmd.Flags().Bool("overwrite-instructions", false, "覆盖已存在的 IDE-specific 指令文件（CLAUDE.md / .cursorrules / 等）")
	initCmd.Flags().Bool("overwrite-agents", false, "DEPRECATED: 重命名为 --overwrite-instructions，传一次后 stderr 警告")
}

// runIdeInit resolves the --ide flag and writes hooks + agent instruction
// files. The two write streams are:
//  1. `.kiro/AGENTS.md` — workspace-level steering doc (always written,
//     loaded by the steering store regardless of IDE choice).
//  2. Per-IDE project-root instruction files (CLAUDE.md, .cursorrules,
//     AGENTS.md, etc.) — written only when an IDE id is resolved; each
//     IDE gets the file its own agent loader reads.
//
// All errors bubble up so the caller can decide whether to make them fatal.
func runIdeInit(cmd *cobra.Command, target, workspaceRoot, lang string, overwrite bool) error {
	// Write AGENTS.md FIRST — always, even when --ide none skips hook
	// config. The steering store depends on this file regardless of
	// which IDE is in use. Per-IDE instruction files (CLAUDE.md,
	// .cursorrules, …) are written further down via
	// WriteAgentInstructions.
	agentsPath, err := ide.WriteAgentsMD(workspaceRoot, lang, overwrite)
	if err != nil {
		return ferrors.Wrap("init.writeAgents", err, "write AGENTS.md")
	}
	writeOut(cmd.OutOrStdout(), "已写入 %s\n", agentsPath)

	// Resolve "auto" / "none" / explicit.
	resolved, err := resolveIdeTarget(target)
	if err != nil {
		return err
	}
	if resolved == nil {
		writeOutln(cmd.OutOrStdout(), "已跳过 IDE 配置（--ide none）")
		return nil
	}

	// Per-IDE project-root instruction files (e.g. CLAUDE.md,
	// .cursorrules). Dedupe happens inside WriteAgentInstructions.
	written, err := ide.WriteAgentInstructions(workspaceRoot, lang, overwrite, resolved)
	if err != nil {
		return ferrors.Wrap("init.writeInstructions", err, "write IDE instruction files")
	}
	for _, rel := range written {
		writeOut(cmd.OutOrStdout(), "✓ 已写入 %s\n", rel)
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
		writeOut(cmd.OutOrStdout(), "✓ %s\n", note)
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
		writeOut(os.Stderr,
			"⚠️  未检测到 Claude Code / CodeBuddy / Cursor / Continue / OpenCode 目录；跳过 IDE 配置\n"+
				"   安装其中一个并重跑 `free-kiro init --ide auto`，或显式指定 --ide <name>\n")
		return nil, nil
	}
	return present, nil
}

func sampleSteering() map[string]string {
	return map[string]string{
		"product.md":   "---\nmode: always\ndescription: 这个产品的目标、核心能力与边界\n---\n# Product\n\n用 1-3 段说明产品的目标和必须提供的核心能力。\n本文档作为 always 上下文注入到每次生成中。\n",
		"structure.md": "---\nmode: always\ndescription: 代码组织与架构约定\n---\n# Structure\n\n- 模块保持小且单一职责。\n- IO 隔离在显式适配器后面。\n- 除非有明确理由，优先使用标准库而非新增依赖。\n",
		"tech.md":      "---\nmode: auto\ndescription: Go 项目技术栈与开发规范\n---\n# Tech\n\n- 语言：Go 1.22+。\n- 不引入非必要第三方依赖。\n- 测试用标准库 testing 包，断言用 testify。\n",
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
