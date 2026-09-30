---
name: free-kiro
description: |
  Spec-driven development workflow for AI coding agents. Use this skill when
  the user wants to (a) start or plan a new feature / bugfix as a spec with
  EARS-style requirements, design, and tasks, (b) run quality gates via
  `free-kiro lint`, (c) fetch a PRD from a URL or browser tab to bootstrap a
  spec, (d) inspect watch / serve / report / doctor / upgrade / demo output,
  or (e) add or trigger event-driven hooks in a `.kiro/` workspace.
  Triggers: "spec", "spec-driven", "EARS", "requirements first", "lint",
  "tasks.md waves", "PRD", "from browser", "spec from URL", ".kiro/",
  "wave plan", "free-kiro", "quick spec", "spec sync", "drift", "看板",
  "升级", "demo".
license: MIT
allowed-tools: Bash(free-kiro:*) Read Write Edit Glob Grep
metadata:
  free-kiro-min-version: "0.8.0"
  apps: [claude-code, opencode, codex, codebuddy]
---

# free-kiro

把"先想清楚 → 再动手"做成强约束门禁的规约式开发工作流。AI 助手在用户谈及
spec / EARS / lint / PRD / wave / .kiro / demo / upgrade / drift 等意图时，
自动调用本 skill 包装的 `free-kiro` 二进制，让规格驱动开发的整条流水线
（需求 → 设计 → 任务 → 实现 → 完成）对用户透明可用。

## 当用户出现以下意图时主动调用

- "写个 spec"、"规划下"、"先做个需求分析"、"用 EARS 写验收标准"
- "过一下 lint"、"lint 一下"、"质量门禁"
- "quick 一下"、"免审批"、"Quick Spec"
- "spec sync"、"看 drift"、"消除漂移"
- "分析一下 spec"、"advisory 分析"
- "从这个 PRD 拉个 spec"、"spec new --from-prd"、"把这个 issue 转成 spec"
- "从浏览器抓下来做 spec"（需要 `bsk`，见 [references/prd-fetch.md](references/prd-fetch.md)）
- "看看 wave 怎么排"、"tasks.md 长啥样"、"parallel waves"
- "添加个 hook"、"事件触发"、"PreToolUse 拦截"
- "watch 一下 .kiro/"、"serve 看板"、"跑下 doctor"、"出份 report"
- "升级 free-kiro"、"check 一下最新版本"、"upgrade"
- "跑下 demo"、"5 分钟 onboarding"

## 快速参考（一个屏幕）

