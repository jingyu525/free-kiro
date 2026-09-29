# free-kiro

> **规约式开发工作流引擎** —— Kiro Spec 工作流的免费、跨平台、单二进制克隆版。
> 把"先想清楚 → 再动手"做成强约束门禁，让 AI coding 工具的可能性空间收敛，而不是发散失控。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/jingyu525/free-kiro)](https://github.com/jingyu525/free-kiro/releases)

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
| **hook 信封互操作** | 写出兼容 Claude Code event-keyed JSON，官方 Kiro IDE 可直接加载 |
| **drift 漂移检测** | approve 时锁定 baseline，编辑后自动暴露差异 |
| **advisory 一致性分析** | vague / 重复 AC / 可追溯性问题（不阻塞） |
| **可视化** | ASCII tree + Mermaid 图 + 本地 web dashboard（real-time SSE 推送） |
| **跨多 spec 状态** | `.kiro/.current` 自动标记活跃 spec；SessionStart hook 准度提升 |
| **issue → spec** | `free-kiro spec new --from-issue <url>` 一键从 GitHub issue 生成 spec |
| **PRD → spec** | `free-kiro spec new --from-prd <url>` 从任意网页（HTML/Markdown）拉取 PRD 内容 |
| **自升级** | `free-kiro upgrade` 下载最新 release + SHA256 校验 + re-exec |
| **零新外部依赖** | 全部用 Go 标准库 + `golang.org/x/net/html`（唯一新 dep） |

## 安装（三种方式任选）

### 1. Homebrew（macOS / Linux 推荐）

```bash
brew install jingyu525/tap/free-kiro
```

### 2. curl|bash 一行脚本（任何平台）

```bash
curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/install.sh | bash
```

脚本自动：检测平台、下载二进制、SHA256 校验、写入 `~/.local/bin`、追加 PATH 到 shell rc。

### 3. 从 GitHub Releases 下载

打开 [github.com/jingyu525/free-kiro/releases](https://github.com/jingyu525/free-kiro/releases)，
下载对应平台的 tarball，解压即可。

### 4. go install（需要 Go 1.22+）

```bash
go install github.com/jingyu525/free-kiro/cmd/free-kiro@latest
```

### CI 集成

```yaml
- uses: jingyu525/free-kiro/.github/actions/setup-free-kiro@v0
  with: { version: v0.6.0 }
- run: free-kiro lint
- run: free-kiro doctor --strict
```

## 30 秒上手

```bash
cd your-project
free-kiro init --ide claude-code            # 初始化 + 自动配置 IDE hook + 写 AGENTS.md
free-kiro doctor                            # 一键诊断
free-kiro spec quick demo --prompt "add login"  # Quick Spec：一次性生成 + 免审批
# 编辑 requirements.md / design.md / tasks.md 填真实内容
free-kiro lint demo                         # 过 lint 门禁（错误末尾附 docs 链接）
free-kiro spec start demo                   # 开始实现
free-kiro spec status demo                  # 人类可读状态（drift 提顶层）
free-kiro task list demo                    # 并行 wave 视图
free-kiro spec show demo --tree             # ASCII 树
free-kiro spec show demo --graph | pbcopy   # Mermaid 图（贴 GitHub 自动渲染）
free-kiro report                            # 完整 markdown 报告 → .kiro/REPORT.md
free-kiro serve                             # 启动本地 web dashboard
free-kiro spec complete demo                # 收尾
```

## 杀手路径：从 GitHub issue 一键生成 spec

```bash
# 从 issue 拉 title + body 自动生成 spec
free-kiro spec new --from-issue https://github.com/owner/repo/issues/42
# → 自动创建 spec "fix-login-bug"，prompt 是 issue 标题 + 正文

# 简写
free-kiro spec new --from-issue owner/repo#42
```

## 杀手路径：从 PRD URL 一键生成 spec

```bash
# 任意网页（Notion / Confluence / 公司 wiki / Google Docs）
free-kiro spec new --from-prd https://www.notion.so/your-team/PRD-Login-Flow-abc123

# 自定义名字
free-kiro spec new login-v2 --from-prd https://confluence.company.com/x/login-v2
```

`--from-prd` 自动：
- HTTP GET 抓取页面（HTML / Markdown）
- 过滤 nav/footer/script/style 噪音
- 提取 `<title>` + 可见文本
- 1 MB body cap + 30s timeout
- 用 title 生成 kebab-case slug
- 输出 "created (from PRD)" 标签

## 接入 AI coding 工具（以 Claude Code 为例）

写到 `~/.claude/settings.json`（用 `free-kiro init --ide claude-code` 自动配置）：

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
      "name": "free-kiro-session-next",
      "trigger": "SessionStart",
      "action": {
        "type": "command",
        "command": "SPEC=$(cat .kiro/.current 2>/dev/null) && free-kiro spec next \"$SPEC\""
      }
    }
  ]
}
```

效果：每次 Edit/Write 前自动 lint，发现 ERROR 直接拦截写入；会话开始
时自动锁定当前 spec 阶段。`exit 2` 才能拦截（普通 `exit 1` 会被 IDE 忽略）。

## 命令速查

```
# 工作流
free-kiro init [--ide auto|claude-code|codebuddy|none]    初始化 + IDE hook
free-kiro doctor [--strict]                              一键诊断
free-kiro upgrade [--check|--force]                      自升级

