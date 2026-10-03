# per-ide-instructions

把 `free-kiro init --ide <id>` 写入的项目级 agent 指令从单一 `.kiro/AGENTS.md`
升级为每个 IDE 各自的真实加载位置。当前 AGENTS.md 对 Claude Code、Cursor、Continue
完全不生效——它们各自有专属的指令文件协议。本次改造让 free-kiro 的"先想清楚 → 再
动手"工作流能在所有受支持的 IDE 里被 agent 主动读到，而不只是 hook 拦截层。

## User Stories

- As a developer using Claude Code I want `free-kiro init --ide claude-code` to
  write a project-root `CLAUDE.md` so that my agent actively loads the free-kiro
  workflow instructions on every session start (not just via PreToolUse gate).
- As a Cursor user I want free-kiro to write `.cursorrules` and a modular
  `.cursor/rules/free-kiro.md` so that Cursor's rules system loads free-kiro
  instructions on every agent invocation.
- As a Continue user I want free-kiro to write `.continuerules` and
  `.continue/rules/free-kiro.md` so that Continue loads free-kiro instructions
  per its rules system.
- As an OpenCode / CodeBuddy user I want free-kiro to keep writing `AGENTS.md`
  at the project root so that my agent's instruction loader picks it up
  unchanged from current behavior.

## Acceptance Criteria

- [AC-1] WHEN the user runs `free-kiro init --ide <id>` for a known IDE id,
  THE SYSTEM SHALL write every project-level instruction file declared for
  that id in the canonical table
  (`Claude Code → ["CLAUDE.md"]`,
  `Cursor → [".cursorrules", ".cursor/rules/free-kiro.md"]`,
  `Continue → [".continuerules", ".continue/rules/free-kiro.md"]`,
  `OpenCode → ["AGENTS.md"]`,
  `CodeBuddy → ["AGENTS.md"]`),
  prepending the `# free-kiro-managed:` marker on the first content line so the
  file can be identified as free-kiro-owned.
- [AC-2] WHILE the target instruction file already exists,
  THE SYSTEM SHALL skip writing that file and reuse the existing contents,
  unless `--overwrite-instructions` is passed, in which case THE SYSTEM SHALL
  overwrite the file unconditionally.
- [AC-3] WHERE `--ide none` is supplied,
  THE SYSTEM SHALL skip IDE-level instruction writes and only write
  `.kiro/AGENTS.md` so the workspace-level steering store can still load it.
- [AC-4] UNLESS the workspace root is not writable,
  THE SYSTEM SHALL create any missing parent directories
  (including `.cursor/`, `.cursor/rules/`, `.continue/`, `.continue/rules/`)
  with permission `0o755` before writing.
- [AC-5] IF the same instruction file path is targeted by multiple selected IDEs in
  one init invocation (e.g. `--ide opencode --ide codebuddy` both target
  `AGENTS.md`), THEN THE SYSTEM SHALL write that path exactly once per
  invocation and return the single write in the result list.
- [AC-6] THE SYSTEM SHALL continue to preserve the user's other top-level keys in
  `~/.{ide}/settings.json` (`model`, `enabledPlugins`, etc.) when upserting
  the two canonical free-kiro hooks
  (`PreToolUse matcher="Edit|Write"` and `SessionStart`),
  matching the existing `# free-kiro-managed:` upsert contract.
- [AC-7] THE SYSTEM SHALL accept the legacy `--overwrite-agents` flag, forward its
  boolean to the same overwrite switch as `--overwrite-instructions`, and emit
  a stderr warning naming the new flag so existing scripts keep working for
  one minor version.
- [AC-8] THE SYSTEM SHALL keep writing `.kiro/AGENTS.md` regardless of which IDE id
  is selected (including `--ide none`) so the existing steering-store loader
  at `internal/steering/store.go` continues to find it.
- [AC-9] THE SYSTEM SHALL expose the IDE→instruction-files mapping as a single
  `instructionFiles map[ID][]string` constant in `internal/ide/ide.go` so a
  future IDE can be added by editing one line.

## Out of Scope

- Rewriting the SKILL.md bundle distributed by `free-kiro skill install`.
- Changing the hook event-keyed JSON envelope schema.
- Introducing new third-party Go modules.
- Verifying whether CodeBuddy officially reads `AGENTS.md` (we follow the
  existing Tencent domestic convention; if CodeBuddy proves otherwise in a
  later audit, a follow-up spec will add a CodeBuddy-specific filename).
- Removing the existing `.kiro/AGENTS.md` write path (kept as workspace
  steering doc; see AC "THE SYSTEM SHALL keep writing").
