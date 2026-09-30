// Package text provides shared text helpers used across packages
// that parse multi-line documents (lint rules, taskgraph, etc.).
//
// Kept small on purpose — every helper here must earn its place by
// being used from at least two other packages. If a helper is only
// used in one package, it should live next to its caller.
package text

// RangeLines returns each line of text with the trailing newline
// stripped. Empty input returns nil. A trailing line without a
// terminating newline is preserved as-is.
//
// Allocation-light: no strings.Split, no scanner. Identical to the
// previous hand-rolled copies in internal/lint/ears.go and
// internal/taskgraph/parse.go before housekeeping-cleanup landed.
func RangeLines(text string) []string {
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