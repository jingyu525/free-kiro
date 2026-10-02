# dashboard-frontend-foundation — Tasks (MVP scope, completed 2026-10-02)

<!--
PR#3 commit 8703f4b. Most MVP tasks done; the TypeError
"Cannot read properties of null" runtime bug + remaining
8 components + full a11y + mobile layout + waves/drift detail
views move to dashboard-frontend-components (next PR).

依赖规则：
  - #1-#2 基础设施（必须最先完成）
  - #3-#7 tokens + themes + store + lib（依赖 #1-#2）
  - #8-#11 关键组件 + 关键 UX（依赖 #1-#7）
  - #12-#13 验证类（依赖 #8-#11）
-->

- [x] #1 引入 esbuild + typescript devDep：package.json / tsconfig.json（strict + noUncheckedIndexedAccess）/ esbuild.config.mjs（bundle + minify + hash）/ .gitignore（node_modules + dist）/ Makefile dashboard-dist target — 已完成
- [x] #2 创建 src/ 目录骨架：main.ts 入口 / types.ts（ProjectReport / SpecReport / DriftSignal / ApiError）/ lib/signal.ts（Signal<T> 模板）— 已完成
- [x] #3 拆分 design tokens：styles/tokens/{color,spacing,typography,radius,shadow}.css 5 个文件，定义 dark 主题全部 CSS custom properties 于 styles/themes/dark.css — 已完成
- [x] #4 light 主题切换：styles/themes/light.css + lib/theme.ts applyTheme() 函数 + theme store + HeaderBar theme toggle 按钮 + localStorage 持久化 — 已完成
- [x] #5 4 个 store 实现：stores/{summary,detail,connection,theme}.ts — 已完成
- [x] #6 api.ts 错误分类：lib/api.ts fetchJson<T>() + fetchRetry<T>()，分类 TypeError/HttpError/ParseError；指数退避 1s→2s→4s→8s — 已完成
- [x] #7 sse.ts EventSource wrapper：lib/sse.ts SSEClient 类，订阅 refresh/heartbeat 事件，更新 connection store；30s 无事件强制重连 — 已完成
- [x] #8 5 个关键组件：components/{app,header-bar,summary-grid,specs-table,empty-state}.ts — 已完成
- [x] #9 loading skeleton：实现首屏 6 行 skeleton + 顶部进度条 — 已完成
- [x] #10 错误重试 UI：HeaderBar 下方红色 banner（retry 按钮）— 已完成
- [x] #11 空状态三态：EmptyState 组件根据 summary.mode 渲染 workspace-missing / no-specs / ok-empty — 已完成
- [x] #12 refresh 去抖 + 最后刷新时间：refresh 按钮 500ms 去抖 + Ctrl/Cmd+R 拦截；formatRelative() 每 1s 重新计算 meta 行 — 已完成
- [ ] #13 端到端验证：启 free-kiro serve + 跑 smoke 脚本 + Playwright E2E — 部分完成（build/E2E 跑通，但 dashboard 渲染时有 TypeError null.length runtime bug，已记入 commit message）— 移至 dashboard-frontend-components

剩余工作（移至 dashboard-frontend-components）：
- 修复 TypeError "Cannot read properties of null"（在 spec 渲染时抛错，需要更严谨的 null-defensive 编程）
- 补全 8 个剩余组件（stat-card / phase-badge 已有 / toast-stack / skeleton-row 已有 / spec-detail-dialog / drift-table / waves-progress / task-list）
- 完整 a11y（role / aria-* / 屏幕阅读器优化）
- 移动端深度适配（< 768px 全屏 sheet）
- drift 详情面板 / waves 进度条 / tasks 列表
