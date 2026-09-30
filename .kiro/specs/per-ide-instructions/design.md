# per-ide-instructions — Design

## Architecture

- `internal/ide/ide.go` is the single owner of the IDE→instruction-files
  mapping and the writer entry point. We extend it with a new
  `instructionFiles map[ID][]string` constant and a new exported
  `WriteAgentInstructions(root, lang, overwrite, ids)` function. The legacy
  `WriteAgentsMD` is preserved as a thin deprecated wrapper so any third-party
  caller or skill bundle keeps compiling for one minor version.
- The two existing templates (`agents_zh.md`, `agents_en.md`) keep serving
  the OpenCode / CodeBuddy path. We add two more templates
  (`instructions_zh.md`, `instructions_en.md`) sharing the same prose but
  addressed at the Claude Code / Cursor / Continue convention.
- `internal/cli/init.go` calls the new writer once after the existing
  `InstallHooks` loop. The CLI flag surface gets one new
  `--overwrite-instructions` flag; the existing `--overwrite-agents` flag is
  kept but its usage string marks it deprecated and the new flag is the
  canonical surface going forward.
- `internal/cli/doctor_checks.go` gets a new check that, for every IDE id
  the user has configured, asserts the per-IDE instruction file exists and
  contains the `# free-kiro-managed:` marker. The existing hook count check
  stays unchanged.

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `instructionFiles` | Canonical `map[ID][]string` of IDE to instruction file paths. | `const instructionFiles` |
| `WriteAgentInstructions` | Iterate `ids`, call `WriteSingleInstruction`, dedupe by absolute path, collect results. | `func WriteAgentInstructions(root, lang string, overwrite bool, ids []ID) ([]string, error)` |
| `WriteSingleInstruction` | For one IDE id: resolve paths, dedupe intra-id, `os.MkdirAll` parents, `os.Stat` skip or write, return written paths. | `func WriteSingleInstruction(root, lang string, overwrite bool, id ID) ([]string, error)` |
| `AgentTemplate` | Pick `instructions_{zh,en}.md` vs `agents_{zh,en}.md` based on `id` (OpenCode/CodeBuddy → agents; others → instructions). | `func agentTemplate(lang string, id ID)` |
| `MarkFreeKiro` | Constant first-line marker for free-kiro-managed files. | `const freeKiroInstructionMarker = "# free-kiro-managed:"` |
| `IsFreeKiroInstruction` | Detect whether a path is owned by free-kiro (read first line, marker prefix). | `func IsFreeKiroInstruction(path string) (bool, error)` |
| `DoctorCheckInstructions` | New doctor check: per configured IDE, every file in `instructionFiles[id]` exists and is marked. | `func doctorCheckInstructions(home string, detected []Info) error` |
| `WriteAgentsMD` (deprecated) | Thin wrapper that calls `WriteAgentInstructions` for the AGENTS.md file only; emits deprecation comment in godoc. | `func WriteAgentsMD(root, lang string, overwrite bool) (string, error)` |

## Data Model

- No new exported struct types.
- One new package-level constant `instructionFiles` of type `map[ID][]string`,
  initialised inside an `init()` block (mirroring the style of `freeKiroHooks`
  in the same file at `internal/ide/ide.go:172`).
- One new constant `freeKiroInstructionMarker string` for the marker.
- No new Go module dependencies. We stay on stdlib + the existing
  `embed.FS` already used for templates.

## Error Handling

- All file write failures are wrapped via `ferrors.Wrap(scope, err, "write
  <path>")` using the existing `ferrors` package, matching the style at
  `internal/ide/ide.go:340`.
- The deprecated `--overwrite-agents` flag does not raise an error; it
  prints a `Stderr` warning to the existing `writeOut(os.Stderr, ...)`
  helper in `internal/cli/init.go:152`.
- The new writer returns the joined error from any per-file failure; the
  caller (`runIdeInit`) treats IDE-level failures as non-fatal, matching
  the existing pattern at `internal/cli/init.go:73-77`.
- `IsFreeKiroInstruction` reads the first line via `bufio.Scanner` and
  returns `(false, nil)` on `os.IsNotExist`, `false, err` on other read
  failures, `(true, nil)` when the marker prefix matches.

## Testing Strategy

- New file `internal/ide/ide_instructions_test.go` with table-driven cases
  for `WriteAgentInstructions`:
  - Each of the 5 IDE ids writes the canonical files.
  - Pre-existing files are skipped without `--overwrite-instructions`.
  - Pre-existing files are overwritten with `--overwrite-instructions=true`.
  - `OpenCode` and `CodeBuddy` together produce a single `AGENTS.md`.
  - `--ide none` skips IDE writes but still produces `.kiro/AGENTS.md`
    through the legacy path.
- New test for `IsFreeKiroInstruction` covering empty file, missing file,
  and file with marker.
- New test for `DoctorCheckInstructions` happy and missing-file paths.
- Existing `internal/ide/ide_test.go` continues to pass (we keep
  `WriteAgentsMD` exported signature).
- All new tests use `t.TempDir()` per docs/CODING_STYLE.md §5.2.
- CI gate: `go test -race ./...` and `golangci-lint run` keep passing per
  docs/CODING_STYLE.md §8.5.

## Migration / Rollout

- No data migration. `.kiro/AGENTS.md` is preserved on disk and continues
  to be rewritten on `init`.
- Users opt in by re-running `free-kiro init --ide <id>`; the new files
  appear at the project root and in `.cursor/rules/` / `.continue/rules/`.
- The deprecated `--overwrite-agents` flag survives for one minor version;
  release notes flag the rename.
- Rollback: revert the PR. No persisted state outside the files we add.