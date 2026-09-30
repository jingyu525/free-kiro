// Package cli — status.go: aggregate workspace state into a single
// human-readable or JSON snapshot.
//
// Reuses visualize.BuildReport as the single source of truth so the
// JSON shape stays in lock-step with `serve /api/summary`.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/spec"
	"github.com/jingyu525/free-kiro/internal/visualize"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// statusCmdFactory returns the cobra command for `free-kiro status`.
//
// `status` is read-only — it never touches .kiro/specs/, .kiro/steering/,
// .kiro/hooks/, or .meta.json. It only reads .kiro/.current (via
// workspace.ReadCurrent, inside BuildReport) to mark the active spec.
func statusCmdFactory() *cobra.Command {
	var asJSON bool
	var asHuman bool
	c := &cobra.Command{
		Use:   "status",
		Short: "聚合展示 workspace 状态（所有 spec 阶段 / 漂移 / tasks 进度）",
		Long: `status 打印当前 .kiro/ workspace 整体状态：

  - 每个 spec 的 phase / approved / drift / tasks 进度
  - 标记当前活跃 spec（来自 .kiro/.current）
  - 默认人类可读表格，加 --json 输出与 serve /api/summary 同源结构`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ws, err := workspace.Find("").Require()
			if err != nil {
				return exitWithError(err)
			}
			rep, err := visualize.BuildReport(ws, spec.New(ws))
			if err != nil {
				return exitWithError(ferrors.Wrap("status", err, "build report failed"))
			}
			out := cmd.OutOrStdout()
			switch {
			case asJSON:
				return renderStatusJSON(out, rep)
			case asHuman, true:
				// Default branch — `asHuman` is also true when neither
				// flag was passed; the explicit flag is here so callers
				// like `free-kiro watch --preset reactive` can document
				// intent without breaking.
				return renderStatusHuman(out, rep)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "输出 JSON（与 serve /api/summary 同源）")
	c.Flags().BoolVar(&asHuman, "human", false, "输出人类可读表格（默认）")
	return c
}

// renderStatusJSON marshals the ProjectReport and writes to w.
// ProjectReport already carries json tags, so no struct duplication.
func renderStatusJSON(w io.Writer, r *visualize.ProjectReport) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ferrors.Wrap("status", err, "marshal report failed")
	}
	writeOutln(w, string(data))
	return nil
}

// renderStatusHuman writes a fixed-column plain-text table.
// No ANSI escapes — output is greppable / pipe-safe.
func renderStatusHuman(w io.Writer, r *visualize.ProjectReport) error {
	if len(r.Specs) == 0 {
		writeOutln(w, "No specs found")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	writeOutln(tw, "*\tNAME\tPHASE\tAPPROVED\tDRIFT\tTASKS")
	for _, s := range r.Specs {
		marker := ""
		if s.Active {
			marker = "*"
		}
		drift := "none"
		if len(s.Drift) > 0 {
			drift = fmt.Sprintf("%d keys", len(s.Drift))
		}
		tasks := fmt.Sprintf("%d/%d", s.Tasks.Done, s.Tasks.Total)
		if s.Tasks.Total == 0 {
			tasks = "—"
		}
		writeOut(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			marker, s.Meta.Name, s.Meta.Phase,
			yesNoHuman(s.Meta.Approved), drift, tasks)
	}
	if err := tw.Flush(); err != nil {
		return ferrors.Wrap("status", err, "flush table failed")
	}
	// If .current names a spec absent from the list, surface it as a
	// footer so the user knows their active pointer is dangling.
	if r.Active != "" && !activeExists(r.Specs, r.Active) {
		writeOut(w, "\n(current=%s not found)\n", r.Active)
	}
	return nil
}

// yesNoHuman renders a bool as the literal "yes"/"no" token — no
// Unicode checkmarks (those would break greps in CJK locales).
func yesNoHuman(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func activeExists(specs []*visualize.SpecReport, name string) bool {
	for _, s := range specs {
		if s.Meta.Name == name {
			return true
		}
	}
	return false
}
