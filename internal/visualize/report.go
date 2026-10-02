// Package visualize — report.go: aggregate project state into a
// Markdown report suitable for committing alongside the code (or
// pasting into PR descriptions / docs).
package visualize

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// SpecReport is the per-spec subset included in a project report.
type SpecReport struct {
	Meta    *models.SpecMeta `json:"meta"`
	Current map[string]int   `json:"current"`
	Drift   []DriftEntry     `json:"drift"`
	Tasks   TaskProgress     `json:"tasks"`
	Active  bool             `json:"active"`
}

// TaskProgress captures the implementation status for one spec.
type TaskProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
	Waves int `json:"waves"`
}

// ProjectReport is the top-level report aggregation.
type ProjectReport struct {
	GeneratedAt time.Time     `json:"generated_at"`
	Specs       []*SpecReport `json:"specs"`
	Active      string        `json:"active"` // name of active spec (from .kiro/.current)
	Mode        string        `json:"mode"`   // "workspace-missing" | "no-specs" | "ok" — drives dashboard empty-state tri-state
}

// BuildReport gathers everything needed for a project report. Engine
// + Workspace are the only collaborators — keeps the visualisation
// layer independent of CLI concerns.
//
// `ws` is the WorkspacePaths interface declared in server.go — Go
// interfaces are structural so report.go can use the wider Server
// interface as long as it only touches ReadCurrent().
func BuildReport(ws WorkspacePaths, eng *spec.Engine) (*ProjectReport, error) {
	specs, err := eng.ListSpecs()
	if err != nil {
		return nil, err
	}
	active := ws.ReadCurrent()
	out := &ProjectReport{
		GeneratedAt: time.Now().UTC(),
		Active:      active,
		Mode:        computeReportMode(ws, specs),
	}
	for _, m := range specs {
		st, err := eng.Status(m.Name)
		if err != nil {
			// Skip on per-spec failure rather than abort the whole report.
			continue
		}
		sr := &SpecReport{
			Meta:    m,
			Active:  m.Name == active,
			Current: extractIntMap(st["current"]),
			Tasks:   taskProgressFromStatus(st),
		}
		if drift, ok := st["drift"].([]any); ok {
			for _, raw := range drift {
				dm, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				sr.Drift = append(sr.Drift, DriftEntry{
					Spec:     m.Name,
					Key:      asString(dm["key"]),
					Baseline: asInt(dm["baseline"]),
					Current:  asInt(dm["current"]),
					Delta:    asInt(dm["delta"]),
				})
			}
		}
		out.Specs = append(out.Specs, sr)
	}
	return out, nil
}

func extractIntMap(v any) map[string]int {
	m, _ := v.(map[string]any)
	out := map[string]int{}
	for k, val := range m {
		out[k] = asInt(val)
	}
	return out
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}

func taskProgressFromStatus(st map[string]any) TaskProgress {
	t, _ := st["tasks"].(map[string]any)
	return TaskProgress{
		Done:  asInt(t["done"]),
		Total: asInt(t["total"]),
		Waves: asInt(t["waves"]),
	}
}

