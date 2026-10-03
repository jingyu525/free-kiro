// Package perf provides shared benchmark fixtures for free-kiro's
// hot-path benchmarks. The fixtures live in ../../testdata/perf/ at
// the repository root and are read once + cached at package load so
// `b.N` iterations reuse the same backing array (no IO noise inside
// the timed region).
//
// Three sizes are pre-baked:
//
//   - Small  → ~30 lines / 10 AC
//   - Medium → ~120 lines / 100 AC
//   - Large  → ~1000 lines / 1000 AC
//
// Benchmarks across `internal/lint`, `internal/spec`, `internal/visualize`
// etc. all import this package, so the path resolution is anchored at
// the location of this source file (runtime.Caller(0)) rather than the
// caller's working directory — `go test ./internal/lint/` and
// `go test ./internal/visualize/` must resolve to the same bytes.
package perf

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// Size selects a perf fixture by AC line count. Benchmarks typically
// use b.Run("small"/"medium"/"large", ...) subtests over the three
// predefined sizes.
type Size int

const (
	// Small holds 10 AC lines (typical draft spec).
	Small Size = iota
	// Medium holds 100 AC lines (typical mature spec).
	Medium
	// Large holds 1000 AC lines (stress test for lint engine).
	Large
)

// sizes maps Size to fixture filename. Kept as a map (not array) so that
// `sizes[size]` lookup panics with a clean "missing key" rather than a
// raw runtime bound-check on out-of-range Size values.
var sizes = map[Size]string{
	Small:  "small.md",
	Medium: "medium.md",
	Large:  "large.md",
}

// fixtureBasePath is computed once at package load from the location of
// this source file (internal/testutil/perf/perf.go). Three levels up
// is the repository root; testdata/perf/ holds the fixtures.
//
// Anchoring at the source file (not at runtime.Caller(1) of the caller)
// means the path is stable regardless of which package the benchmark
// lives in — `go test ./internal/lint/`, `./internal/spec/templates/`,
// and `./internal/visualize/` all resolve to the same bytes.
var fixtureBasePath = func() string {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		panic("perf: runtime.Caller(0) failed; cannot locate fixture root")
	}
	// internal/testutil/perf/perf.go → 3 dirs up = repo root.
	return filepath.Clean(filepath.Join(filepath.Dir(here), "..", "..", "..", "testdata", "perf"))
}()

var (
	cache   = map[Size][]byte{}
	loaders = map[Size]*sync.Once{}
)

func init() {
	for sz := range sizes {
		loaders[sz] = &sync.Once{}
	}
}

// Load returns the fixture bytes for size. The bytes are read from disk
// once (per Size) and shared across all subsequent calls — the same
// backing array is returned every time. Callers that mutate the slice
// must copy it first.
//
// Failures (missing file, unknown size) are reported via tb.Fatalf so
// the calling test/benchmark aborts cleanly. We deliberately do not
// return an error: every benchmark in this repository should fail fast
// if its fixture is missing, not silently produce zero data.
//
// Implementation note: cache and loaders are maps rather than arrays,
// so the path for an out-of-range Size falls into the map-miss branch
// (clean Fatalf) instead of a raw runtime "index out of range" panic
// that would be impossible for the surrounding test to recover from.
func Load(tb testing.TB, size Size) []byte {
	o, ok := loaders[size]
	if !ok {
		tb.Fatalf("perf: unknown Size %d", size)
		return nil
	}
	o.Do(func() {
		name := sizes[size] // safe: ok above implies name exists
		path := filepath.Join(fixtureBasePath, name)
		data, err := os.ReadFile(path)
		if err != nil {
			tb.Fatalf("perf: read fixture %s: %v", path, err)
		}
		cache[size] = data
	})
	return cache[size]
}

// Reset clears the in-memory cache so a subsequent Load re-reads from
// disk. Used by tests that mutate fixture bytes in place; production
// benchmarks never call this.
//
// Safe to call concurrently with Load from other goroutines — the
// sync.Once state is rebuilt atomically with the cache slice.
func Reset() {
	for sz := range sizes {
		loaders[sz] = &sync.Once{}
		cache[sz] = nil
	}
}

// FixturePath returns the absolute path to the named fixture. Useful
// for debug logging and for the rare benchmark that needs the path
// itself (e.g. fsnotify watching the file).
func FixturePath(size Size) string {
	return filepath.Join(fixtureBasePath, sizes[size])
}