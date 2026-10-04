# enforce-golang-standards-zero-dead-code

`golangci-lint v2.14.0` + `staticcheck` 全仓库扫描后，仍残留 16 处 lint 问题
（1 errcheck + 13 gofmt + 2 revive）以及 19 处死代码（staticcheck U1000 命中）。
本 spec 收尾遗留违规、把 U1000（unused）纳入硬门禁，并保证 CI 在新增死代码 /
errcheck / gofmt / revive 问题时阻断 merge。

## User Stories

- As a free-kiro 项目维护者 I want 仓库 0 lint 告警 + 0 死代码 so that 规范在自家核心代码上彻底落地，避免双标。
- As a CI 流水线 I want staticcheck `U1000` 加入硬门禁 so that 新增的死函数 / 常量在 PR 阶段就被拦截。
- As a 新加入的贡献者 I want 剩余 16 处 lint 问题被一次性清理 so that 本地 `make lint-go` 与 CI 一次跑通。

## Acceptance Criteria

[AC-1] WHEN 开发者执行 `golangci-lint run ./...`（或 `make lint-go`）THE SYSTEM SHALL 以退出码 0 退出且 stdout/stderr 输出 0 issue，覆盖 `cmd/` 与全部 `internal/` 子包。
[AC-2] WHEN CI 运行 `.github/workflows/ci.yml` 的 `lint-go` job THE SYSTEM SHALL 在 60 秒内完成；若发现 errcheck 未处理返回、revive 规则违反、staticcheck U1000 命中或 gofmt diff，job 必须以非 0 退出码失败。
[AC-3] WHILE `.golangci.yml` 的 `linters.enable` 数组包含 `- unused` THE SYSTEM SHALL 对 `internal/` 与 `cmd/` 下所有非 `*_test.go` / `*_gen.go` / `*.pb.go` 文件运行 `unused`（对应 staticcheck U1000）并报告未引用的 package-level 函数、常量、类型、变量；命中即视为硬门禁违规，CI 阻断 merge。
[AC-4] WHERE `internal/visualize/server_static.go` 被 build tag 排除（`//go:build !visualize_static`）THE SYSTEM SHALL 允许其保留未引用符号并在 lint 输出中以 `unused (U1000)` 标记，最多产生 4 处此类豁免。
[AC-5] WHERE `internal/visualize/server_static.go` 默认编译进入二进制 THE SYSTEM SHALL 启用 `detectBrowserAutoOpen` / `openBrowser` / `detectBrowserOpen` / `atoiLocal` 全部 4 个入口，否则在 30 天内视为死代码清理对象。
[AC-6] UNLESS 资源释放调用（`defer xxx.Close()` / best-effort `os.RemoveAll`）附 `//nolint:reason` 注释且总数 ≤ 5 THE SYSTEM SHALL 在 1 秒内拒绝任何静默吞错的改动进入 main 分支。
[AC-7] IF 任意 Go 文件被 `gofmt -s` 或 `goimports -local github.com/jingyu525/free-kiro` 改写 THEN `golangci-lint run` 必须在 5 秒内以非 0 退出码失败并打印 diff 摘要。
[AC-8] THE SYSTEM SHALL 在 `.kiro/steering/coding-style.md` 第 8 章"AI 协作"段落末尾追加 1 条 "零死代码（U1000 zero-tolerance）" 条款，与既有"零 `//nolint` 豁免"条款并列，全文字数增加不少于 30 字。
[AC-9] THE SYSTEM SHALL 在本 spec 的 `tasks.md` 全部 16 项勾选完成且 `go test ./...` 在 60 秒内 100% 通过后调用 `free-kiro spec complete enforce-golang-standards-zero-dead-code` 把 spec 标记为 DONE。
[AC-6] IF 任意 Go 文件被 `gofmt -s` 或 `goimports -local github.com/jingyu525/free-kiro` 改写 THEN `golangci-lint run` 必须在 5 秒内以非 0 退出码失败并打印 diff 摘要。
[AC-7] THE SYSTEM SHALL 在 `.kiro/steering/coding-style.md` 第 8 章"AI 协作"段落末尾追加 1 条 "零死代码（U1000 zero-tolerance）" 条款，与既有"零 `//nolint` 豁免"条款并列，全文字数增加不少于 30 字。
[AC-8] THE SYSTEM SHALL 在本 spec 的 `tasks.md` 全部 16 项勾选完成且 `go test ./...` 在 60 秒内 100% 通过后调用 `free-kiro spec complete enforce-golang-standards-zero-dead-code` 把 spec 标记为 DONE。

## Out of Scope

- 不修改上游 spec `enforce-golang-standards-zero-exemptions` 已落地的 `.golangci.yml` v2 schema、`formatters` 拆分、CI `continue-on-error` 移除机制；本 spec 仅在其基础上扩展 `staticcheck.checks`。
- 不升级 `golangci-lint-action`（保留 v6）或切换 `golangci-lint` 到 v3 major；版本升级由后续独立 spec 处理。
- 不引入新 linter（如 `gosec` / `gocritic` / `gocyclo`）；本 spec 范围限于既有 6 类（govet / staticcheck / errcheck / revive / gofmt / goimports）的规则深度调整。
- 不动 `internal/lint` / `internal/spec` 等核心包的功能行为；本 spec 仅清理其残留的 gofmt / revive 问题。
- 不删除 `internal/visualize/server_static.go` 等 build-tag 隔离文件；本 spec 仅在文件确实编译时引用其符号，若实际编译排除则保持原样。
- 不处理 `dist/` 构建产物与 `testdata/` 测试数据（已通过 `exclude-dirs` 排除）。
