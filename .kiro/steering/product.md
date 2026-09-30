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

## 核心能力（v0.7.0）

| 能力 | 说明 |
|---|---|
| spec 状态机 | `draft → requirements → design → tasks → approved → implementing → done` |
| EARS lint | 5 种 EARS 句式 + 无条件基线，缺漏即拦截 |
| bugfix 三段式 | Current / Expected / Unchanged Behavior 契约 |
| tasks wave | `tasks.md` 依赖图 → 拓扑排序 → 并行 wave 调度 |
| steering 两级作用域 | workspace 覆盖 global，4 种 inclusion 模式 |
| hook 信封 | 写出的 JSON 兼容 Kiro IDE event-keyed 格式 |
| drift 检测 | approve 时锁 baseline，编辑后自动暴露差异 |
| advisory | vague / 重复 AC / 可追溯性问题（不阻塞） |
| 可视化 | ASCII tree + Mermaid + 本地 web dashboard（SSE 实时） |
| 跨多 spec | `.kiro/.current` 自动标记活跃 spec |
| issue→spec | `free-kiro spec new --from-issue <gh-url>` |
| PRD→spec | `--from-prd <url>` / `--from-browser <url>` |
| 实时 watch | `free-kiro watch --preset reactive` |
| 自升级 | `free-kiro upgrade` 校验 SHA256 + re-exec |

## 安装与分发

- Homebrew tap：`brew install jingyu525/tap/free-kiro`
- `curl|bash` 一行脚本（自动检测平台 + 写 `~/.local/bin`）
- GitHub Releases 直下（GoReleaser 跨 6 个平台）

## 非目标（防止 scope creep）

- **不是** Kiro IDE 的复刻 —— 不做 IDE、不做云端协作、不做账号体系
- **不替代** lint 工具 —— 只做 spec 文档的结构与语义门禁
- **不引入** 新外部依赖 —— 保持 Go stdlib + 3 个直接依赖的精简哲学
- **不绑定** 单一 AI coding 工具 —— 通过 hook 信封兼容所有支持 hook 的工具