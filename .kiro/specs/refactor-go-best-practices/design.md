<!-- TODO: 描述技术架构与关键决策。这一段是给"自己"看的——30 天后回看能否立刻想起来为什么这么做 -->

# refactor-go-best-practices — Design

## Architecture

本次重构保持现有包边界不变（`cmd/` + 13 个 `internal/` 包），只做
**横向拆分**（单文件 → 多文件，职责分离）与**纵向抽取**（散落逻辑 →
新公共包）。新引入两个公共包：

- `internal/frontmatter` — YAML frontmatter 解析/序列化/校验
- `internal/cli/runner.go` — cli 子命令的 cmd runner 模板

依赖方向：

```
cmd/free-kiro
  └─ internal/cli          (cobra commands; uses runner + frontmatter)
       ├─ internal/spec    (engine)
       ├─ internal/errors  (typed errors + ExitCode)
       ├─ internal/frontmatter  (NEW — extracted from 4 sites)
       └─ internal/runner  (NEW — extracted from cli/*)

internal/steering ──┐
internal/skill    ──┼──> internal/frontmatter
internal/spec     ──┘
```

不创建新的"service"或"usecase"层（避免过度分层）；不引入接口抽象
（只在 ≥3 个真实实现时才抽接口）。本次也不引入 DI 容器——保留现有的
package-level 函数风格。

### 拆分策略（超大文件）

| 原文件 | 行数 | 拆分后 | 关注点 |
|---|---|---|---|
| `cli/spec_lifecycle.go` | 333 | `spec_lifecycle.go` (cmd wiring) + `spec_state.go` (state transition helpers) + `spec_format.go` (output formatting) | 命令 → 状态机 → 打印 |
| `cli/skill.go` | 387 | `skill.go` (cmd + flags) + `skill_install.go` (bundle copy + hook install) + `skill_sync.go` (update + show) | 命令 → 安装 → 同步 |
| `cli/doctor.go` | 343 | `doctor.go` (cmd + reporter) + `doctor_checks.go` (each check as funcs) + `doctor_report.go` (markdown formatter) | 入口 → 检查 → 报告 |
| `spec/engine.go` | 468 | `engine.go` (Engine struct + New/load/save) + `engine_state.go` (phase transitions) + `engine_io.go` (file IO + frontmatter) | 结构 → 状态 → IO |
| `visualize/server.go` | 444 | `server.go` (router + lifecycle) + `server_sse.go` (SSE handler) + `server_static.go` (static asset handler) | 路由 → SSE → 静态 |

每个新文件 ≤ 250 行（lint 用 `wc -l` 卡线），public API 不变。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `internal/frontmatter` (NEW) | YAML frontmatter parse/marshal/validate，body 透传 | `Parse(io.Reader) (Frontmatter, []byte, error)` / `Marshal(Frontmatter, []byte) ([]byte, error)` / `Validate(Frontmatter, Schema) error` |
| `internal/cli/runner.go` (NEW) | cli 子命令的公共骨架：engine 解析、typed-error 映射、panic recover、next 提示 | `RunCmd(ctx, *cobra.Command, []string, func(ctx, eng) error) error` |
| `internal/cli/spec_lifecycle.go` (refactored) | spec approve/start/complete 子命令 | 保持原签名 |
| `internal/cli/skill_install.go` (NEW) | skill install 逻辑 | 同 `runSkillInstall` |
| `internal/cli/doctor_checks.go` (NEW) | 每个 check 独立函数 | `checkPATH`, `checkWorkspace`, ... |
| `internal/spec/engine_state.go` (NEW) | 状态机转移 | `transition(current, target)` typed |
| `internal/visualize/server_sse.go` (NEW) | SSE 流处理器 | `sseHandler(w, r)` |

## Data Model

### Frontmatter 包

```go
package frontmatter

type Frontmatter map[string]any

type FieldRule struct {
    Required bool
    Type     string // "string"|"int"|"bool"|"[]string"
    Allowed  []any  // enum
}

type Schema map[string]FieldRule

// Parse reads `--- yaml\n---\n<body>` from r.
// On missing `---` separator → returns (nil, body, ErrNoFrontmatter).
// On invalid YAML → wraps *yaml.SyntaxError via fmt.Errorf %w.
func Parse(r io.Reader) (Frontmatter, []byte, error)
func Marshal(fm Frontmatter, body []byte) ([]byte, error)
func Validate(fm Frontmatter, s Schema) error
```

零外部依赖——用 stdlib `encoding/json` 替代 `gopkg.in/yaml.v3`
（现有 yaml 解析只在 init.go 一处且简单到可以用 regex + json 表达，
但避免行为变更风险，**保持现有 YAML 格式**，实际用
`gopkg.in/yaml.v3` 已有的间接依赖或 stdlib `sigs.k8s.io/yaml`）。
**决定**：沿用现有的 yaml 解析方式（`internal/cli/init.go` 里的
手写 split-on-`---`），先不引入 yaml 库；Frontmatter 在内存里就是
`map[string]string`（覆盖现有 4 个 call-site 的字段类型），Serialize
时输出 `key: value` 单行格式。

### 错误分类（保留）

