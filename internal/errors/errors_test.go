package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestNewUsage(t *testing.T) {
	e := NewUsage("cli.spec.new", "spec name is required")
	if e == nil {
		t.Fatal("NewUsage returned nil")
	}
	if e.Op != "cli.spec.new" {
		t.Errorf("Op = %q, want cli.spec.new", e.Op)
	}
	if e.Msg != "spec name is required" {
		t.Errorf("Msg = %q, want 'spec name is required'", e.Msg)
	}
	if e.Err != nil {
		t.Errorf("Err = %v, want nil", e.Err)
	}
	// ExitCode path.
	if got := ExitCode(e); got != 3 {
		t.Errorf("ExitCode(UsageError) = %d, want 3", got)
	}
	// errors.As target type works.
	var u *UsageError
	if !stderrors.As(e, &u) {
		t.Errorf("errors.As(*UsageError) failed")
	}
	// Error() string contains Op and Msg.
	got := e.Error()
	if !contains(got, "cli.spec.new") || !contains(got, "spec name is required") {
		t.Errorf("Error() = %q, missing op/msg", got)
	}
}

func TestWrapUsage(t *testing.T) {
	cause := stderrors.New("stdin: file already closed")
	e := WrapUsage("cli.spec.new.readPrompt", cause, "读取 stdin 失败")
	if e == nil {
		t.Fatal("WrapUsage returned nil")
	}
	if e.Err != cause {
		t.Errorf("Err not preserved: got %v, want %v", e.Err, cause)
	}
	if !stderrors.Is(e, cause) {
		t.Errorf("errors.Is should match wrapped cause")
	}
	if got := ExitCode(e); got != 3 {
		t.Errorf("ExitCode(WrapUsage) = %d, want 3", got)
	}
	if got := e.Error(); !contains(got, "读取 stdin 失败") {
		t.Errorf("Error() = %q, missing msg", got)
	}
}

func TestExitCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"UsageError", NewUsage("op", "msg"), 3},
		{"LintGateError", &LintGateError{KiroError: New("op", "msg")}, 2},
		{"TaskGraphError", NewTaskGraphError("op", "msg"), 1},
		{"KiroError", New("op", "msg"), 2},
		{"unknown", stderrors.New("plain"), 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ExitCode(c.err); got != c.want {
				t.Errorf("ExitCode(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

func TestWrapAndUnwrap(t *testing.T) {
	cause := stderrors.New("disk full")
	e := Wrap("init.write", cause, "写入 AGENTS.md 失败")
	if !stderrors.Is(e, cause) {
		t.Errorf("errors.Is should find cause")
	}
	if got := fmt.Sprint(e); !contains(got, "写入 AGENTS.md 失败") {
		t.Errorf("Error() = %q, missing msg", got)
	}
}

// contains is a tiny strings.Contains shim to avoid importing strings
// only for the helper.
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
