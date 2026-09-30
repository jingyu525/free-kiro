# cli-version-flag — Design

## Architecture

在 `internal/cli/root.go` 给 rootCmd 设置 `Version` 字段，并调用
`SetVersionTemplate()` 控制输出格式。Cobra 在发现 `Version` 字段非空时
会自动注册 `--version` flag，不需要在 main.go 加任何代码（main.go 已经
走 `cli.Execute()`，Cobra 内部会拦截 flag 并调用模板打印）。

`buildVersion` / `buildCommit` / `buildDate` 三个 package-level 变量已
在 `internal/cli/upgrade.go` 定义，由 `.goreleaser.yaml` 的 ldflags 注入。
本 spec 只读取，不新增。

Version template 处理 dev build：commit / date 为空时退化为 `"unknown"`
而不是 panic 或空字符串；`buildVersion == "dev"` 时保持 `"dev"` 字面量。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `internal/cli/root.go` (rootCmd) | 设置 `Version` 字段 + `SetVersionTemplate` | cobra.Command |
| `internal/cli/upgrade.go` (已有) | 提供 `buildVersion` / `buildCommit` / `buildDate` package-level var | (无改动) |
| `internal/cli/root_test.go` (新) | 单元测试：Version 字段非空 + 输出格式 + dev build 路径 | `TestRootVersionFlag` |

## Data Model

无新结构。复用现有 package-level 变量。

## Error Handling

- `buildCommit == ""` → 输出 `"unknown"`
- `buildDate == ""` → 输出 `"unknown"`
- `buildVersion == "dev"` → 正常输出 `"dev"`，不视为错误
- `--version` 路径不调用任何可能 panic 的函数；不在 SetVersionTemplate
  里做 IO（避免 race / 测试 flaky）

## Testing Strategy

- **单元测试** (`internal/cli/root_test.go`)：
  - `TestRootCmdHasVersion` — 反射或直接读 `rootCmd.Version`，断言非空
  - `TestVersionOutput` — 用 `rootCmd.SetArgs([]string{"--version"})` +
    `bytes.Buffer` 捕获 stdout，断言：
    - 输出单行、含 `"free-kiro version"` 前缀
    - exit code 0（Cobra 返回 nil error）
    - dev build 时含 `"dev"`
- **手工冒烟**：构建 snapshot（`goreleaser build --snapshot --clean`），
  跑 `dist/free-kiro_darwin_arm64/free-kiro --version` 验证输出与 ldflags
  注入值一致

## Migration / Rollout

无 migration。GoReleaser 下一次 release 就会自然带上正确 version。
不需要 feature flag。