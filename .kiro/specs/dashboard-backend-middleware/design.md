# dashboard-backend-middleware — Design

<!--
本 spec 是纯后端硬化（中间件 + 静态资源 handler），不引入业务端点、不引入 fsnotify。
唯一 Go 代码新增/修改：
  - internal/visualize/middleware.go: 新文件，≤100 行
  - internal/visualize/server.go: 装配中间件链 + 新增 handleStatic + embed.FS 扩 static/dist

唯一架构原则：中间件装配顺序 withRecover → withLogger → withCacheHeaders → mux（外层兜底 panic，内层细粒度控制）。
-->

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ HTTP Request                                                     │
│   ↓                                                              │
│ withRecover (middleware.go)                                      │
│   ↓ defer recover() → 500 + log if panic                          │
│ withLogger (middleware.go)                                       │
│   ↓ defer log.Printf("%s %s %d %s") after next.ServeHTTP         │
│ withCacheHeaders (middleware.go)                                 │
│   ↓ 按路径前缀 set Cache-Control BEFORE next.ServeHTTP           │
│ mux (http.ServeMux)                                              │
│   ├── /                  → handleIndex (existing)                │
│   ├── /assets/*          → handleStatic (NEW)                    │
│   ├── /api/summary       → handleSummary (existing)              │
│   ├── /api/specs         → handleSpecs (existing)                │
│   ├── /api/spec/         → handleSpec (existing)                 │
│   └── /api/events        → handleEvents (SSE, existing)          │
└─────────────────────────────────────────────────────────────────┘
```

不变：
- 5 个现有端点（`/`、`/api/summary`、`/api/specs`、`/api/spec/<name>`、`/api/events`）。
- SSE 事件名 `ping`/`refresh`（dashboard-sse-bugfix 固化）。
- `Shutdown()` 5s timer 释放（dashboard-sse-bugfix 修复保留）。
- stdlib `net/http`，零第三方依赖。

新增：
- `internal/visualize/middleware.go`（≤ 100 行）。
- `Server.handleStatic`（≤ 50 行）。
- `//go:embed static/dist` 一行（与现有 `static/*` 并列）。

## Components

| Component | File | Responsibility |
|---|---|---|
| `withRecover` | `middleware.go` | `defer recover()` → `log.Printf` + `WriteHeader(500)` + JSON body |
| `withLogger` | `middleware.go` | 记录 `method path status latency`；`/assets/*` 降级为简洁日志 |
| `withCacheHeaders` | `middleware.go` | 按前缀分支：`/assets/*` 1y immutable；`/`、`/api/*` no-cache；其他透传 |
| `Server.handleStatic` | `server.go` | 从 embed.FS 读 `static/dist/<path>` 返回；防路径穿越 |
| `Server.embedStaticDist` | `server.go` | `//go:embed static/dist` 一行 |

### middleware.go 结构（≤ 100 行）

```go
package visualize

import (
    "encoding/json"
    "log"
    "net/http"
    "runtime/debug"
    "strings"
    "time"
)

func withRecover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("panic: %v\n%s", r, debug.Stack())
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                _ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
            }
        }()
        next.ServeHTTP(w, r)
    })
}

func withLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        // 包装 ResponseWriter 抓 status code
        rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rw, r)
        log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
    })
}

type statusRecorder struct {
    http.ResponseWriter
    status      int
    wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
    if !r.wroteHeader {
        r.status = code
        r.wroteHeader = true
        r.ResponseWriter.WriteHeader(code)
    }
}

func withCacheHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case strings.HasPrefix(r.URL.Path, "/assets/"):
            w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
        case r.URL.Path == "/api/events":
            // SSE: no-cache (代理不缓冲)
            w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        case strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/":
            w.Header().Set("Cache-Control", "no-cache")
        }
        next.ServeHTTP(w, r)
    })
}
```

### handleStatic 实现（≤ 50 行）

```go
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.Header().Set("Allow", http.MethodGet)
        http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
        return
    }
    // 路径前缀 /assets/，去掉前缀
    rel := strings.TrimPrefix(r.URL.Path, "/assets/")
    // 防路径穿越
    if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
        return
    }
    fsPath := "static/dist/" + rel
    data, err := staticFS.ReadFile(fsPath)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotFound)
        _ = json.NewEncoder(w).Encode(map[string]string{"error": "asset not found"})
        return
    }
    // Content-Type by extension
    switch {
    case strings.HasSuffix(rel, ".js"):
        w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
    case strings.HasSuffix(rel, ".css"):
        w.Header().Set("Content-Type", "text/css; charset=utf-8")
    case strings.HasSuffix(rel, ".json"):
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
    case strings.HasSuffix(rel, ".svg"):
        w.Header().Set("Content-Type", "image/svg+xml")
    default:
        w.Header().Set("Content-Type", "application/octet-stream")
    }
    _, _ = w.Write(data)
}
```

### server.go 改动（≤ 30 行 diff）

```go
// 现有：//go:embed static/*
var staticFS embed.FS

// 新增一行：
//go:embed static/dist
// (Go embed 语法要求 directive 相邻；放同一 var 块即可)
var staticFS embed.FS

// NewServer 中 mux 装配
func NewServer(addr string, ws WorkspacePaths, eng *spec.Engine) *Server {
    s := &Server{
        ws:         ws,
        eng:        eng,
        addr:       addr,
        subscribers: make([]chan struct{}, 0),
        stop:       make(chan struct{}),
    }
    mux := http.NewServeMux()
    mux.HandleFunc("/", s.handleIndex)
    mux.HandleFunc("/api/summary", s.handleSummary)
    mux.HandleFunc("/api/specs", s.handleSpecs)
    mux.HandleFunc("/api/spec/", s.handleSpec)
    mux.HandleFunc("/api/events", s.handleEvents)
    mux.HandleFunc("/assets/", s.handleStatic)  // 新增

    // 中间件链：recover → logger → cache → mux
    handler := withRecover(withLogger(withCacheHeaders(mux)))
    s.srv = &http.Server{Handler: handler}
    return s
}
```

## Error Handling

| 场景 | 行为 |
|---|---|
| handler panic | `withRecover` recover + log + 500 + JSON body |
| handler 写多次 WriteHeader | `statusRecorder` 去重，保留最后一次 status 给 logger |
| `/assets/..` 路径穿越 | `handleStatic` 400 + JSON |
| `/assets/` 不存在文件 | `handleStatic` 404 + JSON |
| `/assets/` method != GET | `handleStatic` 405 + Allow header |
| `/assets/` 含 `/` 多级 | 保留 prefix 匹配（如 `/assets/foo/bar.js` → `static/dist/foo/bar.js`） |
| `static/dist/` 目录为空 | `go build` 失败（embed 强制非空） |

**零吞错误**（.kiro/steering/agent-rules.md §2）：所有 `embed.FS.ReadFile` / `json.NewEncoder.Encode` 失败必须显式忽略原因（已写入 header 后无法改 status，silent ignore 是正确选择，但要在注释中说明）。

## Testing Strategy

### 单元测试（覆盖率新增行 ≥ 80%）

```go
// internal/visualize/middleware_test.go (新增)
func TestWithRecover_PanicReturns500(t *testing.T) {
    h := withRecover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        panic("boom")
    }))
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
    if rec.Code != 500 { t.Fatalf("got %d", rec.Code) }
    if !strings.Contains(rec.Body.String(), "internal server error") { t.Fatal(...) }
}

func TestWithLogger_LogsRequest(t *testing.T) { /* 捕获 log 输出验证 */ }

