# dashboard-frontend-components — Design

<!--
决策记录（30 天后回看"为什么这么做"）：
  - normalize 在 entity 层而非 widget 层：单点修复，所有 widget 自动受益；vs widget 层到处加 ?? [] 散布逻辑
  - 原生 <dialog> 而不是 react-modal：浏览器原生焦点陷阱 + Esc + ::backdrop，免 90KB lib（与 dashboard-frontend-react-vite-fsd bundle 预算一致）
  - react-query 的 useRefetch 触发 dialog 详情：与 dashboard 一致
  - 4 tab 而非 4 page：spec 详情是 drill-down 场景，dialog 模态正确；不是新页面
  - Skip-link 用纯 anchor（#main）：不依赖 JS、不阻塞主路径
-->

## Architecture

### TypeError null.length 修复路径

```
┌────────────────────────────────────────────────────────────────────┐
│ 后端:        BuildReport (internal/visualize/report.go)            │
│              ↓ JSON 序列化                                          │
│              drift: [] 或 null                                       │
│              current: {} 或 null                                     │
│              tasks: {done,total,waves} 或 null                       │
├────────────────────────────────────────────────────────────────────┤
│ 浏览器:                                                           │
│  fetchSummary() → entities/summary/api.ts                          │
│       ↓                                                            │
│  normalizeProjectReport(raw)  ←─ 单点 null-safe 规约               │
│       ↓                                                            │
│  useSummaryQuery() 返回 ProjectReport (规范化后)                     │
│       ↓                                                            │
│  SummaryGrid / SpecsTable / SpecDetailDialog  全部安全 .map/.length│
└────────────────────────────────────────────────────────────────────┘
```

### normalize 函数

```ts
// entities/summary/normalize.ts (新文件)

export function normalizeProjectReport(raw: ProjectReportRaw): ProjectReport {
  return {
    ...raw,
    specs: (raw.specs ?? []).map((s) => ({
      meta: s.meta,
      current: s.current ?? {},
      drift: s.drift ?? [],
      tasks: s.tasks ?? { done: 0, total: 0, waves: 0 },
      active: Boolean(s.active),
    })),
    mode: raw.mode ?? 'ok',
  };
}
```

所有 widgets 通过 `useSummaryQuery()` 的 `select` option 调用 normalize：

```ts
export function useSummaryQuery(): UseQueryResult<ProjectReport, Error> {
  return useQuery({
    queryKey: SUMMARY_QUERY_KEY,
    queryFn: fetchSummary,
    select: normalizeProjectReport,  // ← 单点规约
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}
```

这样 widgets 拿到的 data 永远是规约后的，无需各自 null check。

### spec-detail-dialog 状态机

```
States:
  - 'closed': dialog 无可见
  - 'opening': hash 变化 → spec 存在 → 准备 dialog content
  - 'loading': useSpecOverviewQuery.isLoading
  - 'ready':  query 完成 → 渲染 4 tab
  - 'error':  query 失败 → RetryBanner
  - 'closing': Esc / close 按钮 / backdrop 点击 → dialog close + 焦点归还

URL 同步：
  /              → closed
  #/spec/<name>  → opening → loading → ...
                                ↓ 用户按 Esc
                                → closing → closed (hash 清空)
```

### 4 Tab 内组件

| Tab | 组件 | 数据源 |
|---|---|---|
| `overview` | inline `<dl>` meta 字段表 | spec.meta |
| `drift` | `<DriftTable drift={spec.drift}>` | spec.drift（已 normalize） |
| `tasks` | `<TaskList>` + `<WavesProgress>` | useSpecTasksQuery |
| `timeline` | inline `<ol>` 事件流 | useSpecTimelineQuery |

`drift-table` / `waves-progress` / `task-list` 是独立 widgets 子目录，dialog 内 import；FSD 单向保留。

### a11y 焦点流

```
Tab 顺序:
  skip-link → header(theme toggle / refresh / conn-dot) → main(stats → table rows → links)
                                                  → dialog 内容 (only if open)

打开 dialog 时:
  1. <dialog>.showModal()
  2. useEffect: dialog.querySelector('button, [href], input, select, textarea').focus()
  3. 焦点陷阱: Tab 在最后一个 focusable → 跳到第一个（监听 keydown）
  4. Esc: dialog.close() + history.back() (清空 hash) + trigger.focus()

关闭 dialog 后:
  hash = '' → useHashRoute 返回 '' → AppRouter 不 render Dialog → SpecsTable 重 mount
  → React focus 自然丢失（这是正常的；用户已 navig away）
```

