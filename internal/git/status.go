package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"twiggit/internal/core"
)

// statusOp is the namespace prefix for OperationError.Op values
// raised by ReadWorktreeStatus and helpers.
const statusOp = "status.worktree"

// ReadWorktreeStatus populates the diagnostic projection for a single
// worktree. The composite *Client implements it as a direct method
// because the read spans the go-git reader half (RepositoryStatus,
// ListBranches) and the CLI half (IsBranchMerged, GetTrackedBase,
// rev-list) — neither embedded half can host it alone.
//
// Contract:
//   - On any per-worktree read failure the returned core.WorktreeStatus
//     carries IsSkipped = true and a lowercase SkipReason; the walk
//     continues across other worktrees.
//   - IsStale is intentionally left at the zero value (false). The cmd
//     layer fills it per row via WorktreeStatus.ComputeIsStale(cfg) so
//     per-invocation --stale-behind / --stale-days overrides take effect.
//   - LastChecked is set to time.Now() once at entry; the value is
//     constant across every populated field in the returned row.
//
// A non-nil error is returned only when a precondition fails: empty
// repoPath / wtPath or a nil cfg.
func (c *Client) ReadWorktreeStatus(ctx context.Context, cfg *core.Config, repoPath, wtPath string) (core.WorktreeStatus, error) {
	if cfg == nil {
		return core.WorktreeStatus{}, &core.OperationError{
			Op:      statusOp,
			Entity:  wtPath,
			Message: "config is required",
		}
	}
	if repoPath == "" {
		return core.WorktreeStatus{}, &core.OperationError{
			Op:      statusOp,
			Message: "repository path is required",
		}
	}
	if wtPath == "" {
		return core.WorktreeStatus{}, &core.OperationError{
			Op:      statusOp,
			Message: "worktree path is required",
		}
	}

	now := time.Now()
	row := core.WorktreeStatus{LastChecked: now}

	branch, err := c.resolveWorktreeBranch(ctx, repoPath, wtPath, &row)
	if err != nil {
		return row, err
	}
	row.Base = c.resolveBaseForWorktree(ctx, wtPath, cfg)

	c.populateDirtyAndAheadBehind(ctx, cfg, wtPath, branch, &row)
	ensureBaseSkip(&row)
	c.populateMerged(ctx, wtPath, branch, &row)
	c.populateLastCommitDate(ctx, repoPath, branch, &row)

	return row, nil
}

// resolveWorktreeBranch reads the worktree's branch from
// `git worktree list --porcelain`.
//
// Discriminator between the two failure modes:
//   - ListWorktrees itself fails: precondition failure (the walk
//     cannot read the worktree's branch). The row is marked skipped
//     with the underlying error in SkipReason and nil is returned so
//     the rest of the walk continues across other worktrees.
//   - The worktree is not in the returned list: precondition failure
//     for this per-call race (the caller iterated worktrees that came
//     from the same list, so this branch only fires under a TOCTOU
//     race). The row is marked skipped and a *core.OperationError
//     wrapping core.ErrWorktreeNotFound is returned. cmd/status
//     discards the error (the row carries the skip fields); cmd/delete
//     honors the sentinel via errors.Is to refuse deletion of a
//     vanished worktree.
//
// Per-wt read failures (ahead/behind via rev-list, IsBranchMerged,
// ListBranches lookup) live in the downstream populate helpers and
// populate the row's IsSkipped + SkipReason without returning an
// error.
func (c *Client) resolveWorktreeBranch(ctx context.Context, repoPath, wtPath string, row *core.WorktreeStatus) (string, error) {
	worktrees, listErr := c.ListWorktrees(ctx, repoPath)
	if listErr != nil {
		row.IsSkipped = true
		row.SkipReason = "failed to list worktrees: " + listErr.Error()
		//nolint:nilerr // ListWorktrees failure marks the row skipped; the other worktrees in the walk still surface.
		return "", nil
	}
	absWt, err := filepath.Abs(wtPath)
	if err != nil {
		absWt = wtPath
	}
	for _, wt := range worktrees {
		if wt.Path == absWt {
			if row.Worktree == nil {
				row.Worktree = &core.Worktree{Path: wt.Path, Branch: wt.Branch}
			} else {
				row.Worktree.Branch = wt.Branch
			}
			return wt.Branch, nil
		}
	}
	row.IsSkipped = true
	if row.SkipReason == "" {
		row.SkipReason = "worktree not found in project worktree list"
	}
	return "", &core.OperationError{
		Op:     statusOp + ".worktree-not-found",
		Entity: wtPath,
		Cause:  core.ErrWorktreeNotFound,
	}
}

