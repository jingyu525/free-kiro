# improve-visualize-hooks-reliability — Design

## Architecture

三个改动点分别落在 `internal/visualize/`、`internal/hooks/`、`internal/lint/`
三个独立的包里，互不依赖（除 visualize 内部的 fix 之间）。不改公共 API，
只动内部实现 + 新增字段。零 breaking change。

## 2.1 visualize 包（fix S1 / S8 / V1）

### 当前问题

1. `server_sse.go:14-46` 把 `notifier` 声明为 package-level 全局变量，
   `subscribe` / `unsubscribe` / `broadcast` / `watchChanges` 都直接读写
   `notifier.subscribers`，无任何同步 → 多 goroutine 并发 data race。
2. `server_sse.go:124` `go s.watchChanges()` 每次 SSE client 连接都
   启动一个新 watcher → 多 client 重复启动。
3. `server.go:99-111` `Shutdown()` 在 `watchChanges` 从未启动时永久
   hang 在 `<-s.done`；并且用 `srv.Close()` 而不是 graceful
   `srv.Shutdown(ctx)`，活跃 SSE 连接不会被主动关闭。

### 设计

把 `notifier` 从全局变量挪到 `Server` struct，加 `sync.Mutex` 保护
`subscribers` 切片。`watchChanges` 用 `sync.Once` 保证全 Server 生命周期
只启动一次。`Shutdown` 走 `srv.Shutdown(ctx)` + 5 秒 deadline，
`<-s.done` 加 `time.After` fallback。

```go
type Server struct {
    // ...existing fields...
    notifierMu     sync.Mutex
    subscribers    []chan struct{}
    watcherOnce    sync.Once      // S8: 一次性启动 watcher
    watcherRunning chan struct{}  // 关闭后表示 watcher 已退出
}

const shutdownTimeout = 5 * time.Second  // 写常量，不硬编码 magic number

func (s *Server) subscribe() chan struct{} {
    ch := make(chan struct{}, 1)
    s.notifierMu.Lock()
    s.subscribers = append(s.subscribers, ch)
    s.notifierMu.Unlock()
    return ch
}

func (s *Server) unsubscribe(ch chan struct{}) {
    s.notifierMu.Lock()
    defer s.notifierMu.Unlock()
    for i, c := range s.subscribers {
        if c == ch {
            s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
            close(c)
            return
        }
    }
}

func (s *Server) broadcast() {
    s.notifierMu.Lock()
    subs := append([]chan struct{}{}, s.subscribers...)  // 快照，避免持锁 IO
    s.notifierMu.Unlock()
    for _, ch := range subs {
        select {
        case ch <- struct{}{}:
        default:
        }
    }
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
    // ...
    ch := s.subscribe()
    defer s.unsubscribe(ch)
    // ...
    s.watcherOnce.Do(func() {
        go s.watchChanges()
    })
    // ...
}

func (s *Server) Shutdown() error {
    s.watcherOnce.Do(func() {})  // 幂等标记，让 <-s.done 有界
    select {
    case <-s.watcherRunning:
    case <-time.After(shutdownTimeout):
    }
    ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
    defer cancel()
    return s.srv.Shutdown(ctx)
}
```

## 2.2 hooks 包（fix H8）

### 当前问题

`Registry.Match()` 每次都重新 `LoadAll()` → 每次 Match 都重新 IO 读
所有 hook JSON 文件。PreToolUse 高频场景下浪费明显。

### 设计

在 `Registry` struct 加 `sync.RWMutex` 保护的缓存：

```go
type Registry struct {
    ws        *workspace.Workspace
    cacheMu   sync.RWMutex
    cache     []*models.Hook  // LoadAll 结果
    cacheTime time.Time       // 缓存填充时刻
}

const hookCacheTTL = 1 * time.Second  // 简单 TTL，避免 stale

func (r *Registry) cachedAll() ([]*models.Hook, error) {
    r.cacheMu.RLock()
    if time.Since(r.cacheTime) < hookCacheTTL && r.cache != nil {
        out := r.cache
        r.cacheMu.RUnlock()
        return out, nil
    }
    r.cacheMu.RUnlock()
    return r.refreshCache()
}

func (r *Registry) refreshCache() ([]*models.Hook, error) {
    r.cacheMu.Lock()
    defer r.cacheMu.Unlock()
    all, err := r.LoadAll()  // 复用已有 IO 逻辑
    if err != nil {
        return nil, err
    }
    r.cache = all
    r.cacheTime = time.Now()
    return all, nil
}

func (r *Registry) Add(h *models.Hook) (string, error) {
    // ...现有逻辑...
    // 写盘成功后让下次 Match 强制刷新
    r.cacheMu.Lock()
    r.cache = nil
    r.cacheMu.Unlock()
    return path, nil
}
```

