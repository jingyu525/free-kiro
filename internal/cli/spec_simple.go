package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/models"
	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/taskgraph"
	"github.com/jingyu525/free-kiro/internal/visualize"
)

// specShowCmd prints the contents of one phase document (or every
// planning doc when --phase all).
func specShowCmd() *cobra.Command {
	var treeFlag bool
	c := &cobra.Command{
		Use:   "show <name>",
		Short: "打印某个 phase 的文档（--tree 看 ASCII 树形）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			phaseStr, _ := cmd.Flags().GetString("phase")
			if treeFlag {
				return renderSpecTree(cmd.OutOrStdout(), eng, args[0])
			}
			if phaseStr == "all" {
				var meta *models.SpecMeta
				meta, err = loadMetaViaEngine(eng, args[0])
				if err != nil {
					return exitWithError(err)
				}
				for _, doc := range models.PlanningOrder(meta.Workflow, meta.SpecType) {
					body, _ := eng.Show(args[0], phaseFromDoc(models.Phase(doc)))
					writeOut(cmd.OutOrStdout(), "=== %s ===\n%s\n\n", doc, body)
				}
				return nil
			}
			body, err := eng.Show(args[0], models.Phase(phaseStr))
			if err != nil {
				return exitWithError(err)
			}
			writeOutln(cmd.OutOrStdout(), body)
			return nil
		},
	}
	c.Flags().String("phase", "all", "phase: requirements | design | tasks | all")
	c.Flags().BoolVar(&treeFlag, "tree", false, "render as ASCII tree (phase + docs + wave schedule)")
	return c
}

// specListCmd enumerates every spec in the workspace as a fixed-width table.
func specListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出全部 specs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			rows, err := eng.StatusForList()
			if err != nil {
				return exitWithError(err)
			}
			if len(rows) == 0 {
				writeOutln(cmd.OutOrStdout(), "(no specs)")
				return nil
			}
			// Stable sort by name.
			sort.Slice(rows, func(i, j int) bool {
				return rows[i]["name"].(string) < rows[j]["name"].(string)
			})
			writeOut(cmd.OutOrStdout(), "%-24s %-14s %-9s %s\n", "NAME", "PHASE", "APPROVED", "ACTIVE")
			for _, r := range rows {
				marker := ""
				if active, _ := r["active"].(bool); active {
					marker = "*"
				}
				writeOut(cmd.OutOrStdout(), "%-24s %-14s %-9s %s\n",
					r["name"], r["phase"], yesNo(r["approved"].(bool)), marker)
			}
			// Hint about how to switch.
			writeOutln(cmd.OutOrStdout(),
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

// renderSpecTree prints an ASCII tree summary of the spec: phase /
// docs / wave schedule. Mirrors what `free-kiro report` puts in
// REPORT.md but inline for terminal use.
func renderSpecTree(w io.Writer, eng *spec.Engine, name string) error {
	// We need tasks.md content + status for the tree. Read tasks.md
	// directly (cheap; ~12-line file).
	tasksData, err := os.ReadFile(filepath.Join(eng.WS().SpecDir(name), "tasks.md"))
	if err != nil {
		// No tasks yet — render a leaner tree.
		tasksData = nil
	}
	var tasks []models.Task
	if tasksData != nil {
		tasks = taskgraph.ParseTasks(string(tasksData))
	}

	st, _ := eng.Status(name)
	root := buildSpecTreeNode(name, st, tasks)
	writeOutln(w, visualize.RenderTree(root))
	return nil
}

// buildSpecTreeNode assembles the tree data structure from a spec's
// status + parsed tasks. Pure function — testable.
func buildSpecTreeNode(name string, st map[string]any, tasks []models.Task) *visualize.TreeNode {
	phase, _ := st["phase"].(string)
	workflow, _ := st["workflow"].(string)
	specType, _ := st["spec_type"].(string)
	approved, _ := st["approved"].(bool)

	root := &visualize.TreeNode{
		Text: name,
		Meta: map[string]string{
			"phase":    phase,
			"workflow": workflow,
			"type":     specType,
			"approved": yesNo(approved),
		},
	}

	// Drift (actionable alert at top).
	if drift, ok := st["drift"].([]any); ok && len(drift) > 0 {
		driftNode := &visualize.TreeNode{Text: "DRIFT (baseline → current):"}
		for _, raw := range drift {
			dm, _ := raw.(map[string]any)
			key, _ := dm["key"].(string)
			base := asInt(dm["baseline"])
			cur := asInt(dm["current"])
			delta := asInt(dm["delta"])
			sign := " "
			if delta > 0 {
				sign = "+"
			} else if delta < 0 {
				sign = "-"
			}
			driftNode.Children = append(driftNode.Children, &visualize.TreeNode{
				Text: fmt.Sprintf("%s: %d → %d (%s%d)", key, base, cur, sign, absDeltaLocal(delta)),
			})
		}
		root.Children = append(root.Children, driftNode)
	}

	// Tasks line.
	if t, ok := st["tasks"].(map[string]any); ok {
		done := asInt(t["done"])
		total := asInt(t["total"])
		waves := asInt(t["waves"])
		root.Children = append(root.Children, &visualize.TreeNode{
			Text: fmt.Sprintf("tasks: %d/%d done, %d wave(s)", done, total, waves),
		})
	}

	// Wave schedule.
	if len(tasks) > 0 {
		waveNode := &visualize.TreeNode{Text: "wave schedule:"}
		waves := taskgraph.ExecutionWaves(tasks)
		for i, w := range waves {
			wn := &visualize.TreeNode{
				Text: fmt.Sprintf("Wave %d", i+1),
			}
			for _, t := range w {
				mark := " "
				if t.Done {
					mark = "x"
				}
				deps := "-"
				if len(t.Deps) > 0 {
					deps = ""
					for j, d := range t.Deps {
						if j > 0 {
							deps += ","
						}
						deps += fmt.Sprintf("#%d", d)
					}
				}
				wn.Children = append(wn.Children, &visualize.TreeNode{
					Text: fmt.Sprintf("[%s] #%d %s  (deps: %s)", mark, t.ID, t.Title, deps),
				})
			}
			waveNode.Children = append(waveNode.Children, wn)
		}
		root.Children = append(root.Children, waveNode)
	}

	return root
}

func absDeltaLocal(d int) int {
	if d < 0 {
		return -d
	}
	return d
}

// asInt converts a value (int / int64 / float64) to int. Returns 0
// on unknown types. Used by the tree renderer to coerce status JSON
// values without panicking.
func asInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
