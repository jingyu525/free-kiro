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
	"bufio"
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

// freeKiroInstructionMarker is the marker prepended on the first line of every
// free-kiro-written IDE instruction file (CLAUDE.md, .cursorrules, AGENTS.md,
// etc.). Lets us identify our files on subsequent installs without clobbering
// user-managed files at the same path.
const freeKiroInstructionMarker = "# free-kiro-managed:"

// instructionFiles is the canonical mapping from supported IDE id to the
// project-root instruction file paths that the IDE's agent loader reads on
// session start. Adding a new IDE means appending one entry here.
//
//	Claude Code — single CLAUDE.md at repo root.
//	Cursor       — legacy .cursorrules plus modular .cursor/rules/free-kiro.md.
//	Continue     — legacy .continuerules plus modular .continue/rules/free-kiro.md.
//	OpenCode     — AGENTS.md at repo root (cross-tool convention).
//	CodeBuddy    — AGENTS.md at repo root (Tencent domestic convention).
//
// Paths intentionally omit leading "./"; `instructionFiles[id]` values are
// joined to the workspace root by `WriteSingleInstruction`.
var instructionFiles = map[ID][]string{
	ClaudeCode: {"CLAUDE.md"},
	Cursor:     {".cursorrules", ".cursor/rules/free-kiro.md"},
	Continue:   {".continuerules", ".continue/rules/free-kiro.md"},
	OpenCode:   {"AGENTS.md"},
	CodeBuddy:  {"AGENTS.md"},
}

// InstructionFiles returns a copy of the canonical project-root
// instruction file paths for the given IDE id. Returns nil for unknown
// ids; callers should treat that as "this IDE has no per-project
// instruction files to check". The copy lets callers iterate freely
// without risking mutation of the package-level table.
func InstructionFiles(id ID) []string {
	rels := instructionFiles[id]
	if len(rels) == 0 {
		return nil
	}
	out := make([]string, len(rels))
	copy(out, rels)
	return out
}

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

// WriteAgentsMD creates .kiro/AGENTS.md (workspace-level steering doc).
//
// Note: prefer WriteAgentInstructions in new code. WriteAgentInstructions
// writes every canonical project-root file for each selected IDE id
// (including the AGENTS.md that the OpenCode and CodeBuddy loaders
// actually read). WriteAgentsMD only covers the .kiro/AGENTS.md path used
// by the steering store — which is a workspace-level doc, NOT a
// per-IDE instruction file. The steering-store loader still depends on
// this path, so the function stays in the API until the steering store
// migrates off it.
//
// Idempotent unless overwrite=true. The first line of the body is
// guaranteed to start with the free-kiro-managed marker so future
// readers can identify free-kiro-authored content.
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
	body = prependMarker(body)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", ferrors.Wrap("ide.agents", err, "mkdir "+filepath.Dir(dest))
	}
	if err := os.WriteFile(dest, []byte(body), 0o644); err != nil {
		return "", ferrors.Wrap("ide.agents", err, "write "+dest)
	}
	return dest, nil
}

func agentsTemplate(lang string) (string, error) {
	return instructionTemplate(lang, OpenCode)
}

// instructionTemplate picks the correct embedded template for the given IDE id
// and language. OpenCode and CodeBuddy share the AGENTS.md cross-tool
// convention and so use the original `agents_*.md` template. Claude Code,
// Cursor, and Continue have their own per-IDE instruction file conventions
// and use the `instructions_*.md` template addressed at those loaders.
//
// Adding a new language: drop `instructions_<lang>.md` or
// `agents_<lang>.md` next to the existing templates and add a branch on
// `strings.HasPrefix(strings.ToLower(lang), "<prefix>")`.
func instructionTemplate(lang string, id ID) (string, error) {
	prefix := "instructions"
	switch id {
	case OpenCode, CodeBuddy:
		prefix = "agents"
	}
	name := prefix + "_en.md"
	if strings.HasPrefix(strings.ToLower(lang), "zh") {
		name = prefix + "_zh.md"
	}
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return "", ferrors.Wrap("ide.template", err, "read template "+name)
	}
	return string(data), nil
}

