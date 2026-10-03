package lint

import (
	"testing"

	"github.com/jingyu525/free-kiro/internal/testutil/perf"
)

// BenchmarkEARSRe_FindAllString measures the cost of scanning an entire
// spec document with the combined EARSRe. Production code calls
// FindAllString inside lint.Requirements for every AC line candidate, so
// the per-spec cost is what matters — the time to compile the regex is
// excluded because package-level var initialisation runs once at import.
//
// Subtests cover the three fixture sizes declared in testdata/perf/.
// `b.ResetTimer()` is not needed: the fixture is loaded from the
// in-memory cache (testutil/perf.Load), so the first iteration cost is
// the matcher's hot path, not file IO.
func BenchmarkEARSRe_FindAllString(b *testing.B) {
	for _, size := range []perf.Size{perf.Small, perf.Medium, perf.Large} {
		b.Run(sizeName(size), func(b *testing.B) {
			data := perf.Load(b, size)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				_ = EARSRe.FindAllString(string(data), -1)
			}
		})
	}
}

// BenchmarkEARSRe_MatchString measures the cheaper per-call cost used by
// the per-template regexes (WHENRe, WHILERe, …) and by single-line AC
// classification. Each iteration calls MatchString on the same string;
// the work per call is O(K) on input length.
func BenchmarkEARSRe_MatchString(b *testing.B) {
	for _, size := range []perf.Size{perf.Small, perf.Medium, perf.Large} {
		b.Run(sizeName(size), func(b *testing.B) {
			data := perf.Load(b, size)
			doc := string(data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				_ = EARSRe.MatchString(doc)
			}
		})
	}
}

// BenchmarkUsesIFTHEN covers the IF-THEN template check, which is the
// only EARS template not expressible as a single regex (two lazy
// `.+?\s+` segments cannot backtrack across each other in RE2). The
// bench keeps the IF-THEN path on par with the cheaper regex tests.
func BenchmarkUsesIFTHEN(b *testing.B) {
	for _, size := range []perf.Size{perf.Small, perf.Medium, perf.Large} {
		b.Run(sizeName(size), func(b *testing.B) {
			data := perf.Load(b, size)
			doc := string(data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				_ = UsesIFTHEN(doc)
			}
		})
	}
}

func sizeName(s perf.Size) string {
	switch s {
	case perf.Small:
		return "small"
	case perf.Medium:
		return "medium"
	case perf.Large:
		return "large"
	default:
		return "unknown"
	}
}