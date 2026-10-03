# unify-ears-template-and-hints

让 EARS spec 模板的 AC 示范行自带 `[AC-N]` 前缀，避免新生成的 spec 一过
`free-kiro lint` 就被 `ears-ac-missing-id` WARNING 淹没；同时把
`internal/lint/` 各处 linter `Hint:` 字符串指向的英文 slug
（`#five-templates` / `#vague-words` / `#semantic-quality-gates`）跟
`docs/EARS.md` 中文 heading 实际可解析的 anchor 对齐，消除 GitHub 上点不开的
死链。

## User Stories

- As a free-kiro 用户 I want `free-kiro spec new` 生成的 `requirements.md` 一跑
  `lint` 就自带 `[AC-N]` 编号的合规 AC 模板 so that 我不需要再给 6 条占位符
  手动加编号就能通过 `ears-ac-missing-id` 检查。
- As a free-kiro 用户 I want `free-kiro lint` 输出里的 `Hint:` 链接在 GitHub
  渲染的 docs/EARS.md 上点击能定位到对应 heading so that 我能在不离开终端的
  情况下追溯到权威文档。

## Acceptance Criteria

[AC-1] WHEN 工程师执行 `free-kiro spec new <name>` 后立刻跑 `free-kiro lint <name>` 时 THE SYSTEM SHALL 不再因模板生成的 6 条 AC 示范行缺少 `[AC-N]` 前缀而触发 `ears-ac-missing-id` WARNING；实现方式是在 `internal/spec/templates/requirements.md.tmpl` 第 26-31 行的 6 条 AC 示范前各加 `[AC-N]` 前缀（保留 N 为占位符），并把示范响应从 `-- [占位描述].` 改成 `<占位描述>.` 的紧凑形式，让用户 0 改动即可通过 `ears-ac-missing-id` 检查。

[AC-2] WHEN 工程师执行 `go test ./internal/lint/...` 时 THE SYSTEM SHALL 在 100 ms 内对 `internal/lint/requirements.go`、`internal/lint/bugfix.go`、`internal/lint/quality.go` 这 3 个文件中每条形如 `docs/EARS.md#<anchor>` 的 Hint 在 `docs/EARS.md` 里 100% 都能被解析（用 markdown parser 拿所有 heading slug 集合做字符串匹配），任何不可解析的 anchor 必须让测试 fail 并指出文件:行号 + anchor 名。

[AC-3] THE SYSTEM SHALL 在 docs/EARS.md 的 4 个 heading（"五种模板 + 无条件基线"对应 `{#five-templates}`、"模糊词黑名单"对应 `{#vague-words}`、"语义质量门禁（新增 10 条）"对应 `{#semantic-quality-gates}`、"bugfix.md（bugfix spec）"对应 `{#bugfix-spec-bugfixmd}`）各加 1 个 Pandoc 风格 `{#english-slug}` 显式锚点，与现有 linter Hint 字符串里的 slug 对齐，确保 GitHub 渲染的 anchor 链接 100% 点击可定位。

[AC-4] WHERE `free-kiro lint <name>` 输出含 `Hint:` 字段时 THE SYSTEM SHALL 让该 Hint 在 GitHub 渲染的 docs/EARS.md 上点击可定位到对应 heading；端到端验证：跑 1 次 `free-kiro spec new <test-spec>` + 1 次 `free-kiro lint <test-spec>` 不报 `ears-ac-missing-id`，且 1 个新测试 `TestLintHintsResolve` 对 `internal/lint/requirements.go`、`internal/lint/bugfix.go`、`internal/lint/quality.go` 全部 Hint 字符串 100% 通过 docs/EARS.md heading slug 集合校验。

## Out of Scope

- 不修改 `internal/spec/templates/bugfix.md.tmpl`（bugfix 当前不在 `ears-ac-missing-id` 规则覆盖范围内，且其 `[AC-N]` 编号语义与 feature spec 不同）。
- 不引入新的 lint rule；只修复 Hint 锚点解析 + 模板示范行格式。
- 不为 `internal/lint/` 其它 `docs/*.md#anchor` 引用做批量迁移；本 spec 只修与 EARS 文档相关的 Hint 锚点。
- 不动 free-kiro CLI 输出格式；只确保 Hint 字段里的 anchor 在 docs/EARS.md 上可解析。
- 不引入 markdown 解析新依赖；优先使用 Go stdlib + 已有依赖，或自实现 heading slug 提取（若 goldmark 不在 go.mod 则走 stdlib 方案）。