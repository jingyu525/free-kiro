package spec

import (
	"fmt"
	"os"
)

// warnf prints a non-fatal warning to stderr. Best-effort: errors
// from Fprintf are discarded (mirroring the policy of cli/writeOutln
// — partial writes can't be meaningfully recovered during a CLI run).
//
// Kept package-local so spec doesn't need to depend on the cli
// package's print helpers, and so all spec-engine warnings share a
// single format ("warning: <message>").
func warnf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "warning: "+format+"\n", a...)
}
