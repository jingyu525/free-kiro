# dashboard-frontend-components

<!--
目标：修 free-kiro serve dashboard React 渲染期的 TypeError "Cannot read
properties of null" runtime bug（dashboard-frontend-foundation MVP 已知 +
dashboard-frontend-react-vite-fsd 迁移后仍存在，ErrorBoundary 兜底但首屏
不完整）；补 4 个剩余 widget（spec-detail-dialog / drift-table / waves-progress /
task-list）+ 完整 a11y（role / aria-* / 屏幕阅读器 / focus management /
keyboard navigation）+ Playwright E2E 端到端验证。

零后端 API 变更：复用 dashboard-frontend-react-vite-fsd 的 FSD 6 层 +
React 18 + react-query 5 + ErrorBoundary + Vite 5 + pnpm 9；复用
dashboard-backend-api-extensions 的 /api/spec/<name>/{tasks,drift,timeline} 端点。

移动端深度适配（< 768px 全屏 sheet、卡片堆叠）属于 dashboard-mobile-a11y 独立 spec。
fsnotify / ETag 属于 dashboard-realtime-fsnotify。
-->

## User Stories

- As a dashboard 使用者 I want 点击 spec name 弹出详情面板 so that 能查看该 spec 的 phase / drift / tasks / timeline 详情，不用离开 dashboard 主页。
- As a dashboard 使用者 I want React 渲染稳定无 TypeError so that 首次加载就能看到 stats grid + specs table，ErrorBoundary fallback 不再触发。
- As a 屏幕阅读器使用者 I want 所有交互元素带 role / aria-* 标签 so that 能用 NVDA / VoiceOff 完整浏览 dashboard。
- As a 键盘使用者 I want Tab / Shift+Tab / Enter / Esc 完整操作 dashboard so that 不用鼠标也能导航、打开详情、关闭弹层、刷新。
- As a 维护者 I want Playwright E2E 覆盖 dashboard 渲染 + 交互 so that 后续 spec 修改不会回归到 TypeError 崩树。

## Acceptance Criteria

### 修 TypeError null.length bug（dashboard 渲染稳定）

- [AC-1] THE SYSTEM SHALL dashboard 首次加载 `<= 3 秒` 内完成 stats grid + specs table 渲染，控制台 0 个 error。
- [AC-2] WHEN `useSummaryQuery().data` 为 `undefined` 且 `isLoading` 为 `true` THE SYSTEM SHALL 显示 `<SkeletonOverlay rows=6>` 占位，不调用任何 `spec.summary` 的 `length` 或 `map` 属性。
- [AC-3] WHEN `useSummaryQuery().data.specs` 为 `null` 或 `undefined` THE SYSTEM SHALL 把它规范化为 `[]` 后再调 `.map(...)` 或 `.length`，所有 widget 不得直接读取 `data.specs.length` 不带 `??[]`。
- [AC-4] WHEN `spec.drift` 为 `null`（来自后端 `BuildReport` 空 drift 序列化为 null slice） THE SYSTEM SHALL 规范化为 `[]` 后渲染。
- [AC-5] WHEN `spec.current` 为 `null`（同上） THE SYSTEM SHALL 规范化为 `{}` 后渲染。
- [AC-6] WHEN `spec.tasks` 为 `null` 或字段缺失 (`done`/`total`/`waves`) THE SYSTEM SHALL 降级为 `{done: 0, total: 0, waves: 0}` 后渲染。
- [AC-7] THE SYSTEM SHALL 在 `entities/summary/types.ts` 把 `drift: DriftSignal[]`、`current: Record<string, number>`、`tasks: SpecTaskProgress` 标为**可选**并提供 `normalizeSpecReport()` 工具函数做 null-safe 降级，调用方在 widgets 内统一调 `normalize()` 后再渲染。
- [AC-8] THE SYSTEM SHALL `pnpm exec tsc --noEmit` 0 error；`pnpm run lint:fsd` 通过；`pnpm build` 通过；浏览器 Playwright `goto /` 后 `consoleErrors === []`。

