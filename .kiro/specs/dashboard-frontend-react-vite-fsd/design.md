# dashboard-frontend-react-vite-fsd — Design

<!--
决策记录（30 天后回看问"为什么这么做"）：
  - Vite 5 而不是 Next.js：dashboard 是单页应用、SSR 不必要、Go embed 产物只需要静态资源
  - React 18 而不是 19：react-query 生态 18 兼容性验证更广、concurrent features 已够用
  - pnpm 9 而不是 npm：硬链接复用磁盘、monorepo-ready、CI 缓存友好
  - FSD 而不是 DDD/Atomic：FSD 天然匹配"自顶向下 import 方向"硬约束、避免循环依赖
  - @tanstack/react-query 替代手写 Signal：re-fetch / 缓存 / 失效 / 退避全包
  - CSS Modules + design tokens（不用 Tailwind）：与 dashboard-frontend-foundation 的 token 系统一脉相承
-->

## Architecture

### 整体数据流

```
┌────────────────────────────────────────────────────────────────────┐
│ 浏览器                                                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │ React 18 root   │  │ react-query     │  │ EventSource     │    │
│  │ (app/index.tsx) │←→│ QueryClient     │←→│ /api/events     │    │
│  │                 │  │ + cache         │  │ ping/refresh    │    │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘    │
│           │                    │                     │            │
│           ↓ render             ↓ fetch               ↓ SSE push   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ widgets/  features/  entities/   pages/  app/                │ │
│  └─────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────┘
                                ↕ HTTP
┌────────────────────────────────────────────────────────────────────┐
│ free-kiro serve (Go, stdlib net/http + embed.FS)                  │
│  ├── GET /                 → index.html (no-cache)               │
│  ├── GET /assets/*         → dist/assets/index-<hash>.{js,css}    │
│  ├── GET /api/summary      → ProjectReport JSON                  │
│  ├── GET /api/specs        → SpecReport[] JSON                   │
│  ├── GET /api/spec/<n>/... → 详情/drift/timeline/hooks/health     │
│  └── GET /api/events       → SSE text/event-stream (ping/refresh)│
└────────────────────────────────────────────────────────────────────┘
```

### FSD 分层（强制 import 方向）

```
src/
├── app/                      ← 应用初始化、Providers、入口
│   ├── providers/            ← QueryClientProvider / ThemeProvider / ToastProvider
│   ├── router/               ← hash 路由
│   └── index.tsx             ← ReactDOM.createRoot
├── pages/                    ← 页面级组合
│   └── dashboard-page/       ← 单页 dashboard（hash 路由 #spec/<name> 预留）
├── widgets/                  ← 独立 UI block
│   ├── header-bar/
│   ├── summary-grid/         ← 7 张 stat 卡 grid
│   ├── specs-table/
│   ├── empty-state/          ← workspace-missing / no-specs / ok 三态
│   ├── toast-stack/
│   ├── progress-bar/
│   └── skeleton-row/
├── features/                 ← 用户场景功能（可独立测试、可独立删除）
│   ├── theme-toggle/         ← HeaderBar 内的明暗切换
│   ├── refresh-dashboard/    ← Ctrl+R 拦截 + refresh 按钮去抖
│   ├── retry-fetch/          ← 错误 banner 的 retry 按钮
│   └── skeleton-overlay/     ← 首屏 6 行骨架
├── entities/                 ← 业务实体（types + fetcher + hooks）
│   ├── summary/              ← ProjectReport 类型 + useSummaryQuery
│   ├── spec/                 ← SpecReport 类型 + useSpecQuery / useSpecTasksQuery
│   └── connection/           ← SSE connection state machine
└── shared/                   ← 可复用基础设施（无业务语义）
    ├── api/                  ← fetchJson / fetchRetry + 错误分类
    ├── sse/                  ← SSEClient + useSSESubscription
    ├── config/               ← API base URL / asset base URL
    ├── lib/                  ← format / hash-router
    ├── ui/                   ← 无业务语义的可复用组件（Button / Card / Spinner / Skeleton）
    └── styles/               ← tokens / themes / globals / components
```

