# dashboard-spec-detail-view — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

Wave 划分：
  - Wave 1：#1-#3   hash-router 扩展 + sparkline 数据 hook + sparkline widget
  - Wave 2：#4-#7   widget 修改（spec-detail-dialog + task-list + drift-table + dialog-toolbar）
  - Wave 3：#8-#11  a11y 增强（roving tabindex + toolbar role + CSS + announce）
  - Wave 4：#12-#15 E2E 升级 + 验证 + commit + complete
-->

## Wave 1 — 路由 + 数据层

- [ ] #1 `shared/lib/hash-router.ts` 修改：`useHashRoute()` 返回 `{name: string, tab: 'overview'|'drift'|'tasks'|'timeline'}`；保留 `HashRoute` type（向后兼容）；非法 tab 降级为 `'overview'` — [deps: —]
- [ ] #2 `entities/spec/sparkline.ts` 新建：`useSparklineData(name, key)` hook 调 `fetchSpecTimeline(name)` 用 `"k="<n>"|"<key>=<n>"` regex 提取历史 samples，返回 `{samples: number[], latest: number, delta: number}`；timeline 缺 key 维度时 samples=[]、latest=0、delta=0 — [deps: —]
- [ ] #3 `widgets/sparkline-cell/index.tsx` 新建：`<SparklineCell name={string} key={string} />` 渲染 `<svg class="sparkline" viewBox="0 0 60 16" role="img" aria-label="history of <key>: <N> samples, latest delta <delta>">` + path + current circle；samples < 2 时仅画一个 circle；空 samples 时 `<span class="sparkline-empty">—</span>` — [deps: #2]

## Wave 2 — widget 修改

- [ ] #4 `widgets/spec-detail-dialog/components/dialog-toolbar.tsx` 新建：`<DialogToolbar>` 包裹 prev / next / copy / close 4 个 `<button>`，外层 `<div role="toolbar" aria-label="Spec detail actions">` — [deps: —]
- [ ] #5 `widgets/spec-detail-dialog/index.tsx` 修改：(a) tab 切换时 `navigateToHash(`#/spec/${name}?tab=${newTab}`)`；(b) `useHashRoute()` 替代 destructured string 提取 `name` + `tab`；(c) prev/next 在 specs 列表上下跳当前+1/-1，列表首尾 disabled；(d) copy button 调 `navigator.clipboard.writeText(location.href)` + `announce('Link copied')` + 失败时 `announce('Copy failed', 'assertive')`；(e) `<h2>` 加 `id="dialog-title-${name}"`，tab buttons 加 `aria-controls="panel-${tab}"` — [deps: #1,#4]
- [ ] #6 `widgets/spec-detail-dialog/index.tsx` 修改：tab buttons 加 `tabIndex={tab === currentTab ? 0 : -1}` (roving tabindex) + `onKeyDown` 监听 ArrowLeft / ArrowRight / Home / End 切换 — [deps: #5]
- [ ] #7 `widgets/task-list/index.tsx` 修改：deps 渲染为 `<a href="#/spec/${dep}?tab=tasks#task-${id}">` 取代 `<code>` 文本；每 task `<li>` 加 `id="task-${id}"`；点击 dep → 跨 spec 跳转 — [deps: —]
- [ ] #8 `widgets/drift-table/index.tsx` 修改：每行末加 `<td><SparklineCell name={specName} key={row.k} /></td>` — [deps: #3]

## Wave 3 — a11y 增强

- [ ] #9 `shared/styles/components.css` 加：`.sparkline` (width:60px, height:16px, stroke:currentColor) + `.sparkline path`（曲线） + `.sparkline circle`（当前点 r=2） + `.sparkline-empty`（灰色 —） + `.dialog-toolbar` (display: flex, gap: var(--space-2)) — [deps: —]
- [ ] #10 `widgets/spec-detail-dialog/index.tsx` 修改：dialog 关闭时 `announce('Spec details closed', 'polite')`；打开时 `announce(`Spec ${name} details, tab ${tab}`, 'polite')` — [deps: #5]
- [ ] #11 `widgets/spec-detail-dialog/index.tsx` 修改：跨 spec 链接触发后 `requestAnimationFrame(() => { dialogRef.current?.querySelector(`#task-${id}`)?.scrollIntoView({ block: 'nearest' }) })` — [deps: #7]

## Wave 4 — E2E + 验证

- [ ] #12 `scripts/e2e-dashboard.mjs` 加场景：7) `/#/spec/<name>?tab=drift` URL 渲染（HTTP 200 + body 含 `tab=drift`）；8) `/api/spec/<name>/timeline` 返回非空数组 — [deps: #5]
- [ ] #13 `pnpm exec tsc --noEmit` 0 error + `pnpm run lint:fsd` ✓ + `pnpm build` OK；Playwright 浏览器验证：a) URL `?tab=drift` 切到 drift tab；b) ArrowRight 切到 timeline；d) copy button clipboard 写入；e) 跨 spec 跳转 dialog content 重 mount + scrollIntoView — [deps: #1-#12 全部]
- [ ] #14 commit — [deps: #13]
- [ ] #15 `free-kiro spec complete dashboard-spec-detail-view` — [deps: #14]