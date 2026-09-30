# housekeeping-cleanup — Design

## Architecture

纯 cleanup，每个 finding 独立。最大改动是抽出 `internal/text` 包 + 共享
`RangeLines` helper，其余都是单文件小改动。

## 2.1 共享 helper（AC-6）

新建 `internal/text/text.go`，放 `RangeLines(text string) []string`。行为
与现 `lint/ears.go:rangeLines` 完全一致（output 不含 trailing newline，
空 text 返回 nil）。`lint/ears.go` 和 `taskgraph/parse.go` 删本地 copy，
改 import `internal/text`。

包名选择 `text` 是因为：
- 不与 stdlib `text/template` 等冲突（小写包路径）
- 职责清晰（任何文本操作 helper 都汇到这里）

## 2.2 Hook.Disabled 字段（AC-4）

`internal/models/hook.go` 加 `Disabled bool` 字段。`hooks/envelope.go` 的
normaliseHook 从 `Disabled` 字段读取（v1 envelope 用 `enabled: false`
等价）。`hooks/dispatch.go:64-70` 删 timeout=0 解析分支，改读
`h.Disabled`：

```go
if h.Disabled {
    res.OK = false
    res.Error = "hook disabled"
    return res
}
```

## 2.3 URL() IPv6 安全（AC-3）

`visualize/server.go:URL()` 把 `addr` (host:port) 拼到 `http://` 后面。
IPv6 addr 形如 `[::1]:8080`，直接拼接得到 `http://[::1]:8080` — 但
现有 `portOnly` helper（server_static.go）可能误处理。统一用
`net.SplitHostPort` 拆 host + port，再判断是否需要 IPv6 brackets。

```go
func (s *Server) URL() string {
    addr := s.srv.Addr
    host, port, err := net.SplitHostPort(addr)
    if err != nil {
        return "http://" + addr
    }
    if strings.Contains(host, ":") { // IPv6 literal
        return "http://[" + host + "]:" + port
    }
    return "http://" + host + ":" + port
}
```

## 2.4 Baseline JSON tag（AC-5）

字段 `IgnoredCodes` 不动（godoc 习惯：plural field name 描述
"the codes ignored"）；把 JSON tag `ignored_issues` 改成 `ignored_codes`，
与字段名一致。但**注意**这会破坏现存 `.baseline.json` 文件（spec 历史
资产）— 需要同步给 7 个历史 spec 改 baseline.json。

更保守的方案：保留 `ignored_issues` JSON tag，**仅加 unit test** 钉住
schema。这样既有资产不破坏，新作者不会被歧义困惑。我选保守。

## 2.5 dead code 清理（AC-8）

- `lint/baseline.go:90-92` `Empty()` 方法 — grep 全仓无 caller，删
- `hooks/dispatch.go:117-118` `var _ = ferrors.New` — 删 + 删 `ferrors` import
- `spec/engine.go:228-229` `var _ = json.Marshal` — 删 + 删 `encoding/json` import

## 2.6 错误处理（AC-1, AC-2）

`spec/engine.go:83` 的 stderr 直接打印改用 `cli/print.go` 已有的 logger
（如果有），或者加一个 minimal helper。若没有 logger，沿用 stderr 但加
注释说明这是临时方案。

`spec/engine.go:207-209` 的 `_ = e.ws.ClearCurrent()` 改成：
```go
if err := e.ws.ClearCurrent(); err != nil {
    fmt.Fprintf(os.Stderr, "warning: could not clear .current: %v\n", err)
}
```

## 2.7 godoc 同步（AC-9）

`visualize/server.go:12` 注释里 "A future enhancement can add fsnotify
+ SSE for real-time updates" 与 server_sse.go 已有 SSE 矛盾 — 改写为
"Polling watcher triggers SSE broadcasts every 2 seconds. fsnotify
support is a future enhancement for sub-second updates."

`visualize/server.go:107-109` "Best-effort: 200 ms is enough on CI"
应改为 "Best-effort: bounded by `shutdownTimeout` (5s) so a never-
started watcher doesn't block Shutdown forever."

## Components

| Component | Responsibility | API |
|---|---|---|
| `internal/text.RangeLines` | 共享行迭代 helper | `RangeLines(text string) []string` |
| `internal/models.Hook.Disabled` | 显式禁用标志 | bool |
| `visualize.Server.URL()` | IPv6 安全 URL 拼接 | `URL() string` |
| `spec.Engine.NewSpec/Complete` | 不直接打 stderr | wrap or helper |

## Data Model

- `internal/models/hook.go`: 加 `Disabled bool` 字段（零值 false）
- `internal/text/text.go`: 新建包，含 RangeLines
- `internal/lint/baseline.go`: 删 `Empty()` 方法
- `internal/lint/ears.go`: 删 `rangeLines`，import `internal/text`
- `internal/taskgraph/parse.go`: 删 `rangeLines` + `atoi`，import `internal/text` + `strconv`

## Error Handling

- AC-2: ClearCurrent 错误 wrap 报出，不 panic
- AC-1: warning 路径用统一 logger/stderr helper，格式一致

## Testing Strategy

| 测试 | 覆盖 | 关键 case |
|---|---|---|
| `internal/text/text_test.go`（新） | AC-6 | RangeLines 对各种 input（含末尾无 newline、纯 newline、empty）输出与旧实现一致 |
| `internal/hooks/dispatch_test.go`（增） | AC-4 | `Disabled=true` → res.OK=false res.Error 包含 "disabled" |
| `internal/visualize/server_sse_test.go`（增） | AC-3 | URL() 在 IPv6 / IPv4 / 仅 port 输入下都返回合法 URL |
| `internal/lint/baseline_test.go`（增） | AC-8 | Baseline JSON tag 仍含 `ignored_issues`（钉住 schema） |
| `internal/spec/engine_test.go`（增） | AC-1/AC-2 | 用 mock WS 验证 WriteCurrent/ClearCurrent 错误被 surface |

每个 refactor 至少跑既有测试 + 新增 1-2 个 case。

## Migration / Rollout

- AC-5 选保守方案：不改 JSON tag，只钉 schema test → 既有 .baseline.json
  文件零影响。
- AC-4: `Hook.Disabled` 是新字段，零值 false，老 envelope 解析时不读
  此字段则 hook 仍按 enabled=true 处理（既有 envelope 都显式有
  `enabled: true`，所以零值不破坏）。
- AC-6: `internal/text` 新包，不影响其他代码路径。