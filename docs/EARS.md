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
- 必须包含 `User Stories` 段（缺则 `no-user-stories` **ERROR**，从 WARNING 升级）

#### 语义质量门禁（新增 10 条）

| Code | 严重级别 | 触发条件 | ❌ 反例 |
|---|---|---|---|
| `ears-etc-list` | ERROR | AC 行内出现 `\b(etc\|and/or)\b` | `WHEN user configures foo THE SYSTEM SHALL validate etc. inputs.` |
| `ears-multi-shall-line` | WARNING | 单行内 SHALL 出现 ≥ 2 次（跨行续行不算） | `WHEN login THE SYSTEM SHALL redirect AND WHEN token expires THE SYSTEM SHALL refresh.` |
| `ears-few-ac` | WARNING | EARSRe 命中行数 < `models.MinAcceptanceCriteria`（默认 3） | 只有 1 条 AC 的 spec |
| `ears-low-template-diversity` | WARNING | 6 种模板命中种类数 < 2 | 8 条 AC 全是 WHEN + ubiquitous,无 WHERE/UNLESS/IF |
| `ears-ac-missing-id` | WARNING | AC 行未匹配 `^\s*-\s*\[AC-\d+\]\s+` | `WHEN foo THE SYSTEM SHALL bar.`（无 `[AC-1]` 前缀） |
| `ears-response-immeasurable` | WARNING | SHALL 之后子串不含数字/时间单位/状态码/百分比/`within`/`at most` | `WHEN user clicks login THE SYSTEM SHALL respond fast.` |
| `ears-trigger-unobservable` | WARNING | WHEN/WHILE 触发词含 `busy\|slow\|normal\|large\|small\|many\|recently\|soon\|fast\|robust\|flexible\|seamless\|intuitive` | `WHEN system is busy THE SYSTEM SHALL ...` |
| `ears-keyword-misuse-while-as-when` | WARNING | WHILE 行状态子串以过去式动词（`logged`/`submitted`/`clicked`/...）结尾 | `WHILE user logs in THE SYSTEM SHALL ...`（应为 WHEN） |
| `ears-passive-response` | WARNING | SHALL 之后子串以 `be`/`is`/`are`/`been` 开头 | `THE SYSTEM SHALL be fast.` |

> 模板多样性检查使用 `WHENRe` / `WHILERe` / `WHERERe` / `UNLESSRe` + `UsesIFTHEN()` 独立判定
> （RE2 不支持跨两个 `.+?\s+` 的反向引用,IF-THEN 必须用 prefix+suffix 拼接）,
> 无条件基线（`THE SYSTEM SHALL`）也算一种模板。

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

## 模糊词黑名单

`etc` 和 `and/or` 由 lint `ears-etc-list` **ERROR** 拦截（直接阻断 advance/approve）；
其余 9 个词仍由 `spec analyze` 以 **info** 提示（advisory，不阻断）。

| 词 | 处理 |
|---|---|
| `etc` | lint ERROR（拒绝） |
| `and/or` | lint ERROR（拒绝） |
| `maybe` / `some` | analyze info |
| `user-friendly` / `seamless` / `intuitive` | analyze info |
| `robust` / `flexible` | analyze info |
| `tbd` / `todo` | analyze info |

## 自动化校验

```bash
# Lint 当前 spec 的 EARS 合规性
free-kiro lint my-spec

# Advisory 一致性分析（vague / 重复 AC / 可追溯性）
free-kiro spec analyze my-spec

# 强制忽略 .baseline.json（严格模式）
free-kiro lint my-spec --strict-baseline
```

Lint 失败的 ERROR 会阻止 `spec approve` / `spec generate` 的前向转移。这是 free-kiro
的核心约束：可能性空间收敛。

## Baseline 白名单机制

历史 spec 在新增规则上线时不应瞬时变红。`.kiro/specs/<name>/.baseline.json`
提供一份该 spec 的 lint 白名单，schema 为：

```json
{
  "schema_version": 1,
  "spec_name": "my-spec",
  "ignored_issues": ["ears-few-ac", "ears-low-template-diversity"]
}
```

- `schema_version`: 必填，当前为 1；不匹配时报错（不静默忽略）
- `spec_name`: 可选；填写时必须与所在目录名一致，防止误粘贴
- `ignored_issues`: Issue Code 数组；命中的 issue 在 `free-kiro lint` 输出中
  仍列出但 prefix 为 `[baseline]`，且不计入 `Gate()` 的 ERROR 阻断集合

`--strict-baseline` flag 临时禁用白名单（用于历史 spec 真正想"重新审视"时）。

空 `ignored_issues` 的 baseline 也是合法的——它让历史 spec 文件存在以接入
机制，同时不放过任何新发现的 issue。

---

参考：[WORKFLOW.md](WORKFLOW.md)（完整工作流）/ [CLI.md](CLI.md)（lint 命令）