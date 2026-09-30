// Package errors defines the error taxonomy and exit-code contract for free-kiro.
//
// Exit-code contract (also documented in docs/CLI.md and kiro-spec skill):
//
//	0 = success
//	1 = lint gate failure (lint found at least one ERROR)
//	2 = engine / internal failure (KiroError: workspace missing, illegal phase
//	    transition, IO error, etc.). IDE PreToolUse hooks MUST exit 2 to actually
//	    block a write — write `free-kiro lint || exit 2`.
//	3 = user input error (missing arg, name collision, etc.)
//
// Every error in free-kiro implements this taxonomy via one of the typed
// errors below; the CLI layer (internal/cli) maps them to exit codes via
// errors.Is / errors.As.
package errors

import (
	"errors"
	"fmt"
)

// KiroError is the root of the engine-error hierarchy. Any error not derived
// from KiroError is treated as a programming bug (panic in debug, generic
// exit-2 in release).
type KiroError struct {
	Op  string // operation that produced the error (e.g. "spec.new")
	Err error  // wrapped cause (may be nil)
	Msg string // human-readable explanation (Chinese, user-facing)
}

func (e *KiroError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Msg)
}

func (e *KiroError) Unwrap() error { return e.Err }

// New constructs a KiroError without a wrapped cause.
func New(op, msg string) *KiroError {
	return &KiroError{Op: op, Msg: msg}
}

// Wrap attaches a cause to an engine error.
func Wrap(op string, err error, msg string) *KiroError {
	return &KiroError{Op: op, Err: err, Msg: msg}
}

// NewUsage constructs a UsageError (caller-misuse; exit code 3) without
// a wrapped cause. Sugar for `&UsageError{KiroError: New(op, msg)}` so
// call sites in cli/ don't have to type the wrapper every time.
func NewUsage(op, msg string) *UsageError {
	return &UsageError{KiroError: New(op, msg)}
}

// WrapUsage constructs a UsageError (exit code 3) that wraps an
// underlying IO / parse error. Use when a caller misuses the CLI in a
// way that surfaces a low-level error (e.g. stdin read failure with no
// prompt supplied).
func WrapUsage(op string, err error, msg string) *UsageError {
	return &UsageError{KiroError: Wrap(op, err, msg)}
}

// NewTaskGraphError constructs a TaskGraphError (exit code 1) for
// unparseable / cyclic tasks.md. Sugar mirroring NewUsage so cli/task.go
// can return it directly.
func NewTaskGraphError(op, msg string) *TaskGraphError {
	return &TaskGraphError{KiroError: New(op, msg)}
}

// NewLintFailureError constructs a LintFailureError (exit code 1) for
// `free-kiro lint [name]` ERROR-severity findings. Distinct from
// LintGateError (exit 2, used by `spec approve`) so callers can tell
// "the lint subcommand itself failed" apart from "an upstream
// operation was blocked by lint".
func NewLintFailureError(op, msg string) *LintFailureError {
	return &LintFailureError{KiroError: New(op, msg)}
}

// Typed KiroError subclasses — each represents a distinct failure mode that
// the CLI layer may want to react to specifically. Use these instead of bare
// New() when the failure mode is recognisable.

// WorkspaceError indicates the .kiro directory is missing or unreadable.
// Maps to exit code 2.
type WorkspaceError struct{ *KiroError }

// TransitionError indicates an illegal phase transition in the spec
// state machine. Maps to exit code 2.
type TransitionError struct{ *KiroError }

// LintGateError indicates a lint gate blocked an advance/approve.
// Maps to exit code 2.
type LintGateError struct{ *KiroError }

// LintFailureError indicates the `free-kiro lint [name]` subcommand
// found at least one ERROR-severity issue. Maps to exit code 1 —
// matches the contract in docs/CLI.md / errors.go header
// ("1 = lint gate failure"). Distinct from LintGateError (exit 2)
// which is used when an upstream caller like `spec approve` is blocked
// by lint and the operation itself failed.
type LintFailureError struct{ *KiroError }

// TaskGraphError indicates tasks.md has a cycle or other unparseable
// structure. Maps to exit code 1.
type TaskGraphError struct{ *KiroError }

// SteeringError indicates a steering doc is malformed (bad frontmatter,
// invalid mode). Maps to exit code 2.
type SteeringError struct{ *KiroError }

// HookError indicates a hook definition is malformed or its action
// failed. Maps to exit code 2.
type HookError struct{ *KiroError }

// UsageError indicates the caller misused the CLI (missing arg, bad
// name, etc.). Maps to exit code 3.
type UsageError struct{ *KiroError }

// ExitCode maps an error to the CLI exit code per the contract above.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var u *UsageError
	if errors.As(err, &u) {
		return 3
	}
	var l *LintFailureError
	if errors.As(err, &l) {
		return 1
	}
	var lge *LintGateError
	if errors.As(err, &lge) {
		return 2
	}
	var t *TaskGraphError
	if errors.As(err, &t) {
		return 1
	}
	var k *KiroError
	if errors.As(err, &k) {
		return 2
	}
	// Unknown error type — treat as engine failure.
	return 2
}
