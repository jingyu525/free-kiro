# dashboard-frontend-react-vite-fsd

<!--
目标：把 free-kiro serve dashboard 的前端从 esbuild + vanilla TS + npm 迁移到
Vite 5 + React 18 + TypeScript strict + pnpm 9，并按 Feature-Sliced Design (FSD)
分层重构 src/（app / pages / widgets / features / entities / shared）。

零后端 API 变更：保留 dashboard-sse-bugfix 的 SSE 事件名（ping/refresh）、
dashboard-backend-api-extensions 的 5 个端点 + ProjectReport.mode 三态字段；
保留 //go:embed static/* + handleStatic Cache-Control: immutable 链路。

本次不修 TypeError null.length runtime bug — 该 bug 移至
dashboard-frontend-components（PR#3 已记入 commit message）。
本次不补 spec 详情弹层、waves 可视化、drift 详情面板 — 属于
dashboard-spec-detail-view 独立 spec。
本次不做移动端深度适配 — 属于 dashboard-mobile-a11y。
-->

## User Stories

- As a dashboard 维护者 I want 前端构建链迁移到 Vite + pnpm so that HMR < 200ms、依赖装得更快、CI 缓存更可靠。
- As a dashboard 维护者 I want 源码按 FSD（app / pages / widgets / features / entities / shared）分层 so that 后续加 spec 详情视图、drift 详情、移动端适配时改动半径最小。
- As a dashboard 维护者 I want 用 React 18 + 严格 TypeScript so that 组件树可调试、SSR-ready、props 类型即文档。
- As a dashboard 使用者 I want 渲染链路保持等价 so that 迁移前后 SSE 推送、错误重试、双主题切换、loading skeleton、空状态三态都不断。
- As a Go 维护者 I want //go:embed static/dist 嵌入产物不破 so that `go build` 不需要任何额外 npm 步骤就能跑（除 dashboard-dist target 之外）。

## Acceptance Criteria

### 构建链迁移（pnpm + Vite + React 18 + TS strict）

- [AC-1] THE SYSTEM SHALL 在 `internal/visualize/static/package.json` 声明 `"packageManager": "pnpm@9.x"` 与 `"engines": { "node": ">=20" }`，并提交 `pnpm-lock.yaml` 到 git。
- [AC-2] WHEN 开发者运行 `pnpm install` THE SYSTEM SHALL 在 60 秒内完成（首次冷启不计），CI 用 `pnpm install --frozen-lockfile`。
- [AC-3] THE SYSTEM SHALL 在 `package.json` 声明运行时依赖 `react`、`react-dom`（版本固定到 patch 级），开发依赖 `vite`、`@vitejs/plugin-react`、`typescript`、`@types/react`、`@types/react-dom`。
- [AC-4] THE SYSTEM SHALL 启用 TypeScript `"strict": true`、`"noUncheckedIndexedAccess": true`、`"jsx": "react-jsx"`，CI 跑 `pnpm exec tsc --noEmit` 必须 0 error。
- [AC-5] WHEN 开发者运行 `pnpm dev` THE SYSTEM SHALL 启动 dev server 于 `http://localhost:5173`（或下一个空闲端口），HMR 修改单文件到浏览器更新 ≤ 500ms。
- [AC-6] WHEN 开发者运行 `pnpm build` THE SYSTEM SHALL 调用 `vite build` 把 `index.html` 与所有依赖打包到 `dist/assets/`，并输出至少 `dist/assets/index-<hash>.js` 与 `dist/assets/index-<hash>.css`。
- [AC-7] THE SYSTEM SHALL `dist/assets/index-<hash>.js` gzip 后 ≤ 80 KB（包含 React 18 runtime + dashboard 代码），`index-<hash>.css` gzip 后 ≤ 12 KB，`index.html` ≤ 4 KB。
- [AC-8] THE SYSTEM SHALL 在 `static/.gitignore` 包含 `node_modules/` 与 `dist/`，`pnpm-lock.yaml` **必须**入库。
- [AC-9] WHERE `process.env.NODE_ENV === 'production'` THE SYSTEM SHALL Vite 在 `vite.config.ts` 设置 `build.minify: 'esbuild'`、`build.sourcemap: false`、`build.target: 'es2022'`。

