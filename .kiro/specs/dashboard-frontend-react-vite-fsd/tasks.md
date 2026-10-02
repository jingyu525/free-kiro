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
-->

## Wave 1 — 基础工具链（pnpm + Vite + TS + React）

- [ ] #1 `internal/visualize/static/package.json` 改写：声明 `packageManager: pnpm@9.x`、`engines.node >= 20`；runtime deps `react@18.3.x`、`react-dom@18.3.x`；devDeps `vite@5.x`、`@vitejs/plugin-react@4.x`、`typescript@5.6.x`、`@types/react@18`、`@types/react-dom@18`；scripts `dev` / `build` / `preview` / `lint` / `type-check`
- [ ] #2 `internal/visualize/static/tsconfig.json` 改写：启用 `strict`、`noUncheckedIndexedAccess`、`jsx: react-jsx`；target `ES2022`；module `ESNext`；moduleResolution `bundler`；`noEmit: true`；include `src/**/*.{ts,tsx}`
- [ ] #3 `internal/visualize/static/vite.config.ts` 新建：root `src`、base `/assets/`、plugin-react、manualChunks 拆 `react-vendor` / `query-vendor`、target `es2022`、minify `esbuild`、sourcemap 由 mode 决定 — [deps: #1,#2]
- [ ] #4 `internal/visualize/static/.gitignore` 校验：确认 `node_modules/`、`dist/`、`*.log`、`.DS_Store` 全部忽略；新增 `pnpm-lock.yaml` **不忽略**（强制入库）— [deps: #1]

## Wave 2 — shared 层

- [ ] #5 `shared/config/index.ts` 新建：导出 `API_BASE = '/api'`、`ASSET_BASE = '/assets'`，从 `process.env` 读取 Vite define 注入（dev = `'http://localhost:7374/api'`，prod = `'/api'`）— [deps: #3]
- [ ] #6 `shared/api/client.ts` 新建：实现 `fetchJson<T>(url, init)` + `fetchRetry<T>(url, opts)`，分类 `ApiError: TypeError | HttpError(code) | ParseError`；退避 `attempt² * 1000ms`、cap `8000ms`；保留 dashboard-frontend-foundation 的 `api.ts` 全部行为 — [deps: #5]
- [ ] #7 `shared/sse/client.ts` 新建：实现 `SSEClient` 类：EventSource wrapper、订阅 `ping`/`refresh`、emit `open` / `refresh` / `error` / `heartbeat`、30s 无事件自动 close 重连；与 `lib/sse.ts` 行为等价 — [deps: #5]
- [ ] #8 `shared/sse/use-sse-subscription.ts` 新建：React hook，on emit → `queryClient.invalidateQueries(['summary'])` + `['specs']` — [deps: #7]
- [ ] #9 `shared/lib/format.ts` + `shared/lib/hash-router.ts` 新建：`formatRelative` / `formatDelta` / `formatTimeHHMMSS` / `formatNumber`；`useHashRoute()` hook（用 `useSyncExternalStore` 订阅 `hashchange`）— [deps: #3]
- [ ] #10 `shared/ui/{Button,Card,Spinner,Skeleton}.tsx` 新建：可复用 UI kit，无业务语义；样式走 `shared/styles/components.css` — [deps: #3]
- [ ] #11 `shared/styles/{globals,tokens/{color,spacing,typography,radius,shadow},themes/{dark,light},components}.css` 从老 `src/styles/` 迁移，结构不变 — [deps: #3]

## Wave 3 — entities 层

- [ ] #12 `entities/summary/{types,api,use-summary-query}.ts` 新建：`type ProjectReport`（复用 `internal/visualize/report.go` 字段）、`fetchSummary()` 调 `/api/summary`、`useSummaryQuery()` react-query wrapper（staleTime 30s、retry 3）— [deps: #6]
- [ ] #13 `entities/spec/{types,api,use-spec-query}.ts` 新建：`type SpecReport` / `DriftSignal` / `TaskProgress`、`fetchSpec(name)` / `fetchSpecTasks(name)` 调 `/api/spec/<name>/tasks`、`useSpecQuery(name)` react-query wrapper — [deps: #6]
- [ ] #14 `entities/connection/{state,provider}.tsx` 新建：reducer 管理 `'online' \| 'reconnecting' \| 'offline'`、`attempt`、`lastEventAt`；`<ConnectionProvider>` Context 包装 SSEClient — [deps: #7]
- [ ] #15 `app/providers/query-provider.tsx` 新建：`<QueryProvider>` 包装 `QueryClientProvider`，staleTime 30s、retry 3，onError → toast — [deps: #12,#13]

## Wave 4 — features 层

- [ ] #16 `features/theme-toggle/index.tsx` 新建：明暗切换按钮，写 `<html data-theme>` + localStorage 三态（dark/light/auto）— [deps: #14]
- [ ] #17 `features/refresh-dashboard/index.tsx` 新建：refresh 按钮 + 500ms 去抖；`useEffect` 拦截 Ctrl+R / Cmd+R 调 `useQuery.refetch()` — [deps: #12]
- [ ] #18 `features/retry-fetch/index.tsx` 新建：错误 banner（`border-left: 3px solid var(--color-err)`）+ retry 按钮调 `useQuery.refetch()` — [deps: #12]
- [ ] #19 `features/skeleton-overlay/index.tsx` 新建：首屏 6 行骨架（`SkeletonRow × 6`），从 `entities/summary` 读 `isLoading && !data` 决定渲染 — [deps: #12]

