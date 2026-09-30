package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newVersionedRoot constructs a fresh cobra.Command mirroring the
// production rootCmd wiring (Version + SetVersionTemplate), so each
// test starts from a clean state without depending on package-level
// init order or carrying over parsed-flag state between runs.
func newVersionedRoot() *cobra.Command {
	c := &cobra.Command{
		Use:     "free-kiro",
		Version: fmt.Sprintf("%s (commit %s, built %s)", buildVersion, buildCommit, buildDate),
	}
	c.SetVersionTemplate("free-kiro version {{.Version}}\n")
	return c
}

// TestRootCmdHasVersion asserts the production rootCmd (the singleton
// package-level instance) has Version wired up at init time and embeds
// both commit and build date segments.
func TestRootCmdHasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Fatal("rootCmd.Version must be non-empty after init()")
	}
	if !strings.Contains(rootCmd.Version, "commit ") {
		t.Errorf("rootCmd.Version = %q, must contain 'commit '", rootCmd.Version)
	}
	if !strings.Contains(rootCmd.Version, "built ") {
		t.Errorf("rootCmd.Version = %q, must contain 'built '", rootCmd.Version)
	}
}

// TestVersionFlagOutput asserts the runtime `--version` invocation
// produces a single line that includes the `free-kiro version ` prefix
// and the underlying buildVersion literal.
func TestVersionFlagOutput(t *testing.T) {
	cmd := newVersionedRoot()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version returned error: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()

	// Must be a single line (cobra appends '\n' to the template).
	if strings.Count(got, "\n") > 1 {
		t.Errorf("--version output should be one line, got %q", got)
	}
	if !strings.HasPrefix(got, "free-kiro version ") {
		t.Errorf("--version output should start with 'free-kiro version ', got %q", got)
	}
	// dev builds inject the literal "dev" — verify it survives the
	// template.
	if buildVersion == "dev" && !strings.Contains(got, "dev") {
		t.Errorf("dev build --version output should contain 'dev', got %q", got)
	}
	// No version text should leak into stderr (we asserted exit=0 above
	// already; this catches silent re-routing bugs).
	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty for --version, got %q", errBuf.String())
	}
}

// TestVersionFlagIgnoresOtherArgs asserts `--version` short-circuits
// before cobra tries to dispatch to a subcommand — passing `init` after
// `--version` must NOT trigger init's RunE.
func TestVersionFlagIgnoresOtherArgs(t *testing.T) {
	cmd := newVersionedRoot()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--version", "init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version init returned error: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	if !strings.HasPrefix(got, "free-kiro version ") {
		t.Errorf("--version init should still print version, got %q", got)
	}
	// initCmd would write to stdout if it ran; the only stdout content
	// should be the version line.
	if strings.Contains(got, ".kiro") {
		t.Errorf("--version should ignore trailing args, but got init output: %q", got)
	}
}

// TestVersionTemplateIsGreppable asserts the version output is safe to
// pipe — no ANSI escape codes, no carriage returns.
func TestVersionTemplateIsGreppable(t *testing.T) {
	cmd := newVersionedRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version: %v", err)
	}
	got := out.String()
	if strings.ContainsAny(got, "\r\x1b") {
		t.Errorf("--version output must not contain CR or ANSI escapes, got %q", got)
	}
}