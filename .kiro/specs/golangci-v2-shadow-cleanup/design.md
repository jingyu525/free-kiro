# golangci-v2-shadow-cleanup — Design

## 根因

仓库 `.golangci.yml` 历史上写 `version: "2"` 但内容是 v1 schema（`linters-settings`
顶层键），v2 parser 静默忽略，`govet.enable-all: true` 从未生效。`fix(ci)`
PR（PR #9）将 config 完整迁移到 v2 schema 后，`enable-all: true` 真正启用，
暴露 13 处历史 govet shadow 违规，阻塞 `lint-go` job。

本 spec 目标：修干净 13 处违规，让 `enable-all: true` 真正可持续。

## Architecture

最小改动，仅在违规处做变量重命名 / 复用外层 err，不改任何业务逻辑。
**禁止用 `//nolint:govet` 绕过**（.kiro/steering/policy.md §5 零豁免政策）。

## Components

| Component | Responsibility |
|---|---|
| 各违规所在文件 | 在 shadow 处替换变量名或复用外层 err |
| `.golangci.yml` | `linters.settings.govet.enable-all: true`（最后一关） |

## Data Model

13 处 shadow 违规清单（来自 `golangci-lint run --timeout 5m` 在
`enable-all: true` 模式下的输出）：

| # | 文件 | 行 | 修法 |
|---|---|---|---|
| 1 | `internal/cli/demo.go` | 78 | `err = ...; err != nil`（复用外层 err） |
| 2 | `internal/cli/spec_from_browser.go` | 41 | 同上 |
| 3 | `internal/cli/spec_from_browser.go` | 56 | 同上 |
| 4 | `internal/cli/spec_generate.go` | 32 | 同上 |
| 5 | `internal/cli/spec_simple.go` | 36 | 同上 |
| 6 | `internal/cli/status_test.go` | 196 | 测试文件可改 `err` → `_` 或新名 |
| 7 | `internal/ide/ide.go` | 239 | `err = ...; err != nil`（复用外层 err） |
| 8 | `internal/skill/install.go` | 55 | 同上 |
| 9 | `internal/skill/install.go` | 101 | 同上 |
| 10 | `internal/skill/skill_test.go` | 222 | 测试文件可改 `err` → `_` 或新名 |
| 11 | `internal/skill/skill_test.go` | 226 | 同上 |
| 12 | `internal/skill/skill_test.go` | 229 | 同上 |
| 13 | `internal/workspace/workspace.go` | 173 | `err = ...; err != nil`（复用外层 err） |

修法二选一：
- **复用外层 err**（推荐，大多数情况）：`if err := foo(); err != nil` →
  `if err = foo(); err != nil`（`=` 而非 `:=`，复用外层 err）
- **不同变量名**（外层 err 不能复用的场景）：`if err := foo(); err != nil` →
  `if fooErr := foo(); fooErr != nil` 并同步替换该块内 `err` 引用

## Error Handling

无变化。每个 shadow 违规修完后，原错误处理路径（return err / log）保持
不变——只是声明的变量名 / 复用方式调整。

## Testing Strategy

不需要新增单元测试。验证：
- `golangci-lint run --timeout 5m` 在 `enable-all: true` 模式下输出 `0 issues.`
- `go test ./...` 全绿（业务逻辑未改，但确认无回归）
- `go test -race ./...`（.kiro/steering/agent-rules.md §4）

## Migration / Rollout

无 expand-contract / feature flag：纯代码清理 + 启用历史 lint 规则。
随本 spec 一起落地一次。
