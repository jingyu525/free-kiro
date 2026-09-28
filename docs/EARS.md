# EARS — 验收标准句式

> EARS（Easy Approach to Requirements Syntax）是一套结构化句式，让需求从"模糊描述"
> 变成"可被机器 + 人双重校验的契约"。free-kiro 的 lint 模块强制所有 requirements.md
> / bugfix.md 必须用 EARS 写验收标准。

## 五种模板 + 无条件基线

| 场景 | 句式 | 触发词 |
|---|---|---|
| **事件驱动** | `WHEN <trigger> THE SYSTEM SHALL <response>` | 某个动作发生 |
| **状态驱动** | `WHILE <state> THE SYSTEM SHALL <behavior>` | 持续在某状态期间 |
| **可选特性** | `WHERE <feature> THE SYSTEM SHALL <behavior>` | 仅在某个 feature flag / option 开启时 |
| **异常** | `UNLESS <exception> THE SYSTEM SHALL <default>` | 例外情况下的默认行为 |
| **复杂条件** | `IF <condition> THEN <event> THE SYSTEM SHALL <response>` | 多条件复合 |
| **无条件基线** | `THE SYSTEM SHALL <requirement>` | 始终成立 |

大小写不敏感（lint 用 `(?i)` 匹配），但建议大写以保持视觉一致。

## 真实示例

### Feature spec（requirements.md）

```markdown
## User Stories

As a user I want to log in so that I can access my dashboard.

## Acceptance Criteria

WHEN the user submits valid credentials THE SYSTEM SHALL redirect to /dashboard.
WHEN authentication fails THE SYSTEM SHALL display an error message within 500 ms.
WHILE the session is active THE SYSTEM SHALL refresh the auth token every 30 minutes.
WHERE two-factor authentication is enabled THE SYSTEM SHALL require a TOTP code.
UNLESS the user is an admin THE SYSTEM SHALL hide the audit log link.
IF the user submits the form 3 times within 1 minute THEN rate limit THE SYSTEM SHALL return 429.
THE SYSTEM SHALL persist user preferences across browser sessions.
```

### Bugfix spec（bugfix.md）

```markdown
## Current Behavior (Defect)

WHEN a user logs in with valid credentials the system redirects to /home instead of /dashboard.

## Expected Behavior (Correct)

WHEN the user submits valid credentials THE SYSTEM SHALL redirect to /dashboard.

## Unchanged Behavior (Regression Prevention)

WHEN authentication fails THE SYSTEM SHALL CONTINUE TO display the error message.
WHEN the session expires THE SYSTEM SHALL CONTINUE TO redirect to /login.
```

## 关键约束

### requirements.md（feature spec）

- 至少一条 AC 用以上六种 EARS 模板之一（缺则 `no-ears` ERROR）
- AC 行里**禁止**残留 `<TODO:...>` 占位符（缺则 `placeholder-ac` ERROR）

### bugfix.md（bugfix spec）

- **Current Behavior** 段**禁止**使用 `THE SYSTEM SHALL`（缺陷是错的，不是"应当"——`defect-uses-shall` ERROR）
- **Expected Behavior** 段必须含 `THE SYSTEM SHALL`（缺则 `no-ears-expected` ERROR）
- **Unchanged Behavior** 段必须含 `THE SYSTEM SHALL CONTINUE TO`（缺则 `missing-unchanged` ERROR；非 CONTINUE 形式报 `no-shall-continue` WARNING）

## 写好 EARS 的实践

1. **触发词要具体**：避免 `WHEN the system is busy THE SYSTEM SHALL ...`——"busy" 不可观测。改 `WHEN CPU usage > 90% for 5 seconds THE SYSTEM SHALL ...`。

2. **响应要可测量**：避免 `THE SYSTEM SHALL respond fast`——"fast" 是模糊词。改 `THE SYSTEM SHALL respond within 200 ms`。

3. **每个 AC 独立一行**：不要把多个 SHALL 塞进一个 bullet。

4. **避免 placeholder-ac**：写文档时用 `<TODO:...>` 标记未决项，但**必须在 lint 前替换为真实内容**。

5. **WHILE 用于持续行为**：WHILE 描述的是持续期间的状态机，与 WHEN（瞬时事件）区分。

6. **WHERE 是 feature flag**：WHERE 应该对应明确的开关（如配置、ab test、user role）。

7. **UNLESS 写默认**：UNLESS 用于说明"例外之外"的行为，等价于默认行为 + 豁免条件。

## 模糊词黑名单（`spec analyze` 检测）

advisory 检查，列出但**不拦截**：

- `etc` — 用具体列表替代
- `and/or` — 拆成两条 AC
- `maybe` / `some` — 给出确定值
- `user-friendly` / `seamless` / `intuitive` — 用具体指标（点击数、响应时间）
- `robust` / `flexible` — 描述容错场景或扩展点
- `tbd` / `todo` — 删掉或补完

## 自动化校验

```bash
# Lint 当前 spec 的 EARS 合规性
free-kiro lint my-spec

# Advisory 一致性分析（vague / 重复 AC / 可追溯性）
free-kiro spec analyze my-spec
```

Lint 失败的 ERROR 会阻止 `spec approve` / `spec generate` 的前向转移。这是 free-kiro
的核心约束：可能性空间收敛。

---

参考：[WORKFLOW.md](WORKFLOW.md)（完整工作流）/ [CLI.md](CLI.md)（lint 命令）