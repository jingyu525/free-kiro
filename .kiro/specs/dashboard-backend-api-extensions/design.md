# dashboard-backend-api-extensions — Design

<!--
本 spec 是纯后端 API 扩展，不引入中间件、不引入 fsnotify、不引入 SSE 改造。
唯一 Go 代码新增/修改：
  - internal/spec/engine_state.go: 新增 Engine.TaskList(name) 方法
  - internal/visualize/server.go: 新增 5 个 handler + ProjectReport.Mode 字段透传
  - internal/visualize/report.go: ProjectReport 加 Mode 字段；BuildReport 计算 mode
  - internal/workspace/workspace.go: WorkspacePaths 接口加 KiroDirExists() bool

唯一架构原则：复用现有 taskgraph / hooks / workspace 模块，不重新实现（AGENT_RULES §6 零死代码）。
-->

## Architecture

```
GET /api/spec/<name>/tasks
  → handleSpecTasks(w, r)
    → url.PathUnescape + 取 name
    → eng.TaskList(name)
      → specDir := ws.SpecDir(name)
      → tasksMd := specDir + "/tasks.md"
      → taskgraph.ParseTasks(tasksMd)         # 复用现有解析器
      → taskgraph.ExecutionWaves(tasks)       # 复用现有 wave 计算
      → map 为 []taskgraph.Wave                # 包装返回结构
    → writeJSON 200 OK {spec, waves, summary}

GET /api/spec/<name>/drift
  → handleSpecDrift(w, r)
    → eng.Status(name).drift                  # 复用现有 engine.Status 字段
    → writeJSON 200 OK {spec, signals}

GET /api/spec/<name>/timeline
  → handleSpecTimeline(w, r)
    → Server.collectSpecMtimes(name)          # 新方法：仅遍历 specDir
    → 排序 mtime desc + 截断 50
    → writeJSON 200 OK {spec, files}

GET /api/health
  → handleHealth(w, r)
    → uptime := time.Since(s.startedAt)       # Server 字段新增 startedAt
    → lastRefreshAt := s.lastRefreshAt.Load() # atomic.Int64 新字段
    → writeJSON 200 OK {status, uptime_seconds, last_refresh_at, subscribers, version}

GET /api/hooks
  → handleHooks(w, r)
    → hooks.Registry.List()                   # 复用现有 registry
    → writeJSON 200 OK {hooks}
    → IF registry 未初始化 → {hooks: [], registry_disabled: true}

BuildReport:
  → ws.KiroDirExists()                        # 新接口方法
    → os.Stat(.kiro/) 失败 → false
  → mode = "workspace-missing" | "no-specs" | "ok"
  → ProjectReport.Mode = mode
```

不变：
- `/api/summary` payload 仅加 `mode` 字段（向后兼容）。
- `/api/spec/<name>`（现有）行为不变。
- SSE 事件名 `ping`/`refresh` 不动。
- stdlib `net/http` + embed.FS 不变。

## Components

| Component | File | Responsibility |
|---|---|---|
| `Engine.TaskList(name)` | `internal/spec/engine_state.go` | 解析 tasks.md + 计算 waves + 返回结构化结果 |
| `Server.handleSpecTasks` | `internal/visualize/server.go` | 路由 `/api/spec/<name>/tasks` |
| `Server.handleSpecDrift` | `internal/visualize/server.go` | 路由 `/api/spec/<name>/drift` |
| `Server.handleSpecTimeline` | `internal/visualize/server.go` | 路由 `/api/spec/<name>/timeline` |
| `Server.collectSpecMtimes` | `internal/visualize/server_sse.go` 或 `server.go` | 仅遍历 `.kiro/specs/<name>/` 目录的 mtime + size |
| `Server.handleHealth` | `internal/visualize/server.go` | 路由 `/api/health` |
| `Server.handleHooks` | `internal/visualize/server.go` | 路由 `/api/hooks` |
| `Server.startedAt` | `internal/visualize/server.go` | `NewServer` 时 `time.Now()`（新增字段） |
| `Server.lastRefreshAt` | `internal/visualize/server.go` | `atomic.Int64`，broadcast 时 `Store(time.Now().UnixNano())`（新增字段） |
| `Workspace.KiroDirExists` | `internal/workspace/workspace.go` | `os.Stat(KiroDir())` 实现 |
| `ProjectReport.Mode` | `internal/visualize/report.go` | JSON 字段 + BuildReport 计算 |

## Data Model