简单 TTL（1s）而不是 fsnotify 失效，够用 + 实现简单。TTL 过期或
`Add` 写盘后强制下一次 reload。

## 2.3 lint 包（fix F1 / L3）

### 当前问题

1. `linter.go:67-81` `os.ReadFile` 失败时除 `os.IsNotExist` 外的
   错误被静默吞掉 → 权限错误被误报为 `missing-requirements`。
2. `linter.go:120-135` `Gate()` 用 `strings.HasPrefix(i.Message,
   "[baseline] ")` 判断 baseline → 二次信息 antipattern。

### 设计

**F1** — 改用 `errors.Is(err, fs.ErrNotExist)` 显式判断：

```go
data, err := os.ReadFile(firstPath)
switch {
case err == nil:
    text := string(data)
    // ...
case errors.Is(err, fs.ErrNotExist):
    out = append(out, Issue{...missing-requirements...})
default:
    out = append(out, Issue{
        Severity: SeverityError,
        Code:     "lint-requirements-read-error",
        Message:  fmt.Errorf("read %s: %w", firstPath, err).Error(),
        Location: firstPath,
    })
}
```

**L3** — 给 `Issue` 加 `Baseline bool` 字段；`Spec` 在 baseline 命中时
置位；`Gate` 直接用字段：

```go
type Issue struct {
    Severity string
    Code     string
    Message  string
    Location string
    Hint     string
    Baseline bool   // ← 新字段，gate 直接读
}

// 在 Spec() 里（linter.go:105 附近）：
if applyBaseline && base.ShouldIgnore(out[i].Code) {
    out[i].Baseline = true
    out[i].Message = baselinePrefix + out[i].Message  // 保留旧格式给 lint 输出
}

// 在 Gate() 里（linter.go:120）：
for _, i := range Spec(specDir) {
    if i.Severity != SeverityError { continue }
    if isMissingCode(i.Code) { continue }
    if i.Baseline { continue }  // ← 取代 strings.HasPrefix 判断
    gate = append(gate, i)
}
```

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `visualize.Server.subscribe/unsubscribe/broadcast` | SSE 订阅广播，加 mutex 保护 | `subscribe() chan struct{}` |
| `visualize.Server.watcherOnce` | 保证 watcher 启动幂等 | `sync.Once` |
| `visualize.Server.Shutdown` | 带 timeout 的 graceful shutdown | `Shutdown() error` |
| `hooks.Registry.cachedAll/refreshCache` | 1s TTL 的 hook 缓存 | `cachedAll() ([]*models.Hook, error)` |
| `lint.Issue.Baseline` | 显式标记 baseline 命中 | `bool` 字段 |
| `lint.Gate` | 直接用 `Issue.Baseline` 而非 Message 前缀 | 内部实现改动 |

## Data Model

- `Issue` struct 新增 `Baseline bool` 字段（零值 false，向后兼容）。
- `Server` struct 新增 `notifierMu / subscribers / watcherOnce / watcherRunning`。
- `Registry` struct 新增 `cacheMu / cache / cacheTime`。
- 新常量：`shutdownTimeout = 5 * time.Second`、`hookCacheTTL = 1 * time.Second`。

## Error Handling

- 修 `LoadFile` / `Stat` 失败时：缺文件走 `missing-*` Issue（保持原行为），
  其他错误用 `lint-*-error` Issue wrap `err` 暴露根因。
- `Registry.cachedAll` 的 IO 错误透传给 caller，不吞。
- `subscribe/unsubscribe` 不返回 error（内部函数），mutex 失败 panic
  即可（Go runtime 已保证 mutex 不会出错）。

## Testing Strategy

| 测试文件 | 覆盖 | 关键 case |
|---|---|---|
| `internal/visualize/server_sse_test.go`（新） | S1+S2+S3 | 50 goroutine 并发 subscribe/unsubscribe + `go test -race`；watcher 启动计数；Shutdown timeout |
| `internal/hooks/registry_test.go`（增） | H8 | 100 次 Match < 100ms；Add 后立即 reload |
| `internal/lint/baseline_test.go`（增） | L3 | baseline 命中 `Issue.Baseline==true`；Gate 输出不含 baseline 命中；Message 自定义前缀不影响 |
| `internal/lint/requirements_test.go`（增） | F1 | 用 chmod 000 制造 EACCES，断言 `lint-requirements-read-error` Issue 含 root cause |

每个 fix 至少 1 阳 + 1 阴 case。`go test -race ./...` 在 CI 必跑。

## Migration / Rollout

无 migration — 内部字段 + 内部行为变更，公共 API 不变。
`.meta.json` 的 `baseline` 字段不动；老的 `.baseline.json` 文件继续可用。
`Issue.Baseline` 是新字段，老代码（如果有反射 marshal Issue）零值
就是 `false`，向后兼容。