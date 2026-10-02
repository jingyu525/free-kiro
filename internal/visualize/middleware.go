// Package visualize — middleware.go: HTTP middleware chain for the
// dashboard server. Four composable middlewares wrap the mux:
//
//	withRecover     → defer recover(); panic → 500 + JSON body + log stack
//	withLogger      → access log (method/path/status/latency) with statusRecorder
//	withCacheHeaders → per-prefix Cache-Control (assets 1y immutable,
//	                   api no-cache, events no-store + no-cache)
//	withETag        → strong ETag via SHA-256 over response body; 304 on
//	                   If-None-Match match; skipped for /api/events (SSE
//	                   is a never-ending stream that can't be buffered)
//
// Assembly order in NewServer: withRecover(withLogger(withCacheHeaders(withETag(mux))))
// — outer recover catches panics from any inner layer including the
// logger itself, so a panic mid-logging still surfaces as a 500 instead
// of crashing the http.Server.
package visualize

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// withRecover wraps next in a panic guard. On panic: logs stack,
// writes 500 + JSON body, returns. http.Server stays alive for the
// next request.
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, debug.Stack())
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code
// for withLogger. Implements http.Flusher by delegating to the
// underlying writer when supported (required for /api/events SSE
// compatibility — without it EventSource handshakes stall).
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.status = http.StatusOK
		r.wroteHeader = true
	}
	return r.ResponseWriter.Write(b)
}

// Flush forwards to the underlying writer if it supports http.Flusher.
// The dashboard-realtime-fsnotify spec leans on this for SSE keep-alive
// heartbeats; without delegation, `Content-Type: text/event-stream`
// writes get buffered by middleware proxies and break the stream.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// withLogger writes one access-log line per request. Path is the
// request URL path with no query string (avoids leaking tokens).
// Assets get a terse line because they are hot and uninteresting.
func withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

// withCacheHeaders sets Cache-Control based on URL path prefix.
// Must run BEFORE next.ServeHTTP so headers are attached before the
// first WriteHeader (which would lock headers in). Three branches:
//
//	/assets/*  → immutable, 1y (built artifacts are content-hashed)
//	/api/events → no-cache + no-store (SSE proxies must not buffer)
//	/api/*, /  → no-cache (HTML/JSON reflect .kiro/ state)
//	other      → pass through (no Cache-Control header)
func withCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/assets/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case r.URL.Path == "/api/events":
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		case strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/":
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

// etagRecorder buffers the response body so withETag can compute a
// SHA-256 hash over it after the handler returns. Headers pass through
// to the underlying writer immediately so withCacheHeaders (which sets
// Cache-Control before the first byte) still works correctly.
//
// WriteHeader does NOT propagate to the underlying writer — withETag
// decides 200 vs 304 in its post-handler block and calls WriteHeader
// on the underlying directly. statusRecorder wraps us and forwards
// that decision through.
type etagRecorder struct {
	http.ResponseWriter
	body        bytes.Buffer
	status      int
	wroteHeader bool
}

func (r *etagRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
}

func (r *etagRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.status = http.StatusOK
		r.wroteHeader = true
	}
	return r.body.Write(b)
}

// Header returns the underlying writer's header map so handlers and
// upstream middleware (Cache-Control) keep working.
func (r *etagRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// withETag attaches a strong ETag (SHA-256 over response body) to
// every cacheable endpoint, and short-circuits with 304 Not Modified
// when the client's If-None-Match matches.
//
// The ETag is keyed by (path, content-fingerprint) so the response
// stays cacheable across requests even when the handler embeds
// volatile fields like `generated_at`. We compute a stable "content
// hash" by stripping known-volatile JSON keys (timestamps) before
// SHA-256 — but for unknown JSON shapes we fall back to a request-
// path-keyed cache keyed on a hash the server computed last time it
// saw an identical body length + status. That's enough for the
// dashboard where bodies change only when the underlying spec files
// change (so the cache invalidates naturally on next request).
//
// Skips:
//
//	- /api/events: SSE streams can never be buffered for an ETag.
//	- Assets: served from static FS with their own Cache-Control (immutable).
func withETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SSE / static assets: pass through untouched.
		if r.URL.Path == "/api/events" || strings.HasPrefix(r.URL.Path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}
		rec := &etagRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		body := rec.body.Bytes()
		// Compute a stable ETag. Volatile JSON keys (timestamps) are
		// stripped before hashing — see stripVolatileKeys. The hash is
		// memoised per (path, status) so two requests with identical
		// content (only differing in `generated_at`) get the same ETag.
		tag := stableETag(r.URL.Path, rec.status, body)

		w.Header().Set("ETag", `"`+tag+`"`)

		underlying := rec.ResponseWriter
		if rec.status >= 200 && rec.status < 300 && inmMatches(r.Header.Get("If-None-Match"), tag) {
			underlying.WriteHeader(http.StatusNotModified)
			return
		}
		underlying.WriteHeader(rec.status)
		_, _ = underlying.Write(body)
	})
}