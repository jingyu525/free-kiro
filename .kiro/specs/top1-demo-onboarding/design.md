# top1-demo-onboarding — Design

## Architecture

本 spec 把"价值兑现"分三层落地,**严格遵循 .kiro/steering/coding-style.md
的零豁免条款**(不引入 `//nolint`):

```
cmd/free-kiro
  └─ internal/cli (demo subcommand 新增;init 扩 multi-ide)
       └─ internal/ide (扩展 5 种 IDE 的 ID/ConfigPath/InstallHooks)
            └─ internal/spec (engine 行为不变,仅 examples/ 复用)

仓库根
  └─ examples/todo-app/       (NEW — 完整走通的真实样例)
       └─ .kiro/specs/add-task-priority/
            ├─ requirements.md  (lint 全绿)
            ├─ design.md        (lint 全绿)
            └─ tasks.md         (lint 全绿;至少 4 任务、2 wave)

.github/workflows/spec-lint.yml.example   (NEW — CI 模板)
docs/DEMO_VERIFICATION.md              (NEW — 复盘指引)
docs/CI_INTEGRATION.md                 (NEW — CI 接入步骤)
```

不创建新的"service"/"usecase"层(沿用 refactor-go-best-practices 的
架构纪律);不引入接口抽象(≥3 真实实现才抽,本次只扩常量);不引入
新 Go 依赖。

### 三块独立的交付边界

| 子模块 | 文件改动 | 复用现有 |
|---|---|---|
| **A. `free-kiro demo`** | `internal/cli/demo.go` (NEW) | `internal/spec`、`internal/visualize`、`internal/workspace` |
| **B. `init --ide auto` 扩 5 IDE** | `internal/ide/ide.go`(扩)、`internal/ide/ide_test.go`(扩) | 现有 `InstallHooks` / `DetectAll` |
| **C. 真实样例 + README + CI 模板** | `examples/todo-app/`(NEW)、`README.md`、`docs/CI_INTEGRATION.md`(NEW)、`.github/workflows/spec-lint.yml.example`(NEW)、`docs/DEMO_VERIFICATION.md`(NEW) | 现有 `spec_lifecycle` / `lint` 命令 |

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `internal/cli/demo.go` (NEW) | 演示路径:打印 5 步 onboarding 摘要 + 写 marker + 可选打印 hook 片段 | `runDemo(cmd, args)` |
| `internal/ide/ide.go` (扩) | 增 `Cursor`/`Continue`/`OpenCode` 三个 `ID` 常量;扩 `configPathFor` 覆盖 5 种路径;扩 `Parse` 接受 5 种字符串 | `ID`, `Parse(s)`, `DetectAll(home)`, `configPathFor` |
| `internal/ide/ide_test.go` (扩) | 新增 ≥15 个 case:5 IDE × 3 路径 | `TestInstallHooks_*` |
| `examples/todo-app/` (NEW) | 完整走通的 .kiro/ workspace + add-task-priority spec | 复用 free-kiro 现有 spec 命令 |
| `.github/workflows/spec-lint.yml.example` (NEW) | CI 模板 | 复用 `setup-free-kiro` action |
| `docs/CI_INTEGRATION.md` (NEW) | 5 步接入文档 | 文本 |
| `docs/DEMO_VERIFICATION.md` (NEW) | demo 复盘 + 验证记录 | 文本 |

## Data Model

### `internal/ide/ide.go` 新增常量

```go
const (
    ClaudeCode ID = "claude-code"   // 已有
    CodeBuddy  ID = "codebuddy"     // 已有
    Cursor     ID = "cursor"        // NEW — ~/.cursor/settings.json
    Continue   ID = "continue"      // NEW — ~/.continue/config.json (or .continue/settings.json)
    OpenCode   ID = "opencode"      // NEW — ~/.opencode/settings.json
)
```