```go
// internal/visualize/server.go (新增 handler 局部类型)

// TasksResponse
type TasksResponse struct {
    Spec    string             `json:"spec"`
    Waves   []TaskWaveResponse `json:"waves"`
    Summary TaskSummaryResponse `json:"summary"`
}
type TaskWaveResponse struct {
    Index int              `json:"index"`         // 1-based
    Done  int              `json:"done"`
    Total int              `json:"total"`
    Tasks []TaskItemResponse `json:"tasks"`
}
type TaskItemResponse struct {
    ID    int      `json:"id"`
    Title string   `json:"title"`
    Done  bool     `json:"done"`
    Deps  []int    `json:"deps"`              // 从 [deps: #N1,#N2] 解析
    Wave  int      `json:"wave"`              // 1-based（冗余 Index，便于前端过滤）
}
type TaskSummaryResponse struct {
    Total     int `json:"total"`
    Done      int `json:"done"`
    Remaining int `json:"remaining"`
    Waves     int `json:"waves"`
}

// DriftResponse
type DriftResponse struct {
    Spec    string                 `json:"spec"`
    Signals []spec.DriftSignal     `json:"signals"`  // 复用 internal/spec/drift.go 现有类型
}

// TimelineResponse
type TimelineResponse struct {
    Spec  string           `json:"spec"`
    Files []TimelineFileEntry `json:"files"`
}
type TimelineFileEntry struct {
    Path  string `json:"path"`         // 相对 specDir，如 "requirements.md"
    Mtime string `json:"mtime"`        // RFC3339
    Size  int64  `json:"size"`
}

// HealthResponse
type HealthResponse struct {
    Status         string `json:"status"`            // "ok" | "starting"
    UptimeSeconds  int64  `json:"uptime_seconds"`
    LastRefreshAt  string `json:"last_refresh_at"`   // RFC3339，可能为空（启动后无 refresh）
    Subscribers    int    `json:"subscribers"`
    Version        string `json:"version"`           // 来自 internal/version.Version
}

// HooksResponse
type HooksResponse struct {
    Hooks            []HookEntry `json:"hooks"`
    RegistryDisabled bool        `json:"registry_disabled,omitempty"`
}
type HookEntry struct {
    ID         string `json:"id"`
    Event      string `json:"event"`
    Enabled    bool   `json:"enabled"`
    ActionType string `json:"action_type"`   // "shell" | "agent"
    Glob       string `json:"glob,omitempty"`
}

// ProjectReport (扩)
type ProjectReport struct {
    GeneratedAt time.Time     `json:"generated_at"`
    Specs       []*SpecReport `json:"specs"`
    Active      string        `json:"active"`
    Mode        string        `json:"mode"`        // 新增："workspace-missing" | "no-specs" | "ok"
}

// internal/spec/engine_state.go (新增)
func (e *Engine) TaskList(name string) ([]taskgraph.Wave, error) {
    specDir := e.ws.SpecDir(name)
    if _, err := os.Stat(specDir); err != nil {
        if os.IsNotExist(err) {
            return nil, ErrSpecNotFound
        }
        return nil, fmt.Errorf("stat spec dir: %w", err)
    }
    tasksMdPath := filepath.Join(specDir, "tasks.md")
    tasks, err := taskgraph.ParseTasks(tasksMdPath)
    if err != nil {
        if errors.Is(err, taskgraph.ErrTasksFileNotFound) {
            return []taskgraph.Wave{}, nil  // 空切片而非 nil
        }
        return nil, fmt.Errorf("parse tasks: %w", err)
    }
    waves := taskgraph.ExecutionWaves(tasks)
    summary := taskgraph.Summary(tasks)  // TaskSummary{Total, Done, Remaining, Waves}
    return waves, nil
}
```

## Error Handling

| 错误 | 触发条件 | 响应 |
|---|---|---|
| `SpecNotFound` | `ws.SpecDir(name)` 不存在 | `404` + `{"error":"spec not found","spec":"<name>"}` |
| `tasks.md 不存在` | spec 存在但无 tasks.md | `200` + 空 waves（不报错） |
| `tasks.md 解析失败` | 行级正则不匹配 / IO 失败 | `500` + `{"error":"parse tasks: <wrapped>"}` |
| `path 含 ..` | URL 路径穿越 | `400` + `{"error":"invalid spec name"}` |
| `registry 未初始化` | hooks.Registry nil | `200` + `{"hooks":[],"registry_disabled":true}`（降级不报错） |
| `panic` | handler 内任意 panic | `500` + `{"error":"internal server error"}`（middleware 提供） |

**零吞错误**（AGENT_RULES §2）：所有 `os.Stat` / `os.ReadFile` / `taskgraph.ParseTasks` 失败都必须显式处理或 wrap 错误向上抛；不允许 `_ = doX()` 或 `log.Print(err)` 后继续。

## Testing Strategy

### 单元测试（覆盖率新增行 ≥ 80%）

