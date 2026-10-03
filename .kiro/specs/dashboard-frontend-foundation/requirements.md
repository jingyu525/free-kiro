# dashboard-frontend-foundation

<!--
为 free-kiro serve dashboard 建立现代前端基建：
  - esbuild + TypeScript strict 构建链
  - 单文件 index.html 拆分为模块化结构
  - dark + light 双主题 token 系统
  - loading skeleton / 错误重试 / 空状态三态 / refresh 去抖 / 最后刷新时间 / SSE 状态指示

零后端依赖：复用现有 /api/summary + /api/events + /api/specs + /api/spec/<name>
四个端点，保留 dashboard-sse-bugfix 的 SSE 事件名（ping/refresh）链路。
本 spec 不引入 spec 详情弹层、waves 可视化、drift 详情、fsnotify、移动端深度适配 —
这些属于 dashboard-spec-detail-view / dashboard-realtime-fsnotify /
dashboard-mobile-a11y 三个独立后续 spec。
-->

## User Stories

- As a dashboard 使用者 I want 在网络慢或首次加载时看到骨架占位 so that 不会误判页面"挂了"。
- As a dashboard 使用者 I want 看到清晰的错误提示和重试按钮 so that 网络断开时知道发生了什么并能恢复。
- As a dashboard 使用者 I want 区分"还没初始化 .kiro"和"没创建任何 spec"两种空状态 so that 知道该跑哪个命令。
- As a dashboard 使用者 I want 在亮光环境使用浅色主题 so that 长时间盯屏不疲劳。
- As a dashboard 使用者 I want 知道上一次成功刷新是几秒前 so that 能判断页面是否卡死。
- As a dashboard 使用者 I want 看到 SSE 连接状态（在线/重连中/离线）so that 知道当前是实时推送还是降级轮询。
- As a dashboard 维护者 I want 前端是模块化结构 so that 后续加详情视图/移动端不用再拆 184 行单 HTML。
- As a 维护者 I want 前后端数据类型对齐（TypeScript 复用后端 JSON 字段名）so that 改字段名时编译器报错而不是运行时崩。

## Acceptance Criteria

### 构建链（esbuild + TypeScript strict）

- [AC-1] THE SYSTEM SHALL 在 `internal/visualize/static/dist/` 下产出至少一个 `main.<hash>.js`（gzip ≤ 30 KB）和一个 `main.<hash>.css`（gzip ≤ 10 KB），通过 `//go:embed static/dist` 嵌入 Go 二进制。
- [AC-2] WHEN 开发者运行 `npm run build` 或 `make dashboard-dist` THE SYSTEM SHALL 调用 esbuild 把 `src/main.ts` 与所有依赖打包到 `dist/`，并在文件名中嵌入内容哈希。
- [AC-3] WHEN 开发者运行 `npm run watch` THE SYSTEM SHALL 在 200ms 内增量重编译变更文件，并在终端输出状态。
- [AC-4] THE SYSTEM SHALL 启用 TypeScript strict（`"strict": true, "noUncheckedIndexedAccess": true`），CI 跑 `tsc --noEmit` 必须 0 error。
- [AC-5] THE SYSTEM SHALL 在 `package.json` 声明 devDependency：`esbuild` 与 `typescript`，无运行时依赖（无 react / vue / preact）。

### 模块拆分

- [AC-6] THE SYSTEM SHALL 按 `src/{main.ts,types.ts,components/,stores/,lib/,styles/}` 拆分，单文件 ≤ 500 行（POLICY §6），函数 ≤ 50 行。
- [AC-7] THE SYSTEM SHALL 提供至少 9 个组件骨架：`App`、`HeaderBar`、`SummaryGrid`、`SpecsTable`、`StatCard`、`PhaseBadge`、`EmptyState`、`ToastStack`、`SkeletonRow`，以及 4 个 store：`summary`、`detail`、`connection`、`theme`。
- [AC-8] THE SYSTEM SHALL 在 `src/lib/` 提供 `api.ts`（fetch 封装）、`sse.ts`（EventSource wrapper）、`format.ts`（时间/数字格式化）、`hash-router.ts`（URL hash 路由）、`theme.ts`（token 应用）5 个工具模块。

### Design Tokens（双主题）

