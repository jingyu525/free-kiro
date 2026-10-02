# dashboard-frontend-components — Tasks

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
  - Wave 1：#1-#4   修 TypeError bug（normalize + types 可选化 + select）
  - Wave 2：#5-#9   补 4 widget（drift-table / waves-progress / task-list / spec-detail-dialog + 4 tabs）
  - Wave 3：#10-#13 a11y 增强（skip-link / focus management / announce / CSS）
  - Wave 4：#14-#18 集成 + 验证（dialog 集成 + E2E + smoke + 删除 fallback + commit）

提交：spec 已 implementing，commit 待落地。
-->

## Wave 1 — 修 TypeError null.length bug

- [x] #1 `entities/summary/normalize.ts` 新建：`normalizeProjectReport(raw)` 把 `specs[]/drift[]/current{}/tasks{}` 全 null-safe 规约；同时导出 `normalizeSpecReport(raw)`
- [x] #2 `entities/summary/use-summary-query.ts` 修改：`useQuery` 加 `select: normalizeProjectReport`（单点规约，所有 widget 自动拿到规范化 data）
- [x] #3 `entities/spec/types.ts` 修改：标 `drift` / `current` / `tasks` 可选（`drift?: DriftSignal[]` 等）— [deps: —]
- [x] #4 `entities/spec/api.ts` 加 `fetchSpecOverview(name)`：聚合 `/api/spec/<name>` + `/api/spec/<name>/drift` + tasks 端点；同时把后端 `{waves:[{tasks:[]}]}` 包装响应规约为 `TaskProgress[]`（flatMap），`{signals:[]}` 规约为 `DriftDetail[]`

## Wave 2 — 补 4 个 widget

- [x] #5 `widgets/drift-table/index.tsx` 新建：列 `key` / `baseline` / `current` / `delta`；delta 列正红/负绿/零灰；空数据 `"No drift signals."`；`<table><caption>...</caption><thead>...</thead><tbody>...</tbody></table>`；a11y `aria-rowindex` — [deps: #3]
- [x] #6 `widgets/waves-progress/index.tsx` 新建：每条 wave `<progress max={size} value={done}>` + 文字 `${done}/${total}`；不同 wave 不同色相（`--color-wave-1..8`）；空数据 `"No tasks."` — [deps: #3]
- [x] #7 `widgets/task-list/index.tsx` 新建：列 `#id` / `title` / `deps` / `done` / `wave`；`<ul role="list">` + 每 task `<li role="listitem">`；deps `<code>` + `aria-label="depends on #N1, #N2"`；done checkbox `aria-label` — [deps: #3]
- [x] #8 `widgets/spec-detail-dialog/index.tsx` 新建：原生 `<dialog>` + 4 tab + 焦点陷阱 + Esc 关闭 + backdrop 关闭 + 焦点归还 trigger；`useHashRoute()` 驱动显示/隐藏；`useEffect` 把 hash 变化同步到 `dialog.showModal/close`；打开时 `announce()` — [deps: #5,#6,#7]
- [x] #9 `widgets/spec-detail-dialog/tabs/{overview,drift,tasks,timeline}.tsx` 新建：4 tab 内容；overview `<dl>` meta；drift 包 DriftTable；tasks 包 TaskList + WavesProgress；timeline `<ol>` 事件流 — [deps: #8]

## Wave 3 — a11y 增强

- [x] #10 `shared/lib/a11y/announce.ts` 新建：`announce(message, priority?)` 向 `<div role="status" aria-live="polite|assertive">` 写入文字；mount 时创建该 div 到 `<body>`（singleton，priority 切换时重建） — [deps: —]
- [x] #11 `app/index.tsx` 加 `<a href="#main" class="skip-link">Skip to main content</a>` 在 HeaderBar 之前 — [deps: —]
- [x] #12 `widgets/spec-detail-dialog/index.tsx` 焦点管理：打开 → 焦点到 dialog 内首 focusable + `announce('Spec X details opened')`；关闭 → 焦点归还 trigger（`useRef` 记录） — [deps: #8]
- [x] #13 `shared/styles/components.css` 新增样式：`.skip-link`（默认 `top: -40px` + `:focus { top: 0 }`）、`.spec-detail-dialog`（`::backdrop` 灰色蒙层）、`.drift-table` / `.waves-progress` / `.task-list` / `.timeline` / `.overview-meta` 基础布局 + `@media (prefers-reduced-motion: reduce)` 关闭动画 — [deps: —]

## Wave 4 — 集成 + 验证

- [x] #14 `pages/dashboard-page/index.tsx` 把 `<main>` 加 `id="main"`（与 #11 skip-link 联动）；保留 ErrorBoundary 作 safety net（不删） — [deps: #11]
- [x] #15 `app/index.tsx` render `<SpecDetailDialog />` 在 `<AppRouter>` 兄弟位置；dialog 内部用 `useHashRoute()` 判断 `spec/<name>` — [deps: #8]
- [x] #16 `scripts/e2e-dashboard.mjs` 新建：HTTP-level Playwright 替代方案，验证 6 场景（/ 200 + cache / /api/health / /api/summary / /api/spec/<name> / /api/spec/<name>/tasks / /api/spec/<name>/drift）— [deps: #14,#15]
- [x] #17 跑 `pnpm exec tsc --noEmit` 0 error + `pnpm run lint:fsd` ✓ + `pnpm build` OK + 启 serve + E2E 6 场景全通；Playwright 浏览器手动验证 dialog 打开 + 4 tab 切换 OK；smoke `scripts/smoke-dashboard.sh` 不破（dashboard-frontend-foundation 已建） — [deps: #1-#16 全部]
- [x] #18 commit + `free-kiro spec complete dashboard-frontend-components` — [deps: #17]