**Import 方向硬约束**（由 `eslint-plugin-boundaries` 或等价 import 顺序 lint 强制）：

```
app → pages → widgets → features → entities → shared
                  ↑           ↑           ↑
                  └─── 不允许反向 ──────────┘
```

跨层违规示例（lint 阻断）：
- `entities/summary` import `widgets/header-bar` ❌
- `shared/ui` import `features/theme-toggle` ❌
- `widgets/summary-grid` import `pages/dashboard-page` ❌

同层之间允许 import（如 `entities/spec` import `entities/summary/types` 的子集类型），但鼓励"将类型 → 而非 instance →"。

### React 组件树

```
<QueryClientProvider client={qc}>
  <ThemeProvider>
    <ToastProvider>
      <HashRouter>
        <Routes>
          <Route path="/" element={<DashboardPage/>} />
          {/* #spec/<name> 预留位：dashboard-spec-detail-view 实现 */}
        </Routes>
      </HashRouter>
    </ToastProvider>
  </ThemeProvider>
</QueryClientProvider>

DashboardPage
├── <HeaderBar/>                    ← widget
│   ├── <ThemeToggle/>            ← feature（嵌入 header-bar 右上角）
│   ├── <ConnectionDot/>          ← entity/connection 展示
│   └── <RefreshButton/>          ← feature（点击去抖 500ms）
├── <SummaryGrid/>                 ← widget
│   └── <StatCard/> × 7           ← widget
├── <SpecsTable/>                  ← widget
│   ├── <SpecRow/> × N             ← shared/ui + entity/spec 类型
│   └── <PhaseBadge/>             ← widget
├── <EmptyState mode=.../>         ← widget（三态）
├── <RetryBanner error={...}/>    ← feature（错误状态）
└── <ToastStack/>                 ← widget（错误 toast 堆叠）
```

### 状态管理（react-query + Context）

| 状态 | 工具 | 位置 |
|---|---|---|
| `summary` server state | `useQuery({ queryKey: ['summary'], queryFn: fetchSummary })` | `entities/summary/` |
| `specs` server state | `useQuery({ queryKey: ['specs'], queryFn: fetchSpecs })` | `entities/spec/` |
| `connection` SSE 状态 | `useReducer` + Context（`'online' \| 'reconnecting' \| 'offline'`） | `entities/connection/` |
| `theme` UI 偏好 | React Context + `localStorage` 持久化 | `app/providers/ThemeProvider` |
| `toast` 全局通知 | `useReducer` + Context（push / dismiss） | `app/providers/ToastProvider` |
| `hash` 路由 | `useSyncExternalStore` 订阅 `hashchange` | `app/router/` |

### 数据契约（`entities/*/types.ts`）

```ts
// entities/summary/types.ts
export type ProjectReport = {
  generated_at: string;          // RFC3339
  specs: SpecReport[];
  active: string;                // .kiro/.current
  mode?: 'workspace-missing' | 'no-specs' | 'ok';
  workspace_ref?: string;        // 来自 dashboard-backend-api-extensions
  started_at?: string;           // 服务端启动时间
  last_refresh_at?: string;     // 最近一次 SSE refresh 时间
};

// entities/spec/types.ts
export type SpecReport = {
  meta: {
    name: string;
    phase: 'draft' | 'requirements' | 'design' | 'tasks' | 'approved' | 'implementing' | 'done';
    workflow: string;
    spec_type: 'feature' | 'bugfix';
    quick: boolean;
    approved: boolean;
    generator: string;
    prompt: string;
    created_at: string;
    updated_at: string;
  };
  current: Record<string, number>;
  drift: Array<{key: string; baseline: number; current: number; delta: number}>;
  tasks: {done: number; total: number; waves: number};
  active: boolean;
};
```