**关键决策**:5 种 IDE 都沿用现有 `event-keyed` JSON 信封
(详见 `internal/ide/ide.go` 的 `settingsShape` + `freeKiroHooks()`)。
**不**为每种 IDE 写一份 envelope — 这避免重复并保证单一事实源。
**前提条件**:Cursor / Continue / OpenCode 必须支持 Claude Code 风格的
`hooks.{EventName}[].matcher + hooks[]` JSON schema。如果不支持,后续
spec 单独处理(Out of Scope)。

### Cursor 路径约定(2026-Q4)

- `~/.cursor/settings.json` — 主流已确认
- Continue: `~/.continue/config.json` (settings 字段内嵌)

按需挑选其一;为简化,本次只实现 Cursor + Claude Code + CodeBuddy 三种
**已确认路径**的硬支持,Continue / OpenCode 标记为 "wip" 但 `Parse`
仍接受(返回 `DirExists=false`,与 auto 模式兼容)。

### demo 命令数据流

```
$ free-kiro demo [--no-color] [--ide <name>]
  ├─ os.Getwd()                                          # 检查 cwd 必须在 repo root
  ├─ repoRootHasExamples()                                                      # 检查 examples/todo-app 存在
  ├─ writeMarker()                                                              # 写 .kiro/.demostart (idempotent)
  ├─ printSteps()                                                              # 5 行 onboarding 摘要
  ├─ if --ide != none: printHookSnippet(ide)                                   # 打印一段
  └─ if recentMarker(): print "already running" + skip rewrite                 # 30 秒去重
```

### examples/todo-app 的 spec 设计

`add-task-priority` spec 是给"AI coding 工具接入 free-kiro"做的
**最小可演示用例**:

- requirements.md:`WHEN a user creates a task with priority P0 THE
  SYSTEM SHALL sort tasks ascending by priority then create tasks`.
- design.md:描述字段(`priority enum P0/P1/P2`)+ 数据流
  (CLI arg → JSON store → 列表渲染)。
- tasks.md:5 个任务,2 个 wave:
  - Wave 1:`#1 data model + JSON store`、`#2 CLI arg parser`
  - Wave 2:`#3 sort logic`、`#4 list renderer`、`#5 e2e test`
    `[deps: #1, #2, #3]`

这样 `free-kiro task list` 出来 2 wave,可演示并行调度。

## Error Handling

### demo 命令错误

| 场景 | 退出码 | 类型 |
|---|---|---|
| cwd 不在 repo root(找不到 `examples/todo-app`) | 3 | `UsageError` |
| `--ide` 值非法 | 3 | `UsageError` |
| 写 marker 失败(IO) | 2 | `Wrap("demo.writeMarker", err)` |

### init --ide 错误

| 场景 | 退出码 | 类型 |
|---|---|---|
| `--ide` 值不在 5 种 cli 名中 | 3 | `UsageError`(`errors.NewUsage`) |
| auto 检测 0 个 IDE | 0 + stderr 警告 | 函数无错误,但人类可读 |
| 写 settings.json 失败 | 2 | `Wrap("ide.install", err)` |

注意:**auto 检测 0 个 IDE 不视作 error** — workspace 已就绪,IDE 配
置是"加强项"。当前 `internal/cli/init.go:150` 的 stderr 警告已实现,
本次新增的 5 种 IDE 沿用同一模式。

### examples/todo-app 的 spec lint 失败

如果在 demo 现场跑 `free-kiro lint examples/todo-app` 报错,意味
样例 spec 本身没维护好。**不**在 demo 命令里自动修复(违反
free-kiro "先想清楚再动手"原则);改为在 `docs/DEMO_VERIFICATION.md`
里标注"如 lint 失败 → 修复 `examples/todo-app/` 内的 spec"。

## Testing Strategy

### 单元测试(必须)

| 文件 | 新增 case | 目标 |
|---|---|---|
| `internal/cli/demo_test.go` (NEW) | ≥6 个 case:cwd 正确 / cwd 错 / marker 写成功 / 30 秒去重 / `--ide none` / `--ide claude-code` / `--no-color` 无 ANSI | demo 命令全路径 |
| `internal/ide/ide_test.go` (扩) | ≥15 个 case:5 IDE × (Parse OK / configPath OK / DetectAll DirExists logic) | IDE 枚举 |
| `internal/ide/install_hooks_test.go` (NEW 或扩) | ≥3 case:Claude Code 既有 hooks 保留 / Cursor 首次安装 / 5 IDE 全部 upsert | hook 安装 |

