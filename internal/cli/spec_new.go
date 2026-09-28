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
		prompt    string
		workflow  string
		specType  string
		quick     bool
		fromIssue string
	)
	c := &cobra.Command{
		Use:   "new [<name>]",
		Short: "新建 spec（含 workflow / type / quick / from-issue 变体）",
		Args:  cobra.MaximumNArgs(1),
		Long: `创建一个新的 spec：

  --prompt "<一句话需求>"   必填（或通过 stdin 传入）
  --from-issue <url>        从 GitHub issue 自动拉取标题 + body 作为 prompt
                            URL 形式: https://github.com/<o>/<r>/issues/<n>
                            或简写: <o>/<r>#<n>
                            配合 --type bugfix 可快速生成修复 spec
  --workflow <wf>          requirements-first（默认）| design-first
  --type <type>            feature（默认）| bugfix
  --quick                  免审批变体（Quick Spec）

当 --from-issue 被使用时，<name> 可省略 —— 会从 issue 标题自动生成 kebab-case slug。

下一步：kiro spec generate <name> --phase all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}

			// Resolve prompt + name from various input modes.
			p, resolvedName, err := resolvePromptAndName(prompt, fromIssue, args)
			if err != nil {
				return exitWithError(err)
			}

			meta, err := eng.NewSpec(resolvedName, p, workflow, specType, quick)
			if err != nil {
				return exitWithError(err)
			}
			tag := "created"
			if fromIssue != "" {
				tag = "created (from issue)"
			}
			if quick {
				tag = "created (quick)"
			}
			printSpecMeta(cmd, meta, tag)
			if !quick {
				fmt.Fprintf(cmd.OutOrStdout(),
					"next: free-kiro spec generate %s --phase all\n", meta.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(),
					"next: free-kiro spec start %s\n", meta.Name)
			}
			return nil
		},
	}
	c.Flags().StringVar(&prompt, "prompt", "", "spec prompt text (or pipe via stdin)")
	c.Flags().StringVar(&fromIssue, "from-issue", "", "GitHub issue URL to derive prompt from (requires `gh` CLI)")
	c.Flags().StringVar(&workflow, "workflow", "requirements-first", "planning order: requirements-first or design-first")
	c.Flags().StringVar(&specType, "type", "feature", "spec type: feature or bugfix")
	c.Flags().BoolVar(&quick, "quick", false, "quick spec: waive the formal approve gate")
	return c
}

// resolvePromptAndName centralizes the input resolution logic for
// `spec new`. Three modes are supported:
//
//  1. --from-issue: fetch the issue title + body via `gh`. The spec
//     name defaults to a kebab-case slug from the title (or stays as
//     the explicit positional arg when provided).
//  2. --prompt + positional name: classic path.
//  3. stdin pipe (read by readPrompt) + positional name: classic path.
//
// Returns the resolved prompt text and the final spec name.
func resolvePromptAndName(promptFlag, fromIssue string, args []string) (string, string, error) {
	if fromIssue != "" {
		ref, err := ParseGitHubIssueURL(fromIssue)
		if err != nil {
			return "", "", err
		}
		body, err := FetchIssueTitleAndBody(ref)
		if err != nil {
			return "", "", err
		}
		// Derive name from first line of body when caller didn't pass one.
		var name string
		if len(args) >= 1 {
			name = args[0]
		} else {
			name = slugFromText(firstLine(body))
		}
		return body, name, nil
	}

	p, err := readPrompt(promptFlag)
	if err != nil {
		return "", "", err
	}
	var name string
	if len(args) >= 1 {
		name = args[0]
	} else {
		return "", "", fmt.Errorf("spec name is required when --from-issue is not used")
	}
	return p, name, nil
}

// firstLine returns the first non-empty line of text.
func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		t := strings.TrimSpace(l)
		if t != "" {
			return t
		}
	}
	return ""
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
		return "", fmt.Errorf("a prompt is required (use --prompt, --from-issue, or pipe via stdin)")
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
		return "", fmt.Errorf("a prompt is required (use --prompt, --from-issue, or pipe via stdin)")
	}
	return out, nil
}