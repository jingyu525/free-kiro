# dashboard-frontend-foundation — Design

<!--
本 spec 是纯前端基建 + UX 反馈强化，不引入新后端 API、不引入 fsnotify、不引入新 Go 依赖。
唯一后端相关的变更是 BuildReport 的 ProjectReport 新增 mode 字段（由 dashboard-backend-api-extensions 提供，本 spec 容忍字段缺失）。
唯一架构变化是：单文件 index.html → 模块化 src/ + esbuild 构建 → embed dist。
-->

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ 开发时：                                                          │
│   src/main.ts ──esbuild--> dist/main.<hash>.js                   │
│   src/styles/* ──esbuild--> dist/main.<hash>.css                 │
│                                                                 │
│ 构建时：                                                          │
│   make dashboard-dist                                           │
│   └── npm run build (esbuild --bundle --minify)                 │
│   └── go build (//go:embed static/dist)                         │
│                                                                 │
│ 运行时：                                                          │
│   GET /             → index.html (no-cache)                     │
│   GET /assets/*     → dist/main.<hash>.{js,css} (1y immutable)  │
│   GET /api/*        → JSON (no-cache)                           │
│   GET /api/events   → SSE text/event-stream (no-cache)          │
│                                                                 │
│ 前端数据流：                                                        │
│   EventSource(/api/events) ──refresh──> summary.refetch()        │
│                                       └──> GET /api/summary    │
│                                       └──> render stats + table │
└─────────────────────────────────────────────────────────────────┘
```

- 不变：Go 端 stdlib net/http + embed.FS；SSE 事件名 `ping`/`refresh`（dashboard-sse-bugfix 固化）。
- 唯一 Go 端变化：`//go:embed static/dist`（在 `internal/visualize/server.go` 现有 `//go:embed static/*` 旁加一行）。
- 唯一新依赖：`package.json` 加 devDependency `esbuild` + `typescript`（开发时用，运行时无）。
- 路由：`/` 不变；新增 `/assets/*` 走 `handleStatic`（由 dashboard-backend-middleware spec 提供，本 spec 假定已就绪）。
- 旧 `/index.html` 单文件保留在 `static/legacy/index.html`，供 `/legacy` 路由回滚访问（dashboard-sse-bugfix 当前实现不变，只是物理路径改了）。

## Components

| Component | Responsibility | 依赖 |
|---|---|---|
| `App` (`src/components/app.ts`) | 顶层 layout：mount HeaderBar + main(SummaryGrid + SpecsTable) + ToastStack；订阅 `summary`、`connection`、`theme` 三个 store | summary, connection, theme |
| `HeaderBar` (`header-bar.ts`) | 标题 + meta（specs count + relativeTime + HH:MM:SS）+ refresh 按钮 + theme toggle + conn-dot | summary, connection, theme |
| `SummaryGrid` (`summary-grid.ts`) | 渲染 7 张 stat 卡片 | summary |
| `SpecsTable` (`specs-table.ts`) | 渲染 spec 表格（name / phase / approved / drift / tasks），name 列为 `<a href="#spec/<name>">`（点击本 spec 不弹层，hash 路由预留供后续 spec 用） | summary |
| `StatCard` (`stat-card.ts`) | 单张 stat 卡片（label + value + optional delta） | props |
| `PhaseBadge` (`phase-badge.ts`) | phase chip（颜色 + 文字） | props |
| `EmptyState` (`empty-state.ts`) | 三态分桶：workspace-missing / no-specs / ok-empty | summary.mode |
| `ToastStack` (`toast-stack.ts`) | 错误/通知堆叠，5s 自动消失 | toast store (内部) |
| `SkeletonRow` (`skeleton-row.ts`) | 表格 loading 行 × 6，渐变动画 | summary.loadState |

### Stores（手写 Signal 模型）

```ts
// src/lib/signal.ts (≈ 60 行)
class Signal<T> {
  private _v: T;
  private subs = new Set<(v: T) => void>();
  constructor(v: T) { this._v = v; }
  get(): T { return this._v; }
  set(next: T) { this._v = next; this.subs.forEach(fn => fn(next)); }
  subscribe(fn: (v: T) => void): () => void {
    this.subs.add(fn);
    return () => { this.subs.delete(fn); };
  }
}
```

| Store | 类型 | 信号 |
|---|---|---|
| `summary` | `Signal<{value: ProjectReport \| null; loadState: 'idle'\|'loading'\|'error'\|'success'; lastError: ApiError \| null}>` | 同时驱动 stats + table + lastRefreshAt |
| `detail` | `Map<string, Signal<SpecDetail>>`（LRU 50，本 spec 不消费，预留给 spec-detail-view） | 预留 |
| `connection` | `Signal<{state: 'online'\|'reconnecting'\|'offline'; attempt: number; lastEventAt: number}>` | 驱动 conn-dot + sse.ts 状态机 |
| `theme` | `Signal<'dark'\|'light'\|'auto'>` | 驱动 `<html data-theme>` + toggle 按钮 |

### Lib（5 个工具模块）

| Module | Responsibility |
|---|---|
| `api.ts` | `fetchJson<T>(url)` 封装：分类 `TypeError`/`HttpError`/`ParseError`；`fetchRetry(url, opts)` 指数退避 |
| `sse.ts` | `SSEClient` 类：EventSource wrapper + heartbeat 监控 + 重连状态机；emit `'open'\|'refresh'\|'error'\|'heartbeat'` 事件 |
| `format.ts` | `formatRelative(ts)`、`formatDelta(n)`、`formatTimeHHMMSS(date)`、`formatNumber(n)` |
| `theme.ts` | `applyTheme(value)`：写 `<html data-theme>` + localStorage + matchMedia 监听 |
| `hash-router.ts` | `subscribeHash(fn)`：监听 `hashchange`，emit `''` 或 `'spec/<name>'`；本 spec 不消费，仅占位 |

### 数据契约（`src/types.ts`）

```ts
// 复用后端 internal/visualize/report.go 的字段命名
export type ProjectReport = {
  generated_at: string;          // RFC3339
  specs: SpecReport[];
  active: string;                // = .kiro/.current
  mode?: 'workspace-missing' | 'no-specs' | 'ok';  // 后端 dashboard-backend-api-extensions 提供；本 spec 缺省 'ok'
};

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
  tasks: {done: number; total: number; waves: number};  // waves 字段本 spec 不渲染
  active: boolean;
};
```

## Error Handling

| 错误类型 | 来源 | UI |
|---|---|---|
| `TypeError`（`fetch` 失败） | 网络断开 / DNS / CORS | 顶部红色 banner "Network error" + retry + toast |
| `HttpError(4xx)` | 端点 404 / 401 | 顶部红色 banner "HTTP 404" + retry + toast |
| `HttpError(5xx)` | 服务端 panic | 顶部红色 banner "Server error 5xx" + retry + toast |
| `ParseError` | JSON.parse 失败 | 顶部红色 banner "Invalid response" + retry + toast |
| SSE `onerror` | EventSource 断开 | conn-dot → `reconnecting` / `offline`（不弹 toast，避免噪音） |
| SSE 30s 无事件 | 服务端 hang | conn-dot → `offline`，强制 `es.close()` 重连 |

- **零吞错误**（.kiro/steering/agent-rules.md §2）：所有 `try/catch` 必须显式处理或重新抛出；不允许 `_ = doX()`。
- **错误分类前置**：在 `api.ts` 一处分类，UI 层不再判断 `instanceof Error`。
- **退避**：`fetchRetry` 用 `attempt * attempt * 1000ms`，cap 8000ms；成功一次归零；UI 按钮 disable 直到下一次 success。

## Testing Strategy

### 类型检查（CI）
```bash
tsc --noEmit                                          # 0 error
esbuild --bundle src/main.ts --outfile=/tmp/check.js  # 仅验证可编译
```

### Bundle 体积（CI）
```bash
gzip -c dist/main.<hash>.js | wc -c   # ≤ 30 KB
gzip -c dist/main.<hash>.css | wc -c  # ≤ 10 KB
```

### E2E（沿用 `.playwright-mcp/`，不入 CI）
```bash
free-kiro serve --port 7373 &
sleep 1
mcp__playwright__browser_navigate("http://127.0.0.1:7373")
mcp__playwright__browser_snapshot()        # 验证 a11y tree 含 dialog-free, refresh, theme toggle, conn-dot
mcp__playwright__browser_click(ref=refresh_button)
mcp__playwright__browser_click(ref=theme_toggle)   # 验证 <html data-theme> 切换
mcp__playwright__browser_press_key("Control+R")   # 验证自定义 refresh
# 验证 SSE
curl -N http://127.0.0.1:7373/api/events &
sleep 5
touch .kiro/specs/<any>/requirements.md
sleep 1
# 应收到 event: refresh
kill %1
kill %2
```

### Smoke 脚本（CI 跑）
```bash
# scripts/smoke-dashboard.sh
curl -sf http://127.0.0.1:7373/api/health        | jq -e '.status == "ok"'
curl -sf http://127.0.0.1:7373/api/summary       | jq -e '.generated_at'
curl -sf http://127.0.0.1:7373/api/specs          | jq -e '.specs'
curl -sfI http://127.0.0.1:7373/                  | grep -i 'cache-control:.*no-cache'
curl -sfI http://127.0.0.1:7373/assets/main.HASH.js | grep -i 'cache-control:.*immutable'
```

## Migration / Rollout

### 单文件 → 多文件的拆分步骤（保留 dashboard-sse-bugfix 行为不变）

```
1. 复制 static/index.html → static/legacy/index.html（保留原样）
2. 创建 static/src/ 目录骨架 + package.json + esbuild.config.mjs + tsconfig.json
3. 新建 static/index.html → 仅 <div id="app"></div> + <link>/<script> 引用 dist/main.<hash>.{css,js}
4. server.go 加一行 //go:embed static/dist（与现有 static/* 并列）
5. /legacy 路由 → 返回 static/legacy/index.html（让用户可手动回滚）
6. npm install (devDep: esbuild + typescript)
7. npm run build → 产出 dist/
8. go build ./cmd/free-kiro → 嵌入 dist
9. 跑 smoke + E2E 验证 dashboard-sse-bugfix 行为不破
```

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| `free-kiro serve` CLI | flag 不变 |
| `/api/summary` payload | 向后兼容：新增 `mode` 字段，旧客户端忽略 |
| SSE 事件名 | `ping`/`refresh` 不变；客户端忽略未识别事件 |
| `<dialog>` 元素 | 本 spec 不使用；`/legacy` 路径仍可用 |
| 浏览器 | 锁定 ≥ Chrome 100 / Safari 15 / Firefox 100；用 `<dialog>` 在 dashboard-spec-detail-view 才引入 |
| 旧 dashboard 用户 | `/legacy` 路由可直接访问原单文件 184 行版本 |

### 风险与回滚

- **风险 A**：esbuild 失败导致 dist 为空 → `//go:embed` panic → `go build` 报错。**对策**：`Makefile dashboard-dist` target 必须在 `go build` 之前；加 `make check-dashboard` 跑一次 `test -f dist/main.*.js`。
- **风险 B**：前端 ESM 语法在老浏览器跑不起来 → 首屏白屏。**对策**：浏览器锁定 ≥ Chrome 100；`/legacy` 路由兜底。
- **风险 C**：bundle 超 30KB → 首屏慢。**对策**：CI 跑 `gzip -c | wc -c` 断言 ≤ 30 KB，超出则 fail。
- **风险 D**：SSE 事件名不匹配（dashboard-sse-bugfix 链路断）→ 自动刷新失效。**对策**：保留 dashboard-sse-bugfix 的 `ping`/`refresh` 不变；E2E 用 `touch .kiro/specs/<name>/requirements.md` 验证 2s 内收 `refresh`。

### 落地前置（按 .kiro/steering/agent-rules.md §1）

- 本 spec 三件套（requirements / design / tasks）必须 `free-kiro lint` 全绿。
- 落地后由用户显式 `free-kiro spec approve dashboard-frontend-foundation` 后再 `spec start`。
- 实施期改了 AC 计数 → 跑 `free-kiro spec sync dashboard-frontend-foundation` 重 baseline，不重生成。
