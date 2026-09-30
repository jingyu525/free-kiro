package cli

import (
	"fmt"
	"io"
)

// writeOut writes a formatted message to w.
//
// Errors from the underlying io.Writer are intentionally discarded.
// Cobra command output targets stdout/stderr that the user is already
// watching; a partial write cannot be meaningfully recovered, and the
// process is about to exit. Returning the error would force every call
// site to either propagate it (impossible — Cobra RunE returns after
// printing) or wrap it in another swallowed error, neither of which
// improves UX.
func writeOut(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

// writeOutln writes a non-formatted line to w. Same best-effort policy
// as writeOut.
func writeOutln(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, a...)
}
