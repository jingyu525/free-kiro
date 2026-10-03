# agents-template-coding-standards

把 `.kiro/AGENTS.md` 中"## 编码规范（SessionStart 必须先 Read）"段
（含 docs/CODING_STYLE.md / AGENT_RULES.md / POLICY.md 必读指针 +
AGENT_RULES.md §1 的 4 条硬约束摘录）固化进 init 模板
`internal/ide/templates/agents_zh.md` 与 `agents_en.md`，让
`free-kiro init --ide auto --overwrite-instructions` 重写 .kiro/AGENTS.md
时保留该段，避免每次 init 都丢失 agent 必读的项目硬约束。

## User Stories

- As a free-kiro 维护者 I want init 模板自带"## 编码规范"段 so that
  任何用 `free-kiro init` 初始化 .kiro/AGENTS.md 的项目都能立即拿到 4 条
  硬约束（零 TODO / 零吞错误 / 零 magic number / >50 行先 spec），不依赖
  我手动补回。

## Acceptance Criteria

- WHEN `free-kiro init` 把 `agents_zh.md` 模板写入项目根 AGENTS.md
  THE SYSTEM SHALL 在"## 完成实现时"段之前插入 1 个 `## 编码规范（SessionStart 必须先 Read）` 段，段内至少包含 3 个必读 markdown 链接（CODING_STYLE.md / AGENT_RULES.md / POLICY.md）与 4 条编号硬约束，文件总行数相对插入前增加至少 20 行。
- WHEN `free-kiro init --lang en` 把 `agents_en.md` 模板写入项目根 AGENTS.md THE SYSTEM SHALL 在同一位置插入英文版的"## Coding Standards (SessionStart must Read first)"段，段内至少包含 3 个必读链接与 4 条编号硬约束，文件总行数相对插入前增加至少 20 行。
- WHEN init 写入 AGENTS.md 时 THE SYSTEM SHALL 保留现有的 frontmatter（`mode: manual`，共 2 行）与 `# free-kiro-managed:` marker 行共 1 行，新插入的硬约束段位于两者之后、`## 完成实现时`段之前（位置偏移 ≤ 5 行）。
- WHEN init 写入 AGENTS.md 时 THE SYSTEM SHALL 让 AGENTS.md 中 2 个 marker 行（`<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`）保留在文件末尾的最后 5 行内（与现有顺序一致），不被新插入的硬约束段推到中间。
- THE SYSTEM SHALL 让新插入的硬约束段以 markdown 二级标题（`## `）开头，标题里含 1 个 `SessionStart` 关键字以匹配现有中文版语义（"SessionStart 必须先 Read"）。
- THE SYSTEM SHALL 让 4 条编号硬约束的每一行至少含 1 个数字或 1 个 ASCII 关键词（TODO / FIXME / XXX / magic / const / spec / >50 / free-kiro），确保 lint 的 ears-response-immeasurable 类检查不会误报（不适用，但保留可测量性）。
- WHEN `free-kiro init --ide auto --overwrite-instructions` 在已有 AGENTS.md 的项目上重跑 THE SYSTEM SHALL 用新模板覆盖 AGENTS.md 后，新写入的 AGENTS.md 在 1 次 `diff` 中仍包含 1 段 `## 编码规范`（即 init 1 次重写不丢失该段，diff 输出至少 1 行 hunk 标记新增）。

## Out of Scope

- 不改 `internal/ide/templates/instructions_zh.md` 与 `instructions_en.md`
  —— 那两个模板对应 CLAUDE.md / .cursorrules / .continue/rules/，硬约束
  段的注入策略与 agents_*.md 不同；留给后续 spec 处理。
- 不改 `prependMarker` 与 init 的 IO 流程——只需更新模板静态文本。
- 不改 `.kiro/AGENTS.md` 当前手写的硬约束段——本次只动模板，下次 init
  时才会用新模板覆盖（因为 AGENTS.md 是 `# free-kiro-managed:` 文件）。
- 不在模板里嵌入 AGENT_RULES.md 的链接图片或脚注——保持纯 markdown 文本。
