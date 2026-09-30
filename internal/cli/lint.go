package cli

import (
	"fmt"
	"path/filepath"

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
	return &cobra.Command{
		Use:   "lint [name]",
		Short: "离线质量门禁（exit 1 = ERROR，可被 IDE PreToolUse 拦截）",
		Long: `静态检查 spec 文档的质量问题，不调用任何模型。

  $ free-kiro lint            # lint 全部 specs
  $ free-kiro lint my-spec    # 只 lint my-spec

退出码：
  0 = 全部 OK
  1 = 发现 ERROR（no-ears / placeholder-ac / tasks 环/悬挂/自引用 …）
  2 = 引擎错误（workspace 不存在、spec 不存在 等）

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
				return lintOne(holder.ws, args[0])
			}
			return lintAll(holder.ws, cmd)
		},
	}
}

func lintOne(ws *workspace.Workspace, name string) error {
	dir := ws.SpecDir(name)
	issues := lint.LintSpec(dir)
	printIssues(name, issues)
	return gateExitCode(issues)
}

func lintAll(ws *workspace.Workspace, cmd *cobra.Command) error {
	specsDir := ws.SpecsDir()
	entries, err := readDir(specsDir)
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "(no specs to lint)")
		return nil
	}
	failed := 0
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		dir := filepath.Join(specsDir, name)
		issues := lint.LintSpec(dir)
		printIssues(name, issues)
		if anyError(issues) {
			failed++
		}
	}
	if failed > 0 {
		// Surface a non-zero exit code via the cobra error machinery.
		// LintGateError maps to exit 2, which keeps the IDE hook contract
		// `free-kiro lint || exit 2` working (parent process receives 2
		// regardless of whether it comes from lint or from the fallback).
		return ferrors.New("lint.all", fmt.Sprintf("%d spec(s) failed lint", failed))
	}
	return nil
}

func printIssues(name string, issues []lint.LintIssue) {
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

func anyError(issues []lint.LintIssue) bool {
	for _, i := range issues {
		if i.Severity == lint.SeverityError {
			return true
		}
	}
	return false
}

// gateExitCode returns an error so cobra exits with a non-zero code when
// any ERROR-severity issue is present, 0 otherwise. Used by `lint <name>`.
// Uses the plain KiroError type so main.go's ExitCode maps it to exit 2,
// preserving the existing IDE hook contract (`free-kiro lint || exit 2`).
func gateExitCode(issues []lint.LintIssue) error {
	if anyError(issues) {
		return ferrors.New("lint.one", "lint gate failed")
	}
	return nil
}