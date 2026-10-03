# free-kiro-managed:
# AGENTS.md — 本项目 AI 编码 agent 指令

本项目使用 **free-kiro** 管理规约工作流。free-kiro 是一个单二进制 CLI，
把代码改动门控在已写好的 spec 文档上。所有改动**先 spec 后代码**。

## free-kiro 在这个项目做什么

- 所有 spec 在 `.kiro/specs/<name>/` 下，包含 `requirements.md` /
  `design.md` / `tasks.md`（或 bugfix 的 `bugfix.md`）。
- spec 阶段流转：draft → requirements → design → tasks → approved →
  implementing → done。
- 阶段流转由 `free-kiro lint` 门控（EARS + 结构规则）。
- tasks.md 里的任务用 `[deps: #N1,#N2]` 声明依赖；
  `free-kiro task list <name>` 显示并行 wave 调度。

## 每个工具调用前（除 Read）

    SPEC=$(cat .kiro/.current 2>/dev/null) && free-kiro spec next "$SPEC"

这是 oracle 的下一步动作建议。如果 lint 失败，先修再继续。
SessionStart hook 会自动跑同样的命令。

如果 `.kiro/.current` 不存在（没活跃 spec），先跑 `free-kiro spec new <name>`，
新 spec 会自动被标记为活跃。

## 每次 Edit / Write 之前

    free-kiro lint || exit 2

IDE 的 PreToolUse hook 会自动跑。如果看到 lint ERROR，先编辑 spec 文档
（而不是代码），直到 `free-kiro lint` 全绿，再继续。

## 完成实现时

    free-kiro spec complete <name>

## 帮助

- `free-kiro --help` — 命令树
- `free-kiro doctor` — 诊断安装 + hook 问题
- `docs/EARS.md` — EARS 句式参考
- `docs/HOOKS.md` — hook 配置参考
- `docs/STEERING.md` — 项目上下文注入规则