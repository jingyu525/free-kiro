# improve-visualize-hooks-reliability

修复 codereview 报告里的 6 个高优先级 correctness 缺陷，覆盖 3 个包：
`internal/visualize`、`internal/hooks`、`internal/lint`。目标是让 dashboard
SSE 长连接、hook 派发、lint 门禁在并发负载与局部 IO 失败下仍然正确。

## User Stories

- As a free-kiro maintainer I want the 6 high-priority correctness bugs from
  the `internal/` code-review fixed so that the visualize dashboard, hook
  dispatch, and lint gate behave reliably under concurrent access and
  partial IO failures (race detector clean, no Shutdown hang, no silent
  error swallow).
- As a free-kiro user running `free-kiro serve` I want the dashboard SSE
  endpoint to support multiple concurrent browser clients without dropping
  refresh events so that I can keep the dashboard open in several tabs
  without surprises.

## Acceptance Criteria

[AC-1] WHEN 2 or more browser clients concurrently connect to the SSE endpoint THE SYSTEM SHALL serialize all subscribe/unsubscribe/broadcast operations under a mutex so that `go test -race ./internal/visualize/...` reports 0 race findings within 30 seconds (≥ 50 concurrent goroutines).

[AC-2] WHEN a second SSE client connects after the first has already opened its subscription THE SYSTEM SHALL NOT spawn an additional watchChanges goroutine so that at most 1 watcher goroutine is alive per Server lifetime.

[AC-3] WHEN `Server.Shutdown` is called and watchChanges was never started THE SYSTEM SHALL return within 5 seconds instead of blocking forever.

[AC-4] WHEN `Server.Shutdown` is called while SSE clients are still connected THE SYSTEM SHALL close them within 5 seconds using `http.Server.Shutdown(ctx)` instead of `srv.Close()`.

[AC-5] WHEN requirements.md or design.md exists but is unreadable due to a non-`fs.ErrNotExist` IO error THE SYSTEM SHALL emit a lint-*-read-error Issue with severity ERROR whose Message wraps the underlying error via `fmt.Errorf("...: %w", err)` within the same lint pass.

[AC-6] WHILE no hook JSON file in `.kiro/hooks/` has been modified since the last cache fill THE SYSTEM SHALL complete 100 consecutive `Registry.Match(...)` calls in under 100 ms total wall-clock so that PreToolUse hooks do not re-read disk on every file save.

[AC-7] WHEN an `Issue.Code` is listed in the spec's `.baseline.json` `ignored_issues` THE SYSTEM SHALL exclude that Issue from `lint.Gate()` output within 0 microseconds based on an explicit `Issue.Baseline bool` field rather than a `strings.HasPrefix(Issue.Message, ...)` heuristic.

[AC-8] WHERE the 6 fixes land in code THE SYSTEM SHALL keep every fix covered by unit tests (≥ 1 positive + 1 negative case per fix) AND `go test -race ./internal/...` SHALL report 0 failures AND `go vet ./...` SHALL report 0 warnings.

[AC-9] THE SYSTEM SHALL land all 6 fixes within a single PR titled `fix(reliability): visualize SSE race + watchChanges dedup + Shutdown timeout + lint IO error wrap + hooks Match cache + Gate baseline field`, with no public API breakage.

## Out of Scope

- Fixing the lower-priority findings (dead `var _ = xxx` placeholders,
  `IgnoredCodes` JSON tag pluralisation, `TaskGraph.atoi` hand-roll,
  duplicated `rangeLines` between `lint` and `taskgraph`, etc.) —
  these are tracked for a future housekeeping spec.
- Replacing `visualize/server_sse.go`'s polling watcher with
  `fsnotify` — that's a separate "real-time file watcher" spec.
- Adding per-IDE hook retry/backoff — separate resilience spec.