package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// hookCacheTTL bounds how long Match() will reuse a previously-loaded
// hook list before forcing a refresh. Short enough that authors see
// their edits within a second; long enough that high-frequency
// PreToolUse invocations don't re-read disk on every Save.
const hookCacheTTL = 1 * time.Second

// Registry loads, normalises, matches, and dispatches hooks. Each
// .kiro/hooks/*.json file may contain a single hook (flat), an array of
// hooks, or a Kiro v1 envelope; all forms are accepted.
//
// Match() consults a TTL cache (hookCacheTTL) so the host can call it
// repeatedly without re-reading every JSON file on disk. Add()
// invalidates the cache so a freshly-written hook takes effect on the
// next Match().
type Registry struct {
	ws *workspace.Workspace

	cacheMu   sync.RWMutex
	cache     []*models.Hook // last LoadAll result; nil ⇒ not yet filled
	cacheTime time.Time      // wall-clock when cache was filled
}

// NewRegistry constructs a registry rooted at the given workspace.
func NewRegistry(ws *workspace.Workspace) *Registry {
	return &Registry{ws: ws}
}

// LoadAll reads every JSON file in .kiro/hooks and returns the
// normalised Hook records. Invalid files raise HookError.
func (r *Registry) LoadAll() ([]*models.Hook, error) {
	dir := r.ws.HooksDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ferrors.Wrap("hooks.load", err, "read "+dir)
	}
	var out []*models.Hook
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, ent.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, ferrors.Wrap("hooks.load", err, "read "+path)
		}
		// Detect shape: object → single or envelope; array → list of flat.
		var probe any
		if err := json.Unmarshal(data, &probe); err != nil {
			return nil, &ferrors.HookError{KiroError: ferrors.Wrap("hooks.load", err, "parse "+path)}
		}
		switch v := probe.(type) {
		case []any:
			// Top-level array: each element is a flat hook.
			for _, item := range v {
				rm, ok := item.(map[string]any)
				if !ok {
					return nil, ferrors.New("hooks.load", path+": array element is not an object")
				}
				h, err := normaliseHook(rawHook(rm))
				if err != nil {
					return nil, ferrors.Wrap("hooks.load", err, path)
				}
				out = append(out, h)
			}
		case map[string]any:
			rm := rawHook(v)
			if _, ok := rm["hooks"]; ok {
				// Kiro v1 envelope.
				var env envelopeFile
				if err := json.Unmarshal(data, &env); err != nil {
					return nil, ferrors.Wrap("hooks.load", err, path)
				}
				for _, h := range env.Hooks {
					nh, err := normaliseHook(h)
					if err != nil {
						return nil, ferrors.Wrap("hooks.load", err, path)
					}
					out = append(out, nh)
				}
			} else {
				// Single flat hook.
				h, err := normaliseHook(rm)
				if err != nil {
					return nil, ferrors.Wrap("hooks.load", err, path)
				}
				out = append(out, h)
			}
		default:
			return nil, ferrors.New("hooks.load", path+": unsupported JSON shape")
		}
	}
	return out, nil
}

// Match returns the hooks that should fire for the given event + file
// path. Disabled hooks are skipped. Filters are matched in the order:
// regex (Kiro matcher) → glob (free-kiro glob).
//
// Results come from a hookCacheTTL-bounded cache so a sequence of
// Match() calls without underlying file changes reuses the disk read.
// Add() invalidates the cache.
func (r *Registry) Match(event string, file string) ([]*models.Hook, error) {
	all, err := r.cachedAll()
	if err != nil {
		return nil, err
	}
	var matched []*models.Hook
	for _, h := range all {
		if !h.Enabled {
			continue
		}
		if h.Event != event {
			continue
		}
		if h.Glob == "" {
			matched = append(matched, h)
			continue
		}
		if file == "" {
			continue
		}
		if h.IsRegex {
			if rx, err := regexp.Compile(h.Glob); err == nil && rx.MatchString(file) {
				matched = append(matched, h)
			}
		} else {
			if globMatch(h.Glob, file) {
				matched = append(matched, h)
			}
		}
	}
	return matched, nil
}

// cachedAll returns the cached hook list if it's still within the TTL
// window; otherwise it refreshes and replaces the cache. The fast path
// holds a read lock; the slow path holds the write lock and runs the
// disk read.
func (r *Registry) cachedAll() ([]*models.Hook, error) {
	r.cacheMu.RLock()
	if r.cache != nil && time.Since(r.cacheTime) < hookCacheTTL {
		// Copy the slice so callers can't mutate our cache under a
		// stale read lock.
		out := make([]*models.Hook, len(r.cache))
		copy(out, r.cache)
		r.cacheMu.RUnlock()
		return out, nil
	}
	r.cacheMu.RUnlock()

	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	// Re-check inside the write lock — another goroutine may have
	// refreshed while we were upgrading.
	if r.cache != nil && time.Since(r.cacheTime) < hookCacheTTL {
		out := make([]*models.Hook, len(r.cache))
		copy(out, r.cache)
		return out, nil
	}
	all, err := r.LoadAll()
	if err != nil {
		return nil, err
	}
	r.cache = all
	r.cacheTime = time.Now()
	return all, nil
}

// invalidateCache clears the cached hook list so the next Match() will
// re-read disk. Called by Add() after a successful write.
func (r *Registry) invalidateCache() {
	r.cacheMu.Lock()
	r.cache = nil
	r.cacheTime = time.Time{}
	r.cacheMu.Unlock()
}

// Add writes one hook as a Kiro-compatible v1 envelope. The returned
// path is .kiro/hooks/<id>.json.
func (r *Registry) Add(h *models.Hook) (string, error) {
	if h.ID == "" {
		return "", ferrors.New("hooks.add", "hook requires an id")
	}
	if err := os.MkdirAll(r.ws.HooksDir(), 0o755); err != nil {
		return "", ferrors.Wrap("hooks.add", err, "mkdir")
	}
	data, err := encodeEnvelope(h)
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(r.ws.HooksDir(), h.ID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", ferrors.Wrap("hooks.add", err, "write "+path)
	}
	// Successful write invalidates the cache so the next Match() picks
	// up the new hook. Failure paths don't invalidate because the
	// disk state hasn't actually changed.
	r.invalidateCache()
	return path, nil
}