func TestWithCacheHeaders_AssetsImmutable(t *testing.T) {
    h := withCacheHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    }))
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, httptest.NewRequest("GET", "/assets/main.abc.js", nil))
    if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
        t.Fatalf("got %q", got)
    }
}

func TestWithCacheHeaders_APINoCache(t *testing.T) { /* ... */ }
func TestWithCacheHeaders_SSENoStore(t *testing.T) { /* ... */ }
func TestWithCacheHeaders_RootNoCache(t *testing.T) { /* ... */ }
func TestWithCacheHeaders_OtherPassthrough(t *testing.T) { /* 不设置 Cache-Control */ }

// internal/visualize/server_static_test.go (新增)
func TestHandleStatic_ExistingJS(t *testing.T) { /* 嵌入测试 dist 文件 */ }
func TestHandleStatic_NotFound(t *testing.T) { /* 404 */ }
func TestHandleStatic_PathEscape(t *testing.T) { /* 400 */ }
func TestHandleStatic_NonGetMethod(t *testing.T) { /* 405 */ }
func TestHandleStatic_ContentTypeByExtension(t *testing.T) {
    cases := []string{".js", ".css", ".json", ".svg", ".png"}
    for _, ext := range cases {
        t.Run(ext, ...)
    }
}
```

### 回归测试（保留现有）

- `TestServer_ShutdownWithoutWatcher`
- `TestServer_ConcurrentSubscribeBroadcast`
- `TestServer_WatcherStartedOnce`
- `TestServer_URL`

### E2E

```bash
free-kiro serve --port 7373 &
sleep 1
# 中间件行为
curl -sfI http://127.0.0.1:7373/ | grep -i 'cache-control:.*no-cache' || echo FAIL
curl -sfI http://127.0.0.1:7373/api/summary | grep -i 'cache-control:.*no-cache' || echo FAIL
# SSE no-store
curl -sfI http://127.0.0.1:7373/api/events | grep -i 'cache-control:.*no-store' || echo FAIL
# 静态资源（前提 dist/ 已构建）
curl -sfI http://127.0.0.1:7373/assets/main.abc.js | grep -i 'cache-control:.*immutable' || echo "FAIL or no dist yet"
# panic recover
# (难以从外部触发；通过单元测试覆盖)
kill %1
```

## Migration / Rollout

### 装配顺序

```
NewServer:
  mux := http.NewServeMux()
  mux.HandleFunc(...)         // 路由注册
  handler := withRecover(      // 1. 最外层：panic 兜底
              withLogger(     // 2. 中间层：所有请求日志
                withCacheHeaders(  // 3. 细粒度：路径分支
                  mux)))     // 4. 内层：业务路由
  s.srv = &http.Server{Handler: handler}
