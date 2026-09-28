# CLI 参考

> free-kiro 的完整命令参考。每个命令都列出了参数、行为、与退出码的契约。
> 不熟悉的概念（如 EARS、Spec 状态机）请先看 [WORKFLOW.md](WORKFLOW.md) 和 [EARS.md](EARS.md)。

## 顶层结构

```
free-kiro
├── init                              初始化 .kiro 工作区
├── spec                              管理 spec
│   ├── new <name>                    新建 spec
│   ├── generate <name>               生成 planning 文档
│   ├── quick <name>                  Quick Spec（免审批）
│   ├── show <name>                   打印某个 phase 的文档
│   ├── list                          列出全部 specs
│   ├── approve <name>                审批 + 捕获 baseline
│   ├── start <name>                  标记开始实现
│   ├── complete <name>               标记完成
│   ├── status <name>                 状态 + 漂移（JSON）
│   ├── next <name>                   预言机：下一步动作（JSON）
│   ├── sync <name>                   重新基线化
│   └── analyze <name>                advisory 一致性分析
├── steering                          项目约定文档
│   ├── list                          列出全部
│   ├── show <name>                   打印单个
│   └── context                       组装上下文（agent 注入用）
├── task list <spec>                  并行 wave 视图
├── hook                              事件驱动 hook
│   ├── list                          列出全部 hook
│   ├── add --id <id> ...             新增 hook
│   └── run <event>                   触发匹配 hook
└── lint [<spec>]                     离线质量门禁
```

## `init`

初始化 `.kiro/` 工作区 + 写入示例 steering 文档（product.md / structure.md / tech.md）。

```bash
free-kiro init --path <dir>     # 默认: .
```

幂等：重复运行不会覆盖已有文件。

## `spec new`

创建新 spec。

```bash
free-kiro spec new <name> --prompt "<text>" [--workflow requirements-first|design-first] [--type feature|bugfix] [--quick]
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `name` | ✅ | spec 名称（小写短横线） |
| `--prompt` | ✅ | 一句话需求（也可通过 stdin 传入） |
| `--workflow` |  | `requirements-first`（默认）or `design-first` |
| `--type` |  | `feature`（默认）or `bugfix` |
| `--quick` |  | Quick Spec 变体（一次性生成 + 免审批） |

下一步：`free-kiro spec generate <name> --phase all`

## `spec generate`

生成 planning 文档。默认生成全部（按 workflow 顺序）；可指定单 phase。

```bash
free-kiro spec generate <name> [--phase requirements|design|tasks|all] [--force]
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `name` | ✅ | spec 名称 |
| `--phase` |  | 默认 `all`；可选 `requirements` / `design` / `tasks` |
| `--force` |  | 覆盖已存在的文档（默认不覆盖，保护作者已写内容） |

## `spec quick`

Quick Spec 变体：新建 spec + 一次性生成 + 免审批仪式。

```bash
free-kiro spec quick <name> --prompt "<text>" [--type feature|bugfix]
```

适合小改动 / 已胸有成竹的变更。大功能仍走正式 `spec new` → `spec approve`。

## `spec show`

```bash
free-kiro spec show <name> --phase requirements|design|tasks|all
```

`--phase all` 一次打印三份文档。文档不存在时显示 `(not generated)`。

## `spec list`

```bash
free-kiro spec list
```

固定宽度表格：`NAME / PHASE / APPROVED`。

## `spec approve`

```bash
free-kiro spec approve <name>
```

lint gate 强约束：spec 文档存在 ERROR（`no-ears` / `placeholder-ac` / tasks 环/悬挂/自引用）时直接拒绝。

通过后捕获 baseline（`ac_count` / `task_count` / `design_sections`），后续编辑可被 drift 检测到。

## `spec start` / `spec complete`

纯记账命令（free-kiro 不运行代码）：

```bash
free-kiro spec start <name>       # APPROVED → IMPLEMENTING
free-kiro spec complete <name>    # IMPLEMENTING → DONE
```

## `spec status`

```bash
free-kiro spec status <name>
```

返回 JSON：

