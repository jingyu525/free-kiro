package lint

import (
	"testing"

	"github.com/jingyu525/free-kiro/internal/testutil/perf"
)

// BenchmarkRequirements measures the cost of the full shape gate that
// `free-kiro lint` runs on every spec: EARSRe placement, the 10 quality
// rules in quality.go (ears-few-ac, ears-low-template-diversity,
// ears-ac-missing-id, ears-response-immeasurable, ears-trigger-unobservable,
// ears-keyword-misuse-while-as-when, ears-passive-response, ears-etc-list,
// ears-multi-shall-line, placeholder-ac), and the structural checks in
// requirements.go. The benchmark is the headline number for any refactor
// of the lint engine — anything that pushes it above baseline should be
// either justified or reverted.
func BenchmarkRequirements(b *testing.B) {
	for _, size := range []perf.Size{perf.Small, perf.Medium, perf.Large} {
		b.Run(sizeName(size), func(b *testing.B) {
			data := perf.Load(b, size)
			doc := string(data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				_ = Requirements(doc)
			}
		})
	}
}

// BenchmarkExtractEARSLines measures the cost of the per-AC extraction
// helper used by both the lint gate and spec analyze. Decoupling from
// BenchmarkRequirements lets us attribute regressions to the right place
// when the issue is in the regex-based line classifier.
func BenchmarkExtractEARSLines(b *testing.B) {
	for _, size := range []perf.Size{perf.Small, perf.Medium, perf.Large} {
		b.Run(sizeName(size), func(b *testing.B) {
			data := perf.Load(b, size)
			doc := string(data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				_ = extractEARSLines(doc)
			}
		})
	}
}