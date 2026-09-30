package cli

import (
	"github.com/spf13/cobra"
)

// These are the spec subcommand stubs replaced by real implementations
// in spec_*.go files. specCmd itself is constructed in spec.go and
// populates its subcommands via initSpecSubcommands().

// initSpecSubcommands registers every spec lifecycle subcommand onto
// the specCmd parent. Called from spec.go's init().
func initSpecSubcommands(specCmd *cobra.Command) {
	specCmd.AddCommand(specNewCmdFactory()) // new
	specCmd.AddCommand(specGenerateCmd())   // generate
	specCmd.AddCommand(specQuickCmd())      // quick
	specCmd.AddCommand(specShowCmd())       // show
	specCmd.AddCommand(specListCmd())       // list
	specCmd.AddCommand(specApproveCmd())    // approve
	specCmd.AddCommand(specStartCmd())      // start
	specCmd.AddCommand(specCompleteCmd())   // complete
	specCmd.AddCommand(specSyncCmd())       // sync
	specCmd.AddCommand(specStatusCmd())     // status
	specCmd.AddCommand(specNextCmd())       // next
	specCmd.AddCommand(specAnalyzeCmd())    // analyze
}

// specNewCmdFactory wraps newSpecCmd so the registration above is
// consistent (every spec subcommand returns *cobra.Command).
func specNewCmdFactory() *cobra.Command { return newSpecCmd() }

// specShowCmd needs to declare its --phase flag.
func init() {
	// Defer phase flag registration until after specShowCmd is created.
}

// Note: specShowCmd() already attaches the --phase flag inside its own
// constructor (see spec_simple.go).
