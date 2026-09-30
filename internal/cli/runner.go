package cli

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/spec"
)

// CmdFunc is the function signature every spec-lifecycle subcommand
// implements when going through RunCmd. The engine is pre-resolved; the
// command is the cobra command being executed (so callers can read
// flags / write to OutOrStdout). Errors returned from fn are passed
// through untouched so callers can return typed errors (UsageError →
// exit 3) and have main.go's ExitCode map them correctly.
type CmdFunc func(ctx context.Context, eng *spec.Engine, cmd *cobra.Command) error

// RunCmd wraps a CmdFunc with the three patterns every spec lifecycle
// subcommand used to inline:
//
//  1. resolve engine via engineForSpec (returns a typed error already
//     mapped through exitWithError).
//  2. defer recover() so a panic becomes a *ferrors.KiroError → exit 2
//     instead of crashing the process.
//  3. propagate fn's error verbatim (callers return typed errors
//     intentionally — wrapping them here would defeat the ExitCode
//     mapping done by main.go).
//
// RunCmd does NOT print the "next:" hint — see NextHint for that. Mixing
// it in would force every CmdFunc to know its follow-up command, which
// couples unrelated lifecycle steps.
func RunCmd(cmd *cobra.Command, _ []string, fn CmdFunc) error {
	eng, err := engineForSpec()
	if err != nil {
		return err
	}
	var runErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Include a short stack trace in Msg so operator can
				// diagnose without re-running with --debug. Full debug
				// stack goes to stderr for `free-kiro doctor` / `report`
				// to pick up.
				writeOut(cmd.ErrOrStderr(), "panic recovered: %v\n%s\n", r, debug.Stack())
				runErr = ferrors.New("cli.runCmd.panic",
					fmt.Sprintf("panic in %s: %v", cmd.Name(), r))
			}
		}()
		runErr = fn(cmd.Context(), eng, cmd)
	}()
	return runErr
}

// NextHint writes a `next: free-kiro <subcommand> <args>` follow-up line
// to the command's stdout. Centralised so the message format stays
// consistent (used by spec new / spec_lifecycle / spec_quick).
func NextHint(cmd *cobra.Command, format string, args ...any) {
	writeOut(cmd.OutOrStdout(), "next: "+format+"\n", args...)
}