```

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| 现有 5 个端点 | 路由不变，handler 不动；中间件仅在外部包装 |
| SSE 事件名 | `ping`/`refresh` 不变 |
| `Shutdown()` | timer 释放路径不变 |
| `Server.URL()` | IPv6 处理不变 |
| stdlib 依赖 | 不引入任何第三方依赖 |
| Go embed | `//go:embed static/*` 保留；新增 `//go:embed static/dist` |

### 风险与回滚

- **风险 A**：中间件装配顺序错误 → recover 不生效。**对策**：单测 `TestWithRecover_PanicReturns500` 强制验证；E2E 故意制造 panic handler 验证。
- **风险 B**：cache headers 在 `WriteHeader` 后设置导致 ignored。**对策**：`withCacheHeaders` 在 `next.ServeHTTP` **之前**设置 header（设计即如此）；单测验证 header 在 body 写入前已设置。
- **风险 C**：`/assets/` 路径被现有路由 `/api/spec/` 误匹配。**对策**：`http.ServeMux` prefix 匹配按注册顺序；`/api/spec/` 与 `/assets/` 互不重叠（前者以 `/api/` 开头，后者以 `/assets/` 开头）。
- **风险 D**：`statusRecorder` 包装 `http.ResponseWriter` 破坏 `http.Flusher` 接口（SSE 需要）。**对策**：`/api/events` 走 SSE handler 不会被 `withCacheHeaders` 提前 buffer；如需兼容 Flusher，可实现 `http.Flusher` 转发（dashboard-realtime-fsnotify 处理）。

### 落地前置（按 .kiro/steering/agent-rules.md §1）

- 本 spec 三件套必须 `free-kiro lint` 全绿。
- 软依赖 dashboard-frontend-foundation 的 `dist/` 构建产物（embed 要求目录非空）；若 dist 不存在，`go build` 失败提示开发者跑 `make dashboard-dist`。
- 落地后由用户显式 `free-kiro spec approve dashboard-backend-middleware` 后再 `spec start`。
