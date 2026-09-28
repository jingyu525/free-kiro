# 与官方 Kiro 的兼容性

> free-kiro 设计目标：**完全兼容官方 Kiro 的 spec 工作流**，**完全兼容 Kiro 官方
> v1 hook 信封**。你可以从官方 Kiro IDE 迁移到 free-kiro，所有 spec 文档、steering
> 配置、hook 文件**直接复用**，无需任何转换。

## 兼容性矩阵

| 维度 | 官方 Kiro IDE | free-kiro (Go) | 兼容性 |
|---|---|---|---|
| `.kiro/` 目录结构 | ✅ | ✅ | 100% |
| `specs/<name>/{requirements\|design\|tasks\|bugfix}.md` | ✅ | ✅ | 100% |
| `steering/*.md` | ✅ | ✅ | 100% |
| `hooks/*.json`（v1 envelope） | ✅ | ✅（同时接受 flat shape） | 100% |
| EARS 5 种句式 + 无条件基线 | ✅ | ✅ | 100% |
| Bugfix 三段式 | ✅ | ✅ | 100% |
| Spec 状态机 | ✅ | ✅ | 100% |
| Wave 并行调度（依赖图拓扑） | ✅ | ✅ | 100% |
| Drift baseline 锁定 | ✅ | ✅ | 100% |
| Analyze advisory | ✅ | ✅ | 100% |
| 退出码契约（lint 1 / 引擎 2 / 用法 3） | n/a | ✅ | n/a |
| IDE PreToolUse hook 拦截 | ✅ | ✅（同样的 `\|\| exit 2`） | n/a |

## 不兼容的部分（free-kiro 不做）

- ❌ **AI 模型调用** — spec 文档是骨架，由 agent 写内容。free-kiro 是被动的规划层。
- ❌ **IDE 集成 / 编辑器插件** — hook 系统是 IDE 的职责，free-kiro 只暴露 CLI。
- ❌ **TUI 交互** — CLI only，避免分心。
- ❌ **自动驱动 hook**（agent 主动响应 IDE 事件） — free-kiro 保持被动，只在被 `hook run` 调用时执行。
- ❌ **模型驱动的 spec 自动生成** — free-kiro 保持 offline / template only。

## free-kiro 相对官方 Kiro IDE 的差异化优势

| 维度 | 官方 Kiro IDE | free-kiro |
|---|---|---|
| 启动时间 | ~150ms（IDE 集成） | ~3ms（Go 编译产物） |
| 分发 | IDE 安装（仅 Kiro IDE 用户） | 单二进制 `curl\|bash` 一行安装 |
| 跨平台 | IDE 内嵌 | darwin/linux/windows × amd64/arm64 原生二进制 |
| Hook 信封 | v1 envelope | 同时接受 flat + v1 envelope |
| 退出码 | n/a | 0/1/2/3 区分（lint / 引擎 / 用法） |
| 定价 | AWS 付费 | 免费 + MIT |

## hook 文件迁移

### 官方 Kiro → free-kiro

官方 Kiro 的 hook 文件通常长这样：

```json
{
  "version": "v1",
  "hooks": [
    {
      "name": "kiro-lint",
      "trigger": "PostFileSave",
      "matcher": "\\.tsx$",
      "action": {
        "type": "command",
        "command": "npx eslint --fix"
      }
    }
  ]
}
```

`free-kiro hook list` 会自动归一化读取，显示：

```
ID                   EVENT            FILTER        TYPE    TO    ENABLED
kiro-lint            PostFileSave     \.tsx$ (re)   shell   -     true
```

注意官方 Kiro 的 `matcher`（regex）和 `type: "command"` 都被正确归一化。

### free-kiro → 官方 Kiro

free-kiro 写出的 hook 自动符合 Kiro v1 信封。直接复制 `.kiro/hooks/<id>.json` 到
官方 Kiro IDE 的 hooks 目录即可加载。

或者把整个 `.kiro/hooks/` 目录拷过去。

## spec 文档格式：完全兼容

官方 Kiro IDE 生成的 spec 目录，free-kiro 可以直接读：

```bash
# 官方 Kiro IDE 创建的 spec：
ls .kiro/specs/demo/
# → .meta.json  requirements.md  design.md  tasks.md

# 用 free-kiro 接管：
cd same-project
free-kiro lint demo                    # 同样的 lint 报告
free-kiro spec approve demo            # 同样的 approve 流程
free-kiro spec status demo             # JSON 输出格式
free-kiro spec sync demo               # 重新基线化
```

反之亦然——free-kiro 生成的 spec 用官方 Kiro IDE 也能管理。

## steering 文档：完全兼容

free-kiro 与官方 Kiro 都接受 `inclusion:` 键（Kiro 官方风格）。

当文件同时存在 `mode:` 与 `inclusion:` 时，`inclusion:` 优先（Kiro 官方优先）。

也支持 Kiro 的 `fileMatchPattern` glob 语法（`**` / `*` / `?`）。

## 退出码契约

| 退出码 | 含义 | IDE hook 用法 |
|---|---|---|
| 0 | 成功 | 命令完成 |
| 1 | lint ERROR | 配合 `\|\| exit 2` 在 PreToolUse 中拦截写入 |
| 2 | 引擎错误 | workspace 缺失 / 非法 phase 转移 / IO 等 |
| 3 | 用户输入错误 | 缺参数 / 名称冲突 |

**关键**：Claude Code / CodeBuddy 的 `PreToolUse` hook 必须 `exit 2` 才拦截写入，
所以配置是：

```bash
free-kiro lint || exit 2
```

## 从官方 Kiro IDE 迁移到 free-kiro

1. 安装 free-kiro：`curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/install.sh | bash`
2. 在项目目录跑 `free-kiro init`（如果之前已经有 `.kiro/`，是幂等的）
3. 现有 spec 文档、steering、hook 全部不动——free-kiro 直接接管
4. 如果你之前用官方 Kiro IDE 的 hook 配置，**JSON 直接复用**，free-kiro 自动归一化读取
5. CI hook 把 IDE 集成命令替换成 `free-kiro lint || exit 2` 即可

## 与官方 Kiro IDE 共存

free-kiro 不与官方 Kiro IDE 冲突——它们共享 hook JSON 格式：

```bash
# 项目里有 .kiro/，可以被两者同时打开
# spec 文档两边都能读
# hook 文件格式相同，复制即可
```

## 已知差异（小细节）

| 细节 | 官方 Kiro IDE | free-kiro | 影响 |
|---|---|---|---|
| 默认 generator | template | template | 无差异 |
| `prompt` 元数据存储 | ✅ | ✅ | 无差异 |
| `created_at` 时间格式 | RFC3339 UTC | RFC3339 UTC | 无差异 |
| `.meta.json` 缩进 | 2 空格 | 2 空格 | 无差异 |
| EARS 大小写 | 大小写不敏感 | 大小写不敏感 | 无差异 |
| `tasks.md` 行格式 | `- [ ] #N Title [deps: #N1,#N2]` | 同 | 无差异 |

## 总结

free-kiro 是官方 Kiro Spec 工作流引擎的跨平台免费实现。spec 工作流的所有
概念（状态机 / EARS / 漂移 / wave / steering / hook）在两者之间完全可移植。

唯一的差异是：**单一职责 CLI** + **跨平台分发** + **完全免费 + MIT**。

---

参考：[WORKFLOW.md](WORKFLOW.md)（状态机）/ [EARS.md](EARS.md)（句式）/ [STEERING.md](STEERING.md)（约定）/ [HOOKS.md](HOOKS.md)（信封）