### FSD 分层重构（强制 import 方向）

- [AC-10] THE SYSTEM SHALL 在 `internal/visualize/static/src/` 下按 FSD 标准 6 层建目录：`app/`、`pages/`、`widgets/`、`features/`、`entities/`、`shared/`，每层一个 index barrel 文件。
- [AC-11] THE SYSTEM SHALL import 方向严格自顶向下：`app → pages → widgets → features → entities → shared`，下层不得 import 上层（`eslint-plugin-boundaries` 或等价 lint 强制）。
- [AC-12] THE SYSTEM SHALL `shared/` 层包含 `shared/api/`（fetchJson / fetchRetry）、`shared/sse/`（SSEClient）、`shared/config/`（API base URL）、`shared/ui/`（Button / Card / Spinner / Skeleton）、`shared/lib/`（format / hash-router）、`shared/styles/`（tokens / globals）。
- [AC-13] THE SYSTEM SHALL `entities/` 层包含 `entities/spec/`（types 与 fetcher）、`entities/summary/`（types 与 fetcher）、`entities/connection/`（types 与 SSE 状态机）。
- [AC-14] THE SYSTEM SHALL `features/` 层包含 `features/theme-toggle/`、`features/refresh-dashboard/`、`features/retry-fetch/`、`features/skeleton-overlay/`。
- [AC-15] THE SYSTEM SHALL `widgets/` 层包含 `widgets/header-bar/`、`widgets/summary-grid/`、`widgets/specs-table/`、`widgets/empty-state/`、`widgets/toast-stack/`、`widgets/progress-bar/`、`widgets/skeleton-row/`。
- [AC-16] THE SYSTEM SHALL `pages/` 层包含 `pages/dashboard-page/`（单页面 dashboard，hash 路由 `#spec/<name>` 预留位置）。
- [AC-17] THE SYSTEM SHALL `app/` 层包含 `app/providers/`（React Query / QueryClient / ThemeProvider / ToastProvider）、`app/router/`（hash 路由）、`app/index.tsx`（ReactDOM.createRoot 挂载点）。
- [AC-18] WHEN 任一文件跨层违规 import（上 → 下反向）THEN `pnpm lint` SHALL 报错并 exit 1。

### React 组件化与状态管理

- [AC-19] THE SYSTEM SHALL 所有 widget 渲染为 React FC（function component），每个 widget 在自己的目录内有 `index.ts` + `ui.tsx` + `model.ts`（如需本地 store）。
- [AC-20] THE SYSTEM SHALL 全局 server state 用 `@tanstack/react-query` 管理（`useQuery` / `useMutation`），客户端 UI state 用 React `useState` / `useReducer`，跨页面共享 state 走 `React.Context`。
- [AC-21] WHEN `summary` query 处于 `loading` 且 `data` 为 `undefined`（首次加载） THE SYSTEM SHALL 在表格区显示 `SkeletonRow × 6`（背景 `--color-skel-base` 与 `--color-skel-shine` 1.4s 渐变动画）。
- [AC-22] WHEN `summary` query 处于 `loading` 且 `data` 已存在（刷新中）THE SYSTEM SHALL 保留旧数据渲染，顶部显示 `<div class="top-bar-progress">` 进度条 250ms 滑动。
- [AC-23] WHEN `summary` query 处于 `error` THE SYSTEM SHALL 在 HeaderBar 下方显示红色 banner（`border-left: 3px solid var(--color-err)`），文案为 "Couldn't refresh dashboard." 并附带 retry 按钮，点击后调 `summary.refetch()`。

