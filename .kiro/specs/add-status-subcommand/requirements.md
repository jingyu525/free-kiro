# add-status-subcommand

给 free-kiro CLI 加一个聚合的 `status` 子命令，展示当前 `.kiro/`
workspace 里所有 spec 的整体状态（phase / approved / drift / tasks 进度），
并标记活跃 spec。人类可读（默认）+ 机器可读（`--json`）双输出，供
`watch` / `serve` / CI 复用。

## User Stories

- As a developer running free-kiro on a project, I want to run
  `free-kiro status` and see a one-screen summary of every spec's
  lifecycle stage, so that I can spot drift and blocked work at a
  glance without opening every `requirements.md` / `tasks.md`.
- As an automation script (watch reactive preset, serve dashboard, CI
  pipeline), I want `free-kiro status --json` to return a stable
  machine-readable snapshot, so that I can drive dashboards or fail
  builds without parsing human output.
- As a developer context-switching between specs, I want
  `free-kiro status` to mark the currently active spec (from
  `.kiro/.current`), so that I can confirm which spec is "in focus"
  before editing.

## Acceptance Criteria

- WHEN the user runs `free-kiro status` in a workspace with one or
  more specs, THE SYSTEM SHALL print a human-readable table to stdout
  that contains, for each spec, the columns `name` / `phase` /
  `approved` (yes/no) / `drift` (clean/drift/n-a) / `tasks` (done /
  total / waves), and SHALL mark the active spec (read from
  `.kiro/.current`) with a leading `*` or an `active` column set to
  `yes`.
- WHEN the user runs `free-kiro status --json`, THE SYSTEM SHALL
  print a JSON object with shape `{ "generated_at": RFC3339,
  "active": "<name>|"", "specs": [ ...per-spec objects... ] }` to
  stdout, with the same field set per spec (name / phase / approved /
  drift / tasks-done / tasks-total / tasks-waves), and SHALL exit 0.
- WHERE the workspace `.kiro/` directory does not exist, THE SYSTEM
  SHALL print an error message on stderr instructing the user to run
  `free-kiro init`, and exit with code 3 (usage error).
- WHERE `.kiro/specs/` exists but contains no spec subdirectories, THE
  SYSTEM SHALL print `No specs found` to stdout and exit 0 (in JSON
  mode: `{"active":"","specs":[]}`).
- WHILE `.kiro/.current` exists but its value does not match any
  existing spec directory, THE SYSTEM SHALL still print the spec list
  and SHALL either omit the active marker or set `active` to the
  pointer value with a `(not found)` suffix, without erroring out.
- UNLESS `--json` is passed, THE SYSTEM SHALL emit plain text only
  (no ANSI color escapes) so the output is safe to pipe to `less`,
  `grep`, or `tee`.
- THE SYSTEM SHALL not modify any spec, steering, hook, or `.kiro/`
  metadata — `status` is read-only.

## Out of Scope

- A `free-kiro status <name>` form for deep-diving into one spec —
  that is already covered by `free-kiro spec show <name>`.
- Re-running `lint` or recomputing drift inside `status` — callers
  use `free-kiro lint` separately.
- Emitting webhooks / log events on status changes — `status` is a
  snapshot tool, not a trigger.
- ANSI colors / TTY-only niceties — keep output greppable by default;
  no `--color` flag in this spec.
