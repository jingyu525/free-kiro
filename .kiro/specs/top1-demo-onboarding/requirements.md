# top1-demo-onboarding

为 free-kiro 落地"5 分钟端到端 demo + 多 IDE hook 自动发现"——把
0 stars / 0 下载的开源现状，从"装上但跑不起来"推进到"clone 后 5 分钟
看到完整 spec → lint → dashboard → wave 调度闭环"。包含三块独立可交付
增量：`free-kiro demo` 子命令 + `examples/todo-app` 真实样例、`init
--ide auto` 扩展支持 5 种 AI 编码工具、内置 GitHub Actions
`spec-lint.yml` 模板。

## User Stories

- As a new free-kiro evaluator, I want to clone the repo, run one
  command, and see a complete spec → lint → status → dashboard loop
  within 5 minutes, so that I can judge the project's value without
  reading the README end-to-end or setting up my own project.
- As a maintainer of an existing project who just installed free-kiro,
  I want `free-kiro init --ide auto` to detect whichever AI coding
  tool I have installed (Claude Code / CodeBuddy / Cursor / Continue
  / OpenCode) and register hooks for all of them in one pass, so that
  I don't have to read 5 different integration docs to get the gate
  enabled.
- As a CI engineer integrating free-kiro into a team repo, I want a
  drop-in `.github/workflows/spec-lint.yml` workflow template that
  runs `free-kiro lint` on every PR with the right setup action and
  exit-code contract, so that the team can adopt it with a single file
  copy rather than reverse-engineering the docs.

## Acceptance Criteria

### Real example spec (`examples/todo-app/`)

- THE SYSTEM SHALL ship a directory `examples/todo-app/` in the
  repository root containing a fully-initialized `.kiro/` workspace
  with one spec named `add-task-priority` whose `requirements.md`,
  `design.md`, and `tasks.md` are all present with no placeholder
  markers, and pass `free-kiro lint examples/todo-app` with exit
  code 0.
- WHEN `examples/todo-app/.kiro/specs/add-task-priority/tasks.md` is
  inspected, THE SYSTEM SHALL show at least 4 tasks with realistic
  dependency annotations such that `free-kiro task list
  examples/todo-app` resolves into 2 or more parallel waves (i.e.,
  at least one wave with 2+ tasks).
- WHEN `free-kiro spec approve add-task-priority` is run inside
  `examples/todo-app/` and then 3 lines of an approval-time baseline
  are manually edited in `requirements.md`, THE SYSTEM SHALL report
  `drift: drift` via `free-kiro spec status` and SHALL list the
  exact diff hunks via `free-kiro spec status --json` (so the demo
  is ready to show "drift detected" without further setup).

### `free-kiro demo` subcommand

- WHEN the user runs `free-kiro demo` in the repository root, THE
  SYSTEM SHALL print a 5-step onboarding summary to stdout, write a
  `demostart.log` marker into `./.kiro/.demostart`, and exit with
  code 0.
- WHEN the user runs `free-kiro demo` outside the repository root,
  THE SYSTEM SHALL print a one-line Chinese message to stderr
  directing the user to `cd <repo-root>` first and exit with code 3.
- WHEN the user runs `free-kiro demo --no-color`, THE SYSTEM SHALL
  not emit any ANSI escape sequence in stdout or stderr (useful for
  CI logs and piping into `tee`).
- WHERE the `free-kiro demo` command is invoked with `--ide
  <claude-code|codebuddy|none>` (default `none`), THE SYSTEM SHALL
  additionally print the corresponding hook snippet that the user
  can paste into their AI assistant settings, and exit 0.
- WHERE the `free-kiro demo` command is invoked twice within 30
  seconds, THE SYSTEM SHALL print a warning indicating the demo is
  already running and exit 0 without rewriting the marker file.

### README 顶部重写

