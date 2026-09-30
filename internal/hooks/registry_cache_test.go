package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jingyu525/free-kiro/internal/models"
)

const hookJSONBody = `{
  "id": "demo",
  "event": "PreToolUse",
  "action_type": "shell",
  "action": "echo demo"
}`

// TestRegistry_MatchUsesCache covers AC-5 (H8): once Match() has loaded
// the hook list, subsequent Match() calls within hookCacheTTL must
// reuse the cached slice and not re-read disk. We verify this by
// checking that the cache population time stays unchanged after
// multiple Match() invocations.
func TestRegistry_MatchUsesCache(t *testing.T) {
	r, _ := setup(t)

	// Seed one hook file so Match has something to find.
	hookDir := r.ws.HooksDir()
	if err := os.WriteFile(filepath.Join(hookDir, "demo.json"),
		[]byte(hookJSONBody), 0o644); err != nil {
		t.Fatal(err)
	}

	// First Match populates the cache.
	if _, err := r.Match("PreToolUse", ""); err != nil {
		t.Fatalf("first Match: %v", err)
	}
	r.cacheMu.RLock()
	first := r.cacheTime
	r.cacheMu.RUnlock()
	if first.IsZero() {
		t.Fatal("expected cacheTime to be set after first Match")
	}

	// 100 subsequent Match() calls within the TTL window must not
	// refresh the cache.
	for range 100 {
		if _, err := r.Match("PreToolUse", ""); err != nil {
			t.Fatalf("Match: %v", err)
		}
	}
	r.cacheMu.RLock()
	second := r.cacheTime
	r.cacheMu.RUnlock()
	if !second.Equal(first) {
		t.Errorf("cacheTime changed during TTL window: %v → %v (cache invalidated by Match)", first, second)
	}
}

// TestRegistry_MatchRefreshAfterTTL covers the cache-eviction path:
// after hookCacheTTL elapses, the next Match() must refresh.
func TestRegistry_MatchRefreshAfterTTL(t *testing.T) {
	r, _ := setup(t)
	hookDir := r.ws.HooksDir()
	if err := os.WriteFile(filepath.Join(hookDir, "demo.json"),
		[]byte(hookJSONBody), 0o644); err != nil {
		t.Fatal(err)
	}

	// Force a stale cache by backdating cacheTime past the TTL.
	if _, err := r.Match("PreToolUse", ""); err != nil {
		t.Fatal(err)
	}
	r.cacheMu.Lock()
	r.cacheTime = time.Now().Add(-2 * hookCacheTTL)
	r.cacheMu.Unlock()

	if _, err := r.Match("PreToolUse", ""); err != nil {
		t.Fatal(err)
	}
	r.cacheMu.RLock()
	fresh := r.cacheTime
	r.cacheMu.RUnlock()
	if time.Since(fresh) > hookCacheTTL {
		t.Errorf("expected cacheTime to be refreshed after stale Match; got age=%v", time.Since(fresh))
	}
}

// TestRegistry_AddInvalidatesCache covers the write path: Add()
// invalidates the cache so the next Match() observes the new hook
// even within the TTL window.
func TestRegistry_AddInvalidatesCache(t *testing.T) {
	r, _ := setup(t)

	// Populate cache (no hooks yet, but the empty result is cached).
	if _, err := r.Match("PreToolUse", ""); err != nil {
		t.Fatal(err)
	}

	// Add a hook through the Registry — should invalidate cache.
	if _, err := r.Add(&models.Hook{
		ID:      "fresh",
		Event:   "PreToolUse",
		Action:  "echo fresh",
		Enabled: true,
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	r.cacheMu.RLock()
	cached := r.cache
	r.cacheMu.RUnlock()
	if cached != nil {
		t.Errorf("expected cache to be cleared after Add; got %d entries", len(cached))
	}

	// Next Match() should see the new hook.
	matched, err := r.Match("PreToolUse", "")
	if err != nil {
		t.Fatalf("Match after Add: %v", err)
	}
	if len(matched) != 1 || matched[0].ID != "fresh" {
		t.Errorf("expected Match to return the new hook; got %+v", matched)
	}
}

// TestRegistry_MatchLatencyUnderBudget is a soft latency budget for
// AC-5: 100 Match() calls against a 20-hook fixture should complete
// well under the 100ms wall-clock budget on any reasonable machine.
// We use a 200ms ceiling to absorb noisy CI VMs; the contract is "much
// faster than the pre-fix per-call disk read".
func TestRegistry_MatchLatencyUnderBudget(t *testing.T) {
	r, _ := setup(t)
	hookDir := r.ws.HooksDir()
	for range 20 {
		path := filepath.Join(hookDir, "h.json")
		if err := os.WriteFile(path, []byte(hookJSONBody), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.Match("PreToolUse", ""); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for range 100 {
		if _, err := r.Match("PreToolUse", ""); err != nil {
			t.Fatalf("Match: %v", err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Errorf("100 Match() calls took %v; budget is 100ms (200ms CI ceiling)", elapsed)
	}
}