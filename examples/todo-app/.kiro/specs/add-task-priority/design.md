# add-task-priority — Design

## Architecture

最小命令行工具,3 文件:

```
examples/todo-app/
  ├── main.go           (NEW — cobra CLI; 2 子命令:add / list)
  ├── store.go          (NEW — ~/.todo-tasks.json 原子读写)
  └── sort.go           (NEW — 按 priority asc, created_at asc)
```

依赖:**仅 Go 标准库**(`encoding/json` / `os` / `time` / `sort`) +
`github.com/spf13/cobra` — 不引入第三方持久化 / 排序库。复用 free-kiro
项目根的 `go.mod` 通过 `replace` 指令直接借用(`replace
github.com/jingyu525/free-kiro => ../..`)。

```
TodoStore (struct)
  ├─ Path string                  # 默认 ~/.todo-tasks.json
  ├─ Load() ([]Task, error)       # 缺文件 → ([], nil)
  ├─ Save(tasks []Task) error     # 原子写:tmp + rename
  └─ Append(t Task) (Task, error) # 分配 id + 持久化 + 返回

Task (struct, JSON 字段:id/text/priority/done/created_at)
```

## Data Model

```go
type Priority string
const (
    PriorityP0 Priority = "P0"
    PriorityP1 Priority = "P1"
    PriorityP2 Priority = "P2"
)

func ParsePriority(s string) (Priority, error) {
    switch s {
    case "P0", "P1", "P2":
        return Priority(s), nil
    case "":
        return PriorityP2, nil  // default per AC
    }
    return "", fmt.Errorf("invalid priority: %s (must be P0, P1, or P2)", s)
}

type Task struct {
    ID        int64     `json:"id"`
    Text      string    `json:"text"`
    Priority  Priority  `json:"priority"`
    Done      bool      `json:"done"`
    CreatedAt time.Time `json:"created_at"`
}
```

`id` 全局单调递增;加载时取 max(id)+1 作为下一个 id(避免重启后 id 回退)。
`created_at` 写入时取 `time.Now().UTC()`。

## Error Handling

| 场景 | 退出码 | 类型 |
|---|---|---|
| `--priority` 不在 P0/P1/P2 | 3 | `errors.NewUsageError` |
| `~/.todo-tasks.json` 内容是无效 JSON | 2 | `Wrap("store.load", err, ...)` |
| 写文件失败(磁盘满 / 权限) | 2 | `Wrap("store.save", err, ...)` |
| `~/.todo-tasks.json.tmp` 已存在(并发?) | 2 | 覆盖前先 `os.Remove` |

`store.Save` 流程:`open tmp with O_CREATE|O_TRUNC|O_WRONLY` →
`json.NewEncoder(tmp).Encode(tasks)` → `tmp.Sync()` → `tmp.Close()`
→ `os.Rename(tmp, path)`(原子)。任意中间失败回滚:删 tmp,返回 error。

## Testing Strategy

### 单元测试(每个文件至少 1 个表驱动测试)

| 文件 | 关键 case |
|---|---|
| `sort_test.go` | 混合优先级 / 同优先级不同 created_at / 空数组 / 单元素 |
| `store_test.go` | Load 缺文件 → ([], nil) / Save 原子写(tmp 不残留) / Save 后 Load 字段一致 |
| `main_test.go` | `add` 默认 P2 / `add --priority X` 拒绝 / `list` 排序 / `list` 空文件 |

### 端到端冒烟(本 spec 不强制,但 spec complete 前手动跑)

```bash
cd examples/todo-app && go run . add "buy milk" --priority P0
cd examples/todo-app && go run . add "write spec"
cd examples/todo-app && go run . list
# 预期:第一行 P0,第二行 P2
```

## Migration / Rollout

无迁移(全新代码)。`~/.todo-tasks.json` 是用户数据,首次运行才创建。
无 feature flag、无 staged rollout。