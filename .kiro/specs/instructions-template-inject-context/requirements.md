# instructions-template-inject-context

把 `CLAUDE.md` / `.cursorrules` / `.cursor/rules/free-kiro.md` 顶部手动
加的"## 项目上下文（自动注入的 steering）"段固化进 init 模板
`internal/ide/templates/instructions_zh.md` 与 `instructions_en.md`，让
`free-kiro init --ide auto --overwrite-instructions` 重写这些 IDE 指令
文件时保留该段，不再被 init 模板覆盖丢失。

## User Stories

- As a Claude Code / Cursor 用户 I want init 模板自带"## 项目上下文"段 so that 我不需要在每次 `free-kiro init` 后手动把 marker 块说明补回 CLAUDE.md 顶部——agent 能立即识别文件末尾的 inject 块是机器生成的 steering 快照，而不是用户手写内容。

## Acceptance Criteria

- [AC-1] WHEN `free-kiro init` 把 `instructions_zh.md` 模板写入项目根 CLAUDE.md THE SYSTEM SHALL 在"## 编码规范（SessionStart 必须先 Read）"段之前插入 1 个 `## 项目上下文（自动注入的 steering）` 段，段内至少含 2 个 marker 链接与 1 个 always-mode 文档列表（`product.md` / `structure.md` / `tech.md`），文件总行数相对插入前增加至少 10 行。
- [AC-2] WHEN `free-kiro init --lang en` 把 `instructions_en.md` 模板写入项目根 CLAUDE.md THE SYSTEM SHALL 在同一位置插入英文版 `## Project context (auto-injected steering)` 段，结构对齐 zh 版（含 2 个 marker 链接 + 1 个 always-mode 文档列表），文件总行数相对插入前增加至少 10 行。
- [AC-3] WHEN init 写入 CLAUDE.md 时 THE SYSTEM SHALL 保留现有的 frontmatter（`mode: manual`）与 `# free-kiro-managed:` marker 行共 2 个边界行，新插入的段位于两者之后、`## 编码规范`段之前（位置偏移 ≤ 3 行）。
- [AC-4] WHEN init 写入 CLAUDE.md 时 THE SYSTEM SHALL 让 CLAUDE.md 中 2 个 inject marker 行（`<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`）保留在文件末尾的最后 5 行内，不被新插入的项目上下文段推到 marker 块与"## 帮助"段之间。
- [AC-5] WHEN `free-kiro init --ide auto --overwrite-instructions` 在已有 CLAUDE.md 的项目上重跑 THE SYSTEM SHALL 用新模板覆盖 CLAUDE.md 后，新写入的 CLAUDE.md 在 1 次 `diff` 中仍包含 1 段 `## 项目上下文（自动注入的 steering）`（即 init 1 次重写不丢失该段，diff 输出至少 1 行 hunk 标记新增）。
- [AC-6] THE SYSTEM SHALL 让 init 模板中的"## 项目上下文"段固定包含 2 个 marker 字符串 `<!-- free-kiro-managed:start -->` 与 `<!-- free-kiro-managed:end -->` 的字面引用，agent 读完该段即可定位文件末尾的 inject 块（不依赖 grep）。

## Out of Scope

- 不改 `internal/ide/templates/agents_*.md` —— agents 模板对应 OpenCode / CodeBuddy 的 AGENTS.md，由 spec `.kiro/specs/agents-template-coding-standards/` 处理。
- 不改 `internal/cli/init.go` 与 `internal/ide/ide.go` —— 本 spec 只动模板静态文本。
- 不改 inject 内容本身（marker 块内由 `free-kiro steering inject` 写入）—— 本 spec 只动 marker 块**之前**的"项目上下文"说明段。
- 不动 docs/STEERING.md §"自动注入到 IDE 指令文件" —— 那段已写明 inject 机制；项目上下文段是 init 模板里给 agent 的本地化提示，不重复文档内容。