// populateDirtyAndAheadBehind fills IsClean, HasUncommittedChanges,
// and the RepositoryStatus snapshot (including Ahead/Behind) by
// combining the go-git RepositoryStatus call with the rev-list CLI
// invocation. The row is marked skipped on either failure but the
// other populated fields survive.
func (c *Client) populateDirtyAndAheadBehind(ctx context.Context, _ *core.Config, wtPath, branch string, row *core.WorktreeStatus) {
	repoStatus, err := c.RepositoryStatus(ctx, wtPath)
	if err != nil {
		row.IsSkipped = true
		row.SkipReason = "failed to read repository status: " + err.Error()
		return
	}
	row.RepositoryStatus = &repoStatus
	row.IsClean = repoStatus.IsClean
	row.HasUncommittedChanges = !repoStatus.IsClean

	if branch == "" {
		// branch lookup failed earlier; we still carry the dirty
		// state but the counts cannot be derived.
		return
	}
	base := row.Base
	if base == "" {
		// No base available; the ensureBaseSkip helper has already
		// (or will) mark the row skipped. Leave counts at zero.
		return
	}
	ahead, behind, err := c.revListCount(ctx, wtPath, base, branch)
	if err != nil {
		row.IsSkipped = true
		row.SkipReason = "failed to compute ahead/behind: " + err.Error()
		return
	}
	repoStatus.Ahead = ahead
	repoStatus.Behind = behind
}

// ensureBaseSkip marks the row skipped when the base resolution
// produced no candidate. Called after populateDirtyAndAheadBehind so
// the ahead/behind fields still populate when the read succeeds
// against an empty base. Preserves any earlier skip reason
// (a per-wt read failure takes precedence over the base-absence
// reason in the diagnostic chain).
func ensureBaseSkip(row *core.WorktreeStatus) {
	if row.Base != "" {
		return
	}
	row.IsSkipped = true
	if row.SkipReason == "" {
		row.SkipReason = "no tracked base and no protected branches"
	}
}

// resolveBaseForWorktree walks the per-worktree tracked-base config
// first, then falls back to the first protected branch. Returns the
// empty string when no source yields a value; never errors.
func (c *Client) resolveBaseForWorktree(ctx context.Context, wtPath string, cfg *core.Config) string {
	tracked, err := c.GetTrackedBase(ctx, wtPath)
	if err == nil && strings.TrimSpace(tracked) != "" {
		return strings.TrimSpace(tracked)
	}
	if len(cfg.Validation.ProtectedBranches) > 0 {
		return cfg.Validation.ProtectedBranches[0]
	}
	return ""
}

// populateMerged calls cliClient.IsBranchMerged and records the
// outcome. IsBranchMerged's first argument is the worktree path
// (not the project path) — the parameter name in the role signature
// is misleading per task 3.4's note; this follows the existing
// call site in cmd/delete.go:209.
func (c *Client) populateMerged(ctx context.Context, wtPath, branch string, row *core.WorktreeStatus) {
	if branch == "" {
		return
	}
	merged, err := c.IsBranchMerged(ctx, wtPath, branch)
	if err != nil {
		row.IsSkipped = true
		row.SkipReason = "failed to check merge status: " + err.Error()
		return
	}
	row.IsMerged = merged
}

// populateLastCommitDate looks up the branch in the go-git branch
// list and copies the matching core.Branch.Date. When the branch is
// not present (or the lookup fails) the zero time stays and the row
// is not marked skipped — a missing last-commit date is not, by
// itself, a per-worktree failure.
func (c *Client) populateLastCommitDate(ctx context.Context, repoPath, branch string, row *core.WorktreeStatus) {
	if branch == "" {
		return
	}
	branches, err := c.ListBranches(ctx, repoPath)
	if err != nil {
		return
	}
	for _, b := range branches {
		if b.Name == branch {
			row.LastCommitDate = b.Date
			return
		}
	}
}

// revListCount runs the two rev-list --count invocations and parses
// the integer stdout. The result is (ahead, behind, nil) on success.
func (c *Client) revListCount(ctx context.Context, wtPath, base, branch string) (int, int, error) {
	ahead, err := c.runRevListCount(ctx, wtPath, base+".."+branch)
	if err != nil {
		return 0, 0, err
	}
	behind, err := c.runRevListCount(ctx, wtPath, branch+".."+base)
	if err != nil {
		return 0, 0, err
	}
	return ahead, behind, nil
}

// runRevListCount runs `git rev-list --count <range>` inside wtPath
// and parses the integer stdout. A non-zero exit code is wrapped as
// a status.worktree OperationError.
func (c *Client) runRevListCount(ctx context.Context, wtPath, revRange string) (int, error) {
	result, err := c.executor.ExecuteWithTimeout(
		ctx,
		wtPath,
		CmdGit,
		c.defaultTimeout,
		"rev-list", "--count", revRange,
	)
	if err != nil {
		if result != nil {
			return 0, &core.OperationError{
				Op:      statusOp,
				Entity:  wtPath,
				Message: fmt.Sprintf("git rev-list --count %s failed: %s", revRange, result.Stderr),
				Cause:   err,
			}
		}
		return 0, &core.OperationError{
			Op:      statusOp,
			Entity:  wtPath,
			Message: fmt.Sprintf("git rev-list --count %s failed: %v", revRange, err),
			Cause:   err,
		}
	}
	if result == nil || result.ExitCode != 0 {
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}
		return 0, &core.OperationError{
			Op:      statusOp,
			Entity:  wtPath,
			Message: fmt.Sprintf("git rev-list --count %s exited non-zero: %s", revRange, stderr),
		}
	}
	parsed, parseErr := strconv.Atoi(strings.TrimSpace(result.Stdout))
	if parseErr != nil {
		return 0, &core.OperationError{
			Op:      statusOp,
			Entity:  wtPath,
			Message: fmt.Sprintf("git rev-list --count %s returned %q: %v", revRange, result.Stdout, parseErr),
		}
	}
	return parsed, nil
}
