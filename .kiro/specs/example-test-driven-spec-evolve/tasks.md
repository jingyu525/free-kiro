# example-test-driven-spec-evolve — Tasks

每完成一条把 `[ ]` 改成 `[x]`；wave 视图用 `free-kiro task list example-test-driven-spec-evolve`。

本 tasks 复刻 `unify-ears-template-and-hints` 的 5-task 节奏，但 **#1 改为先写测试**（"再写实现"），体现 "测试驱动发现漏想" 的核心动作。

## Wave 1 — 测试先行（暴露漏想）

- [x] #1 写 tasks.md（本文件）：把 5 个 task 的 wave 调度画出来，对比 unify-ears-template-and-hints/tasks.md 的 wave 顺序；显式标注"测试先行"vs"实现先行"的差异 [no deps]
- [x] #2 跑 `free-kiro lint example-test-driven-spec-evolve` 验证 requirements.md 的 4 条 EARS AC 通过（同时充当"反向印证"测试：如果 AC-1 / AC-2 写漏了 slug 引用，lint 会失败）[no deps]
- [x] #3 跑 `free-kiro task list example-test-driven-spec-evolve` 验证本 tasks.md 的 wave 调度图与 AC-4 "完整 wave 调度"承诺一致（如果 task 编号错误或 deps 循环，task list 会报）[no deps]

## Wave 2 — 文档同步（依赖 Wave 1 测试结果）

- [x] #4 把 Wave 1 测试暴露的任何漏想同步写进 design.md 的"漏想发现点 / 同步策略"两节，并加 cross-reference 指向 unify-ears-template-and-hints 的对应 AC 与文件:行号 [deps: #1, #2, #3]

## Wave 3 — 审批上线

- [x] #5 `free-kiro spec approve example-test-driven-spec-evolve` → `free-kiro spec start example-test-driven-spec-evolve` → git commit + push → `free-kiro spec complete example-test-driven-spec-evolve` 进入 done 状态；commit message 引用 unify-ears-template-and-hints 的 commit hash 作为"本流程被复用"的实证 [deps: #4]

## 与 unify-ears-template-and-hints 的对照表

| # | unify-ears-template-and-hints 的 task | 本 example spec 的 task |
|---|---|---|
| #1 | 改 requirements.md.tmpl（实现先行） | 写 tasks.md（任务重排先行） |
| #2 | 改 docs/EARS.md 加 3 个锚点 | 跑 lint（测试先行，暴露 AC 漏想） |
| #3 | 新建 hints_test.go | 跑 task list（测试先行，暴露依赖图问题） |
| #4 | 端到端验证 | 同步 design.md |
| #5 | spec complete + commit | spec complete + commit |

**核心差异**：本 spec 的 Wave 1 没有任何"写代码"动作——全是"写测试 / 跑 lint / 测试代码"动作。这是**测试驱动方法论 spec 化**的关键证据：本仓库有能力在不写新代码的前提下完成一个 spec，靠的就是"测试反向印证"。

## 复用 checklist（与 design.md 同步）

8 步（见 design.md "复用到其它场景的 checklist"）：
1. 建 spec → 2. 写 AC → 3. lint advance gate → 4. Wave 1 先写测试 → 5. 跑测试看 FAIL → 6. 同步扩 AC / 改 docs / 改代码 → 7. 重跑测试 PASS → 8. approve → start → 实现 → complete