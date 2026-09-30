// Package ide handles integration with AI coding assistants (Claude Code,
// CodeBuddy, etc.). It detects which IDEs are installed and writes the
// hook configuration that turns free-kiro into a real-time lint gate.
//
// Claude Code's actual settings.json schema (as of late 2026) is:
//
//	{
//	  "hooks": {
//	    "<EventName>": [
//	      {
//	        "matcher": "<regex>",
//	        "hooks": [
//	          {"type": "command", "command": "..."}
//	        ]
//	      }
//	    ]
//	  }
//	}
//
// We preserve the user's existing event-keyed hooks (and any other top-
// level keys like `model`, `enabledPlugins`) and upsert our entries
// alongside them. Each free-kiro entry's command starts with a
// `# free-kiro-managed:` marker so we can identify (and remove) our
// entries on subsequent installs without clobbering user-managed hooks.
package ide

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

//go:embed templates/*.md
var templatesFS embed.FS

// ID is the canonical name of a supported IDE.
type ID string

// Known IDE identifiers free-kiro can configure.
const (
	// ClaudeCode is the Anthropic Claude Code CLI.
	ClaudeCode ID = "claude-code"
	// CodeBuddy is the Tencent CodeBuddy IDE.
	CodeBuddy ID = "codebuddy"
	// Cursor is the Cursor AI editor.
	Cursor ID = "cursor"
	// Continue is the Continue.dev VS Code / JetBrains extension.
	Continue ID = "continue"
	// OpenCode is the OpenCode CLI / IDE.
	OpenCode ID = "opencode"
)

// All returns the list of IDEs free-kiro knows how to configure.
func All() []ID {
	return []ID{ClaudeCode, CodeBuddy, Cursor, Continue, OpenCode}
}

// String pretty-prints the ID.
func (i ID) String() string { return string(i) }

// Parse converts a CLI flag value to an ID. Empty / "auto" / "none" all
// map to the empty ID — the caller resolves the meaning.
func Parse(s string) (ID, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto", "none":
		return "", nil
	case "claude-code", "claudecode", "claude":
		return ClaudeCode, nil
	case "codebuddy":
		return CodeBuddy, nil
	case "cursor":
		return Cursor, nil
	case "continue":
		return Continue, nil
	case "opencode":
		return OpenCode, nil
	}
	return "", ferrors.NewUsage("ide.parse",
		"unknown IDE: "+s+" (supported: claude-code, codebuddy, cursor, continue, opencode, none)")
}

// Info describes one supported IDE and where its settings live.
type Info struct {
	ID         ID
	ConfigPath string // absolute path to settings.json
	DirExists  bool
}

// DetectAll enumerates the supported IDEs.
func DetectAll(home string) []Info {
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			home = "."
		}
	}
	out := make([]Info, 0, len(All()))
	for _, id := range All() {
		cfg := configPathFor(id, home)
		dir := filepath.Dir(cfg)
		_, statErr := os.Stat(dir)
		out = append(out, Info{ID: id, ConfigPath: cfg, DirExists: statErr == nil})
	}
	return out
}

func configPathFor(id ID, home string) string {
	switch id {
	case ClaudeCode:
		return filepath.Join(home, ".claude", "settings.json")
	case CodeBuddy:
		return filepath.Join(home, ".codebuddy", "settings.json")
	case Cursor:
		return filepath.Join(home, ".cursor", "settings.json")
	case Continue:
		return filepath.Join(home, ".continue", "config.json")
	case OpenCode:
		return filepath.Join(home, ".opencode", "settings.json")
	}
	return ""
}

// FreeKiroHookPrefix is the prefix every free-kiro-managed hook's `name`
// starts with. Used to identify our hooks when merging into existing
// settings — we never clobber hooks owned by other tools.
const FreeKiroHookPrefix = "free-kiro-"

// freeKiroMarker is the comment marker prepended to every free-kiro hook
// command in the IDE's settings. Lets us identify our entries on
// subsequent installs.
const freeKiroMarker = "# free-kiro-managed:"

// HookSpec is one event entry in the IDE's settings — what most
// documentation calls a "hook" (a matcher + a list of commands).
type HookSpec struct {
	Matcher string   `json:"matcher,omitempty"`
	Hooks   []HookIn `json:"hooks"`
}

// HookIn is one command within a hook entry.
type HookIn struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// settingsShape captures the bits of the IDE's settings.json that
// free-kiro touches. Everything else (model, enabledPlugins, …) is
// preserved as raw JSON so we round-trip the file untouched.
type settingsShape struct {
	Hooks map[string][]HookSpec `json:"hooks"`
	// RawMessage preserves the rest of the settings file verbatim. We
	// re-marshal by encoding only the well-known fields + this blob.
	Raw json.RawMessage `json:"-"`
}

// FreeKiroHookSpec is one hook entry that free-kiro knows how to install.
type FreeKiroHookSpec struct {
	Event   string
	Matcher string
	Command string
}

// freeKiroHooks is the canonical set of hooks installed by `init --ide`.
// Designed to be a single source of truth — adding a new hook here
// auto-propagates to every project.
func freeKiroHooks() []FreeKiroHookSpec {
	marker := freeKiroMarker + " "
	return []FreeKiroHookSpec{
		{
			Event:   "PreToolUse",
			Matcher: "Edit|Write",
			Command: marker + "lint-gate: free-kiro lint || exit 2",
		},
		{
			Event:   "SessionStart",
			Matcher: "",
			Command: marker + "session-next: free-kiro spec next $(cat .kiro/.current 2>/dev/null) || free-kiro spec list 2>/dev/null | awk 'NR==2{print $1}'",
		},
	}
}

