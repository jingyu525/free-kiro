# dashboard-sse-bugfix — Design

<!--
本次是 bugfix + 小幅 UI 调整，全部走最小改动路径：
  - SSE 事件名修复：单端对齐
  - loadSpecs 端点切换 + 去重：JS 端 2 行
  - footer 文案：1 行
  - Shutdown timer 释放：3 行 Go 重写

无架构变更、无新依赖、无新接口。
-->

## Architecture

- 不变：dashboard 仍是 `internal/visualize.Server` 上的 stdlib HTTP 服务，前端是单文件 `static/index.html`。
- 唯一的事件链路：`.kiro/` 文件 → `watchChanges()` 每 2s tick 检测 mtime → `broadcast()` → 已订阅 channel → SSE handler 写出 `event: refresh\ndata: 1\n\n` → 浏览器 `EventSource.addEventListener('refresh', ...)` → `loadAll()`。
- 唯一的数据链路：`loadAll()` 并行打 `/api/summary` + `/api/specs`，分别用于 stat 卡渲染与表格渲染。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `visualize.Server.handleEvents` (Go) | SSE 写出 + 触发 watcher | 事件名固定为 `ping` / `refresh` |
| `visualize.Server.watchChanges` (Go) | 2s tick mtime diff + broadcast | 无 API 变更 |
| `visualize.Server.Shutdown` (Go) | drain watcher + http.Server.Shutdown | 用 `time.NewTimer` 替代 `time.After` |
| `static/index.html` (JS) | EventSource 订阅 + fetch + DOM 渲染 | `loadAll()` 拆为两个端点 |

## Data Model

- 不变。无 schema / struct / 迁移。

## Error Handling

- 不变。SSE handler 在 `fmt.Fprint` 返回错误时退出循环，与现有逻辑一致。
- `loadAll()` 的 `try/catch` 已在 UI 文本中显示 `"load error: " + e`，不变。

## Testing Strategy

- **单元测试**：
  - 现有 `TestServer_ShutdownWithoutWatcher` / `TestServer_ConcurrentSubscribeBroadcast` / `TestServer_WatcherStartedOnce` 不变（验证 #5 修复不破坏回归保护）。
  - 新增 `TestServer_ShutdownTimerReleased`：在 `watcherRunning` 已关闭后调用 `Shutdown`，断言不再有 pending timer（Go runtime 暂无直接观测 API，改为断言 `Shutdown` 路径在 `watcherRunning` 立即就绪时 < 100ms 返回，间接证明 timer 路径不会被无谓等待）。
- **手动验证**（写在 docs 或 PR 描述里）：
  - Playwright/curl 验证 SSE 在 `.kiro/specs/*/requirements.md` `touch` 后 2-3s 内收到 `event: refresh`。
  - 浏览器 Network 面板验证 5s 周期内 `/api/summary` 与 `/api/specs` 各 1 次。

## Migration / Rollout

- 不适用。本次无 API break（服务端 SSE 事件名固定，客户端单文件本地刷新）。
- 风险：若用户已有 fork / 第三方客户端订阅 `event: change`，会在合并后看到 `change` 事件不再触发。本次服务端从未发出过 `event: change`，所以零兼容性问题（grep 验证：服务端只发 `ping` / `refresh`）。
