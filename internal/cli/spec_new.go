package cli

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
)

func newSpecCmd() *cobra.Command {
	var (
		prompt      string
		workflow    string
		specType    string
		quick       bool
		fromIssue   string
		fromPRD     string
		fromBrowser string
	)
	c := &cobra.Command{
		Use:   "new [<name>]",
		Short: "新建 spec（含 workflow / type / quick / from-issue / from-prd 变体）",
		Args:  cobra.MaximumNArgs(1),
		Long: `创建一个新的 spec：

  --prompt "<一句话需求>"   必填（或通过 stdin 传入）
  --from-issue <url>        从 GitHub issue 自动拉取标题 + body 作为 prompt
                            URL 形式: https://github.com/<o>/<r>/issues/<n>
                            或简写: <o>/<r>#<n>
                            配合 --type bugfix 可快速生成修复 spec
  --from-prd <url>          从任意网页（HTML / Markdown）拉取 PRD 内容
                            适合 Notion / Confluence / Google Docs / 公司 wiki
                            走 'free-kiro spec new <name> --from-prd <url>'
  --from-browser <url>     用 browser-skill 的 bsk CLI 抓 JS 渲染页
                            适合 Notion 私有页 / SPA / 需登录态的页面
                            走 'free-kiro spec new <name> --from-browser <url>'
  --workflow <wf>          requirements-first（默认）| design-first
  --type <type>            feature（默认）| bugfix
  --quick                  免审批变体（Quick Spec）

当 --from-* 被使用时，<name> 可省略 —— 会自动生成 kebab-case slug。
--from-issue / --from-prd / --from-browser 三者互斥。

下一步：kiro spec generate <name> --phase all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := engineForSpec()
			if err != nil {
				return err
			}

			// Resolve prompt + name from various input modes.
			p, resolvedName, source, err := resolvePromptAndNameEx(cmd.Context(), prompt, fromIssue, fromPRD, fromBrowser, args)
			if err != nil {
				return exitWithError(err)
			}

			meta, err := eng.NewSpec(resolvedName, p, workflow, specType, quick)
			if err != nil {
				return exitWithError(err)
			}
			tag := "created"
			switch source {
			case "from-issue":
				tag = "created (from issue)"
			case "from-prd":
				tag = "created (from PRD)"
			case "from-browser":
				tag = "created (from browser)"
			}
			if quick {
				tag = "created (quick)"
			}
			printSpecMeta(cmd, meta, tag)
			if !quick {
				writeOut(cmd.OutOrStdout(),
					"next: free-kiro spec generate %s --phase all\n", meta.Name)
			} else {
				writeOut(cmd.OutOrStdout(),
					"next: free-kiro spec start %s\n", meta.Name)
			}
			return nil
		},
	}
	c.Flags().StringVar(&prompt, "prompt", "", "spec prompt text (or pipe via stdin)")
	c.Flags().StringVar(&fromIssue, "from-issue", "", "GitHub issue URL to derive prompt from (requires `gh` CLI)")
	c.Flags().StringVar(&fromPRD, "from-prd", "", "PRD URL (HTML/Markdown) to derive prompt from")
	c.Flags().StringVar(&fromBrowser, "from-browser", "", "URL to render via browser-skill `bsk` (requires `bsk` CLI on PATH)")
	c.Flags().StringVar(&workflow, "workflow", "requirements-first", "planning order: requirements-first or design-first")
	c.Flags().StringVar(&specType, "type", "feature", "spec type: feature or bugfix")
	c.Flags().BoolVar(&quick, "quick", false, "quick spec: waive the formal approve gate")
	return c
}

// resolvePromptAndNameEx is the unified input resolver. Returns the
// resolved prompt, the final spec name, and a source label so the
// caller can print an accurate "created (from ...)" tag.
//
// Source precedence (mutually exclusive):
//
//  1. --from-issue    fetch GitHub issue title + body via `gh`
//  2. --from-prd      fetch any web page (HTML/Markdown), extract title
//     + visible text
//  3. --from-browser  fetch JS-rendered page via browser-skill's `bsk`,
//     extract title + visible text from rendered HTML
//  4. --prompt / stdin with positional name
func resolvePromptAndNameEx(ctx context.Context, promptFlag, fromIssue, fromPRD, fromBrowser string, args []string) (string, string, string, error) {
	// Mutually exclusive check across all three --from-* flags.
	n := 0
	if fromIssue != "" {
		n++
	}
	if fromPRD != "" {
		n++
	}
	if fromBrowser != "" {
		n++
	}
	if n > 1 {
		return "", "", "", ferrors.NewUsage("spec.new", "--from-issue, --from-prd, and --from-browser are mutually exclusive")
	}
	if fromIssue != "" {
		ref, err := ParseGitHubIssueURL(fromIssue)
		if err != nil {
			return "", "", "", err
		}
		body, err := FetchIssueTitleAndBody(ref)
		if err != nil {
			return "", "", "", err
		}
		name := nameFromArgsOrSlug(args, firstLine(body))
		return body, name, "from-issue", nil
	}
	if fromPRD != "" {
		title, body, err := FetchPRD(ctx, fromPRD)
		if err != nil {
			return "", "", "", err
		}
		name := nameFromArgsOrSlug(args, title)
		// Prefix the prompt with the source URL so the author has the
		// reference handy while filling in requirements.
		full := "# Source: " + fromPRD + "\n\n# " + title + "\n\n" + body
		return full, name, "from-prd", nil
	}
	if fromBrowser != "" {
		title, body, err := FetchBrowserHTML(ctx, fromBrowser)
		if err != nil {
			return "", "", "", err
		}
		name := nameFromArgsOrSlug(args, title)
		// Same prefix shape as --from-prd for downstream consistency.
		full := "# Source (browser-rendered): " + fromBrowser + "\n\n# " + title + "\n\n" + body
		return full, name, "from-browser", nil
	}
	p, err := readPrompt(promptFlag)
	if err != nil {
		return "", "", "", err
	}
	if len(args) < 1 {
		return "", "", "", ferrors.NewUsage("spec.new", "spec name is required when --from-* is not used")
	}
	return p, args[0], "", nil
}

// nameFromArgsOrSlug returns the explicit positional name when given,
// otherwise a kebab-case slug derived from `source`.
func nameFromArgsOrSlug(args []string, source string) string {
	if len(args) >= 1 {
		return args[0]
	}
	return slugFromText(firstLine(source))
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
		return "", ferrors.Wrap("spec.new.readPrompt", err, "stdin stat")
	}
	if fi.Mode()&os.ModeCharDevice != 0 {
		// Stdin is a TTY, not a pipe — no prompt available.
		return "", ferrors.NewUsage("spec.new", "a prompt is required (use --prompt, --from-issue, or pipe via stdin)")
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
			return "", ferrors.Wrap("spec.new.readPrompt", err, "read stdin")
		}
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", ferrors.NewUsage("spec.new", "a prompt is required (use --prompt, --from-issue, or pipe via stdin)")
	}
	return out, nil
}
