# enforce-golang-standards-zero-dead-code — Design

## Architecture

按"配置先行 → 自动可修 → 人工清死代码 → CI 复核"四阶段推进。每阶段都是
"可验证产出"，review 时逐段 sign-off。

```
   阶段 1                 阶段 2                阶段 3                阶段 4
   .golangci.yml   →    gofmt / goimports  →  删 19 处 U1000  →   golangci-lint
   staticcheck          -w 一次性修 13        死代码（按包        v2.14.0 +
   加 U1000              文件                拆 wave）          staticcheck
                                                                    全绿
```

`unused` 在 golangci-lint 中是独立 linter，**不是** `staticcheck.checks`
数组里的子检查——`golangci-lint v2` 的 staticcheck 集成刻意排除了
U1000（避免误报未引用符号）。需要把 `unused` 加到 `linters.enable` 数组
才能触发 U1000 报告。`golangci-lint v2.x` 已内置该 linter，无需额外安装。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `.golangci.yml` | 在 `linters.enable:` 数组下追加 `- unused`（独立 linter，非 staticcheck 子检查——golangci-lint v2 的 staticcheck 集成不包含 U1000，必须用 standalone linter）；移除 `linters-settings.staticcheck.checks:` 块（如有） | YAML |
| `internal/ide/ide.go:426` | `defer f.Close()` 改为 `_ = f.Close()` 或拆出显式 `if err := f.Close(); err != nil { ... }` | n/a |
| `internal/hooks/hooks_test.go:400` | 函数签名 `fn := func(p string) (...)` 改为 `func(_ string) (...)` | n/a |
| `internal/lint/quality.go:171` | `func CheckFewAC(doc string, min int) []Issue` 参数 `min` 改名为 `minAC`（消除 builtin 遮蔽） | `CheckFewAC(doc string, minAC int) []Issue` |
| `internal/{lint,models,spec,text,visualize}/*.go` | `gofmt -s -w` + `goimports -local github.com/jingyu525/free-kiro -w` 一次性格式化 | n/a |
| `internal/cli/stubs.go` (3 函数) | 删 `attachPhaseFlags` / `stubCmd` / `stubFor`，仅保留 `internal/cli` 中实际被 import 的 helper | n/a |
| `internal/spec/generator.go` (6 函数) | 删 `writeIfMissing` / `ensureMeta` / `listSpecDocs` / `firstDocExists` / `fmtMissing` / `phaseForDoc` | n/a |
| `internal/steering/frontmatter.go:115` | 删 `splitLines`（已被 `internal/text/text.SplitLines` 替代） | n/a |
| `internal/steering/store.go:175` | 删 `deriveName`（无 caller） | n/a |
| `internal/upgrade/upgrade.go:33` | 删未引用常量 `httpTimeout` | n/a |
| `internal/visualize/mermaid.go:122` | 删 `taskWaves`（已由 `internal/visualize/taskgraph.go` 接管） | n/a |
| `internal/visualize/server.go:45,180` | 删 `osDirEntry` 类型别名 + `(*Server).loadWorkspace` 方法（无 caller） | n/a |
| `internal/visualize/server_static.go` (4 函数) | 删 `atoiLocal` / `detectBrowserAutoOpen` / `openBrowser` / `detectBrowserOpen`（前提：build tag 排除确认） | n/a |
| `.kiro/steering/coding-style.md` 第 8 章 | 追加"零死代码（U1000 zero-tolerance）"段落 | Markdown |

## Data Model

无。这是 lint 规约收尾工程，无数据结构变更。

## Error Handling

本 spec 涉及 1 处 errcheck（`internal/ide/ide.go:426` `defer f.Close()`），
按既有 `enforce-golang-standards-zero-exemptions` design 文档约定的
"模式 A：`_ = <call>`"处理——改为 `_ = f.Close()`，并在同函数内补一段
注释解释为何吞错（写文件失败无法恢复，CLI 用户从 stdout 提示已感知）。
不使用 `//nolint:reason` 注释，保持豁免总数 ≤ 5 的硬约束。

## Testing Strategy

- **单元测试**：`go test ./...` 必须全绿；死代码删除前先 grep 整个仓库
  确认无引用，删除后再次 grep 确认无回归。
- **集成测试**：`free-kiro lint` 在 spec 落地后必须以 0 错误退出。
- **门禁验证**：`golangci-lint run ./... --timeout=5m` 退出码 0；
  `staticcheck ./...` 退出码 0（两者行为一致后即可信任）。
- **回归保护**：在 `.kiro/steering/coding-style.md` 第 8 章追加 U1000 条款，让
  后续 AI agent 在写代码前自觉避免引入未引用符号。

## Migration / Rollout

无数据库迁移、无 feature flag 切换。本次仅：

1. 配置先行（`.golangci.yml` 改 1 行 + `checks` 数组）。
2. 自动可修（`gofmt -s -w .` + `goimports -local ... -w .`，13 文件）。
3. 人工清死代码（19 个符号，按所属包分 4 个 wave 并行）。
4. 复核（`golangci-lint` + `staticcheck` + `go test ./...` + `free-kiro
   lint` 四件套全绿后 `free-kiro spec complete`）。
