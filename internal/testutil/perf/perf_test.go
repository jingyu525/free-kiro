package perf

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// captureTB records the first call to Fatalf and forwards nothing to
// the underlying *testing.T — used to observe perf.Load's bound check
// without aborting the surrounding test. Implements testing.TB.
type captureTB struct {
	testing.TB
	fatalfCalled bool
	fatalfMsg    string
}

func (c *captureTB) Fatalf(format string, args ...any) {
	c.fatalfCalled = true
	c.fatalfMsg = format
	if len(args) > 0 {
		// Cheap formatting: we only care about whether it ran.
		c.fatalfMsg = format + " (with args)"
	}
	// Do not forward to the real testing.TB — we want Load to return
	// gracefully so the test can assert on the captured state.
}

func TestLoad_ReturnsNonEmptyBytes(t *testing.T) {
	for _, size := range []Size{Small, Medium, Large} {
		t.Run(sizeName(size), func(t *testing.T) {
			data := Load(t, size)
			if len(data) == 0 {
				t.Fatalf("fixture must be non-empty for size %d", size)
			}
			if !strings.HasPrefix(string(data), "# perf-") {
				t.Errorf("fixture must start with H1 title, got %q",
					string(data[:min(20, len(data))]))
			}
		})
	}
}

func TestLoad_ContainsMatchingACCount(t *testing.T) {
	// Cross-check that the fixture on disk actually contains the
	// declared number of AC entries. Catches the case where the
	// generator drifts from the expected density.
	cases := []struct {
		size     Size
		wantMin  int
		wantName string
	}{
		{Small, 10, "small"},
		{Medium, 100, "medium"},
		{Large, 1000, "large"},
	}
	for _, tc := range cases {
		t.Run(tc.wantName, func(t *testing.T) {
			data := Load(t, tc.size)
			newlineHits := bytes.Count(data, []byte("\n[AC-"))
			leadingHits := 0
			if bytes.HasPrefix(data, []byte("[AC-")) {
				leadingHits = 1
			}
			total := newlineHits + leadingHits
			if total < tc.wantMin {
				t.Errorf("fixture must contain ≥ %d AC lines, got %d",
					tc.wantMin, total)
			}
		})
	}
}

func TestLoad_CachedAcrossCalls(t *testing.T) {
	// Two consecutive Load calls for the same size MUST return the
	// same backing array — that's the cache contract.
	a := Load(t, Medium)
	b := Load(t, Medium)
	if len(a) == 0 || len(b) == 0 {
		t.Fatalf("Load returned empty bytes")
	}
	if &a[0] != &b[0] {
		t.Errorf("Load must return the same backing array on repeated calls")
	}
}

func TestReset_ReReadsFromDisk(t *testing.T) {
	// After Reset(), the cache pointer goes nil and a fresh ReadFile
	// runs. We assert the slice identity changed; bytes could happen
	// to be identical even on a re-read, but the pointer must differ.
	first := Load(t, Small)
	firstPtr := &first[0]

	Reset()

	second := Load(t, Small)
	secondPtr := &second[0]

	if firstPtr == secondPtr {
		t.Fatalf("Reset() did not invalidate cache: same backing array")
	}
}

func TestLoad_FatalOnUnknownSize(t *testing.T) {
	// testing.TB.Fatalf aborts the calling goroutine via runtime.Goexit,
	// not panic — recover() never sees it. Wrap a mock TB that captures
	// the call instead so the surrounding test can keep running.
	c := &captureTB{TB: t}
	Load(c, Size(99))
	if !c.fatalfCalled {
		t.Errorf("Load(Size(99)) must invoke tb.Fatalf, got nothing")
	}
}

func TestFixturePath_IsAbsolute(t *testing.T) {
	for _, size := range []Size{Small, Medium, Large} {
		t.Run(sizeName(size), func(t *testing.T) {
			p := FixturePath(size)
			if !filepath.IsAbs(p) {
				t.Errorf("FixturePath must return an absolute path, got %q", p)
			}
			if !strings.HasSuffix(p, ".md") {
				t.Errorf("FixturePath must end in .md, got %q", p)
			}
		})
	}
}

func sizeName(s Size) string {
	switch s {
	case Small:
		return "small"
	case Medium:
		return "medium"
	case Large:
		return "large"
	default:
		return "unknown"
	}
}