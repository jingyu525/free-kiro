package models

// Task is a single executable task parsed from tasks.md.
//
// Format (matches kiro-clone / Kiro exactly):
//
//	- [ ] #1 Title
//	- [x] #2 Implement core [deps: #1]
//	- [X] #3 Polish [deps: #1,#2]
//
// Done reflects the [x] / [X] checkbox state. Deps are 1-based task ids
// parsed from [deps: #N,#M]; the engine schedules tasks into parallel
// *waves* using these.
type Task struct {
	ID    int
	Title string
	Deps  []int
	Done  bool
	Raw   string
}

// TaskSummary is the progress summary printed by `task list`.
type TaskSummary struct {
	Total     int
	Done      int
	Remaining int
	Waves     int
}