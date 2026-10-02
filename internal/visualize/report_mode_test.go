package visualize

import (
	"encoding/json"
	"testing"

	"github.com/jingyu525/free-kiro/internal/models"
)

// stubWSMode wraps stubWS and lets each test toggle KiroDirExists()
// independently. The dashboard-backend-api-extensions spec requires
// this branching in BuildReport → Mode.
type stubWSMode struct {
	stubWS
	exists bool
}

func (s stubWSMode) KiroDirExists() bool { return s.exists }

// TestComputeReportMode covers the three-way tri-state the dashboard
// uses to render empty-state CTAs. Pure function, no fixtures.
func TestComputeReportMode(t *testing.T) {
	spec := &models.SpecMeta{Name: "demo", Phase: models.PhaseDraft}
	cases := []struct {
		name   string
		exists bool
		specs  []*models.SpecMeta
		want   string
	}{
		{"workspace missing", false, nil, "workspace-missing"},
		{"no specs but workspace exists", true, nil, "no-specs"},
		{"workspace + specs", true, []*models.SpecMeta{spec}, "ok"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws := stubWSMode{exists: c.exists}
			got := computeReportMode(ws, c.specs)
			if got != c.want {
				t.Errorf("computeReportMode = %q; want %q", got, c.want)
			}
		})
	}
}

// TestProjectReport_ModeField covers the JSON wire format: the new
// `mode` field must be present (frontend tri-state UI depends on it)
// and must serialize as a plain string (no nesting).
func TestProjectReport_ModeField(t *testing.T) {
	r := &ProjectReport{Mode: "workspace-missing"}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(b), `"mode":"workspace-missing"`) {
		t.Errorf("JSON = %s; want contains '\"mode\":\"workspace-missing\"'", string(b))
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