### 补 spec-detail-dialog（点击 name 列打开）

- [AC-9] WHEN 用户在 SpecsTable 点击 name 链接 `<a href="#/spec/<name>">` THE SYSTEM SHALL 在 100ms 内打开 `<dialog>` 模态，显示该 spec 的完整详情（meta / drift / tasks / timeline 4 tab）。
- [AC-10] THE SYSTEM SHALL `<dialog>` 使用原生 `<dialog>` 元素（`showModal()` API），自动获取焦点陷阱、Esc 关闭、`::backdrop` 灰色蒙层。
- [AC-11] WHEN 用户按 Esc 或点击 backdrop THE SYSTEM SHALL SHALL 关闭 dialog，并把焦点归还到触发链接（`useRef` 记录 trigger 元素）。
- [AC-12] THE SYSTEM SHALL dialog 标题含 `<h2>{spec.meta.name}</h2>` + phase badge + close 按钮（`aria-label="Close"`）。
- [AC-13] THE SYSTEM SHALL dialog 内 4 个 tab：`overview`（meta 字段）/ `drift`（drift-table）/ `tasks`（task-list）/ `timeline`；tab 切换用 `<button role="tab" aria-selected>` + `<section role="tabpanel">`。
- [AC-14] WHEN dialog 打开且 spec 详情 query 正在加载 THE SYSTEM SHALL 显示 6 行 skeleton（per-tab）。
- [AC-15] WHEN dialog 详情 query 失败 THE SYSTEM SHALL 显示 retry 按钮（与 dashboard-page 同 RetryBanner）。

### 补 drift-table

- [AC-16] THE SYSTEM SHALL `widgets/drift-table/index.tsx` 渲染 spec 的 drift 详情，列：`key` / `baseline` / `current` / `delta`；delta 列正数显示 `+N` 红色背景、负数显示 `-N` 绿色背景、零显示 `±0` 灰色。
- [AC-17] WHEN drift query `data` 为 `null` 或 `[]` THE SYSTEM SHALL 显示 `No drift signals.` 占位。
- [AC-18] THE SYSTEM SHALL drift-table 外层 `<table><caption>{spec.name} drift</caption><thead><tr><th scope="col">…</th></tr></thead><tbody>{rows.map}</tbody></table>`。

### 补 waves-progress

- [AC-19] THE SYSTEM SHALL `widgets/waves-progress/index.tsx` 渲染 spec 的 wave 进度：每条 wave 一个 `<progress>` 元素（`max={waveSize}`、`value={waveDone}`） + 文字 `${done}/${total}`。
- [AC-20] WHEN waves 数量为 0（无 task list） THE SYSTEM SHALL 显示 `No tasks.` 占位。
- [AC-21] THE SYSTEM SHALL 不同 wave 用不同色相（`--color-wave-1` 到 `--color-wave-N`，N ≤ 8）以 a11y 列.

### 补 task-list

- [AC-22] THE SYSTEM SHALL `widgets/task-list/index.tsx` 渲染 spec 的 task 列表，列：`#id` / `title` / `deps` / `done` / `wave`。
- [AC-23] WHEN task 已完成 THE SYSTEM SHALL 在 `<input type="checkbox" checked aria-label>` 显示 ✓；WHEN 未完成 THE SYSTEM SHALL 显示空 checkbox。
- [AC-24] THE SYSTEM SHALL task-list 用 `<ul role="list">` + 每 task `<li role="listitem">`；deps 列显示为 `<code>` 标签链 + `aria-label="depends on #N1, #N2"`。

### 完整 a11y

