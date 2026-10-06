package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"twiggit/internal/core"
)

// Rebase runs `git rebase <onto>` inside the worktree at wtPath. The
// returned RebaseOutcome classifies the result: Clean, Conflicted
// (worktree left in mid-rebase), NothingToDo, or Aborted on a
// non-conflict failure. The adapter wraps the cause via the
// rebase.* OperationError namespace so callers can branch on
// core.ErrRebaseConflict via errors.Is.
func (c *cliClient) Rebase(ctx context.Context, wtPath, onto string) (core.RebaseOutcome, error) {
	if wtPath == "" {
		return core.RebaseOutcomeAborted, NewRebaseError("rebase", "worktree path cannot be empty", nil)
	}
	if onto == "" {
		return core.RebaseOutcomeAborted, NewRebaseError("rebase", "onto branch cannot be empty", nil)
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, wtPath, CmdGit, c.defaultTimeout, "rebase", onto)
	if err != nil {
		if result != nil && isRebaseConflictStderr(result.Stderr) {
			return core.RebaseOutcomeConflicted, NewRebaseError("rebase.worktree", "git rebase hit a conflict: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseConflict, result.Stderr))
		}
		return core.RebaseOutcomeAborted, NewRebaseError("rebase.worktree", "git rebase failed: "+result.Stderr, err)
	}
	if result == nil {
		return core.RebaseOutcomeAborted, NewRebaseError("rebase.worktree", "command executor returned nil result for rebase", nil)
	}
	if result.ExitCode != 0 {
		if isRebaseConflictStderr(result.Stderr) {
			return core.RebaseOutcomeConflicted, NewRebaseError("rebase.worktree", "git rebase hit a conflict: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseConflict, result.Stderr))
		}
		if isNothingToDoStderr(result.Stderr) || isNothingToDoStdout(result.Stdout) {
			return core.RebaseOutcomeNothingToDo, nil
		}
		return core.RebaseOutcomeAborted, NewRebaseError("rebase.worktree", "git rebase failed: "+result.Stderr, result.Err)
	}
	if isNothingToDoStdout(result.Stdout) {
		return core.RebaseOutcomeNothingToDo, nil
	}
	return core.RebaseOutcomeClean, nil
}

// Abort runs `git rebase --abort` in the worktree. Returns the wrapped
// error only when the underlying git call fails; a successful abort
// (including the no-op case where no rebase is in progress) returns
// nil because the end state — "no mid-rebase state" — is what the
// caller wanted.
func (c *cliClient) Abort(ctx context.Context, wtPath string) error {
	if wtPath == "" {
		return NewRebaseError("abort", "worktree path cannot be empty", nil)
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, wtPath, CmdGit, c.defaultTimeout, "rebase", "--abort")
	if err != nil {
		return NewRebaseError("abort.worktree", "git rebase --abort failed: "+result.Stderr, err)
	}
	if result == nil {
		return NewRebaseError("abort.worktree", "command executor returned nil result for rebase --abort", nil)
	}
	if result.ExitCode != 0 {
		return NewRebaseError("abort.worktree", "git rebase --abort failed: "+result.Stderr, result.Err)
	}
	return nil
}

// Continue runs `git rebase --continue` in the worktree. When no
// rebase is in progress git exits non-zero with a "No rebase in
// progress" message; the adapter maps that to
// (RebaseOutcomeAborted, ErrRebaseInProgress) so callers can branch
// on the absent-rebase state without string-matching.
func (c *cliClient) Continue(ctx context.Context, wtPath string) (core.RebaseOutcome, error) {
	if wtPath == "" {
		return core.RebaseOutcomeAborted, NewRebaseError("continue", "worktree path cannot be empty", nil)
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, wtPath, CmdGit, c.defaultTimeout, "rebase", "--continue")
	if err != nil {
		if result != nil && isRebaseConflictStderr(result.Stderr) {
			return core.RebaseOutcomeConflicted, NewRebaseError("continue.worktree", "git rebase --continue hit a conflict: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseConflict, result.Stderr))
		}
		if result != nil && isNoRebaseInProgressStderr(result.Stderr) {
			return core.RebaseOutcomeAborted, NewRebaseError("continue.worktree", "no rebase in progress: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseInProgress, result.Stderr))
		}
		return core.RebaseOutcomeAborted, NewRebaseError("continue.worktree", "git rebase --continue failed: "+result.Stderr, err)
	}
	if result == nil {
		return core.RebaseOutcomeAborted, NewRebaseError("continue.worktree", "command executor returned nil result for rebase --continue", nil)
	}
	if result.ExitCode != 0 {
		if isRebaseConflictStderr(result.Stderr) {
			return core.RebaseOutcomeConflicted, NewRebaseError("continue.worktree", "git rebase --continue hit a conflict: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseConflict, result.Stderr))
		}
		if isNoRebaseInProgressStderr(result.Stderr) {
			return core.RebaseOutcomeAborted, NewRebaseError("continue.worktree", "no rebase in progress: "+result.Stderr, fmt.Errorf("%w: %s", core.ErrRebaseInProgress, result.Stderr))
		}
		return core.RebaseOutcomeAborted, NewRebaseError("continue.worktree", "git rebase --continue failed: "+result.Stderr, result.Err)
	}
	return core.RebaseOutcomeClean, nil
}

// Fetch runs `git fetch <remote> <ref>` in the repo at repoPath.
// Used by `twiggit rebase --fetch` to refresh the tracked base
// before rebasing; the rebase command is offline by default.
func (c *cliClient) Fetch(ctx context.Context, repoPath, remote, ref string) error {
	if repoPath == "" {
		return NewRebaseError("fetch", "repository path cannot be empty", nil)
	}
	if remote == "" {
		return NewRebaseError("fetch", "remote cannot be empty", nil)
	}
	if ref == "" {
		return NewRebaseError("fetch", "ref cannot be empty", nil)
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, CmdGit, c.defaultTimeout, "fetch", remote, ref)
	if err != nil {
		return NewRebaseError("fetch.worktree", "git fetch failed: "+result.Stderr, err)
	}
	if result == nil {
		return NewRebaseError("fetch.worktree", "command executor returned nil result for fetch", nil)
	}
	if result.ExitCode != 0 {
		return NewRebaseError("fetch.worktree", "git fetch failed: "+result.Stderr, result.Err)
	}
	return nil
}

// SetTrackedBase writes `twiggit.tracked-base` to per-worktree config
// via `git -C <wtPath> config --worktree twiggit.tracked-base <base>`.
// The `--worktree` flag scopes the write to the worktree's private
// config so the value travels with the worktree on `git worktree move`
// and is removed on `git worktree remove`.
func (c *cliClient) SetTrackedBase(ctx context.Context, wtPath, base string) error {
	if wtPath == "" {
		return NewBaseTrackerError("set", "worktree path cannot be empty", nil)
	}
	if base == "" {
		return NewBaseTrackerError("set", "base branch cannot be empty", nil)
	}

	absWt, err := filepath.Abs(wtPath)
	if err != nil {
		absWt = wtPath
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, absWt, CmdGit, c.defaultTimeout, "config", "--worktree", "twiggit.tracked-base", base)
	if err != nil {
		return NewBaseTrackerError("set.tracked-base", "git config --worktree failed: "+result.Stderr, err)
	}
	if result == nil {
		return NewBaseTrackerError("set.tracked-base", "command executor returned nil result for tracked-base set", nil)
	}
	if result.ExitCode != 0 {
		return NewBaseTrackerError("set.tracked-base", "git config --worktree failed: "+result.Stderr, result.Err)
	}
	return nil
}

// GetTrackedBase reads `twiggit.tracked-base` from per-worktree config
// via `git -C <wtPath> config --worktree --get twiggit.tracked-base`.
// Returns ("", nil) — not an error — when the key is absent so the
// cmd layer can decide whether to fall back to a protected branch.
func (c *cliClient) GetTrackedBase(ctx context.Context, wtPath string) (string, error) {
	if wtPath == "" {
		return "", NewBaseTrackerError("get", "worktree path cannot be empty", nil)
	}

	absWt, err := filepath.Abs(wtPath)
	if err != nil {
		absWt = wtPath
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, absWt, CmdGit, c.defaultTimeout, "config", "--worktree", "--get", "twiggit.tracked-base")
	if err == nil && result != nil && result.ExitCode == 0 {
		return strings.TrimSpace(result.Stdout), nil
	}

	// git --get exits 1 with empty stdout when the key is absent; map
	// that to ("", nil). Any other exit code is a real failure.
	if result != nil && result.ExitCode == 1 && strings.TrimSpace(result.Stderr) == "" {
		return "", nil
	}
	if result != nil {
		return "", NewBaseTrackerError("get.tracked-base", "git config --worktree --get failed: "+result.Stderr, result.Err)
	}
	return "", NewBaseTrackerError("get.tracked-base", "command executor returned nil result for tracked-base get", err)
}

// isRebaseConflictStderr reports whether a git rebase stderr line
// indicates a conflict that left the worktree in mid-rebase state.
// Centralised so a future git-version skew that adds a new phrase
// only needs one site updated.
func isRebaseConflictStderr(stderr string) bool {
	return strings.Contains(stderr, "could not apply") ||
		strings.Contains(stderr, "Resolve all conflicts manually") ||
		strings.Contains(stderr, "CONFLICT")
}

// isNothingToDoStderr reports whether git rebase stderr announces
// the branch already contains the upstream tip.
func isNothingToDoStderr(stderr string) bool {
	return strings.Contains(stderr, "up to date") ||
		strings.Contains(stderr, "Your branch is up to date")
}

// isNothingToDoStdout mirrors isNothingToDoStderr for the stdout
// channel. Some git versions print the "up to date" notice on
// stdout when the rebase had nothing to apply.
func isNothingToDoStdout(stdout string) bool {
	return strings.Contains(stdout, "up to date") ||
		strings.Contains(stdout, "Your branch is up to date")
}

// isNoRebaseInProgressStderr reports whether `git rebase --continue`
// (or `git rebase --abort`) was invoked with no rebase in flight.
// Distinct from the conflict phrase so callers can map the absent
// state to ErrRebaseInProgress.
func isNoRebaseInProgressStderr(stderr string) bool {
	return strings.Contains(stderr, "No rebase in progress") ||
		strings.Contains(stderr, "no rebase in progress")
}
