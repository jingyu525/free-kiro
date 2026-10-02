# dashboard-realtime-fsnotify

<!--
目标：把 free-kiro serve dashboard 的实时更新链路从 2 秒 mtime 轮询
升级到 fsnotify（OS 级 inotify/FSEvents）+ 给 7 个 API 端点加
ETag/If-None-Match → 304 缓存复用。同时承担 dashboard-spec-detail-view
遗留的 spec 详情 dialog 在 prod 模式 React #185 死循环 bug 的修复
（dialog 组件代码已存在但 app/index.tsx 注释挂载）。

零破坏：
- 保留 dashboard-sse-bugfix 的 ping/refresh 事件名 + 后端 SSE 链路
- 保留 dashboard-backend-api-extensions 的 5 个端点契约
- 保留 dashboard-frontend-react-vite-fsd 的 FSD 6 层 + react-query 5
- 保留 dashboard-frontend-components 的 normalize + a11y

不在本 spec 范围：
- 移动端深度适配（dashboard-mobile-a11y）
- 写操作（dashboard 当前 read-only viewer）
- WebSocket 替代 SSE
-->

## User Stories

- As a dashboard 使用者 I want 改 .kiro/specs 文件后 < 200ms 看到 dashboard 刷新 so that 实时反馈我改了什么。
- As a dashboard 使用者 I want 重复打开 dashboard 时不重复下载未变的 spec so that 网络与服务器负载更省（304）。
- As a dashboard 维护者 I want spec 详情 dialog 在 production build 也能正常打开 so that 我能看到 phase / drift / tasks / timeline。
- As a Go 维护者 I want visualize 包复用 internal/watch 的 fsnotify watcher so that `free-kiro watch` 与 `free-kiro serve` 用同一个文件监听实现，避免分叉维护。

## Acceptance Criteria

### fsnotify 替换 mtime 轮询

- [AC-1] THE SYSTEM SHALL `internal/visualize/server_sse.go` 的 `watchChanges` 把当前 mtime polling ticker 替换为 `internal/watch.Watcher`（已 fsnotify-based）；`.kiro/specs/<name>/{requirements,design,tasks}.md` 任一文件 `Write/Create/Remove/Rename` 事件在 < 200ms 内触发 SSE broadcast `refresh` 事件。
- [AC-2] WHEN `.kiro/specs/<name>/` 目录被新增或删除 THE SYSTEM SHALL SSE broadcast `refresh`（fsnotify Create/Remove on directory）。
- [AC-3] THE SYSTEM SHALL fsnotify watcher debounce 300ms（与 `free-kiro watch --preset reactive --debounce 300ms` 默认值一致），避免编辑器 truncate+write 两次事件触发两次 broadcast。
- [AC-4] IF fsnotify 初始化失败（kernel 资源耗尽）THEN THE SYSTEM SHALL fallback 到 2s mtime polling（保留现有 watchTickInterval 路径），不返回 `CODESIGNATURE/IO` 错误给 SSE 客户端。
- [AC-5] THE SYSTEM SHALL `internal/visualize/server_sse.go` 现有 `watcherOnce sync.Once`、`watcherRunning chan struct{}`、`watcherStartCount atomic.Int32` 保持不变（dashboard-sse-bugfix 回归保护）。

### ETag/If-None-Match → 304

- [AC-6] THE SYSTEM SHALL `/api/summary`、`/api/specs`、`/api/spec/<name>`、`/api/spec/<name>/tasks`、`/api/spec/<name>/drift`、`/api/spec/<name>/timeline`、`/api/health`、`/api/hooks` 8 个端点响应含 `ETag` header，值为内容 SHA-256 hex（`fmt.Sprintf("%x", sha256.Sum256(body))`）。
- [AC-7] WHEN 客户端带 `If-None-Match: <etag>` 头且 ETag 匹配当前响应 THE SYSTEM SHALL 返回 304 Not Modified 且 body 为空，`ETag` header 仍存在。
- [AC-8] WHEN 客户端带 `If-None-Match: <stale-etag>` 头且 ETag 不匹配 THE SYSTEM SHALL 返回 200 + 完整 body + 当前 ETag（标准 200 路径）。
- [AC-9] THE SYSTEM SHALL ETag 计算复用现有 handler 不重复 JSON encoding：维护 `etagByHash` map，handler render 后写 ETag；或单次 `BuildReport()` 后 SHA-256 body bytes。
- [AC-10] THE SYSTEM SHALL 304 响应不写 `Content-Length` body 但保留 `Content-Type` 与 `ETag` headers。

