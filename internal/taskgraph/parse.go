// Package taskgraph parses tasks.md and computes parallel execution waves.
//
// Tasks are written as:
//
//   - [ ] #1 Set up module layout
//   - [x] #2 Implement core domain model [deps: #1]
//   - [X] #3 Polish            [deps: #1,#2]
//
// The engine groups tasks into *waves* (tasks within a wave share no
// dependencies and may run concurrently). Cycle detection uses DFS with
// WHITE / GRAY / BLACK colouring; a back-edge into a GRAY (currently on
// the DFS stack) node is a cycle.
package taskgraph

import (
	"regexp"
	"strconv"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/text"
)

// lineRe matches `- [ ] #1 Title [deps: #2,#3]` and friends.
var lineRe = regexp.MustCompile(`^\s*-\s*\[( |x|X)\]\s*#(\d+)\s+(.*?)\s*(?:\[deps:\s*([^\]]*)\])?\s*$`)

// ParseTasks extracts tasks from a tasks.md document. Lines that don't
// match the canonical format are silently skipped (they're probably
// free-form prose — headers, notes, etc.).
func ParseTasks(doc string) []models.Task {
	var out []models.Task
	for _, line := range text.RangeLines(doc) {
		m := lineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		mark, idStr, title, depsStr := m[1], m[2], m[3], m[4]
		deps := parseDepIDs(depsStr)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue // lineRe guarantees digits, but be defensive
		}
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
		id, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out
}

var depIDRe = regexp.MustCompile(`#(\d+)`)
