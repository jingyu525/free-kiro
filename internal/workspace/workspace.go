// Package workspace owns the project root and its .kiro subdirectory layout.
//
// A Workspace is the directory that contains a .kiro folder. Commands resolve
// the nearest ancestor containing .kiro; if none exists, they fall back to
// the current directory so `free-kiro init` can bootstrap a fresh layout.
//
// The layout mirrors Kiro exactly:
//
//	.kiro/
//	├── settings.json   (created by init)
//	├── specs/<name>/   (one subdirectory per spec)
//	├── steering/       (project-scoped steering docs)
//	└── hooks/          (event-driven automations)
package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

const (
	// KiroDir is the hidden directory that marks a workspace root.
	KiroDir = ".kiro"
	// SpecsDirName lives inside KiroDir; one subdirectory per spec.
	SpecsDirName = "specs"
	// SteeringDirName lives inside KiroDir; .md files inside are steering docs.
	SteeringDirName = "steering"
	// HooksDirName lives inside KiroDir; .json files inside are hook defs.
	HooksDirName = "hooks"
	// SettingsFileName lives at the top of KiroDir; generator configuration.
	SettingsFileName = "settings.json"
)

// Workspace owns a project root and its .kiro subdirectory.
type Workspace struct {
	root string // absolute path
}

// New constructs a Workspace for an arbitrary directory. The directory does
// not need to contain .kiro yet (use Find for discovery, New for tests).
func New(root string) *Workspace {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root // fall back to the literal; Path ops will surface errors
	}
	return &Workspace{root: abs}
}

// Find walks up from start (default: cwd) looking for the nearest .kiro
// directory. If none is found, returns a Workspace rooted at start (so
// `free-kiro init` can bootstrap there).
func Find(start string) *Workspace {
	start = startOrCwd(start)
	cur, err := filepath.Abs(start)
	if err != nil {
		cur = start
	}
	for {
		if isDir(filepath.Join(cur, KiroDir)) {
			return &Workspace{root: cur}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached filesystem root — fall back to the original start.
			return &Workspace{root: start}
		}
		cur = parent
	}
}

func startOrCwd(s string) string {
	if s == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return cwd
	}
	return s
}

// Root returns the absolute project root (the directory that owns .kiro).
func (w *Workspace) Root() string { return w.root }

// KiroDir is <root>/.kiro.
func (w *Workspace) KiroDir() string { return filepath.Join(w.root, KiroDir) }

// SpecsDir is <root>/.kiro/specs.
func (w *Workspace) SpecsDir() string { return filepath.Join(w.root, KiroDir, SpecsDirName) }

// SteeringDir is <root>/.kiro/steering.
func (w *Workspace) SteeringDir() string { return filepath.Join(w.root, KiroDir, SteeringDirName) }

// HooksDir is <root>/.kiro/hooks.
func (w *Workspace) HooksDir() string { return filepath.Join(w.root, KiroDir, HooksDirName) }

// SettingsPath is <root>/.kiro/settings.json.
func (w *Workspace) SettingsPath() string { return filepath.Join(w.root, KiroDir, SettingsFileName) }

// SpecDir is <root>/.kiro/specs/<name>.
func (w *Workspace) SpecDir(name string) string { return filepath.Join(w.SpecsDir(), name) }

// Exists returns true if .kiro is present at the root.
func (w *Workspace) Exists() bool { return isDir(w.KiroDir()) }

// Require returns the workspace if .kiro exists; otherwise raises a
// WorkspaceError pointing the user at `free-kiro init`.
func (w *Workspace) Require() (*Workspace, error) {
	if !w.Exists() {
		return nil, ferrors.Wrap(
			"workspace.require",
			nil,
			"no .kiro workspace found at or above "+w.root+"; run `free-kiro init` first",
		)
	}
	return w, nil
}

// EnsureLayout creates .kiro, .kiro/specs, .kiro/steering, .kiro/hooks and
// writes a default settings.json if one does not already exist. Existing
// files are preserved.
func (w *Workspace) EnsureLayout() error {
	for _, d := range []string{w.KiroDir(), w.SpecsDir(), w.SteeringDir(), w.HooksDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return ferrors.Wrap("workspace.ensure", err, "create "+d)
		}
	}
	if _, err := os.Stat(w.SettingsPath()); os.IsNotExist(err) {
		if err := w.SaveSettings(defaultSettings()); err != nil {
			return err
		}
	} else if err != nil {
		return ferrors.Wrap("workspace.ensure", err, "stat settings.json")
	}
	return nil
}

// LoadSettings reads .kiro/settings.json. Returns an empty map when the file
// does not exist; raises WorkspaceError on JSON parse failure.
func (w *Workspace) LoadSettings() (map[string]any, error) {
	data, err := os.ReadFile(w.SettingsPath())
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, ferrors.Wrap("workspace.settings", err, "read settings.json")
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, ferrors.Wrap("workspace.settings", err, "parse settings.json")
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// SaveSettings writes .kiro/settings.json. Marshals with 2-space indent.
func (w *Workspace) SaveSettings(data map[string]any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ferrors.Wrap("workspace.settings", err, "marshal settings.json")
	}
	b = append(b, '\n')
	if err := os.WriteFile(w.SettingsPath(), b, 0o644); err != nil {
		return ferrors.Wrap("workspace.settings", err, "write settings.json")
	}
	return nil
}

func defaultSettings() map[string]any {
	return map[string]any{
		"generator": "template",
		"model": map[string]any{
			"provider":    "template",
			"base_url":    "",
			"model":       "",
			"api_key_env": "",
		},
	}
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}