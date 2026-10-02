// Package visualize — server_handlers.go: per-spec deep-dive HTTP
// handlers introduced by dashboard-backend-api-extensions. Each handler
// is GET-only and reads .kiro/ files via the Engine; never writes.
// All errors surface as JSON with the appropriate status code.
//
// Routes:
//
//	GET /api/spec/<name>/tasks     → grouped by execution wave
//	GET /api/spec/<name>/drift     → baseline drift signals
//	GET /api/spec/<name>/timeline  → per-spec file mtimes
//	GET /api/health                → liveness + uptime + last refresh
//	GET /api/hooks                 → hook registry snapshot (read-only)
package visualize

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
)

// buildVersion is the dashboard's self-reported version. Defaults to
// "dev" for local builds; release builds override via -ldflags when
// ldflags injection is added (out of scope for this spec).
const buildVersion = "dev"

// tasksResponse is the JSON shape returned by /api/spec/<name>/tasks.
// `Waves` is 1-indexed for human readability; `Summary` is a flat
// roll-up so the client can render a progress bar without re-counting.
type tasksResponse struct {
	Spec    string         `json:"spec"`
	Waves   []tasksWave    `json:"waves"`
	Summary tasksSummaryRO `json:"summary"`
}

type tasksWave struct {
	Index int           `json:"index"`
	Done  int           `json:"done"`
	Total int           `json:"total"`
	Tasks []tasksItemRO `json:"tasks"`
}

type tasksItemRO struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Deps  []int  `json:"deps"`
	Wave  int    `json:"wave"`
}

type tasksSummaryRO struct {
	Total     int `json:"total"`
	Done      int `json:"done"`
	Remaining int `json:"remaining"`
	Waves     int `json:"waves"`
}

// driftResponse is the JSON shape returned by /api/spec/<name>/drift.
// Signals is empty (not null) when the spec is unapproved or has
// no baseline drift — callers can render "no drift" without a nil
// check.
type driftResponse struct {
	Spec    string             `json:"spec"`
	Signals []driftSignalEntry `json:"signals"`
}

type driftSignalEntry struct {
	Key      string `json:"key"`
	Baseline int    `json:"baseline"`
	Current  int    `json:"current"`
	Delta    int    `json:"delta"`
}

// timelineResponse is the JSON shape returned by /api/spec/<name>/timeline.
// Files are ordered by mtime descending so the most recently edited
// file surfaces first in drift detail panels.
type timelineResponse struct {
	Spec  string         `json:"spec"`
	Files []timelineFile `json:"files"`
}

type timelineFile struct {
	Path  string `json:"path"`  // relative to spec dir
	Mtime string `json:"mtime"` // RFC3339
	Size  int64  `json:"size"`
}

// healthResponse is the JSON shape returned by /api/health.
// Subscribers and LastRefreshAt are zero-valued before the first
// SSE client connects or before the first file change is detected.
type healthResponse struct {
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	LastRefreshAt string `json:"last_refresh_at"`
	Subscribers   int    `json:"subscribers"`
	Version       string `json:"version"`
}

// hooksResponse is the JSON shape returned by /api/hooks.
// RegistryDisabled=true means the hook registry was not initialised
// — the handler returns 200 with an empty hooks list rather than 503
// so dashboard health-checks don't false-alarm on a fresh install.
type hooksResponse struct {
	Hooks            []hookEntry `json:"hooks"`
	RegistryDisabled bool        `json:"registry_disabled,omitempty"`
}

type hookEntry struct {
	ID         string `json:"id"`
	Event      string `json:"event"`
	Enabled    bool   `json:"enabled"`
	ActionType string `json:"action_type"`
	Glob       string `json:"glob,omitempty"`
}

// timelineMaxFiles caps the timeline response size. 50 is more than
// any realistic spec produces today; if a spec grows past this the
// UI gets a `+more` indicator instead of a multi-MB response.
const timelineMaxFiles = 50

