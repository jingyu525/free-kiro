package taskgraph

import (
	"testing"

	"github.com/jingyu525/free-kiro/internal/models"
)

func TestParseTasks_BasicAndDone(t *testing.T) {
	doc := `# Tasks

- [ ] #1 First
- [x] #2 Second [deps: #1]
- [X] #3 Third   [deps: #1,#2]
not a task line
- [ ] #4 no deps
`
	tasks := ParseTasks(doc)
	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks; got %d", len(tasks))
	}
	if tasks[0].ID != 1 || tasks[0].Done {
		t.Errorf("task 0 wrong: %+v", tasks[0])
	}
	if tasks[1].ID != 2 || !tasks[1].Done || len(tasks[1].Deps) != 1 || tasks[1].Deps[0] != 1 {
		t.Errorf("task 1 wrong: %+v", tasks[1])
	}
	if !tasks[2].Done {
		t.Errorf("task 2 should be Done (uppercase X)")
	}
	if len(tasks[2].Deps) != 2 {
		t.Errorf("task 2 deps: %v", tasks[2].Deps)
	}
	if tasks[3].Title != "no deps" {
		t.Errorf("task 3 title: %q", tasks[3].Title)
	}
}

func TestParseTasks_Empty(t *testing.T) {
	if got := ParseTasks("# Tasks\n\nNo items.\n"); len(got) != 0 {
		t.Errorf("expected empty result; got %v", got)
	}
}

func TestDetectCycle_None(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Deps: nil},
		{ID: 2, Deps: []int{1}},
		{ID: 3, Deps: []int{1}},
		{ID: 4, Deps: []int{2, 3}},
	}
	if c := DetectCycle(tasks); c != nil {
		t.Errorf("expected no cycle; got %v", c)
	}
}

func TestDetectCycle_Simple(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Deps: []int{2}},
		{ID: 2, Deps: []int{1}},
	}
	c := DetectCycle(tasks)
	if c == nil {
		t.Fatal("expected cycle")
	}
	if len(c) < 3 || c[0] != c[len(c)-1] {
		t.Errorf("cycle should close: %v", c)
	}
}

func TestDetectCycle_ThreeNodes(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Deps: []int{2}},
		{ID: 2, Deps: []int{3}},
		{ID: 3, Deps: []int{1}},
	}
	c := DetectCycle(tasks)
	if c == nil {
		t.Fatal("expected cycle in 1->2->3->1")
	}
}

func TestExecutionWaves(t *testing.T) {
	tasks := []models.Task{
		{ID: 1},
		{ID: 2},
		{ID: 3, Deps: []int{1}},
		{ID: 4, Deps: []int{1}},
		{ID: 5, Deps: []int{3, 4}},
	}
	waves := ExecutionWaves(tasks)
	if len(waves) != 3 {
		t.Fatalf("expected 3 waves; got %d", len(waves))
	}
	if len(waves[0]) != 2 {
		t.Errorf("wave 1 should have 2 tasks (#1,#2); got %d", len(waves[0]))
	}
	if len(waves[1]) != 2 {
		t.Errorf("wave 2 should have 2 tasks (#3,#4); got %d", len(waves[1]))
	}
	if len(waves[2]) != 1 || waves[2][0].ID != 5 {
		t.Errorf("wave 3 should be just #5; got %v", waves[2])
	}
}

func TestExecutionWaves_DanglingDepsIgnored(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Deps: []int{99}}, // dangling
	}
	waves := ExecutionWaves(tasks)
	if len(waves) != 1 || waves[0][0].ID != 1 {
		t.Errorf("dangling deps should still place task at level 1; got %v", waves)
	}
}

func TestExecutionWaves_CycleReturnsNil(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Deps: []int{2}},
		{ID: 2, Deps: []int{1}},
	}
	if waves := ExecutionWaves(tasks); waves != nil {
		t.Errorf("cycle should produce nil; got %v", waves)
	}
}

func TestSummary(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Done: true},
		{ID: 2, Done: false},
		{ID: 3, Deps: []int{1}},
	}
	s := Summary(tasks)
	if s.Total != 3 || s.Done != 1 || s.Remaining != 2 {
		t.Errorf("summary wrong: %+v", s)
	}
	if s.Waves != 2 {
		t.Errorf("expected 2 waves; got %d", s.Waves)
	}
}

func TestAtoi(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"", 0},
		{"abc", 0},
		{"12x", 12}, // stops at non-digit
	}
	for _, c := range cases {
		if got := atoi(c.in); got != c.want {
			t.Errorf("atoi(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseDepIDs(t *testing.T) {
	got := parseDepIDs("#1,#2, #3")
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("parseDepIDs: %v", got)
	}
	if got := parseDepIDs(""); got != nil {
		t.Errorf("empty deps should be nil; got %v", got)
	}
}
