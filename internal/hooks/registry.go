package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ferrors "github.com/liujingyu/free-kiro/internal/errors"

	"github.com/liujingyu/free-kiro/internal/models"
	"github.com/liujingyu/free-kiro/internal/workspace"
)

// Registry loads, normalises, matches, and dispatches hooks. Each
// .kiro/hooks/*.json file may contain a single hook (flat), an array of
// hooks, or a Kiro v1 envelope; all forms are accepted.
type Registry struct {
	ws *workspace.Workspace
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
			return nil, &ferrors.HookError{ferrors.Wrap("hooks.load", err, "parse "+path)}
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
func (r *Registry) Match(event string, file string) ([]*models.Hook, error) {
	all, err := r.LoadAll()
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
	return path, nil
}