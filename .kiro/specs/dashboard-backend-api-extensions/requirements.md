# dashboard-backend-api-extensions

<!--
为 free-kiro serve dashboard 提供后端 API 扩展，让前端能展示详情视图所需数据：
  - /api/spec/<name>/tasks   完整 task 列表按 wave 分组（驱动后续 spec-detail-view）
  - /api/spec/<name>/drift   drift signal 独立 URL（驱动后续 spec-detail-view）
  - /api/spec/<name>/timeline 该 spec 文件 mtime 列表（驱动后续 drift 详情"源文件"列）
  - /api/health              liveness + 上次 refresh 时间（驱动前端 health indicator）
  - /api/hooks               hook 注册表 + enabled 状态（驱动后续 settings 视图）
  - BuildReport.ProjectReport.mode  workspace-missing|no-specs|ok 三态（驱动前端空状态分桶）

所有新端点必须向后兼容：现有 /api/summary 只增字段不删字段，老客户端零影响。
本 spec 不引入 fsnotify、不引入 SSE heartbeat、不引入中间件（属于 dashboard-backend-middleware）。
本 spec 不写前端 UI（属于 dashboard-frontend-foundation）。
-->

## User Stories

- As a dashboard 前端开发者 I want 拿到 spec 的完整 task 列表（按 wave 分组，含 id/title/done/deps/wave_index）so that 详情视图能展示 waves 进度条与可点击的任务列表。
- As a dashboard 前端开发者 I want drift signals 独立 URL so that 详情视图能 deep-link 与 ETag 缓存。
- As a dashboard 前端开发者 I want 知道每个 spec 目录下文件最近修改时间 so that drift 详情能展示"哪个文件改了"。
- As a dashboard 前端开发者 I want 知道 .kiro/ 工作区是否真的存在 so that 区分"没初始化"与"没创建 spec"两种空状态。
- As a dashboard 用户 I want health 端点告诉前端服务是否健康 + 上次成功刷新时间 so that 监控集成 / health-check 探针可对接。
- As a dashboard 用户 I want 看到当前 hook 注册表 + enabled 状态 so that 知道哪些事件触发哪些动作。
- As a 后端维护者 I want 新增端点完全向后兼容 so that 现有 dashboard-sse-bugfix / `free-kiro status --json` 等消费者零回归。

## Acceptance Criteria

### Engine.TaskList 新方法

- [AC-1] THE SYSTEM SHALL 在 `internal/spec` 包新增 `Engine.TaskList(name string) ([]taskgraph.Wave, error)` 方法（`taskgraph.Wave` 由 `internal/taskgraph/waves.go` 现有 `ExecutionWaves` 返回）。
- [AC-2] WHEN 调用 `Engine.TaskList(name)` 且 spec 存在且 `tasks.md` 可读 THE SYSTEM SHALL 返回非 nil 的 `[]Wave`，每个 `Wave` 包含 `Index int`（1-based）、`Tasks []models.Task`、`Done int`、`Total int`。
- [AC-3] WHEN spec 存在但 `tasks.md` 不存在或为空 THE SYSTEM SHALL 返回空切片 `[]Wave{}` 与 nil error（不报错）。
- [AC-4] WHEN spec 不存在 THE SYSTEM SHALL 返回 nil 与 `SpecNotFound` 错误（用 `errors.Is(err, ErrSpecNotFound)` 判定）。
- [AC-5] THE SYSTEM SHALL 复用现有 `taskgraph.ParseTasks` + `taskgraph.ExecutionWaves`，不重新实现 wave 计算（零重复，AGENT_RULES §6）。
- [AC-6] THE SYSTEM SHALL `TaskList` 不调 `BuildReport`，不触发整 workspace 扫描，单 spec O(n) 其中 n = task 数。

### `/api/spec/<name>/tasks` 端点