- THE SYSTEM SHALL replace the existing README "30 秒上手" section
  with a new "5 分钟 demo" section that links to `examples/todo-app/`
  and embeds the literal command sequence `cd examples/todo-app
  && ../../dist/free-kiro_*/free-kiro serve` followed by `xdg-open
  http://127.0.0.1:7373` (or platform-specific open command).
- WHEN the README is read on GitHub, THE SYSTEM SHALL ensure the
  topmost `## Why free-kiro` paragraph contains the literal phrase
  `5 分钟端到端 demo` so that visitors immediately discover the demo
  path.

### `init --ide auto` 多 IDE 支持

- THE SYSTEM SHALL extend `internal/ide.All()` to return at least
  five IDE identifiers: `claude-code`, `codebuddy`, `cursor`,
  `continue`, and `opencode`; `Parse(...)` SHALL accept all five
  identifiers (case-insensitive) and unknown identifiers SHALL
  surface a `UsageError` listing the five supported names.
- WHEN the user runs `free-kiro init --ide auto` on a machine where
  exactly 2 of the 5 supported IDE config directories are detected
  (e.g., `~/.claude` and `~/.codebuddy`), THE SYSTEM SHALL register
  the canonical free-kiro hooks into both `settings.json` files,
  print one `✓ <id>` line per installed IDE, and exit 0.
- WHEN the user runs `free-kiro init --ide cursor`, THE SYSTEM SHALL
  write to `~/.cursor/settings.json` (creating the directory if
  missing) using the same free-kiro marker convention as the
  existing Claude Code / CodeBuddy integrations.
- WHEN the user runs `free-kiro init --ide auto` on a machine with
  zero detected IDEs, THE SYSTEM SHALL print a stderr warning that
  lists the 5 supported IDEs and that recommends `--ide <name>` for
  explicit override, then exit 0 (the workspace is still usable).
- THE SYSTEM SHALL add `internal/ide/ide_test.go` cases for at
  least 5 identifiers × 3 paths (legacy compat, fresh install,
  settings-with-existing-user-hooks), totalling ≥15 new
  assertions.

### 内置 GitHub Actions 模板

- THE SYSTEM SHALL ship a template file
  `.github/workflows/spec-lint.yml.example` (committed in-repo) that
  references `jingyu525/free-kiro/.github/actions/setup-free-kiro@v0`
  and runs `free-kiro lint` then `free-kiro doctor` on every
  pull_request to main, with `exit 1` propagated to the job.
- WHEN a downstream consumer copies
  `.github/workflows/spec-lint.yml.example` to
  `.github/workflows/spec-lint.yml` (renaming away the suffix), THE
  system SHALL document the copy step in
  `docs/CI_INTEGRATION.md` (NEW) with a 5-step procedure covering
  install / setup action / lint / doctor / status badge.

### Cross-cutting

- THE SYSTEM SHALL not introduce any new external Go dependency
  beyond those already present in `go.mod` (i.e., the demo is built
  with stdlib + `fsnotify` + `cobra` only).
- WHEN any of the acceptance criteria in this spec is verified
  manually, THE SYSTEM SHALL record the verification result (command
  + exit code + observed output) in `docs/DEMO_VERIFICATION.md` so
  that future maintainers can replay the demo without re-deriving
  the steps.

## Out of Scope

- 不为五种 IDE 分别定制 hook envelope — 它们都沿用现有 Claude Code
  风格的 `event-keyed` JSON 信封(详见 `docs/HOOKS.md`)。
- 不实现 demo GIF/录屏自动生成 — README 中 demo 路径改用文字 + 截图占位,
  后续单独 spec 处理。
- 不发布到 npm / PyPI 等其他包管理器 — 现有 Homebrew + curl|bash 渠道
  维持不变。
- 不改动 free-kiro 二进制对用户的核心命令(`spec` / `lint` / `init` 等
  的 flag surface) — `init` 的 `--ide` 值集合是本次唯一扩展点。
- 不实现 `free-kiro demo` 的交互式向导(Wizard TUI)— 5 分钟 demo 只做
  命令行纯文本输出。