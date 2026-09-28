package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/models"
)

// specShowCmd prints the contents of one phase document (or every
// planning doc when --phase all).
func specShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "打印某个 phase 的文档",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			phaseStr, _ := cmd.Flags().GetString("phase")
			if phaseStr == "all" {
				meta, err := loadMetaViaEngine(eng, args[0])
				if err != nil {
					return exitWithError(err)
				}
				for _, doc := range models.PlanningOrder(meta.Workflow, meta.SpecType) {
					body, _ := eng.Show(args[0], phaseFromDoc(models.Phase(doc)))
					fmt.Fprintf(cmd.OutOrStdout(), "=== %s ===\n%s\n\n", doc, body)
				}
				return nil
			}
			body, err := eng.Show(args[0], models.Phase(phaseStr))
			if err != nil {
				return exitWithError(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), body)
			return nil
		},
	}
}

// specListCmd enumerates every spec in the workspace as a fixed-width table.
func specListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出全部 specs",
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			rows, err := eng.StatusForList()
			if err != nil {
				return exitWithError(err)
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no specs)")
				return nil
			}
			// Stable sort by name.
			sort.Slice(rows, func(i, j int) bool {
				return rows[i]["name"].(string) < rows[j]["name"].(string)
			})
			fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-14s %-9s %s\n", "NAME", "PHASE", "APPROVED", "ACTIVE")
			for _, r := range rows {
				marker := ""
				if active, _ := r["active"].(bool); active {
					marker = "*"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-14s %-9s %s\n",
					r["name"], r["phase"], yesNo(r["approved"].(bool)), marker)
			}
			// Hint about how to switch.
			fmt.Fprintln(cmd.OutOrStdout(),
				"\n(* = active spec, used by IDE SessionStart hook. "+
					"To switch: edit .kiro/.current.)")
			return nil
		},
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// phaseFromDoc converts a doc filename back to its phase (helper for spec show).
func phaseFromDoc(p models.Phase) models.Phase {
	return p
}