package taskgraph

import (
	"github.com/liujingyu/free-kiro/internal/models"
)

// DetectCycle returns the cycle path (task ids) if tasks has a dependency
// cycle, or nil if the graph is acyclic.
//
// Uses DFS with WHITE / GRAY / BLACK colouring. A back-edge into a GRAY
// (currently on the DFS stack) node is a cycle; the returned slice is the
// stack from that node to the current one, with the node appended again
// to make the cycle explicit, e.g. [1, 2, 1] for #1 -> #2 -> #1.
//
// Dangling dependencies (ids with no matching task) are ignored — they
// are reported separately by lint as ERROR `dangling-dep`.
func DetectCycle(tasks []models.Task) []int {
	byID := map[int]models.Task{}
	for _, t := range tasks {
		byID[t.ID] = t
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[int]int{}
	for _, t := range tasks {
		color[t.ID] = white
	}
	var stack []int

	var dfs func(id int) []int
	dfs = func(id int) []int {
		color[id] = gray
		stack = append(stack, id)
		t := byID[id]
		for _, d := range t.Deps {
			if _, ok := byID[d]; !ok {
				continue
			}
			switch color[d] {
			case gray:
				// Back edge: extract the cycle slice from the stack.
				for i := len(stack) - 1; i >= 0; i-- {
					if stack[i] == d {
						cycle := append([]int{}, stack[i:]...)
						cycle = append(cycle, d)
						return cycle
					}
				}
				return []int{d, d} // shouldn't happen but defensive
			case white:
				if res := dfs(d); res != nil {
					return res
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
		return nil
	}

	for _, t := range tasks {
		if color[t.ID] == white {
			if res := dfs(t.ID); res != nil {
				return res
			}
		}
	}
	return nil
}