字段命名与 `internal/visualize/report.go` 完全一致；后续新增字段（如 `task_list` 端点）只需在 `entities/spec/types.ts` 加 type，调用方编译器报错而不是 runtime 崩。

### Vite 配置（`vite.config.ts`）

```ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'node:path';

export default defineConfig(({ mode: viteMode }) => ({
  root: resolve(__dirname, 'src'),
  base: '/assets/',
  plugins: [react()],
  server: { port: 5173, strictPort: false },
  build: {
    outDir: resolve(__dirname, 'dist'),
    emptyOutDir: true,
    target: 'es2022',
    minify: 'esbuild',
    sourcemap: viteMode !== 'production' ? 'inline' : false,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('node_modules/react') || id.includes('node_modules/react-dom')) {
            return 'react-vendor';
          }
          if (id.includes('node_modules/@tanstack/react-query')) {
            return 'query-vendor';
          }
          return undefined;
        },
      },
    },
  },
}));
```

`root: 'src'` 把 Vite 入口设为 `src/index.html`（Vite 约定），构建后 HTML 与 assets 都在 `dist/`。

### CSS 架构（保留 dashboard-frontend-foundation 的 token 系统）

- `shared/styles/globals.css` — 入口 import 所有 token + theme + components
- `shared/styles/tokens/{color,spacing,typography,radius,shadow}.css` — 5 个 token 文件
- `shared/styles/themes/{dark,light}.css` — 双主题 CSS custom properties
- `shared/styles/components.css` — Button / Card / Skeleton 等 UI kit 样式

Vite 通过 `import 'shared/styles/globals.css'` 在 `app/index.tsx` 入口加载，运行时打包到 `index-<hash>.css`。

### 与 dashboard-frontend-foundation 的关系

| 老规范（已 done） | 本规范覆盖 | 不覆盖（移至后续） |
|---|---|---|
| esbuild 构建链 | ✅ 替换为 Vite | — |
| vanilla TS | ✅ 替换为 React 18 | — |
| npm | ✅ 替换为 pnpm 9 | — |
| 手写 Signal store | ✅ 替换为 react-query | — |
| 9 个组件 | ✅ 重写为 React widgets | — |
| 双主题 token | ✅ 保留 | — |
| SSE `ping`/`refresh` | ✅ 保留 | — |
| TypeError null bug | ❌ 不修 | dashboard-frontend-components |
| spec 详情 dialog | ❌ 不做 | dashboard-spec-detail-view |
| drift 详情面板 | ❌ 不做 | dashboard-spec-detail-view |
| waves 进度条 | ❌ 不做 | dashboard-spec-detail-view |
| 移动端深度适配 | ❌ 不做 | dashboard-mobile-a11y |
| 完整 a11y 校验 | ❌ 不做 | dashboard-frontend-components |
| `/legacy` 路径 | ✅ 保留 | — |

## Components

### shared 层

| Module | Responsibility | Key API |
|---|---|---|
| `shared/api/client.ts` | fetch 封装 + 错误分类 + 退避 | `fetchJson<T>(url)`, `fetchRetry<T>(url, opts)`, `ApiError` |
| `shared/sse/client.ts` | EventSource wrapper + heartbeat + state machine | `SSEClient` class |
| `shared/sse/use-sse-subscription.ts` | React hook：SSE 推 → react-query invalidate | `useSSESubscription(handler)` |
| `shared/config/index.ts` | API / asset base URL | `API_BASE`, `ASSET_BASE` |
| `shared/lib/format.ts` | 时间/数字格式化 | `formatRelative`, `formatDelta`, `formatTimeHHMMSS`, `formatNumber` |
| `shared/lib/hash-router.ts` | `hashchange` 监听 + `useHashRoute` hook | `useHashRoute()` |
| `shared/ui/Button.tsx` | 可复用 Button（primary/ghost/danger） | `<Button variant=...>` |
| `shared/ui/Card.tsx` | 可复用 Card 容器 | `<Card>` |
| `shared/ui/Spinner.tsx` | 加载 spinner | `<Spinner size=...>` |
| `shared/ui/Skeleton.tsx` | 通用骨架占位 | `<Skeleton variant=...>` |

