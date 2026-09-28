# 工作流图解

> free-kiro 把 Kiro 的 spec-driven workflow 拆成 6 个明确的状态 + 强约束 lint 门禁。
> 这是状态机的图解和"为什么这么设计"的说明。

## 状态机全景

```
                        ┌─────────────────┐
                        │     DRAFT       │ ← spec new
                        └────────┬────────┘
                                 │ spec generate (force overwrite OK)
                  ┌──────────────┴──────────────┐
                  ▼                             ▼
        ┌─────────────────┐           ┌─────────────────┐
        │  REQUIREMENTS   │◄─────────►│     DESIGN      │
        │  requirements.md │           │    design.md     │
        │  (or bugfix.md)  │           │                 │
        └────────┬────────┘           └────────┬────────┘
                 │                             │
                 │      spec generate           │
                 └──────────────┬──────────────┘
                                ▼
                       ┌─────────────────┐
                       │     TASKS       │
                       │    tasks.md     │
                       └────────┬────────┘
                                │ spec approve (lint gate)
                                ▼
                       ┌─────────────────┐
                       │    APPROVED     │  (baseline captured)
                       └────────┬────────┘
                                │ spec start
                                ▼
                       ┌─────────────────┐
                       │  IMPLEMENTING   │
                       └────────┬────────┘
                                │ spec complete
                                ▼
                       ┌─────────────────┐
                       │      DONE       │  (terminal)
                       └─────────────────┘
```

## 三种规划顺序（workflow 变体）

### Requirements-First（默认）

```
draft → requirements.md → design.md → tasks.md → approved
```

最常用：先把"做什么"讲清楚，再设计"怎么做"，再拆任务。

### Design-First

```
draft → design.md → requirements.md → tasks.md → approved
```

适合有强技术约束的场景：先确定架构可行，再回头写需求。

### Quick Spec（免审批）

```
draft → requirements.md → design.md → tasks.md → implementing
                              (跳过 approved 仪式)
```

一次性生成 + 跳过人工签字。lint 仍生效。小改动用。

## Spec 类型（type 变体）

### Feature（默认）

第一份 planning 文档 = `requirements.md`，用 EARS 句式。

### Bugfix

第一份规划文档 = `bugfix.md`，三段式契约：
- **Current Behavior**（缺陷描述）：**禁止**使用 `THE SYSTEM SHALL`（缺陷是错的）
- **Expected Behavior**（正确行为）：必须含 `THE SYSTEM SHALL`
- **Unchanged Behavior**（回归预防）：必须含 `THE SYSTEM SHALL CONTINUE TO`

## Lint Gate（核心约束）

每个前向转移都被 lint 拦截。规则分两层：

**ERROR（拦截进阶）：**
- `no-ears` — requirements.md 缺 EARS 句式
- `placeholder-ac` — AC 行里残留 `<TODO:...>` 占位符
- `defect-uses-shall` — bugfix Current Behavior 用了 SHALL
- `missing-expected` — bugfix 缺 Expected Behavior
- `missing-unchanged` — bugfix 缺 Unchanged Behavior
- `no-ears-expected` — bugfix Expected Behavior 缺 SHALL
- `self-dep` / `dangling-dep` / `cycle` — tasks 依赖图错误

**WARNING（不拦截）：**
- `no-user-stories` — 缺 User Stories section
- `no-shall-continue` — bugfix Unchanged 没用到 CONTINUE TO 形式
- `missing-current` — bugfix 缺 Current Behavior
- `missing-tasks` / `missing-design` — 文档还没写
- `empty-tasks` — tasks.md 没解析出任务

为什么这样分层：
- **ERROR** 拦截的是结构错误（错的、缺的），不修复就无法正确流转
- **WARNING** 是质量提示（建议有，但作者有最终决定权）

完整 EARS 句式参考 [EARS.md](EARS.md)。

## Drift 漂移检测

approve 时 `baseline` 快照：

```json
{ "ac_count": 3, "task_count": 3, "design_sections": 6 }
```

之后任何 `status` 都对比 `current`：

- 无 drift → `drift: []`
- 有 drift → 列出每个 key 的 baseline vs current vs delta

合法编辑后用 `spec sync` 重置 baseline。

## Wave 并行调度

`tasks.md` 每行格式：

```
- [ ] #N Title
- [x] #N Title  (done)
- [ ] #N Title [deps: #N1,#N2]
```

`task list <spec>` 把依赖图拓扑分层：

```
Wave 1: 无依赖的任务 → 可并行
Wave 2: 仅依赖 Wave 1 的任务 → Wave 1 完成后并行
Wave 3: ...
```

DFS 三色染色检测循环（white / gray / black），发现环就 ERROR 拦截。

## 退出码契约（IDE hook 友好）

```
0  成功
1  lint ERROR（可被 IDE 拦截）
2  引擎错误（workspace 缺失 / 非法 phase 转移 / IO 等）
3  用户输入错误（缺参数 / 名称冲突）
```

**关键**：Claude Code / CodeBuddy 的 `PreToolUse` hook 必须 `exit 2` 才拦截写入，所以配置是：

```bash
free-kiro lint || exit 2
```

这才是"门禁"——`exit 1` 只是普通失败，IDE 会继续。

## 全流程示例

```bash
# 1. 项目初始化
cd your-project
free-kiro init

# 2. 写需求
free-kiro spec new user-auth --prompt "add JWT-based user login"
free-kiro spec generate user-auth --phase all
# 编辑 requirements.md：写 User Story + EARS 句式

# 3. 过 lint
free-kiro lint user-auth
# exit 0 即 OK；exit 1 即有 ERROR，看输出修

# 4. 设计 + 拆任务
# 编辑 design.md + tasks.md
free-kiro spec generate user-auth --phase all  # 重新生成 design + tasks（requirements 不动）

# 5. wave 视图
free-kiro task list user-auth
# 看完决定实现顺序

# 6. 审批
free-kiro spec approve user-auth

# 7. 开始实现
free-kiro spec start user-auth

# 8. 实现期间，agent 可随时跑 status / next 防断线
free-kiro spec next user-auth

# 9. 完成
free-kiro spec complete user-auth
```

---

参考：[CLI.md](CLI.md)（命令详细参数）/ [EARS.md](EARS.md)（句式参考）/ [STEERING.md](STEERING.md)（项目约定注入）