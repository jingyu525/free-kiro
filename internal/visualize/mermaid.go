package visualize

import (
	"fmt"
	"io"
	"strings"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
)

// RenderMermaidSpec emits a Mermaid `graph LR` block describing one spec's
// tasks and their dependencies. Suitable for pasting into GitHub /
// GitLab markdown — both renderers understand the syntax.
//
// Output shape:
//
//	graph LR
//	  spec_<name>["my-spec<br/>phase: approved"]
//	  spec_<name> --> w1["Wave 1: 1 task"]
//	  spec_<name> --> w2["Wave 2: 1 task"]
//	  w1 --> w2
//
// Tasks within the same wave are grouped under a single node to keep the
// diagram readable. Use the `wave N` link to navigate to the textual
// task list (`free-kiro task list <spec>`).
func RenderMermaidSpec(w io.Writer, name string, phase models.Phase, tasks []models.Task, waves [][]models.Task) {
	fmt.Fprintf(w, "graph LR\n")
	// Spec node.
	fmt.Fprintf(w, "  spec_%s[\"%s<br/>phase: %s\"]\n", mermaidID(name), mermaidLabel(name), phase)
	for i, wave := range waves {
		wn := fmt.Sprintf("w%d_%s", i+1, mermaidID(name))
		count := len(wave)
		word := "tasks"
		if count == 1 {
			word = "task"
		}
		fmt.Fprintf(w, "  spec_%s --> %s[\"Wave %d: %d %s\"]\n", mermaidID(name), wn, i+1, count, word)
		_ = word
	}
	// Wave-to-wave dependencies.
	for i := 0; i+1 < len(waves); i++ {
		prev := fmt.Sprintf("w%d_%s", i+1, mermaidID(name))
		next := fmt.Sprintf("w%d_%s", i+2, mermaidID(name))
		fmt.Fprintf(w, "  %s --> %s\n", prev, next)
	}
}

// RenderMermaidProject emits a top-level diagram of all specs in the
// workspace, with their current phase. Useful as a project-wide overview
// at the top of REPORT.md.
func RenderMermaidProject(w io.Writer, specs []*models.SpecMeta) {
	fmt.Fprintln(w, "graph LR")
	for _, s := range specs {
		fmt.Fprintf(w, "  spec_%s[\"%s<br/>phase: %s\"]\n",
			mermaidID(s.Name), mermaidLabel(s.Name), s.Phase)
	}
}

// RenderMermaidDrift emits a table listing all drift signals across specs.
// Mermaid tables aren't rendered on all platforms — fall back to a
// markdown table when the consumer doesn't understand the syntax.
func RenderMermaidDrift(w io.Writer, entries []DriftEntry) {
	if len(entries) == 0 {
		return
	}
	fmt.Fprintln(w, "| spec | key | baseline | current | delta |")
	fmt.Fprintln(w, "|------|-----|----------|---------|-------|")
	for _, e := range entries {
		sign := ""
		switch {
		case e.Delta > 0:
			sign = "+"
		case e.Delta < 0:
			sign = "-"
		}
		fmt.Fprintf(w, "| %s | %s | %d | %d | %s%d |\n",
			e.Spec, e.Key, e.Baseline, e.Current, sign, absDelta(e.Delta))
	}
}

// DriftEntry is one row of the drift table emitted by RenderMermaidDrift.
type DriftEntry struct {
	Spec     string `json:"spec"`
	Key      string `json:"key"`
	Baseline int    `json:"baseline"`
	Current  int    `json:"current"`
	Delta    int    `json:"delta"`
}

// mermaidID makes a name safe to embed in a Mermaid node id. Strips
// dashes and any non-alphanumeric characters; idempotent.
func mermaidID(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// mermaidLabel sanitises a label for use inside square brackets.
// Quotes / brackets / pipes would break the diagram syntax.
func mermaidLabel(s string) string {
	return strings.NewReplacer("\"", "'", "[", "(", "]", ")", "\n", " ").Replace(s)
}

func absDelta(d int) int {
	if d < 0 {
		return -d
	}
	return d
}

// taskWaves is a convenience for callers that already have a Task slice
// and want the wave grouping without going through Engine. Always
// returns at least one wave (even an empty placeholder).
func taskWaves(tasks []models.Task) [][]models.Task {
	w := taskgraph.ExecutionWaves(tasks)
	if len(w) == 0 {
		return [][]models.Task{{}}
	}
	return w
}