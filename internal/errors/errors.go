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

// Typed KiroError subclasses — each represents a distinct failure mode that
// the CLI layer may want to react to specifically. Use these instead of bare
// New() when the failure mode is recognisable.

// WorkspaceError: .kiro directory missing or unreadable. Exit 2.
type WorkspaceError struct{ *KiroError }

// TransitionError: illegal phase transition in the spec state machine. Exit 2.
type TransitionError struct{ *KiroError }

// LintGateError: lint gate blocked an advance/approve. Exit 2.
type LintGateError struct{ *KiroError }

// TaskGraphError: tasks.md has a cycle or other unparseable structure. Exit 1.
type TaskGraphError struct{ *KiroError }

// SteeringError: steering doc malformed (bad frontmatter, invalid mode). Exit 2.
type SteeringError struct{ *KiroError }

// HookError: hook definition malformed or action failed. Exit 2.
type HookError struct{ *KiroError }

// UsageError: caller misused the CLI (missing arg, bad name, etc.). Exit 3.
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
	var l *LintGateError
	if errors.As(err, &l) {
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