| 类型 | exit code | 触发条件 |
|---|---|---|
| `UsageError` | 3 | 缺参数、互斥 flag、非法输入 |
| `LintGateError` | 2 | lint 失败但尝试 approve/start |
| `WorkspaceError` | 2 | `.kiro/` 缺失或不可读 |
| `TransitionError` | 2 | 非法 phase 转移 |
| `TaskGraphError` | 1 | tasks.md 环依赖 |
| 其他 `KiroError` | 2 | 默认 engine error |

`errors.ExitCode()` 已实现，本次重构只在 cli 层把
`fmt.Errorf("...")` 替换为 `errors.NewUsage(op, msg)` /
`errors.Wrap(op, err, msg)`。

## Error Handling

### 当前问题

```go
// internal/cli/spec_new.go:163
return "", "", "", fmt.Errorf("spec name is required when --from-* is not used")
```

↑ 这是典型的 UsageError，但走 `fmt.Errorf` 后会被 `ExitCode()` 兜底
为 2（engine error）。用户用错应该退出 3 而不是 2。

### 目标写法

```go
import kiroerr "github.com/jingyu525/free-kiro/internal/errors"

// 普通 UsageError（裸 op）
return kiroerr.NewUsage("spec.new", "spec name is required when --from-* is not used")

// 带 wrap 的 IO 错误
return kiroerr.Wrap("init.writeAgents", err, "写入 AGENTS.md 失败")
```

为减少 import 别名丑陋，在 `internal/errors/errors.go` 加两个便捷
构造器（不改 `ExitCode`）：

```go
func NewUsage(op, msg string) *UsageError {
    return &UsageError{KiroError: New(op, msg)}
}
func WrapUsage(op string, err error, msg string) *UsageError {
    return &UsageError{KiroError: Wrap(op, err, msg)}
}
```

### 错误返回契约

- cli 层所有错误：经 `exitWithError(err)`（已在 cli/root.go）→
  `ExitCode(err)` → 写入 stderr → 进程退出对应码
- 包内 helper：直接 `return errors.Wrap(op, err, msg)`，
  不在中间层打 log（避免"log+return"双重汇报）
- `exitWithError` 内部统一调用 `cobra.SilenceErrors = true` + 自己打印

## Testing Strategy

### 单元测试

每个新抽取的包至少 1 个 `*_test.go`：

| 文件 | 覆盖目标 | 关键 case |
|---|---|---|
| `frontmatter/frontmatter_test.go` | Parse / Marshal / Validate | 8 个 case 见 AC |
| `cli/runner_test.go` | RunCmd 4 个职责 | engine nil / panic recover / typed error → exit code / next 提示 |
| `cli/spec_state_test.go` | 状态机转移 | 合法转移 / 非法转移 / TransitionError |
| `cli/doctor_checks_test.go` | 每个 check | PATH OK / PATH missing / IDE hook missing |
| `spec/engine_io_test.go` | frontmatter 读写 round-trip | spec meta + body 保持 |

### 回归测试

每次 wave 完成后：

```bash
go vet ./...
go build ./...
go test ./...
free-kiro lint refactor-go-best-practices
free-kiro doctor
goreleaser build --snapshot --clean && \
  dist/free-kiro_darwin_arm64/free-kiro --version
```

CLI 行为不变通过现有 `_test.go` + `free-kiro doctor` + 手工冒烟
（README 列出的子命令各跑一遍 flag-set）保障。

### 性能 / 内存

无性能目标（Out of Scope）。只需确认无明显回归（go test 无超时）。

## Migration / Rollout

### 落地顺序（4 wave）

按依赖方向自底向上：

```
Wave 1: frontmatter 抽取 + 单元测试 (无 cli 改动)
Wave 2: cli 层 typed error 替换 + runner.go 抽出 (依赖 Wave 1 的 frontmatter 在 spec/engine 调用方)
Wave 3: 拆超大文件 (cli/spec_lifecycle / cli/skill / cli/doctor / spec/engine / visualize/server)
Wave 4: 回归 + lint + 文档同步
```

### 兼容策略

- 不引入新 flag、不删除旧 flag、不重命名 public 函数
- frontmatter 包 v1：4 个 call-site 全部迁移后，删除旧的 inline 实现
  （**未迁移前**保留兼容 shim，避免半步迁移导致 test 失败）
- cli/runner.go 引入后**不强制**所有 subcommand 立刻迁移——本次
  重构只迁移 ≥3 处重复的子命令（spec generate / approve / start /
  complete / skill install / update / doctor），其余保留原样

### 风险与回滚

- 风险：frontmatter 解析行为微妙变化（CRLF / BOM / 空 body）
- 缓解：先用 frontmatter_test.go + 现有 testdata/*.md 做"双轨"
  验证（同时跑新旧两套，确认输出相等），再删旧实现
- 回滚：每个 task 一个独立 commit；如 lint/test 失败，
  `git revert <commit>` 即可

### 发布

不需要 feature flag，不需要 staged rollout。重构完成 → 走常规 PR 流程
→ CI 通过 → merge → 下一版本（v0.7.1 或 v0.8.0）自然包含。
