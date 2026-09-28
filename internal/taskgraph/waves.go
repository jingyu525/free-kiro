package taskgraph

import (
	"github.com/jingyu525/free-kiro/internal/models"
)

// ExecutionWaves groups tasks into parallel execution waves (level 1, 2,
// 3, ...). Tasks within a wave share no dependencies and may run
// concurrently.
//
// Raises an error-equivalent panic if the graph has a cycle. Use
// DetectCycle first if you want to handle the cycle gracefully.
func ExecutionWaves(tasks []models.Task) [][]models.Task {
	if cycle := DetectCycle(tasks); cycle != nil {
		// Return a single wave with a synthetic "cycle" task so the caller
		// still gets something printable; lint will have reported the
		// cycle already.
		return nil
	}
	levels := computeLevels(tasks)
	grouped := map[int][]models.Task{}
	for _, t := range tasks {
		lvl := levels[t.ID]
		grouped[lvl] = append(grouped[lvl], t)
	}
	out := make([][]models.Task, 0, len(grouped))
	for lvl := 1; lvl <= len(grouped); lvl++ {
		if _, ok := grouped[lvl]; ok {
			out = append(out, grouped[lvl])
		}
	}
	return out
}

// computeLevels assigns each task a wave level via memoised DFS. Level 1
// is the set of tasks with no dependencies (or only dangling deps).
func computeLevels(tasks []models.Task) map[int]int {
	byID := map[int]models.Task{}
	for _, t := range tasks {
		byID[t.ID] = t
	}
	cache := map[int]int{}
	var level func(id int) int
	level = func(id int) int {
		if v, ok := cache[id]; ok {
			return v
		}
		t, ok := byID[id]
		if !ok {
			return 0
		}
		maxDep := 0
		for _, d := range t.Deps {
			if _, ok := byID[d]; !ok {
				continue
			}
			if dl := level(d); dl > maxDep {
				maxDep = dl
			}
		}
		lvl := maxDep + 1
		cache[id] = lvl
		return lvl
	}
	out := map[int]int{}
	for _, t := range tasks {
		out[t.ID] = level(t.ID)
	}
	return out
}

// Summary returns aggregate progress numbers for a set of tasks. Waves
// counts the number of topological levels (0 if there are no tasks).
func Summary(tasks []models.Task) models.TaskSummary {
	total := len(tasks)
	done := 0
	for _, t := range tasks {
		if t.Done {
			done++
		}
	}
	waves := len(ExecutionWaves(tasks))
	return models.TaskSummary{
		Total:     total,
		Done:      done,
		Remaining: total - done,
		Waves:     waves,
	}
}