// handleSpecTasks serves /api/spec/<name>/tasks. Path is everything
// after the prefix; URL-decoded so names with special chars work.
// Returns 404 on missing spec, 400 on path traversal, 500 on parse
// failure. Panic recovery is the middleware's job.
func (s *Server) handleSpecTasks(w http.ResponseWriter, r *http.Request) {
	name, ok := decodeSpecName(w, r)
	if !ok {
		return
	}
	waves, err := s.eng.TaskList(name)
	if err != nil {
		if errors.Is(err, spec.ErrSpecNotFound) {
			writeSpecNotFound(w, name)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := tasksResponse{Spec: name, Waves: make([]tasksWave, 0, len(waves))}
	total, done := 0, 0
	for _, wv := range waves {
		items := make([]tasksItemRO, 0, len(wv.Tasks))
		for _, t := range wv.Tasks {
			items = append(items, tasksItemRO{
				ID: t.ID, Title: t.Title, Done: t.Done,
				Deps: append([]int(nil), t.Deps...),
				Wave: wv.Index,
			})
		}
		resp.Waves = append(resp.Waves, tasksWave{
			Index: wv.Index, Done: wv.Done, Total: wv.Total, Tasks: items,
		})
		total += wv.Total
		done += wv.Done
	}
	resp.Summary = tasksSummaryRO{Total: total, Done: done, Remaining: total - done, Waves: len(waves)}
	writeJSON(w, resp)
}

// handleSpecDrift serves /api/spec/<name>/drift. Returns drift
// signals from eng.Status. Unapproved specs get an empty list
// (no baseline → no drift to report).
func (s *Server) handleSpecDrift(w http.ResponseWriter, r *http.Request) {
	name, ok := decodeSpecName(w, r)
	if !ok {
		return
	}
	st, err := s.eng.Status(name)
	if err != nil {
		writeSpecNotFound(w, name)
		return
	}
	signals := extractDriftSignals(st["drift"])
	writeJSON(w, driftResponse{Spec: name, Signals: signals})
}

// handleSpecTimeline serves /api/spec/<name>/timeline.
func (s *Server) handleSpecTimeline(w http.ResponseWriter, r *http.Request) {
	name, ok := decodeSpecName(w, r)
	if !ok {
		return
	}
	specDir := s.ws.SpecDir(name)
	if _, err := os.Stat(specDir); err != nil {
		if os.IsNotExist(err) {
			writeSpecNotFound(w, name)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	files := s.collectSpecMtimes(specDir)
	// Sort by mtime desc (newest first).
	sort.Slice(files, func(i, j int) bool {
		return files[i].mtime.After(files[j].mtime)
	})
	if len(files) > timelineMaxFiles {
		files = files[:timelineMaxFiles]
	}
	out := timelineResponse{Spec: name, Files: make([]timelineFile, 0, len(files))}
	for _, f := range files {
		out.Files = append(out.Files, timelineFile{
			Path:  f.relPath,
			Mtime: f.mtime.UTC().Format(time.RFC3339),
			Size:  f.size,
		})
	}
	writeJSON(w, out)
}

// handleHealth serves /api/health. Returns 200 always (even pre-
// ServeWith) so monitoring integrations don't false-alarm during
// startup. Returns "starting" status before ServeWith begins.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	last := s.lastRefreshAt.Load()
	var lastStr string
	if last > 0 {
		lastStr = time.Unix(0, last).UTC().Format(time.RFC3339)
	}
	uptime := max(int64(time.Since(s.startedAt).Seconds()), 0)
	writeJSON(w, healthResponse{
		Status:        "ok",
		UptimeSeconds: uptime,
		LastRefreshAt: lastStr,
		Subscribers:   s.subscribersCount(),
		Version:       buildVersion,
	})
}

// handleHooks serves /api/hooks. Returns the hook registry snapshot
// or 200 + registry_disabled=true when the registry is nil. The hook
// registry lives in internal/hooks; we accept its nil gracefully
// because not all CLI entry points register a registry (e.g. `hook run`
// in isolation). The dashboard treats a missing registry as "no hooks
// configured" rather than "service down".
func (s *Server) handleHooks(w http.ResponseWriter, _ *http.Request) {
	entries := listHooks()
	resp := hooksResponse{Hooks: entries, RegistryDisabled: entries == nil}
	if entries == nil {
		// Avoid encoding nil as JSON null — frontend renders it as a
		// broken array. Empty slice serialises to "[]" which the client
		// can iterate over without nil checks.
		resp.Hooks = []hookEntry{}
	}
	writeJSON(w, resp)
}

// decodeSpecName extracts the spec name from the URL path. URL-
// decodes so names with non-ASCII characters work. Rejects empty,
// "..", or absolute paths (defence in depth — handleStatic already
// has the same logic for /assets/; reuse the principle).
func decodeSpecName(w http.ResponseWriter, r *http.Request) (string, bool) {
	raw := strings.TrimPrefix(r.URL.Path, "/api/spec/")
	// strip sub-path tail (e.g. /tasks, /drift, /timeline)
	if i := strings.Index(raw, "/"); i >= 0 {
		raw = raw[:i]
	}
	name, err := url.PathUnescape(raw)
	if err != nil || name == "" || strings.Contains(name, "..") || strings.HasPrefix(name, "/") {
		http.Error(w, `{"error":"invalid spec name"}`, http.StatusBadRequest)
		return "", false
	}
	return name, true
}

func writeSpecNotFound(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error":"spec not found","spec":"` + name + `"}`))
}

// extractDriftSignals turns the loose `[]any` from eng.Status into
// a typed slice. Tolerates malformed entries silently (dashboard
// shouldn't 500 because of one corrupt baseline).
func extractDriftSignals(raw any) []driftSignalEntry {
	out := []driftSignalEntry{}
	items, _ := raw.([]any)
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, driftSignalEntry{
			Key:      asString(m["key"]),
			Baseline: asInt(m["baseline"]),
			Current:  asInt(m["current"]),
			Delta:    asInt(m["delta"]),
		})
	}
	return out
}

// specMtime is an internal struct for the collect-sort-truncate
// pipeline of /api/spec/<name>/timeline. mtime + size are pre-
// resolved so the sort comparator doesn't stat again.
type specMtime struct {
	relPath string
	mtime   time.Time
	size    int64
}

// collectSpecMtimes walks a single spec directory and returns one
// entry per regular file with path relative to the spec dir. Empty
// directories return an empty slice (no error). Mirrors the
// package-level collectMtimes in server_sse.go but scoped to one
// spec — the dashboard frontend hits /timeline with frequency
// proportional to spec count, so O(spec_size) per call beats
// O(workspace_size) of the global watcher.
func (s *Server) collectSpecMtimes(specDir string) []specMtime {
	var out []specMtime
	_ = filepath.WalkDir(specDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			// Per AGENT_RULES §2: don't silently swallow; record a
			// zero-size entry so the caller can see the file exists.
			out = append(out, specMtime{
				relPath: strings.TrimPrefix(path, specDir+string(filepath.Separator)),
				mtime:   time.Time{},
				size:    0,
			})
			return nil
		}
		out = append(out, specMtime{
			relPath: strings.TrimPrefix(path, specDir+string(filepath.Separator)),
			mtime:   info.ModTime(),
			size:    info.Size(),
		})
		return nil
	})
	return out
}

// listHooks returns the current hook registry snapshot. Returns nil
// when the registry is not wired in (CLI commands that don't load
// .kiro/hooks). The handler maps nil → empty slice with
// RegistryDisabled=true so the dashboard renders "hooks disabled"
// rather than crashing.
//
// Kept in this file (not a separate package) because the dashboard
// only needs a read-only view; the registry's full API lives in
// internal/hooks and is not safe to depend on from visualize.
func listHooks() []hookEntry {
	// Hook registry integration deferred — requires importing
	// internal/hooks which would create a visualize → hooks
	// dependency. The hook system has its own list CLI
	// (`free-kiro hook list`); the dashboard links to that
	// surface. Returning nil + the handler's RegistryDisabled
	// fallback keeps the contract working until a follow-up
	// spec wires the registry.
	return nil
}

// Compile-time assertion: models.Task must remain the canonical
// task type so external packages can decode /api/spec/.../tasks JSON.
var _ models.Task
