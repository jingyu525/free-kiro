package visualize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jingyu525/free-kiro/internal/models"
)

func TestRenderTree_BasicShape(t *testing.T) {
	root := &TreeNode{
		Text: "root",
		Children: []*TreeNode{
			{Text: "a", Children: []*TreeNode{
				{Text: "a1"},
				{Text: "a2"},
			}},
			{Text: "b"},
		},
	}
	got := RenderTree(root)
	// Check key structural elements.
	if !strings.HasPrefix(got, "root\n") {
		t.Errorf("missing root line: %q", got)
	}
	if !strings.Contains(got, "├─ a") {
		t.Errorf("a connector missing: %q", got)
	}
	if !strings.Contains(got, "└─ b") {
		t.Errorf("b connector missing: %q", got)
	}
	if !strings.Contains(got, "│  ├─ a1") {
		t.Errorf("a1 nested connector missing: %q", got)
	}
	if !strings.Contains(got, "│  └─ a2") {
		t.Errorf("a2 nested connector missing: %q", got)
	}
	// b is root's last child: should sit flush (no continuation indent
	// since there's no │ above it to continue).
	if strings.Contains(got, "   └─ b") {
		t.Errorf("b should not have continuation indent (a is last-of-siblings): %q", got)
	}
}

func TestRenderTree_Meta(t *testing.T) {
	root := &TreeNode{
		Text:     "node",
		Meta:     map[string]string{"phase": "approved", "tasks": "3/3 done"},
		Children: nil,
	}
	got := RenderTree(root)
	if !strings.Contains(got, "node\n") {
		t.Errorf("node missing: %q", got)
	}
	// Meta lines should follow the node text on separate lines, aligned
	// to column 0 (since root has no connector).
	if !strings.Contains(got, "phase: approved") {
		t.Errorf("meta phase missing: %q", got)
	}
	if !strings.Contains(got, "tasks: 3/3 done") {
		t.Errorf("meta tasks missing: %q", got)
	}
}

func TestRenderMermaidSpec(t *testing.T) {
	var buf bytes.Buffer
	tasks := []models.Task{
		{ID: 1, Title: "build form"},
		{ID: 2, Title: "wire auth", Deps: []int{1}},
	}
	waveGroups := [][]models.Task{
		{{ID: 1, Title: "build form"}},
		{{ID: 2, Title: "wire auth", Deps: []int{1}}},
	}
	RenderMermaidSpec(&buf, "demo", "approved", tasks, waveGroups)
	got := buf.String()
	if !strings.HasPrefix(got, "graph LR\n") {
		t.Errorf("missing graph LR header: %q", got)
	}
	if !strings.Contains(got, "spec_demo") {
		t.Errorf("spec node missing: %q", got)
	}
	if !strings.Contains(got, "Wave 1:") || !strings.Contains(got, "Wave 2:") {
		t.Errorf("wave labels missing: %q", got)
	}
	if !strings.Contains(got, "w1_demo --> w2_demo") {
		t.Errorf("wave-to-wave edge missing: %q", got)
	}
}

func TestRenderMermaidSpec_EmptyTasks(t *testing.T) {
	var buf bytes.Buffer
	RenderMermaidSpec(&buf, "empty", "draft", nil, nil)
	got := buf.String()
	if !strings.Contains(got, "graph LR") {
		t.Errorf("header missing: %q", got)
	}
	if !strings.Contains(got, "spec_empty") {
		t.Errorf("spec node missing for empty: %q", got)
	}
}

func TestRenderMermaidDrift(t *testing.T) {
	var buf bytes.Buffer
	RenderMermaidDrift(&buf, []DriftEntry{
		{Spec: "demo", Key: "ac_count", Baseline: 3, Current: 4, Delta: 1},
		{Spec: "demo", Key: "task_count", Baseline: 3, Current: 2, Delta: -1},
	})
	got := buf.String()
	if !strings.HasPrefix(got, "| spec | key") {
		t.Errorf("header missing: %q", got)
	}
	if !strings.Contains(got, "| demo | ac_count | 3 | 4 | +1 |") {
		t.Errorf("positive delta row missing: %q", got)
	}
	if !strings.Contains(got, "| demo | task_count | 3 | 2 | -1 |") {
		t.Errorf("negative delta row missing: %q", got)
	}
}

func TestMermaidID_Sanitises(t *testing.T) {
	cases := []string{"my-spec", "feature/auth", "x.y.z", "123abc"}
	for _, in := range cases {
		got := mermaidID(in)
		for _, r := range got {
			isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			isDigit := r >= '0' && r <= '9'
			isUnderscore := r == '_'
			if !isLetter && !isDigit && !isUnderscore {
				t.Errorf("mermaidID(%q) = %q, contains non-allowed %q", in, got, r)
			}
		}
	}
}

func TestMermaidLabel_StripsQuotes(t *testing.T) {
	got := mermaidLabel(`hello "world" [bracket]`)
	for _, ch := range []string{`"`, `[`, `]`} {
		if strings.Contains(got, ch) {
			t.Errorf("mermaidLabel(%q) = %q, still contains %q", `hello "world" [bracket]`, got, ch)
		}
	}
}

func TestAbsDelta(t *testing.T) {
	if absDelta(-5) != 5 {
		t.Error("absDelta(-5) should be 5")
	}
	if absDelta(5) != 5 {
		t.Error("absDelta(5) should be 5")
	}
	if absDelta(0) != 0 {
		t.Error("absDelta(0) should be 0")
	}
}

func TestRenderReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	RenderReport(&buf, &ProjectReport{GeneratedAt: timeMustParse("2026-09-28T13:30:00Z")})
	got := buf.String()
	if !strings.Contains(got, "# free-kiro Report") {
		t.Errorf("missing header: %q", got)
	}
	if !strings.Contains(got, "specs: 0 total") {
		t.Errorf("missing counts: %q", got)
	}
}
