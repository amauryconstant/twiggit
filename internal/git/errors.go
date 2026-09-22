// Package git error type.
//
// ExternalError is the single error type returned by every public method in
// internal/git. It wraps the underlying tool failure (go-git error chain or
// os/exec output), exposes a stable Operation tag for cmd-side dispatch,
// and embeds *core.OperationError so callers can keep using errors.As with
// the existing core.OperationError target during the migration.
package git

import (
	"context"
	"errors"
	"fmt"

	"twiggit/internal/core"
)

// ErrorKind classifies ExternalError instances for callers that need to
// branch on failure shape without walking the cause chain.
type ErrorKind int

const (
	// ErrorKindOther is the default for failures that don't map cleanly to
	// one of the named kinds (network blips, malformed CLI output, etc.).
	ErrorKindOther ErrorKind = iota
	// ErrorKindNotFound marks a missing resource (branch, ref, worktree).
	ErrorKindNotFound
	// ErrorKindPermission marks EACCES / sandbox / hook-permission failures.
	ErrorKindPermission
	// ErrorKindTimeout marks context-deadline or command-timeout failures.
	ErrorKindTimeout
)

// ExternalError is the git adapter's stable error surface.
//
// Tool is always "git" for this package; Operation names the public method
// (e.g. "OpenRepository", "CreateWorktree"); Message is a human-readable
// summary safe to render directly; Cause is the underlying error chain.
//
// *core.OperationError is embedded so errors.As walks to the canonical
// core type for cmd-side dispatch. The embedded OperationError.Op is
// the high-level tag (e.g. "git.worktree"); the public Operation field
// above is the per-call method name (e.g. "git worktree add failed: ...").
type ExternalError struct {
	Tool      string
	Operation string
	Message   string
	Cause     error
	Kind      ErrorKind

	*core.OperationError
}

// Error returns "<Tool> <Operation>: <Message>". The cause chain is not
// inlined; callers that want the full chain can use errors.Unwrap or
// output.FormatError which walks the chain under TWIGGIT_DEBUG.
func (e *ExternalError) Error() string {
	return fmt.Sprintf("%s %s: %s", e.Tool, e.Operation, e.Message)
}

// Unwrap returns Cause so errors.Is and errors.As walk to the originating
// error (typically a go-git plumbing error or *exec.ExitError). The
// embedded *core.OperationError also implements Unwrap, so the chain
// visits the operation context as well.
func (e *ExternalError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return e.OperationError
}

// Is matches target against the wrapped NotFound sentinels when the
// ExternalError's Kind is ErrorKindNotFound. This preserves the contract
// callers had with the legacy core.NewGitRepositoryError /
// core.NewGitWorktreeError constructors so errors.Is(err,
// core.ErrGitRepoNotFound) keeps working across the migration.
func (e *ExternalError) Is(target error) bool {
	if e.Kind != ErrorKindNotFound {
		return false
	}
	switch target {
	case core.ErrGitRepoNotFound:
		return e.OperationError != nil && e.OperationError.Op == "git.repository"
	case core.ErrWorktreeNotFound:
		return e.OperationError != nil && e.OperationError.Op == "git.worktree"
	}
	return false
}

// NewRepoError constructs an ExternalError tagged with Op="git.repository"
// and a Kind derived from Cause. This is the canonical repo-level error
// constructor for the read-side methods.
func NewRepoError(op, msg string, cause error) *ExternalError {
	return newExternalError("git.repository", op, msg, cause)
}

// NewWorktreeError constructs an ExternalError tagged with Op="git.worktree".
// Cause is the underlying error from the git CLI invocation or go-git call.
// Used by the write-side methods (CreateWorktree, DeleteWorktree, etc.).
func NewWorktreeError(op, msg string, cause error) *ExternalError {
	return newExternalError("git.worktree", op, msg, cause)
}

// NewCommandError constructs an ExternalError tagged with Op="git.command".
// Used by the command executor wrapper to surface os/exec failures with
// consistent Tool/Operation prefixes.
func NewCommandError(op, msg string, cause error) *ExternalError {
	return newExternalError("git.command", op, msg, cause)
}

// newExternalError is the shared constructor. tag is the high-level
// Op identifier (e.g. "git.repository"); op is the public method name;
// msg and cause flow through unchanged. The embedded *core.OperationError
// carries the high-level tag for errors.As dispatch.
func newExternalError(tag, op, msg string, cause error) *ExternalError {
	return &ExternalError{
		Tool:      "git",
		Operation: op,
		Message:   msg,
		Cause:     cause,
		Kind:      classifyKind(cause),
		OperationError: &core.OperationError{
			Op:      tag,
			Message: msg,
			Cause:   cause,
		},
	}
}

// classifyKind maps a cause error to an ErrorKind. The heuristic is
// deliberately conservative: anything we can match returns its kind;
// everything else falls through to ErrorKindOther.
func classifyKind(cause error) ErrorKind {
	if cause == nil {
		return ErrorKindOther
	}
	if errors.Is(cause, context.DeadlineExceeded) {
		return ErrorKindTimeout
	}
	return ErrorKindOther
}
