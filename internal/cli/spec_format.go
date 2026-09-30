package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
	"github.com/jingyu525/free-kiro/internal/visualize"
)

// renderHumanStatus flattens the spec status into a top-level view.
// Drift signals are surfaced prominently at the top — they're the
// reason most users run `status`.
func renderHumanStatus(w io.Writer, name string, s map[string]any) {
	fmt.Fprintf(w, "%s\n", name)
	fmt.Fprintf(w, "  phase:      %s\n", s["phase"])
	fmt.Fprintf(w, "  workflow:   %s\n", s["workflow"])
	fmt.Fprintf(w, "  spec_type:  %s\n", s["spec_type"])
	if approved, _ := s["approved"].(bool); approved {
		fmt.Fprintln(w, "  approved:   yes")
	} else {
		fmt.Fprintln(w, "  approved:   no")
	}

	// Tasks line (handy quick view).
	if t, ok := s["tasks"].(map[string]any); ok {
		done, _ := t["done"].(int)
		total, _ := t["total"].(int)
		waves, _ := t["waves"].(int)
		fmt.Fprintf(w, "  tasks:      %d/%d done, %d wave(s)\n", done, total, waves)
	}

	// Drift — promoted to the top because it's the action item.
	if drift, ok := s["drift"].([]any); ok && len(drift) > 0 {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  DRIFT (baseline → current):")
		for _, item := range drift {
			d, ok := item.(map[string]any)
			if !ok {
				continue
			}
			key, _ := d["key"].(string)
			base, _ := d["baseline"].(int)
			cur, _ := d["current"].(int)
			delta, _ := d["delta"].(int)
			sign := " "
			if delta > 0 {
				sign = "+"
			} else if delta < 0 {
				sign = "-"
			}
			fmt.Fprintf(w, "    %s: %d → %d  (%s%d)\n",
				key, base, cur, sign, absDelta(delta))
		}
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  fix: either revert the change, or run `free-kiro spec sync <name>` to accept it as the new baseline")
	} else {
		fmt.Fprintln(w, "  drift:      none")
	}

	// Baseline / current snapshot at the bottom (for context).
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  baseline:")
	if b, ok := s["baseline"].(map[string]int); ok {
		for k, v := range b {
			fmt.Fprintf(w, "    %s: %d\n", k, v)
		}
	}
	fmt.Fprintln(w, "  current:")
	if c, ok := s["current"].(map[string]int); ok {
		for k, v := range c {
			fmt.Fprintf(w, "    %s: %d\n", k, v)
		}
	}
}

func absDelta(d int) int {
	if d < 0 {
		return -d
	}
	return d
}

// renderMermaidStatus emits a Mermaid graph LR block for one spec.
// Falls back to a no-op graph (with a comment line) when the spec
// has no tasks.md yet.
func renderMermaidStatus(w io.Writer, eng *spec.Engine, name string, status map[string]any) {
	phase := models.Phase("")
	if p, ok := status["phase"].(string); ok {
		phase = models.Phase(p)
	}
	tasks, waves := loadTasksForMermaid(eng, name)
	if len(tasks) == 0 {
		fmt.Fprintln(w, "graph LR")
		fmt.Fprintf(w, "  spec_%s[\"%s<br/>phase: %s<br/>no tasks yet\"]\n",
			name, name, phase)
		return
	}
	visualize.RenderMermaidSpec(w, name, phase, tasks, waves)
}

// loadTasksForMermaid reads tasks.md and returns the wave grouping.
func loadTasksForMermaid(eng *spec.Engine, name string) ([]models.Task, [][]models.Task) {
	data, err := os.ReadFile(filepath.Join(eng.WS().SpecDir(name), "tasks.md"))
	if err != nil {
		return nil, nil
	}
	tasks := taskgraph.ParseTasks(string(data))
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks, taskgraph.ExecutionWaves(tasks)
}
