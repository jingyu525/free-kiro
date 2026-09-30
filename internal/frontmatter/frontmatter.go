// Package frontmatter parses, marshals, and validates the simple YAML-ish
// `---\nkey: value\n---\n<body>` metadata block shared by steering docs,
// spec docs, and the init templates.
//
// Deliberately stdlib-only: no external YAML library. The wire format is a
// strict subset of YAML (no nested maps, no lists, no flow syntax) — enough
// for the {mode, description, fileMatchPattern, type, …} keys the project
// actually uses, and simple enough to keep predictable under CRLF / BOM /
// empty-body edge cases.
//
// The accepted shape:
//
//	---
//	mode: always
//	description: this doc is always loaded
//	---
//	# Heading
//
//	body text…
//
// Values may be single- or double-quoted (quotes are stripped). Keys are
// `[A-Za-z_][\w-]*`. Unknown lines inside the fence are skipped (forward
// compatible). Missing or unclosed fence → ErrNoFrontmatter (Parse) or
// empty frontmatter (lower-level helpers), so callers can decide policy.
package frontmatter

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Frontmatter is the parsed key/value metadata block. Values are kept as
// strings at parse time; Validate coerces per its FieldRule.Type.
type Frontmatter = map[string]any

// ErrNoFrontmatter is returned by Parse when the input has no opening
// `---` fence.
var ErrNoFrontmatter = errors.New("frontmatter: missing opening `---` fence")

// fenceOpen matches the literal `---` line that opens / closes a
// frontmatter block. Whitespace is allowed around the dashes.
const fenceOpen = "---"

// Parse reads r and returns the frontmatter map and the body bytes that
// follow the closing fence. Errors:
//
//   - ErrNoFrontmatter — r has no opening `---` line.
//   - fmt.Errorf("frontmatter: …") — malformed (unclosed fence, etc.).
//
// A leading UTF-8 BOM (\xEF\xBB\xBF) is stripped before scanning. Body
// bytes preserve the original line terminators so round-tripping a doc
// (Parse → Marshal) is byte-stable except for key ordering. CRLF inputs
// surface as lines that still carry a trailing \r (the validator and
// callers handle it via strings.TrimSpace on the value side).
func Parse(r io.Reader) (Frontmatter, []byte, error) {
	if r == nil {
		return nil, nil, ErrNoFrontmatter
	}
	// Steering / spec docs are KB-scale; cap at 4 MiB to refuse runaway
	// inputs early.
	const maxReadBytes = 4 << 20
	data, err := io.ReadAll(io.LimitReader(r, maxReadBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("frontmatter: read: %w", err)
	}
	text := string(data)
	// Strip leading UTF-8 BOM (3-byte sequence).
	text = strings.TrimPrefix(text, "\xEF\xBB\xBF")
	return parseLines(splitLines(text))
}

// splitLines breaks text on '\n', preserving the raw line content (CRLF
// leaves a trailing \r on the line, mirrors the previous inline behaviour
// in internal/steering/frontmatter.go so callers don't see a silent
// behavioural change). Matches strings.Split semantics, including a
// trailing empty element when text ends with '\n'.
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
	// Always append the trailing segment (may be empty when text ends with
	// '\n'). Mirrors strings.Split, which is the contract callers expect.
	out = append(out, text[start:])
	return out
}

// parseLines does the actual scan. Exposed for tests so they can drive
// edge cases (CRLF, BOM, empty body) without going through an io.Reader.
func parseLines(lines []string) (Frontmatter, []byte, error) {
	if len(lines) == 0 {
		return nil, nil, ErrNoFrontmatter
	}
	if strings.TrimSpace(lines[0]) != fenceOpen {
		return nil, nil, ErrNoFrontmatter
	}
	fm := Frontmatter{}
	i := 1
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == fenceOpen {
			body := strings.Join(lines[i+1:], "\n")
			return fm, []byte(body), nil
		}
		key, val, ok := parseKV(lines[i])
		if ok {
			fm[key] = val
		}
		i++
	}
	// Closing fence missing → surface as ErrNoFrontmatter so callers can
	// decide (ParseFrontmatter-style lenient callers wrap; strict callers
	// error).
	return nil, nil, fmt.Errorf("%w (no closing fence)", ErrNoFrontmatter)
}

// parseKV returns (key, value, true) when line is a `key: value` line with
// the key matching `[A-Za-z_][\w-]*`. Otherwise ok=false. Surrounding
// single / double quotes are stripped from value.
func parseKV(line string) (string, string, bool) {
	idx := strings.IndexByte(line, ':')
	if idx <= 0 {
		return "", "", false
	}
	key := line[:idx]
	if !validKey(key) {
		return "", "", false
	}
	rest := strings.TrimSpace(line[idx+1:])
	rest = unquote(rest)
	return key, rest, true
}

// validKey reports whether s is a non-empty identifier-shaped token.
func validKey(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_':
		case (r >= '0' && r <= '9') && i > 0:
		case r == '-':
		default:
			return false
		}
	}
	return true
}

// unquote strips a single matching pair of ' or " from s.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

// Marshal produces the canonical bytes:
//
//	---\nkey1: val1\nkey2: val2\n---\n<body>
//
// Key order follows map iteration (intentionally non-deterministic —
// callers that need stable output should sort beforehand or pass an
// ordered Map). Values are stringified via fmt.Sprint.
func Marshal(fm Frontmatter, body []byte) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(fenceOpen)
	sb.WriteByte('\n')
	for k, v := range fm {
		if !validKey(k) {
			return nil, fmt.Errorf("frontmatter: invalid key %q", k)
		}
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(stringify(v))
		sb.WriteByte('\n')
	}
	sb.WriteString(fenceOpen)
	sb.WriteByte('\n')
	if len(body) > 0 {
		sb.Write(body)
	}
	return []byte(sb.String()), nil
}

// stringify renders any value as a single-line frontmatter value. Strings
// pass through; other types use fmt.Sprint (single-line, no trailing
// newline).
func stringify(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}
