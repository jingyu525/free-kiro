# spec — 规格生命周期深挖

完整状态机：`draft → (requirements → design → tasks) | (design → requirements → tasks) → approved → implementing → done`

## 阶段速查

| 状态 | 下一步动作（用 `kiro spec next <name>` 拿预言机） |
|---|---|
| `draft` | `free-kiro spec generate <name> --phase requirements`（或 `design` 起步走 Design-First） |
| `requirements` | 就地编辑 `.kiro/specs/<name>/requirements.md`，跑 `free-kiro lint <name>` 直到 exit 0；再 `spec generate --phase design` |
| `design` | 就地编辑 `design.md`；`lint` 通过后 `spec generate --phase tasks` |
| `tasks` | 就地编辑 `tasks.md`（任务格式 `- [ ] #N … [deps: …]`）；`lint` 通过后请用户 *显式* 同意后 `spec approve <name>` |
| `approved` | `free-kiro spec start <name>` 进入实现 |
| `implementing` | 按 `free-kiro task list <name>` 显示的 waves 顺序实现；每完成一个把 `- [ ]` 改成 `- [x]`；每完成 wave 跑 `free-kiro spec sync <name>` |
| `done` | 终态；`free-kiro spec complete <name>` 写入 |

## EARS 句式（requirements.md 必用，否则 lint 拦）

| 场景 | 句式 |
|---|---|
| 事件 | `WHEN <trigger> THE SYSTEM SHALL <response>` |
| 状态 | `WHILE <state> THE SYSTEM SHALL <behavior>` |
| 可选 | `WHERE <feature> THE SYSTEM SHALL <behavior>` |
| 异常 | `UNLESS <exemption> THE SYSTEM SHALL <default>` |
| 复杂条件 | `IF <cond> THEN <branch> THE SYSTEM SHALL <behavior>` |
| 基线 | `THE SYSTEM SHALL <behavior>` |

例：

```markdown
## Acceptance Criteria
- WHEN user submits a valid login form THE SYSTEM SHALL redirect to /home.
- WHILE the user is unauthenticated THE SYSTEM SHALL hide all admin links.
- WHERE the workspace has a `.kiro/` directory THE SYSTEM SHALL auto-load its steering.
```

## tasks.md 写法（lint 也会拦）

```markdown
- [ ] #1 Set up module layout
- [ ] #2 Implement core domain model [deps: #1]
- [ ] #3 Implement persistence layer [deps: #1]
- [ ] #4 Wire API [deps: #2,#3]
```

- 任务编号必须连续无跳号
- 依赖只能引用更早的编号（不能反向）
- 不能环 / 悬空 / 自引用
- 每个任务小到一次可实现

## Workflow 变体

```bash
# 默认 Requirements-First：需求 → 设计 → 任务
free-kiro spec new <name> --prompt "..."

# Design-First：你已有架构设计
free-kiro spec new <name> --prompt "..." --workflow design-first

# Bugfix 变体（用 bugfix.md 而非 requirements.md）
free-kiro spec new <name> --prompt "<复现>" --type bugfix

# Quick Spec（免审批 + 一次性生成三份文档）
free-kiro spec quick <name> --prompt "<一句话改动>"
```

## 漂移处理

实现期间 AC 或任务数变化 → `free-kiro spec status <name>` 看漂移 → 改 spec 后 `free-kiro spec sync <name>` 重新基线。不要 `spec generate` 重生成已经填了真实内容的文档（会覆盖）。

## 何时停手

- 三份文档都过 lint（`exit 0`）
- tasks.md 所有 `- [ ]` 都改成 `- [x]`
- `free-kiro spec status <name>` 显示 phase=done
- 用户接受最终交付

## 一致性分析（advisory）

```bash
free-kiro spec analyze <name>            # 模糊语言 / 重复 AC / 需求↔任务脱节
```

`analyze` 永不出错，exit 永远 0 —— 它是给作者看的，不拦状态机。