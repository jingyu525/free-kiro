---
mode: always
description: free-kiro 产品定位、核心能力、目标用户与非目标
---

# Product

`free-kiro` 是一个**规约式开发工作流引擎**。它是 AWS Kiro IDE 核心
spec-driven 工作流的免费、跨平台、单二进制克隆版。

## 一句话定位

> 把"先想清楚 → 再动手"做成**强约束门禁**：agent 必须先把需求/设计/
> 任务写成 spec 文档、过 lint，才能写代码。

## 目标用户

使用 Claude Code / CodeBuddy / Cursor / Continue 等 AI coding 工具的
开发者。痛点是：agent 一上手就开始改文件，可能性空间迅速发散，复杂
任务经常失控。

## 核心能力（v0.8.0）

| 能力 | 说明 |
|---|---|
| spec 状态机 | 主流路径为线性 `draft → requirements → design → tasks → approved → implementing → done`；planning 阶段（requirements / design / tasks）可互转；APPROVED → TASKS 是合法回填；完整合法迁移表见 `internal/models/phase.go` 的 `allowed` 切片 |
| EARS lint | 5 种 EARS 触发句式（when / while / where / unless / if-then）+ 1 个 ubiquitous 基线，共 6 类正则分支（`internal/lint/ears.go` 的 `EARSRe` 含 6 个 alternation），缺漏即拦截 |
| bugfix 三段式 | Current / Expected / Unchanged Behavior 契约，附 `defect-uses-shall` 反向规则（缺陷绝不能 SHALL 化） |
| tasks wave | `tasks.md` 依赖图 → 拓扑排序 → 并行 wave 调度 |
| steering 两级作用域 | workspace 覆盖 global，4 种 inclusion 模式（always / auto / manual / filematch） |
| hook 信封 | 写出的 JSON 兼容 Kiro IDE event-keyed 格式 |
| drift 检测 | approve 时锁 baseline，编辑后自动暴露差异 |
| advisory | 4 类 finding 代码：`vague-language` / `duplicate-acceptance-criteria` / `uncovered-acceptance-criteria` / `tasks-without-requirements`（不阻塞；详见 `internal/spec/analyze.go:53-115`） |
| 可视化 CLI | ASCII tree + Mermaid |
| 可视化 Web | 本地 dashboard，fsnotify 实时推送 + ETag/304 缓存；fsnotify 在容器 / 网络 FS 不可用时降级为 2s mtime 轮询（`internal/visualize/server_sse.go` `watchChangesFallback`），端到端延迟可达 2s+ |
| Dashboard UI | spec 详情面板（hash deep-link + deps 链接 + sparkline + a11y）+ 4 个核心 widget（`stat-card` / `sparkline-cell` / `phase-badge` / `waves-progress`）+ E2E；另有 9 个辅助 widget 见 `internal/visualize/static/src/widgets/` |
| 跨多 spec | `.kiro/.current` 自动标记活跃 spec |
| issue→spec | `free-kiro spec new --from-issue <gh-url>`；前置依赖 `gh` CLI（`brew install gh` + `gh auth login`） |
| PRD→spec | `--from-prd <url>`（仅需网络可达）/ `--from-browser <url>`（前置依赖 `bsk` CLI + browser-skill + Chromium 登录态） |
| 实时 watch | `free-kiro watch --preset <name>`，5 个 preset：`default` / `lint` / `status` / `reactive`（=lint+status）/ `full`（=lint+status+report） |
| 自升级 | `free-kiro upgrade` 校验 SHA256 + re-exec；POSIX (Linux/macOS) 路径 `os.Exec` 替换当前进程；Windows 下不自动 re-exec，升级后需手动重跑 `free-kiro upgrade` 或重启进程 |
| demo 流程 | `top1-demo-onboarding` spec 提供首次跑通的端到端示例 |

## 安装与分发

- Homebrew tap：`brew install jingyu525/tap/free-kiro`
- `curl|bash` 一行脚本（自动检测平台 + 写 `~/.local/bin`）
- GitHub Releases 直下（GoReleaser 跨 6 个平台）

## 非目标（防止 scope creep）

- **不是** Kiro IDE 的复刻 —— 不做 IDE、不做云端协作、不做账号体系
- **不替代** lint 工具 —— 只做 spec 文档的结构与语义门禁
- **不引入** 与 spec 工作流无关的新外部依赖 —— 保持 Go stdlib + 3 个直接依赖的精简哲学
- **不绑定** 单一 AI coding 工具 —— 通过 hook 信封兼容所有支持 hook 的工具
