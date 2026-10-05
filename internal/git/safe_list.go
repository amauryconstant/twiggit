package git

import (
	"context"
	"twiggit/internal/core"
)

// SafeListWorktrees returns the worktree list for repoPath and a
// "safe" flag. The flag is true when the listing succeeded (even if
// the slice is empty); it is false when the listing failed with an
// I/O error the caller should surface. The underlying error is
// intentionally NOT returned; the boolean is the entire contract.
// Callers that need the underlying error must invoke
// Client.ListWorktrees directly.
//
// The repoPath parameter deviates from the task brief's two-arg
// signature because *Client has no implicit "current repo" state —
// the repository to inspect must be passed explicitly.
//
// Rationale: cmd/create.go and cmd/prune.go had two near-identical
// "list worktrees, swallow on failure" guards. The shared helper
// makes the intent explicit — "I want the slice, do not fail the
// caller" — while keeping the failure path observable through the
// bool so the few callers that DO want to act on failure can branch
// on it.
func SafeListWorktrees(ctx context.Context, c *Client, repoPath string) (worktrees []core.Worktree, safe bool) {
	if c == nil {
		return nil, false
	}
	worktrees, err := c.ListWorktrees(ctx, repoPath)
	if err != nil {
		return nil, false
	}
	return worktrees, true
}
