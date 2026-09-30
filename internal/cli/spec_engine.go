package cli

import (
	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
)

// engineHelper builds a SpecEngine on demand. Kept here (not next to
// loadEngine) so the import direction is cli → spec (no cycle).
func (e *engineAndWS) engine() *spec.Engine {
	return spec.New(e.ws)
}

// engineForSpec is a tiny convenience that loads the workspace + engine
// in one step for subcommands that need both.
func engineForSpec() (*spec.Engine, error) {
	holder, err := loadEngine()
	if err != nil {
		return nil, err
	}
	return holder.engine(), nil
}

// printSpecMeta renders the post-action state of a spec meta to stdout.
// Used by spec new / approve / start / complete for consistent output.
func printSpecMeta(cmd *cobra.Command, m *models.SpecMeta, tag string) {
	writeOut(cmd.OutOrStdout(),
		"%s spec %q (phase: %s, workflow: %s, type: %s)\n",
		tag, m.Name, m.Phase, m.Workflow, m.SpecType)
}
