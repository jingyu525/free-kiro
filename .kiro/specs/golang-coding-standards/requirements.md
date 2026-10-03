# golang-coding-standards

本 spec 为 free-kiro 项目（Go CLI）建立一份**项目级 Go 编码规范**，作为所有
后续 Go 代码（含 AI agent 自动生成的代码）的风格基线。规范以一份 Markdown
主文档承载，外加 `golangci-lint` 配置文件把可机械检查的规则变成 CI 硬门禁。

## User Stories

- As a free-kiro 项目维护者, I want 一份统一的 Go 编码规范文档, so that
  贡献者（含 AI agent）产出的代码风格一致、可读性可预期。
- As a CI 流水线, I want 一份 `golangci-lint` 配置把规范里"可机器检查"
  的规则变成构建门禁, so that 违规代码在合并前被自动拦截。
- As a 新加入的 Go 开发者, I want 在 `CONTRIBUTING.md` 入口就能看到规范
  索引, so that 5 分钟内就能定位"项目推荐怎么写 Go"。

## Acceptance Criteria

- [AC-1] WHEN 开发者执行 `make lint` 或 `free-kiro lint` 时 THE SYSTEM SHALL
  加载 `golangci.yml` 并对 `./...` 下的所有 Go 文件运行至少 `govet`、
  `staticcheck`、`errcheck`、`gofmt`、`goimports` 五个 linter。
- [AC-2] WHILE 任意 Go 源文件存在未解决的 `goimports` diff 或 `gofmt` 差异时
  THE SYSTEM SHALL 在 `free-kiro lint` 中以 ERROR 退出码 1 拒绝通过。
- [AC-3] WHERE 规范明确要求命名/导出/错误包装等规则时 THE SYSTEM SHALL 在 `.golangci.yml` 中开启对应 linter（如 `revive`/`staticcheck` 的 `ST1003`/`ST1019`/`ST1023` 等）使其在 CI 中自动报错。
- [AC-4] UNLESS 该文件位于 `vendor/`、`*.pb.go`、`*_gen.go` 生成的目录外 THE
  SYSTEM SHALL 默认对所有 `.go` 文件应用上述检查。
- [AC-5] IF 提交信息中包含 `//nolint:<linter>` 注释且未附 `//nolint:reason`
  说明时 THEN `revive` 的 `exported`/`revive` 规则 THE SYSTEM SHALL
  在 lint 输出中产生 WARNING 提示要求补充原因。
- [AC-6] THE SYSTEM SHALL 在仓库根目录的 `CONTRIBUTING.md` 提供指向 `docs/CODING_STYLE.md` 的链接，且 `docs/CODING_STYLE.md` 至少覆盖： 命名约定、错误处理、并发、接口设计、测试、注释与文档、依赖管理、 与 AI agent 协作的硬性要求 8 个章节。
- [AC-7] THE SYSTEM SHALL 在 `docs/CODING_STYLE.md` 的每一节给出至少一个
  ✅ 推荐 写法 与一个 ❌ 反例，并以可复制的 Go 代码片段呈现。

## Out of Scope

- 不重写或格式化历史已合并代码；规范只对新代码与修改行强制生效。
- 不引入除 `golangci-lint` 之外的额外静态分析二进制（如 `gopls`、
  `semgrep`）作为硬门禁；可在文档里"推荐"但不强制。
- 不为 monorepo 多模块场景设计；本项目是单 module 单 binary。
- 不替项目选定 license header 模板；保留现状（BSD-3 / Apache-2 由
  后续 license-spec 决定）。