### entities 层

| Module | Responsibility | Key API |
|---|---|---|
| `entities/summary/types.ts` | ProjectReport / DetailShapes 等 | `type ProjectReport` |
| `entities/summary/api.ts` | `fetchSummary()` 调用 `/api/summary` | `fetchSummary(): Promise<ProjectReport>` |
| `entities/summary/use-summary-query.ts` | react-query wrapper | `useSummaryQuery()` |
| `entities/spec/types.ts` | SpecReport / DriftSignal / TaskProgress | `type SpecReport` |
| `entities/spec/api.ts` | fetchSpec / fetchSpecTasks / fetchSpecDrift | `fetchSpec(name)`, `fetchSpecTasks(name)` |
| `entities/spec/use-spec-query.ts` | react-query wrapper | `useSpecQuery(name)` |
| `entities/connection/state.ts` | SSE 状态机 reducer | `connectionReducer(state, action)` |
| `entities/connection/provider.tsx` | React Context 暴露 connection 状态 | `<ConnectionProvider>` |

### features 层

| Module | Responsibility | Key API |
|---|---|---|
| `features/theme-toggle/index.tsx` | HeaderBar 内的明暗切换按钮 | `<ThemeToggle/>` |
| `features/refresh-dashboard/index.tsx` | refresh 按钮 + Ctrl+R 拦截 | `<RefreshButton/>` |
| `features/retry-fetch/index.tsx` | 错误 banner retry 按钮 | `<RetryBanner error=... onRetry=.../>` |
| `features/skeleton-overlay/index.tsx` | 首屏 6 行骨架 | `<SkeletonOverlay rows=6/>` |

### widgets 层

| Module | Responsibility | Key API |
|---|---|---|
| `widgets/header-bar/index.tsx` | 标题 + meta + refresh + theme + conn-dot | `<HeaderBar/>` |
| `widgets/summary-grid/index.tsx` | 7 张 stat 卡 grid | `<SummaryGrid/>` |
| `widgets/specs-table/index.tsx` | spec 表格（name/phase/approved/drift/tasks） | `<SpecsTable/>` |
| `widgets/empty-state/index.tsx` | workspace-missing / no-specs / ok 三态 | `<EmptyState mode=.../>` |
| `widgets/toast-stack/index.tsx` | 错误/通知堆叠（5s 自动消失） | `<ToastStack/>` |
| `widgets/progress-bar/index.tsx` | 顶部进度条（刷新中 250ms 滑动） | `<ProgressBar active=.../>` |
| `widgets/skeleton-row/index.tsx` | 表格 loading 行 × N | `<SkeletonRow/>` |
| `widgets/stat-card/index.tsx` | 单张 stat 卡（label + value + delta） | `<StatCard label=... value=...>` |
| `widgets/phase-badge/index.tsx` | phase chip（颜色 + 文字） | `<PhaseBadge phase=.../>` |

### pages / app

| Page / App | Responsibility | Key API |
|---|---|---|
| `pages/dashboard-page/index.tsx` | 单页 dashboard 组合 | `<DashboardPage/>` |
| `app/providers/query-provider.tsx` | QueryClient 配置（staleTime 30s / retry 3） | `<QueryProvider>` |
| `app/providers/theme-provider.tsx` | Theme Context + localStorage 持久化 | `<ThemeProvider>` |
| `app/providers/toast-provider.tsx` | Toast Context + push/dismiss reducer | `<ToastProvider>` |
| `app/providers/connection-provider.tsx` | SSE connection Context 包装 | `<ConnectionProvider>` |
| `app/router/index.tsx` | hash 路由 + `<Routes>` 配置 | `<AppRouter/>` |
| `app/index.tsx` | ReactDOM.createRoot 挂载点 | `createRoot(document.getElementById('app')!)` |

