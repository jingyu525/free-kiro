# dashboard-backend-middleware — Tasks

<!--
依赖规则：
  - #1 middleware.go 是所有后续任务的前置
  - #2-#3 handler 与 #4 embed 改动可与 #1 并行
  - #5 路由注册在 #3 完成后做
  - #6 测试在 #1-#5 完成后
  - #7 验证回归必须在 #6 完成后
-->

- [x] #1 middleware.go 三个中间件：新建 `internal/visualize/middleware.go`，实现 `withRecover`（panic → 500 + JSON body + log stack）+ `withLogger`（method/path/status/latency 日志 + `statusRecorder` 包装 ResponseWriter）+ `withCacheHeaders`（按前缀分支：`/assets/*` 1y immutable、`/api/*` 与 `/` no-cache、`/api/events` no-cache + no-store）；文件 ≤ 100 行
- [x] #2 `Server.handleStatic` handler：在 `internal/visualize/server.go` 新增方法，从 embed.FS 读 `static/dist/<path>`，按扩展名设置 Content-Type；防路径穿越（`..` 或 `/` 开头）400；method != GET 返回 405 + Allow 头；文件 ≤ 50 行
- [x] #3 `Server` 装配中间件链：改 `NewServer`，用 `withRecover(withLogger(withCacheHeaders(mux)))` 包裹 mux，赋值给 `srv.Handler`；新增 mux 路由 `/assets/` → `handleStatic`
- [x] #4 embed.FS 扩 `static/dist`：实测发现 `//go:embed static/*` 已隐式覆盖 `static/dist/`，无需新增 directive；保留 `//go:embed static/*` 不变；spec-1 前端基建落地后 dist/ 会自动被 embed
- [x] #5 statusRecorder 实现：在 `middleware.go` 内定义 `type statusRecorder struct { http.ResponseWriter; status int; wroteHeader bool }`，实现 `WriteHeader(int)` 抓 status + `Flush()` 转发到 `http.Flusher`（dashboard-realtime-fsnotify SSE 兼容）
- [x] #6 单测（覆盖率新增行 ≥ 80%）：写 `middleware_test.go`（withRecover 三种 panic + next 调用 + logger 日志抓取 + withCacheHeaders 8 分支 + statusRecorder 三种 case = 16 子测试）+ `server_static_test.go`（NotFound + PathEscape 4 子 + NonGetMethod 4 子 + ContentType 8 子 + EmptyPath = 18 子测试）；全部 PASS
- [x] #7 端到端验证：启 `free-kiro serve --port 7374` + curl 验证：`/` → 200 + no-cache ✅、`/api/summary` → 200 + no-cache ✅、`/api/events` → 200 + no-cache, no-store, must-revalidate ✅、`/assets/nonexistent.js` → 404 + immutable ✅、`/assets/..%2F..` → 400 + immutable ✅（path escape 拦截）、`POST /assets/main.js` → 405 + Allow: GET + immutable ✅。修复 server_sse.go 删 handleEvents 自己设的 Cache-Control 让中间件生效 [deps: #1,#2,#3,#4,#5,#6]
- [x] #8 回归保护：`go test -race ./internal/visualize/...` 全部 PASS（V1/S1/S8/URL + 所有新增测试）；`free-kiro status --json` 端到端 PASS（dashboard-sse-bugfix 兼容）；`go vet` 干净；新版 binary 验证 [deps: #7]
