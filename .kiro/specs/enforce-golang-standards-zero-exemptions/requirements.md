# enforce-golang-standards-zero-exemptions

把 `golang-coding-standards` spec 的"软门禁 + 历史代码豁免"模式切换为
"全代码库硬门禁 + 零豁免"模式：移除 `.golangci.yml` 中对
`internal/(spec|lint|steering|hooks|skill|upgrade|workspace|ide)` 的
exclude-rule、把 `.golangci.yml` 从 v1 schema 迁移到 v2 schema（兼容本地
golangci-lint v2.13.1 与 CI golangci-lint-action v6/v8），逐包清理历史欠账
（实测 ~327 处 lint 问题：192 errcheck + 86 gofmt + 47 revive + 2 staticcheck），
CI lint job 从 `continue-on-error: true`（WARN）切换为硬阻断（ERROR 阻断 merge）。

## User Stories

- As a free-kiro 项目维护者, I want 所有 Go 代码（包括 `internal/spec`、
  `internal/lint` 等核心包）对软门禁 linter 一视同仁, so that 规范的
  "零吞错误 / 导出必有注释 / 命名一致"等条款在自家核心代码上同样适用，
  不会出现"规范只约束边缘代码、放纵核心代码"的双标。
- As a CI 流水线, I want `.golangci.yml` 的 `issues.exclude-rules` 不再
  包含任何针对 `internal/...` 的目录级豁免, so that 新提交立刻按
  全规则检查，无需等"逐包收紧"的后续任务。
- As a 新加入的贡献者, I want `make lint-go` 与 `golangci-lint run ./...`
  在本地一次跑通、CI 一次跑通, so that 不需要维护私人 `//nolint` 列表
  也能保持 0 lint 告警。

## Acceptance Criteria

- [AC-1] WHEN 开发者执行 `golangci-lint run ./...`（或 `make lint-go`）时 THE
  SYSTEM SHALL 退出码 0，stdout/stderr 不输出任何 issue。
- [AC-2] WHILE `.golangci.yml` 仍存在 `path: 'internal/(spec|lint|...)/...'` 这类
  目录级 exclude-rule 时 THE SYSTEM SHALL 把软门禁 linter（`staticcheck`、
  `errcheck`、`revive`）视为"规范未全面落地"，CI `lint-go` job 必须以
  ERROR 退出。
- [AC-3] WHERE `.golangci.yml` 使用 v1 schema（如 `output.formats` 为 slice、`gofmt`/`goimports` 在 `linters.enable` 内、未声明 `version: "2"`）时 THE SYSTEM SHALL 在迁移完成后使用 v2 schema（`version: "2"` + `formatters.enable` 拆分 `gofmt`/`goimports` + `output.formats` 改为 map），并通过 `golangci-lint v2.x` 的 config 加载校验。
- [AC-4] UNLESS linter 在该行加注 `//nolint:reason <具体原因>` 时 THE SYSTEM
  SHALL 默认对所有非 `vendor/` / `third_party/` / `testdata/` / `_gen.go` /
  `*.pb.go` 的 Go 文件应用 5 类 linter（`govet` + `gofmt` + `goimports` +
  `staticcheck` + `errcheck` + `revive`）；本 spec 落地后，全仓库
  `//nolint:reason` 注释总数 ≤ 5（仅允许在 OS 信号处理、不可恢复 stderr
  写入等极少数场景保留）。
- [AC-5] IF 任意包新增/修改 Go 代码后存在 errcheck 未处理返回、revive
  `exported` 缺注释、revive `unused-parameter` 未改名 `_`、
  staticcheck `QF1xxx`/`SA4xxx` 告警、或 gofmt diff 时 THEN
  `.github/workflows/ci.yml` 的 `lint-go` job 必须以非 0 退出码失败，
  阻断 PR merge。
- [AC-6] THE SYSTEM SHALL 在 `.kiro/steering/coding-style.md` 第 8 章"AI 协作"明确
  写入："本项目 lint 配置零豁免；任何新增 `//nolint` 必须附带理由，
  由 PR reviewer 在 review 时逐条签字。"

## Out of Scope

- 不修改 `golang-coding-standards` spec 已落地的 8 章节文档内容（命名 /
  错误处理 / 并发 / 接口 / 测试 / 注释 / 依赖 / AI 协作）；本 spec 只
  在第 8 章追加一条"零豁免"条款。
- 不调整 `revive` 的 `rules` 字段中 9 条规则的 `severity`（仍为 warning，
  本次升级后由 CI ERROR 阻断）；如需新增/裁剪规则由后续独立 spec 处理。
- 不升级 `golangci-lint-action` 主版本（保留 v6，与 `golangci-lint v2.x`
  兼容）；版本升级留作未来 spec。
- 不处理 `dist/` 下的构建产物和 `testdata/` 测试数据（已通过
  `exclude-dirs` 排除）。
- 不动 `//go:generate` 生成的 `*_gen.go`、`*.pb.go`；按既有
  `exclude-dirs` 排除，未来若引入生成代码需另起 spec。
