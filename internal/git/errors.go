package git

import (
	"context"
	"errors"
	"fmt"
	"twiggit/internal/core"
)

// ErrorKind classifies ExternalError instances for callers that need to
// branch on failure shape without walking the cause chain. Only the kinds
// actually produced by classifyKind are exported; ErrorKindNotFound and
// ErrorKindPermission were retired in the test-observability-error-discipline
// change (NotFound dispatch walks through *core.NotFoundError.Is() over the
// cause chain instead).
type ErrorKind int

const (
	// ErrorKindOther is the default for failures that don't map cleanly to
	// one of the named kinds (network blips, malformed CLI output, etc.).
	ErrorKindOther ErrorKind = iota
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

// Unwrap returns both the embedded *core.OperationError and the underlying
// Cause so errors.As walks to OperationError first (preserving the Op/
// Message contract that callers dispatch on) and then to the originating
// tool error (go-git plumbing error or *exec.ExitError). When only one
// side is populated, Unwrap returns that side alone.
func (e *ExternalError) Unwrap() []error {
	switch {
	case e.Cause != nil && e.OperationError != nil:
		return []error{e.OperationError, e.Cause}
	case e.OperationError != nil:
		return []error{e.OperationError}
	case e.Cause != nil:
		return []error{e.Cause}
	}
	return nil
}

// Is is intentionally absent: NotFound dispatch walks through
// *core.OperationError.Is() (embedded) over the cause chain so
// errors.Is(extErr, core.ErrGitRepoNotFound) reaches the right
// sentinel via the dot-prefixed Op mapping in OperationError.Is.
// The retired *ExternalError.Is method short-circuited on
// ErrorKindNotFound, but no caller ever set that kind, so the
// method was unreachable. Removing it eliminates the dead branch
// while preserving the contract for every live NotFound path.

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

// NewBranchError constructs an ExternalError tagged with Op="git.branch.<op>".
// Used by branch-mutation methods (DeleteBranch, IsBranchMerged). The
// per-method op is concatenated into the high-level tag so callers can
// dispatch on the precise operation via errors.As + OperationError.Op.
func NewBranchError(op, msg string, cause error) *ExternalError {
	return newExternalError("git.branch", op, msg, cause)
}

// newExternalError is the shared constructor. tag is the high-level
// namespace (e.g. "git.worktree"); op is the per-method suffix (e.g.
// "create"). The embedded *core.OperationError.Op carries the
// dot-concatenated form "tag.op" so callers can dispatch on the precise
// operation via errors.As + OperationError.Op. ExternalError.Operation
// retains the raw per-method suffix for human-readable Error() output.
func newExternalError(tag, op, msg string, cause error) *ExternalError {
	return &ExternalError{
		Tool:      "git",
		Operation: op,
		Message:   msg,
		Cause:     cause,
		Kind:      classifyKind(cause),
		OperationError: &core.OperationError{
			Op:      tag + "." + op,
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
