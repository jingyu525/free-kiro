# dashboard-spec-detail-view

<!--
目标：在 dashboard-frontend-components 已落地（dialog + 4 tab + a11y + E2E）
基础上增强 spec 详情视图：
  - deep-link 同步 hash 当前选中的 tab（#/spec/<name>?tab=drift）
  - 跨 spec 链接（task deps 链可跳转）
  - drift trend sparkline（从 timeline 端点取 baseline/current 历史）
  - URL hash 持久化（浏览器刷新后恢复 tab 状态）
  - "back to list" 关闭按钮 + 跳到下一个/上一个 spec 的快速导航

零后端 API 变更：复用 /api/spec/<name>/{tasks,drift,timeline} 3 端点；
保留 dashboard-frontend-react-vite-fsd 的 FSD 6 层 + React 18 + react-query
5 + Vite 5 + ErrorBoundary。

移动端深度适配属 dashboard-mobile-a11y；fsnotify / ETag 属
dashboard-realtime-fsnotify。
-->

## User Stories

- As a dashboard 使用者 I want 浏览器 URL 反映当前 spec 详情 + tab so that 复制 URL 粘贴给同事可直接跳到该 spec 的 drift 子视图。
- As a dashboard 使用者 I want 刷新浏览器后 spec 详情 + tab 状态保留 so that 不会因刷新丢失上下文。
- As a dashboard 使用者 I want 点击 task 的 deps 链接直接跳到依赖 spec 详情 so that 不用回列表再选。
- As a dashboard 使用者 I want drift 列显示历史趋势 sparkline so that 能直观看到 baseline/current/delta 是最近变多还是稳定。
- As a dashboard 使用者 I want dialog 关闭按钮旁边有 prev/next spec 快速导航 so that 不用回列表切换。

## Acceptance Criteria

### Deep-link 同步 tab

- [AC-1] THE SYSTEM SHALL dialog 打开时解析 `window.location.hash`，匹配 `#/spec/<name>?tab=<overview|drift|tasks|timeline>` 时自动切到对应 tab（默认 overview）。
- [AC-2] WHEN 用户点击 dialog tab 切换 THE SYSTEM SHALL 100ms 内更新 `window.location.hash` 为 `#/spec/<name>?tab=<newTab>`，不触发 react-query 重新挂载 dialog。
- [AC-3] THE SYSTEM SHALL 浏览器后退按钮（history.back()）按 hash 时间顺序回退到上一个 dialog tab 状态或关闭 dialog（如果上一个 hash 不含 `spec/`）。
- [AC-4] THE SYSTEM SHALL 解析非法 tab 值（如 `?tab=foo`）时降级为 `overview` tab，不抛错。

### URL hash 持久化 + 刷新恢复

- [AC-5] WHEN 浏览器刷新（Cmd+R / F5）THE SYSTEM SHALL 从 `window.location.hash` 恢复 spec name + tab，dialog 自动 `showModal()`，焦点到对应 tab 的 panel。
- [AC-6] THE SYSTEM SHALL hash 不带 tab query 时默认 `overview` tab（如 `#/spec/dashboard-frontend-components`）。
- [AC-7] THE SYSTEM SHALL `useHashRoute()` hook 扩展支持 `?tab=` query string，返回 `{name, tab}` 而不是裸 string。

### 跨 spec 链接（task deps 跳转）

- [AC-8] WHEN task-list 中 task 有 deps THE SYSTEM SHALL 把每个 dep 渲染为 `<a href="#/spec/<dep>?tab=tasks">#<dep></a>` 而不是裸 `<code>` 文本。
- [AC-9] WHEN 用户点击 dep 链接 THE SYSTEM SHALL 100ms 内跳到该 dep spec 的 dialog，且 tab 自动切到 `tasks`，焦点归还到对应 task（按 name 锚点 `#task-<id>`）。
- [AC-10] WHEN dep spec 的 dialog 关闭（按 Esc / 点 Close / 点 backdrop）THE SYSTEM SHALL hash 清空 `?tab=` query，但保留 `#/spec/<current-name>`（回到原 spec），焦点归还当前 task。

### Drift trend sparkline

