# DEMO_VERIFICATION — 5 分钟 demo 复盘与验证记录

本文档配套 [`top1-demo-onboarding`](../.kiro/specs/top1-demo-onboarding/)。
它有两件事:
1. **§1 验证记录** —— 记录每次 demo 端到端跑通的命令 + 退出码 + 真实输出,作为将来回放依据。
2. **§2 迁移审计** —— README 重写前的"30 秒上手"原内容快照,作为内容审计依据。

任何 reviewer 想复盘 demo:按 §1 命令顺序跑即可看到完整闭环。

---

## §1 验证记录

### §1.1 examples/todo-app 真实样例(写于 2026-09-30)

**目标**:验证 `examples/todo-app/` 是一个 lint 全绿 / 多 wave 拓扑 / 可直接 clone 后跑 `free-kiro demo` 的真实样例。

#### 步骤 1 — 创建 workspace

```bash
$ mkdir -p examples/todo-app
$ cd examples/todo-app
$ free-kiro init --ide none --path .
已初始化 free-kiro 工作区于 /Users/liujingyu/CodeBuddy/free-kiro/examples/todo-app/.kiro
已写入示例 steering：product.md、structure.md、tech.md
已写入 /Users/liujingyu/CodeBuddy/free-kiro/examples/todo-app/.kiro/AGENTS.md
已跳过 IDE 配置（--ide none）

# 退出码: 0
```

#### 步骤 2 — 生成 spec 骨架

```bash
$ free-kiro spec quick add-task-priority --prompt "todo app priority sorting"
quick spec "add-task-priority" generated (requirements/design/tasks); approval waived
next: free-kiro spec start add-task-priority   (or review, then free-kiro spec approve add-task-priority)

# 退出码: 0
```

随后手工精修 `requirements.md` / `design.md` / `tasks.md`,把 `<TODO:…>`
占位符替换为真实 EARS AC + 任务(详见 spec 三份文档)。

#### 步骤 3 — lint 验证(在 examples/todo-app 内)

```bash
$ cd examples/todo-app
$ free-kiro lint
add-task-priority: OK

# 退出码: 0
```

> **UX 注意**:`free-kiro lint examples/todo-app` 在主项目根跑会失败
> (free-kiro 把参数当 spec 名称)。演示给用户时务必说清"必须先 cd
> 到样例目录内"或通过 `free-kiro demo` 自动跳转。

#### 步骤 4 — task list 看 wave 拓扑

```bash
$ cd examples/todo-app
$ free-kiro task list add-task-priority
Wave 1:
  [ ] #1 在 `examples/todo-app/store.go` 实现 `Task` 结构体 ...  (deps: -)
  [ ] #2 在 `examples/todo-app/main.go` 用 cobra 搭 CLI 框架 ...  (deps: -)
Wave 2:
  [ ] #3 在 `examples/todo-app/sort.go` 实现 `SortTasks(...)` ...  (deps: #1)
Wave 3:
  [ ] #4 在 `examples/todo-app/main_test.go` 用 `cmd.SetArgs` + ...  (deps: #1,#2,#3)
Wave 4:
  [ ] #5 跑 `cd examples/todo-app && go build ./...` + `go test ./...` ...  (deps: #4)

add-task-priority: 0/5 done, 4 wave(s)

# 退出码: 0
```

