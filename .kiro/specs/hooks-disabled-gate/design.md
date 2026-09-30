# hooks-disabled-gate — Design

<!-- 描述技术架构与关键决策。这一段是给"自己"看的——30 天后回看能否立刻想起来为什么这么做 -->

## Architecture

Single-pass fix in `internal/hooks/`. No new package, no new public surface.

Three call sites change:

1. **`Registry.Match` (registry.go)** — extend the per-hook filter loop to also skip `Disabled`. This is the canonical gate; everything else is defence-in-depth.
2. **`runAgentAction` (dispatch.go)** — mirror the existing `runShellAction` guard at the top of the function. Same error string (`"hook disabled"`) so logs stay greppable across both action types.
3. **`Registry.cachedAll` (registry.go)** — replace `make + copy` (shallow slice header copy) with a `cloneHooks` helper that returns `[]*models.Hook` pointing to fresh `models.Hook` values. The helper lives next to `cachedAll` so the immutability guarantee stays local to the cache.

Defence-in-depth rationale: keeping `runAgentAction`'s guard means a future code path that constructs `Result`s without going through `Match()` (e.g. a manual `runAgentAction(nil, h)` from a unit test, or a new dispatcher in a downstream caller) still honours `Disabled`. The Match-layer fix is the user-visible behaviour; the dispatch-layer fix is the invariant.

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `Registry.Match` | Return hooks that fire for `(event, file)`, honouring both `Enabled` and `Disabled`. | `Match(event, file) ([]*models.Hook, error)` |
| `Registry.cachedAll` | TTL-bounded disk read returning deep-copied hooks. | `cachedAll() ([]*models.Hook, error)` |
| `cloneHooks` (new, unexported) | Deep-copy a `[]*models.Hook` slice so callers can't mutate cache. | `cloneHooks(in []*models.Hook) []*models.Hook` |
| `runShellAction` | Existing shell exec; keep existing `Disabled` guard unchanged. | unchanged |
| `runAgentAction` | Add mirror `Disabled` guard; otherwise unchanged. | unchanged otherwise |

## Data Model

`models.Hook` is unchanged. `Disabled bool` already exists from `4e1f45a`. No new fields, no migration.

`cloneHooks` allocates one new `*models.Hook` per element and one new slice header. The `Timeout` field is `*int`; the clone copies the pointer (cheap) rather than dereferencing-and-reallocating. This is intentional — callers don't mutate the pointed-to int, and copying the pointer preserves semantics. If a future caller needs to mutate it, we can revisit; today it's overcautious to allocate an extra `int`.

## Error Handling

No new error paths. The new `runAgentAction` guard returns `Result{OK: false, Error: "hook disabled"}` matching the existing `runShellAction` shape, so consumers that already filter on `Error == "hook disabled"` (CLI output, hooks report) keep working.

## Testing Strategy

| Test | Layer | What it asserts |
|---|---|---|
| `TestMatch_SkipsDisabled` | unit, registry | Two fixtures (`{enabled:true, disabled:true}`, `{enabled:false}`); both absent from `Match()` result. |
| `TestRunAgentAction_DisabledGuard` | unit, dispatch | Spy `AgentFn` records `calls`; with `Disabled=true` the spy stays at 0 and `Result.OK==false` with `"hook disabled"`. |
| `TestCachedAll_ReturnsIndependentCopies` | unit, registry | Call `Match()`, mutate `h.Enabled = false` on a returned hook, call `Match()` again within TTL, assert the original enabled hook still appears. |
| `TestCachedAll_Race` | existing | `go test -race` continues to pass; the new deep copy is allocation-only and adds no new shared state. |

All tests live next to the code they exercise (`internal/hooks/registry_test.go`, `internal/hooks/dispatch_test.go`). No new test helper packages.

## Migration / Rollout (if applicable)

None. The behavioural change is internal to `internal/hooks/`. No public CLI surface, no JSON schema change, no `.kiro/hooks/*.json` files affected. Existing users who set `disabled: true` will see their agent hooks now actually skip — this is the intended fix, not a regression.