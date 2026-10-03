# cli-version-flag

让 free-kiro CLI 支持 `--version` flag，遵循 GNU 工具惯例：单行输出到
stdout、exit 0，输出含 version / commit / build date。dev build 也能
正常显示。

## User Stories

- As a developer using free-kiro, I want to run `free-kiro --version` to
  see the exact build version, commit hash, and build date, so that I can
  cite a concrete version in bug reports and upgrade decisions.
- As a contributor building free-kiro from `main` (without `-ldflags`
  injection), I want `free-kiro --version` to print a sane `dev`
  placeholder, so that local builds don't panic or print empty strings.

## Acceptance Criteria

- [AC-1] WHEN the user runs `free-kiro --version`, THE SYSTEM SHALL print a single line to stdout in the form `free-kiro version <version> (commit <short-sha>, built <RFC3339-date>)` and exit with code 0.
- [AC-2] WHILE the binary is a dev build (buildVersion == "dev"), THE SYSTEM
  SHALL print a version string that contains the literal `dev` token and
  SHALL NOT panic, print an empty version, or omit the version field.
- [AC-3] WHERE `--version` is combined with other flags or subcommand arguments
  (e.g. `free-kiro --version init`), THE SYSTEM SHALL only print the
  version line, ignore the remaining arguments, and exit 0.
- [AC-4] WHEN the stdout is not a TTY (e.g. `free-kiro --version | cat` or
  redirected to a file), THE SYSTEM SHALL still write exactly one line
  to stdout and SHALL NOT write any version text to stderr.
- [AC-5] UNLESS the version string itself is requested, THE SYSTEM SHALL NOT
  print stack traces, debug info, or internal error details on the
  `--version` code path.
- [AC-6] THE SYSTEM SHALL derive version / commit / date from the `buildVersion`,
  `buildCommit`, `buildDate` package-level variables in `internal/cli`
  (already injected by GoReleaser ldflags in `.goreleaser.yaml`), and
  SHALL NOT introduce new ldflags variables.

## Out of Scope

- A new `free-kiro version` subcommand — `--version` is enough and avoids
  duplicate semantics.
- A `-v` short flag — keep the surface minimal; users can rely on
  `--version` only.
- Any change to `upgrade.go` `displayVersion()` semantics — that function
  stays as-is.
- Introducing new ldflags variables — reuse the existing
  `buildVersion` / `buildCommit` / `buildDate`.
- Changing the version output format of `upgrade` or `skill show`
  subcommands — those keep their existing templates.
