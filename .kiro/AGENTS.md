# AGENTS.md — 本项目 AI 编码 agent 指令

本项目使用 **free-kiro** 管理规约工作流。free-kiro 是一个单二进制 CLI，
把代码改动门控在已写好的 spec 文档上。所有改动**先 spec 后代码**。

## free-kiro 在这个项目做什么

- 所有 spec 在 `.kiro/specs/<name>/` 下，包含 `requirements.md` /
  `design.md` / `tasks.md`（或 bugfix 的 `bugfix.md`）。
- spec 阶段流转：draft → requirements → design → tasks → approved →
  implementing → done。
- 阶段流转由 `free-kiro lint` 门禁（EARS + 结构规则）。
- tasks.md 里的任务用 `[deps: #N1,#N2]` 声明依赖；
  `free-kiro task list <name>` 显示并行 wave 调度。

## 每个工具调用前（除 Read）

    SPEC=$(cat .kiro/.current 2>/dev/null) && free-kiro spec next "$SPEC"

这是 oracle 的下一步动作建议。如果 lint 失败，先修再继续。
SessionStart hook 会自动跑同样的命令。

如果 `.kiro/.current` 不存在（没活跃 spec），先跑 `free-kiro spec new <name>`，
新 spec 会自动被标记为活跃。

## 编码规范（SessionStart 必须先 Read）

**在写任何 Go 代码之前，先 Read [`docs/CODING_STYLE.md`](../docs/CODING_STYLE.md)**。

该文档 8 章覆盖命名 / 错误处理 / 并发 / 接口 / 测试 / 注释 / 依赖 /
AI 协作。第 8 章"AI agent 协作"是硬约束（违反任意一条 = PR 拒收）。

硬约束摘录：

1. 零 `// TODO` / `// FIXME` / `// XXX` — 未完成的功能**不要写代码**，先
   回 spec 阶段补 requirements/design。
2. 零吞错误（`_ = doX()` / `log.Print(err)` 后继续）。
3. 零硬编码 magic number（端口、超时、阈值都要 `const` 或在 `internal/config`）。
4. 改动 > 50 行先 `free-kiro spec new`，三件套全绿再动代码。

## 每次 Edit / Write 之前

    free-kiro lint || exit 2

IDE 的 PreToolUse hook 会自动跑。如果看到 lint ERROR，先编辑 spec 文档
（而不是代码），直到 `free-kiro lint` 全绿，再继续。

## 完成实现时

    free-kiro spec complete <name>

## 帮助

- `free-kiro --help` — 命令树
- `free-kiro doctor` — 诊断安装 + hook 问题
- `docs/CODING_STYLE.md` — Go 编码规范（必读）
- `docs/EARS.md` — EARS 句式参考
- `docs/HOOKS.md` — hook 配置参考
- `docs/STEERING.md` — 项目上下文注入规则
