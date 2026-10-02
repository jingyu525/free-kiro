// Package visualize — server.go: minimal HTTP dashboard for free-kiro.
//
// Provides:
//
//	GET /              → static HTML (single-page app)
//	GET /api/summary   → JSON of the project report
//	GET /api/specs      → JSON list of all specs
//	GET /api/spec/<n>   → JSON detail for one spec
//
// All HTML/JS/CSS is embedded via embed.FS so the binary stays
// self-contained. The server uses only stdlib (net/http) — no
// third-party deps. Data refresh uses two complementary mechanisms:
// the polling fallback (`/api/summary` re-fetched every 5 seconds by
// the SPA) and the SSE broadcast endpoint (`/api/events`, implemented
// in server_sse.go) driven by a 2-second file mtime watcher. fsnotify
// is a future enhancement for sub-second updates.
//
// File layout (Wave 5 refactor):
//
//	server.go         — Server struct + lifecycle (NewServer / Start /
//	                    ServeWith / Shutdown) + JSON handlers
//	                    (handleIndex / Summary / Specs / Spec) +
//	                    static embed FS
//	server_sse.go     — SSE handler + fs watcher (broadcast system)
//	server_static.go  — browser-open helpers + small utilities
package visualize

import (
	"context"
	"embed"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jingyu525/free-kiro/internal/spec"
)

//go:embed static/*
//go:embed all:static/dist/assets/*
var staticFS embed.FS

// shutdownTimeout caps how long Shutdown() waits for the watcher and
// active SSE connections to drain. Picked so CI / single-tab shutdown
// completes quickly while multi-tab browser clients have a chance to
// receive their final refresh.
const shutdownTimeout = 5 * time.Second

// Server serves the dashboard over HTTP.
type Server struct {
	addr string
	ws   WorkspacePaths
	eng  *spec.Engine
	srv  *http.Server
	stop chan struct{} // signal watcher goroutine to exit

	// SSE fan-out state. previously a package-level global with no
	// locking — concurrent browser clients triggered -race findings
	// (S1 fix). Moving to per-Server fields keeps multi-tenant
	// instances independent and serialises access under notifierMu.
	notifierMu  sync.Mutex
	subscribers []chan struct{}

	// watcherOnce ensures watchChanges() is started exactly once per
	// Server lifetime (S8 fix). watcherRunning is closed by the
	// watcher when it exits so Shutdown() can wait for it.
	watcherOnce    sync.Once
	watcherRunning chan struct{}

	// watcherStartCount is incremented each time watchChanges actually
	// runs (i.e. once on the first handleEvents call, never again).
	// Tests read this to assert the S8 dedup invariant without poking
	// at unexported runtime state.
	watcherStartCount atomic.Int32

	// startedAt records NewServer wall-clock time so /api/health can
	// report uptime. lastRefreshAt is updated each time the file
	// watcher broadcasts a refresh event so health also surfaces
	// "last successful refresh" for monitoring integrations.
	startedAt      time.Time
	lastRefreshAt atomic.Int64
}

// WorkspacePaths is the subset of workspace.Workspace the Server needs.
// Defined here (not in report.go) because the Server package is the
// primary caller — the report.go package accepts an even thinner
// interface (only ReadCurrent) since the Server has already done the
// path resolution.
type WorkspacePaths interface {
	KiroDir() string
	SpecDir(name string) string
	Root() string
	ReadCurrent() string
	KiroDirExists() bool
}