// RenderReport writes a complete markdown report.
func RenderReport(w io.Writer, r *ProjectReport) {
	_, _ = fmt.Fprintf(w, "# free-kiro Report\n\n")
	_, _ = fmt.Fprintf(w, "generated: %s\n\n", r.GeneratedAt.Format(time.RFC3339))

	phases := map[models.Phase]int{}
	for _, s := range r.Specs {
		phases[s.Meta.Phase]++
	}
	active := 0
	for _, s := range r.Specs {
		if s.Active {
			active++
		}
	}
	_, _ = fmt.Fprintf(w, "specs: %d total (%d done, %d active, %d in planning, %d draft)\n\n",
		len(r.Specs), phases[models.PhaseDone], active,
		phases[models.PhaseRequirements]+phases[models.PhaseDesign]+phases[models.PhaseTasks],
		phases[models.PhaseDraft])

	// Summary table.
	_, _ = fmt.Fprintln(w, "## Summary")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "| spec | phase | approved | drift | tasks |")
	_, _ = fmt.Fprintln(w, "|------|-------|----------|-------|-------|")
	for _, s := range r.Specs {
		marker := ""
		if s.Active {
			marker = " **(active)**"
		}
		drift := "none"
		if len(s.Drift) > 0 {
			drift = fmt.Sprintf("%d keys", len(s.Drift))
		}
		tasks := fmt.Sprintf("%d/%d", s.Tasks.Done, s.Tasks.Total)
		if s.Tasks.Total == 0 {
			tasks = "—"
		}
		_, _ = fmt.Fprintf(w, "| %s%s | %s | %s | %s | %s |\n",
			s.Meta.Name, marker, s.Meta.Phase,
			yesNo(s.Meta.Approved), drift, tasks)
	}
	_, _ = fmt.Fprintln(w)

	// Drift alerts.
	var allDrift []DriftEntry
	for _, s := range r.Specs {
		allDrift = append(allDrift, s.Drift...)
	}
	if len(allDrift) > 0 {
		_, _ = fmt.Fprintln(w, "## Drift alerts")
		_, _ = fmt.Fprintln(w)
		RenderMermaidDrift(w, allDrift)
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "Fix: `free-kiro spec sync <spec>` to accept as new baseline, or revert.")
		_, _ = fmt.Fprintln(w)
	}

	// Project overview mermaid.
	if len(r.Specs) > 0 {
		_, _ = fmt.Fprintln(w, "## Project overview")
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "```mermaid")
		specMetas := make([]*models.SpecMeta, 0, len(r.Specs))
		for _, s := range r.Specs {
			specMetas = append(specMetas, s.Meta)
		}
		RenderMermaidProject(w, specMetas)
		_, _ = fmt.Fprintln(w, "```")
		_, _ = fmt.Fprintln(w)
	}

	// Per-spec detail.
	for _, s := range r.Specs {
		_, _ = fmt.Fprintf(w, "## %s\n\n", s.Meta.Name)
		_, _ = fmt.Fprintf(w, "- phase: %s\n", s.Meta.Phase)
		_, _ = fmt.Fprintf(w, "- workflow: %s\n", s.Meta.Workflow)
		_, _ = fmt.Fprintf(w, "- spec_type: %s\n", s.Meta.SpecType)
		_, _ = fmt.Fprintf(w, "- tasks: %d/%d done, %d wave(s)\n",
			s.Tasks.Done, s.Tasks.Total, s.Tasks.Waves)
		_, _ = fmt.Fprintln(w)
		// Embedded mermaid for this spec.
		tasks, waves := loadSpecTasks(s.Meta.Name)
		if len(tasks) > 0 {
			_, _ = fmt.Fprintln(w, "```mermaid")
			RenderMermaidSpec(w, s.Meta.Name, s.Meta.Phase, tasks, waves)
			_, _ = fmt.Fprintln(w, "```")
			_, _ = fmt.Fprintln(w)
		}
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// loadSpecTasks reads a spec's tasks.md and returns tasks + wave grouping.
// Returns nil/empty when the file doesn't exist (spec still in planning).
func loadSpecTasks(name string) ([]models.Task, [][]models.Task) {
	// We need the workspace to resolve the path; the caller passes it
	// via the package-level currentWorkspace (set by RenderAndWriteReport).
	data, err := os.ReadFile(currentWorkspace.Sp(name))
	if err != nil {
		return nil, nil
	}
	tasks := taskgraph.ParseTasks(string(data))
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks, taskgraph.ExecutionWaves(tasks)
}

// computeReportMode classifies the workspace state for the dashboard's
// empty-state tri-state. O(1) — single stat + len check. The frontend
// (`dashboard-frontend-foundation`) renders different CTAs per mode:
//   - "workspace-missing" → "Run `free-kiro init`"
//   - "no-specs"          → "Run `free-kiro spec new <name>`"
//   - "ok"                → render the normal table
//
// computeReportMode is a pure function so report_test.go can table-drive
// it without touching the filesystem.
func computeReportMode(ws WorkspacePaths, specs []*models.SpecMeta) string {
	if !ws.KiroDirExists() {
		return "workspace-missing"
	}
	if len(specs) == 0 {
		return "no-specs"
	}
	return "ok"
}

// workspaceRef is a thin pointer-ish struct for the report module to
// resolve per-spec directories without depending on a specific
// Workspace type at compile time.
type workspaceRef struct {
	Sp func(name string) string
}

// currentWorkspace is set by RenderAndWriteReport before any per-spec
// rendering happens. Set-once, read-many semantics; safe in single-
// threaded CLI use (which is the only caller).
var currentWorkspace workspaceRef

// RenderAndWriteReport is the high-level entry point: build the report
// from a workspace + engine, write it to `path` (typically
// .kiro/REPORT.md). Empty path or `-` writes to stdout.
func RenderAndWriteReport(path string, ws *workspace.Workspace, eng *spec.Engine) (*ProjectReport, error) {
	currentWorkspace = workspaceRef{Sp: ws.SpecDir}
	r, err := BuildReport(ws, eng)
	if err != nil {
		return nil, err
	}
	f, err := openWriter(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	RenderReport(f, r)
	return r, nil
}
