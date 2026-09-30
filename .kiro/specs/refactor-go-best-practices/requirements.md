<!-- 用 1-3 句话说明这个 feature 的目标和边界 -->

# refactor-go-best-practices

对 free-kiro 当前 Go 代码（13 个 internal 包、约 13.3K 行）做一次
规约驱动的重构，遵循 Effective Go 与 Go 官方最佳实践。聚焦三大方向：
**复用**（抽 frontmatter 公共包 + 抽 cmd runner 模板 + 拆超大文件）、
**错误处理**（cli 层统一 typed errors，让 ExitCode 真正生效）、
**架构**（拆 cli/spec_lifecycle.go / cli/skill.go / cli/doctor.go /
spec/engine.go 等 5 个 300+ 行超大文件）。

## User Stories

- As a free-kiro maintainer, I want cli-layer errors to flow through the
  typed `internal/errors` taxonomy, so that `ExitCode(err)` returns the
  documented 0/1/2/3 instead of always 2 for every failure mode.
- As a free-kiro contributor, I want frontmatter parsing / markdown section
  reading / atomic-write helpers to live in one shared package, so that
  new specs and steering docs don't re-implement the same YAML frontmatter
  reader four times.
- As a free-kiro reviewer, I want the 5 currently oversized files
  (`cli/spec_lifecycle.go` 333 lines, `cli/skill.go` 387, `cli/doctor.go`
  343, `spec/engine.go` 468, `visualize/server.go` 444) split by
  responsibility, so that no single file mixes command wiring, business
  logic, and IO helpers.
- As a free-kiro release engineer, I want the refactor to be behavior-
  preserving, so that the v0.7.x CLI surface (flags, args, stdout format,
  exit codes) stays byte-identical for end users.

## Acceptance Criteria

### Reuse — frontmatter 公共包

- THE SYSTEM SHALL provide `internal/frontmatter` package exposing
  `Parse(r io.Reader) (Frontmatter, body []byte, err error)`,
  `Marshal(fm Frontmatter, body []byte) ([]byte, error)`, and
  `Validate(fm Frontmatter, schema map[string]FieldRule) error`,
  and SHALL replace the four existing inline frontmatter implementations
  in `internal/cli/init.go`, `internal/steering/store.go`,
  `internal/skill/`, and `internal/spec/engine.go`.
- WHEN any of the four call-sites parses a YAML frontmatter document,
  THE SYSTEM SHALL produce identical `Frontmatter` map output (same keys,
  same string values, same body separator handling) as the original
  per-file implementation for every existing `testdata/*.md` fixture.
- THE SYSTEM SHALL add `internal/frontmatter/frontmatter_test.go`
  covering at least 8 cases (empty body / multi-line body / BOM / CRLF /
  missing separator / invalid YAML / unknown field rejected by schema /
  schema with required field missing).
- UNLESS the consuming package explicitly opts out via a TODO comment
  referencing this spec, THE SYSTEM SHALL delete the four legacy
  inline implementations after migration.

### Reuse — cmd runner 模板

- THE SYSTEM SHALL provide `internal/cli/runner.go` exposing a
  `RunCmd` helper that takes `(ctx context.Context, eng *spec.Engine,
  args []string) error` and centralises the four repeated patterns
  observed across `spec new / generate / approve / start / complete`,
  `skill install / update / show`, and `doctor / watch / serve`: (a)
  resolve engine from cobra flags, (b) map typed errors through
  `errors.ExitCode` and `exitWithError`, (c) print `next:` hint
  after success, (d) recover from panics into a `KiroError`.
- WHEN a refactored subcommand delegates to `RunCmd`, THE SYSTEM SHALL
  behave identically to its pre-refactor version for: every documented
  flag combination, every stdin/TTY path, and every exit code (0/1/2/3).
- THE SYSTEM SHALL reduce the total line count of the affected subcommand
  implementations by at least 25 % measured before-and-after.

### Error handling — typed-error 收敛

