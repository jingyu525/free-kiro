package lint

import (
	"fmt"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
)

// LintTasks checks tasks.md for parseability, dangling/self deps, and
// dependency cycles. The format itself is fixed by taskgraph.ParseTasks;
// here we apply the quality rules on top of the parsed structure.
//
// Rules:
//
//   - No parseable tasks → WARNING `empty-tasks` (advisory; the author may
//     still be drafting).
//   - Task depending on itself → ERROR `self-dep` (blocks advance/approve).
//   - Task depending on a non-existent id → ERROR `dangling-dep`.
//   - Dependency cycle → ERROR `cycle` (blocks advance/approve).
func LintTasks(text string) []LintIssue {
	var out []LintIssue
	tasks := taskgraph.ParseTasks(text)
	if len(tasks) == 0 {
		out = append(out, LintIssue{
			Severity: SeverityWarning,
			Code:     "empty-tasks",
			Message:  "tasks.md has no parseable tasks",
			Location: "tasks.md",
		})
		return out
	}

	ids := map[int]bool{}
	for _, t := range tasks {
		ids[t.ID] = true
	}
	for _, t := range tasks {
		if containsInt(t.Deps, t.ID) {
			out = append(out, LintIssue{
				Severity: SeverityError,
				Code:     "self-dep",
				Message:  fmt.Sprintf("task #%d depends on itself", t.ID),
				Location: fmt.Sprintf("tasks.md:#%d", t.ID),
				Hint:     "remove the self-reference; a task cannot depend on itself",
			})
		}
		for _, d := range t.Deps {
			if !ids[d] {
				out = append(out, LintIssue{
					Severity: SeverityError,
					Code:     "dangling-dep",
					Message:  fmt.Sprintf("task #%d depends on missing #%d", t.ID, d),
					Location: fmt.Sprintf("tasks.md:#%d", t.ID),
					Hint:     fmt.Sprintf("either add a task #%d or remove this dependency", d),
				})
			}
		}
	}

	if cycle := taskgraph.DetectCycle(tasks); cycle != nil {
		out = append(out, LintIssue{
			Severity: SeverityError,
			Code:     "cycle",
			Message:  "dependency cycle detected: " + formatCycle(cycle),
			Location: "tasks.md",
			Hint:     "break the cycle by removing one dependency; run `free-kiro task list <spec>` to visualise the graph",
		})
	}
	return out
}

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func formatCycle(cycle []int) string {
	out := ""
	for i, id := range cycle {
		if i > 0 {
			out += " -> "
		}
		out += fmt.Sprintf("#%d", id)
	}
	return out
}

// ensure models import is referenced (build-only).
var _ = models.Task{}