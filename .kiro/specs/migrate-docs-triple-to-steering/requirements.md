# migrate-docs-triple-to-steering

把 `docs/CODING_STYLE.md` + `docs/AGENT_RULES.md` + `docs/POLICY.md` 三件套迁移到
`.kiro/steering/` 下，作为项目级 steering 文档（mode: always）注入 IDE 指令文件。
统一约束文档的注入链路，消除文档位置与文档语义的错位。

## Background

仓库当前有 3 份位于 `docs/` 的约束文档（`CODING_STYLE.md` ~700 行、
`AGENT_RULES.md` ~120 行、`POLICY.md` ~94 行），加上项目级 steering
`.kiro/steering/{product,structure,tech}.md`（mode: always 注入 IDE）。

根因问题：**code style 的归属位置错了**——它是项目事实而非 PHP 反向引用，
应该归到 `.kiro/steering/` 体系，通过 steering inject 自动进入 IDE 上下文，
而不是依赖 IDE 指令文件里 hand-written 链接。

### 当前问题

1. **注入策略分裂**：`docs/CODING_STYLE.md` / `docs/AGENT_RULES.md` /
   `docs/POLICY.md` 依赖 5 份 IDE 指令文件（CLAUDE.md / AGENTS.md /
   .cursorrules / .cursor/rules/free-kiro.md / .kiro/AGENTS.md）+ 4 份 IDE
   模板（`internal/ide/templates/agents_{en,zh}.md` +
   `instructions_{en,zh}.md`）的 hand-written 链接手动同步；项目级 steering
   则由 `free-kiro steering inject` 自动注入。同一类约束有两条注入链路，
   版本漂移风险高。
2. **覆盖重叠与冲突**：`docs/CODING_STYLE.md` 与 `.kiro/steering/tech.md`
   在命名、错误处理、测试上有重叠，且**测试框架口径冲突**
   （CODING_STYLE 默认 testify/assert+require，tech.md 明令"不引入 testify"）；
   覆盖率硬阈值在 POLICY.md（70%/80%）、tech.md（"不强制"）、
   CODING_STYLE.md（"CI 报告非 lint 强制"）三处说法不一致。
3. **既有 spec 方向不匹配**：spec `separate-codestyle-projectspecific`
   方向是「docs/CODING_STYLE.md 内部瘦身」，与本次「迁到 steering」
   的方向不一致，需要明确处置。

## User Stories

- As a free-kiro 贡献者, I want `docs/CODING_STYLE.md` 的 Go 编码规范内容
  迁移到 `.kiro/steering/coding-style.md`，so that 我能在 IDE SessionStart
  时通过 steering inject 自动看到该文档，无需依赖 IDE 指令文件中的手写链接。
- As a free-kiro 贡献者, I want AI agent 协作硬性要求（AGENT_RULES）与
  项目策略（POLICY）迁移到 `.kiro/steering/` 下，so that 所有约束文档
  走同一条 steering inject 注入链路，版本漂移风险可控。
- As a maintainer, I want迁移后所有引用方（IDE 指令文件 / 模板 / spec 文档 /
  配置文件）的路径全部更新，so that `grep -r "docs/CODING_STYLE.md" .`
  残留 0 命中，且 steering inject 底部块包含全部 3 份新文档。
- As a CI reviewer, I want跨文档冲突（testify vs stdlib testing、覆盖率阈值、
  `-race` 必跑、commit 模板、`//nolint` 上限）在迁移过程中一次性统一口径，
  so that 新贡献者不会因文档矛盾产生困惑。

## Acceptance Criteria

### Steering 文件搬迁

- [AC-1] WHEN 开发者阅读 `.kiro/steering/coding-style.md` 时 THE SYSTEM SHALL 在文档内找到原 `docs/CODING_STYLE.md` 的 7 章通用 Go 规范（命名约定、错误处理、并发、接口设计、测试通用部分、注释与文档通用部分、依赖管理通用部分），且 frontmatter 含 `mode: always` 与 `description` 字段，可验证条款数 ≥7。
- [AC-2] WHEN 开发者查阅 free-kiro 对 AI agent 的协作要求时 THE SYSTEM SHALL 在 `.kiro/steering/agent-rules.md` 找到等价内容（零豁免、零死代码、上下文注入、失败处置、零吞错误、硬编码约束），且 frontmatter 含 `mode: always` 与 `description`，可验证条款数 ≥6。
- [AC-3] WHEN 开发者查阅 free-kiro 项目策略时 THE SYSTEM SHALL 在 `.kiro/steering/policy.md` 找到等价内容（覆盖率门槛 ≥70% / ≥80%、TODO owner、`//nolint` ≤5、协议合规、内部依赖路径、commit 格式、函数 ≤50 行、文件 ≤500 行、PR 范围约束），且 frontmatter 含 `mode: always` 与 `description`，可验证条款数 ≥8。
- [AC-4] WHERE `docs/CODING_STYLE.md` / `docs/AGENT_RULES.md` / `docs/POLICY.md` 三份文档存在时 THE SYSTEM SHALL 在迁移完成后这三份文档被删除，残留文件数为 0。

