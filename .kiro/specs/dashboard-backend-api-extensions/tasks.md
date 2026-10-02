# dashboard-backend-api-extensions — Tasks

<!--
依赖规则：
  - #1 Engine.TaskList 是 #2 的前置（handler 复用方法）
  - #3-#6 handler 可与 #2 并行（互不依赖）
  - #7 BuildReport.mode 与所有 handler 并行
  - #8 测试在 #1-#7 完成后做
  - #9-#10 验证类必须在 #8 完成后
-->

- [x] #1 Engine.TaskList 新方法：实现 `internal/spec/engine_state.go` 的 `(*Engine).TaskList(name string) ([]Wave, error)` + `Wave` 类型 + `ErrSpecNotFound` 常量；复用 `taskgraph.ParseTasks` + `taskgraph.ExecutionWaves`；spec 不存在返回 ErrSpecNotFound；tasks.md 缺失返回空切片
- [x] #2 `/api/spec/<name>/tasks` handler：实现 `Server.handleSpecTasks`，复用 `eng.TaskList(name)`；JSON 序列化 `tasksResponse`（spec + waves[] + summary）
- [x] #3 `/api/spec/<name>/drift` handler：实现 `Server.handleSpecDrift`，复用 `eng.Status(name).drift`；JSON 序列化 `driftResponse`；spec 未 approve 返回空 signals
- [x] #4 `/api/spec/<name>/timeline` handler：实现 `Server.handleSpecTimeline` + `Server.collectSpecMtimes(name)`；filepath.WalkDir 仅遍历 specDir；按 mtime desc 排序截断 50；JSON 序列化 `timelineResponse`
- [x] #5 `/api/health` handler：实现 `Server.handleHealth` + `Server` 字段新增 `startedAt time.Time`（NewServer 时初始化）+ `lastRefreshAt atomic.Int64`（broadcast 时更新）；JSON 序列化 `healthResponse`
- [x] #6 `/api/hooks` handler：实现 `Server.handleHooks`；listHooks 返回 nil 时降级返回 `{hooks:[],registry_disabled:true}`；JSON 序列化 `hooksResponse`
- [x] #7 BuildReport.mode 字段：扩 `ProjectReport` 加 `Mode string` json tag；`BuildReport()` 调用 `computeReportMode(ws, specs)`；`WorkspacePaths` 接口加 `KiroDirExists() bool`；`internal/workspace/workspace.go` 加 `KiroDirExists() bool` 方法
- [x] #8 单测（覆盖率新增行 ≥ 80%）：写 `engine_tasklist_test.go`（6 个表驱动 case + WaveIndexing = 7 子测试 PASS）+ `server_handlers_test.go`（8 个 handler test：happy/not found/path escape/drift/timeline sort/health/health after broadcast/hooks disabled/router dispatch = 15 子测试 PASS）+ `report_mode_test.go`（3 mode case + JSON wire format = 4 子测试 PASS）
- [x] #9 端到端验证：启 `free-kiro serve --port 7374` + curl 验证：`/api/health` 200 ok/uptime/dev ✅、`/api/hooks` 200 registry_disabled=true ✅、`/api/spec/dashboard-sse-bugfix/tasks` 200 + 5 tasks 2 waves ✅、`/api/spec/.../drift` 200 empty signals ✅、`/api/spec/.../timeline` 200 + 4 files mtime desc ✅、`/api/spec/nonexistent/tasks` 404 ✅、`/api/spec/..foo/tasks` 400 ✅、`/api/summary` mode="ok" ✅
- [x] #10 回归保护：`go test -race ./...` 全部 OK（spec + visualize + 全部子包）；`free-kiro status --json` 含 mode="ok" + 20 specs；E2E 验证 dashboard-sse-bugfix SSE 事件名 ping/refresh 链路未破；server_sse.go:168 死代码 `var _ = context.Background` 已清；`max()` 替换 if uptime<0 {uptime=0}