### Design Tokens 保留 + 双主题

- [AC-24] THE SYSTEM SHALL 在 `:root[data-theme="dark"]` 与 `:root[data-theme="light"]` 分别定义全部 CSS custom properties，dark 主题 `--color-fg` (`#c9d1d9`) on `--color-bg` (`#0e1116`) 对比度 ≥ 11:1，light 主题对应 ≥ 14:1。
- [AC-25] WHEN 用户在 HeaderBar 点击 theme toggle 按钮 THE SYSTEM SHALL 切换 `<html>` 的 `data-theme` 属性，并在 `localStorage["fk-theme"]` 持久化。
- [AC-26] WHEN 页面首次加载且 `localStorage["fk-theme"]` 不存在 THE SYSTEM SHALL 使用 `matchMedia('(prefers-color-scheme: dark)')` 的值。
- [AC-27] THE SYSTEM SHALL 任何状态指示同时使用颜色 + icon + 文字三重冗余（a11y 色盲友好），不依赖单一颜色。

### SSE 链路保留

- [AC-28] THE SYSTEM SHALL `shared/sse/` 实现 `SSEClient` 类：订阅 `/api/events`，处理 `ping`/`refresh` 事件名，emit `open` / `refresh` / `error` / `heartbeat` 内部事件。
- [AC-29] WHEN EventSource `onopen` 触发 THEN connection store SHALL 状态变为 `online`、attempt 归零。
- [AC-30] IF EventSource `onerror` 触发 THEN connection store SHALL 状态变为 `reconnecting`、`attempt` 加 1。
- [AC-31] IF `Date.now() - lastEventAt > 30000`（30 秒无事件）THEN SSEClient SHALL `es.close()` 强制重连，并把 connection store 状态置为 `offline`。
- [AC-32] THE SYSTEM SHALL `useSSESubscription` hook 调用 `react-query` 的 `queryClient.invalidateQueries(['summary'])` 触发自动 refetch（替代 vanilla 时代的 `loadAll()`）。

### 后端 embed 链路兼容

- [AC-33] THE SYSTEM SHALL Vite 配置 `base: '/assets/'` 输出相对路径，HTML 入口引用 `dist/assets/index-<hash>.{js,css}`。
- [AC-34] THE SYSTEM SHALL `internal/visualize/server.go` 的 `//go:embed static/*` 不需改动，Vite 产出直接落到 `static/dist/assets/`，与现有 dev workflow 兼容。
- [AC-35] THE SYSTEM SHALL 新增 `Makefile` target `dashboard-dev`（`pnpm dev`）与保留 `dashboard-dist`（`pnpm build` → `go install ./cmd/free-kiro`），不删除 dashboard-frontend-foundation 留下的 target。

### 性能预算

- [AC-36] THE SYSTEM SHALL `dist/assets/index-<hash>.js` gzip 后 ≤ 80 KB，`dist/assets/index-<hash>.css` gzip 后 ≤ 12 KB，`index.html` ≤ 4 KB。
- [AC-37] THE SYSTEM SHALL 在 `index.html` 注入 `<link rel="preload" as="style" href="/assets/index-<hash>.css">` 与 `<script type="module" src="/assets/index-<hash>.js" defer>`。
- [AC-38] WHERE `process.env.NODE_ENV === 'production'` THE SYSTEM SHALL 启用 Vite `build.rollupOptions.output.manualChunks` 把 `react`、`react-dom` 拆到独立 chunk 以利用浏览器长缓存。

### A11y 基础

