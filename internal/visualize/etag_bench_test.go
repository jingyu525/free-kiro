package visualize

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"testing"
)

// BenchmarkStableETag measures the cost of computing the ETag header
// value for an HTTP response body. The benchmark covers three payload
// sizes that mirror real dashboard responses:
//
//   - 1 KB   — /api/health
//   - 10 KB  — /api/spec/<name>  (full spec status)
//   - 100 KB — /api/summary (full project report with all specs)
//
// stableETag also writes through an in-process cache (etagCache) keyed
// on (path, status, bodyLen), so the second hit is faster than the
// first — `b.Run` reuses the same (path, status, bodyLen) tuple, which
// is exactly what the dashboard does in steady state.
func BenchmarkStableETag(b *testing.B) {
	for _, size := range []int{1024, 10 * 1024, 100 * 1024} {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			body := make([]byte, size)
			if _, err := rand.Read(body); err != nil {
				b.Fatalf("rand.Read: %v", err)
			}
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for b.Loop() {
				_ = stableETag("/api/summary", 200, body)
			}
		})
	}
}

// BenchmarkStableETag_JSON covers the realistic case where body is a
// serialised ProjectReport. JSON encoding + SHA-256 is the hottest path
// the ETag middleware hits on every dashboard render.
func BenchmarkStableETag_JSON(b *testing.B) {
	type payload struct {
		Name      string `json:"name"`
		Phase     string `json:"phase"`
		Tasks     int    `json:"tasks"`
		LongField string `json:"long_field"`
	}
	makeBody := func(specs int) []byte {
		out := make([]payload, specs)
		for i := range out {
			out[i] = payload{
				Name:      fmt.Sprintf("spec-%d", i),
				Phase:     "implementing",
				Tasks:     10 + i,
				LongField: "abcdefghijklmnopqrstuvwxyz-0123456789-padding-for-realistic-payload-size",
			}
		}
		buf, err := json.Marshal(out)
		if err != nil {
			b.Fatalf("json.Marshal: %v", err)
		}
		return buf
	}
	for _, specs := range []int{5, 50, 500} {
		b.Run(fmt.Sprintf("specs=%d", specs), func(b *testing.B) {
			body := makeBody(specs)
			b.SetBytes(int64(len(body)))
			b.ReportAllocs()
			for b.Loop() {
				_ = stableETag("/api/specs", 200, body)
			}
		})
	}
}