## Wave 5 — widgets 层

- [ ] #20 `widgets/header-bar/index.tsx` 新建：标题 + spec count + relativeTime + HH:MM:SS + `<RefreshButton/>` + `<ThemeToggle/>` + `<ConnectionDot/>` — [deps: #16,#17,#14]
- [ ] #21 `widgets/summary-grid/index.tsx` + `widgets/stat-card/index.tsx` 新建：7 张 stat 卡 grid，每卡 `<StatCard label value delta>` — [deps: #12]
- [ ] #22 `widgets/specs-table/index.tsx` + `widgets/phase-badge/index.tsx` 新建：spec 表格（name / phase / approved / drift / tasks），name 列 `<a href="#spec/<name>">` 预留 hash 路由 — [deps: #12,#13]
- [ ] #23 `widgets/empty-state/index.tsx` 新建：根据 `summary.mode` 渲染三态：workspace-missing / no-specs / ok-empty，CTA 文案与 `dashboard-frontend-foundation` 一致 — [deps: #12]
- [ ] #24 `widgets/toast-stack/index.tsx` 新建：Toast 堆叠 + 5s 自动消失；订阅 `ToastContext` — [deps: #29]
- [ ] #25 `widgets/progress-bar/index.tsx` 新建：顶部进度条，summary `isFetching && hasData` 时 250ms 滑动 — [deps: #12]
- [ ] #26 `widgets/skeleton-row/index.tsx` 新建：单行骨架（gradient 1.4s 动画）；接受 `rows` prop — [deps: #11]

## Wave 6 — pages + app 层

- [ ] #27 `pages/dashboard-page/index.tsx` 新建：组合 `<HeaderBar/>` + `<SummaryGrid/>` + `<SpecsTable/>` + `<EmptyState/>` + `<RetryBanner/>` + `<ToastStack/>` + `<ProgressBar/>` + `<SkeletonOverlay/>` — [deps: #20,#21,#22,#24,#25,#26,#18,#19]
- [ ] #28 `app/providers/theme-provider.tsx` 新建：`<ThemeProvider>` Context + `applyTheme()` 写 `<html data-theme>` + localStorage + matchMedia 监听 — [deps: #11]
- [ ] #29 `app/providers/toast-provider.tsx` 新建：`<ToastProvider>` Context + reducer（`{toasts: Toast[]}`，push / dismiss）— [deps: #3]
- [ ] #30 `app/router/index.tsx` 新建：`<AppRouter>` 包装 `<Route path="/" element={<DashboardPage/>} />`，`#spec/<name>` 路由预留 — [deps: #27]
- [ ] #31 `app/index.tsx` 新建：`ReactDOM.createRoot(document.getElementById('app')!).render(<Root />)`；`<Root>` 套 `<QueryProvider>` + `<ThemeProvider>` + `<ConnectionProvider>` + `<ToastProvider>` + `<AppRouter>` — [deps: #15,#28,#29,#30]

## Wave 7 — 验证 + 清理

- [ ] #32 `pnpm exec tsc --noEmit` 通过，0 error；记录命令到 Makefile `dashboard-typecheck` target — [deps: #31]
- [ ] #33 `pnpm build` 通过：产 `dist/assets/index-<hash>.{js,css}` + `dist/index.html`；Makefile `dashboard-dist` target 改为 `pnpm build && go install ./cmd/free-kiro` — [deps: #31]
- [ ] #34 跑 `scripts/smoke-dashboard.sh`（dashboard-frontend-foundation 已建，需更新断言路径 `/assets/index-<HASH>.js`）+ 新增 FSD 分层检查 `scripts/check-fsd-layers.mjs`（解析 import → 阻断反向依赖）— [deps: #32,#33]
- [ ] #35 Playwright E2E 验证：刷新按钮 / 主题切换 / Ctrl+R / SSE 推送（`touch .kiro/specs/<name>/requirements.md` 2s 内收到 `refresh`）— [deps: #33]
- [ ] #36 删除老文件：`esbuild.config.mjs`、`src/main.ts`（除 flat 这一个老入口已迁移至 `app/index.tsx`）、`src/lib/signal.ts`（已用 react-query 替代）、`src/components/`（已迁移至 `widgets/`）、`src/stores/`（已迁移至 `entities/` + Context）；保留 `src/types.ts` 仅作字段注释（看是否还有 import，决定删除还是保留）— [deps: #33]
- [ ] #37 `static/legacy/index.html` 保留（dashboard-frontend-foundation 已建，作为回滚路径）；`static/index.html` 校验 `<div id="app">` 入口 + `<link>` / `<script>` 指向 `/assets/index-<hash>.{css,js}`（Vite 注入）；`//go:embed static/*` 不需改动 — [deps: #33]