### Skip-link CSS

```css
.skip-link {
  position: absolute;
  top: -40px;          /* 默认在视口外 */
  left: 0;
  padding: 8px 16px;
  background: var(--color-bg-elevated);
  color: var(--color-link);
  z-index: 1000;
  text-decoration: underline;
}
.skip-link:focus {
  top: 0;              /* 聚焦时滑入 */
}
```

## Components

### 新建 widgets

| Module | Responsibility | Key API |
|---|---|---|
| `widgets/spec-detail-dialog/index.tsx` | 4 tab 模态，hash 路由驱动 | `<SpecDetailDialog name={string} />` |
| `widgets/spec-detail-dialog/tabs/overview.tsx` | meta 字段 dl 表 | `<OverviewTab spec={spec} />` |
| `widgets/spec-detail-dialog/tabs/drift.tsx` | 包 DriftTable + 加载态 | `<DriftTab name={string} />` |
| `widgets/spec-detail-dialog/tabs/tasks.tsx` | 包 TaskList + WavesProgress | `<TasksTab name={string} />` |
| `widgets/spec-detail-dialog/tabs/timeline.tsx` | 事件流 ol | `<TimelineTab name={string} />` |
| `widgets/drift-table/index.tsx` | drift 表格（颜色 + 文字双重 a11y） | `<DriftTable drift={DriftSignal[]} />` |
| `widgets/waves-progress/index.tsx` | 多 wave progress bar | `<WavesProgress tasks={TaskProgress[]} />` |
| `widgets/task-list/index.tsx` | task 列表（checkbox + deps） | `<TaskList tasks={TaskProgress[]} />` |

### 新建 shared / entities

| Module | Responsibility | Key API |
|--- all lg---|---|
| `shared/lib/a11y/announce.ts` | aria-live 写入 | `announce(message, priority)` |
| `entities/summary/normalize.ts` | null-safe 规约 | `normalizeProjectReport(raw)` |
| `entities/spec/api.ts` (+ add) | `fetchSpecOverview(name)` 聚合端点 | `fetchSpecOverview(name)` |

### 修改

| Module | 改动 |
|---|---|
| `entities/summary/use-summary-query.ts` | `select: normalizeProjectReport` |
| `entities/spec/types.ts` | 标 `drift` / `current` / `tasks` 为 optional |
| `widgets/specs-table/index.tsx` | `<a href="#/spec/<name>">` 触发 dialog |
| `pages/dashboard-page/index.tsx` | 移除 ErrorBoundary（不再需要，但保留作 safety net） |
| `app/index.tsx` | render `<SpecDetailDialog />` 在 Root 下 |
| `shared/styles/components.css` | `.skip-link` / `.spec-detail-dialog` / `.drift-table` / `.waves-progress` / `.task-list` 样式 |
| `widgets/empty-state/index.tsx` | 用 `data?.specs.length === 0` 而非 `data.specs.length === 0`（normalize 后是 OK，但保险） |

## Data Model

无 Go 端 schema 变更。前端 `entities/summary/types.ts` 标字段为可选：

```ts
// 修改前（dashboard-frontend-react-vite-fsd）
export interface SpecReport {
  meta: SpecMeta;
  current: Record<string, number>;  // 强制
  drift: DriftSignal[];              // 强制
  tasks: SpecTaskProgress;           // 强制
  active: boolean;
}

// 修改后（dashboard-frontend-components）
export interface SpecReport {
  meta: SpecMeta;
  current?: Record<string, number>;  // 可选
  drift?: DriftSignal[];             // 可选
  tasks?: SpecTaskProgress;          // 可选
  active?: boolean;
}
```

`ProjectReport.specs: SpecReport[]` 仍然非空数组（空时是 `[]`），但 `ProjectReportRaw.specs` 可以是 `null`（后端序列化空切片时）。

## Error Handling

| 错误 | 来源 | UI |
|---|---|---|
| TypeError null.length（已知 bug） | 后端 drift/current/tasks 为 null | `normalizeProjectReport()` 规约；widget 不再 throw |
| dialog 详情 query 失败 | `/api/spec/<name>/{tasks,drift,timeline}` 4xx/5xx | dialog 内 RetryBanner（per-tab 独立 retry） |
| SSE 推送 + react-query invalidate | 同 dashboard-frontend-react-vite-fsd | 不变 |
| Playwright E2E 检测到 console.error | dashboard 渲染 | E2E exit 1，CI fail |

