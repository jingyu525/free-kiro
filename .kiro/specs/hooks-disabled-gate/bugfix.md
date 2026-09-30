# hooks-disabled-gate — Bug Fix

<!--
bugfix.md 用三段式契约：
  ## Current Behavior    描述观察到的缺陷（禁止用 THE SYSTEM SHALL — 缺陷是错的）
  ## Expected Behavior   正确行为（必须含 THE SYSTEM SHALL）
  ## Unchanged Behavior  回归预防（必须用 THE SYSTEM SHALL CONTINUE TO）

详见 docs/EARS.md#bugfix-spec-bugfixmd
-->

## Current Behavior (Defect)

<!-- 描述观察到的错误行为。不要用 SHALL（缺陷是错的，不是"应当"） -->

- `Registry.Match()` filters out hooks with `enabled=false` but does not skip hooks with `disabled=true`. A hook with `{enabled: true, disabled: true, action_type: "agent"}` is returned to callers and dispatched.
- `runAgentAction()` has no `h.Disabled` guard. Even if `Match()` were fixed, the agent-path bypass means any code path that reaches `runAgentAction` (current or future) executes the prompt regardless of `disabled`. `runShellAction()` has the same guard, so this is an inconsistency introduced when `Disabled` was added.
- `Registry.cachedAll()` returns a freshly-allocated slice but only copies the slice header; the underlying `*models.Hook` pointers still alias the cache. The doc comment promises "callers can't mutate our cache" — this promise is currently upheld only because no caller mutates. Any future caller that writes through the returned slice (debug logging, per-call bookkeeping, race-detector test scaffolding) will silently corrupt the cache within the TTL window.

## Expected Behavior (Correct)

<!-- 正确行为；必须含 THE SYSTEM SHALL -->

- WHEN a hook has `disabled=true` THE SYSTEM SHALL skip it in `Match()` and not include it in the returned slice, regardless of `enabled`.
- WHEN `Dispatch()` invokes `runAgentAction()` for a hook with `disabled=true` THE SYSTEM SHALL return a `Result{OK: false, Error: "hook disabled"}` without calling `agentFn` and without invoking the placeholder-printing branch.
- WHEN `cachedAll()` returns its cached slice THE SYSTEM SHALL return Hook structs that are deep copies, so mutating any field on a returned hook does not affect the next `Match()` call within the TTL window.

## Unchanged Behavior (Regression Prevention)

<!-- 必须保留的行为；用 THE SYSTEM SHALL CONTINUE TO 形式 -->

- WHEN a hook has `enabled=true` and `disabled=false` THE SYSTEM SHALL CONTINUE TO include it in `Match()` results and execute its action through the existing `Dispatch` path.
- WHEN `runShellAction()` receives a hook with `disabled=true` THE SYSTEM SHALL CONTINUE TO return `Result{OK: false, Error: "hook disabled"}` (existing guard at `internal/hooks/dispatch.go:60`).
- WHEN `cachedAll()` is called within `hookCacheTTL` (1 second) of the previous load THE SYSTEM SHALL CONTINUE TO return without re-reading disk, and `Add()` SHALL CONTINUE TO invalidate the cache before returning.
- WHEN the cache is empty or stale THE SYSTEM SHALL CONTINUE TO call `LoadAll()` and replace `r.cache` under the write lock; concurrent readers SHALL CONTINUE TO see the snapshot they took under the read lock.

## Root Cause (optional)

<!-- 简要分析根因，方便 design.md / tasks.md 落地修复 -->

`Hook.Disabled` was added in commit `4e1f45a` ("fix(hooks): Match TTL 缓存 + Hook.Disabled 显式字段 + 删 timeout=0 歧义") to replace the overloaded `Timeout=0` signal. The replacement was applied to `runShellAction` and the JSON normaliser but not to:
1. The `Match()` filter loop, which still only consults `Enabled`.
2. The `runAgentAction` branch, which has no `Disabled` check at all.

The cache shallow-copy concern is independent: `cachedAll` predates `Disabled`, but its doc comment overpromises immutability. The fix is to either honour the promise (deep copy) or downgrade the comment. This spec chooses deep copy to keep callers' mental model simple.

## Acceptance Criteria

- [AC-1] `Registry.Match(event, file)` returns no hook whose `Disabled` is true (verified by unit test with at least one `{enabled:true, disabled:true}` fixture and one `{enabled:false}` fixture asserting both are skipped).
- [AC-2] `runAgentAction(agentFn, hook)` where `hook.Disabled == true` returns `Result{OK: false, Error: "hook disabled"}` and does NOT call `agentFn` (verified by spy AgentFn that fails the test if invoked).
- [AC-3] Mutating any field on a hook returned from `Match()` does NOT change the value seen by a subsequent `Match()` call within the TTL window (verified by setting `Enabled=false` on a returned hook and re-asserting the next `Match()` still includes it).
- [AC-4] `go test -race ./internal/hooks/...` exits 0.
- [AC-5] `free-kiro lint hooks-disabled-gate` exits 0.