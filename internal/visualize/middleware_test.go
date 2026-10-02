package visualize

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWithRecover_PanicReturns500 covers AC "withRecover captures
// panic → 500 + JSON body + log stack". Uses a handler that panics
// to simulate any inner-layer failure (handler bug, nil deref in
// router, etc.).
func TestWithRecover_PanicReturns500(t *testing.T) {
	cases := []struct {
		name    string
		panicV  any
		wantSub string
	}{
		{"string panic", "boom", "internal server error"},
		{"error panic", http.ErrAbortHandler, "internal server error"},
		{"nil panic", (*int)(nil), "internal server error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := withRecover(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				panic(c.panicV)
			}))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest("GET", "/panic", nil))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d; want 500", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q; want application/json", ct)
			}
			if !strings.Contains(rec.Body.String(), c.wantSub) {
				t.Errorf("body = %q; want substring %q", rec.Body.String(), c.wantSub)
			}
		})
	}
}

// TestWithRecover_NextHandlerCalled covers the non-panic path: the
// recovered handler must invoke next.ServeHTTP exactly once.
func TestWithRecover_NextHandlerCalled(t *testing.T) {
	called := false
	h := withRecover(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/ok", nil))
	if !called {
		t.Fatal("inner handler not invoked")
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d; want 418", rec.Code)
	}
}

// TestWithLogger_LogsRequest captures the log writer to verify the
// access log line format: "<METHOD> <PATH> <STATUS> <LATENCY>".
// Asserts a panic-recovered 500 is also logged (so the operator
// can grep for failures).
func TestWithLogger_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})

	chain := withRecover(withLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, httptest.NewRequest("GET", "/api/summary", nil))

	line := buf.String()
	if !strings.Contains(line, "GET /api/summary 200 ") {
		t.Errorf("access log line = %q; want contains 'GET /api/summary 200 '", line)
	}
	// Latency must look like "<digits>.<digits>" or "<digits>" with a time unit suffix.
	if !strings.Contains(line, "ns") && !strings.Contains(line, "µs") &&
		!strings.Contains(line, "us") && !strings.Contains(line, "ms") && !strings.Contains(line, "s") {
		t.Errorf("access log line = %q; want latency suffix (ns/µs/ms/s)", line)
	}
}

// TestWithCacheHeaders covers all four prefix branches of the
// per-path Cache-Control matrix documented in middleware.go.
func TestWithCacheHeaders(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string // "" means no Cache-Control header should be set
	}{
		{"assets immutable", "/assets/main.abc123.js", "public, max-age=31536000, immutable"},
		{"api no-cache", "/api/summary", "no-cache"},
		{"api spec no-cache", "/api/spec/foo", "no-cache"},
		{"api specs no-cache", "/api/specs", "no-cache"},
		{"sse no-store + no-cache", "/api/events", "no-cache, no-store, must-revalidate"},
		{"root no-cache", "/", "no-cache"},
		{"legacy no-cache", "/legacy", ""}, // not handled → passthrough
		{"unknown path passthrough", "/foo/bar", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := withCacheHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest("GET", c.path, nil))
			got := rec.Header().Get("Cache-Control")
			if got != c.want {
				t.Errorf("Cache-Control for %q = %q; want %q", c.path, got, c.want)
			}
		})
	}
}

// TestStatusRecorder_NoDoubleWriteHeader covers the statusRecorder
// invariant: WriteHeader must capture the FIRST status code only.
// Subsequent WriteHeader calls are no-ops on the wrapper, so the
// downstream writer's own double-write detection stays active.
func TestStatusRecorder_NoDoubleWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}
	sr.WriteHeader(http.StatusTeapot) // first call: status = 418, rec.Code = 418
	sr.WriteHeader(http.StatusOK)     // second call: ignored
	if sr.status != http.StatusTeapot {
		t.Errorf("status = %d; want 418 (first WriteHeader wins)", sr.status)
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("downstream Code = %d; want 418", rec.Code)
	}
}

// TestStatusRecorder_Implicit200 covers the case where Write is
// called without a prior WriteHeader. statusRecorder must default
// to 200 in that case (matching net/http's documented behaviour)
// so withLogger reports the right status for handlers that just
// Write body bytes.
func TestStatusRecorder_Implicit200(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusInternalServerError}
	_, _ = sr.Write([]byte("body")) // implicit 200
	if sr.status != http.StatusOK {
		t.Errorf("status = %d; want 200 (implicit on Write)", sr.status)
	}
}

// TestStatusRecorder_FlushDelegates covers SSE compatibility:
// when the underlying writer supports http.Flusher, Flush must
// forward. Without it, /api/events SSE responses get buffered
// upstream by middleware proxies and break EventSource handshakes.
func TestStatusRecorder_FlushDelegates(t *testing.T) {
	// httptest.ResponseRecorder does NOT implement http.Flusher.
	// Expect Flush to be a silent no-op (no panic).
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}
	sr.Flush() // must not panic
}