## Data Model

无 Go 端 schema 变更。本 spec 不引入新端点、不调 `internal/visualize/report.go`。

新引入的 TypeScript 类型在 `entities/{summary,spec,connection}/types.ts`，字段命名复用 `internal/visualize/report.go` 的 JSON tag。

## Error Handling

| 错误类型 | 来源 | UI 表现 |
|---|---|---|
| `ApiError: TypeError`（网络断开） | `fetch` reject | 红色 banner + toast（5s 自动消失）+ connection state `offline` |
| `ApiError: HttpError(4xx)` | 端点 404 / 401 | 红色 banner + toast |
| `ApiError: HttpError(5xx)` | 服务端 panic | 红色 banner + toast |
| `ApiError: ParseError` | JSON.parse 失败 | 红色 banner + toast |
| SSE `onerror` | EventSource 断开 | conn-dot → `reconnecting` / `offline`（不弹 toast） |
| SSE 30s 无事件 | 服务端 hang | conn-dot → `offline`，SSEClient `es.close()` 强制重连 |
| React Query `error` 状态 | react-query 触发 | `<RetryBanner error=... onRetry={() => query.refetch()} />` |

- **零吞错误**（.kiro/steering/agent-rules.md §2）：所有 `try/catch` 必须显式处理或重新抛出；不允许 `_ = doX()`。
- **错误分类前置**：在 `shared/api/client.ts` 一处分类为 `ApiError`（union type），UI 层不再判断 `instanceof Error`。
- **退避**：`fetchRetry` 用 `attempt * attempt * 1000ms`，cap 8000ms；成功一次归零；UI 按钮 disable 直到下一次 success。

## Testing Strategy

### 类型检查（CI）

```bash
pnpm exec tsc --noEmit                                # 0 error
pnpm exec vite build --mode production --logLevel info # 验证可构建
```

### Bundle 体积（CI）

```bash
gzip -c dist/assets/index-<hash>.js | wc -c   # ≤ 80 KB（含 React 18 runtime）
gzip -c dist/assets/index-<hash>.css | wc -c  # ≤ 12 KB
```

### FSD 分层检查（CI）

```bash
# 用 eslint-plugin-boundaries 或自写 import 顺序脚本
pnpm exec eslint --ext .ts,.tsx src/                # 跨层违规阻断
# 或自写 script/scripts/check-fsd-layers.mjs：
#   解析 import 路径 → 检查 (from_layer, to_layer) 关系
#   违规 → exit 1
```

### E2E（沿用 `.playwright-mcp/`，不入 CI）

```bash
free-kiro serve --port 7374 &
sleep 1
mcp__playwright__browser_navigate("http://127.0.0.1:7374")
mcp__playwright__browser_snapshot()                # 验证 a11y tree 含 refresh / theme toggle / conn-dot
mcp__playwright__browser_click(ref=refresh_button)
mcp__playwright__browser_click(ref=theme_toggle)   # 验证 <html data-theme> 切换
mcp__playwright__browser_press_key("Control+R")    # 验证 Ctrl+R → summary refetch
# SSE 验证
curl -N http://127.0.0.1:7374/api/events &
sleep 5
touch .kiro/specs/<any>/requirements.md
sleep 1
# 应收到 event: refresh（react-query 自动 invalidate + refetch）
kill %1
```

### Smoke（CI 跑）

```bash
scripts/smoke-dashboard.sh
# 与 dashboard-frontend-foundation 相同：
curl -sf http://127.0.0.1:7374/api/health    | jq -e '.status == "ok"'
curl -sf http://127.0.0.1:7374/api/summary   | jq -e '.generated_at'
curl -sf http://127.0.0.1:7374/api/specs      | jq -e '.specs'
curl -sfI http://127.0.0.1:7374/              | grep -i 'cache-control:.*no-cache'
curl -sfI http://127.0.0.1:7374/assets/index.HASH.js | grep -i 'cache-control:.*immutable'
```

