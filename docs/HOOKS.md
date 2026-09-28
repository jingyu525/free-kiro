# Hooks — 事件驱动自动化

> Hook 让 IDE / 编辑器把"事件"接入 free-kiro：每次保存文件、每次工具调用、每次会话
> 开始时，自动触发匹配的命令或 agent prompt。free-kiro 写出兼容 Kiro 官方 v1 信封的
> JSON，**官方 Kiro IDE 可以直接加载同一份 hook 文件**。

## 概念

`.kiro/hooks/*.json` 下的 JSON 文件，每个文件声明：

```
trigger event + optional filter (glob or regex) + action (shell command or agent prompt)
```

匹配后：
- **shell action**：通过 `os/exec` 执行，event context 作为 JSON 写到 STDIN
- **agent action**：delegate point，由 host（IDE / CLI）注入模型 runner；free-kiro
  自身**不运行模型**

## 事件命名（两套兼容）

| free-kiro 风格 | Kiro 官方风格 | 触发时机 |
|---|---|---|
| `file.save` | `PostFileSave` | agent 保存文件后 |
| `file.create` | `PostFileCreate` | agent 创建新文件后 |
| `file.delete` | `PostFileDelete` | agent 删除文件后 |
| `prompt.submit` | `UserPromptSubmit` | 用户给 agent 发消息时 |
| `task.run` | `PreTaskExecution` / `PostTaskExecution` | spec 任务开始/结束时（IDE only） |
| `manual` | — | 手动触发（`hook run manual`） |
| — | `SessionStart` | 新会话开始（IDE only） |
| — | `PreToolUse` | 工具调用前 |
| — | `PostToolUse` | 工具调用后 |

任选一套皆可。free-kiro 写出的 hook 用 Kiro 风格（更广泛兼容），但读取时两种都接受。

## JSON 格式（v1 envelope）

`free-kiro hook add` 写出的格式：

```json
{
  "version": "v1",
  "hooks": [
    {
      "name": "lint-on-save",
      "trigger": "PostFileSave",
      "matcher": "\\.tsx$",
      "action": {
        "type": "command",
        "command": "npx eslint --fix $FILE"
      },
      "enabled": true,
      "description": "lint tsx files on save",
      "timeout": 30
    }
  ]
}
```

| 字段 | 必需 | 说明 |
|---|---|---|
| `version` | ✅ | 必须是 `"v1"` |
| `hooks` | ✅ | hook 数组 |
| `hooks[].name` | ✅ | hook 唯一 ID |
| `hooks[].trigger` | ✅ | 事件名（PascalCase 或小写） |
| `hooks[].matcher` |  | 正则，匹配文件路径（Kiro 风格） |
| `hooks[].glob` |  | glob，匹配文件路径（free-kiro 风格） |
| `hooks[].action.type` | ✅ | `"command"`（shell）或 `"agent"` |
| `hooks[].action.command` |  | shell 命令（action.type=command） |
| `hooks[].action.prompt` |  | agent prompt（action.type=agent） |
| `hooks[].enabled` |  | 默认 `true`，设为 `false` 不删除但禁用 |
| `hooks[].timeout` |  | shell 命令超时秒数（`0` = 禁用超时） |
| `hooks[].description` |  | 人类可读描述 |

`matcher`（regex）与 `glob` 二选一。Kiro v1 用 `matcher`，free-kiro 兼容两种。

## STDIN 信封（shell action 收到的 JSON）

```json
{
  "event": "PostFileSave",
  "file": "components/Button.tsx",
  "cwd": "/Users/you/project"
}
```

可用 `jq` 解析：

```bash
jq -r '.file' | xargs npx eslint --fix
```

或读取后操作：

```bash
FILE=$(jq -r '.file' /dev/stdin)
EVENT=$(jq -r '.event' /dev/stdin)
echo "hook fired for $FILE on $EVENT"
```

## CLI 命令

### 列出 hook

```bash
free-kiro hook list
```

```
ID                   EVENT            FILTER     TYPE    TO    ENABLED
lint-on-save         file.save        *.md       shell   -     true
```

### 新增 hook

