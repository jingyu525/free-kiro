# example-test-driven-spec-evolve

把 `unify-ears-template-and-hints`（phase=done）的端到端实现过程作为一份
**方法论示例 spec** 永久保留，供未来 agent / 工程师复用"先写测试暴露 spec
漏想 → 同步扩展 → 重测 PASS"的标准模板。

这是**示例文档**而非功能 spec：AC 描述的不是"要构建什么"，而是"这份
example spec 必须满足的契约"——即"它本身要够资格作为模板被复用"。

## User Stories

- As a 未来维护者 I want 翻阅 `.kiro/specs/` 时能直接定位到一份完整的
  "测试驱动 spec 演进"端到端流程参考 so that 我不用重新发明方法论就能套用
  同一套做法。
- As a 未来 AI agent I want 拿到 `example-test-driven-spec-evolve` 的
  requirements / design / tasks 三件套后能在 5 分钟内理解 "写测试 → 暴露漏
  想 → 扩 spec → 重测 PASS"的标准循环 so that 我在面对类似场景时能直接套
  模板。

## Acceptance Criteria

[AC-1] WHEN 工程师翻阅 `.kiro/specs/` 下历史 spec 时 THE SYSTEM SHALL 能从
`example-test-driven-spec-evolve` 三件套看到一份完整的「测试驱动 spec 演进」
端到端流程参考，含具体 `unify-ears-template-and-hints` 内对应 commit hash +
文件:行号。

[AC-2] WHEN 工程师阅读本 spec 时 THE SYSTEM SHALL 在 requirements.md / design.md
/ tasks.md 三件套内分别找到「漏想发现点 / 同步策略 / 任务重排」三段叙事，每段
叙事显式引用 `unify-ears-template-and-hints` spec 内对应 AC、文件、行号与
commit。

[AC-3] WHERE 本 spec 完成实现时 THE SYSTEM SHALL 在 commit diff 里只看到
`.kiro/specs/example-test-driven-spec-evolve/` 目录新增（除 `.meta.json` 自动
生成），不修改任何已有代码或文档（含 `.baseline.json` / 其它已 done 的
spec）。

[AC-4] THE SYSTEM SHALL 在本 spec 的 requirements.md 用 ≥ 3 种 EARS 句式
覆盖本 example 的契约，并通过 `free-kiro lint example-test-driven-spec-evolve`
exit code 0 + 0 ERROR；端到端验证包含 `free-kiro task list
example-test-driven-spec-evolve` 看到完整 wave 调度。

## Out of Scope

- 不修改任何已有 spec / 代码 / 文档；本 spec 仅新增示例。
- 不引入新的 lint rule 或 workflow 阶段；只复用现有 spec 状态机。
- 不写新 Go 代码或 markdown 模板；本 spec 的"实现"就是 requirements / design /
  tasks 三件套本身。
- 不为 `example-` 前缀的 spec 创建独立的目录或 lifecycle 钩子；按标准 spec 流
  程走完 draft → done 即可。
- 不在 docs/ 或 steer/ 内放独立「方法论」页面；本 example spec 本身就是那份
  文档。