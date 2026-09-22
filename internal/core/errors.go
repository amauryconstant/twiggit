// Package core error type hierarchy.
//
// Four canonical types collapse the legacy 20-type taxonomy. Errors are
// distinguishable via Op/Entity fields, not via per-type concrete structs.
//
//	ValidationError — input / argument validation failure
//	NotFoundError   — a referenced resource does not exist
//	OperationError  — runtime operation failure (git, config, shell, etc.)
//	UsageError      — invocation-level usage failure (cobra / pflag)
//
// Sentinels: only the four NotFound sentinels remain. Walk-to-sentinel
// behavior is implemented by OperationError.Is / ValidationError.Is for
// the resource they previously identified.
package core

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError marks an input or argument validation failure.
//
// Op / Entity / Field distinguish the original source constructor
// (ServiceError, WorktreeServiceError, ProjectServiceError,
// NavigationServiceError, ResolutionError, ConflictError, or direct
// NewValidationError use). The ValidationError is terminal: its Unwrap
// returns nil so the cmd-side formatter can match it via errors.As
// without a chain hop.
type ValidationError struct {
	Op          string
	Entity      string
	Field       string
	Value       string
	Message     string
	Suggestions []string
}

func (e *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation failed")
	if e.Op != "" {
		sb.WriteString(" for ")
		sb.WriteString(e.Op)
	}
	if e.Entity != "" {
		sb.WriteString(" (")
		sb.WriteString(e.Entity)
		sb.WriteString(")")
	}
	if e.Field != "" {
		sb.WriteString(".")
		sb.WriteString(e.Field)
	}
	sb.WriteString(": ")
	sb.WriteString(e.Message)
	if e.Value != "" {
		sb.WriteString(" (value: ")
		sb.WriteString(e.Value)
		sb.WriteString(")")
	}
	return strings.TrimRight(sb.String(), " \t")
}

// Unwrap returns nil: ValidationError is terminal and has no chain to
// walk. This makes it typed-nil-safe (golang-safety).
func (e *ValidationError) Unwrap() error { return nil }

// Is maps the Operation identifier to one of the four NotFound sentinels.
// ConflictError / ServiceError / direct NewValidationError do not map.
func (e *ValidationError) Is(target error) bool {
	switch e.Op {
	case "worktree.service", "git.worktree", "WorktreeService":
		return target == ErrWorktreeNotFound
	case "project.service", "ProjectService":
		return target == ErrProjectNotFound
	case "navigation.service", "NavigationService", "resolution", "Resolution":
		return target == ErrResolutionNotFound
	}
	return false
}

// NewValidationError constructs a ValidationError with the canonical
// 3-arg signature. Callers needing Op-driven sentinel matching must set
// err.Op directly after construction.
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:       field,
		Value:       value,
		Message:     message,
		Suggestions: []string{},
	}
}

// NewOpValidationError constructs a ValidationError with the Op field
// pre-populated. Use this when the call site needs the operation name
// to drive errors.Is sentinel walking (e.g. "worktree.service" →
// ErrWorktreeNotFound). Functionally a thin convenience over
// NewValidationError + Op assignment; keeps the 4-arg shape at
// call sites that previously used the 4-arg signature.
func NewOpValidationError(op, field, value, message string) *ValidationError {
	ve := NewValidationError(field, value, message)
	ve.Op = op
	return ve
}

// NotFoundError marks a missing resource. Entity names the resource
// kind; Name names the specific instance.
type NotFoundError struct {
	Entity string
	Name   string
}

func (e *NotFoundError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s %q not found", e.Entity, e.Name)
	}
	return fmt.Sprintf("%s not found", e.Entity)
}

// Unwrap returns nil: NotFoundError is terminal.
func (e *NotFoundError) Unwrap() error { return nil }

// Is reports membership in the four canonical NotFound sentinels.
func (e *NotFoundError) Is(target error) bool {
	switch target {
	case ErrGitRepoNotFound, ErrWorktreeNotFound, ErrProjectNotFound, ErrResolutionNotFound:
		return true
	}
	return false
}

// OperationError marks a runtime operation failure (git, config, shell,
// service). Op names the operation (e.g., "git.repository",
// "shell.already_installed"); Entity names the resource the operation
// targets (path / shell type / repo); Cause is the underlying error
// wrapped for errors.Is / errors.As chain walking; Suggestions are
// user-actionable hints rendered after the message.
type OperationError struct {
	Op          string
	Entity      string
	Field       string
	Message     string
	Cause       error
	Suggestions []string
}

func (e *OperationError) Error() string {
	var sb strings.Builder
	if e.Op != "" {
		sb.WriteString(e.Op)
		sb.WriteString(": ")
	}
	sb.WriteString(e.Message)
	if e.Entity != "" {
		sb.WriteString(" (entity: ")
		sb.WriteString(e.Entity)
		sb.WriteString(")")
	}
	if e.Field != "" {
		sb.WriteString(" (field: ")
		sb.WriteString(e.Field)
		sb.WriteString(")")
	}
	if e.Cause != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Cause.Error())
	}
	return sb.String()
}

// Unwrap returns the wrapped cause so errors.Is / errors.As walk the
// chain to the originating error.
func (e *OperationError) Unwrap() error { return e.Cause }

// Is maps the operation identifier to a NotFound sentinel when
// appropriate. The Cause chain is walked first, then the Op table.
func (e *OperationError) Is(target error) bool {
	if errors.Is(e.Cause, target) {
		return true
	}
	switch e.Op {
	case "git.repository", "GitRepository":
		return target == ErrGitRepoNotFound
	case "git.worktree", "GitWorktree":
		return target == ErrWorktreeNotFound
	}
	return false
}