// FormatHookReport renders a human-friendly one-liner per installed hook.
func FormatHookReport(id ID, cfgPath string) string {
	return fmt.Sprintf("%s: %s", id, cfgPath)
}

// IsFreeKiroInstruction reports whether the file at path begins with the
// free-kiro-managed instruction marker. A missing file returns
// (false, nil); other read errors return (false, err). Used by callers
// that need to distinguish files free-kiro owns from user-managed files at
// the same path.
func IsFreeKiroInstruction(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, ferrors.Wrap("ide.isInstruction", err, "open "+path)
	}
	// Close() error is intentionally dropped: this is a read-only probe
	// whose only failure mode is the scanner.Err() path below; the file
	// handle will be released by the runtime when `f` goes out of scope.
	// We still defer the close (rather than calling it eagerly) so the
	// scanner below gets to read the file end-to-end.
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		if scanErr := scanner.Err(); scanErr != nil {
			return false, ferrors.Wrap("ide.isInstruction", scanErr, "scan "+path)
		}
		return false, nil
	}
	return strings.HasPrefix(scanner.Text(), freeKiroInstructionMarker), nil
}

// WriteSingleInstruction writes every project-root instruction file declared
// in `instructionFiles[id]` for the given language, creating any missing
// parent directories. Existing files are skipped unless overwrite is true.
// Returns the relative paths that were actually written.
//
// Caller responsibilities: `root` should be an absolute workspace root (or
// the cwd); `lang` matches the languages listed in instructionTemplate's
// godoc.
func WriteSingleInstruction(root, lang string, overwrite bool, id ID) ([]string, error) {
	rels, ok := instructionFiles[id]
	if !ok {
		return nil, ferrors.New("ide.writeInstruction", "unsupported IDE: "+string(id))
	}
	body, err := instructionTemplate(lang, id)
	if err != nil {
		return nil, err
	}
	body = prependMarker(body)
	seen := map[string]bool{}
	var written []string
	for _, rel := range rels {
		if seen[rel] {
			continue
		}
		seen[rel] = true
		abs := filepath.Join(root, rel)
		if !overwrite {
			if _, statErr := os.Stat(abs); statErr == nil {
				continue
			}
		}
		dir := filepath.Dir(abs)
		if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
			return written, ferrors.Wrap("ide.writeInstruction", mkdirErr, "mkdir "+dir)
		}
		if writeErr := os.WriteFile(abs, []byte(body), 0o644); writeErr != nil {
			return written, ferrors.Wrap("ide.writeInstruction", writeErr, "write "+abs)
		}
		written = append(written, rel)
	}
	return written, nil
}

// WriteAgentInstructions writes the canonical instruction files for every
// selected IDE id, deduping by absolute path so that the same file targeted
// by multiple ids (e.g. OpenCode and CodeBuddy both targeting AGENTS.md) is
// written exactly once. Returns the relative paths that were actually
// written, in the order they were first attempted.
//
// Failures from any single id are returned immediately along with whatever
// paths had been collected so far; callers can decide whether to abort or
// continue.
func WriteAgentInstructions(root, lang string, overwrite bool, ids []ID) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, id := range ids {
		rels, err := WriteSingleInstruction(root, lang, overwrite, id)
		if err != nil {
			return out, err
		}
		for _, rel := range rels {
			abs, absErr := filepath.Abs(filepath.Join(root, rel))
			if absErr != nil {
				abs = filepath.Join(root, rel)
			}
			if seen[abs] {
				continue
			}
			seen[abs] = true
			out = append(out, rel)
		}
	}
	return out, nil
}

// prependMarker ensures the body has the free-kiro-managed marker on its
// first line. If the template already starts with the marker (new
// convention), the body is returned unchanged; otherwise the marker is
// prepended. Used so the legacy `agents_*.md` templates — which predate
// the marker convention — and the new `instructions_*.md` templates —
// which already start with the marker — produce identical on-disk files.
func prependMarker(body string) string {
	first, rest, hasRest := strings.Cut(body, "\n")
	if strings.HasPrefix(first, freeKiroInstructionMarker) {
		return body
	}
	if !hasRest {
		return freeKiroInstructionMarker + "\n"
	}
	return freeKiroInstructionMarker + "\n" + first + "\n" + rest
}
