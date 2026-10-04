---
mode: always
description: Go 社区通用编码规范（命名 / 错误处理 / 并发 / 接口 / 测试 / 注释与文档 / 依赖管理）
---

# Go 编码规范（通用）

> 适用对象：任何遵循 Go 社区通用风格基线的 Go 项目；本仓库
> （`github.com/jingyu525/free-kiro`）作为首选落地参照。
> 规范与 `golangci-lint` 配置（`.golangci.yml`）一一对应：可机器检查的规则由
> linter 强制，其余规则由人工 review 保证。
>
> **范围声明**：本文档**仅承载 Go 社区通用编码规范**（命名、错误处理、并发、
> 接口、测试通用部分、注释与文档通用部分、依赖管理通用部分），共 7 章。
> 本仓库的**项目特定策略**（覆盖率硬阈值、TODO owner、协议合规、内部依赖路径、
> commit message 中文、代码规模上限、PR 范围约束）见
> [`.kiro/steering/policy.md`](./policy.md)；**AI agent 协作硬性要求**见
> [`.kiro/steering/agent-rules.md`](./agent-rules.md)。
>
> **配套文档**：CI 门禁见 `.github/workflows/ci.yml` 的 `lint-go` job；
> 入口索引见 `CONTRIBUTING.md`；技术栈事实见
> [`.kiro/steering/tech.md`](./tech.md)。

## 目录

