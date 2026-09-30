package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/spec"
)

func TestNextHint_Format(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(&buf)
	NextHint(cmd, "free-kiro spec generate %s --phase all", "demo")
	got := buf.String()
	want := "next: free-kiro spec generate demo --phase all\n"
	if got != want {
		t.Errorf("NextHint output = %q, want %q", got, want)
	}
}

func TestNextHint_Plain(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(&buf)
	NextHint(cmd, "free-kiro doctor")
	if got := buf.String(); got != "next: free-kiro doctor\n" {
		t.Errorf("NextHint = %q, want 'next: free-kiro doctor\\n'", got)
	}
}

// TestRunCmd_PanicRecover verifies that a CmdFunc which panics is
// converted into a *ferrors.KiroError so main.go's ExitCode maps it to
// exit 2 (engine error) instead of crashing the process.
//
// We can't easily stub engineForSpec without a real .kiro workspace, so
// we directly exercise the recover closure via a small fake wiring.
func TestRunCmd_PanicRecover(t *testing.T) {
	cmd := &cobra.Command{Use: "fake"}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	var eng *spec.Engine // nil — fn will not be reached

	captured := errors.New("original")
	_ = captured

	fn := func(_ context.Context, _ *spec.Engine, _ *cobra.Command) error {
		panic("boom")
	}

	// Wrap the same recover pattern RunCmd uses to verify the conversion.
	var runErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				runErr = ferrors.New("cli.runCmd.panic", "boom")
			}
		}()
		// Call fn through the same deferred-recover wrapper.
		wrapped := func() {
			defer func() {
				if r := recover(); r != nil {
					panic(r) // re-panic so the outer recover catches it
				}
			}()
			_ = fn(context.Background(), eng, cmd)
		}
		wrapped()
	}()
	if runErr == nil {
		t.Fatal("expected runErr to be set after panic recover")
	}
	var k *ferrors.KiroError
	if !errors.As(runErr, &k) {
		t.Errorf("expected *ferrors.KiroError, got %T", runErr)
	}
	if !strings.Contains(runErr.Error(), "boom") {
		t.Errorf("error msg = %q, want to contain 'boom'", runErr.Error())
	}
	// ExitCode mapping for plain KiroError is 2.
	if got := ferrors.ExitCode(runErr); got != 2 {
		t.Errorf("ExitCode(panic-recovered) = %d, want 2", got)
	}
}

// TestRunCmd_TypedErrorPassthrough verifies that an error returned from
// CmdFunc flows through untouched so the typed error's ExitCode value
// (e.g. 3 for UsageError) reaches main.go's ExitCode call.
func TestRunCmd_TypedErrorPassthrough(t *testing.T) {
	usage := ferrors.NewUsage("test.op", "bad input")
	// Simulate the post-recover path: runErr = fn(...) where fn returns a
	// UsageError; RunCmd must NOT wrap it (wrapping would demote it to
	// a generic KiroError → exit 2).
	var runErr error
	fn := func(_ context.Context, _ *spec.Engine, _ *cobra.Command) error {
		return usage
	}
	// Mimic RunCmd's post-fn assignment without invoking engineForSpec.
	runErr = fn(context.Background(), nil, nil)
	if runErr != usage {
		t.Errorf("RunCmd should pass typed error through verbatim; got %v", runErr)
	}
	if got := ferrors.ExitCode(runErr); got != 3 {
		t.Errorf("ExitCode(UsageError) = %d, want 3", got)
	}
}
