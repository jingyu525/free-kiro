# dashboard-realtime-fsnotify — Design

<!--
决策记录（30 天后回看"为什么这么做"）：
  - 用 internal/watch.Watcher（已 fsnotify）而不是新依赖：避免引入新 lib，watch 与
    serve 两个命令统一同一个文件监听实现
  - ETag 用 SHA-256 over response bytes：与现有 middleware chain 兼容；不引入 weak ETag /
    W 3c 等额外语义
  - Dialog 不再订阅 useSummaryQuery：之前 prod 模式 React #185 死循环根因为 dialog
    + react-query 5 useSyncExternalStore forceStoreRerender 互锁；改读 cache snapshot
    后 dialog 不参与 react-query 订阅通道
  - 304 fallback：当 fsnotify 不可用（容器 / 远程 FS）保留 mtime polling 2s tick
    作为 graceful degradation
  - SSE refresh 仍 invalidate queries：ETag 304 不会触发 React Query cache 改；
    只有 fsnotify 推送 refresh 才强制 invalidate
-->

## Architecture

### fsnotify 替换 mtime 轮询

```
┌────────────────────────────────────────────────────────────────────┐
│ Before (dashboard-frontend-foundation / dashboard-sse-bugfix):      │
│   watchTickInterval = 2s                                            │
│   for every tick {                                                  │
│     mtimes := walk(.kiro/).mtimes()                                 │
│     if changed: broadcast("refresh")                                │
│   }                                                                 │
└────────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────────┐
│ After (dashboard-realtime-fsnotify):                               │
│   w := watch.New(debounce=300ms)                                    │
│   w.Add(.kiro/)                                                    │
│   w.OnEvent(func(ev) { broadcast("refresh") })                     │
│   IF fsnotify.NewWatcher fails → fall back to mtime ticker          │
└────────────────────────────────────────────────────────────────────┘
```

**关键不变量**（来自 dashboard-sse-bugfix 回归保护）：
- `watcherOnce sync.Once` 保证 watcher 单次启动
- `watcherRunning chan struct{}` Shutdown() 阻塞等 watcher 退出
- `watcherStartCount atomic.Int32` 测试用
- SSE 事件名仍为 `ping` / `refresh`
- 客户端 `/api/summary` 5s polling fallback 保留（dashboard-sse-bugfix V1）

### ETag/If-None-Match → 304

```
客户端                                 中间件链 (withETag)
  │                                          │
  ├─ GET /api/summary                          │
  │  If-None-Match: "abc123" ───────────────►   │
  │                                          ├─ handler.render() → bytes
  │                                          ├─ etag = sha256(bytes).hex
  │                                          ├─ If etag == "abc123":
  │   ◄── 304 Not Modified ────────────────  │     return 304 + etag
  │   body 空                                │   Else:
  │                                          ├─ Set ETag header
  │   ◄── 200 OK + body ───────────────────  │     return 200 + body + etag
```

### 端点覆盖（8 端点全部启用 ETag）

| 端点 | handler | ETag source |
|---|---|---|
| `/api/summary` | `handleSummary` | BuildReport() JSON bytes |
| `/api/specs` | `handleSpecs` | StatusForList() JSON bytes |
| `/api/spec/<name>` | `handleSpec` | status JSON bytes |
| `/api/spec/<name>/tasks` | `handleSpecTasks` | tasks JSON bytes |
| `/api/spec/<name>/drift` | `handleSpecDrift` | drift JSON bytes |
| `/api/spec/<name>/timeline` | `handleSpecTimeline` | timeline JSON bytes |
| `/api/health` | `handleHealth` | health JSON bytes |
| `/api/hooks` | `handleHooks` | hooks JSON bytes |

### 前端 fetch + ETag 缓存

```
fetchJsonWithEtag<T>(url, opts):
  cached = queryClient.getQueryData<EtagCache<T>>([url])
  headers = opts.headers ?? {}
  if cached?.etag:
    headers["If-None-Match"] = cached.etag
  response = await fetchJson<T>(url, { ...opts, headers, allow404: true })
  if response === null:    // 304 path
    return cached?.data
  newEtag = response.headers["ETag"]
  queryClient.setQueryData([url], { etag: newEtag, data: response.data })
  return response.data
```

### Dialog React #185 修复

**之前（dashboard-spec-detail-view）**：
```tsx
export function SpecDetailDialog() {
  const route = useHashRoute();        // useSyncExternalStore
  const { data } = useSummaryQuery();  // react-query 5 useSyncExternalStore
  // 双 useSyncExternalStore 订阅同一 queryKey，prod minified 触发 forceStoreRerender
  // 循环 → React error #185
}
```

**之后（dashboard-realtime-fsnotify）**：
```tsx
export function SpecDetailDialog() {
  const route = useHashRoute();
  const qc = useQueryClient();
  // 读 cache snapshot — 不订阅；changes 由 SSE refresh + invalidate 推送
  const data = qc.getQueryData(SUMMARY_QUERY_KEY);
  ...
}
```

这样 dialog 只订阅 hash 路由（useSyncExternalStore 由 hash-router 自己管），不再与 react-query 5 store 互锁。

## Components

### 后端（internal/visualize/）

