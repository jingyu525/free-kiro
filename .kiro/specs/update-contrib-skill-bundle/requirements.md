# update-contrib-skill-bundle

把 `contrib/skills/free-kiro/` 的 SKILL.md bundle 与当前 `free-kiro` 二进制
的能力面重新对齐：补齐自 c7ed84f 之后 binary 新增的命令 / flag / 概念
（`upgrade`、`demo`、`spec quick` / `spec sync` / `spec show --tree` /
`spec list` / `spec analyze` / `spec status --json`、steering context、
`watch --command` / `--preset full` / `--root` / `--debounce` /
`--verbose`、`serve --bind` / `--port` / `--open`、`skill path` /
`skill version`、`upgrade --check --force`、`demo --ide` 等），并刷新
`skill.json` 里的 sha256 校验值，使 `free-kiro skill install` / `update`
重新通过校验。

## User Stories

- As a free-kiro 安装者（Claude Code / OpenCode / Codex / CodeBuddy 用户）I want 装到的 SKILL.md 反映 binary 当前真实能力 so that AI 助手不会引用已废弃的命令 / flag、也不会漏掉新增的入口。
- As a 维护者 I want 改动 skill.json 的 sha256 与 bundle 实际文件保持一致 so that `free-kiro skill install` 校验通过、`free-kiro skill show` 不报 sha 不匹配。
- As a 排错者 I want troubleshooting.md 覆盖 binary 真实存在的错误类型 so that 用户对照表能直接查到对应解法。

## Acceptance Criteria

[AC-1] WHEN `contrib/skills/free-kiro/SKILL.md` 被读取时 THE SYSTEM SHALL 在"快速参考"块列出至少 14 类命令（spec / task / steering / hook / lint / watch / serve / doctor / report / skill / upgrade / demo / status / init），覆盖全部新增子命令。

[AC-2] WHILE `skill.json` 的 `sha256.*` 字段未刷新时 THE SYSTEM SHALL 在 `free-kiro skill install` / `update` 流程中以 exit 1 终止（sha256 mismatch 校验失败）。

[AC-3] WHEN SKILL.md 提及 spec 状态机预言机时 THE SYSTEM SHALL 使用 `free-kiro spec next <name>` 前缀（而非 `kiro spec next`），全文 0 处出现 `kiro spec next`。

[AC-4] WHEN `references/ops.md` 描述 `watch` 时 THE SYSTEM SHALL 在同一表格里覆盖 5 个 preset（`default` / `lint` / `status` / `reactive` / `full`）+ `--command` / `--debounce`（默认 500ms）/ `--root` / `--verbose` 共 4 个 flag。

[AC-5] WHERE `references/ops.md` 描述 `serve` 时 THE SYSTEM SHALL 列出 `--bind`（默认 `127.0.0.1`）/ `--port`（默认 `7373`）/ `--open` 共 3 个 flag 与 4 个端点（`/` / `/api/summary` / `/api/specs` / `/api/spec/<n>`）。

[AC-6] WHERE `references/hooks.md` 描述事件命名时 THE SYSTEM SHALL 同时列出 free-kiro 内部命名（`file.save` / `file.create` / `file.delete` / `prompt.submit` / `task.run` / `manual`，共 6 项）与 Kiro 官方命名（`SessionStart` / `UserPromptSubmit` / `PreToolUse` / `PostToolUse` / `PostFileSave` / `PostFileCreate`，共 6 项）。

[AC-7] WHEN `references/spec.md` 描述 spec 工作流时 THE SYSTEM SHALL 覆盖全部 12 个 spec 子命令（new / generate / approve / start / complete / quick / sync / show / list / analyze / status / next）与 `spec new --workflow` / `--type` / `--quick` 共 3 个变体 flag。

[AC-8] WHEN `references/troubleshooting.md` 描述 lint 错误时 THE SYSTEM SHALL 至少覆盖 6 种错误码（`no-ears` / `dependency cycle` / `dependency dangling` / `self-reference` / `illegal phase transition` / `cannot approve: tasks.md must be generated first`）。

[AC-9] IF `scripts/compute-skill-sha.sh` 重算 hash 后 THEN THE SYSTEM SHALL 在 1 次 `free-kiro skill install --dry-run --app claude-code` 调用内 exit 0（即 sha 与 skill.json 字段 1:1 自洽）。

[AC-10] WHEN `SKILL.md` 的"自检"段被引用时 THE SYSTEM SHALL 同时列出 `free-kiro skill show` 与 `free-kiro doctor` 两条命令，且 docstring 不超过 5 行。

[AC-11] UNLESS `spec.json` 的 `version` 字段被用户显式覆盖时 THE SYSTEM SHALL 保持 `0.7.0-dev`（与本仓库当前 dev 构建对齐）。

[AC-12] THE SYSTEM SHALL 在 SKILL.md 的"退出码契约"段内 1 张表覆盖全部 4 个 exit code（0 OK / 1 lint ERROR / 2 engine error / 3 usage error），无遗漏行。

## Out of Scope

- 不修改 free-kiro 二进制本身的代码、命令、flag、行为。
- 不修改 `docs/` 下的开发者文档（POLICY.md / AGENT_RULES.md / CODING_STYLE.md）。
- 不改 `install.sh` 的安装逻辑与 `package.json` 的依赖。
- 不改 `contrib/skills/free-kiro/references/prd-fetch.md`（已与 binary 对齐）。
- 不引入新的 reference 文件，不删除现有 reference 文件。
- 不动 `.kiro/specs/` 下其它已完成 spec 的内容。
- 不触发 release pipeline（`skill.json.version` 不在本 spec 内升正式版）。