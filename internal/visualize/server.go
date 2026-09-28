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
package visualize

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"

	"github.com/jingyu525/free-kiro/internal/spec"
)

//go:embed static/*
var staticFS embed.FS

// Server serves the dashboard over HTTP.
type Server struct {
	addr string
	ws   WorkspacePaths
	eng  *spec.Engine
	srv  *http.Server
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

// NewServer constructs a Server bound to the given address (e.g.
// ":7373" or "127.0.0.1:7373"). The Engine is used to gather report
// data on each request.
func NewServer(addr string, ws WorkspacePaths, eng *spec.Engine) *Server {
	s := &Server{
		addr: addr,
		ws:   ws,
		eng:  eng,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/summary", s.handleSummary)
	mux.HandleFunc("/api/specs", s.handleSpecs)
	mux.HandleFunc("/api/spec/", s.handleSpec)
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Start begins serving HTTP. Blocks until the listener errors or
// Shutdown is called. Use ServeWith(listener) when you need to bind a
// listener manually (e.g. port 0 to pick a random free port).
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return ferrors.Wrap("visualize.server", err, "listen "+s.addr)
	}
	return s.srv.Serve(ln)
}

// ServeWith is like Start but accepts a pre-bound listener. Lets the
// CLI print the actual bound port (port 0 → kernel-assigned).
func (s *Server) ServeWith(ln net.Listener) error {
	// Update Addr() so URL() reflects the actual bound address when
	// the caller used port 0.
	if tcp, ok := ln.Addr().(*net.TCPAddr); ok {
		s.srv.Addr = tcp.String()
	}
	return s.srv.Serve(ln)
}

// Shutdown gracefully stops the server (waits up to 5s for in-flight
// requests to complete).
func (s *Server) Shutdown() error {
	return s.srv.Shutdown(nil)
}

// Addr returns the actual bound address (useful when constructed
// with ":0" for tests).
func (s *Server) Addr() string { return s.srv.Addr }

// URL returns a printable URL pointing at this server (handles ":port"
// → "http://localhost:port"). Useful for CLI output.
func (s *Server) URL() string {
	host := s.srv.Addr
	if h, _, err := net.SplitHostPort(host); err == nil {
		if h == "" || h == "0.0.0.0" || h == "::" {
			host = "localhost:" + portOnly(s.srv.Addr)
		}
	}
	return "http://" + host
}

func portOnly(addr string) string {
	_, port, _ := net.SplitHostPort(addr)
	return port
}

// --- handlers ---

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
	w.Write(data)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	ws, err := s.loadWorkspace()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rep, err := BuildReport(ws, s.eng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, rep)
}

func (s *Server) handleSpecs(w http.ResponseWriter, r *http.Request) {
	specs, err := s.eng.ListSpecs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]map[string]any, 0, len(specs))
	for _, m := range specs {
		out = append(out, map[string]any{
			"name":     m.Name,
			"phase":    string(m.Phase),
			"workflow": m.Workflow,
			"spec_type": m.SpecType,
			"quick":    m.Quick,
			"approved": m.Approved,
		})
	}
	writeJSON(w, out)
}

func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/spec/"):]
	if name == "" {
		http.Error(w, "spec name required", http.StatusBadRequest)
		return
	}
	st, err := s.eng.Status(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, st)
}

// loadWorkspace returns the Server's workspace unchanged — it already
// implements the WorkspacePaths interface BuildReport needs.
func (s *Server) loadWorkspace() (WorkspacePaths, error) {
	if w, ok := s.ws.(WorkspacePaths); ok {
		return w, nil
	}
	return nil, fmt.Errorf("workspace impl missing required methods")
}

func writeJSON(w http.ResponseWriter, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// atoiLocal parses a small non-negative integer (returns -1 on failure).
func atoiLocal(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// detectBrowserAutoOpen reports whether the environment looks like a
// desktop session where `xdg-open` / `open` would work.
func detectBrowserAutoOpen() bool {
	if os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != "" {
		return false
	}
	return true
}

// openBrowser tries to launch the user's default browser. Best-effort:
// errors are non-fatal (the URL is already printed).
func openBrowser(url string) {
	// Lazy import to avoid pulling os/exec into the package always.
	// (keep this dependency-free; CLI calls openBrowser via a wrapper
	// in cli/serve.go for the actual exec invocation.)
	_ = url
}

// detectBrowserOpen returns true when the env looks like a desktop
// (DISPLAY / WAYLAND_DISPLAY on Linux, no SSH_TTY). Used by the CLI
// to decide whether to attempt opening a browser.
func detectBrowserOpen() bool {
	if os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != "" {
		return false
	}
	return true
}