package spec

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jingyu525/free-kiro/internal/workspace"
)

// TestEngine_TaskList_TableDriven covers the six outcomes documented
// in spec-5 task #8 AC. Uses a real workspace.Workspace rooted at a
// t.TempDir() so we exercise the same code path production does
// (Engine.ws is *workspace.Workspace, not an interface).
func TestEngine_TaskList_TableDriven(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T, root string) // creates .kiro/specs/<name>/ + tasks.md
		specName  string
		wantErr   error
		wantWaves int
		wantDone  int
		wantTotal int
	}{
		{
			name: "happy path: 3 tasks, 1 wave (no deps)",
			setup: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".kiro", "specs", "demo"))
				mustWrite(t, filepath.Join(root, ".kiro", "specs", "demo", "tasks.md"),
					"- [ ] #1 Task one\n- [x] #2 Task two\n- [ ] #3 Task three\n")
			},
			specName: "demo", wantWaves: 1, wantDone: 1, wantTotal: 3,
		},
		{
			name: "two waves via deps",
			setup: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".kiro", "specs", "deps"))
				mustWrite(t, filepath.Join(root, ".kiro", "specs", "deps", "tasks.md"),
					"- [ ] #1 first\n- [ ] #2 second [deps: #1]\n- [ ] #3 third [deps: #2]\n")
			},
			specName: "deps", wantWaves: 3, wantDone: 0, wantTotal: 3,
		},
		{
			name: "empty tasks.md",
			setup: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".kiro", "specs", "empty"))
				mustWrite(t, filepath.Join(root, ".kiro", "specs", "empty", "tasks.md"), "")
			},
			specName: "empty", wantWaves: 0, wantDone: 0, wantTotal: 0,
		},
		{
			name: "tasks.md with only free-form prose (no checkbox lines)",
			setup: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".kiro", "specs", "prose"))
				mustWrite(t, filepath.Join(root, ".kiro", "specs", "prose", "tasks.md"),
					"# heading\n\njust prose.\n")
			},
			specName: "prose", wantWaves: 0, wantDone: 0, wantTotal: 0,
		},
		{
			name: "missing tasks.md (spec in planning)",
			setup: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".kiro", "specs", "planning"))
			},
			specName: "planning", wantWaves: 0, wantDone: 0, wantTotal: 0,
		},
		{
			name:     "missing spec dir → ErrSpecNotFound",
			setup:    func(_ *testing.T, _ string) {}, // no .kiro/specs/<name>/
			specName: "ghost",
			wantErr:  ErrSpecNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			c.setup(t, root)
			ws := workspace.New(root)
			eng := &Engine{ws: ws}

			waves, err := eng.TaskList(c.specName)
			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("err = %v; want %v", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			gotWaves := len(waves)
			gotDone := 0
			gotTotal := 0
			for _, wv := range waves {
				gotDone += wv.Done
				gotTotal += wv.Total
			}
			if gotWaves != c.wantWaves {
				t.Errorf("waves = %d; want %d", gotWaves, c.wantWaves)
			}
			if gotDone != c.wantDone {
				t.Errorf("done = %d; want %d", gotDone, c.wantDone)
			}
			if gotTotal != c.wantTotal {
				t.Errorf("total = %d; want %d", gotTotal, c.wantTotal)
			}
		})
	}
}

// TestEngine_TaskList_WaveIndexing verifies the 1-based index invariant:
// wave 1 = no-dep tasks, wave 2 = depends on wave 1, etc.
func TestEngine_TaskList_WaveIndexing(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, ".kiro", "specs", "w"))
	mustWrite(t, filepath.Join(root, ".kiro", "specs", "w", "tasks.md"),
		"- [ ] #1 alpha\n- [ ] #2 beta [deps: #1]\n- [x] #3 gamma [deps: #1]\n")
	ws := workspace.New(root)
	eng := &Engine{ws: ws}

	waves, err := eng.TaskList("w")
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if len(waves) != 2 {
		t.Fatalf("len(waves) = %d; want 2", len(waves))
	}
	if waves[0].Index != 1 {
		t.Errorf("waves[0].Index = %d; want 1", waves[0].Index)
	}
	if waves[1].Index != 2 {
		t.Errorf("waves[1].Index = %d; want 2", waves[1].Index)
	}
	// Wave 1 has 1 task (#1); wave 2 has 2 tasks (#2, #3).
	if waves[0].Total != 1 || waves[1].Total != 2 {
		t.Errorf("wave sizes = (%d, %d); want (1, 2)", waves[0].Total, waves[1].Total)
	}
	if waves[1].Done != 1 {
		t.Errorf("wave 2 Done = %d; want 1", waves[1].Done)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