```go
// internal/spec/engine_tasklist_test.go (新增)
func TestEngine_TaskList_TableDriven(t *testing.T) {
    cases := []struct {
        name      string
        setup     func(t *testing.T, dir string)  // 写 spec 目录 + tasks.md
        wantErr   error
        wantWaves int
        wantDone  int
    }{
        {"happy path: 3 tasks 1 wave", setupSimpleTasks, nil, 1, 1},
        {"deps: 2 waves", setupTasksWithDeps, nil, 2, 0},
        {"empty tasks.md", setupEmptyTasks, nil, 0, 0},
        {"missing tasks.md", setupNoTasksMd, nil, 0, 0},
        {"missing spec dir", setupNoSpec, ErrSpecNotFound, 0, 0},
        {"cycle detection", setupCyclicDeps, taskgraph.ErrCycle, 0, 0},  // 现有 cycle.go 检测
    }
    // ...
}

// internal/visualize/server_handlers_test.go (新增)
func TestServer_HandleSpecTasks(t *testing.T) { /* ... */ }
func TestServer_HandleSpecDrift(t *testing.T) { /* ... */ }
func TestServer_HandleSpecTimeline(t *testing.T) { /* ... */ }
func TestServer_HandleHealth(t *testing.T) { /* ... */ }
func TestServer_HandleHooks_Disabled(t *testing.T) { /* ... */ }
func TestServer_HandleSpecTasks_NotFound(t *testing.T) { /* ... */ }
func TestServer_HandleSpecTasks_PathEscape(t *testing.T) { /* ... */ }

// internal/visualize/report_test.go (扩)
func TestBuildReport_Mode_WorkspaceMissing(t *testing.T) { /* ... */ }
func TestBuildReport_Mode_NoSpecs(t *testing.T) { /* ... */ }
func TestBuildReport_Mode_Ok(t *testing.T) { /* ... */ }
```

### 回归测试（保留现有）

- `TestServer_ShutdownWithoutWatcher`（server_sse_test.go:24）
- `TestServer_ConcurrentSubscribeBroadcast`（server_sse_test.go:50）
- `TestServer_WatcherStartedOnce`（server_sse_test.go:82）
- `TestServer_URL`（server_sse_test.go:116）

### E2E（沿用 `.playwright-mcp/`，不入 CI）

```bash
free-kiro serve --port 7373 &
sleep 1
curl -sf http://127.0.0.1:7373/api/health | jq -e '.status == "ok"'
curl -sf http://127.0.0.1:7373/api/summary | jq -e '.mode'  # 新字段存在
curl -sf http://127.0.0.1:7373/api/spec/dashboard-sse-bugfix/tasks | jq -e '.waves | length'
curl -sf http://127.0.0.1:7373/api/spec/dashboard-sse-bugfix/drift | jq -e '.signals'
curl -sf http://127.0.0.1:7373/api/spec/dashboard-sse-bugfix/timeline | jq -e '.files | length'
curl -sf http://127.0.0.1:7373/api/spec/nonexistent/tasks | jq -e '.error == "spec not found"'
curl -sf http://127.0.0.1:7373/api/hooks | jq -e '.hooks'
kill %1
```

## Migration / Rollout

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| `/api/summary` | **新增** `mode` 字段；老客户端忽略未知字段，零 break |
| `/api/spec/<name>` | **不变**（现有 handler 不动） |
| SSE 事件名 | **不变**（`ping`/`refresh`，dashboard-sse-bugfix 固化） |
| Go embed | **不变**（仅 dashboard-frontend-foundation / dashboard-backend-middleware 加 `static/dist`） |
| 客户端 `free-kiro status --json` | 复用 `BuildReport`，自动拿到 `mode` 字段；零 break |
| `Engine.Status(name)` map | **不变**（不删字段） |

### 风险与回滚

- **风险 A**：新 handler 写错路径导致 panic → 整个 dashboard 挂。**对策**：dashboard-backend-middleware 的 withRecover 中间件（前提）或本 spec 内部 `defer recover()` 兜底（双保险）。
- **风险 B**：`BuildReport.mode` 计算 stat 慢 → 整体 dashboard 卡。**对策**：`os.Stat` 单次调用，O(1)；CI benchmark 验证 ≤ 1ms。
- **风险 C**：路径穿越（`/api/spec/..%2F..%2F/tasks`）。**对策**：`url.PathUnescape` 后 `filepath.Clean`，检查是否含 `..` 或以 `/` 开头，命中则 400。
- **风险 D**：hook registry 未初始化导致 503 → 前端 health check 误报。**对策**：handler 降级返回 `registry_disabled: true`，让前端 UI 显示"hook 不可用"而非"服务挂了"。

### 落地前置（按 AGENT_RULES §1）

- 本 spec 三件套（requirements / design / tasks）必须 `free-kiro lint` 全绿。
- `dashboard-backend-middleware` 是本 spec 的**软依赖**（withRecover 提供 panic 兜底）；若 middleware 还没合入，本 spec 应在 handler 顶部加 `defer func(){ if r := recover(); r != nil { ... } }()` 自保。
- 落地后由用户显式 `free-kiro spec approve dashboard-backend-api-extensions` 后再 `spec start`。
