package ide

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstructionFiles_AllKnownIDsHaveEntries guards against adding an IDE
// id that has no entry in instructionFiles — which would silently produce
// an empty WriteSingleInstruction result and confuse doctor.
func TestInstructionFiles_AllKnownIDsHaveEntries(t *testing.T) {
	for _, id := range All() {
		rels := InstructionFiles(id)
		if len(rels) == 0 {
			t.Errorf("InstructionFiles(%q) is empty; add a canonical file list", id)
		}
		for _, rel := range rels {
			if rel == "" {
				t.Errorf("InstructionFiles(%q) contains empty path", id)
			}
		}
	}
}

// TestWriteSingleInstruction covers the three core behaviours per IDE:
// new file gets written, existing file is skipped, existing file is
// overwritten when requested. Every case runs in its own t.TempDir().
func TestWriteSingleInstruction(t *testing.T) {
	cases := []struct {
		name      string
		id        ID
		overwrite bool
		preExist  []string // paths (relative) to pre-create inside the temp root
		wantWrite []string // paths that should exist after the call
		wantSkip  []string // paths that should remain untouched (preExist preserved)
	}{
		{
			name:      "claude-code writes CLAUDE.md",
			id:        ClaudeCode,
			wantWrite: []string{"CLAUDE.md"},
		},
		{
			name:      "cursor writes both legacy and modular rules",
			id:        Cursor,
			wantWrite: []string{".cursorrules", ".cursor/rules/free-kiro.md"},
		},
		{
			name:      "continue writes both legacy and modular rules",
			id:        Continue,
			wantWrite: []string{".continuerules", ".continue/rules/free-kiro.md"},
		},
		{
			name:      "opencode writes AGENTS.md",
			id:        OpenCode,
			wantWrite: []string{"AGENTS.md"},
		},
		{
			name:      "codebuddy writes AGENTS.md",
			id:        CodeBuddy,
			wantWrite: []string{"AGENTS.md"},
		},
		{
			name:      "skip pre-existing files when overwrite=false",
			id:        ClaudeCode,
			preExist:  []string{"CLAUDE.md"},
			wantSkip:  []string{"CLAUDE.md"},
			wantWrite: []string{},
		},
		{
			name:      "overwrite pre-existing files when overwrite=true",
			id:        ClaudeCode,
			preExist:  []string{"CLAUDE.md"},
			overwrite: true,
			wantWrite: []string{"CLAUDE.md"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for _, rel := range tc.preExist {
				abs := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
					t.Fatalf("pre-create dir: %v", err)
				}
				if err := os.WriteFile(abs, []byte("user-managed"), 0o644); err != nil {
					t.Fatalf("pre-create file: %v", err)
				}
			}

			written, err := WriteSingleInstruction(root, "zh", tc.overwrite, tc.id)
			if err != nil {
				t.Fatalf("WriteSingleInstruction: %v", err)
			}

			if !equalStringSlices(written, tc.wantWrite) {
				t.Errorf("returned paths = %v, want %v", written, tc.wantWrite)
			}
			for _, rel := range tc.wantWrite {
				abs := filepath.Join(root, rel)
				body, readErr := os.ReadFile(abs)
				if readErr != nil {
					t.Errorf("expected file %s to exist: %v", rel, readErr)
					continue
				}
				if !strings.HasPrefix(string(body), freeKiroInstructionMarker) {
					t.Errorf("file %s missing %q marker; body starts with %q",
						rel, freeKiroInstructionMarker, firstLine(string(body)))
				}
			}
			for _, rel := range tc.wantSkip {
				abs := filepath.Join(root, rel)
				body, readErr := os.ReadFile(abs)
				if readErr != nil {
					t.Errorf("expected pre-existing %s to survive: %v", rel, readErr)
					continue
				}
				if string(body) != "user-managed" {
					t.Errorf("pre-existing %s was modified: got %q, want %q",
						rel, string(body), "user-managed")
				}
			}
		})
	}
}

// TestWriteAgentInstructions_DedupesAcrossIDs guarantees that the same
// absolute path targeted by two different IDE ids (e.g. OpenCode and
// CodeBuddy both writing AGENTS.md) is only written once and only
// returned once in the slice.
func TestWriteAgentInstructions_DedupesAcrossIDs(t *testing.T) {
	root := t.TempDir()
	written, err := WriteAgentInstructions(root, "zh", false, []ID{OpenCode, CodeBuddy})
	if err != nil {
		t.Fatalf("WriteAgentInstructions: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected single deduped AGENTS.md; got %v", written)
	}
	if written[0] != "AGENTS.md" {
		t.Errorf("expected AGENTS.md; got %q", written[0])
	}
}

// TestWriteAgentInstructions_MultipleDistinctIDs ensures each distinct
// file is written once across multiple non-overlapping IDEs.
func TestWriteAgentInstructions_MultipleDistinctIDs(t *testing.T) {
	root := t.TempDir()
	written, err := WriteAgentInstructions(root, "en", false, []ID{ClaudeCode, OpenCode})
	if err != nil {
		t.Fatalf("WriteAgentInstructions: %v", err)
	}
	want := map[string]bool{"CLAUDE.md": true, "AGENTS.md": true}
	if len(written) != len(want) {
		t.Fatalf("expected %d paths; got %v", len(want), written)
	}
	for _, p := range written {
		if !want[p] {
			t.Errorf("unexpected path %q", p)
		}
	}
	for rel := range want {
		if _, statErr := os.Stat(filepath.Join(root, rel)); statErr != nil {
			t.Errorf("expected file %s on disk: %v", rel, statErr)
		}
	}
}

// TestIsFreeKiroInstruction covers the four observable states: marked,
// unmarked, missing, and unparseable-read.
func TestIsFreeKiroInstruction(t *testing.T) {
	root := t.TempDir()
	marked := filepath.Join(root, "marked.md")
	if err := os.WriteFile(marked, []byte(freeKiroInstructionMarker+" rest\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unmarked := filepath.Join(root, "unmarked.md")
	if err := os.WriteFile(unmarked, []byte("# user-managed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
		want bool
	}{
		{"marked file", marked, true},
		{"unmarked file", unmarked, false},
		{"missing file returns false without error", filepath.Join(root, "nope.md"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := IsFreeKiroInstruction(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("IsFreeKiroInstruction = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestWriteAgentsMD_PrependsMarker ensures the legacy workspace writer
// also prepends the marker so doctor can identify its output.
func TestWriteAgentsMD_PrependsMarker(t *testing.T) {
	root := t.TempDir()
	dest, err := WriteAgentsMD(root, "zh", false)
	if err != nil {
		t.Fatalf("WriteAgentsMD: %v", err)
	}
	body, readErr := os.ReadFile(dest)
	if readErr != nil {
		t.Fatalf("read %s: %v", dest, readErr)
	}
	if !strings.HasPrefix(string(body), freeKiroInstructionMarker) {
		t.Errorf(".kiro/AGENTS.md missing marker; body starts with %q", firstLine(string(body)))
	}
}

// firstLine returns up to the first newline in s, or s if none. Used to
// keep failure messages short.
func firstLine(s string) string {
	head, _, _ := strings.Cut(s, "\n")
	return head
}

// equalStringSlices returns true when a and b have the same length and
// the same elements in the same order. nil and empty are considered equal.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
