// free-kiro is a single-binary clone of the Kiro Spec workflow engine.
//
// All commands route through internal/cli.Execute, which maps engine errors
// to exit codes per the contract in internal/errors:
//
//	0 = success, 1 = lint gate failure, 2 = engine error, 3 = usage error.
package main

import (
	"fmt"
	"os"

	"github.com/jingyu525/free-kiro/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		// cli.Execute already prints the error and returns the exit code.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}