- THE SYSTEM SHALL replace every `fmt.Errorf(...)` call inside
  `internal/cli/*.go` (excluding `osutil.go` and `version.go`) with
  either `errors.UsageError`, `errors.Wrap`, or `errors.New` from the
  `internal/errors` package, so that `errors.ExitCode(err)` returns 3
  for caller-misuse conditions (missing arg, mutually-exclusive flags,
  invalid input) instead of the current default 2.
- WHEN a cli subcommand is invoked with a missing required flag, THE
  SYSTEM SHALL exit with code 3 (UsageError), not code 2, AND SHALL
  write a single human-readable Chinese message to stderr matching the
  pre-refactor text byte-for-byte.
- WHEN a cli subcommand fails because the `.kiro/` workspace is missing
  or unreadable, THE SYSTEM SHALL exit with code 2 (WorkspaceError),
  wrapping the underlying `os.PathError` via `%w` so `errors.Is` and
  `errors.Unwrap` work for callers.
- THE SYSTEM SHALL NOT introduce new external dependencies for error
  wrapping — only stdlib (`errors`, `fmt`) and the existing
  `internal/errors` package.
- UNLESS the call site is inside a test file (`*_test.go`), THE SYSTEM
  SHALL NOT retain any `fmt.Errorf("...")` call in `internal/cli/`
  whose error message describes a caller-misuse condition.

### Architecture — 拆超大文件

- THE SYSTEM SHALL split each of the five oversized files into 2–3
  files, each ≤ 250 lines, organised by responsibility (one file per
  concern: command wiring / engine helpers / IO / formatting):
  - `cli/spec_lifecycle.go` (333) → `cli/spec_lifecycle.go` + `cli/spec_state.go` + `cli/spec_format.go`
  - `cli/skill.go` (387) → `cli/skill.go` + `cli/skill_install.go` + `cli/skill_sync.go`
  - `cli/doctor.go` (343) → `cli/doctor.go` + `cli/doctor_checks.go` + `cli/doctor_report.go`
  - `spec/engine.go` (468) → `spec/engine.go` + `spec/engine_state.go` + `spec/engine_io.go`
  - `visualize/server.go` (444) → `visualize/server.go` + `visualize/server_sse.go` + `visualize/server_static.go`
- WHEN a file is split, THE SYSTEM SHALL keep the original package name
  and original exported identifiers, AND SHALL update only the new
  internal call-sites (no CLI-surface or test-surface changes beyond
  what was already public).
- THE SYSTEM SHALL add or update at least one unit test file per
  extracted package (e.g. `internal/frontmatter/frontmatter_test.go`,
  `internal/cli/runner_test.go`) so that coverage on the new packages
  is ≥ 60 %.

### Regression — 行为不变

- THE SYSTEM SHALL pass `go vet ./...`, `go build ./...`, and
  `go test ./...` after each task in the implementation plan,
  measured as zero non-vet output and zero failing tests.
- WHEN the user runs any of the existing CLI subcommands listed in
  the README (`init / spec new|generate|approve|start|complete / lint /
  task / steering / hook / watch / serve / doctor / upgrade / skill /
  status / --version`), THE SYSTEM SHALL produce byte-identical stdout
  and exit codes compared to v0.7.0 for the documented flag set.
- THE SYSTEM SHALL keep the exit-code contract from
  `internal/errors/errors.go` lines 1–14 unchanged (0/1/2/3 mapping).
- THE SYSTEM SHALL NOT introduce new third-party `go.mod` entries;
  the only allowed additions are stdlib packages.

## Out of Scope

- Public CLI surface changes — no new flags, no renamed flags, no
  reordered arguments, no new subcommands.
- Adding new functionality (lint rules, IDE integrations, transport
  backends, etc.).
- Rewriting the spec state-machine algorithm in `spec/engine.go` —
  only file-level splits are allowed; logic is moved verbatim.
- Replacing `cobra` with another CLI framework.
- Generating gRPC / OpenAPI / protobuf bindings.
- Migrating from `errors.ExitCode` to a new error library.
- Performance optimisation (profiling, allocation reduction,
  concurrency rewrites) — only structural changes qualify.
- Adding new lint rules to `internal/lint/`.
- Documentation rewrites in `docs/` beyond the minimum required to
  reflect the new package layout.
