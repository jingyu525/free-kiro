# example-test-driven-spec-evolve — Design

## Architecture

本 spec 是**叙事型 meta doc**，不涉及任何代码 / 模块 / 数据结构改动。三件套
按 `requirements / design / tasks` 分别承担 3 段叙事，**互不重复、各有侧重**：

| 文件 | 叙事主题 | 内容指向 |
|---|---|---|
| `requirements.md` | 契约：本 example 必须满足的 4 条 AC | 用 EARS 句式表达，**通过 lint 验证** |
| `design.md`（本文件） | **漏想发现点 + 同步策略**：抽象出方法论 | 引用 `unify-ears-template-and-hints` 的真实过程 |
| `tasks.md` | **任务重排**：方法论落到 wave 调度 | 用 `[deps:]` 展示 "测试先行 → 实现 → 文档同步" |

## Components

### 1. 漏想发现点（AC 关联）

参考 `unify-ears-template-and-hints` 的真实轨迹：

| 阶段 | 漏想内容 | 暴露方式 |
|---|---|---|
| 写 spec 时 | AC-3 字面写"3 处 heading 各加 1 个锚点"，实际 `internal/lint/bugfix.go:44` 引用 `#bugfix-spec-bugfixmd`（第 4 个锚点） | 仅扫文档看不出来；只列了 3 个 slug |
| 跑测试时 | `TestLintHintsResolve` 扫所有 `internal/lint/*.go` 的 `Hint:` 字符串 → 报告 `bugfix.go:44 Hint ... anchors "bugfix-spec-bugfixmd" which is not in docs/EARS.md` | **测试驱动发现** |
| 修复时 | 同步扩 AC-3 字面（3 → 4）+ 在 docs/EARS.md `## bugfix.md（bugfix spec）` heading 加 `{#bugfix-spec-bugfixmd}` 锚点 | drift 检测无 → spec 仍合法 |

### 2. 同步策略

测试驱动发现漏想后的同步动作清单（来自本项目实测）：

1. **先看 spec 是否还有效**：扩 AC 字面（如 `3 个` → `4 个`）+ 扩 Out of Scope 边界
2. **再修实现**：docs/EARS.md 加锚点；必要时改 template / 源代码
3. **再跑测试**：positive path 必须 PASS
4. **再跑 spec drift 检测**：`free-kiro spec status <name>` 看 `drift: none`
5. **最后 commit**：1 个 squash commit，message 引用"测试暴露 → 同步扩展"链条

### 3. 任务重排（tasks.md 视角）

`unify-ears-template-and-hints` 的 5 个 tasks 实际执行顺序被**测试结果反向重排**：

```
原始 wave 顺序:
  Wave 1 (#1 模板 / #2 docs 锚点 / #3 测试) → 全并行
  Wave 2 (#4 端到端验证) → 依赖 Wave 1
  Wave 3 (#5 commit + complete) → 依赖 Wave 2

实际执行顺序:
  Wave 1 #3 测试 先写 → 跑 → FAIL（暴露 #2 漏一个锚点）
  → 紧急扩 #2 docs (追加锚点)
  → 紧急扩 AC-3 字面（spec sync / baseline re-capture）
  → Wave 2 #4 验证 → PASS
  → Wave 3 #5 commit + complete
```

**关键差异**：测试**前置**于实现暴露漏想，而不是"实现后再补测试"。这是 TDD 的
真正价值——把 spec 漏洞在 commit 前堵住。

## Data Model

无。本 spec 不引入任何数据结构。

## Implementation Details

### 测试驱动的判定信号

`TestLintHintsResolve` 这类"反向引用一致性测试"是本流程的核心。本仓库已具备的
同类工具：

| 测试 | 作用 |
|---|---|
| `internal/lint/ears_test.go:TestEARSRe_AllTemplates` | 反向印证 6 种 EARS 模板 |
| `internal/lint/hints_test.go:TestLintHintsResolve` | 反向印证 Hint 锚点一致性 |
| `internal/spec/spec_lifecycle.go` drift 检测 | 反向印证 spec 与实现脱钩 |

**所有类似测试都能在 spec 阶段暴露漏想**，因为它们的输入是"实现侧引用"，输出
是"spec 侧应承"。任何 mismatch 都是 spec 漏想。

### 复用到其它场景的 checklist

把本流程套到新场景时的 8 步 checklist：

1. 建 spec（feature / bugfix）
2. 写 requirements.md AC
3. 跑 `free-kiro lint <name>` 通过 advance gate
4. **Wave 1：先写测试**（不是先写实现）
5. **跑测试 → 看是否 FAIL**（FAIL 通常是好消息：暴露了漏想）
6. 同步扩 AC / 改 docs / 改代码
7. 重跑测试 → PASS
8. spec approve → start → 实现 → complete

## Error Handling

无。本 spec 是示例文档，不处理错误。

## Testing Strategy

| 层级 | 覆盖 | 文件 |
|---|---|---|
| Lint | `free-kiro lint example-test-driven-spec-evolve` exit 0 | `free-kiro` CLI |
| Task | `free-kiro task list example-test-driven-spec-evolve` 展示 wave | `free-kiro` CLI |
| Drift | `free-kiro spec status` 报 `drift: none` | `free-kiro` CLI |

不写新 Go 测试。本 spec 的"测试"是 free-kiro CLI 自带的 3 个命令。

## Migration / Rollout

无 schema / 代码迁移。本 spec 在仓库 `.kiro/specs/example-test-driven-spec-evolve/`
目录下新增 4 个文件：

```
.meta.json           # 自动生成
requirements.md      # 4 条 AC 描述本 example 的契约
design.md            # 漏想发现点 + 同步策略（本文档）
tasks.md             # 任务重排 + 复用 checklist
```

回滚：删除该目录 + 1 个 commit revert。

## 引用

- `unify-ears-template-and-hints` 实现 commit：`26b25f7 fix(spec): 统一 EARS 模板 AC 前缀、docs 锚点与 lint Hint 解析`
- `internal/lint/hints_test.go:TestLintHintsResolve`（暴露漏想的测试）
- `internal/lint/bugfix.go:44`（被暴露的 broken Hint）
- `docs/EARS.md` "## bugfix.md（bugfix spec）" heading（追加 `{#bugfix-spec-bugfixmd}` 的位置）
- `unify-ears-template-and-hints/requirements.md` AC-3（原文："3 个 heading 改为 4 个 heading"）