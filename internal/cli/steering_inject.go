package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jingyu525/free-kiro/internal/steering"
)

// Exit-code contract for `free-kiro steering inject`. Per spec
// steering-inject-to-ide §AC:
//
//	0 — all 5 target files (or the --only subset) successfully updated
//	3 — no always-mode docs found in .kiro/steering/
//	4 — --only glob pattern is invalid
//	1..5 — that many target files were skipped (missing marker, IO err);
//	        capped at 5 because there are at most 5 targets
const (
	injectExitOK          = 0
	injectExitNoDocs      = 3
	injectExitBadGlob     = 4
	injectExitMaxSkipped  = 5
)

// steeringInjectCmd wires `free-kiro steering inject` — assemble the
// always-mode steering docs into a markdown block and write it into the
// marker region of each IDE project-root instruction file. See spec
// `.kiro/specs/steering-inject-to-ide/` for the full contract.
func steeringInjectCmd() *cobra.Command {
	var dryRun bool
	var onlyGlob string

	c := &cobra.Command{
		Use:   "inject",
		Short: "把 always-mode steering 注入 IDE 指令文件的 marker 区域",
		Long: `读取 .kiro/steering/*.md 中 mode: always 的文档，按文件名字母序拼成 1 个
markdown 块，写入项目根 IDE 指令文件（CLAUDE.md / AGENTS.md / .cursorrules
等 5 个文件）的 marker 区域内。marker 之外的内容不会被修改。

  --dry-run       只打印将注入的内容，不写文件
  --only <glob>   只处理路径匹配 glob 的目标文件（* 与 ?，不接受 **）

退出码：
  0   全部目标文件写入成功
  3   未找到 always-mode steering 文档
  4   --only glob 语法非法
  1-5 N 个目标文件被跳过（缺 marker / IO 错），上限 5`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := newSteeringStore()
			if err != nil {
				return err
			}
			// Validate --only glob before touching the store, so a bad
			// pattern fails fast with exit 4 (per spec).
			if onlyGlob != "" {
				if _, matchErr := filepath.Match(onlyGlob, "dummy"); matchErr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(),
						"error: invalid glob pattern: %s\n", onlyGlob)
					os.Exit(injectExitBadGlob)
				}
			}
			res := store.InjectAll(onlyGlob)
			runInject(cmd, &res, dryRun)
			return nil
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "只打印将注入的内容，不写文件")
	c.Flags().StringVar(&onlyGlob, "only", "", "只处理路径匹配 glob 的目标文件")
	return c
}

// runInject prints the result of an InjectAll call and exits with the
// appropriate code. Extracted from RunE so tests can drive it via the
// cobra command's stdout/stderr buffers.
func runInject(cmd *cobra.Command, res *steering.InjectResult, dryRun bool) {
	if res.DocCount == 0 {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"error: no always-mode steering docs found in .kiro/steering/\n")
		os.Exit(injectExitNoDocs)
	}
	if dryRun {
		// Print the would-be block so users can eyeball it before
		// committing. Trailing newline keeps the terminal tidy.
		fmt.Fprintln(cmd.OutOrStdout(), res.Block)
		return
	}
	// Real write — surface per-target results.
	for _, rel := range res.Written {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ wrote %s\n", rel)
	}
	for _, sk := range res.Skipped {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"warning: file %s: %s\n", sk.Path, sk.Reason)
	}
	if n := len(res.Skipped); n > 0 {
		if n > injectExitMaxSkipped {
			n = injectExitMaxSkipped
		}
		os.Exit(n)
	}
}