| Module | 改动 |
|---|---|
| `server_sse.go` | 删 mtime ticker；改用 `internal/watch.Watcher`；保留 watcherOnce/watcherRunning/watcherStartCount |
| `middleware.go` | 加 `withETag(next)` 中间件 |
| `server.go` | 把 `withETag` 插入中间件链（在 `withCacheHeaders` 之后） |
| `etag.go` 新建 | `etagFor(body []byte) string` + `setETagHeader(w, body)` helpers |

### 前端（internal/visualize/static/src/）

| Module | 改动 |
|---|---|
| `shared/api/client.ts` | 加 `fetchJsonWithEtag<T>` + `EtagCache<T>` 类型 |
| `entities/summary/use-summary-query.ts` | 用 `fetchJsonWithEtag` 替代 `fetchJson` |
| `entities/spec/use-spec-query.ts` | 同上 |
| `widgets/spec-detail-dialog/index.tsx` | 用 `useQueryClient().getQueryData()` 替代 `useSummaryQuery()` |
| `app/index.tsx` | 重新挂载 `<SpecDetailDialog />`（撤销 dashboard-spec-detail-view 临时 unmount） |

## Data Model

无 Go 端 schema 变更。无 wire format break（304 与 200 都带 ETag header）。

新增 Go 类型：
- `internal/visualize/etag.go`：`etag string` 类型 + `etagFor(body []byte) string`

新增前端类型：
- `EtagCache<T> = { etag: string; data: T }`

## Testing Strategy

### 类型检查
```bash
pnpm exec tsc --noEmit                # 0 error
go build ./...                       # 0 error
```

### E2E（HTTP-level）
```bash
# scripts/e2e-dashboard.mjs 加场景：
9) GET /api/summary + If-None-Match: <etag> → 304
10) GET /api/summary 不带 header → 200 + ETag header
```

### 单元测试（Go）
```go
// internal/visualize/etag_test.go
func TestEtagFor_Deterministic()
func TestWithETag_IfNoneMatch_Matches()
func TestWithETag_IfNoneMatch_Differs()
func TestWithETag_NoHeader_PassesThrough()
```

### fsnotify 集成测试
```go
// internal/visualize/server_sse_test.go (新增 case)
func TestWatchChanges_FsnotifyEvent_BroadcastsRefresh()
func TestWatchChanges_FsnotifyInitFails_FallsBackToPolling()
```

### Smoke（手动）
```bash
# 1. touch requirements.md → 浏览器 SSE 2s 内收 refresh（< 200ms 实际）
touch .kiro/specs/foo/requirements.md
# 2. 重复打开 /api/summary with same ETag → 304
curl -i -H "If-None-Match: <上次 ETag>" http://localhost:7374/api/summary
# 3. dialog 重新挂载 → URL `#/spec/foo?tab=overview` 不再抛 React #185
# 浏览器 navigate → 0 console error
```

### Bundle 预算
```bash
gzip -c dist/assets/index-*.js | wc -c   # 入口 ≤ 32 KB（新增 fetchJsonWithEtag + dialog fix）
```

## Compatibility / Rollout

### 兼容性边界
- 后端 ETag 是新加的，旧客户端忽略 → 兼容
- 后端 SSE 事件名不变 → 兼容
- 前端 dialog 重新挂载 — 行为不变，只是修 bug → 兼容
- fsnotify 在 macOS / Linux 正常；Windows 需要 fsnotify 1.6+（已 lock）

### 风险与回滚

- **风险 A**：fsnotify 在某些容器（macOS docker volume 或 NFS）不可用。**对策**：AC-4 fallback mtime polling 2s tick；watcher 初始化 try/catch。
- **风险 B**：ETag SHA-256 over response bytes 增加 CPU。**对策**：handler 直接对 bytes 计算（无 parse）；8 端点每次请求 < 100µs。
- **风险 C**：Dialog React #185 修复不彻底（仍 forceStoreRerender）。**对策**：本 spec 落地 dialog 改用 `getQueryData()` 不订阅；如仍崩，加 React DevTools Profiler 确认循环位置。
- **风险 D**：前端 fetchJsonWithEtag 与 SSE invalidate 互不协调（304 + invalidate 同时发生）。**对策**：invalidate 永远覆盖 cache；fetchJsonWithEtag 仅在 staleTime 内有效。

### 落地步骤

```
1. internal/visualize/etag.go 新建 (etagFor / setETagHeader)
2. internal/visualize/etag_test.go 新建
3. internal/visualize/middleware.go 加 withETag 中间件
4. internal/visualize/server.go 把 withETag 插入中间件链
5. internal/visualize/server_sse.go 用 internal/watch.Watcher 替换 ticker
6. internal/visualize/server_sse_test.go 加 fsnotify event case
7. shared/api/client.ts 加 fetchJsonWithEtag
8. entities/{summary,spec}/use-summary-query.ts / use-spec-query.ts 用 fetchJsonWithEtag
9. widgets/spec-detail-dialog/index.tsx 改 getQueryData
10. app/index.tsx 重新挂载 <SpecDetailDialog />
11. scripts/e2e-dashboard.mjs 加 304 场景
12. pnpm tsc + lint:fsd + build
13. go build + go test ./internal/visualize/...
14. 浏览器 Playwright 验证 dialog + URL ?tab + 304 + fsnotify < 200ms
15. commit + spec complete
```