// InstallHooks writes the canonical free-kiro hook set into the IDE's
// settings.json. Existing user hooks are preserved; free-kiro hooks are
// upserted (matched on command-line marker).
func InstallHooks(id ID, home string) (path string, note string, err error) {
	cfgPath := configPathFor(id, home)
	if cfgPath == "" {
		return "", "", ferrors.New("ide.install", "unsupported IDE: "+string(id))
	}
	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", ferrors.Wrap("ide.install", err, "mkdir "+dir)
	}

	settings, rawBlob, err := readOrLoadSettings(cfgPath)
	if err != nil {
		return "", "", err
	}
	if settings.Hooks == nil {
		settings.Hooks = map[string][]HookSpec{}
	}
	for _, fk := range freeKiroHooks() {
		settings.Hooks[fk.Event] = upsertHookEntry(settings.Hooks[fk.Event], fk)
	}

	out, err := marshalSettings(settings, rawBlob)
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return "", "", ferrors.Wrap("ide.install", err, "write "+cfgPath)
	}
	note = "free-kiro hooks written to " + cfgPath
	return cfgPath, note, nil
}

// upsertHookEntry removes any entry in `entries` whose first command
// starts with our marker, then appends the new entry. Other (user-
// managed) entries are left alone.
func upsertHookEntry(entries []HookSpec, fk FreeKiroHookSpec) []HookSpec {
	out := make([]HookSpec, 0, len(entries)+1)
	for _, e := range entries {
		if isFreeKiroEntry(e) {
			continue
		}
		out = append(out, e)
	}
	return append(out, HookSpec{
		Matcher: fk.Matcher,
		Hooks: []HookIn{
			{Type: "command", Command: fk.Command},
		},
	})
}

// isFreeKiroEntry reports whether the entry's first command was written
// by us. We only inspect the first hook's command (a hook entry always
// has at least one command by convention).
func isFreeKiroEntry(e HookSpec) bool {
	if len(e.Hooks) == 0 {
		return false
	}
	return strings.HasPrefix(e.Hooks[0].Command, freeKiroMarker)
}

// readOrLoadSettings reads the IDE's settings.json and returns both
// the structured (hooks) part and the raw JSON blob (other keys).
// Missing file → empty settings, nil blob. Tolerant of arbitrary
// extra top-level keys (model, enabledPlugins, …) — those are
// preserved verbatim in the raw blob.
func readOrLoadSettings(path string) (settingsShape, json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return settingsShape{}, nil, nil
		}
		return settingsShape{}, nil, ferrors.Wrap("ide.install", err, "read "+path)
	}
	// Parse the full file: we only care about the "hooks" key, but we
	// need the rest as a blob to round-trip unchanged.
	var full struct {
		Hooks map[string][]HookSpec `json:"hooks"`
	}
	if err := json.Unmarshal(data, &full); err != nil {
		return settingsShape{}, nil, ferrors.New("ide.install",
			"settings.json is malformed — refusing to overwrite")
	}
	return settingsShape{Hooks: full.Hooks}, data, nil
}

// marshalSettings writes a settings.json that preserves the original
// raw blob (minus the hooks key) and overlays our hooks. This keeps
// every other top-level key (model, enabledPlugins, etc.) intact.
func marshalSettings(s settingsShape, rawBlob json.RawMessage) ([]byte, error) {
	var base map[string]json.RawMessage
	if len(rawBlob) > 0 {
		if err := json.Unmarshal(rawBlob, &base); err != nil {
			return nil, ferrors.Wrap("ide.install", err, "parse settings")
		}
	} else {
		base = map[string]json.RawMessage{}
	}
	hooksBytes, err := json.Marshal(s.Hooks)
	if err != nil {
		return nil, ferrors.Wrap("ide.install", err, "marshal hooks")
	}
	base["hooks"] = hooksBytes
	out, err := json.MarshalIndent(base, "", "  ")
	if err != nil {
		return nil, ferrors.Wrap("ide.install", err, "marshal settings")
	}
	return append(out, '\n'), nil
}

// countFreeKiroHooks reads an IDE's settings.json and counts entries
// whose first command starts with the free-kiro marker. Returns
// (true, count, nil) when the file parses successfully.
func countFreeKiroHooks(path string) (bool, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, 0, nil
		}
		return false, 0, err
	}
	var s settingsShape
	if err := json.Unmarshal(data, &s); err != nil {
		return false, 0, err
	}
	count := 0
	for _, entries := range s.Hooks {
		for _, e := range entries {
			if isFreeKiroEntry(e) {
				count++
			}
		}
	}
	return true, count, nil
}

// WriteAgentsMD creates .kiro/AGENTS.md. Idempotent unless overwrite=true.
func WriteAgentsMD(workspaceRoot, lang string, overwrite bool) (string, error) {
	dest := filepath.Join(workspaceRoot, ".kiro", "AGENTS.md")
	if !overwrite {
		if _, err := os.Stat(dest); err == nil {
			return dest, nil
		}
	}
	body, err := agentsTemplate(lang)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, []byte(body), 0o644); err != nil {
		return "", ferrors.Wrap("ide.agents", err, "write "+dest)
	}
	return dest, nil
}

func agentsTemplate(lang string) (string, error) {
	name := "agents_en.md"
	if strings.HasPrefix(strings.ToLower(lang), "zh") {
		name = "agents_zh.md"
	}
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return "", ferrors.Wrap("ide.agents", err, "read template "+name)
	}
	return string(data), nil
}

// FormatHookReport renders a human-friendly one-liner per installed hook.
func FormatHookReport(id ID, cfgPath string) string {
	return fmt.Sprintf("%s: %s", id, cfgPath)
}
