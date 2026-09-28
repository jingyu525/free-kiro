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
// This package deliberately has no external deps: the frontmatter parser
// is a tiny recursive-descent over `key: value` lines, the glob matcher is
// a hand-rolled scanner. Same shape as kiro-clone, in idiomatic Go.
package steering

import (
	"regexp"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

// ValidModes are the four inclusion modes a steering doc can declare.
var ValidModes = map[string]bool{
	"always":    true,
	"auto":      true,
	"manual":    true,
	"filematch": true,
}

// fmOpen is the first line of a YAML-ish frontmatter block.
var fmOpen = regexp.MustCompile(`^---\s*$`)

// fmKey is one `key: value` line inside the frontmatter.
var fmKey = regexp.MustCompile(`^([A-Za-z_][\w-]*)\s*:\s*(.*)$`)

// ParseFrontmatter splits a doc into (meta, body). Accepts a leading
// `---\n key: value\n---\n` block. Values may be single- or double-
// quoted. Missing or malformed frontmatter returns ({}, text) — the
// caller decides what to do (e.g. fall back to default mode).
func ParseFrontmatter(text string) (map[string]string, string) {
	lines := splitLinesForSteering(text)
	if len(lines) == 0 || !fmOpen.MatchString(lines[0]) {
		return map[string]string{}, text
	}
	meta := map[string]string{}
	i := 1
	for i < len(lines) {
		if fmOpen.MatchString(lines[i]) {
			body := joinLines(lines[i+1:])
			return meta, body
		}
		m := fmKey.FindStringSubmatch(lines[i])
		if m == nil {
			i++
			continue
		}
		val := m[2]
		if len(val) >= 2 && val[0] == val[len(val)-1] && (val[0] == '\'' || val[0] == '"') {
			val = val[1 : len(val)-1]
		}
		meta[m[1]] = val
		i++
	}
	// No closing fence — treat the whole thing as body.
	return map[string]string{}, text
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

// splitLinesForSteering returns lines without trailing newlines.
func splitLinesForSteering(text string) []string {
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
	if start < len(text) {
		out = append(out, text[start:])
	}
	return out
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for i := 1; i < len(lines); i++ {
		out += "\n" + lines[i]
	}
	return out
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

// ensure imports referenced.
var _ = ferrors.New