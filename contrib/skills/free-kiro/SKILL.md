---
name: free-kiro
description: |
  Spec-driven development workflow for AI coding agents. Use this skill when
  the user wants to (a) start or plan a new feature / bugfix as a spec with
  EARS-style requirements, design, and tasks, (b) run quality gates via
  `free-kiro lint`, (c) fetch a PRD from a URL or browser tab to bootstrap a
  spec, (d) inspect watch / serve / report / doctor output, or (e) add or
  trigger event-driven hooks in a `.kiro/` workspace.
  Triggers: "spec", "spec-driven", "EARS", "requirements first", "lint",
  "tasks.md waves", "PRD", "from browser", "spec from URL", ".kiro/",
  "wave plan", "free-kiro".
license: MIT
allowed-tools: Bash(free-kiro:*) Read Write Edit Glob Grep
metadata:
  free-kiro-min-version: "0.7.0"
  apps: [claude-code, opencode, codex, codebuddy]
---

# free-kiro

把"先想清楚 → 再动手"做成强约束门禁的规约式开发工作流。AI 助手在用户谈及
spec / EARS / lint / PRD / wave / .kiro/ 等意图时，自动调用本 skill 包装的
`free-kiro` 二进制，让规格驱动开发的整条流水线（需求 → 设计 → 任务 → 实现 →
完成）对用户透明可用。

## 当用户出现以下意图时主动调用

- "写个 spec"、"规划下"、"先做个需求分析"、"用 EARS 写验收标准"
- "过一下 lint"、"lint 一下"、"质量门禁"
- "从这个 PRD 拉个 spec"、"spec new --from-prd"、"把这个 issue 转成 spec"
- "从浏览器抓下来做 spec"（需要 `bsk`，见 [references/prd-fetch.md](references/prd-fetch.md)）
- "看看 wave 怎么排"、"tasks.md 长啥样"、"parallel waves"
- "添加个 hook"、"事件触发"、"PreToolUse 拦截"
- "watch 一下 .kiro/"、"serve 看板"、"跑下 doctor"、"出份 report"

## 快速参考（一个屏幕）

```bash
# 一次性：把仓库接入 free-kiro 工作流（写入 .kiro/ + IDE hook）
free-kiro init

# 规格驱动开发（核心）
free-kiro spec new <name> --prompt "<一句话需求>"
free-kiro spec new <name> --from-issue <gh-url>
free-kiro spec new <name> --from-prd <url>
free-kiro spec new <name> --from-browser <url>      # 需要 bsk
free-kiro spec generate <name> --phase all
free-kiro spec approve <name>                         # 必须在用户显式同意后
free-kiro spec start <name>
free-kiro spec complete <name>

# 质量门禁（CI 友好）
free-kiro lint [name]                                 # exit 1 = ERROR

# 上下文注入
free-kiro steering list
free-kiro steering show <name>
free-kiro steering context [--file <path>] [--prompt "<text>"]

# 任务视图
free-kiro task list <name>

# 运维
free-kiro watch --preset reactive                     # .kiro 改动自动 lint
free-kiro serve                                       # 本地 web 看板
free-kiro doctor [--strict]
free-kiro report                                      # 写 .kiro/REPORT.md

# 事件钩子
free-kiro hook add --id <id> --event <event> --action-type shell --action "<cmd>"
free-kiro hook run <event>

# Self-install（管理本 skill）
free-kiro skill install [--app all|claude|opencode|codex|codebuddy]
free-kiro skill update [--check]
free-kiro skill show
free-kiro skill uninstall --app <id>
```

## Workflow — Spec 生命周期

详见 [references/spec.md](references/spec.md)。一句话流程：

```
spec new → generate requirements → generate design → generate tasks
        → (用户 approve) → start → 实现（按 wave） → complete
```

每一步都有 `kiro spec next <name>` 预言机告诉你下一步该做什么，调用它即可。
lint 在每步进阶时是硬门禁（exit 1），写错就拦。

## 抓 PRD / Issue / 浏览器页 → spec

详见 [references/prd-fetch.md](references/prd-fetch.md)。

| 输入 | 标志 | 依赖 |
|---|---|---|
| GitHub issue URL | `--from-issue <url>` | `gh` CLI |
| 任意 HTTP(S) HTML/Markdown | `--from-prd <url>` | 仅 `free-kiro`（net/http） |
| JS 渲染的页面 | `--from-browser <url>` | `bsk`（来自 browser-skill） |

三选一互斥。无 `<name>` 时从 title 自动生成 kebab-case slug。

## 运维（watch / serve / report / doctor）

详见 [references/ops.md](references/ops.md)。

| 命令 | 何时用 |
|---|---|
| `free-kiro watch --preset reactive` | 编辑 .kiro/* 时自动跑 lint |
| `free-kiro watch --preset status` | 改动后打印当前 spec 状态 |
| `free-kiro serve` | 浏览器看 dashboard（SSE 实时） |
| `free-kiro doctor` | 单次自检（环境 / workspace / IDE hook） |
| `free-kiro report` | 把当前进度写成 .kiro/REPORT.md |

## 事件钩子（hook）

详见 [references/hooks.md](references/hooks.md)。

事件清单：`PreToolUse`（写文件前 lint 拦截）、`SessionStart`（引导当前 spec）、
`PostToolUse`、`Stop`、`UserPromptSubmit`。`free-kiro init` 默认写入
`PreToolUse`（lint gate）和 `SessionStart`（spec next）。

## 排错

详见 [references/troubleshooting.md](references/troubleshooting.md)。

| 症状 | 处置 |
|---|---|
| `lint` 报 `no-ears` | requirements 缺 EARS 句式（WHEN/SHALL/WHILE …） |
| `lint` 报 `dependency cycle` | tasks.md 依赖写错，参考 [references/spec.md](references/spec.md) |
| `free-kiro` 命令找不到 | 重跑 `install.sh`，确认 PATH 含 `~/.local/bin` |
| `--from-browser` 报 `bsk not found` | 装 browser-skill，或改用 `--from-prd` |
| `kiro spec next` 报 `illegal phase transition` | 改完 spec 不要 `generate` 重生成，就地编辑 + `spec sync` |

## 版本与兼容性

- 本 skill bundle 版本与 free-kiro 二进制版本绑定（同一 `git tag` 发布）。
- `metadata.free-kiro-min-version` 标明所需的最低 free-kiro 版本。
- 旧 binary 装新 skill 会触发 `free-kiro skill show` 里的版本不匹配告警（warn，不阻断）。

## 自检

```bash
free-kiro skill show        # 打印已装的 skill 路径 / 版本 / app
free-kiro doctor            # 跑全量自检（含 --from-browser 的 bsk 探测）
```

## 退出码契约

| Exit | 含义 |
|---|---|
| 0 | OK |
| 1 | lint ERROR（CI/PreToolUse 拦截信号） |
| 2 | engine error（不可恢复） |
| 3 | usage error（参数错） |