```json
{
  "name": "demo",
  "phase": "implementing",
  "workflow": "requirements-first",
  "spec_type": "feature",
  "approved": true,
  "baseline": { "ac_count": 3, "task_count": 3, "design_sections": 6 },
  "current": { "ac_count": 3, "task_count": 3, "design_sections": 6 },
  "drift": [],
  "tasks": { "done": 0, "total": 3, "waves": 3 }
}
```

`drift` 非空时说明编辑后已偏离 baseline，需 `spec sync` 重置或修复。

## `spec next`

```bash
free-kiro spec next <name>
```

返回 JSON 预言机：

```json
{
  "spec": "demo",
  "phase": "implementing",
  "suggested": "complete-when-done",
  "command": "free-kiro spec complete demo"
}
```

每个回合开头跑一次防断线：

```bash
free-kiro spec next $(free-kiro spec list | head -1 | awk '{print $1}')
```

## `spec sync`

```bash
free-kiro spec sync <name>
```

合法编辑 AC / task 列表后，重新捕获 baseline，消除漂移。

## `spec analyze`

advisory 一致性分析（永不阻塞）：

```bash
free-kiro spec analyze <name> [--json]
```

检查三类问题：
- `vague-language` — 模糊措辞（etc / and/or / maybe / user-friendly …）
- `duplicate-acceptance-criteria` — 重复 EARS 行
- `uncovered-acceptance-criteria` — 有 AC 但无 tasks
- `tasks-without-requirements` — 有 tasks 但无 AC

## `steering {list,show,context}`

```bash
free-kiro steering list
free-kiro steering show product
free-kiro steering context [--file <path>] [--prompt <text>]
```

`context` 是 agent 集成入口——每个生成请求前调用，按 mode 规则组装 always / auto / filematch / manual 文档。

## `task list`

```bash
free-kiro task list <spec>
```

按依赖图拓扑分层输出并行 wave：

```
Wave 1:
  [ ] #1 Set up module layout  (deps: -)
Wave 2:
  [ ] #2 Implement core [deps: #1]  (deps: #1)
Wave 3:
  [ ] #4 Polish [deps: #2,#3]  (deps: #2,#3)

demo: 0/3 done, 3 wave(s)
```

## `hook {list,add,run}`

```bash
free-kiro hook list
free-kiro hook add --id <id> --event <event> \
                   --action-type shell|agent --action "<cmd|prompt>" \
                   [--glob <pattern>] [--description <text>] [--timeout <sec>]
free-kiro hook run <event> [--file <path>]
```

add 写出的 JSON 兼容 Kiro v1 信封，官方 Kiro IDE 可直接加载。

事件名兼容两套命名：
- **free-kiro**：`file.save` / `file.create` / `file.delete` / `prompt.submit` / `task.run` / `manual`
- **Kiro 官方**：`SessionStart` / `UserPromptSubmit` / `PreToolUse` / `PostToolUse` / `PostFileSave` / `PostFileCreate` / `PreTaskExecution` / `PostTaskExecution`

## `lint [<spec>]`

```bash
free-kiro lint           # lint 全部 specs
free-kiro lint my-spec    # 只 lint my-spec
```

退出码契约：

| 退出码 | 含义 |
|---|---|
| 0 | 全部 OK |
| 1 | 发现 ERROR（no-ears / placeholder-ac / tasks 环/悬挂/自引用 …） |
| 2 | 引擎错误（workspace 不存在、spec 不存在 等） |

**IDE hook 配置示例**（写到 `~/.claude/settings.json`）：

```json
{
  "hooks": [{
    "name": "kiro-lint-gate",
    "trigger": "PreToolUse",
    "matcher": "Edit|Write",
    "action": {
      "type": "command",
      "command": "free-kiro lint || exit 2"
    }
  }]
}
```

`exit 2` 才能让 Claude Code / CodeBuddy 在 PreToolUse 阶段真正拦截写入。

---

参考：[WORKFLOW.md](WORKFLOW.md)（状态机图解）/ [EARS.md](EARS.md)（句式）/ [HOOKS.md](HOOKS.md)（信封格式）