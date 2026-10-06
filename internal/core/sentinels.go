package core

import "errors"

// Canonical sentinel catalog. Includes the four NotFound sentinels and
// the five shell-related sentinels consolidated from shell_errors.go so
// the package has a single source of truth for errors.Is matching.
//
// Identifiers match the NotFoundError.Is / OperationError.Is
// membership lists.
var (
	ErrGitRepoNotFound       = errors.New("core: git repository not found")
	ErrWorktreeNotFound      = errors.New("core: worktree not found")
	ErrProjectNotFound       = errors.New("core: project not found")
	ErrResolutionNotFound    = errors.New("core: resolution target not found")
	ErrShellAlreadyInstalled = errors.New("core: shell wrapper already installed")
	ErrShellNotInstalled     = errors.New("core: shell wrapper not installed")
	ErrInvalidShellType      = errors.New("core: invalid shell type")
	ErrInferenceFailed       = errors.New("core: could not infer shell type")
	ErrDetectionFailed       = errors.New("core: shell detection failed")
	ErrUncommittedChanges    = errors.New("core: worktree has uncommitted changes")
	ErrRebaseConflict        = errors.New("core: rebase conflict")
	ErrRebaseInProgress      = errors.New("core: no rebase in progress")
	ErrBaseNotSet            = errors.New("core: tracked base not set")
)
