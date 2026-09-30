package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	ferrors "github.com/jingyu525/free-kiro/internal/errors"
	"github.com/jingyu525/free-kiro/internal/lint"
	"github.com/jingyu525/free-kiro/internal/workspace"
)

// lintCmd runs the offline spec-quality gate. Exits 0 when no ERROR is
// found across all specs (or the named spec), 1 when at least one ERROR
// is present. Exit 1 is the IDE hook contract — PreToolUse hooks
// mapping this command to `free-kiro lint || exit 2` use exit 2 to
// actually block a write.
func lintCmdFactory() *cobra.Command {
	var strictBaseline bool
	cmd := &cobra.Command{
		Use:   "lint [name]",
		Short: "离线质量门禁（exit 1 = ERROR，可被 IDE PreToolUse 拦截）",
		Long: `静态检查 spec 文档的质量问题，不调用任何模型。

  $ free-kiro lint                            # lint 全部 specs
  $ free-kiro lint my-spec                    # 只 lint my-spec
  $ free-kiro lint my-spec --strict-baseline  # 忽略 .baseline.json

退出码：
  0 = 全部 OK
  1 = 发现 ERROR（no-ears / placeholder-ac / tasks 环/悬挂/自引用 …）
  2 = 引擎错误（workspace 不存在、spec 不存在 等）

--strict-baseline 强制忽略 .baseline.json 中所有白名单条目,
等同于历史 spec 真正想"重新审视"时的严格模式。

IDE hook 配置示例（写到 ~/.claude/settings.json 或项目 settings.json）：
  {
    "hooks": [{
      "name": "kiro-lint-gate",
      "trigger": "PreToolUse",
      "matcher": "Edit|Write",
      "action": {
        "type": "command",
        "command": "free-kiro lint || exit 2"
      }
    }]
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			holder, err := loadEngine()
			if err != nil {
				return err
			}
			if len(args) >= 1 {
				return lintOne(holder.ws, args[0], strictBaseline)
			}
			return lintAll(holder.ws, cmd, strictBaseline)
		},
	}
	cmd.Flags().BoolVar(&strictBaseline, "strict-baseline", false,
		"ignore .baseline.json (treat as not configured)")
	return cmd
}

func lintOne(ws *workspace.Workspace, name string, strictBaseline bool) error {
	dir := ws.SpecDir(name)
	issues := lintSpec(dir, strictBaseline)
	printIssues(name, issues)
	return gateExitCode(issues)
}

func lintAll(ws *workspace.Workspace, cmd *cobra.Command, strictBaseline bool) error {
	specsDir := ws.SpecsDir()
	entries, err := readDir(specsDir)
	if err != nil {
		writeOutln(cmd.OutOrStdout(), "(no specs to lint)")
		return nil
	}
	failed := 0
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		dir := filepath.Join(specsDir, name)
		issues := lintSpec(dir, strictBaseline)
		printIssues(name, issues)
		if anyError(issues) {
			failed++
		}
	}
	if failed > 0 {
		// Surface a non-zero exit code via the cobra error machinery.
		// LintFailureError maps to exit 1, matching the contract in
		// docs/CLI.md and the smoke test's `expect exit 1 (placeholder)`.
		return ferrors.NewLintFailureError("lint.all",
			fmt.Sprintf("%d spec(s) failed lint", failed))
	}
	return nil
}

// lintSpec dispatches to lint.Spec or lint.SpecStrict based on the
// --strict-baseline flag. SpecStrict is a placeholder that currently
// behaves identically to Spec — task #13 wires up the actual
// "skip baseline" logic.
func lintSpec(specDir string, strictBaseline bool) []lint.Issue {
	if strictBaseline {
		return lint.SpecStrict(specDir)
	}
	return lint.Spec(specDir)
}

func printIssues(name string, issues []lint.Issue) {
	if len(issues) == 0 {
		fmt.Printf("%s: OK\n", name)
		return
	}
	errors, warnings := 0, 0
	for _, i := range issues {
		if i.Severity == lint.SeverityError {
			errors++
		} else {
			warnings++
		}
	}
	fmt.Printf("%s: %d error(s), %d warning(s)\n", name, errors, warnings)
	for _, i := range issues {
		fmt.Printf("  %s\n", i)
	}
}

func anyError(issues []lint.Issue) bool {
	for _, i := range issues {
		if i.Severity == lint.SeverityError {
			return true
		}
	}
	return false
}

// gateExitCode returns an error so cobra exits with code 1 when any
// ERROR-severity issue is present, 0 otherwise. Used by `lint <name>`.
// Uses LintFailureError so main.go's ExitCode maps it to exit 1
// (matches the contract in docs/CLI.md + the smoke test).
func gateExitCode(issues []lint.Issue) error {
	if anyBlockingError(issues) {
		return ferrors.NewLintFailureError("lint.one", "lint gate failed")
	}
	return nil
}

// anyBlockingError mirrors the gate's filtering: ERROR-severity issues
// whose Message starts with "[baseline] " are whitelisted by the spec's
// .baseline.json and should not cause exit 1. Mirrors
// internal/lint.Gate() so that `free-kiro lint <name>` matches the
// approve/advance gate.
func anyBlockingError(issues []lint.Issue) bool {
	for _, i := range issues {
		if i.Severity != lint.SeverityError {
			continue
		}
		if strings.HasPrefix(i.Message, "[baseline] ") {
			continue
		}
		return true
	}
	return false
}