1. [命名约定](#1-命名约定)
2. [错误处理](#2-错误处理)
3. [并发](#3-并发)
4. [接口设计](#4-接口设计)
5. [测试](#5-测试)
6. [注释与文档](#6-注释与文档)
7. [依赖管理](#7-依赖管理)

每个章节的体例：

- 顶部 1 段"原则陈述"——为什么这么定。
- 末尾"✅ 推荐 / ❌ 反例"对照——可直接复制的 Go 代码片段。

---

## 1. 命名约定

### 1.1 原则

- **名字本身就是文档**。Go 名字短而准：包名小写不加下划线，类型
  PascalCase，导出标识符首字母大写，私有标识符首字母小写，常量按
  "导出/未导出"区分大小写而非前缀。
- **不用蛇形、不用匈牙利、不用类型前缀**。`userID` 不是 `iUserID`，
  `HTTPClient` 不是 `HttpClient`（首字母缩写全大写）。
- **包名 = 目录名**。`pkg/spec` 下的包叫 `spec`，调用时是
  `spec.New()` 不是 `spec.NewSpec()`。
- **接口名按行为命名**：单个方法用 `-er` 后缀（`Reader` / `Writer` /
  `Closer`），多方法按角色命名（`FileSystem`）。
- **避免冗余**：`customer.CustomerID` ❌ → `customer.ID` ✅。

### 1.2 标识符命名表

| 类别 | 规则 | 示例 |
|---|---|---|
| 包（package） | 全小写、单词、不复数、无下划线 | `spec`、`lint`、`hook` |
| 文件 | 全小写、下划线分隔、`.go` 后缀 | `spec.go`、`lint_gate.go`、`e2e_test.go` |
| 导出类型/函数/常量 | PascalCase | `LintError`、`NewSpec` |
| 未导出类型/函数/变量 | camelCase | `lintGate`、`parseSpec` |
| 局部变量 | 短名优先，作用域越大名字越长 | `i` < `idx` < `index` < `userIndex` |
| 常量 | 与变量同大小写规则；枚举用 `type X int` + `const` | `const ErrNotFound = ...` |
| 错误变量 | 以 `Err` 开头 | `ErrSpecNotFound` |
| 错误类型 | 以 `Error` 结尾 | `LintFailureError` |
| Sender/Receiver | 1–2 字母、与类型相关 | `(s *Spec)`、`(r *Reader)` |
| 接口（单方法） | 方法名 + `er` | `Reader`、`Closer`、`LintGate` |
| 缩写词 | 全大写或全小写（按导出/未导出） | `HTTP`、`URL` → `HTTPClient`、`parseURL` |

### 1.3 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：包名短、调用简洁
package spec

type Linter interface {
    Lint(name string) error
}

func New(name string) (*Spec, error) { ... }

// 调用侧
spec, err := spec.New("demo")  // 简洁、可读
```

```go
// ❌ 反例 1：包名冗余、调用方需要写两遍 "spec"
package spec

type SpecLinter interface { ... }
func NewSpec(name string) (*Spec, error) { ... }

// 调用侧
spec, err := spec.NewSpec("demo")  // 啰嗦
```

```go
// ❌ 反例 2：匈牙利命名 / 类型前缀
type IUserService interface { ... }   // 不用 I 前缀
var iCount int                        // 不用类型前缀
var bSuccess bool                     // 不用 b 前缀
var strName string                    // 不用 str 前缀
```

```go
// ❌ 反例 3：缩写词大小写错乱
type HttpClient struct { ... }   // 应为 HTTPClient
var urlParser func(*Url)         // 应为 URLParser
```

```go
// ✅ 推荐：Receiver 短、与类型相关、在所有方法上一致
type LintEngine struct { ... }

func (e *LintEngine) Run(spec string) error { ... }
func (e *LintEngine) Reset()                 { ... }

// ❌ 反例：每个方法 receiver 名不一致
func (e *LintEngine) Run(spec string) error { ... }
func (eng *LintEngine) Reset()               { ... }  // 同一个类型用了两个名
```

```go
// ✅ 推荐：错误变量 Err 开头，错误类型 Error 结尾
var ErrSpecNotFound = errors.New("spec not found")

type LintFailureError struct {
    Spec string
    Err  error
}

func (e *LintFailureError) Error() string { ... }

// ❌ 反例：命名混淆
var SpecNotFoundErr = ...                  // 应为 ErrSpecNotFound
type LintFailure struct{ ... }             // 应为 LintFailureError
```

### 1.4 由 linter 强制

- `staticcheck` 的 `ST1003`（命名应采用规范的大小写）
- `staticcheck` 的 `ST1019`（import 与包名重复）
- `revive` 的 `exported`、`package-comments`、`receiver-naming`
- `gofmt` / `goimports` 的格式化

### 1.5 何时可以例外

- 与外部 API（如 JSON tag、protobuf 字段）保持一致时，可在注释中说明
  `//nolint:revive` 的原因。
- 单字母变量名仅用于**极短作用域**（≤10 行）。

---

## 2. 错误处理

### 2.1 原则

- **错误是值**。Go 错误是返回值，不靠异常；要把它当作业务数据流处理。
- **处理或返回，不能默默吞掉**。要么 `if err != nil { ... }`，要么包到
  上一级；不允许 `_ = doX()` 式的忽略。
- **wrap 时保留上下文**。用 `fmt.Errorf("...: %w", err)` 让 `errors.Is`
  / `errors.As` 仍能穿透。
- **不要用 panic 做业务控制流**。`panic` 仅用于"真不可恢复"——如
  `mustXxx` 构造函数在程序启动时配置缺失。
- **sentinel error + custom error type** 二选一：跨包边界用
  `var ErrFoo = errors.New(...)`；本地结构化错误用 type + struct。

### 2.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：wrap 保留链路
func loadSpec(path string) (*Spec, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read spec %s: %w", path, err)
    }
    spec, err := parseSpec(data)
    if err != nil {
        return nil, fmt.Errorf("parse spec %s: %w", path, err)
    }
    return spec, nil
}

// ❌ 反例：吞掉错误
data, _ := os.ReadFile(path)
spec, _ := parseSpec(data)
return spec, nil
```

```go
// ✅ 推荐：sentinel error + 用 errors.Is 判断
var ErrSpecNotFound = errors.New("spec not found")

func Get(name string) (*Spec, error) {
    s, ok := store[name]
    if !ok {
        return nil, fmt.Errorf("spec %q: %w", name, ErrSpecNotFound)
    }
    return s, nil
}

// 调用方
spec, err := Get("demo")
if errors.Is(err, ErrSpecNotFound) {
    // 用户态分支
}

// ❌ 反例：字符串匹配
if strings.Contains(err.Error(), "not found") { ... }
```

```go
// ✅ 推荐：自定义错误类型 + errors.As 提取字段
type LintFailureError struct {
    Spec   string
    Reason string
}

func (e *LintFailureError) Error() string {
    return fmt.Sprintf("lint %s: %s", e.Spec, e.Reason)
}

// 调用方
var lfe *LintFailureError
if errors.As(err, &lfe) {
    log.Printf("lint failed on spec=%s reason=%s", lfe.Spec, lfe.Reason)
}

// ❌ 反例：导出普通 struct 当 error，但没实现 Error() 方法
type LintError struct { Spec string }   // 不会作为 error 流通
```

```go
// ❌ 反例：panic 做业务控制流
func MustParse(path string) *Spec {
    s, err := Parse(path)
    if err != nil {
        panic(err)   // 只在程序启动期、配置缺失等"绝不该失败"处使用
    }
    return s
}

// 调用方
s := MustParse(configPath)   // 启动期 OK
s := MustParse(userInput)    // ❌ 不允许：用户输入可能失败
```

### 2.3 由 linter 强制

- `errcheck`：所有 error 返回必须被检查或显式 `_ =`
- `wrapcheck`：调用方 wrap 返回的 error（`wrapcheck` 默认开启"内层包"豁免）
- `revive` 的 `error-strings`：错误信息不以大写字母或标点结尾
- `staticcheck` 的 `ST1005` / `ST1015`

### 2.4 何时可以例外

- `defer file.Close()` / `defer resp.Body.Close()`：用 `errcheck` 的
  `check-blank` 提示而非 fail。
- 启动期的 `MustXxx` 构造函数：明确命名，不复用做业务调用。

---

## 3. 并发

### 3.1 原则

- **不要通过共享内存通信，要通过通信共享内存**。goroutine 之间优先
  `chan` + 所有权交接；非共享状态不要加锁。
- **goroutine 由谁启动，谁负责关闭**。`go func()` 立刻考虑生命周期与
  `context.Context` 取消链路。
- **永远用 `context.Context` 传递取消与超时**。它是 goroutine 树的"根信号"。
- **`sync.Mutex` 保护的是状态，不是代码**。把锁和数据放在同一 `struct`，
  通过方法暴露访问，不导出锁。
- **不开裸 goroutine 池**。除非有可证的性能瓶颈，否则直接 `go`；worker
  pool 留到确实需要时再引入。

### 3.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：context 取消链路
func (e *Engine) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case job := <-e.jobs:
            if err := e.process(ctx, job); err != nil {
                return err
            }
        }
    }
}

// ❌ 反例：goroutine 无法取消
func (e *Engine) Run(stop chan struct{}) error {
    for job := range e.jobs {
        e.process(job)   // 没有上下文，stop 信号传不到下游
    }
    return nil
}
```

```go
// ✅ 推荐：sync.Mutex 与数据同 struct，方法暴露访问
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// ❌ 反例：把锁导出去
type Counter struct {
    Mu    sync.Mutex   // 导出锁 = 把同步责任甩给调用方
    value int
}
```

```go
// ✅ 推荐：用 chan 交接所有权
func (p *Pipeline) Next() <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for in := range p.src {
            out <- p.process(in)
        }
    }()
    return out
}

// ❌ 反例：多个 goroutine 写同一个 slice/map
var results []Result
for _, in := range inputs {
    go func(in Input) {
        results = append(results, process(in))   // data race
    }(in)
}
```

```go
// ✅ 推荐：errgroup 管理并发 + 取消
import "golang.org/x/sync/errgroup"

func (s *Server) startWorkers(ctx context.Context, n int) error {
    g, ctx := errgroup.WithContext(ctx)
    for i := 0; i < n; i++ {
        g.Go(func() error {
            return s.worker.Run(ctx)
        })
    }
    return g.Wait()
}

// ❌ 反例：sync.WaitGroup + 手动传播第一个 error
var wg sync.WaitGroup
var firstErr error
var mu sync.Mutex
for _, w := range workers {
    wg.Add(1)
    go func(w Worker) {
        defer wg.Done()
        if err := w.Run(); err != nil {
            mu.Lock()
            if firstErr == nil {
                firstErr = err
            }
            mu.Unlock()
        }
    }(w)
}
wg.Wait()
return firstErr
```

### 3.3 由 linter 强制

- `govet` 的 `copylocks`、`printf`、`shadow`（捕捉 `:=` 遮蔽 ctx）
- `gosec` 的 `G104`（未检查的错误）等价检查
- `staticcheck` 的 `SA2002`（atomic 不当使用）、`SA4006`（值被遮蔽）
- `go test -race` 在 CI 必跑

### 3.4 何时可以例外

- 性能热路径里用 `sync/atomic` 替代 Mutex：可接受，但要在注释里
  标注 race detector 测过的场景。
- 测试代码里的 `time.Sleep(...)`：可以用，但优先用 `for { ... if cond { break } ; time.Sleep(...) }` 轮询直到条件成立或 `t.Fatal` 兜底。

---

## 4. 接口设计

### 4.1 原则

- **接口在使用方定义，不在实现方**。定义在消费侧的小接口比放实现侧的
  大接口更容易被 mock 和替换。
- **接口尽量小**（1–3 个方法）。`io.Reader` 一个方法，`http.Handler` 一个
  方法；多方法接口只在角色清晰时使用。
- **接受接口、返回具体类型**。`func NewClient() *Client` ✅；
  `func NewClient() Clienter` ❌。
- **接口命名**：单方法 → 方法名+`er`；多方法 → 角色名；不带 `I` 前缀。
- **避免空接口** `interface{}` / `any` 流到 API 边界；如必须，用泛型或
  类型约束替代。

### 4.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：在使用方定义小接口
package lint

// Linter 是消费侧只需要的能力
type Linter interface {
    Lint(spec string) error
}

func Run(l Linter, spec string) error { ... }

// 实现侧在另一个包里，不需要 import 这层接口
package internal

type Engine struct{ ... }
func (e *Engine) Lint(spec string) error { ... }  // 隐式实现 Linter
```

```go
// ❌ 反例：在实现侧定义大接口，让消费方被迫 import
package internal

type EngineInterface interface {   // 实现侧定义了消费侧不需要的方法
    Lint(spec string) error
    Reset()
    Reload() error
    Stats() Stats
}
```

```go
// ✅ 推荐：接受接口、返回具体类型
func NewEngine(cfg Config) *Engine { ... }   // 返回具体类型
func RunEngine(e *Engine) error { ... }      // 也可以接受具体类型

// 接受接口的写法
func RunLinter(l Linter) error { ... }        // 仅在需要 mock/替换时接受接口
```

```go
// ❌ 反例：返回接口把实现锁死
func NewEngine(cfg Config) LintEngine { ... }   // 调用方拿到接口，无法用未导出字段
```

```go
// ✅ 推荐：单方法接口用 -er 后缀
type Reader interface { Read(p []byte) (int, error) }
type Closer interface { Close() error }
type LintGate interface { Run(spec string) error }

// 多方法接口用角色名
type FileSystem interface {
    Open(name string) (File, error)
    Stat(name string) (os.FileInfo, error)
}
```

```go
// ❌ 反例：接口名带 I 前缀
type ILintEngine interface { ... }   // Go 不推荐匈牙利命名
```

### 4.3 由 linter 强制

- `revive` 的 `unexported-return`（不导出返回类型）
- `revive` 的 `redefines-builtin-id`
- `staticcheck` 的 `ST1009`（方法名应一致）、`ST1016`（receiver 类型一致）

### 4.4 何时可以例外

- 跨包公开 API 的接口需要稳定时，可在实现侧定义，但要标注"consumed by package X/Y/Z"。
- 类型断言 / `any` 用于泛型约束无法表达的情况（如 `encoding/json.RawMessage`）。

---

## 5. 测试

### 5.1 原则

- **测试是代码的第一公民**。与产品代码同包、同 review 标准。
- **表驱动测试**（table-driven）是默认形态。同一逻辑多场景 → 1 个
  `TestXxx(t *testing.T)` + 多 `cases := []struct{...}{}` 子用例。
- **只用标准库 `testing`**：断言用 `if got != want { t.Errorf(...) }` /
  `t.Fatal(...)` / `t.Fatalf(...)`，**不引入 testify / assert / require**
  （与 `.kiro/steering/tech.md` 测试节一致）。初始化 / 前置条件失败用
  `t.Fatal` / `t.Fatalf` 立即停止；普通断言失败用 `t.Errorf` 让后续
  case 继续跑，方便看全部差异。
- **不写无断言测试**。每条 case 必须至少有 1 个 `t.Error*` /
  `t.Fatal*` 调用，否则视为无断言测试。
- **`-race` 必跑**。任何启用 goroutine 的代码都必须 `go test -race`
  验证（详见 `.kiro/steering/tech.md` 测试节）。

### 5.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：表驱动测试
func TestLintEngine_Run(t *testing.T) {
    cases := []struct {
        name    string
        spec    string
        wantErr error
    }{
        {"happy path", "demo", nil},
        {"spec not found", "missing", ErrSpecNotFound},
        {"placeholder AC", "draft", ErrLintFailure},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            err := engine.Run(tc.spec)
            if tc.wantErr != nil {
                if !errors.Is(err, tc.wantErr) {
                    t.Fatalf("Run(%q) = %v, want %v", tc.spec, err, tc.wantErr)
                }
                return
            }
            if err != nil {
                t.Fatalf("Run(%q) unexpected error: %v", tc.spec, err)
            }
        })
    }
}

// ❌ 反例：复制粘贴 N 个 Test 方法
func TestLintEngine_Run_HappyPath(t *testing.T) { ... }
func TestLintEngine_Run_NotFound(t *testing.T) { ... }
func TestLintEngine_Run_Draft(t *testing.T) { ... }   // 重复 boilerplate
```

```go
// ✅ 推荐：t.Fatal 区分初始化失败 vs t.Errorf 区分普通断言
func TestUser_Create(t *testing.T) {
    u, err := NewUser("alice")
    if err != nil {
        t.Fatalf("NewUser 初始化失败: %v", err)   // 初始化失败，后面没意义
    }
    if u == nil {
        t.Fatal("NewUser 返回 nil")
    }

    if got, want := u.Name, "alice"; got != want {
        t.Errorf("u.Name = %q, want %q", got, want)   // 断言失败继续，方便看全部差异
    }
    if !u.Active {
        t.Error("u.Active = false, want true")
    }
}
```

```go
// ❌ 反例：无断言测试（CI 必报）
func TestSomething(t *testing.T) {
    result := DoWork()
    _ = result                          // 没断言
}
```

```go
// ✅ 推荐：临时文件用 t.TempDir()，自动清理
func TestLoadSpec(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "spec.md")
    if err := os.WriteFile(path, []byte("# spec"), 0o644); err != nil {
        t.Fatalf("WriteFile: %v", err)
    }

    spec, err := LoadSpec(path)
    if err != nil {
        t.Fatalf("LoadSpec: %v", err)
    }
    if got, want := spec.Name, "spec"; got != want {
        t.Errorf("spec.Name = %q, want %q", got, want)
    }
}

// ❌ 反例：手工写 defer os.Remove
func TestLoadSpec(t *testing.T) {
    path := "/tmp/spec_test.md"
    if err := os.WriteFile(path, []byte("# spec"), 0o644); err != nil {
        t.Fatal(err)
    }
    defer os.Remove(path)               // 容易忘，测试间互相污染
    ...
}
```

```go
// ✅ 推荐：race detector + 并发测试
func TestCounter_ConcurrentInc(t *testing.T) {
    var c Counter
    const n = 1000
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.Inc()
        }()
    }
    wg.Wait()
    if got, want := c.Value(), n; got != want {
        t.Errorf("c.Value() = %d, want %d", got, want)
    }
}
// 调用：go test -race ./...
```

### 5.3 由 linter 强制

- `govet` 的 `lostcancel`、`nilness`（不直接由 lint 触发但易出 bug）
- 覆盖率硬阈值（≥70% 全包 / ≥80% 新增）见
  [`.kiro/steering/policy.md`](./policy.md)，本文档不重复项目特定阈值

### 5.4 何时可以例外

- E2E 测试（如启动 CLI 子进程跑 smoke）允许用 `exec.Command` + `os.Stderr.Pipe`，
  不强制 `-race`。

---

## 6. 注释与文档

### 6.1 原则

- **注释解释"为什么"，不是"是什么"**。代码本身说明做了什么；注释补
  充业务背景、设计取舍、注意事项。
- **导出符号必须 godoc**。每个导出的 `type` / `func` / `const` / `var`
  都要有完整句子注释，以符号名开头（`// Foo does ...`）。
- **包必须有 package doc**。在 `doc.go` 或包内任一文件顶部写
  `// Package foo ...`。
- **不写废话注释**。`i++ // i 自增` ❌；保留有信息量的注释。

### 6.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：完整 godoc 句子、以符号名开头
// LintEngine 校验 .kiro/specs 下的所有 spec 文档是否符合 EARS 句式。
// 它是无状态、并发的：可在多个 goroutine 之间共享同一个实例。
type LintEngine struct { ... }

// ❌ 反例：缺注释、或注释不以符号名开头
type LintEngine struct { ... }                    // 没注释

// 校验 spec 的引擎
type LintEngine struct { ... }                    // 注释不以符号名开头，godoc 索引会断
```

```go
// ✅ 推荐：解释"为什么"，不是"是什么"
// ErrSpecNotFound 在 store 中查不到 spec 时返回。
// 调用方应该用 errors.Is 判断而非字符串匹配。
var ErrSpecNotFound = errors.New("spec not found")

// ❌ 反例：复述代码
var ErrSpecNotFound = errors.New("spec not found") // 定义一个错误变量
```

### 6.3 由 linter 强制

- `revive` 的 `exported`、`package-comments`
- `revive` 的 `docstring`（结构体字段、`const` 块）
- `staticcheck` 的 `ST1000`（缺包注释）

---

## 7. 依赖管理

### 7.1 原则

- **最小依赖**。每加一个 import 都要有明确理由；能用标准库解决就不引第三方。
- **依赖锁定**：`go.mod` + `go.sum` 必须 100% 提交；CI 跑 `go mod verify`。
- **升级策略**：第三方依赖按 semver 升；major 版本升级单独 PR，并在
  PR 描述里说明 breaking change。

### 7.2 ✅ 推荐 / ❌ 反例

```go
// ✅ 推荐：能用 stdlib 解决就不引第三方
buf, err := io.ReadAll(resp.Body)   // stdlib
if err != nil { ... }

// ❌ 反例：为了 1 行代码引整个第三方库
import "github.com/some/bodyfetcher"   // 仅仅为了把 io.ReadAll 包一下
```

```bash
# ✅ 推荐：升级单个依赖、固定版本、可 review
go get github.com/foo/bar@v1.2.3
go mod tidy

# ❌ 反例：裸 `go get -u` 把所有依赖升一通
go get -u ./...   # 难以 review，可能引入不可控 breaking change
```

> **commit 模板与 major 升级策略**见 `.kiro/steering/policy.md` §3「依赖协议合规」
> 与 §5「commit message 格式」——本文档不重复项目特定条款。

### 7.3 由 linter 强制

- `go mod verify` 与 `go mod tidy -diff` 在 CI 跑
- `govet` 的 `modifies` 检查 import 是否被使用

### 7.4 何时可以例外

- 紧急 hotfix 可临时用 `replace` 指令指定 fork，但必须提 issue 跟进，
  在下一个常规 PR 内切回上游。
- 工具生成的 `*.pb.go` / `*_gen.go` / `mock_*.go` 可豁免 import 分组规则
  （生成器不保证格式），但需在生成脚本里强制 `goimports` 后置处理。

---

> 文档结束。本文档作为项目级 steering（`mode: always`）注入 IDE 指令
> 文件，变更请同步更新 `.kiro/steering/coding-style.md`、`.golangci.yml`、
> `.github/workflows/ci.yml`、`Makefile`、`CONTRIBUTING.md`，并重跑
> `free-kiro steering inject` 重生成 5 份 IDE 指令文件底部块。
