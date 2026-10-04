# golangci-v2-shadow-cleanup

启用 `.golangci.yml` 的 `govet.enable-all: true` 并修仓库里 13 处
因此暴露的 govet shadow 违规。

## 根因（背景）

`.golangci.yml` 历史上写 `version: "2"` + v1 schema 写法（`linters-settings`
顶层键），v2 parser 静默忽略顶层 `linters-settings`，导致历史
`govet.enable-all: true` 实际从未生效。`golangci-v2-shadow-cleanup`
由 `fix(ci)` PR 落地后 v2 schema 真正生效，再保留 `enable-all: true`
就会暴露 13 处 shadow 违规，阻塞 `lint-go` job。

## User Stories

- 作为 free-kiro 维护者，我希望 `golangci-lint run` 启用 `govet.enable-all: true`
  后仍输出 `0 issues.`，以便新代码受到完整 govet 检查（包括 shadow / nilness
  等扩展 check），不留下历史 lint 欠账。

## Acceptance Criteria

- [AC-1] WHEN 执行 `golangci-lint run --timeout 5m` 在仓库根目录，THE SYSTEM SHALL 在 stdout 输出 `0 issues.`（不含 "X issues" 的非零数字）。
- [AC-2] WHERE `.golangci.yml` 中 `linters.settings.govet.enable-all` 设为 `true`，
  THE SYSTEM SHALL 仍输出 `0 issues.`（即启用 enable-all 不暴露新违规）。
- [AC-3] THE SYSTEM SHALL 修干净 13 处 govet shadow 违规（每处都是 `if err := ...;
  err != nil { ... }` 在已声明 `err` 的作用域里 shadow）。修法二选一：
  - 复用外层 `err`：`if _, err := foo(); err != nil { ... }` 改为 `if err = foo();
    err != nil { ... }`（注意 `=` 而非 `:=`，外层已声明）
  - 不同变量名：`if err := foo(); err != nil { ... }` 改为 `if fooErr := foo();
    fooErr != nil { ... }` 并同步替换该块内 `err` 引用
- [AC-4] THE SYSTEM SHALL 在 `golangci-lint run` 输出里 0 个 `//nolint:govet` 注释
  （.kiro/steering/policy.md §5 零豁免政策：通过修代码达成，不通过豁免绕过）。
- [AC-5] IF 任一 shadow 违规通过 `//nolint:govet` 绕过而未修代码，THEN THE SYSTEM
  SHALL 视为本 spec 未完成。

## Out of Scope

- 不引入除 `govet.enable-all` 外的其他新增 linter（如 revive 扩展规则、
  staticcheck 新 checks）。
- 不修改 13 处违规的具体业务逻辑，只调整变量名 / 复用方式。
- 不开新 spec 处理 13 处以外的其他历史 lint 违规（如果未来某次重新启用
  某 linter 后暴露新违规，按当时情况另开 spec）。
