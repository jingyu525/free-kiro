// Package taskgraph parses tasks.md and computes parallel execution waves.
//
// Tasks are written as:
//
//	- [ ] #1 Set up module layout
//	- [x] #2 Implement core domain model [deps: #1]
//	- [X] #3 Polish            [deps: #1,#2]
//
// The engine groups tasks into *waves* (tasks within a wave share no
// dependencies and may run concurrently). Cycle detection uses DFS with
// WHITE / GRAY / BLACK colouring; a back-edge into a GRAY (currently on
// the DFS stack) node is a cycle.
package taskgraph

import (
	"regexp"

	"github.com/jingyu525/free-kiro/internal/models"
)

// lineRe matches `- [ ] #1 Title [deps: #2,#3]` and friends.
var lineRe = regexp.MustCompile(`^\s*-\s*\[( |x|X)\]\s*#(\d+)\s+(.*?)\s*(?:\[deps:\s*([^\]]*)\])?\s*$`)

// ParseTasks extracts tasks from a tasks.md document. Lines that don't
// match the canonical format are silently skipped (they're probably
// free-form prose — headers, notes, etc.).
func ParseTasks(text string) []models.Task {
	var out []models.Task
	for _, line := range splitLines(text) {
		m := lineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		mark, idStr, title, depsStr := m[1], m[2], m[3], m[4]
		deps := parseDepIDs(depsStr)
		id := atoi(idStr)
		out = append(out, models.Task{
			ID:    id,
			Title: title,
			Deps:  deps,
			Done:  mark == "x" || mark == "X",
			Raw:   line,
		})
	}
	return out
}

// parseDepIDs extracts the `#N` tokens from a deps string like "#1,#2".
func parseDepIDs(s string) []int {
	if s == "" {
		return nil
	}
	var out []int
	for _, m := range depIDRe.FindAllStringSubmatch(s, -1) {
		out = append(out, atoi(m[1]))
	}
	return out
}

var depIDRe = regexp.MustCompile(`#(\d+)`)

// atoi parses a small non-negative int. Stops at the first non-digit and
// returns what was accumulated; returns 0 only when the string has no
// leading digits.
func atoi(s string) int {
	n := 0
	saw := false
	for _, r := range s {
		if r < '0' || r > '9' {
			if saw {
				break
			}
			continue
		}
		n = n*10 + int(r-'0')
		saw = true
	}
	return n
}

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
	if start < len(text) {
		out = append(out, text[start:])
	}
	return out
}