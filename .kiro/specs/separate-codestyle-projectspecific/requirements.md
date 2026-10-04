# separate-codestyle-projectspecific

> **Superseded by [`migrate-docs-triple-to-steering`](../migrate-docs-triple-to-steering/)**
> —— 本 spec 的方向（拆分 `docs/CODING_STYLE.md` 项目特定内容到
> `docs/AGENT_RULES.md` + `docs/POLICY.md`）已被升级为「整体迁移到
> `.kiro/steering/`」方案。本 spec 历史保留作为设计讨论记录，**不再推进**。

把 `docs/CODING_STYLE.md` 中混入的项目特定策略从"Go 编码规范"载体中分离出去，
让文档回到 Go 社区通用规范的本职。

## Background

`docs/CODING_STYLE.md` 顶部虽已声明"适用对象：free-kiro 项目及其衍生 Go 代码"，
但其第 5/6/7/8 章混入了不应被冠以"Go 编码规范"的内容：

- **5.1** 测试覆盖率硬阈值 ≥ 70% 全包 / ≥ 80% 新增修改 — Go 社区无此规范
- **6.2** TODO 注释带 owner（`// TODO(jingyu): ...`）— Go 社区默认裸 `// TODO`
- **7.2** 协议合规禁止 GPL 系传染协议 — 项目特定
- **7.2** 内部依赖路径 `github.com/jingyu525/<repo>` — 项目特定
- **8.2** commit message 中文 — Go 社区标准为英文
- **8.2** 函数 ≤ 50 行、文件 ≤ 500 行 — 项目特定
- **8.6** 零 `//nolint` 豁免政策 — 项目特定政策
- **8.7** 零死代码政策 — 项目特定政策

第 8 章标题虽为"AI agent 协作"，但放在名为"Go 编码规范"的文档里，仍然会让
读者把它当作 Go 规范学习。需要把这类项目策略整体下放到独立文档。

## User Stories

- As a Go 开发者阅读本仓库的 Go 编码规范, I want docs/CODING_STYLE.md
  仅承载 Go 社区通用的编码规范, so that 我能放心地把它当作 Go 风格参考，
  而不会被项目特定策略误导。
- As a free-kiro 贡献者, I want AI agent 协作硬性要求与项目策略集中到
  独立文档 (docs/AGENT_RULES.md 与 docs/POLICY.md), so that 我能快速定位
  "这个仓库对 AI 与对项目自身的额外要求"。
- As a 维护者接手新仓库时, I want CLAUDE.md 顶部摘录明确指向 AGENT_RULES.md
  与 POLICY.md, so that IDE SessionStart hook 注入 AI 上下文时能正确分流。

## Acceptance Criteria

[AC-1] WHEN 开发者阅读 docs/CODING_STYLE.md 时 THE SYSTEM SHALL 仅看到 7 章通用 Go 编码规范（命名约定、错误处理、并发、接口设计、测试通用部分、注释与文档通用部分、依赖管理通用部分），文档不再包含第 8 章"AI agent 协作"。
[AC-2] WHEN 开发者阅读 docs/CODING_STYLE.md 时 THE SYSTEM SHALL 在文档顶部 50 行内看到声明"本文档仅覆盖 Go 社区通用规范"，且声明中包含 docs/AGENT_RULES.md 与 docs/POLICY.md 的相对路径链接。
[AC-3] WHEN 开发者查阅 AI agent 协作硬性要求（原 docs/CODING_STYLE.md 第 8.1 节）时 THE SYSTEM SHALL 在 docs/AGENT_RULES.md 中找到等价条款，条款数 ≥ 5 条。
[AC-4] WHEN 开发者查阅原 docs/CODING_STYLE.md 第 8.3 / 8.4 / 8.5 节（上下文注入、失败处置、由工具强制）时 THE SYSTEM SHALL 在 docs/AGENT_RULES.md 中找到等价内容，3 节齐全。
[AC-5] WHEN 开发者查阅 free-kiro 零豁免政策（原第 8.6 节）时 THE SYSTEM SHALL 在 docs/AGENT_RULES.md 中找到等价条款，条款数 ≥ 4 条。
[AC-6] WHEN 开发者查阅 free-kiro 零死代码政策（原第 8.7 节）时 THE SYSTEM SHALL 在 docs/AGENT_RULES.md 中找到等价条款，条款数 ≥ 4 条。
[AC-7] WHEN 开发者查阅覆盖率门槛、TODO 注释 owner、协议合规、内部依赖路径、commit message 格式、函数与文件行数上限等 free-kiro 项目策略时 THE SYSTEM SHALL 在 docs/POLICY.md 中找到全部条款，条款数 ≥ 6 条。
[AC-8] WHERE docs/CODING_STYLE.md 原第 5.1 节包含覆盖率硬阈值（≥ 70% / ≥ 80%）时 THE SYSTEM SHALL 在拆分后该阈值从 CODING_STYLE.md 完全移除并 1:1 迁移到 docs/POLICY.md，人工 `grep` 复核 ≥ 70% 与 ≥ 80% 在 CODING_STYLE.md 残留 0 命中。
[AC-9] WHERE docs/CODING_STYLE.md 原第 6.2 节包含 TODO owner 示例（`// TODO(jingyu):`）时 THE SYSTEM SHALL 在拆分后该规则从 CODING_STYLE.md 完全移除并 1:1 迁移到 docs/POLICY.md，`grep "TODO(jingyu)" docs/CODING_STYLE.md` 残留 0 命中。
[AC-10] WHERE docs/CODING_STYLE.md 原第 7.2 节包含协议合规（GPL 禁止）与内部依赖路径（`github.com/jingyu525/<repo>`）时 THE SYSTEM SHALL 在拆分后这 2 条规则从 CODING_STYLE.md 完全移除并 1:1 迁移到 docs/POLICY.md，`grep -E "GPL|jingyu525"` 在 CODING_STYLE.md 残留 0 命中。
[AC-11] WHEN IDE SessionStart hook 注入 AI 上下文时 THE SYSTEM SHALL 在 CLAUDE.md 的硬约束摘录段（位于"编码规范"小节内）包含至少 1 处对 docs/AGENT_RULES.md 的引用与至少 1 处对 docs/POLICY.md 的引用。
[AC-12] IF docs/CODING_STYLE.md / .golangci.yml / Makefile / CONTRIBUTING.md / .kiro/AGENTS.md / docs/STEERING.md 中存在指向被移除章节的链接或锚点 THEN THE SYSTEM SHALL 在本次改动中同步更新链接目标，不出现悬空引用（人工 grep 复核 0 命中）。
[AC-13] THE SYSTEM SHALL 在改动完成后 `free-kiro lint separate-codestyle-projectspecific` 以退出码 0 通过，且 `make lint-go` 不因此次文档改动出现新 ERROR。

## Out of Scope

- 不重写 docs/CODING_STYLE.md 第 1–7 章的通用 Go 规范内容，只做"删项目特定条款"
  与"顶部声明调整"。
- 不修改 .golangci.yml 的 linter 配置规则（覆盖率、errcheck 等规则集不变）。
- 不修改 Makefile 的 lint 目标。
- 不为 docs/AGENT_RULES.md 与 docs/POLICY.md 增加新的强制工具链（仅 Markdown）。
- 不强制把现有 Go 代码重新对齐到新拆分后的规范；拆分是文档结构变更，不回头
  改造历史代码。
- 不引入新的 CI 校验文档完整性的脚本（spec 阶段不发明工具，靠人工 review）。
