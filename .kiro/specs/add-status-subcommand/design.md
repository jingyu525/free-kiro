# add-status-subcommand — Design

## Architecture

新增 `internal/cli/status.go`，定义 `statusCmd`（Cobra 子命令），挂到
rootCmd 的 init() 里。命令本身只做三件事：

1. 调 `internal/workspace.ListSpecs()` 拿到 spec 列表 + `.kiro/.current`
2. 调 `internal/workspace.ReadSummary()` 聚合每个 spec 的 phase /
   approved / drift / tasks 摘要
3. 按 flag（人类 / JSON）选 renderer

复用 `internal/cli/visualize_cmds.go` 已有的表格渲染 helper（如果存在），
否则新写一个最小表格（`text/tabwriter` 包即可，零新依赖）。

JSON 输出的 schema 跟 `serve` 的 `/api/summary` 完全一致（同一个
`internal/models` 结构 + 同一个 `Summary()` 函数），保证 watch /
serve / status 三处 API 同源。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `internal/cli/status.go` (新) | statusCmd + 人类/JSON 双 renderer | `statusCmdFactory() *cobra.Command` |
| `internal/workspace` (复用) | 列 `.kiro/specs/*` 子目录 + 读 `.kiro/.current` | `ListSpecs()`, `ActiveSpec()` |
| `internal/models` (复用) | Spec / Drift / Tasks summary 结构 | `Spec`, `Summary`, `Tasks` |
| `internal/cli/visualize_cmds.go` (复用) | 表格列宽对齐（若有现成 helper） | 或 `text/tabwriter` 标准库 |
| `internal/cli/root.go` (改) | init() 里 `AddCommand(statusCmdFactory())` | — |

## Data Model

复用现有 `internal/models` 的 `Spec` / `Summary` / `Tasks` 结构。**不
引入新结构**。`Summary` 已经包含 `active` 字段、`drift` 字段、tasks
done/total/waves 字段——本次只需让 `status` 共享同一组装逻辑。

## Error Handling

| 情况 | 行为 |
|---|---|
| `.kiro/` 不存在 | stderr 输出 `.kiro/ workspace not found, run 'free-kiro init' first`，exit 3 |
| `.kiro/specs/` 存在但空 | stdout 输出 `No specs found`（JSON: `{"active":"","specs":[]}`），exit 0 |
| `.kiro/.current` 指向不存在的 spec | 输出时打 `(not found)` 标记，不报错 |
| `--json` 序列化失败 | 防御性 wrap（结构极简，几乎不可能触发），exit 2 |
| 子目录不是合法 spec（缺 `.meta.json`） | 跳过 + stderr warn（不阻断其他 spec） |

## Testing Strategy

| 测试 | 文件 | 覆盖 |
|---|---|---|
| `TestStatusHumanOutput` | `internal/cli/status_test.go` | tmpfs .kiro/ + 2 个 spec + 1 个 .current，断言 stdout 含表头 + spec 行 + active 标记 |
| `TestStatusJSONOutput` | 同上 | `--json` 断言 JSON 解析成功 + `active` 字段正确 |
| `TestStatusNoWorkspace` | 同上 | cwd 在空目录 → exit 3 |
| `TestStatusNoSpecs` | 同上 | .kiro/specs/ 空 → "No specs found" |
| `TestStatusPointerMissing` | 同上 | .current 指向不存在的 spec → 不报错，输出 `(not found)` |
| `TestStatusIsReadOnly` | 同上 | 跑 status 前后 .kiro/specs/foo/.meta.json mtime 不变 |

全部用 stdlib `testing`，复用 cli 包现有的 `setUpTempWorkspace(t)` helper
（如有）；否则按 `t.TempDir()` + 写最小 .kiro/ 文件的常规做法。

## Migration / Rollout

无 migration。`watch --preset reactive` / `watch status` preset 在
升级后自动能用（因为它们调的就是 `free-kiro status`，之前是 bug，
现在自然修复）。**不需要改 watch 包任何代码**。
