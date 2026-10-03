package visualize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jingyu525/free-kiro/internal/models"
)

// BenchmarkProjectReport_JSON measures the cost of serialising a
// ProjectReport to JSON for the dashboard's /api/summary endpoint.
// Subtests cover the three sizes the dashboard will typically face.
//
// Note: BenchmarkRenderReport (the Markdown path) is intentionally NOT
// included here — RenderReport walks r.Specs and calls loadSpecTasks
// which depends on the package-level currentWorkspace. Setting that up
// belongs in an end-to-end test, not a unit benchmark. The JSON path is
// the steady-state hot path for the dashboard anyway.
func BenchmarkProjectReport_JSON(b *testing.B) {
	for _, specs := range []int{5, 20, 50} {
		b.Run(fmt.Sprintf("specs=%d", specs), func(b *testing.B) {
			rep := buildFixtureReport(specs)
			b.SetBytes(int64(specs * 200))
			b.ReportAllocs()
			for b.Loop() {
				data, err := json.Marshal(rep)
				if err != nil {
					b.Fatalf("json.Marshal: %v", err)
				}
				_ = data
			}
		})
	}
}

// BenchmarkTaskProgress_JSON measures the per-spec cost of the most
// frequent dashboard payload. The dashboard re-renders /api/specs on
// every SSE refresh; isolating the per-row cost helps attribute
// regressions.
func BenchmarkTaskProgress_JSON(b *testing.B) {
	for _, specs := range []int{20, 50, 200} {
		b.Run(fmt.Sprintf("specs=%d", specs), func(b *testing.B) {
			progress := make([]TaskProgress, specs)
			for i := range progress {
				progress[i] = TaskProgress{Done: i % 5, Total: 10, Waves: 3}
			}
			b.ReportAllocs()
			for b.Loop() {
				data, err := json.Marshal(progress)
				if err != nil {
					b.Fatalf("json.Marshal: %v", err)
				}
				_ = data
			}
		})
	}
}

// BenchmarkRenderMermaidProject measures the cost of generating the
// top-level Mermaid diagram for the project tree. RenderMermaidProject
// is pure (no IO, no workspace), so it can be benchmarked directly with
// a synthetic SpecMeta slice.
func BenchmarkRenderMermaidProject(b *testing.B) {
	for _, specs := range []int{5, 20, 50} {
		b.Run(fmt.Sprintf("specs=%d", specs), func(b *testing.B) {
			metas := buildFixtureMetas(specs)
			b.ReportAllocs()
			for b.Loop() {
				var buf bytes.Buffer
				RenderMermaidProject(&buf, metas)
				_ = buf.String()
			}
		})
	}
}

// buildFixtureReport builds a synthetic ProjectReport for benchmark
// input. The shape matches what `free-kiro report` actually emits in
// a project with N specs — same field set, same JSON tags, same
// GeneratedAt presence — so the JSON path is exercised realistically
// even though no .kiro/ tree exists on disk.
func buildFixtureReport(specs int) *ProjectReport {
	out := &ProjectReport{
		GeneratedAt: time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC),
		Active:      "perf-bench",
		Mode:        "ok",
	}
	for i := range specs {
		meta := &models.SpecMeta{
			Name:     fmt.Sprintf("spec-%d", i),
			SpecType: models.SpecTypeFeature,
			Workflow: models.WorkflowRequirementsFirst,
			Phase:    models.PhaseImplementing,
		}
		out.Specs = append(out.Specs, &SpecReport{
			Meta:    meta,
			Active:  i == 0,
			Current: map[string]int{"ac_count": 10, "task_count": 8},
			Tasks:   TaskProgress{Done: 5, Total: 8, Waves: 3},
			Drift:   nil,
		})
	}
	return out
}

// buildFixtureMetas returns a synthetic SpecMeta slice for benchmarks
// that only need the spec metadata (e.g. RenderMermaidProject).
func buildFixtureMetas(specs int) []*models.SpecMeta {
	out := make([]*models.SpecMeta, specs)
	for i := range out {
		out[i] = &models.SpecMeta{
			Name:     fmt.Sprintf("spec-%d", i),
			SpecType: models.SpecTypeFeature,
			Workflow: models.WorkflowRequirementsFirst,
			Phase:    models.PhaseImplementing,
		}
	}
	return out
}