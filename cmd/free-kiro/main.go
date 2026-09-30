// free-kiro is a single-binary clone of the Kiro Spec workflow engine.
//
// All commands route through internal/cli.Execute, which maps engine errors
// to exit codes per the contract in internal/errors:
//
//	0 = success, 1 = lint gate failure, 2 = engine error, 3 = usage error.
package main

import (
	"os"

	"github.com/jingyu525/free-kiro/internal/cli"
	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

func main() {
	if err := cli.Execute(); err != nil {
		// cli.Execute already printed the error to stderr; map to the
		// typed exit code so UsageError → 3 (caller misuse), KiroError
		// → 2 (engine), TaskGraphError → 1 (lint/cycle), nil → 0.
		os.Exit(ferrors.ExitCode(err))
	}
}