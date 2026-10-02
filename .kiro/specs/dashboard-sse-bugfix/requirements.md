# dashboard-sse-bugfix

<!--
修复 free-kiro dashboard 的一组已知缺陷：
  1. SSE 事件名错配导致自动刷新链路失效
  2. spec 链接是死的（href="#" 无 handler）
  3. loadSpecs() 错端点 + 重复拉取 /api/summary
  4. footer 文案与实际能力不符
  5. Shutdown() 中 time.After 未释放

本次只动 internal/visualize/{server.go,server_sse.go,static/index.html} 与
internal/cli/serve.go 中的 footer 文案。spec 详情视图（#2 弹层）属于新功能，
不在本次范围内。
-->

## User Stories

- As a free-kiro dashboard 使用者 I want 改动 .kiro/ 下文件后秒级看到表格刷新 so that 编辑 spec 时不用每 5 秒等 polling 兜底。
- As a free-kiro dashboard 使用者 I want spec 名称作为可点击链接 so that 表格不只是装饰，未来能承载详情视图入口。
- As a free-kiro dashboard 使用者 I want 每 5 秒刷新只产生一次网络请求 so that 浏览器开发者工具 Network 面板不被冗余请求淹没。
- As a free-kiro 维护者 I want Shutdown 路径下不再泄漏未释放的 timer so that 长时间运行的进程不会积累微小 GC 压力。

## Acceptance Criteria

### 修复 #1 — SSE 自动刷新链路

- WHEN `.kiro/specs/<name>/*.md` 文件被修改且任意浏览器 SSE 客户端已连接 THE SYSTEM SHALL 在 `watchTickInterval` (2s) 内向所有已订阅客户端发送一条名为 `refresh` 的 SSE 事件，事件 data 字段为 `1`。
- THE SYSTEM SHALL 在 SSE 连接建立时立即向客户端发送一条名为 `ping` 的 SSE 事件，事件 data 字段为 `ok`，以便 EventSource 确认流已就绪。
- WHEN 浏览器 dashboard 收到任意名为 `refresh` 的 SSE 事件 THE SYSTEM SHALL 调用 `loadAll()` 触发 summary 与 specs 表格刷新。
- UNLESS 浏览器 EventSource 不可用 THE SYSTEM SHALL CONTINUE TO 每 5 秒轮询 `/api/summary` 作为刷新兜底。

### 修复 #3 — loadAll 请求去重

- WHEN dashboard 调用 `loadAll()` THE SYSTEM SHALL 向 `/api/summary` 发起恰好一次请求，并同时驱动 stat 卡与表格的渲染。
- THE SYSTEM SHALL 在 5 秒轮询周期内对 `/api/summary` 发起恰好一次请求，零冗余。
- THE SYSTEM SHALL 不再调用 `/api/specs` 端点（保留作为未来 N+1 详情视图的预留接口，不在本次使用）。

### 修复 #4 — footer 文案

- THE SYSTEM SHALL 在 dashboard 页脚展示 "Polling every 5s · SSE auto-refresh on file change. Powered by free-kiro + stdlib net/http."，准确描述实际刷新机制。

### 修复 #5 — Shutdown timer 释放

- WHEN `Server.Shutdown()` 在 `s.watcherRunning` 已关闭后返回 THE SYSTEM SHALL 释放为 `time.After` 分配的 timer 对象，不再依赖其在 `shutdownTimeout` 后自然过期。

### 回归预防 — Unchanged Behavior

- WHEN 任意浏览器 SSE 客户端断开连接 THE SYSTEM SHALL CONTINUE TO 从 `subscribers` 列表中移除对应 channel，且不会 panic（unsubscribe 保持幂等且不 close channel）。
- THE SYSTEM SHALL CONTINUE TO 通过 `sync.Once` 保证 `watchChanges()` 在 Server 生命周期内仅启动一次，无论有多少 SSE 客户端先后连接。
- WHEN 任意 `.kiro/` 文件被外部命令修改 THE SYSTEM SHALL CONTINUE TO 在 2 秒 tick 内检测到 mtime 变化并触发 broadcast。
- WHEN `Server.Shutdown()` 被调用且 `watchChanges` 从未启动 THE SYSTEM SHALL CONTINUE TO 在 `shutdownTimeout` (5s) 内返回，零死锁。
- THE SYSTEM SHALL CONTINUE TO 使用 stdlib `net/http`，不引入任何第三方依赖。

## Out of Scope

- spec 详情视图（点击表格行弹出详情面板）属于新功能，需要单独的 spec 与 design 阶段，本次只保留现有死链接的语义不变（不实现新弹层）。
- SSE keep-alive heartbeat（防止长连接被中间设备切断）属于性能增强，本次不引入。
- `fsnotify` 替换 2 秒 mtime 轮询 watcher 属于未来的 sub-second 实时性增强，本次不动。
- 国际化（footer / stat label 的中文版）本次不动。
