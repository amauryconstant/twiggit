package core

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinels: only the four NotFound sentinels remain. Walk-to-sentinel
// behavior is implemented by OperationError.Is / ValidationError.Is for
// the resource they previously identified.

// ValidationError marks an input or argument validation failure.
//
// Op / Entity / Field distinguish the original source constructor
// (ServiceError, WorktreeServiceError, ProjectServiceError,
// NavigationServiceError, ResolutionError, ConflictError, or direct
// NewValidationError use). The hidden cause field carries the wrapped
// error from constructors that accept one (e.g. NewWorktreeServiceError)
// so errors.Is / errors.As walk through the chain.
type ValidationError struct {
	Op          string
	Entity      string
	Field       string
	Value       string
	Message     string
	Suggestions []string
	cause       error
}

// Error formats the validation failure as a single line: optional Op
// namespace, optional Entity / Field path, the Message, and the offending
// Value when set. Trailing whitespace is trimmed.
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

// Unwrap returns the hidden cause so errors.Is / errors.As walk the
// chain. Returns nil when no cause was supplied at construction time.
func (e *ValidationError) Unwrap() error { return e.cause }

// Is maps the Operation identifier to one of the four NotFound sentinels.
// ConflictError / ServiceError / direct NewValidationError do not map.
// Uses HasPrefix so dot-concatenated Op values (e.g. "git.worktree.create")
// still match the namespace sentinel.
func (e *ValidationError) Is(target error) bool {
	switch {
	case e.Op == "worktree.service" || strings.HasPrefix(e.Op, "git.worktree") || e.Op == "WorktreeService":
		return target == ErrWorktreeNotFound
	case e.Op == "project.service" || e.Op == "ProjectService":
		return target == ErrProjectNotFound
	case e.Op == "navigation.service" || e.Op == "NavigationService" || e.Op == "resolution" || e.Op == "Resolution":
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

// Error renders the missing-resource message: "Entity Name not found"
// when Name is set, else "Entity not found".
func (e *NotFoundError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s %q not found", e.Entity, e.Name)
	}
	return e.Entity + " not found"
}

// Unwrap returns nil: NotFoundError is terminal.
func (e *NotFoundError) Unwrap() error { return nil }

// Is reports membership in the four canonical NotFound sentinels.
// Dispatch is on Entity (not on Name) so the same resource kind maps
// to the same sentinel regardless of the specific instance: project
// resources walk to ErrProjectNotFound, worktree resources to
// ErrWorktreeNotFound, etc. The previous catch-all "true for all four
// sentinels" behaviour was unsound because callers expected errors.Is
// to identify the missing resource.
func (e *NotFoundError) Is(target error) bool {
	switch target {
	case ErrGitRepoNotFound:
		return e.Entity == "git.repository" || e.Entity == "git.repo" || e.Entity == "repo" || e.Entity == "repository" || e.Entity == "git repository"
	case ErrWorktreeNotFound:
		return e.Entity == "worktree"
	case ErrProjectNotFound:
		return e.Entity == "project"
	case ErrResolutionNotFound:
		return e.Entity == "resolution" || e.Entity == "navigation" || e.Entity == "resolution target" || e.Entity == "navigation target"
	case ErrBaseNotSet:
		return e.Entity == "tracked-base" || e.Entity == "base"
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

// Error formats the runtime-failure message: optional Op prefix,
// Message body, optional Entity / Field qualifiers, and the chained
// Cause when set.
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
// HasPrefix matches both "project" and "project.<subop>" namespaces
// (e.g. "project.discover", "navigation.navigate", "resolution.find")
// so the four NotFound families each own a clean namespace.
func (e *OperationError) Is(target error) bool {
	if errors.Is(e.Cause, target) {
		return true
	}
	switch target {
	case ErrProjectNotFound:
		return e.Op != "" && (e.Op == "project" || strings.HasPrefix(e.Op, "project."))
	case ErrWorktreeNotFound:
		return e.Op != "" && (e.Op == "worktree" || strings.HasPrefix(e.Op, "worktree.") || strings.HasPrefix(e.Op, "git.worktree"))
	case ErrGitRepoNotFound:
		return e.Op != "" && (e.Op == "repo" || strings.HasPrefix(e.Op, "git.repository") || strings.HasPrefix(e.Op, "git.repo") || strings.HasPrefix(e.Op, "repository"))
	case ErrResolutionNotFound:
		return e.Op != "" && (e.Op == "resolution" || strings.HasPrefix(e.Op, "resolution.") || e.Op == "navigation" || strings.HasPrefix(e.Op, "navigation."))
	case ErrRebaseConflict:
		return e.Op != "" && (e.Op == "rebase" || strings.HasPrefix(e.Op, "rebase."))
	case ErrBaseNotSet:
		return e.Op != "" && (e.Op == "base" || strings.HasPrefix(e.Op, "base."))
	}
	return false
}

// NewUncommittedChangesError constructs an *OperationError that wraps
// ErrUncommittedChanges so callers can detect uncommitted-change
// safety checks via errors.Is. The worktreePath is surfaced via the
// Entity field for hint rendering.
func NewUncommittedChangesError(worktreePath string) *OperationError {
	return &OperationError{
		Op:      "worktree.uncommitted_changes",
		Entity:  worktreePath,
		Message: "worktree has uncommitted changes (use --force to override)",
		Cause:   ErrUncommittedChanges,
	}
}
