package lint

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// docsAnchorHintRe matches `docs/EARS.md#<anchor>` substrings inside any
// string literal (typically the Hint: field of a lint Issue). The anchor
// is captured in group 1. Used by TestLintHintsResolve to extract every
// docs/EARS.md anchor referenced from internal/lint/*.go.
var docsAnchorHintRe = regexp.MustCompile(`docs/EARS\.md#([\w-]+)`)

// hintLineRe matches a Go source line that defines a Hint: field with a
// string literal — captures the literal contents in group 1.
var hintLineRe = regexp.MustCompile(`(?i)Hint:\s*"([^"]+)"`)

// headingRe matches a markdown ATX heading line (1-6 `#`s followed by a
// title and an optional Pandoc `{#english-slug}` explicit anchor. Group
// 1 = title, group 2 = explicit anchor (empty if absent).
var headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+(.*?)\s*(?:\{#([\w-]+)\})?\s*$`)

// brokenHintFormat is the single Errorf template used by BOTH
// TestLintHintsResolve (production) and TestLintHintsResolve_NegativePath
// (assertion on the rendered message). Keep them in sync by referencing
// this constant from both call sites — no hand-mirrored strings.
const brokenHintFormat = "%s:%d Hint %q anchors %q which is not in docs/EARS.md (valid anchors: %v)"

// TestLintHintsResolve — every Hint: string in the linter source files
// that contains `docs/EARS.md#<anchor>` must point at an anchor that
// actually exists in docs/EARS.md. This catches the broken GitHub-deep-
// link case where someone renames or restructures a heading without
// updating the matching Hint string.
//
// Fails with file:line + Hint literal + broken anchor + the full set of
// valid anchors so the breakage is obvious from test output alone.
//
// Covers:
//   - internal/lint/requirements.go
//   - internal/lint/bugfix.go
//   - internal/lint/quality.go
func TestLintHintsResolve(t *testing.T) {
	docsPath := findDocsEARS(t)
	docsBytes, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("read %s: %v", docsPath, err)
	}
	valid := extractHeadingSlugs(string(docsBytes))

	files := []string{
		"requirements.go",
		"bugfix.go",
		"quality.go",
	}
	for _, name := range files {
		path := name // running `go test` from package dir
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(string(src), "\n")
		for i, line := range lines {
			for _, m := range hintLineRe.FindAllStringSubmatch(line, -1) {
				hint := m[1]
				for _, am := range docsAnchorHintRe.FindAllStringSubmatch(hint, -1) {
					anchor := am[1]
					if !valid[anchor] {
						t.Errorf(brokenHintFormat,
							path, i+1, hint, anchor, sortedKeys(valid))
					}
				}
			}
		}
	}
}

// TestLintHintsResolve_NegativePath verifies the error message format of
// TestLintHintsResolve. We point a synthetic Hint string at an obviously
// invalid anchor ("this-anchor-does-not-exist") and check that the error
// includes (a) the source filename, (b) the line number, (c) the broken
// anchor name, and (d) the hint literal. This guards against silent
// regressions where someone "fixes" the negative-path message to just
// "broken hint" without telling the reader which file / anchor.
//
// The check re-renders the production error format (brokenHintFormat)
// and asserts on the resulting string. If extractHeadingSlugs or
// docsAnchorHintRe change, this test pins the user-visible contract.
func TestLintHintsResolve_NegativePath(t *testing.T) {
	docsPath := findDocsEARS(t)
	docsBytes, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("read %s: %v", docsPath, err)
	}
	valid := extractHeadingSlugs(string(docsBytes))

	const brokenAnchor = "this-anchor-does-not-exist"
	if valid[brokenAnchor] {
		t.Fatalf("test setup: %q unexpectedly exists in docs/EARS.md", brokenAnchor)
	}

	const fakeFile = "hints_negative_test.go"
	const fakeLine = 42
	fakeHint := `see docs/EARS.md#` + brokenAnchor + ` for the foo policy`

	var found bool
	for _, am := range docsAnchorHintRe.FindAllStringSubmatch(fakeHint, -1) {
		if !valid[am[1]] {
			msg := fmt.Sprintf(brokenHintFormat,
				fakeFile, fakeLine, fakeHint, am[1], sortedKeys(valid))
			for _, want := range []string{fakeFile, "42", brokenAnchor, fakeHint} {
				if !strings.Contains(msg, want) {
					t.Errorf("negative-path error missing %q\nfull message: %s", want, msg)
				}
			}
			found = true
		}
	}
	if !found {
		t.Fatal("negative path: expected docsAnchorHintRe to extract the broken anchor")
	}
}

// findDocsEARS locates docs/EARS.md relative to the package directory.
// go test runs each package from its directory, so ../docs/EARS.md is
// the expected layout. We try both `../../docs/EARS.md` (when running
// from internal/lint/) and a relative `docs/EARS.md` fallback for
// unusual invocations.
func findDocsEARS(t *testing.T) string {
	t.Helper()
	candidates := []string{"../../docs/EARS.md", "../docs/EARS.md", "docs/EARS.md"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Fatalf("could not locate docs/EARS.md (tried %v)", candidates)
	return ""
}

// extractHeadingSlugs returns the set of anchorable IDs present in a
// markdown document. Both Pandoc `{#english-slug}` explicit anchors AND
// GitHub-implicit slugs (lowercase, spaces → hyphens, CJK preserved) are
// captured so callers can validate either form of deep link.
//
// GitHub's anchor algorithm (2024+) preserves CJK characters as-is and
// only lowercases / punctuation-strips the Latin parts. Our
// githubImplicitSlug approximates that behaviour with stdlib only.
func extractHeadingSlugs(md string) map[string]bool {
	out := map[string]bool{}
	for _, m := range headingRe.FindAllStringSubmatch(md, -1) {
		title := strings.TrimSpace(m[1])
		if explicit := strings.TrimSpace(m[2]); explicit != "" {
			out[explicit] = true
		}
		out[githubImplicitSlug(title)] = true
	}
	return out
}

// githubImplicitSlug approximates GitHub's auto-generated heading slug:
// 1. lowercase
// 2. spaces and hyphens → '-', collapse runs
// 3. strip other ASCII punctuation (keep ASCII letters / digits)
// 5. preserve CJK (Han) characters and other Unicode letters / digits
func githubImplicitSlug(title string) string {
	title = strings.ToLower(title)
	var b strings.Builder
	prevHyphen := false
	flushHyphen := func() {
		if !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	for _, r := range title {
		switch {
		case r == ' ' || r == '\t' || r == '-':
			flushHyphen()
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		default:
			// CJK + other Unicode letters / digits are preserved by
			// GitHub since 2017; mirror that behaviour.
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
				prevHyphen = false
			}
			// else: drop punctuation / emoji / etc.
		}
	}
	return strings.Trim(b.String(), "-")
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}