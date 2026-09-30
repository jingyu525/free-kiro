# add-task-priority

为命令行 todo 应用增加 priority 字段(P0/P1/P2),并按 priority 升序
排序展示。tasks 持久化到本地 JSON 文件 `~/.todo-tasks.json`。

## User Stories

- As a todo app user, I want to mark a task with priority P0/P1/P2,
  so that I can identify and tackle the most urgent items first.
- As a todo app user, I want the list command to show tasks sorted by
  priority ascending then by created_at ascending, so that ties are
  resolved deterministically.

## Acceptance Criteria

- WHEN the user runs `todo add "buy milk" --priority P0` THE SYSTEM
  SHALL persist a new task with `priority="P0"`, `done=false`,
  `created_at=<RFC3339-now>`, and `id=<next monotonic id>` to
  `~/.todo-tasks.json`, and SHALL print one line
  `created <id> P0 buy milk` to stdout.
- WHEN the user runs `todo add "buy milk"` without `--priority` THE
  SYSTEM SHALL default the new task's priority to `P2`.
- WHEN the user runs `todo list` and there are ≥2 tasks with mixed
  priorities THE SYSTEM SHALL print tasks in ascending priority
  order (P0 first, P2 last); tasks with equal priority SHALL be
  ordered by ascending `created_at`.
- WHEN the user runs `todo list` and `~/.todo-tasks.json` does not
  exist THE SYSTEM SHALL print `no tasks` to stdout and exit 0
  (without creating the file).
- WHEN the user runs `todo add` with `--priority X` (not in
  {P0,P1,P2}) THE SYSTEM SHALL exit with code 3 and print to stderr
  `invalid priority: X (must be P0, P1, or P2)`.
- THE SYSTEM SHALL persist every accepted task write atomically
  (write to `~/.todo-tasks.json.tmp` then rename) so that a process
  crash mid-write leaves the previous valid file untouched.
- THE SYSTEM SHALL store tasks as a JSON array of objects with the
  exact keys `id` (int), `text` (string), `priority` (string in
  {P0,P1,P2}), `done` (bool), `created_at` (RFC3339 string).

## Out of Scope

- 不实现任务编辑 / 删除 / done toggle — 本 spec 只演示 priority +
  list + 持久化三件套
- 不实现优先级筛选 (`--filter`) / 排序
- 不实现 `--help` 之外的文档或 README