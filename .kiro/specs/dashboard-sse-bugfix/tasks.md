# dashboard-sse-bugfix — Tasks

<!--
依赖规则：
  - #1/#2/#3/#4 各自独立（无依赖）
  - #5 验证 #1/#2/#3/#4 全部完成后做端到端冒烟
-->

- [x] #1 SSE 自动刷新：客户端改订阅 `event: refresh`（替换 `event: change`）于 `static/index.html`
- [x] #2 loadAll 去重：合并 loadSummary + loadSpecs 为单次 `fetch('/api/summary')`，渲染两个 view（stat + 表格），消除每 5s 2x 冗余请求于 `static/index.html`
- [x] #3 footer 文案：补 SSE 描述于 `static/index.html`
- [x] #4 Shutdown timer 释放：`time.After(shutdownTimeout)` → `time.NewTimer` + `defer t.Stop()` 于 `internal/visualize/server.go`
- [x] #5 端到端冒烟：启 `free-kiro serve` + Playwright 验证 SSE 事件链路 + Network 面板无冗余请求 [deps: #1,#2,#3,#4]