```bash
# 一次性：把仓库接入 free-kiro 工作流（写入 .kiro/ + IDE hook）
free-kiro init
free-kiro init --ide claude-code          # 只装 Claude Code 的 hook

# 规格驱动开发（核心）
free-kiro spec new <name> --prompt "<一句话需求>"
free-kiro spec new <name> --from-issue <gh-url>
free-kiro spec new <name> --from-prd <url>
free-kiro spec new <name> --from-browser <url>      # 需要 bsk
free-kiro spec new <name> --prompt "..." --workflow design-first
free-kiro spec new <name> --prompt "..." --type bugfix
free-kiro spec new <name> --prompt "..." --quick
free-kiro spec generate <name> --phase all
free-kiro spec generate <name> --phase requirements     # 单 phase
free-kiro spec generate <name> --force                  # 覆盖已存在
free-kiro spec show <name>                              # 打印 phase 文档
free-kiro spec show <name> --tree                       # ASCII 树形
free-kiro spec list                                     # 全部 spec 列表
free-kiro spec status <name>                            # 漂移 + 进度
free-kiro spec status <name> --json                     # 与 serve 同源 JSON
free-kiro spec analyze <name>                           # advisory 一致性
free-kiro spec next <name>                              # 预言机下一步动作
free-kiro spec sync <name>                              # 改完 spec 重 baseline
free-kiro spec approve <name>                           # 必须在用户显式同意后
free-kiro spec start <name>
free-kiro spec complete <name>
free-kiro spec quick <name> --prompt "<一句话改动>"     # 一次性 + 免审批

# 质量门禁（CI 友好）
free-kiro lint [name]                                 # exit 1 = ERROR
free-kiro lint <name> --strict-baseline                # 禁用白名单

# 上下文注入
free-kiro steering list
free-kiro steering show <name>
free-kiro steering context [--file <path>] [--prompt "<text>"]

# 任务视图
free-kiro task list <name>                             # 并行 wave 视图

# 运维
free-kiro watch --preset reactive                      # .kiro 改动自动 lint
free-kiro watch --preset full                          # lint + status + report
free-kiro watch --command "go test ./..." --command "go build"
free-kiro watch --debounce 1s --verbose --root ./docs
free-kiro serve                                        # 本地 web 看板
free-kiro serve --bind 0.0.0.0 --port 8080 --open
free-kiro doctor [--strict] [--verbose]
free-kiro report                                       # 写 .kiro/REPORT.md
free-kiro report --stdout                              # 输出到 stdout
free-kiro report --output ./REPORT.md
free-kiro status [--json|--human]                      # 聚合 workspace 状态

# 事件钩子
free-kiro hook add --id <id> --event <event> --action-type shell --action "<cmd>"
free-kiro hook list
free-kiro hook run <event> [--file <path>]

# Self-install（管理本 skill）
free-kiro skill install [--app all|claude-code|opencode|codex|codebuddy]
free-kiro skill install --dry-run                     # 只打印不写盘
free-kiro skill path --app claude-code                # 打印安装路径
free-kiro skill version                               # bundle + binary 版本
free-kiro skill show                                  # 已装状态
free-kiro skill update [--check]
free-kiro skill uninstall --app <id>

# Binary 升级
free-kiro upgrade                                     # 检查 + 升级
free-kiro upgrade --check                             # 只检查不下载
free-kiro upgrade --force                             # 强制重装当前版本

# Onboarding
free-kiro demo [--ide claude-code|codebuddy|cursor|continue|opencode|none]
free-kiro demo --no-color                             # 禁 ANSI（CI）
```

## Workflow — Spec 生命周期

详见 [references/spec.md](references/spec.md)。一句话流程：

```
spec new → generate requirements → generate design → generate tasks
        → (用户 approve) → start → 实现（按 wave） → complete
```

变体：Quick Spec（`spec new --quick` 或 `spec quick`）跳过 approve 闸口；
Bugfix Spec（`spec new --type bugfix`）用 `bugfix.md` 替代 `requirements.md`；
Design-First Workflow（`spec new --workflow design-first`）从 design 起手。

每一步都有 `free-kiro spec next <name>` 预言机告诉你下一步该做什么，调用它即可。
lint 在每步进阶时是硬门禁（exit 1），写错就拦。

实现期改了 spec：用 `free-kiro spec sync <name>` 重 baseline，
不要 `spec generate` 重生成已经填了真实内容的文档（会覆盖）。

## Spec 变体与旗标

`free-kiro spec new` 支持三种变体，三个 flag 可叠加：

| Flag | 取值 | 效果 |
|---|---|---|
| `--workflow` | `requirements-first`（默认）/ `design-first` | 决定 generate 阶段顺序 |
| `--type` | `feature`（默认）/ `bugfix` | bugfix 用 `bugfix.md` 替代 `requirements.md`，且 Current Behavior 段禁止 `SHALL` |
| `--quick` | （bool） | 跳过 approve 闸口；`spec quick` 子命令等价 |

三选一互斥的"外取" flag：`--from-issue` / `--from-prd` / `--from-browser`。
无 `<name>` 时从 title 自动生成 kebab-case slug。

## 抓 PRD / Issue / 浏览器页 → spec

详见 [references/prd-fetch.md](references/prd-fetch.md)。

| 输入 | 标志 | 依赖 |
|---|---|---|
| GitHub issue URL | `--from-issue <url>` | `gh` CLI |
| 任意 HTTP(S) HTML/Markdown | `--from-prd <url>` | 仅 `free-kiro`（net/http） |
| JS 渲染的页面 | `--from-browser <url>` | `bsk`（来自 browser-skill） |