// NewServer constructs a Server with the dashboard's HTTP routes wired.
// The SSE endpoint + file watcher are added in handleRoutes.
func NewServer(addr string, ws WorkspacePaths, eng *spec.Engine) *Server {
	s := &Server{
		addr:           addr,
		ws:             ws,
		eng:            eng,
		stop:           make(chan struct{}),
		watcherRunning: make(chan struct{}),
		startedAt:      time.Now().UTC(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/specs", s.handleSpecs)
	mux.HandleFunc("/api/spec/", s.handleSpec)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/hooks", s.handleHooks)
	mux.HandleFunc("/assets/", s.handleStatic)

	// Middleware chain: recover → logger → cache → mux. Order matters:
	// recover outermost catches panics from any inner layer (including
	// the logger itself), logger records every request including 500s
	// produced by recover, cache attaches headers before the first
	// WriteHeader (after which headers are locked).
	handler := withRecover(withLogger(withCacheHeaders(withETag(mux))))
	s.srv = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return s
}

// Start runs the HTTP listener. The caller passes a net.Listener
// (built externally so port resolution + friendly URL printing stays
// in the CLI layer).
func (s *Server) Start() error { return s.srv.ListenAndServe() }

// ServeWith runs the server on an already-bound listener.
func (s *Server) ServeWith(ln net.Listener) error { return s.srv.Serve(ln) }

// Shutdown gracefully stops the HTTP server and the file-watcher
// goroutine spawned by handleEvents.
//
// Implementation note (V1 fix): previous version blocked forever on
// `<-s.done` when watchChanges had never been started (no SSE client
// ever connected). We now wait for watcherRunning with a deadline so
// Shutdown() always returns within shutdownTimeout even when the
// watcher was never spawned. We also use http.Server.Shutdown(ctx)
// instead of Close() so active SSE connections receive a clean
// close instead of being dropped mid-event.
func (s *Server) Shutdown() error {
	select {
	case <-s.stop:
		// already closed
	default:
		close(s.stop)
	}
	// Drain the watcher with a bounded wait. watcherRunning is closed
	// by the watcher's defer when it exits, so a never-started watcher
	// would hang here forever; the timer fallback guarantees
	// Shutdown() returns within shutdownTimeout. We use NewTimer +
	// defer Stop instead of time.After so the timer is released as
	// soon as watcherRunning fires (zero GC pressure on the hot
	// shutdown path).
	t := time.NewTimer(shutdownTimeout)
	defer t.Stop()
	select {
	case <-s.watcherRunning:
	case <-t.C:
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// Addr returns the bound TCP address (host:port) of the dashboard
// server. Useful for `openBrowser` and status output.
func (s *Server) Addr() string { return s.srv.Addr }

// URL returns the dashboard's full HTTP URL, suitable for opening in a
// browser (e.g. `open http://127.0.0.1:8080`). Handles IPv6 literals
// (e.g. `[::1]:8080`) by wrapping the host in brackets; falls back to
// returning the raw `Addr` if it isn't a valid `host:port` pair.
func (s *Server) URL() string {
	host, port, err := net.SplitHostPort(s.srv.Addr)
	if err != nil {
		return "http://" + s.srv.Addr
	}
	if strings.Contains(host, ":") {
		// IPv6 literal — must be wrapped in [ ] per RFC 3986.
		return "http://[" + host + "]:" + port
	}
	return "http://" + host + ":" + port
}

// writeJSON encodes v as JSON and writes it to w with a 200 status.
// Errors during encoding surface as a 500 with the error message in
// the response body (caller logs the error separately).
func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(data)
}

// handleIndex serves the embedded single-page dashboard.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// handleSummary returns the JSON snapshot used by the polling client.
func (s *Server) handleSummary(w http.ResponseWriter, _ *http.Request) {
	report, err := BuildReport(s.ws, s.eng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

// handleSpecs returns one row per spec.
func (s *Server) handleSpecs(w http.ResponseWriter, _ *http.Request) {
	specs, err := s.eng.StatusForList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"specs": specs})
}

// handleSpec is the router for /api/spec/<name> and its sub-routes
// /api/spec/<name>/{tasks,drift,timeline}. Distinguishes by counting
// path segments after the prefix. Sub-routes are forwarded to
// dedicated handlers in server_handlers.go; the bare /api/spec/<name>
// shape returns the legacy status map (preserves dashboard-sse-bugfix
// + `free-kiro status --json` consumers).
func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/spec/")
	parts := strings.SplitN(rest, "/", 2)
	name := parts[0]
	if name == "" {
		http.Error(w, "missing spec name", http.StatusBadRequest)
		return
	}
	if len(parts) == 1 {
		status, err := s.eng.Status(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, status)
		return
	}
	// Sub-route: dispatch by second segment.
	switch parts[1] {
	case "tasks":
		s.handleSpecTasks(w, r)
	case "drift":
		s.handleSpecDrift(w, r)
	case "timeline":
		s.handleSpecTimeline(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleStatic serves files embedded under static/dist/ at /assets/*.
// Built artifacts (main.<hash>.js / main.<hash>.css) live there after
// the dashboard-frontend-foundation spec's `make dashboard-dist`
// produces them. Returns 400 on path-traversal attempts, 404 on
// missing files, 405 on non-GET methods.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/assets/")
	if rel == "" || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
		return
	}
	data, err := staticFS.ReadFile("static/dist/" + rel)
	if err != nil {
		// Vite emits bundled assets under dist/assets/<rel> when base is /assets/.
		// Fall back to that subdirectory lookup so legacy handleStatic callers
		// see hashed bundles without rewriting the existing route.
		data, err = staticFS.ReadFile("static/dist/assets/" + rel)
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "asset not found"})
		return
	}
	w.Header().Set("Content-Type", contentTypeForAsset(rel))
	_, _ = w.Write(data)
}

// contentTypeForAsset picks a Content-Type by file extension. Keeps
// the lookup table next to handleStatic so adding a new asset type
// is a one-place change.
func contentTypeForAsset(rel string) string {
	switch {
	case strings.HasSuffix(rel, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(rel, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(rel, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(rel, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(rel, ".png"):
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
