package skill

import (
	"os"
	"path/filepath"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// SkillsDir returns the absolute skills directory for `app` under `home`.
// Path convention: `~/.{app}/skills/{subdir}`. Subdir is the bundle name
// (e.g. "free-kiro") — defaults to the App's name when empty.
//
// Example: SkillsDir(AppClaudeCode, "/Users/x", "free-kiro")
//
//	→ /Users/x/.claude/skills/free-kiro
func SkillsDir(app App, home, subdir string) string {
	if subdir == "" {
		subdir = string(app)
	}
	switch app {
	case AppClaudeCode:
		return filepath.Join(home, ".claude", "skills", subdir)
	case AppOpenCode:
		return filepath.Join(home, ".opencode", "skills", subdir)
	case AppCodex:
		return filepath.Join(home, ".codex", "skills", subdir)
	case AppCodeBuddy:
		return filepath.Join(home, ".codebuddy", "skills", subdir)
	}
	return ""
}

// HomeDir returns the user's home directory (with $HOME honored first).
func HomeDir() (string, error) {
	if h := os.Getenv("FREE_KIRO_HOME"); h != "" {
		return h, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", ferrors.Wrap("skill.paths", err, "locate home directory")
	}
	return h, nil
}

// DetectInstalledApps reports which apps appear to be installed
// (have their config root directory on the filesystem). Returns the
// apps that pass detection, in canonical order.
//
// Detection is intentionally cheap (single os.Stat per app) — we only
// check for the existence of the app's primary config root, not whether
// the app supports skills specifically. Apps that don't ship skills
// will still appear here; downstream code may warn but won't error.
func DetectInstalledApps(home string) []App {
	if home == "" {
		h, err := HomeDir()
		if err != nil {
			return nil
		}
		home = h
	}
	out := make([]App, 0, 4)
	for _, app := range AllApps() {
		root := appRoot(app)
		if root == "" {
			continue
		}
		root = strings.TrimPrefix(root, "~/")
		root = filepath.Join(home, root)
		if _, err := os.Stat(root); err == nil {
			out = append(out, app)
		}
	}
	return out
}

// appRoot returns the per-app config root path used for "is this app
// installed" detection. Mirrors `detected_by` in skill.json.
func appRoot(app App) string {
	switch app {
	case AppClaudeCode:
		return "~/.claude"
	case AppOpenCode:
		return "~/.opencode"
	case AppCodex:
		return "~/.codex"
	case AppCodeBuddy:
		return "~/.codebuddy"
	}
	return ""
}

// EnsureSkillsDir creates the skills directory for `app` if it doesn't
// exist. Returns the absolute path.
func EnsureSkillsDir(app App, home, subdir string) (string, error) {
	dir := SkillsDir(app, home, subdir)
	if dir == "" {
		return "", ferrors.New("skill.paths", "unsupported app: "+string(app))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", ferrors.Wrap("skill.paths", err, "mkdir "+dir)
	}
	return dir, nil
}