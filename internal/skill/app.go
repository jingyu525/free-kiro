// Package skill implements `free-kiro skill {install,uninstall,update,
// show,path,version}` — managing the installable SKILL.md bundle that
// exposes free-kiro to Claude Code / OpenCode / Codex CLI / CodeBuddy.
//
// The bundle itself lives in `contrib/skills/free-kiro/` in the repo and
// is published as `free-kiro-skill_<version>.zip` via GoReleaser. This
// package downloads that zip, verifies its sha256 against the SHA256SUMS
// shipped in the same release, and writes the bundle into each supported
// app's `~/.{app}/skills/<name>/` directory.
//
// Zero external dependencies — stdlib only (net/http, archive/zip,
// encoding/json, crypto/sha256). Mirrors the zero-dep posture of the
// rest of free-kiro.
package skill

import (
	"strings"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// App identifies a target AI coding assistant that consumes skills.
// Each app has its own `~/.{app}/skills/<name>/` directory layout.
type App string

// Skill-app identifiers that the skill bundle can target.
const (
	// AppClaudeCode is the Anthropic Claude Code CLI.
	AppClaudeCode App = "claude-code"
	// AppOpenCode is the opencode CLI.
	AppOpenCode App = "opencode"
	// AppCodex is the OpenAI Codex CLI.
	AppCodex App = "codex"
	// AppCodeBuddy is the Tencent CodeBuddy IDE.
	AppCodeBuddy App = "codebuddy"
)

// AllApps returns every supported App. Used for `--app all` and for
// DetectInstalledApps enumeration.
func AllApps() []App {
	return []App{AppClaudeCode, AppOpenCode, AppCodex, AppCodeBuddy}
}

// ParseApp normalizes a CLI flag value to an App. Empty / "all" return
// ("", nil) — caller resolves the meaning.
func ParseApp(s string) (App, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "all":
		return "", nil
	case "claude-code", "claudecode", "claude":
		return AppClaudeCode, nil
	case "opencode":
		return AppOpenCode, nil
	case "codex":
		return AppCodex, nil
	case "codebuddy":
		return AppCodeBuddy, nil
	}
	return "", ferrors.New("skill.parse-app",
		"unknown app: "+s+" (supported: claude-code, opencode, codex, codebuddy, all)")
}

// String pretty-prints the App.
func (a App) String() string { return string(a) }

// Label returns a human-friendly display name.
func (a App) Label() string {
	switch a {
	case AppClaudeCode:
		return "Claude Code"
	case AppOpenCode:
		return "OpenCode"
	case AppCodex:
		return "Codex CLI"
	case AppCodeBuddy:
		return "CodeBuddy"
	}
	return string(a)
}
