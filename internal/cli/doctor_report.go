package cli

import (
	"fmt"
	"strings"
)

// doctorIssue is one row of the doctor report. Severity is one of
// "ok" / "info" / "warn" / "fatal" — drives both the icon prefix and
// the report-level counters in doctorReport.
type doctorIssue struct {
	Severity string // "ok" | "info" | "warn" | "fatal"
	Title    string
	Detail   string
	Fix      string // optional one-line fix hint
}

func (i doctorIssue) render() string {
	prefix := "✓"
	switch i.Severity {
	case "info":
		prefix = "ℹ"
	case "warn":
		prefix = "⚠"
	case "fatal":
		prefix = "✗"
	}
	out := fmt.Sprintf("%s %s", prefix, i.Title)
	if i.Detail != "" {
		out += "\n    " + i.Detail
	}
	if i.Fix != "" {
		out += "\n    fix: " + i.Fix
	}
	return out
}

// joinStrings is a tiny local helper that mirrors strings.Join but lives
// here so doctor.go can avoid pulling in the strings import when only
// this single call site exists in the file.
func joinStrings(s []string, sep string) string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0]
	}
	var b strings.Builder
	b.WriteString(s[0])
	for _, x := range s[1:] {
		b.WriteString(sep)
		b.WriteString(x)
	}
	return b.String()
}

// keep getwd referenced so the doctor_checks.go shim can be replaced
// with a direct os.Getwd call when needed.
var _ = osGetwd
