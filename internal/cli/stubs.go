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
	specCmd.AddCommand(specNewCmdFactory())     // new
	specCmd.AddCommand(specGenerateCmd())       // generate
	specCmd.AddCommand(specQuickCmd())          // quick
	specCmd.AddCommand(specShowCmd())           // show
	specCmd.AddCommand(specListCmd())           // list
	specCmd.AddCommand(specApproveCmd())        // approve
	specCmd.AddCommand(specStartCmd())          // start
	specCmd.AddCommand(specCompleteCmd())       // complete
	specCmd.AddCommand(specSyncCmd())           // sync
	specCmd.AddCommand(specStatusCmd())         // status
	specCmd.AddCommand(specNextCmd())           // next
	specCmd.AddCommand(specAnalyzeCmd())        // analyze
}

// specNewCmdFactory wraps newSpecCmd so the registration above is
// consistent (every spec subcommand returns *cobra.Command).
func specNewCmdFactory() *cobra.Command { return newSpecCmd() }

// specShowCmd needs to declare its --phase flag.
func init() {
	// Defer phase flag registration until after specShowCmd is created.
}

// attachPhaseFlags attaches the --phase flag (choices: requirements,
// design, tasks, all) to commands that need it. Called explicitly from
// each command's constructor after RunE is set.
func attachPhaseFlags(c *cobra.Command) {
	c.Flags().String("phase", "all", "phase: requirements | design | tasks | all")
}

// Note: specShowCmd() already attaches the --phase flag inside its own
// constructor (see spec_simple.go).
//
// Below: stub commands for steering / task / hook — real implementations
// land in Waves 3 and 4. These keep the root command's --help complete
// in the meantime. Root.go calls stubCmd constructors directly inside
// init() to avoid forward-declaration cycles.
func stubCmd(name, short string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("(子命令 %q 在 Wave 3/4 实现；当前为占位符)\n", name)
			return nil
		},
	}
}

// stubFor is the helper used by root.go's init() to wire the not-yet-
// implemented subcommands (steering / task / hook).
func stubFor(name, short string) *cobra.Command { return stubCmd(name, short) }