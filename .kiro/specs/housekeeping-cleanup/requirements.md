# housekeeping-cleanup

清理 codereview 报告里的 14 个低优先级 housekeeping findings：dead code、DRY
重复实现、JSON tag 一致性、注释与实现不一致、命名歧义、错误处理风格。
所有改动不引入行为差异，纯 refactor + 注释同步 + 测试加固。

## User Stories

- As a free-kiro maintainer I want the leftover code-review findings
  cleaned up so that the codebase has no dead code, no duplicated
  helpers, consistent JSON tags, and accurate godoc that matches the
  implementation.
- As a future contributor reading the code I want shared helpers in
  one place (e.g. `internal/text`) rather than duplicated `rangeLines`
  implementations across `lint` and `taskgraph` so that the next line
  splitter change touches one file.

## Acceptance Criteria

[AC-1] WHEN `spec new` (or any Engine method) needs to log a non-fatal warning to the user THE SYSTEM SHALL route through the project's logger (or a clearly-named stderr helper) within 0 microseconds rather than `fmt.Fprintf(os.Stderr, ...)` so the message format is consistent across the CLI.

[AC-2] WHEN `spec Complete` finishes and `e.ws.ReadCurrent()` still points at the spec THE SYSTEM SHALL wrap and surface any `ClearCurrent` error instead of discarding it via `_ =` within the same call (under 1 ms additional latency).

[AC-3] WHEN the dashboard server binds to an IPv6 address such as `[::1]:8080` THE SYSTEM SHALL return a URL like `http://[::1]:8080` from `Server.URL()` instead of producing a malformed `http://::1]:8080` so users can open the dashboard.

[AC-4] WHEN a hook's timeout is `0` seconds THE SYSTEM SHALL treat it as an explicit "disabled" flag (via a new `Hook.Disabled bool` field) rather than conflating 0-timeout with disable-hook so authors see 1 knob per intent (not 2 overloaded meanings).

[AC-5] WHEN the Baseline struct's JSON tag is inspected THE SYSTEM SHALL have `IgnoredCodes` ↔ `ignored_codes` matching pluralisation (rename field to `IgnoredIssues` OR rename JSON tag — pick one and apply consistently) so authors editing .baseline.json (at least 7 historical files) aren't surprised by a hidden rename.

[AC-6] WHERE the codebase needs to iterate a multi-line text string THE SYSTEM SHALL use a single shared `RangeLines` helper in `internal/text` so `internal/lint/ears.go` and `internal/taskgraph/parse.go` no longer each carry their own private copy (0 duplicates — from 2 down to 1 shared implementation).

[AC-7] WHEN `taskgraph.ParseTasks` extracts the numeric task ID after `#` THE SYSTEM SHALL call `strconv.Atoi` (stdlib) instead of the hand-rolled `atoi` helper within the same line so we stop maintaining at least 14 lines of duplicate stdlib logic.

[AC-8] THE SYSTEM SHALL delete at least 3 pieces of dead code: `Baseline.Empty()` method (no callers), the `var _ = ferrors.New` placeholder in `internal/hooks/dispatch.go`, and the `var _ = json.Marshal` placeholder in `internal/spec/engine.go`, AND each deletion SHALL leave `go build ./...` and `go vet ./...` clean within 0 seconds.

[AC-9] WHERE the `visualize` package's godoc references future enhancements or latency budgets THE SYSTEM SHALL keep every sentence in sync with `server_sse.go` and `server.go` (the "SSE/fsnotify future enhancement" sentence that contradicts the existing SSE handler, and the "200 ms" budget that no longer matches the 5s `shutdownTimeout` constant) within 1 PR.

[AC-10] THE SYSTEM SHALL keep every refactor covered by the existing test suite (no behavior change) AND `go test -race ./internal/...` SHALL report 0 failures AND `go vet ./...` SHALL report 0 warnings AND `free-kiro lint` SHALL report no new findings on this spec.

## Out of Scope

- New functionality (this spec is pure cleanup).
- The `internal/skill/{install,release,manifest,show,paths}.go` packages
  (not touched by the original review).
- Performance changes (micro-optimisations like removing the
  hand-rolled `rangeLines` are correctness-driven, not perf).
- The hand-rolled `atoi` in `taskgraph/waves.go`? wait — there is none;
  only `parse.go` has it. (Confirmed via grep.)