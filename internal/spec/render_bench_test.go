package spec

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"
)

// BenchmarkRenderRequirements measures the cost of text/template
// parsing + execution for the requirements.md.tmpl planning document.
//
// We compile the template exactly once outside the timed loop — the
// production path also pays this parse cost once per process (the
// templateFS embed is loaded at package init, not per Generate call).
// Inside the loop we only measure Execute, which is the steady-state
// per-spec cost users actually feel.
func BenchmarkRenderRequirements(b *testing.B) {
	src := loadTemplateSource(b, "requirements.md.tmpl")
	tmpl := template.Must(template.New("requirements.md").Parse(src))

	data := templateData{Name: "perf-bench", SpecType: "feature"}
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatalf("Execute: %v", err)
		}
		_ = buf.Bytes()
	}
}

// BenchmarkRenderDesign exercises the design template branch — same
// engine, different document. Catches regressions where design-specific
// template syntax gets accidentally slower than the requirements path.
func BenchmarkRenderDesign(b *testing.B) {
	src := loadTemplateSource(b, "design.md.tmpl")
	tmpl := template.Must(template.New("design.md").Parse(src))

	data := templateData{Name: "perf-bench", SpecType: "feature"}
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatalf("Execute: %v", err)
		}
		_ = buf.Bytes()
	}
}

// BenchmarkRenderBugfix exercises the bugfix template branch — same
// engine, different document. Bugfix specs follow the three-section
// (Current / Expected / Unchanged) shape, which may exercise different
// template parsing branches than feature docs.
func BenchmarkRenderBugfix(b *testing.B) {
	src := loadTemplateSource(b, "bugfix.md.tmpl")
	tmpl := template.Must(template.New("bugfix.md").Parse(src))

	data := templateData{Name: "perf-bug", SpecType: "bugfix"}
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			b.Fatalf("Execute: %v", err)
		}
		_ = buf.Bytes()
	}
}

// BenchmarkRenderParse is the upfront cost the user pays the FIRST time
// Generate("foo", PhaseX) is called inside a fresh process:
// text/template parse + validation. Production code amortises this
// across all subsequent specs, but a freshly-started CLI process pays
// it once per template.
func BenchmarkRenderParse(b *testing.B) {
	for _, tpl := range []string{"requirements.md.tmpl", "design.md.tmpl", "tasks.md.tmpl"} {
		b.Run(tpl, func(b *testing.B) {
			src := loadTemplateSource(b, tpl)
			b.ReportAllocs()
			for b.Loop() {
				_, err := template.New(tpl).Parse(src)
				if err != nil {
					b.Fatalf("Parse: %v", err)
				}
			}
		})
	}
}

// loadTemplateSource reads one of the embedded templates via the file
// system (rather than the embed.FS) because the benchmark file lives
// in the same package as generator.go and we don't want to expose the
// embed handle. The read happens once per bench function, outside the
// timed loop.
func loadTemplateSource(tb testing.TB, name string) string {
	tb.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatalf("runtime.Caller(0) failed")
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(here), "templates", name))
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read template %s: %v", path, err)
	}
	return string(data)
}