package taskgraph

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/jingyu525/free-kiro/internal/models"
)

// genLinearTasks builds N tasks in a pure 1 → 2 → 3 → … → N chain. Worst
// case for the recursive DFS in computeLevels (every task depends on the
// previous one, so the call stack must hold N levels).
func BenchmarkExecutionWaves_Linear(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tasks := make([]models.Task, n)
			for i := range n {
				t := models.Task{ID: i + 1, Title: fmt.Sprintf("task %d", i+1)}
				if i > 0 {
					t.Deps = []int{i} // depends on previous
				}
				tasks[i] = t
			}
			b.ReportAllocs()
			for b.Loop() {
				_ = ExecutionWaves(tasks)
			}
		})
	}
}

// genWideTasks builds N tasks with random interconnections but
// artificially capped max-dependency depth (every task depends only on
// the first k tasks). This is a more realistic shape — parallel work
// batches where each task only references an earlier short list.
func BenchmarkExecutionWaves_Wide(b *testing.B) {
	rng := rand.New(rand.NewSource(42)) // deterministic for stable benchmark
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tasks := make([]models.Task, n)
			for i := range n {
				t := models.Task{ID: i + 1, Title: fmt.Sprintf("task %d", i+1)}
				// Each task depends on up to 3 earlier tasks (skips task 1).
				maxDeps := 3
				if i > 0 {
					count := rng.Intn(maxDeps + 1)
					deps := make([]int, 0, count)
					seen := map[int]bool{}
					for len(deps) < count {
						d := rng.Intn(i) + 1
						if !seen[d] {
							seen[d] = true
							deps = append(deps, d)
						}
					}
					t.Deps = deps
				}
				tasks[i] = t
			}
			b.ReportAllocs()
			for b.Loop() {
				_ = ExecutionWaves(tasks)
			}
		})
	}
}

// BenchmarkSummary measures the cost of Summary(), which delegates to
// ExecutionWaves. The benchmark exists so a regression in either wave
// grouping or task-counting is attributable.
func BenchmarkSummary(b *testing.B) {
	n := 200
	tasks := make([]models.Task, n)
	for i := range n {
		t := models.Task{ID: i + 1, Title: fmt.Sprintf("task %d", i+1)}
		if i > 0 {
			t.Deps = []int{i}
		}
		tasks[i] = t
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = Summary(tasks)
	}
}