## Migration / Rollout

### 迁移步骤（保留 legacy 回滚路径）

```
1. pnpm init → 生成 package.json + pnpm-lock.yaml
   ├── 声明 packageManager + engines
   ├── runtime deps: react@18 react-dom@18
   └── devDeps: vite@5 @vitejs/plugin-react typescript@5 @types/react @types/react-dom
2. tsconfig.json 新增 "jsx": "react-jsx" + "noUncheckedIndexedAccess": true
3. vite.config.ts 创建（root: 'src', base: '/assets/'）
4. 把 src/legacy/index.html 保留；新建 src/index.html（Vite 入口）指向 /src/app/index.tsx
5. 按 FSD 分层建目录：app/ pages/ widgets/ features/ entities/ shared/
6. 按 entity → shared → feature → widget → page → app 顺序迁移代码
7. pnpm install
8. pnpm build → dist/assets/index-<hash>.{js,css}
9. go install ./cmd/free-kiro → embed static/dist
10. 跑 smoke + E2E 验证 dashboard-sse-bugfix / dashboard-backend-api-extensions 行为不破
```

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| `free-kiro serve` CLI | flag 不变 |
| `/api/*` payload | 向后兼容：新增字段旧客户端忽略 |
| SSE 事件名 | `ping`/`refresh` 不变；react-query invalidateQueries via 用途不变 |
| `<dialog>` 元素 | 本 spec 不使用；`/legacy` 路径仍可用 |
| 浏览器 | 锁定 ≥ Chrome 100 / Safari 15 / Firefox 100；与 dashboard-frontend-foundation 一致 |
| 旧 dashboard 用户 | `/legacy` 路由可直接访问原单文件版本 |

### 风险与回滚

- **风险 A**：pnpm 9 在某些 CI runner 上需要 corepack 启用 → CI fail。**对策**：CI 跑 `corepack enable pnpm` 或 `npm i -g pnpm@9`；写进 README。
- **风险 B**：Vite build 失败导致 dist 为空 → `//go:embed` panic → `go build` 报错。**对策**：`Makefile dashboard-dist` target 必须在 `go build` 之前；保留 `make check-dashboard` 跑 `test -f dist/assets/index-*.js`。
- **风险 C**：bundle 超 80KB → 首屏慢。**对策**：CI 跑 `gzip -c | wc -c` 断言 ≤ 80 KB，超出则 fail；用 `manualChunks` 拆 react/react-dom 到独立 chunk。
- **风险 D**：SSE 链路断（react-query invalidate 不触发）。**对策**：保留 dashboard-sse-bugfix 的 `ping`/`refresh` 不变；E2E 用 `touch .kiro/specs/<name>/requirements.md` 验证 2s 内收 `refresh`。
- **风险 E**：FSD 分层违规 import（反向依赖）导致循环依赖。**对策**：`pnpm lint` 跑 `eslint-plugin-boundaries` 或自写 `scripts/check-fsd-layers.mjs`，违规阻断。
- **风险 F**：React 18 升级到 19 的 minor 风险（hydration / concurrent 行为变化）。**对策**：锁版本到 `react@18.3.x` patch；不引入 React 19 特性。

### 落地前置（按 .kiro/steering/agent-rules.md §1）

- 本 spec 三件套（requirements / design / tasks）必须 `free-kiro lint dashboard-frontend-react-vite-fsd` 全绿。
- 落地后由用户显式 `free-kiro spec approve dashboard-frontend-react-vite-fsd` 后再 `spec start`。
- 实施期改了 AC 计数 → 跑 `free-kiro spec sync dashboard-frontend-react-vite-fsd` 重 baseline，不重生成。
- 与 dashboard-frontend-components / dashboard-spec-detail-view / dashboard-mobile-a11y 三个后续 spec 解耦：本 spec 不修 TypeError bug、不加 dialog、不做移动端。