- [AC-11] WHEN drift-table 中某 row 的 `key` 在 timeline 端点返回的事件历史里存在 THE SYSTEM SHALL 渲染 `<svg class="sparkline" viewBox="0 0 60 16">` 含 N 个历史点（≤ 60 个）+ 当前点用 `<circle r="2" fill="currentColor">` 高亮。
- [AC-12] WHEN timeline 不含该 key 的历史 THE SYSTEM SHALL 渲染空 `<span class="sparkline sparkline-empty">—</span>` 占位。
- [AC-13] THE SYSTEM SHALL sparkline path stroke 颜色跟 delta trend 对齐：`up` 红 / `down` 绿 / `flat` 灰（颜色 + 文字 `+N`/`-N`/`±0` 双重）。
- [AC-14] THE SYSTEM SHALL sparkline `aria-label="history of <key>: <N> samples, latest delta <delta>"`，`<svg role="img">`。

### Spec 详情快速导航

- [AC-15] THE SYSTEM SHALL dialog 标题旁渲染 `<button aria-label="Previous spec">←</button>` 与 `<button aria-label="Next spec">→</button>`，点击时按 dashboard specs 列表的当前顺序（按 `specs[i]`）跳到上/下一个 spec。
- [AC-16] WHEN 当前 spec 是列表第一个 THE SYSTEM SHALL "Previous" 按钮 `disabled` 且 `aria-disabled="true"`；同理列表最后一个时 "Next" 禁用。
- [AC-17] THE SYSTEM SHALL dialog 标题旁另渲染 `<button aria-label="Copy deep link">🔗</button>`，点击时调 `navigator.clipboard.writeText(location.href)` 并 `announce('Link copied')`。

### a11y 增量

- [AC-18] THE SYSTEM SHALL dialog `<h2>` 含 `id="dialog-title-<name>"`；tab `<button>` 含 `aria-controls="panel-<tab>"` 指向 panel id。
- [AC-19] WHEN 用户用键盘 ArrowLeft / ArrowRight 在 tab 之间切换 THE SYSTEM SHALL 焦点移到上一个 / 下一个 tab button 并自动选中（roving tabindex 模式，参考 WAI-ARIA tab pattern）。
- [AC-20] THE SYSTEM SHALL dialog 内的"close / prev / next / copy"按钮组用 `<div role="toolbar" aria-label="Spec detail actions">` 包裹。
- [AC-21] THE SYSTEM SHALL 尊重 `@media (prefers-reduced-motion: reduce)`：tab 切换无过渡动画、sparkline 无 draw 动画、dialog backdrop fade 0.01ms。

### 文件迁移映射

- [AC-22] THE SYSTEM SHALL `shared/lib/hash-router.ts` 扩展 `useHashRoute()` 返回 `{name: string, tab: 'overview'|'drift'|'tasks'|'timeline'}`，保留原 `HashRoute` type 给 legacy 调用方。
- [AC-23] THE SYSTEM SHALL 新增 `entities/spec/sparkline.ts`：从 `/api/spec/<name>/timeline` 返回的事件流提取 `<key>` → number[] 历史数组；接受 `fetchSpecTimeline(name)` + `<key>` 返回 `{samples: number[], latest: number}`。
- [AC-24] THE SYSTEM SHALL 修改 `widgets/spec-detail-dialog/index.tsx`：标题栏加 prev/next/copy 3 个 button；关闭时调用 `navigateToHash('')`（清空 `?tab=`）。
- [AC-25] THE SYSTEM SHALL 修改 `widgets/task-list/index.tsx`：deps 渲染为 `<a>` 链接而非 `<code>` 文本。
- [AC-26] THE SYSTEM SHALL 修改 `widgets/drift-table/index.tsx`：每行末加 `<td>` 含 `<SparklineCell key={...} name={...} />`。
- [AC-27] THE SYSTEM SHALL 新增 `widgets/sparkline-cell/index.tsx`：调 `useSparklineData(name, key)` 渲染 svg。

## Out of Scope

- spec 详情内的写操作（编辑 phase、approve、start 实施）—— dashboard 当前 read-only viewer。
- spec 详情内的 diff visualization（baseline vs current 对比 diff）—— 后端需新端点，超出本 spec scope。
- Sparkline 数据从 git log 推导（仅时间序列 timestamp + key，不存事件本身）—— 后端 timeline 端点当前返回 kind+message，无 key 维度；本 spec 容许 timeline 不含 key → 显示占位（AC-12）。
- 跨 spec 跳转动画（fade-out / fade-in 过渡）—— 键盘焦点直接是关键路径。
- 移动端 dialog 在小屏幕自适应布局（mobile deep-link 全屏 sheet）—— dashboard-mobile-a11y。
- react-router / history API 替代 url hash —— 与 dashboard-frontend-foundation hash-router 保持一致。