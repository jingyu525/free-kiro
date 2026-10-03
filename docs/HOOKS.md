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

## Agent instructions per IDE

`free-kiro init --ide <id>` 在 hook 之外还会写出一组**项目根指令文件**——这些是
IDE 的 agent loader 在每次会话开始时主动读取的"工作约定"文件。Hook 拦截的是
**能不能写**，指令文件承载的是**该不该写**。

> 为什么 hooks 不够？Hook 只在事件触发时跑，无法替代"会话开始时一次性注入
> 上下文"的机制。Claude Code / Cursor / Continue / OpenCode 各自有自己的
> 指令文件协议约定，互不读取 AGENTS.md / CLAUDE.md / .cursorrules 之外的
> 文件。free-kiro 写出每个 IDE 真正会读的位置。

### 每只写哪里（v0.8.0+）

| IDE | 项目根写入文件 | agent loader 读取位置 |
|---|---|---|
| Claude Code | `CLAUDE.md` | `CLAUDE.md`（项目根 + `~/.claude/CLAUDE.md`） |
| Cursor | `.cursorrules` + `.cursor/rules/free-kiro.md` | `.cursorrules`（legacy）或 `.cursor/rules/*.md`（模块化规则） |
| Continue | `.continuerules` + `.continue/rules/free-kiro.md` | `.continuerules`（legacy）或 `.continue/rules/*.md`（YAML frontmatter 触发） |
| OpenCode | `AGENTS.md` | `./AGENTS.md` → `./CLAUDE.md` → `~/.config/opencode/AGENTS.md` |
| CodeBuddy | `AGENTS.md` | 项目根 AGENTS.md（国内惯例；具体 loader 待官方文档核实） |

`.kiro/AGENTS.md` 仍然会被 `init` 写入，但仅供**workspace 级 steering store**
加载（`internal/steering/store.go`），与上面 IDE instruction 文件是两个独立概念。

### marker 块与 steering 自动注入

`init` 写入的 4 个模板（`instructions_zh/en.md` / `agents_zh/en.md`）以及它们生成
的 5 个项目根 IDE 指令文件，都自带以下 marker 块：

```markdown
<!-- free-kiro-managed:start -->
<!-- auto-generated by free-kiro steering inject; do not edit -->

<!-- free-kiro-managed:end -->
```

这块区域是 `free-kiro steering inject`（见 `docs/STEERING.md` §"自动注入到
IDE 指令文件"）的写入目标：每次运行 inject 会把 `.kiro/steering/*.md`
中 `mode: always` 的文档按字母序拼成一个 markdown 块，覆盖 marker 之间的
内容；marker 自身与之外的段落原样保留。

如果想关闭 steering 自动注入，只需删除 5 个项目根 IDE 指令文件里的 marker
块（连同中间内容）——之后 `steering inject` 会因为找不到 marker 而跳过该文件
（stderr 警告 + 退出码 +1）。

### Marker 与幂等性

所有 free-kiro 写的指令文件首行都是 `# free-kiro-managed:`。这个标记让
`free-kiro init` 在重跑时能识别自己写的文件（**只跳过自己写的、不会误伤用户
手写的同名文件**）。同时 doctor 用这个标记验证"指令文件是否到位"：

```bash
free-kiro doctor
# ...
✓ claude-code instruction files
    CLAUDE.md
✓ opencode instruction files
    AGENTS.md
```

### 覆盖现有文件

默认行为是**跳过已存在**（无论是否 free-kiro 写的）。要强制覆盖，传
`--overwrite-instructions`：

```bash
free-kiro init --ide claude-code --overwrite-instructions
```

`--overwrite-agents` 仍接受，但已废弃，触发一次会 stderr 警告并自动 forward
到 `--overwrite-instructions`（下个 minor 移除）。

### 添加新 IDE

在 `internal/ide/ide.go` 的 `instructionFiles` map 加一行；其余路径
（doctor 检查 + `init` 自动接入）自动跟随。模板按需加
`internal/ide/templates/instructions_<lang>.md`（非 OpenCode/CodeBuddy）或
`agents_<lang>.md`（OpenCode/CodeBuddy 复用）。

## 边界与设计选择

- free-kiro **不主动触发 hook**——只在被 `hook run` 调用时执行。这是"被动规划层"的定位。
- shell 命令通过 `sh -c` 执行，遵守 POSIX shell 语法
- timeout 默认 30s；`0` 禁用超时（不推荐，可能 hang 死）
- agent action 默认给占位符，**不报错**——这样 IDE hook runner 不会因 agent fn 缺失而失败

---

参考：[CLI.md](CLI.md)（hook 命令）/ [WORKFLOW.md](WORKFLOW.md)（状态机集成）/ [COMPATIBILITY.md](COMPATIBILITY.md)（与 Kiro 信封互操作）