- [AC-25] THE SYSTEM SHALL 在 HeaderBar 之前渲染 `<a href="#main" class="skip-link">Skip to main content</a>`，键盘聚焦显示，鼠标点击隐藏。
- [AC-26] WHEN `<dialog>` 打开 THE SYSTEM SHALL 自动把焦点移到 dialog 内第一个 focusable 元素；WHEN 关闭 THE SYSTEM SHALL 把焦点归还到触发元素。
- [AC-27] THE SYSTEM SHALL 所有 `<button>` 含可读文字或 `aria-label`；所有 `<a>` 含可读文字或 `aria-label`；**禁止**裸 `onClick` + `<div>` / `<span>`。
- [AC-28] THE SYSTEM SHALL `widgets/specs-table` 的 `<tr>` 加 `aria-rowindex`；sortable 表头加 `<button aria-sort="ascending|descending|none">`。
- [AC-29] THE SYSTEM SHALL 状态指示同时使用 颜色 + icon + 文字 三重（`conn-dot` online/reconnecting/offline + 🌐/🔄/❌ + 文字 tooltip）。
- [AC-30] THE SYSTEM SHALL 尊重 `@media (prefers-reduced-motion: reduce)`：关闭 skeleton 1.4s 渐变动画、关闭 progress-bar 250ms 滑动、关闭 dialog `showModal` 过渡动画。

### Playwright E2E 验证

- [AC-31] THE SYSTEM SHALL 新建 `scripts/e2e-dashboard.mjs`（或 dashboard.spec.ts），用 Playwright 验证 6 个核心场景：1) dashboard 首次加载 0 console error；2) stats grid 渲染 8 张卡（含 total + 7 phases）；3) specs table 渲染 ≥ 1 行；4) 点击 spec name 列在 100ms 内打开 `<dialog>`；5) 主题切换按钮在 dark ↔ light 切换 `<html data-theme>` 属性；7) SSE 在 `touch .kiro/specs/<any>/requirements.md` 后 2-3s 内收到 `refresh` 事件（react-query 自动 invalidate）。
- [AC-32] THE SYSTEM SHALL E2E 跑通后 0 失败；CI 在 `.github/workflows/ci.yml` 加 `e2e-dashboard` job（不在本 spec 强制落地，但脚本本身必须能本地跑通）。

### 文件迁移映射

- [AC-33] THE SYSTEM SHALL 新建 `widgets/spec-detail-dialog/{index.tsx,tabs/overview,tabs/drift,tabs/tasks,tabs/timeline}.tsx`（tabs 是 widget 内部组件，仍属于 widgets 层）。
- [AC-34] THE SYSTEM SHALL 新建 `widgets/drift-table/index.tsx`、`widgets/waves-progress/index.tsx`、`widgets/task-list/index.tsx`。
- [AC-35] THE SYSTEM SHALL `entities/spec/api.ts` 暴露 `fetchSpecOverview(name)` 返回 `{meta, drift, current}` 复合类型（聚合现有 `fetchSpec` + `fetchSpecDrift`）。
- [AC-36] THE SYSTEM SHALL `shared/lib/a11y/announce.ts` 提供 `announce(message: string, priority?: 'polite'|'assertive')` 工具，向 `<div role="status" aria-live>` 写入文字（屏幕阅读器朗读）。

## Out of Scope

- 移动端深度适配（< 768px 全屏 sheet、卡片堆叠、touch gesture）属于 dashboard-mobile-a11y 独立 spec。
- fsnotify 替换 2s mtime 轮询、ETag 304、SSE heartbeat 强化属于 dashboard-realtime-fsnotify。
- `<dialog>` 内编辑 spec phase、生成 spec、approve spec 等写操作（dashboard 当前是 read-only viewer，写操作走 `free-kiro spec` CLI）。
- i18n 完整接入（dashboard-frontend-foundation 已预留 `shared/lib/messages.ts` key→string map）。
- axe-core 自动 a11y CI 校验（保留 E2E 手动 aria-* 断言；axe-core 接入是后续 PR）。
- React 19 / Server Components 升级探索（本 spec 锁 React 18）。
- WebSocket 替代 upgrade-stats（保留 SSE `ping`/`refresh` 链路）。