```bash
free-kiro hook add \
  --id lint-on-save \
  --event file.save \
  --action-type shell \
  --action "gofmt -w" \
  --glob "*.go" \
  --description "format Go files"
```

参数：

| 参数 | 必需 | 说明 |
|---|---|---|
| `--id` | ✅ | hook 唯一 ID |
| `--event` | ✅ | 事件名 |
| `--action-type` | ✅ | `shell` 或 `agent` |
| `--action` | ✅ | shell 命令 或 agent prompt |
| `--glob` |  | glob 过滤 |
| `--description` |  | 描述 |
| `--timeout` |  | shell 超时秒数（默认 30，0 = 禁用） |
| `--disabled` |  | 写入但默认 enabled=false |

### 触发 hook（手动 / IDE hook runner）

```bash
free-kiro hook run <event> [--file <path>]
```

匹配 event + 可选文件路径的 hook 全部触发。

## IDE 集成示例（Claude Code）

写到 `~/.claude/settings.json`：

```json
{
  "hooks": [
    {
      "name": "free-kiro-lint-gate",
      "trigger": "PreToolUse",
      "matcher": "Edit|Write",
      "action": {
        "type": "command",
        "command": "free-kiro lint || exit 2"
      }
    },
    {
      "name": "free-kiro-session",
      "trigger": "SessionStart",
      "action": {
        "type": "command",
        "command": "free-kiro spec next $(free-kiro spec list 2>/dev/null | head -1 | awk '{print $1}')"
      }
    }
  ]
}
```

效果：
- 每次 `Edit` / `Write` 工具前自动 lint；ERROR 时 exit 2 → IDE 真正拦截写入
- 会话开头自动锁定当前 spec 阶段（打印 next 建议）

## 与官方 Kiro 互操作

`free-kiro hook add` 写出的 JSON 100% 符合官方 Kiro v1 信封：

```bash
# 在 free-kiro 项目中：
free-kiro hook add --id format-go --event file.save --action-type shell --action "gofmt -w" --glob "*.go"

# 生成的 .kiro/hooks/format-go.json 直接拷到 Kiro IDE 用户的 hooks 目录就能加载
```

或者反过来：

```bash
# 拿到 Kiro 官方写的 hook：
cp /path/to/kiro-hooks/lint.json .kiro/hooks/

# free-kiro 自动归一化读取：
free-kiro hook list   # 会显示这条 hook
```

## agent action 的 delegate point

`agent` action 不直接执行，而是把 prompt 交给 host：

```go
// 在 IDE / 自定义 runner 里：
results := registry.Dispatch(ctx, "manual", "", func(prompt string) (string, error) {
    return myModelRunner.Run(prompt)
})
```

`myModelRunner.Run` 可以是 Claude / GPT / 本地模型——free-kiro 不绑定。

直接 `free-kiro hook run <event>` 而不传 agent_fn，会得到占位符输出：

```
[agent hook] would execute prompt: "review the diff"
(free-kiro does not run agents — wire `agent_fn` to your own model/runner)
```

这是故意的：free-kiro 是被动的规划层，主动跑 prompt 是 host 的事。

## 调试技巧

```bash
# 1. 看 hook 是否被加载
free-kiro hook list

# 2. 手动触发 + 看输出
free-kiro hook run file.save --file main.go

# 3. shell action 调试：action 里加 echo
free-kiro hook add --id debug --event file.save --action-type shell --action "jq . /dev/stdin"
free-kiro hook run file.save --file test.go
```

## 边界与设计选择

- free-kiro **不主动触发 hook**——只在被 `hook run` 调用时执行。这是"被动规划层"的定位。
- shell 命令通过 `sh -c` 执行，遵守 POSIX shell 语法
- timeout 默认 30s；`0` 禁用超时（不推荐，可能 hang 死）
- agent action 默认给占位符，**不报错**——这样 IDE hook runner 不会因 agent fn 缺失而失败

---

参考：[CLI.md](CLI.md)（hook 命令）/ [WORKFLOW.md](WORKFLOW.md)（状态机集成）/ [COMPATIBILITY.md](COMPATIBILITY.md)（与 Kiro 信封互操作）