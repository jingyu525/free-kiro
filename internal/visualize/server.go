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
// third-party deps. Polling is the data refresh mechanism: clients
// re-fetch /api/summary every 5 seconds. A future enhancement can
// add fsnotify + SSE for real-time updates.
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
	"embed"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/jingyu525/free-kiro/internal/spec"
)

// osDirEntry aliases os.DirEntry for the package-level osReadDir
// indirection (so tests can stub if needed).
type osDirEntry = os.DirEntry

//go:embed static/*
var staticFS embed.FS

// Server serves the dashboard over HTTP.
type Server struct {
	addr string
	ws   WorkspacePaths
	eng  *spec.Engine
	srv  *http.Server
	stop chan struct{} // signal watcher goroutine to exit
	done chan struct{} // closed when watcher has exited
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
}

// NewServer constructs a Server with the dashboard's HTTP routes wired.
// The SSE endpoint + file watcher are added in handleRoutes.
func NewServer(addr string, ws WorkspacePaths, eng *spec.Engine) *Server {
	s := &Server{
		addr: addr,
		ws:   ws,
		eng:  eng,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/specs", s.handleSpecs)
	mux.HandleFunc("/api/spec/", s.handleSpec)
	mux.HandleFunc("/api/events", s.handleEvents)

	s.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
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
func (s *Server) Shutdown() error {
	select {
	case <-s.stop:
		// already closed
	default:
		close(s.stop)
	}
	// Wait briefly for the watcher to exit before tearing the HTTP
	// server down. Best-effort: 200 ms is enough on CI; production
	// shutdown is bounded by the OS anyway.
	<-s.done
	return s.srv.Close()
}

func (s *Server) Addr() string { return s.srv.Addr }

func (s *Server) URL() string {
	addr := s.srv.Addr
	return "http://" + portOnly(addr)
}

// loadWorkspace re-reads the workspace paths. Used by handlers that
// might run after the workspace has been re-rooted (rare).
func (s *Server) loadWorkspace() (WorkspacePaths, error) {
	return s.ws, nil
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

// handleSpec returns the full status of one named spec.
func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/spec/"):]
	if name == "" {
		http.Error(w, "missing spec name", http.StatusBadRequest)
		return
	}
	status, err := s.eng.Status(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, status)
}