- [AC-39] THE SYSTEM SHALL HeaderBar 含 `<header role="banner">`，main 含 `<main role="main">`，表格使用 `<caption>` + `<th scope="col">`。
- [AC-40] THE SYSTEM SHALL 所有可点击元素是 `<button>` 或 `<a>`，**禁止**裸 `onClick` + `<div>`。
- [AC-41] THE SYSTEM SHALL `:focus-visible { outline: 2px solid var(--color-focus); outline-offset: 2px; }`，**禁止** `outline: none`。
- [AC-42] THE SYSTEM SHALL 错误 banner 与 toast 使用 `role="alert" aria-live="polite"`。
- [AC-43] THE SYSTEM SHALL 尊重 `@media (prefers-reduced-motion: reduce)`，关闭 skeleton 渐变动画与 transition。

### 回归保护 — Unchanged Behavior

- [AC-44] THE SYSTEM SHALL CONTINUE TO 使用 stdlib Go HTTP 与 `embed.FS`，不引入任何第三方 Go 依赖。
- [AC-45] THE SYSTEM SHALL CONTINUE TO 服务 SSE 事件名 `ping`（连接就绪）与 `refresh`（文件变化），客户端订阅 `refresh` 触发 query invalidation。
- [AC-46] THE SYSTEM SHALL CONTINUE TO `/api/summary` payload 包含 `generated_at`、`specs[]`、`active`、`mode`，不删字段。
- [AC-47] THE SYSTEM SHALL CONTINUE TO 5 秒兜底轮询，即使 SSE 不可用也能看到刷新。
- [AC-48] THE SYSTEM SHALL CONTINUE TO 浏览器锁定 ≥ Chrome 100 / Safari 15 / Firefox 100。
- [AC-49] THE SYSTEM SHALL CONTINUE TO 保留 `static/legacy/index.html` 供 `/legacy` 路由回滚访问（dashboard-frontend-foundation 已建）。

### 文件迁移映射

- [AC-50] THE SYSTEM SHALL 老 `src/main.ts` → 新 `src/app/index.tsx`（React 入口）；老 `src/types.ts` 字段按 entity 拆到 `entities/spec/types.ts` + `entities/summary/types.ts`。
- [AC-51] THE SYSTEM SHALL 老 `src/lib/{api,sse,format,theme,hash-router,signal}.ts` → 新 `shared/{api,sse,lib,config}/` 下对应模块；`signal.ts` 删除（用 react-query 替代）。
- [AC-52] THE SYSTEM SHALL 老 `src/stores/{summary,detail,connection,theme}.ts` → 新 `entities/{summary,spec,connection}/` 下 query hooks 与 connection state machine；`theme` store 改用 React Context。
- [AC-53] THE SYSTEM SHALL 老 `src/components/{app,header-bar,summary-grid,specs-table,empty-state,phase-badge,skeleton-row,progress-bar,toast-stack}.ts` → 新 `widgets/` 下对应目录。
- [AC-54] THE SYSTEM SHALL 老 `src/styles/{index,tokens/*,themes/*,components}.css` → 新 `shared/styles/{globals,tokens/*,themes/*,components}.css`；Vite 通过 `import './shared/styles/globals.css'` 自动打包。

## Out of Scope

- 修复 TypeError "Cannot read properties of null" runtime bug（移至 dashboard-frontend-components）。
- spec 详情弹层 `<dialog>` + waves 进度条 + drift 详情面板（移至 dashboard-spec-detail-view）。
- fsnotify 替换 2s mtime 轮询、SSE heartbeat 强化、ETag 304（移至 dashboard-realtime-fsnotify）。
- 移动端深度适配（< 768px 全屏 sheet、卡片堆叠）（移至 dashboard-mobile-a11y）。
- 完整 a11y 自动化校验（axe-core CI 集成）（移至 dashboard-frontend-components）。
- React 19 / Server Components / Server Actions 升级探索（本 spec 锁定 React 18.x）。
- 国际化 i18next 接入（dashboard-frontend-foundation 已预留 `shared/lib/messages.ts` key→string map）。
- 删除 `static/legacy/index.html`（保留作回滚路径，由 dashboard-frontend-foundation 引入）。
- `pnpm` monorepo 多包拆分（本 spec 单包）。