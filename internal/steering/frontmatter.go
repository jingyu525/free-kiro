// Package steering loads and assembles steering documents for the agent.
//
// Steering files are persistent project-context documents (product /
// structure / tech) injected into every generation. They live in two
// scopes:
//
//   - workspace: <root>/.kiro/steering/*.md (plus <root>/AGENTS.md)
//   - global:    ~/.kiro/steering/*.md    (plus ~/.kiro/AGENTS.md)
//
// Workspace docs override global docs by name. Each doc carries a mode
// (always / auto / manual / filematch) that controls when it loads.
//
// This package reuses the shared frontmatter parser in
// internal/frontmatter (Wave 3 of the refactor spec). ParseFrontmatter
// is kept as a thin wrapper for backward compatibility with the existing
// steering tests and store.go call-site.
package steering

import (
	"fmt"
	"strings"

	"github.com/jingyu525/free-kiro/internal/frontmatter"
)

// ValidModes are the four inclusion modes a steering doc can declare.
var ValidModes = map[string]bool{
	"always":    true,
	"auto":      true,
	"manual":    true,
	"filematch": true,
}

// ParseFrontmatter splits a doc into (meta, body). Thin wrapper over
// frontmatter.Parse that returns the legacy (map[string]string, string)
// shape. Missing or malformed frontmatter returns ({}, text) so the
// caller can decide policy (e.g. fall back to default mode) — matching
// the pre-refactor lenient behaviour.
func ParseFrontmatter(text string) (map[string]string, string) {
	fm, body, err := frontmatter.Parse(strings.NewReader(text))
	if err != nil {
		// Lenient fallback: missing fence or unclosed fence → treat the
		// whole input as body (caller may default the mode elsewhere).
		return map[string]string{}, text
	}
	out := make(map[string]string, len(fm))
	for k, v := range fm {
		out[k] = fmt.Sprint(v)
	}
	return out, string(body)
}

// ParseFilePatterns parses a `fileMatchPattern` frontmatter value into a
// list of globs. Accepts either a single quoted glob (`"**/*.tsx"`) or a
// YAML-ish list (`["**/*.ts", "**/*.tsx"]`); the latter is returned as a
// raw string here, so the caller splits on commas after stripping
// brackets. Quoted single-glob values also have their quotes stripped.
func ParseFilePatterns(raw string) []string {
	v := trim(raw)
	if v == "" {
		return nil
	}
	if len(v) >= 2 && (v[0] == '\'' || v[0] == '"') && v[0] == v[len(v)-1] {
		v = v[1 : len(v)-1]
	}
	if v[0] == '[' && v[len(v)-1] == ']' {
		inner := v[1 : len(v)-1]
		var out []string
		for _, p := range splitComma(inner) {
			t := trim(p)
			if t == "" {
				continue
			}
			if len(t) >= 2 && t[0] == t[len(t)-1] && (t[0] == '\'' || t[0] == '"') {
				t = t[1 : len(t)-1]
			}
			out = append(out, t)
		}
		return out
	}
	return []string{v}
}

func trim(s string) string {
	for len(s) > 0 && isSpace(s[0]) {
		s = s[1:]
	}
	for len(s) > 0 && isSpace(s[len(s)-1]) {
		s = s[:len(s)-1]
	}
	return s
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

// splitLines breaks text on '\n', preserving each line verbatim
// (including a trailing empty element when text ends with '\n').
// Retained for steering/store.go's deriveName helper that scans the
// body for the first H1 title.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			out = append(out, text[start:i])
			start = i + 1
		}
	}
	out = append(out, text[start:])
	return out
}