- **零吞错误**（.kiro/steering/agent-rules.md §2）：normalize 失败应 throw（开发期即可见），不允许静默 fallback。
- **错误分类前置**：`isApiError()` 已在 `shared/api/client.ts` 实现。
- **focus 归还**：dialog 关闭时 try/catch 包 trigger.focus()（防止 trigger 已被 unmount）。

## Testing Strategy

### 类型检查（CI）

```bash
pnpm exec tsc --noEmit                                    # 0 error
pnpm run lint:fsd                                         # ✓ FSD layers OK
pnpm exec vite build --mode production                    # build OK
```

### Playwright E2E（新）

```bash
# scripts/e2e-dashboard.mjs（用 MCP server 已有的 playwright 工具也可；脚本版便于 CI）
# 验证 6 场景：
#   1) goto / → 0 console.error
#   2) .summary-grid 含 8 张 .stat-card
#   3) .specs-table tbody ≥ 1 行
#   4) click spec name → <dialog open> 在 100ms 内
#   5) click theme toggle → <html data-theme> 切换
#   6) touch requirements.md → 2-3s 内 SSE refresh 触发 refetch
```

### Smoke (回归)

```bash
scripts/smoke-dashboard.sh
# dashboard-frontend-foundation 已建：
#   /api/health 200
#   /api/summary generated_at 存在
#   /api/specs .specs 数组
#   GET / Cache-Control: no-cache
#   GET /assets/index-*.js Cache-Control: immutable
```

### Bundle 预算（CI）

```bash
gzip -c dist/assets/index-<hash>.js | wc -c   # 入口 ≤ 25 KB (新增 4 widget)
gzip -c dist/assets/index-<hash>.css | wc -c  # ≤ 14 KB (新增 dialog / drift-table / waves-progress / task-list 样式)
```

## Migration / Rollout

### 落地步骤

```
1. entities/summary/normalize.ts 新建 + useSummaryQuery 加 select
2. entities/spec/api.ts 加 fetchSpecOverview(name)
3. entities/spec/types.ts 标字段可选
4. shared/lib/a11y/announce.ts 新建
5. widgets/drift-table / waves-progress / task-list 各新建
6. widgets/spec-detail-dialog 新建 + 4 tabs + 焦点管理
7. shared/styles/components.css 加 .skip-link / .spec-detail-dialog / .drift-table / .waves-progress / .task-list 样式
8. widgets/specs-table 把 <a> href 改为 #/spec/<name>（已是）
9. pages/dashboard-page / app/index.tsx 集成 SpecDetailDialog
10. 跑 tsc + lint:fsd + build
11. 启 free-kiro serve + Playwright E2E
12. 删 app/error-boundary.tsx（保留作 safety net，但测试时确保不触发）
13. commit + PR
```

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| 后端 API | 不变；normalize 处理 null slice |
| SSE 事件 | `ping`/`refresh` 不变 |
| 浏览器 | 锁定 ≥ Chrome 100 / Safari 15 / Firefox 100 |
| React 18 | 锁版本，不引入 React 19 |
| `<dialog>` 元素 | Chrome 37+ / Safari 15.4+ / Firefox 98+；与现有 baseline 一致 |

### 风险与回滚

- **风险 A**：normalize 漏字段导致某 widget 仍读 null。**对策**：每个 widget 都加 `data-testid`，Playwright 渲染后 `data-testid` 存在即说明正常。
- **风险 B**：dialog 焦点陷阱与 React 18 StrictMode 双 mount 冲突。**对策**：focus 放 useEffect 不放 render；trigger element 用 useRef 记录。
- **风险 C**：Playwright E2E CI 跑得慢（≥ 30s）。**对策**：本 spec 不强制 CI 接入；脚本本身可本地跑通即可。
- **风险 D**：a11y skip-link 与现有 toast-stack z-index 冲突。**对策**：skip-link z-index 1000 高于所有层。
- **风险 E**：normalize 把内置报告改成 optional 后，旧老 dashboard 客户端（如 free-kiro status --json）读字段报错。**对策**：本次改动仅 TypeScript 类型层，wire format 不变；旧客户端不受影响。

### 落地前置

- 三件套全绿 `free-kiro lint dashboard-frontend-components`。
- 用户显式 `free-kiro spec approve dashboard-frontend-components` 后再 `spec start`。
- 实施期改 AC 计数 → `free-kiro spec sync dashboard-frontend-components` 重 baseline。