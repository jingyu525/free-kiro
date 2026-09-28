package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newSpecCmd() *cobra.Command {
	var (
		prompt   string
		workflow string
		specType string
		quick    bool
	)
	c := &cobra.Command{
		Use:   "new <name>",
		Short: "新建 spec（含 workflow / type / quick 变体）",
		Args:  cobra.ExactArgs(1),
		Long: `创建一个新的 spec：

  --prompt "<一句话需求>"   必填（或通过 stdin 传入）
  --workflow <wf>          requirements-first（默认）| design-first
  --type <type>            feature（默认）| bugfix
  --quick                  免审批变体（Quick Spec）

下一步：kiro spec generate <name> --phase all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}
			p, err := readPrompt(prompt)
			if err != nil {
				return exitWithError(err)
			}
			meta, err := eng.NewSpec(args[0], p, workflow, specType, quick)
			if err != nil {
				return exitWithError(err)
			}
			tag := "created"
			if quick {
				tag = "created (quick)"
			}
			printSpecMeta(cmd, meta, tag)
			if !quick {
				fmt.Fprintf(cmd.OutOrStdout(),
					"next: free-kiro spec generate %s --phase all\n", meta.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(),
					"next: free-kiro spec quick %s --prompt %q  (or review docs then `spec start`)\n",
					meta.Name, p)
			}
			return nil
		},
	}
	c.Flags().StringVar(&prompt, "prompt", "", "spec prompt text (or pipe via stdin)")
	c.Flags().StringVar(&workflow, "workflow", "requirements-first", "planning order: requirements-first or design-first")
	c.Flags().StringVar(&specType, "type", "feature", "spec type: feature or bugfix")
	c.Flags().BoolVar(&quick, "quick", false, "quick spec: waive the formal approve gate")
	return c
}

// readPrompt returns the explicit --prompt value if set, else reads from
// stdin (when piped). Raises a UsageError if neither is available.
func readPrompt(prompt string) (string, error) {
	if prompt != "" {
		return prompt, nil
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("stdin: %w", err)
	}
	if fi.Mode()&os.ModeCharDevice != 0 {
		// Stdin is a TTY, not a pipe — no prompt available.
		return "", fmt.Errorf("a prompt is required (use --prompt or pipe via stdin)")
	}
	rdr := bufio.NewReader(os.Stdin)
	var sb strings.Builder
	for {
		line, err := rdr.ReadString('\n')
		if line != "" {
			sb.WriteString(line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("a prompt is required (use --prompt or pipe via stdin)")
	}
	return out, nil
}