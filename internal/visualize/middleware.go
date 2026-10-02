// Package visualize — middleware.go: HTTP middleware chain for the
// dashboard server. Three composable middlewares wrap the mux:
//
//	withRecover     → defer recover(); panic → 500 + JSON body + log stack
//	withLogger      → access log (method/path/status/latency) with statusRecorder
//	withCacheHeaders → per-prefix Cache-Control (assets 1y immutable,
//	                   api no-cache, events no-store + no-cache)
//
// Assembly order in NewServer: withRecover(withLogger(withCacheHeaders(mux)))
// — outer recover catches panics from any inner layer including the
// logger itself, so a panic mid-logging still surfaces as a 500 instead
// of crashing the http.Server.
package visualize

import (
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
