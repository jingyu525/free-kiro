# hooks — 事件驱动钩子

free-kiro 有两层 hook：

1. **IDE hook**：写到 `~/.claude/settings.json` 等，由 IDE（Claude Code / CodeBuddy）触发，调 `free-kiro` 二进制
2. **项目 hook**：写到 `.kiro/hooks/<id>.json`，由 `free-kiro hook run <event>` 调度，可被任意脚本触发

## `free-kiro init` 默认装的两条 IDE hook

```jsonc
// ~/.claude/settings.json (excerpt)
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "# free-kiro-managed: lint-gate: free-kiro lint || exit 2"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "# free-kiro-managed: session-next: free-kiro spec next $(cat .kiro/.current 2>/dev/null) || free-kiro spec list 2>/dev/null | awk 'NR==2{print $1}'"
          }
        ]
      }
    ]
  }
}
```

- `PreToolUse` 拦截 Edit/Write，写入前跑 `free-kiro lint`；非零退出（lint ERROR）→ 拦截
- `SessionStart` 在会话开始时引导到当前 spec（`.kiro/.current` 指向的）

## 事件命名（两套命名同时接受）

free-kiro 写出的 hook 同时兼容自家 flat shape 与 Kiro 官方 v1 信封，
可直接被官方 Kiro IDE 加载。

### free-kiro 内部命名

| 事件 | 触发时机 |
|---|---|
| `file.save` | 文件被保存 |
| `file.create` | 文件被创建 |
| `file.delete` | 文件被删除 |
| `prompt.submit` | 用户提交 prompt |
| `task.run` | task 节点跑完 |
| `manual` | 手动 `free-kiro hook run` 触发 |

### Kiro 官方 v1 信封

| 事件 | 触发时机 |
|---|---|
| `SessionStart` | 会话开始 |
| `UserPromptSubmit` | 用户提交 prompt |
| `PreToolUse` | 工具调用前 |
| `PostToolUse` | 工具调用后 |
| `PostFileSave` | 文件保存后 |
| `PostFileCreate` | 文件创建后 |
| `Stop` | 会话停止 |

定位：free-kiro 是被动的规划层——它从不主动触发 hook，只在被调用
`hook run <event>` 时执行匹配的动作。

## 项目 hook（`.kiro/hooks/*.json`）

```bash
free-kiro hook add \
  --id notify-on-spec-done \
  --event spec.complete \
  --action-type shell \
  --action 'curl -X POST https://hooks.slack.com/...'

free-kiro hook list

# 手动触发（调试）
free-kiro hook run spec.complete --file .kiro/specs/foo/requirements.md
```

### Action 类型

- `shell`：`sh -c <action>`，环境变量 `FREE_KIRO_EVENT` / `FREE_KIRO_FILE` / `FREE_KIRO_CWD` 已注入
- `agent`（占位）：未来版本支持，agent_fn 暂未实现

### 自由命名约定

- IDE 标准事件：见上方"Kiro 官方 v1 信封"段（7 个）
- free-kiro 内部事件：`spec.complete` / `spec.approved` / `lint.error`（命名空间 `spec.` 与 `lint.`）

## 写多 hook / 改 hook

```bash
# 覆盖：直接 add 同 id 即替换
free-kiro hook add --id foo --event spec.complete --action-type shell --action 'echo new'

# 禁用：编辑 .kiro/hooks/foo.json，把 action.type 改为 "disabled"
# 删除：直接删文件
```

## 排错

| 症状 | 原因 |
|---|---|
| `free-kiro hook run` 报 `no hook matched` | event 名错或 matcher 不匹配 |
| IDE hook 不触发 | `~/.claude/settings.json` 里 free-kiro-managed 的 marker 被改过，init 会重新覆盖 |
| shell action 不跑 | `--action` 用单引号包整体；调试用 `free-kiro hook run --verbose <event>` |