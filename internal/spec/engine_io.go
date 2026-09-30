package spec

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
)

// loadMeta reads .meta.json from the spec's directory.
func (e *Engine) loadMeta(specName string) (*models.SpecMeta, error) {
	return models.LoadSpecMeta(e.ws.SpecDir(specName))
}

// countTasks / countWaves parse tasks.md just enough to report progress.
// The CLI's `task list` command uses taskgraph directly for the full
// dependency-aware wave view; this is the lightweight summary used in
// status / analyze output.
func countTasks(specDir string) (done, total int) {
	data, err := os.ReadFile(filepath.Join(specDir, "tasks.md"))
	if err != nil {
		return 0, 0
	}
	for _, line := range splitLines(string(data)) {
		m := taskLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		total++
		if m[1] == "x" || m[1] == "X" {
			done++
		}
	}
	return done, total
}

// countWaves returns the number of topological levels in tasks.md. Runs
// the dependency-aware topological leveller via taskgraph.
func countWaves(specDir string) int {
	data, err := os.ReadFile(filepath.Join(specDir, "tasks.md"))
	if err != nil {
		return 0
	}
	tasks := taskgraph.ParseTasks(string(data))
	return len(taskgraph.ExecutionWaves(tasks))
}

var (
	taskLineRe = regexp.MustCompile(`^\s*-\s*\[( |x|X)\]\s*#(\d+)\s+`)
)

func splitLines(text string) []string {
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
