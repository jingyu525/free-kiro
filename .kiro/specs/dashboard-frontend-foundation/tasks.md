# dashboard-frontend-foundation — Tasks (MVP scope)

<!--
本 spec 缩小为 MVP 范围（用户决定 PR#3 拆为两个子 spec）。
剩余工作（13 个组件全套 + a11y + 移动端 + 完整验证）转入 dashboard-frontend-components。

依赖规则：
  - #1-#2 基础设施（必须最先完成）
  - #3-#7 tokens + themes + store + lib（依赖 #1-#2）
  - #8-#11 关键组件 + 关键 UX（依赖 #1-#7）
  - #12-#13 验证类（依赖 #8-#11）
-->

- [ ] #1 引入 esbuild + typescript devDep：写 `package.json` / `tsconfig.json`（strict + noUncheckedIndexedAccess）/ `esbuild.config.mjs`（bundle + minify + hash）/ `.gitignore`（node_modules + dist）/ `Makefile` `dashboard-dist` target 于 `internal/visualize/static/`
- [ ] #2 创建 `src/` 目录骨架：写 `main.ts` 入口（仅 `mount(<App>)`）/ `types.ts`（ProjectReport / SpecReport / DriftSignal / ApiError）/ `lib/signal.ts`（Signal<T> 模板，~60 行）
- [ ] #3 拆分 design tokens：写 `styles/tokens/{color,spacing,typography,radius,shadow}.css` 5 个文件，定义 dark 主题全部 CSS custom properties 于 `styles/themes/dark.css`
- [ ] #4 light 主题切换：写 `styles/themes/light.css` + `lib/theme.ts` `applyTheme()` 函数 + `theme` store + HeaderBar theme toggle 按钮 + localStorage 持久化
- [ ] #5 4 个 store 实现：`stores/{summary,detail,connection,theme}.ts`，`summary` 暴露 `refetch()` + `loadState` + `lastRefreshAt`；`detail` 是 LRU 50（占位）；`connection` 暴露 `state` + `attempt` + `lastEventAt`
- [ ] #6 api.ts 错误分类：写 `lib/api.ts` `fetchJson<T>()` + `fetchRetry<T>()`，分类 `TypeError`/`HttpError`/`ParseError`；指数退避 1s→2s→4s→8s
- [ ] #7 sse.ts EventSource wrapper：写 `lib/sse.ts` `SSEClient` 类，订阅 `refresh`/`heartbeat` 事件，更新 `connection` store；30s 无事件强制重连
- [ ] #8 5 个关键组件：`components/{app,header-bar,summary-grid,specs-table,empty-state}.ts`，每个组件一个文件，函数 ≤ 50 行；剩余 8 个组件（stat-card / phase-badge / toast-stack / skeleton-row / spec-detail-dialog / drift-table / waves-progress / task-list）转 dashboard-frontend-components
- [ ] #9 loading skeleton：实现首屏 6 行 skeleton + 顶部进度条；3s 后追加 "still refreshing…" 文案
- [ ] #10 错误重试 UI：HeaderBar 下方红色 banner（retry 按钮）+ ToastStack 5s 自动消失；retry 按钮 disable + spinning
- [ ] #11 空状态三态：`EmptyState` 组件根据 `summary.mode` 渲染 workspace-missing / no-specs / ok-empty；本 spec 容忍 `mode` 字段缺失降级为 `'ok'`
- [ ] #12 refresh 去抖 + 最后刷新时间：refresh 按钮 500ms 去抖 + Ctrl/Cmd+R 拦截；`formatRelative()` 每 1s 重新计算 meta 行
- [ ] #13 端到端验证：启 `free-kiro serve` + 跑 smoke 脚本 + Playwright E2E（snapshot / click refresh / click theme toggle）+ 验证 dashboard-sse-bugfix SSE 事件名 ping/refresh 链路未破 [deps: #1,#2,#3,#4,#5,#6,#7,#8,#9,#10,#11,#12]