### 前端 fetch + ETag 缓存

- [AC-11] THE SYSTEM SHALL `shared/api/client.ts` 的 `fetchJson<T>` 接收 `If-None-Match` header（可选），304 时返回 `null`（让调用方决定是否复用缓存）。
- [AC-12] THE SYSTEM SHALL `entities/summary/use-summary-query.ts`、`entities/spec/use-spec-query.ts` 等 react-query hooks 配置 `staleTime: 5_000`（缩短）+ 接入 fetchJson If-None-Match 流程：成功 fetch 后缓存 ETag，下次 fetch 时带 `If-None-Match: <etag>`，304 不更新 React Query cache（仍返回上次 data）。
- [AC-13] THE SYSTEM SHALL SSE `refresh` 事件触发 `queryClient.invalidateQueries(['summary'])` + `['specs']` + `['spec', name, ...]` 强制 refetch（不等待 staleTime）。

### spec 详情 dialog 重新挂载 + 修 React #185

- [AC-14] THE SYSTEM SHALL `app/index.tsx` 重新挂载 `<SpecDetailDialog />`（撤销 dashboard-spec-detail-view 临时 unmount）。
- [AC-15] THE SYSTEM SHALL `widgets/spec-detail-dialog/index.tsx` 内部不再直接调 `useSummaryQuery()`（避免 react-query 5 useSyncExternalStore forceStoreRerender 循环）；改用 `useQueryClient().getQueryData(SUMMARY_QUERY_KEY)` 直读 cache snapshot — 不订阅、不强制 re-render。
- [AC-16] THE SYSTEM SHALL dialog 自身 tab 切换（overview/drift/tasks/timeline）走 `navigateToHash({name, tab})`，不触发 react-query 重新 fetch；tab 内的 `useSpecTasksQuery(name)` 等 react-query hooks 走正常 React Query 订阅通道。
- [AC-17] THE SYSTEM SHALL production build（`pnpm build` 后由 `go install ./cmd/free-kiro` 嵌入二进制）启动后，`/api/events` SSE 连接 + URL `?tab=overview` 切到 spec 详情 dialog 不再触发 React error #185；浏览器 console 0 React 错误。

### a11y 增量

- [AC-18] THE SYSTEM SHALL dialog `<dialog>` 元素使用原生 `<dialog>` + `showModal()` + Esc 关闭 + backdrop 关闭；保留 dashboard-frontend-components 已落地的 `announce()` 屏幕阅读器支持。
- [AC-19] THE SYSTEM SHALL dialog 内 tab 用 WAI-ARIA Roving Tabindex 模式（ArrowLeft/Right/Home/End 切换）；保留 dashboard-spec-detail-view 已落地的 `data-tab` + `aria-selected` + `tabIndex={current ? 0 : -1}` 模式。

### 文件迁移映射

- [AC-20] THE SYSTEM SHALL `internal/visualize/server_sse.go` 替换 mtime ticker → `internal/watch.Watcher`；保留 `Watch()` 接口（`Add(root, onStop)` 风格适配）。
- [AC-21] THE SYSTEM SHALL `internal/visualize/server.go` 新增 `etagFor(body []byte) string` helper + 在 8 个 handler 入口调 `setETag(w, body)`。
- [AC-22] THE SYSTEM SHALL `internal/visualize/middleware.go` 加 `withETag` 中间件（与 `withCacheHeaders` 类似，对响应 SHA-256）；接受 `If-None-Match` 时返回 304。
- [AC-23] THE SYSTEM SHALL `shared/api/client.ts` 加 `fetchJsonWithEtag<T>` 包装类型，缓存 ETag 到 closure 变量；调用方传 `queryKey` 即可。

## Out of Scope

- WebSocket 替代 SSE（保留 dashboard-sse-bugfix 的 ping/refresh 事件名）
- fsnotify 在容器/网络文件系统的支持（kernel 不支持时走 mtime fallback，AC-4）
- Etag 跨实例同步（per-process 计算即可，AC-41）
- spec 详情 dialog 内的写操作（dashboard 当前 read-only viewer）
- 移动端深度适配（dashboard-mobile-a11y）
- 单元测试覆盖率提升至 80%（POLICY §1 已 70%；CI 跑 `go test -race -coverprofile` 不破即可）