# Spec 生命周期
free-kiro spec new <name> --prompt "…"                   新建（基本）
free-kiro spec new --from-issue <url>                    从 GitHub issue
free-kiro spec new --from-prd <url>                      从 PRD URL（HTML/MD）
free-kiro spec quick <name> --prompt "…"                 Quick Spec（免审批）
free-kiro spec generate <name> --phase all               生成 3 份文档
free-kiro spec approve <name>                            审批 + 捕获 baseline
free-kiro spec start <name>                              标记开始实现
free-kiro spec complete <name>                           标记完成

# 可视化
free-kiro spec show <name> --tree                        ASCII 树
free-kiro spec show <name>                                原始文档
free-kiro spec status <name> [--json|--graph]            状态（人类可读 / JSON / Mermaid）
free-kiro spec list                                       spec 列表 + 活跃标记
free-kiro spec next <name>                                下一步动作预言机
free-kiro spec analyze <name>                             advisory 一致性分析
free-kiro task list <name>                                并行 wave 视图
free-kiro report                                          完整 markdown 报告

# 项目上下文
free-kiro steering {list,show,context}                   steering 文档
free-kiro hook {list,add,run}                             事件驱动 hook
free-kiro lint [<name>]                                  离线质量门禁

# Dashboard
free-kiro serve [--bind 127.0.0.1] [--port 7373]         本地 web dashboard
```

完整命令参考见 [docs/CLI.md](docs/CLI.md)。

## 退出码契约

| 退出码 | 含义 | IDE hook 用法 |
|---|---|---|
| 0 | 成功 | 命令完成 |
| 1 | lint ERROR | 配合 `\|\| exit 2` 在 PreToolUse 中拦截写入 |
| 2 | 引擎错误 | workspace 缺失 / 非法 phase 转移 / IO 等 |
| 3 | 用户输入错误 | 缺参数 / 名称冲突 |

## 与官方 Kiro 的兼容性

| 维度 | 官方 Kiro IDE | free-kiro |
|---|---|---|
| spec 文档格式（requirements.md / design.md / tasks.md / bugfix.md） | ✅ | ✅ 100% |
| `.kiro/` 目录结构（specs / steering / hooks / settings.json） | ✅ | ✅ 100% |
| Hook JSON 信封 | v1 envelope | 兼容 Claude Code event-keyed schema |
| EARS 5 种句式 + 无条件基线 | ✅ | ✅ lint 结果一致 |
| 跨平台分发 | IDE 安装 | 单二进制（5 平台）+ Homebrew |

不兼容：`dual-kiro` 不做 AI 模型调用、不做 IDE 集成、不做 agent 主动跑 prompt——保持"被动的规划层"定位。

## 跨平台分发

| 平台 | 安装方式 | 大小 |
|---|---|---|
| macOS (Intel / Apple Silicon) | `brew install` 或 `tar.gz` | ~4 MB |
| Linux (amd64 / arm64) | `curl\|bash` / `tar.gz` / `apt`（未来） | ~4 MB |
| Windows (amd64) | `tar.gz` | ~4 MB |

每个 GitHub Release 附带 5 个 tarball + SHA256SUMS。

## 项目生态

| 入口 | 用途 |
|---|---|
| [jingyu525/free-kiro](https://github.com/jingyu525/free-kiro) | 主仓库 |
| [jingyu525/free-kiro Releases](https://github.com/jingyu525/free-kiro/releases) | 跨平台二进制下载 |
| [jingyu525/homebrew-tap](https://github.com/jingyu525/homebrew-tap) | Homebrew formula |
| `setup-free-kiro` GitHub Action | CI 集成 |

## 文档

- [docs/CLI.md](docs/CLI.md) — 完整命令参考
- [docs/WORKFLOW.md](docs/WORKFLOW.md) — 工作流图解（状态机 + lint gate）
- [docs/EARS.md](docs/EARS.md) — EARS 验收标准句式
- [docs/STEERING.md](docs/STEERING.md) — Steering 文档（4 种 inclusion 模式）
- [docs/HOOKS.md](docs/HOOKS.md) — Hooks 信封 + Claude Code 配置
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md) — 与官方 Kiro IDE 兼容性

## License

MIT © 2026 jingyu525