✅ 满足 requirements.md AC:"≥2 wave";Wave 1 含 2 个并行任务(#1 + #2),
满足"至少一个 wave 有 2+ 任务"。

#### 步骤 5 — approve + drift 演示

```bash
$ cd examples/todo-app
$ free-kiro spec approve add-task-priority
approved spec "add-task-priority"; baseline captured for drift detection

# 退出码: 0
```

随后**故意**改一行 `requirements.md`(把 `WHEN the user runs `todo add`
...` 改成 `WHEN the user runs `todo create` ...`)演示 drift:

```bash
$ free-kiro spec status add-task-priority
add-task-priority
  phase:      approved
  workflow:   requirements-first
  spec_type:  feature
  approved:   yes
  tasks:      0/5 done, 4 wave(s)
  drift:      drift      ← 关键:从 none → drift
  ...
```

演示后用 `git checkout .kiro/specs/add-task-priority/requirements.md`
恢复 baseline,再次 `free-kiro spec status` 应回到 `drift: none`。

#### 步骤 6 — 全量门禁(主项目)

```bash
$ cd /Users/liujingyu/CodeBuddy/free-kiro
$ free-kiro lint
top1-demo-onboarding: OK

# 退出码: 0
$ free-kiro doctor
... 7 项全过 ...
```

---

## §2 README "30 秒上手"原内容快照(迁移审计)

> 本节由 `top1-demo-onboarding` 的 task #4 在重写 README 时同步填充。
> 保留原"30 秒上手"完整 markdown 文本,作为内容审计证据。

### §2.1 重写前原文(commit `131235e`)

> **Section header**: `## 30 秒上手`
>
> **Body**:
>
> ```bash
> cd your-project
> free-kiro init --ide claude-code            # 初始化 + 自动配置 IDE hook + 写 AGENTS.md
> free-kiro doctor                            # 一键诊断
> free-kiro spec quick demo --prompt "add login"  # Quick Spec：一次性生成 + 免审批
> # 编辑 requirements.md / design.md / tasks.md 填真实内容
> free-kiro lint demo                         # 过 lint 门禁（错误末尾附 docs 链接）
> free-kiro spec start demo                   # 开始实现
> free-kiro spec status demo                  # 人类可读状态（drift 提顶层）
> free-kiro task list demo                    # 并行 wave 视图
> free-kiro spec show demo --tree             # ASCII 树
> free-kiro spec show demo --graph | pbcopy   # Mermaid 图（贴 GitHub 自动渲染）
> free-kiro report                            # 完整 markdown 报告 → .kiro/REPORT.md
> free-kiro serve                             # 启动本地 web dashboard
> free-kiro watch --preset reactive           # 实时监听：lint + status
> free-kiro spec complete demo                # 收尾
> ```

### §2.2 重写后(2026-09-30,本次 spec 实施)

原"30 秒上手"内容**保留**在 README(不动),仅在 section header 上方
加引导:

> ⚠️ 推荐新用户从上面的 [5 分钟端到端 demo](#5-分钟端到端-demo) 开始。本节保留给
> 已有项目想接入 free-kiro 的老用户;完整迁移审计见
> [`docs/DEMO_VERIFICATION.md` §2](docs/DEMO_VERIFICATION.md)。

新增的"5 分钟端到端 demo" section 在 `## 为什么做 free-kiro` 之后立即
出现,内容涵盖:

- 一行命令 `go run ./cmd/free-kiro demo` 跑通端到端 demo
- `demo` 子命令做什么(cwd 校验、marker 写入、5 步摘要、可选 hook 片段)
- 后续跑 `serve` 看板、浏览器打开、看 drift 的 4 行命令
- 验证完整闭环指向 [`docs/DEMO_VERIFICATION.md` §1](#§1-验证记录)

### §2.3 设计决策记录

**为什么不动原"30 秒上手"**:

1. 它仍然对"已有项目想接入 free-kiro"的用户有用(quick spec + 编辑 + 实现的传统路径)
3. 删除会破坏老用户的肌肉记忆;让他们在 README 找不到熟悉的命令

**为什么新增"5 分钟端到端 demo"在它前面**:

- README 顶部是 README 的 hot zone,90% 的访客只读前 200 行
- 新访客的核心需求是"快判断这个项目值不值得投入",5 分钟 demo 是判断路径
- 已有项目接入是 secondary use case,放 secondary 位置合理

**为什么不彻底重写**:

- 重写有丢失原文的风险(命令参数、链接、边界 case)
- "保留 + 在头部加 demo 路径"是最低风险 + 最高信息密度的折中
- §2.1 保留原文快照,任何 reviewer 想比对前后内容都有依据