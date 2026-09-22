package core

import "errors"

// Canonical sentinel catalog. Only the four NotFound sentinels remain;
// every previously-resource-typed wrapper walks to one of these via
// errors.Is on OperationError.Is / ValidationError.Is.
//
// Identifiers match the NotFoundError Is membership list.
var (
	ErrGitRepoNotFound    = errors.New("core: git repository not found")
	ErrWorktreeNotFound   = errors.New("core: worktree not found")
	ErrProjectNotFound    = errors.New("core: project not found")
	ErrResolutionNotFound = errors.New("core: resolution target not found")
)