### 冲突统一

- [AC-5] WHEN 测试框架口径冲突时 THE SYSTEM SHALL 在 `coding-style.md` 与 `tech.md` 中统一为「不引入 testify，仅用 stdlib `testing`，table-driven 测试优先」，以 tech.md 为准，且 `grep testify go.mod go.sum` 残留命中数为 0。
- [AC-6] WHEN 覆盖率阈值口径冲突时 THE SYSTEM SHALL 在 `policy.md` 保留硬阈值（≥70% 全包 / ≥80% 新增修改），`tech.md` 与 `coding-style.md` 删除冲突措辞并指向 `policy.md`，且 `grep -rE "70%|≥70%" .kiro/steering/coding-style.md .kiro/steering/tech.md` 残留命中数为 0。

### 引用更新

- [AC-7] WHEN IDE SessionStart hook 注入 AI 上下文时 THE SYSTEM SHALL 在 5 份 IDE 指令文件（CLAUDE.md / AGENTS.md / .cursorrules / `.cursor/rules/free-kiro.md` / `.kiro/AGENTS.md`）的硬约束摘录段包含对新 `.kiro/steering/coding-style.md` / `agent-rules.md` / `policy.md` 的引用，替代原 `docs/` 三件套路径，且可验证引用更新文件数 ≥5。
- [AC-8] WHEN 4 份 IDE 模板（`internal/ide/templates/agents_en.md` / `agents_zh.md` / `instructions_en.md` / `instructions_zh.md`）被 `go:embed` 编进 binary 时 THE SYSTEM SHALL 模板中对 `docs/` 三件套的链接全部更新为 `.kiro/steering/` 三件套，且 `grep -r "docs/CODING_STYLE" internal/ide/templates/` 残留命中数为 0。
- [AC-9] THE SYSTEM SHALL 在所有引用方（11 个 spec 文档 / .golangci.yml / Makefile / CONTRIBUTING.md / `.github/workflows/ci.yml` / `internal/ide/ide_test.go` / `internal/visualize/server_handlers.go`）完成批量引用更新，且 `grep -rE "docs/(CODING_STYLE|AGENT_RULES|POLICY)\.md" . --exclude-dir=separate-codestyle-projectspecific --exclude-dir=migrate-docs-triple-to-steering` 残留命中数为 0（`.kiro/specs/migrate-docs-triple-to-steering/` 与 `separate-codestyle-projectspecific/` 的描述性引用豁免；前者为本次迁移 spec 描述"要删哪些文件"的上下文，后者即将被 superseded）。
- [AC-10] WHEN `free-kiro steering inject` 重新执行时 THE SYSTEM SHALL 在 5 份 IDE 指令文件的底部 `<!-- free-kiro-managed:start -->` ... `<!-- free-kiro-managed:end -->` 块内能看到 `coding-style.md` / `agent-rules.md` / `policy.md` 的内容，且可验证注入内容文件数 ≥5。
- [AC-11] THE SYSTEM SHALL 在改动完成后 `free-kiro lint migrate-docs-triple-to-steering` 退出码为 0，`make lint-go` 退出码为 0。

## Out of Scope

- 不重写 `coding-style.md` 第 1-7 章内容（仅搬迁 + 冲突统一）。
- 不调整 `.golangci.yml` 的 linter 配置规则。
- 不为新 steering 文档发明新的 frontmatter schema。
- 不回头改造历史代码以对齐新口径。
- 不修改 `separate-codestyle-projectspecific` spec 的设计（仅加 superseded 声明）。
- 不为 `separate-codestyle-projectspecific` 目录下的文档更新引用路径。
