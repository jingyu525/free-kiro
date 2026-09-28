# free-kiro

> **规约式开发工作流引擎** —— Kiro Spec 工作流的免费、跨平台、单二进制克隆版。
> 把"先想清楚 → 再动手"做成强约束门禁，让 AI coding 工具的可能性空间收敛，而不是发散失控。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://go.dev/)

## 为什么做 free-kiro

[Kiro](https://kiro.dev) 是 AWS 出品的 AI coding IDE，核心价值是把
spec-driven workflow 做成强约束门禁。但 Kiro 官方是付费 + 闭源 +
IDE 强绑定产品。

Claude Code / CodeBuddy / Cursor / Continue 这些主流 AI coding 工具
默认没有规约式门禁——agent 一上手就开始改文件，可能性空间迅速发散，
复杂任务经常失控。

**free-kiro 把 Kiro 的核心 spec 工作流引擎单独抽出来，做成跨平台
单二进制 CLI**，通过 hook 系统接入任何支持 hook 的 AI coding 工具，
让"先写 spec → 过 lint → 再写代码"成为强制流程。

## 核心特性

| 特性 | 说明 |
|---|---|
| **spec 状态机** | `draft → requirements → design → tasks → approved → implementing → done` |
| **EARS 句式 lint** | 5 种 EARS 模板 + 无条件基线自动校验；缺漏直接拦截 |
| **bugfix 三段式** | Current / Expected / Unchanged Behavior 契约 |
| **wave 并行调度** | tasks.md 依赖图 → 拓扑分层 → 并行 wave |
| **steering 两级作用域** | workspace 覆盖 global；4 种 inclusion 模式 |
| **hook 信封互操作** | 写出兼容 Kiro 官方 v1 信封的 JSON，官方 Kiro IDE 可直接加载 |
| **drift 漂移检测** | approve 时锁定 baseline，编辑后自动暴露差异 |
| **advisory 一致性分析** | vague / 重复 AC / 可追溯性问题（不阻塞） |
| **单二进制分发** | `curl -fsSL … \| bash` 一行安装，跨平台（macOS / Linux / Windows） |

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/liujingyu/free-kiro/main/install.sh | bash
```

或从 [GitHub Releases](https://github.com/liujingyu/free-kiro/releases)
下载对应平台的二进制。

## 30 秒上手

```bash
cd your-project
free-kiro init                                          # 初始化 .kiro/
free-kiro spec new user-auth --prompt "add login"       # 建 spec
free-kiro spec generate user-auth --phase all           # 生成 3 份文档
# 编辑 requirements.md / design.md / tasks.md 填真实内容
free-kiro lint user-auth                                # 过 lint 门禁
free-kiro spec approve user-auth                         # 审批 + 锁定 baseline
free-kiro spec start user-auth                          # 开始实现
# ...实现完成后...
free-kiro spec complete user-auth                        # 收尾
```

## 接入 AI coding 工具（以 Claude Code 为例）

写到 `~/.claude/settings.json`：

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
  }, {
    "name": "kiro-session-start",
    "trigger": "SessionStart",
    "action": {
      "type": "command",
      "command": "free-kiro spec next $(free-kiro spec list --format=json | jq -r '.[0].name')"
    }
  }]
}
```

效果：每次 Edit/Write 前自动 lint，发现 ERROR 直接拦截写入；会话开始
时自动锁定当前 spec 阶段。

## 命令速查

```
free-kiro init                                    # 初始化工作区
free-kiro spec new <name> --prompt "…"           # 建 spec
free-kiro spec quick <name> --prompt "…"         # Quick Spec（免审批）
free-kiro spec generate <name> --phase all       # 生成 3 份文档
free-kiro spec approve <name>                    # 审批 + 捕获 baseline
free-kiro spec start <name>                      # 标记开始实现
free-kiro spec complete <name>                   # 标记完成
free-kiro spec status <name>                     # 状态 + drift（JSON）
free-kiro spec next <name>                       # 预言机：下一步动作
free-kiro spec analyze <name>                    # advisory 一致性分析
free-kiro task list <name>                       # 并行 wave 视图
free-kiro steering {list,show,context}           # 项目约定文档
free-kiro hook {list,add,run}                    # 事件驱动 hook
free-kiro lint [<name>]                          # 离线质量门禁
```

完整命令参考见 [docs/CLI.md](docs/CLI.md)。

## 退出码契约

| 退出码 | 含义 |
|---|---|
| 0 | 成功 |
| 1 | lint ERROR（可被 IDE PreToolUse 拦截，配合 `\|\| exit 2`） |
| 2 | 引擎错误（workspace 缺失、非法 phase 转移等） |
| 3 | 用户输入错误 |

## 与 kiro-clone / 官方 Kiro 的兼容性

- ✅ spec 文档格式（requirements.md / design.md / tasks.md / bugfix.md）
- ✅ `.kiro/` 目录结构（specs / steering / hooks / settings.json）
- ✅ Hook JSON 信封（写出的 JSON 可被官方 Kiro IDE 直接加载）
- ✅ EARS 句式（lint 结果与 kiro-clone 一致）

不兼容：
- 不做 AI 模型调用（spec 文档是骨架，由 agent 写内容）
- 不做 IDE 集成（hook 系统是 IDE 的职责，free-kiro 只暴露 CLI 入口）
- 不做 agent 主动跑 prompt（保持"被动的规划层"定位）

## 文档

- [docs/CLI.md](docs/CLI.md) — 完整命令参考
- [docs/WORKFLOW.md](docs/WORKFLOW.md) — 工作流图解
- [docs/EARS.md](docs/EARS.md) — 验收标准句式
- [docs/STEERING.md](docs/STEERING.md) — Steering 文档
- [docs/HOOKS.md](docs/HOOKS.md) — Hooks 信封
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md) — 与 kiro-clone / 官方 Kiro 兼容性

## 致谢

参考实现：[kiro-clone](https://github.com/...)（Python，MIT）。
free-kiro 是其 1:1 Go 重写，专注于跨平台单二进制分发。

## License

MIT © 2026 liujingyu