- [AC-9] THE SYSTEM SHALL 在 `:root[data-theme="dark"]` 与 `:root[data-theme="light"]` 分别定义全部 CSS custom properties，至少包含 `--color-bg`、`--color-bg-elevated`、`--color-border`、`--color-fg`、`--color-fg-muted`、`--color-link`、`--color-focus`、`--color-ok`、`--color-warn`、`--color-err`、`--color-state-{draft,planning,implementing,done,drift}`。
- [AC-10] THE SYSTEM SHALL 文字与背景对比度满足 WCAG AA（≥ 4.5:1），dark 主题 `--color-fg` (`#c9d1d9`) on `--color-bg` (`#0e1116`) ≥ 11:1，light 主题对应 ≥ 14:1。
- [AC-11] THE SYSTEM SHALL 在 `src/styles/tokens/` 提供 `color.css` / `spacing.css` / `typography.css` / `radius.css` / `shadow.css` 5 个 token 文件，spacing 至少 7 档（4/8/12/16/24/32/48 px），radius 至少 3 档（3/6/12 px），text 至少 5 档（11/12/14/16/20/28 px）。
- [AC-12] WHEN 用户在 HeaderBar 点击 theme toggle 按钮 THE SYSTEM SHALL 切换 `<html>` 的 `data-theme` 属性，并在 `localStorage["fk-theme"]` 持久化。
- [AC-13] WHEN 页面首次加载且 `localStorage["fk-theme"]` 不存在 THE SYSTEM SHALL 使用 `matchMedia('(prefers-color-scheme: dark)')` 的值。
- [AC-14] THE SYSTEM SHALL 任何状态指示都同时使用颜色 + icon + 文字三重冗余（a11y 色盲友好），不依赖单一颜色。

### Loading Skeleton

- [AC-15] WHEN `summary` store 处于 `loading` 状态且 `value` 为 `null`（首次加载） THE SYSTEM SHALL 在表格区显示 `SkeletonRow × 6`，背景使用 `--color-skel-base` 与 `--color-skel-shine` 的 1.4s 线性渐变动画。
- [AC-16] WHEN `summary` store 处于 `loading` 状态且 `value` 不为 `null`（刷新中） THE SYSTEM SHALL 保留旧数据渲染，顶部显示 `<div class="top-bar-progress">` 进度条 250ms 滑动。
- [AC-17] IF `summary` store 持续 `loading` 超过 3 秒 THEN THE SYSTEM SHALL 在 HeaderBar meta 行追加 "still refreshing…" 文案。

### 错误重试

- [AC-18] THE SYSTEM SHALL 在 `api.ts` 区分三种错误类型：`TypeError`（网络断开）、`HttpError`（4xx/5xx 响应）、`ParseError`（JSON 解析失败）。
- [AC-19] WHEN `summary` store 处于 `error` 状态 THE SYSTEM SHALL 在 HeaderBar 下方显示红色 banner（`border-left: 3px solid var(--color-err)`），文案为"Couldn't refresh dashboard."并附带 `<button>retry</button>`。
- [AC-20] WHEN 用户点击 retry 按钮 THE SYSTEM SHALL 调用 `summary.refetch()`，按钮立即进入 `disabled + spinning` 状态直到请求完成。
- [AC-21] WHEN 连续 fetch 失败 THE SYSTEM SHALL 按 1s → 2s → 4s → 8s 指数退避自动重试，成功一次后归零。
- [AC-22] THE SYSTEM SHALL 每次失败在 `ToastStack` 显示一条 toast（5 秒后自动消失），文案包含错误类型名（如 `network`、`http 500`、`parse`）。

### 空状态三态

- [AC-23] THE SYSTEM SHALL 在 `/api/summary` 的 `ProjectReport` 增加字段 `mode: 'workspace-missing' | 'no-specs' | 'ok'`（后端由 dashboard-backend-api-extensions spec 提供，本 spec 容忍该字段缺失并降级为 `'ok'`）。
- [AC-24] WHEN `mode === 'workspace-missing'`（即 `.kiro/` 不存在） THE SYSTEM SHALL 显示 `EmptyState` 组件，文案为"No .kiro/ workspace found."并显示 CTA 提示运行 `free-kiro init`。
- [AC-25] WHEN `mode === 'no-specs'`（即 `.kiro/specs/` 为空） THE SYSTEM SHALL 显示 `EmptyState` 组件，文案为"No specs in this workspace."并显示 CTA 提示运行 `free-kiro spec new <name>`。
- [AC-26] WHEN `mode === 'ok'` 且 `specs.length === 0`（容错路径） THE SYSTEM SHALL 与 `no-specs` 显示一致。

### Refresh 去抖 + 最后刷新时间

