# dashboard-backend-middleware

<!--
为 free-kiro serve dashboard 提供后端硬化：三个中间件 + 静态资源 handler。
  - withRecover     panic → 500 + 日志，不挂死 http.Server
  - withLogger      access log（method/path/status/latency）
  - withCacheHeaders 按前缀分支：/assets/* 1y immutable，/ no-cache，/api/* no-cache
  - handleStatic    走 embed.FS 的 static/dist 子树，供前端打包产物加载

本 spec 不引入新业务端点（属于 dashboard-backend-api-extensions）。
本 spec 不引入 fsnotify（属于 dashboard-realtime-fsnotify）。
本 spec 不写前端（属于 dashboard-frontend-foundation）。
-->

## User Stories

- As a dashboard 用户 I want 服务端 panic 时还能继续访问页面 so that 不会因为一个 handler bug 让整个 dashboard 不可用。
- As a 服务端维护者 I want 看到每次请求的 access log（method / path / status / latency）so that 排查性能问题有数据。
- As a 浏览器 I want `/assets/main.<hash>.js` 缓存 1 年（immutable）so that 二次访问秒开。
- As a 浏览器 I want `/` 与 `/api/*` 始终 `no-cache` so that dashboard 改动后我能立刻看到新版本。
- As a 前端构建工具 I want `dist/main.<hash>.{js,css}` 通过 `//go:embed` 嵌入二进制 so that 用户 `go install` 后就有完整 dashboard，零运行时下载。

## Acceptance Criteria

### withRecover 中间件

- [AC-1] THE SYSTEM SHALL 在 `internal/visualize/middleware.go` 新增 `withRecover(next http.Handler) http.Handler`。
- [AC-2] WHEN 被包裹的 handler panic THE SYSTEM SHALL 调用 `recover()`，记录 `log.Printf("panic: %v\n%s", r, debug.Stack())`，返回 `500 Internal Server Error` + `Content-Type: application/json` + `{"error":"internal server error"}`。
- [AC-3] WHEN recover 成功 THE SYSTEM SHALL 后续请求仍正常处理（http.Server 不挂死，连接不被 close）。
- [AC-4] THE SYSTEM SHALL recover 后**不**重新 panic；任何后续写入失败（如 client 已断开）忽略，不影响其他请求。
- [AC-5] THE SYSTEM SHALL 中间件自身不引入 panic 路径（类型转换 / nil deref 等都用 `defer func(){ if r := recover(); r != nil { ... } }()` 包裹）。

### withLogger 中间件

- [AC-6] THE SYSTEM SHALL 新增 `withLogger(next http.Handler) http.Handler`。
- [AC-7] WHEN 任意请求完成 THE SYSTEM SHALL 记录一行 `log.Printf("%s %s %d %s", method, path, status, latency)`，例如 `"GET /api/summary 200 12.3ms"`。
- [AC-8] WHEN 响应 status >= 500 THE SYSTEM SHALL 额外记录 `log.Printf("ERROR %s %s %d ...", ...)`（不抛 panic，warn 级别即可）。
- [AC-9] THE SYSTEM SHALL `latency` 用 `time.Since(start)` 计算，精度 ms。
- [AC-10] THE SYSTEM SHALL path 不包含 query string（避免日志被 token 等敏感参数污染）。
- [AC-11] THE SYSTEM SHALL 中间件零开销路径（避免对静态资源做详细日志 —— 静态资源通过 `withCacheHeaders` 已加 `Vary` 头，由 logger 检测 `/assets/*` 时降级为一行简洁日志 `"200 GET /assets/main.HASH.js"`）。

### withCacheHeaders 中间件

- [AC-12] THE SYSTEM SHALL 新增 `withCacheHeaders(next http.Handler) http.Handler`。
- [AC-13] WHEN 请求路径以 `/assets/` 开头 THE SYSTEM SHALL 设置 `Cache-Control: public, max-age=31536000, immutable`。
- [AC-14] WHEN 请求路径以 `/api/` 开头或为 `/api/events`（SSE） THE SYSTEM SHALL 设置 `Cache-Control: no-cache, no-store, must-revalidate` + `Pragma: no-cache`。
- [AC-15] WHEN 请求路径为 `/` THE SYSTEM SHALL 设置 `Cache-Control: no-cache`。
- [AC-16] WHEN 请求路径为 `/legacy` 或 `/legacy/*` THE SYSTEM SHALL 设置 `Cache-Control: no-cache`。
- [AC-17] WHEN 请求路径不在上述分支 THE SYSTEM SHALL 不设置 Cache-Control（透传给 next handler 决定）。
- [AC-18] THE SYSTEM SHALL 中间件使用 `http.Handler.ServeHTTP` 包装，调用 `next.ServeHTTP(w, r)` 前/后通过 `w.Header().Set` 设置头。
- [AC-19] THE SYSTEM SHALL SSE 端点（`/api/events`）即便无中间件保护也应禁用 proxy buffer（由 `Content-Type: text/event-stream` 自身保证；本 spec 不重复实现）。

### handleStatic handler

- [AC-20] THE SYSTEM SHALL 在 `internal/visualize/server.go` 新增 `Server.handleStatic(w, r)` handler。
- [AC-21] WHEN 请求路径以 `/assets/` 开头 THE SYSTEM SHALL 从 `staticFS` 子树读 `static/dist/<trimmed path>`，返回内容 + 对应 `Content-Type`（js→`application/javascript`、css→`text/css`、其他→`application/octet-stream`）。
- [AC-22] WHEN 文件不存在于 embed.FS THE SYSTEM SHALL 返回 `404 Not Found` + `{"error":"asset not found"}`。
- [AC-23] WHEN 路径含 `..`（路径穿越） THE SYSTEM SHALL 返回 `400 Bad Request`（防 `..%2F..`）。
- [AC-24] THE SYSTEM SHALL handler 不读 `static/index.html`（保留 `handleIndex` 单文件入口职责）。
- [AC-25] THE SYSTEM SHALL handler 性能 ≤ 5ms（embed.FS 在内存中，零磁盘 IO）。

### embed.FS 扩展

- [AC-26] THE SYSTEM SHALL `internal/visualize/server.go:42-43` 现有 `//go:embed static/*` 保留不变。
- [AC-27] THE SYSTEM SHALL 在同文件新增 `//go:embed static/dist` 行（**显式列出目录避免误嵌入 `src/`**）。
- [AC-28] WHEN `static/dist/` 目录为空或不存在 THE SYSTEM SHALL `go build` 编译失败（`//go:embed` 强制要求目录非空），提示开发者跑 `make dashboard-dist`。
- [AC-29] THE SYSTEM SHALL `staticFS` 变量类型不变（保持 `embed.FS`），handleStatic 通过 `staticFS.ReadFile` 读子文件。

### server.go 装配

- [AC-30] THE SYSTEM SHALL `NewServer` 中将 mux 用 `withRecover(withLogger(withCacheHeaders(mux)))` 包裹，赋值给 `srv.Handler`。
- [AC-31] THE SYSTEM SHALL mux 注册路由不变（`/`、`/api/summary`、`/api/specs`、`/api/spec/`、`/api/events`），新增路由由 dashboard-backend-api-extensions 提供；本 spec 自身不新增业务端点。
- [AC-32] THE SYSTEM SHALL mux 注册 `/assets/` 路由指向 `handleStatic`（新增）。
- [AC-33] THE SYSTEM SHALL `Shutdown()` 行为不变（dashboard-sse-bugfix 修复保留）。
- [AC-34] THE SYSTEM SHALL 中间件顺序：`withRecover` 在最外层（兜底 panic）→ `withLogger`（记录所有请求含 panic 后的 500）→ `withCacheHeaders`（路径分支）→ `mux`（业务路由）。

### 回归保护 — Unchanged Behavior

- [AC-35] THE SYSTEM SHALL CONTINUE TO stdlib `net/http`，不引入任何第三方 Go 依赖（`golang.org/x/net` 之类也不引入）。
- [AC-36] THE SYSTEM SHALL CONTINUE TO 现有 5 个端点行为不变（`/`、`/api/summary`、`/api/specs`、`/api/spec/<name>`、`/api/events`）。
- [AC-37] THE SYSTEM SHALL CONTINUE TO SSE 事件名 `ping`/`refresh` 不变（dashboard-sse-bugfix 固化）。
- [AC-38] THE SYSTEM SHALL CONTINUE TO `Shutdown()` 5s 超时 + `time.NewTimer` + `defer t.Stop()` 路径不变。
- [AC-39] THE SYSTEM SHALL CONTINUE TO `Server.URL()` IPv6 字面量处理不变（server.go:166-176）。
- [AC-40] THE SYSTEM SHALL CONTINUE TO `Server.Subscribers()` 返回当前 SSE 连接数（本 spec 不修改该方法）。
- [AC-41] THE SYSTEM SHALL CONTINUE TO `Server.watcherOnce` 与 `Server.watcherRunning` 保证 watcher 单次启动的逻辑不变。

### 错误处理

- [AC-42] WHEN `withRecover` 捕获 panic THE SYSTEM SHALL 不调用 `http.Error` 重复写入 body（避免 `superfluous response.WriteHeader` 警告），改为直接 `w.WriteHeader(500)` + `json.NewEncoder(w).Encode(...)`。
- [AC-43] WHEN `withLogger` 写日志失败（stdin 关闭等极端场景） THE SYSTEM SHALL 不影响响应（`log` 包自身保证不 panic）。
- [AC-44] WHEN `withCacheHeaders` 调用 next.ServeHTTP 后才设置 Cache-Control，则 `WriteHeader` 已发送无法修改 header —— 故**必须**在 next.ServeHTTP 之前设置 header；本 spec 实现遵守此顺序。
- [AC-45] WHEN `handleStatic` 读 embed.FS 失败（理论上不会发生） THE SYSTEM SHALL 返回 500 而非 panic。

## Out of Scope

- 业务端点扩展（`/api/spec/<name>/tasks` 等）—— 属于 dashboard-backend-api-extensions。
- fsnotify 替换 2s mtime 轮询 —— 属于 dashboard-realtime-fsnotify。
- SSE heartbeat / 事件分级 —— 属于 dashboard-realtime-fsnotify。
- 前端构建链（esbuild / package.json）—— 属于 dashboard-frontend-foundation。
- 鉴权 / OAuth / API token —— 本地单用户 dashboard，零鉴权需求。
- CORS —— 同源访问，零跨域需求。
- TLS / HTTPS —— 本地 loopback，零 TLS 需求。
- 分布式追踪（OpenTelemetry） —— 轻量级本地工具，过度设计。