### 端到端验证(手工冒烟,记录在 DEMO_VERIFICATION.md)

```bash
# Wave 1-3 完成后跑一次,记录结果
make build
cd examples/todo-app
free-kiro spec status add-task-priority --human        # ✓ 输出 phase + drift
free-kiro task list add-task-priority                 # ✓ 至少 2 wave
cd ../..
free-kiro demo                                          # ✓ 5 行 onboarding + marker
free-kiro demo --ide claude-code                       # ✓ 打印 hook 片段
free-kiro demo --no-color                              # ✓ 0 ANSI 转义
cd /tmp && free-kiro demo                              # ✓ exit 3 + stderr 引导
free-kiro init --ide auto --path /tmp  (在已存在 .kiro 的目录)  # ✓ 注册 hooks
free-kiro doctor                                       # ✓ 7 项全过

# Wave 4:CI 模板验证
cp .github/workflows/spec-lint.yml.example /tmp/test-spec.yml
# 在临时 repo 用 action linter 验证 yaml 合法(可用 actionlint)
actionlint /tmp/test-spec.yml                          # ✓ 0 错误
```

### 性能 / 内存

无性能目标(Out of Scope)。demo 命令只跑一次,无 hot path。

## Migration / Rollout

### 落地顺序(3 wave)

```
Wave 1: examples/todo-app/ 真实样例 + README 重写 + docs/DEMO_VERIFICATION.md
         (纯静态资产 + 文档;0 行 Go 代码改动;review 最快)
Wave 2: init --ide auto 扩 5 IDE + ide_test.go 扩
         (1 个 Go 文件 + 1 个测试;改动小、复用高)
Wave 3: free-kiro demo 子命令 + 内置 spec-lint.yml 模板 + docs/CI_INTEGRATION.md
         (1 个新 Go 命令 + 1 个 yaml 模板 + 1 个新文档)
```

按 wave 顺序执行,每 wave 完成后跑 `go test ./...` + `free-kiro lint`。
Wave 之间通过 `[deps:]` 在 tasks.md 显式声明。

### 兼容策略

- **不**删除现有 `claude-code` / `codebuddy` flag 值 — 老用户配置
  不变。
- **不**改动 hook envelope 格式 — Cursor / Continue / OpenCode
  沿用 Claude Code 的 `event-keyed` schema(前提:这些 IDE 实际支持;
  若不支持,只 Parse + DetectAll 返回 DirExists=false,不影响 auto)。
- **不**新增 Go 依赖 — `go.mod` 字节级不变。
- **不**修改 `init` 命令的其他 flag — 只扩 `--ide` 的枚举值集合。

### 风险与回滚

| 风险 | 缓解 |
|---|---|
| Cursor / Continue / OpenCode 不支持 event-keyed hook | Parse + DetectAll 已分两层;`DirExists=false` 时 auto 模式自然跳过;不抛错 |
| `examples/todo-app/` 的 spec 写得不规范 | 用 `free-kiro spec quick` + 手工精修;落地前 `free-kiro lint` 0 ERROR |
| demo 命令 cwd 检测误判(`go run` 路径 vs 仓库 root) | 用 `runtime.Caller(0)` + 反推 `examples/todo-app/` 存在性,不依赖 git |
| README 重写后丢失原有"30 秒上手"内容 | 在 `docs/DEMO_VERIFICATION.md` 末尾保留"30 秒上手"原文快照,作为迁移审计 |

每个 task 一个独立 commit;任何 wave 失败 `git revert <commit>`
即可回滚。

### 发布

走常规 PR 流程。改动跨多文件但语义独立,可一次 PR 落地;若担心 review
量大,**wave 1 先 PR) → wave 2 → wave 3 三个连续 PR**(保持 main
随时绿)。