- [AC-27] THE SYSTEM SHALL HeaderBar 的 refresh 按钮在点击后立即进入 `disabled + class="spinning"` 状态，500ms 内的多次点击合并为 1 次 fetch。
- [AC-28] THE SYSTEM SHALL 在 `connection` store 记录 `lastRefreshAt: number`（Date.now()），HeaderBar meta 行显示 `${count} specs · ${relativeTime(lastRefreshAt)} · generated ${HH:MM:SS}`。
- [AC-29] THE SYSTEM SHALL `formatRelative(ts)` 实现：`0-5s → "just now"`、`5-60s → "${n}s ago"`、`60s+ → "${n}m ago"`，每 1s 重新计算一次。
- [AC-30] THE SYSTEM SHALL 拦截 `Ctrl+R` 与 `Cmd+R` 浏览器默认行为，调 `e.preventDefault()` 后触发 `summary.refetch()`。

### SSE 状态指示

- [AC-31] THE SYSTEM SHALL 在 `connection` store 暴露 `state: 'online' | 'reconnecting' | 'offline'` 与 `lastEventAt: number`。
- [AC-32] THE SYSTEM SHALL HeaderBar 渲染 `<span class="conn-dot" data-state="...">` 小灯 + 文字 tooltip（hover/focus 显示 `Live · last event 3s ago` 或 `Reconnecting (attempt 2)…` 或 `Offline — using polling fallback`）。
- [AC-33] WHEN `EventSource.onopen` 触发 THE SYSTEM SHALL `state = 'online'`、`attempt = 0`。
- [AC-34] WHEN `EventSource.onerror` 触发 THE SYSTEM SHALL `state = 'reconnecting'`、`attempt++`，让浏览器默认机制处理重连。
- [AC-35] IF `Date.now() - lastEventAt > 30000`（即 30s 无 heartbeat/refresh） THEN THE SYSTEM SHALL `state = 'offline'` 并 `es.close()` 强制重连。

### A11y 基础

- [AC-36] THE SYSTEM SHALL 表格使用 `<caption>` + `<th scope="col">`，HeaderBar 含 `<header role="banner">`，main 含 `<main role="main">`。
- [AC-37] THE SYSTEM SHALL 所有可点击元素是 `<button>` 或 `<a>`，**禁止**裸 `onclick` + `<div>`。
- [AC-38] THE SYSTEM SHALL `:focus-visible { outline: 2px solid var(--color-focus); outline-offset: 2px; }`，**禁止** `outline: none`。
- [AC-39] THE SYSTEM SHALL 错误 banner 与 toast 使用 `role="alert" aria-live="polite"`。
- [AC-40] THE SYSTEM SHALL 尊重 `@media (prefers-reduced-motion: reduce)`，关闭 skeleton 动画与 transition。

### 性能预算

- [AC-41] THE SYSTEM SHALL `dist/main.<hash>.js` gzip 后 ≤ 30 KB，`dist/main.<hash>.css` gzip 后 ≤ 10 KB，`index.html` ≤ 4 KB。
- [AC-42] THE SYSTEM SHALL `<link rel="preload" href="/assets/main.<hash>.css">` 阻塞渲染最小化，`<script type="module">` 异步加载。
- [AC-43] THE SYSTEM SHALL `static/dist/` 与 `node_modules/` 加入 `.gitignore`。

### 回归保护 — Unchanged Behavior

- [AC-44] THE SYSTEM SHALL CONTINUE TO 使用 stdlib Go HTTP 与 embed.FS，不引入任何第三方 Go 依赖。
- [AC-45] THE SYSTEM SHALL CONTINUE TO 服务 SSE 事件名 `ping`（连接就绪）与 `refresh`（文件变化），客户端订阅 `refresh` 触发 `loadAll()`。
- [AC-46] THE SYSTEM SHALL CONTINUE TO `/api/summary` payload 包含 `generated_at`、`specs[]`、`active`，不删字段。
- [AC-47] THE SYSTEM SHALL CONTINUE TO 5s 兜底轮询（即使 SSE 不可用也能看到刷新）。

## Out of Scope

- spec 详情弹层（点击表格行打开 dialog）属于 dashboard-spec-detail-view 独立 spec。
- waves 可视化（stacked progress bar + 任务列表）属于 dashboard-spec-detail-view。
- drift 详情面板（点击 "N keys drifted" 弹层）属于 dashboard-spec-detail-view。
- fsnotify 替换 2s mtime 轮询、SSE heartbeat、ETag 304 属于 dashboard-realtime-fsnotify。
- 移动端深度适配（< 768px 全屏 sheet、卡片堆叠）属于 dashboard-mobile-a11y。
- 国际化（i18n）—— 本 spec 仅预留 `src/lib/messages.ts` 抽象 key→string map，不引入 i18next。
- 浅色主题色对比度的 pa11y / axe-core 自动校验 —— 本 spec 仅做手工 PR review checklist。
- `<dialog>` 元素使用 —— 当前 spec 不需要详情视图，故不动 `<dialog>` 相关样式。
