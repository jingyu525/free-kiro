# AGENTS.md — instructions for AI coding agents in this workspace

This project uses **free-kiro** to manage its specification workflow.
free-kiro is a single-binary CLI that gates code changes on written spec
documents. Every change should be spec-first.

## What free-kiro does here

- All specs live under `.kiro/specs/<name>/` with `requirements.md`,
  `design.md`, and `tasks.md` (or `bugfix.md` for fixes).
- A spec moves through phases: draft → requirements → design → tasks →
  approved → implementing → done.
- Phase transitions are gated by `free-kiro lint` (EARS + structure rules).
- Tasks have dependencies drawn in `tasks.md`; `free-kiro task list <name>`
  shows the parallel-wave schedule.

## Before every tool call (except Read)

    SPEC=$(cat .kiro/.current 2>/dev/null) && free-kiro spec next "$SPEC"

This tells you the recommended next command. If lint is failing, fix it
before continuing. The SessionStart hook runs the same command.

If `.kiro/.current` is missing (no active spec), run `free-kiro spec new
<name>` to start one — it auto-marks the new spec as active.

## Before every Edit / Write

    free-kiro lint || exit 2

The IDE's PreToolUse hook will run this automatically. If you see lint
ERRORs in the output, edit the spec documents (not the code) until
`free-kiro lint` is green, then continue.

## When you finish implementing a spec

    free-kiro spec complete <name>

## Where to find help

- `free-kiro --help` — command tree
- `free-kiro doctor` — diagnose install + hook problems
- `docs/EARS.md` — EARS acceptance-criteria syntax
- `docs/HOOKS.md` — hook configuration reference
- `docs/STEERING.md` — project-context injection rules