## 运维（watch / serve / report / doctor / status / upgrade / demo）

详见 [references/ops.md](references/ops.md)。

| 命令 | 何时用 |
|---|---|
| `free-kiro watch --preset reactive` | 编辑 .kiro/* 时自动跑 lint |
| `free-kiro watch --preset full` | 改动后跑 lint + status + report |
| `free-kiro watch --command "go test ./..."` | 自定义命令集（可重复 `--command`） |
| `free-kiro serve` | 浏览器看 dashboard（SSE 实时 + 5 s 轮询） |
| `free-kiro serve --open` | 启动后调系统浏览器打开 |
| `free-kiro doctor` | 单次自检（环境 / workspace / IDE hook / GitHub release） |
| `free-kiro doctor --strict` | warn 也算 fail（exit 非零） |
| `free-kiro report` | 把当前进度写成 .kiro/REPORT.md（贴 PR 即可见） |
| `free-kiro status --json` | 与 serve /api/summary 同源 JSON |
| `free-kiro upgrade --check` | 仅检查 GitHub latest，不下载 |
| `free-kiro demo --ide claude-code` | 5 分钟端到端 onboarding |

## 事件钩子（hook）

详见 [references/hooks.md](references/hooks.md)。

事件名两套命名（**同时接受**）：

- **free-kiro 内部**：`file.save` / `file.create` / `file.delete` /
  `prompt.submit` / `task.run` / `manual`
- **Kiro 官方 v1**：`SessionStart` / `UserPromptSubmit` / `PreToolUse` /
  `PostToolUse` / `PostFileSave` / `PostFileCreate` / ...

free-kiro 写出的 hook 同时兼容自家 flat shape 与 Kiro v1 信封，可直接被
官方 Kiro IDE 加载。`free-kiro init` 默认写入 `PreToolUse`（lint gate）和
`SessionStart`（spec next）。

## 排错

详见 [references/troubleshooting.md](references/troubleshooting.md)。

| 症状 | 处置 |
|---|---|
| `lint` 报 `no-ears` | requirements 缺 EARS 句式（WHEN/SHALL/WHILE …） |
| `lint` 报 `dependency cycle` | tasks.md 形成环（A↔B），参考 [references/spec.md](references/spec.md) |
| `lint` 报 `dependency dangling` / `self-reference` | 编号跳号或自引用，修 tasks.md |
| `approve` 报 `tasks.md must be generated first` | 先 `spec generate --phase tasks` 再 approve |
| `spec next` 报 `illegal phase transition` | 改完 spec 不要 `generate` 重生成，就地编辑 + `spec sync` |
| `spec status` 报 drift | 实现期改了 AC / 任务数 → `spec sync` 重 baseline |
| `skill show` 报 `sha mismatch` | 重跑 `scripts/compute-skill-sha.sh` 刷 `skill.json` |
| `free-kiro` 命令找不到 | 重跑 `install.sh`，确认 PATH 含 `~/.local/bin` |
| `--from-browser` 报 `bsk not found` | 装 browser-skill，或改用 `--from-prd` |

## 版本与兼容性

- 本 skill bundle 版本与 free-kiro 二进制版本绑定（同一 `git tag` 发布）。
- `metadata.free-kiro-min-version` 标明所需的最低 free-kiro 版本。
- 旧 binary 装新 skill 会触发 `free-kiro skill show` 里的版本不匹配告警（warn，不阻断）。
- 二进制与 bundle 都用 `free-kiro upgrade` / `free-kiro skill update` 单独更新。

## 自检

```bash
free-kiro skill show        # 已装 SKILL.md 路径 / 版本 / app
free-kiro skill version     # bundle + binary 版本
free-kiro doctor            # 全量自检（含 --from-browser 的 bsk 探测）
```

## 退出码契约

| Exit | 含义 |
|---|---|
| 0 | OK |
| 1 | lint ERROR（CI/PreToolUse 拦截信号） |
| 2 | engine error（不可恢复） |
| 3 | usage error（参数错） |