- [AC-7] WHEN 客户端 GET `/api/spec/<name>/tasks` 且 spec 存在 THE SYSTEM SHALL 返回 `200 OK` + `application/json` + `{ "spec": "<name>", "waves": [{"index": 1, "done": 2, "total": 3, "tasks": [{"id": 1, "title": "...", "done": true, "deps": [2,3]}]}, ...], "summary": {"total": 5, "done": 2, "remaining": 3, "waves": 2} }`。
- [AC-8] WHEN spec 不存在 THE SYSTEM SHALL 返回 `404 Not Found` + `{"error":"spec not found","spec":"<name>"}`。
- [AC-9] WHEN URL path 含特殊字符（如空格、中文、转义） THE SYSTEM SHALL 使用 `net/url.PathUnescape` 解码后再查 spec，找不到则 404。
- [AC-10] THE SYSTEM SHALL handler 不依赖任何 SSE/watcher 状态；纯计算，响应时间 ≤ 50ms（10 个 task 内）。

### `/api/spec/<name>/drift` 端点

- [AC-11] WHEN 客户端 GET `/api/spec/<name>/drift` 且 spec 存在且已 approve THE SYSTEM SHALL 返回 `200 OK` + `{"spec": "<name>", "signals": [{"key":"ac_count","baseline":5,"current":3,"delta":-2}, ...]}`。
- [AC-12] WHEN spec 未 approve THE SYSTEM SHALL 返回 `200 OK` + `{"spec": "<name>", "signals": []}`（空数组，无 baseline 谈何 drift）。
- [AC-13] WHEN spec 不存在 THE SYSTEM SHALL 返回 `404 Not Found`。
- [AC-14] THE SYSTEM SHALL handler 复用现有 `Engine.Drift(name)` 或 `spec.computeDrift`，不重新解析 meta/baseline。

### `/api/spec/<name>/timeline` 端点

- [AC-15] WHEN 客户端 GET `/api/spec/<name>/timeline` 且 spec 存在 THE SYSTEM SHALL 返回 `200 OK` + `{"spec": "<name>", "files": [{"path": "requirements.md", "mtime": "2026-10-02T01:23:45Z", "size": 1234}, ...]}`。
- [AC-16] WHEN spec 不存在 THE SYSTEM SHALL 返回 `404 Not Found`。
- [AC-17] THE SYSTEM SHALL `files` 数组按 `mtime` 倒序（最新在前），最多 50 条。
- [AC-18] THE SYSTEM SHALL 实现 `Server.collectSpecMtimes(name string)` 方法（与现有 `collectMtimes` 解耦），仅遍历 `.kiro/specs/<name>/` 目录。
- [AC-19] THE SYSTEM SHALL 文件大小 < 0（stat 失败）记 `size: 0` 但仍保留 entry；不静默吞错（AGENT_RULES §2）。

### `/api/health` 端点

- [AC-20] THE SYSTEM SHALL 在 `Server.handleHealth` 实现 GET `/api/health`，返回 `200 OK` + `{"status": "ok", "uptime_seconds": <int>, "last_refresh_at": "<RFC3339>", "subscribers": <int>, "version": "<free-kiro 版本>"}`。
- [AC-21] WHEN 服务未启动完成（NewServer 后未 ServeWith） THE SYSTEM SHALL 仍返回 200 + `{"status": "starting"}`。
- [AC-22] THE SYSTEM SHALL `last_refresh_at` 在每次 SSE `refresh` 事件 broadcast 时更新。
- [AC-23] THE SYSTEM SHALL `version` 字段来自 build-time 注入的 `internal/version.Version` 常量。

### `/api/hooks` 端点

- [AC-24] THE SYSTEM SHALL GET `/api/hooks` 返回 `200 OK` + `{"hooks": [{"id": "...", "event": "...", "enabled": true, "action_type": "shell|agent", "glob": "..."}, ...]}`。
- [AC-25] WHEN hook registry 不可用（`hooks.Registry` 未初始化） THE SYSTEM SHALL 返回 `200 OK` + `{"hooks": [], "registry_disabled": true}`（**不**报 503，方便前端降级）。
- [AC-26] THE SYSTEM SHALL handler 不调用 `Dispatch`（不触发 hook 执行），仅读取 registry。

