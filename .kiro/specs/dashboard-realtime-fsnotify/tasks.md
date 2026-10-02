# dashboard-realtime-fsnotify — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

Wave 划分（按依赖自动计算，free-kiro task list 可视化）：
  - Wave 1：#1-#3   后端 fsnotify（server_sse + watch 适配）
  - Wave 2：#4-#6   后端 ETag（etag helper + middleware + 8 端点接入）
  - Wave 3：#7-#10  前端 fetchJsonWithEtag + react-query 接入
  - Wave 4：#11-#13 dialog React #185 修复 + 重新挂载 + CSS（无新增）
  - Wave 5：#14-#16 E2E + Go 单测 + commit + complete
-->

## Wave 1 — 后端 fsnotify

- [ ] #1 `internal/visualize/etag.go` 新建：`etagFor(body []byte) string` (`fmt.Sprintf("%x", sha256.Sum256(body))`) + `setETagHeader(w, body)` (写 header) — [deps: —]
- [ ] #2 `internal/visualize/etag_test.go` 新建：TestEtagFor_Deterministic + TestEtagFor_DiffersForDiffInput + TestSetETagHeader_WritesCorrectValue — [deps: #1]
- [ ] #3 `internal/visualize/server_sse.go` 替换 mtime ticker → `internal/watch.Watcher` 调用；保留 `watcherOnce / watcherRunning / watcherStartCount`；fsnotify init 失败 fallback mtime 2s tick — [deps: —]

## Wave 2 — 后端 ETag middleware

- [ ] #4 `internal/visualize/middleware.go` 加 `withETag(next http.HandlerFunc)`：拦截响应 body bytes → 计算 ETag → 写 `ETag` header → 检查 `If-None-Match` 头若匹配返回 304 + 不写 body；匹配写 body — [deps: #1]
- [ ] #5 `internal/visualize/server.go` 把 `withETag` 插入中间件链（在 `withCacheHeaders` 之后 / mux 之前）；8 个 handler（summary/specs/spec/spec/<n>/tasks/drift/timeline/health/hooks）自动获得 ETag 处理 — [deps: #4]
- [ ] #6 `internal/visualize/middleware_test.go`（已有）加 case：TestWithETag_IfNoneMatch_Matches → 304 + body 空；TestWithETag_NoHeader → 200 + ETag；TestWithETag_StaleETag → 200 + ETag — [deps: #4]

## Wave 3 — 前端 ETag fetch

- [ ] #7 `shared/api/client.ts` 加 `fetchJsonWithEtag<T>(url, opts)`：用 closure Map 缓存 `etag`；带 `If-None-Match` header；304 时返回 `null`；新 ETag 时更新 cache；类型：`EtagCache<T> = { etag: string; data: T }` — [deps: —]
- [ ] #8 `entities/summary/use-summary-query.ts` 改用 `fetchJsonWithEtag`；保留 normalize select（dashboard-frontend-components 已落地） — [deps: #7]
- [ ] #9 `entities/spec/use-spec-query.ts`（useSpecTasksQuery / useSpecDriftQuery）改用 `fetchJsonWithEtag` — [deps: #7]
- [ ] #10 `entities/spec/api.ts` `fetchSpecOverview` 改用 `fetchJsonWithEtag` — [deps: #7]

## Wave 4 — dialog React #185 修复

- [ ] #11 `widgets/spec-detail-dialog/index.tsx` 内部不再调 `useSummaryQuery()`；改用 `useQueryClient().getQueryData(SUMMARY_QUERY_KEY)` 直读 cache snapshot — [deps: —]
- [ ] #12 `app/index.tsx` 重新挂载 `<SpecDetailDialog />`（撤销 dashboard-spec-detail-view 的临时 unmount）；恢复 `<StrictMode>` 包 Root（之前为排查 #185 移除） — [deps: #11]
- [ ] #13 production build (`pnpm build`) 验证 React error #185 不再触发；Playwright 浏览器 navigate `#/spec/<name>?tab=overview` → dialog 打开 + 0 console error — [deps: #11,#12]

## Wave 5 — 验证 + commit

- [ ] #14 `scripts/e2e-dashboard.mjs` 加场景：9) GET `/api/summary` + `If-None-Match: <etag>` → 304；10) GET `/api/specs` + ETag cycle 同样 — [deps: #4,#7]
- [ ] #15 `go test ./internal/visualize/...` 0 error；`go test ./internal/watch/...` 0 error；`pnpm exec tsc --noEmit` 0 error；`pnpm run lint:fsd` ✓；`pnpm build` OK — [deps: #1-#14 全部]
- [ ] #16 commit + `free-kiro spec complete dashboard-realtime-fsnotify` — [deps: #15]