### BuildReport.mode 字段

- [AC-27] THE SYSTEM SHALL 在 `internal/visualize/report.go` 的 `ProjectReport` 结构体新增 `Mode string` 字段（json tag `mode`）。
- [AC-28] THE SYSTEM SHALL `BuildReport()` 计算 mode：`!ws.KiroDir() exists → "workspace-missing"`、`KiroDir exists && len(specs)==0 → "no-specs"`、`else → "ok"`。
- [AC-29] THE SYSTEM SHALL mode 计算是 O(1)（仅一次 stat + 一次 len），不增加 BuildReport 整体耗时超过 1ms。
- [AC-30] THE SYSTEM SHALL `ws.KiroDir() exists` 用 `os.Stat` 实现，找不到返回 `os.IsNotExist` → mode = `workspace-missing`。
- [AC-31] THE SYSTEM SHALL `WorkspacePaths` 接口需新增 `KiroDirExists() bool` 方法（`internal/workspace/workspace.go` 实现）。

### 回归保护 — Unchanged Behavior

- [AC-32] THE SYSTEM SHALL CONTINUE TO `/api/summary` 包含 `generated_at`、`specs[]`、`active` 字段，新增 `mode` 不删任何旧字段。
- [AC-33] THE SYSTEM SHALL CONTINUE TO `/api/spec/<name>`（现有端点）行为不变：返回 `engine.Status(name)` map，handler `internal/visualize/server.go:227-239` 不动。
- [AC-34] THE SYSTEM SHALL CONTINUE TO SSE 事件名 `ping`/`refresh`，不引入新事件类型（事件分级属于 dashboard-realtime-fsnotify）。
- [AC-35] THE SYSTEM SHALL CONTINUE TO stdlib `net/http`，不引入任何第三方依赖。
- [AC-36] THE SYSTEM SHALL CONTINUE TO `/api/spec/<name>/tasks` 失败时不写入任何 `.kiro/` 文件（只读语义）。
- [AC-37] THE SYSTEM SHALL CONTINUE TO `BuildReport()` 单次调用在 100 specs workspace 下 ≤ 100ms（保留现有 benchmark 性能预算）。

### 错误处理

- [AC-38] WHEN 任意新端点 handler panic THE SYSTEM SHALL 返回 500 + `{"error":"internal server error"}`，不挂死 http.Server（recover 由 dashboard-backend-middleware 提供；本 spec 假定 middleware 已就绪或自行 defer recover 兜底）。
- [AC-39] WHEN URL path 含 `/` 多级嵌套（如 `/api/spec/foo/bar/tasks`） THE SYSTEM SHALL 把整段 `foo/bar` 当作 spec name 查找；找不到则 404（保留现有 prefix 匹配行为）。
- [AC-40] THE SYSTEM SHALL 不在响应 body 中泄露绝对路径（即使 stat 失败），仅返回 spec name 与相对路径。

## Out of Scope

- middleware / panic recovery / access log / cache headers —— 属于 dashboard-backend-middleware 独立 spec。
- fsnotify 替换 2s mtime 轮询 —— 属于 dashboard-realtime-fsnotify。
- SSE heartbeat / 事件分级 / ETag 304 —— 属于 dashboard-realtime-fsnotify。
- 移动端适配 / a11y —— 属于 dashboard-mobile-a11y。
- 前端 UI / 组件实现 —— 属于 dashboard-frontend-foundation 与 dashboard-spec-detail-view。
- `<dialog>` 元素使用 —— 属于 dashboard-spec-detail-view。
- hook 启用/禁用的 POST/PATCH 端点 —— 仅 GET（只读），写操作属于 hook 子命令（`free-kiro hook add`）未来扩展。
