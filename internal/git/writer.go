package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/core"
)

// defaultCLITimeout is the per-call timeout for write-side git commands.
// Tunable via the second argument to NewCLIClient.
const defaultCLITimeout = 30 * time.Second

// cliClient implements the write-side git operations via the git CLI
// (CreateWorktree, DeleteWorktree, ListWorktrees, PruneWorktrees,
// DeleteBranch, IsBranchMerged). All worktree-mutating operations go
// through the CLI because go-git does not support them; branch
// mutations are CLI-only for the same reason.
//
// cliClient is an internal collaborator of Client; consumers interact
// with write-side methods through the embedded *Client.
type cliClient struct {
	executor       CommandExecutor
	defaultTimeout time.Duration
}

// CLIClient is the public type alias for the write-side git client
// produced by NewCLIClient. Mirrors the ContextDetector / ContextResolver
// pattern so external callers receive an exported type.
type CLIClient = cliClient

// NewCLIClient creates the write-half of the git Client. The executor
// must be non-nil; timeoutSeconds defaults to 30s when omitted.
//
// In production callers pass the executor returned by NewCommandExecutor.
// Tests inject a MockCommandExecutor to assert call args.
func NewCLIClient(executor CommandExecutor, timeoutSeconds ...int) *CLIClient {
	timeout := defaultCLITimeout
	if len(timeoutSeconds) > 0 {
		timeout = time.Duration(timeoutSeconds[0]) * time.Second
	}

	return &cliClient{
		executor:       executor,
		defaultTimeout: timeout,
	}
}

// parseWorktreeLine parses a single line from git worktree list output
// (pure function, no I/O).
func parseWorktreeLine(line string) *core.WorktreeInfo {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if path, ok := strings.CutPrefix(line, "worktree "); ok {
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		return &core.WorktreeInfo{
			Path: absPath,
		}
	}

	return nil
}

// buildWorktreeAddArgs builds arguments for git worktree add command
// (pure function, no I/O).
func buildWorktreeAddArgs(branchExists bool, branchName, worktreePath, sourceBranch string) []string {
	args := []string{"worktree", "add"}

	if branchExists {
		// Branch already exists, checkout existing branch
		args = append(args, worktreePath, branchName)
	} else if sourceBranch != "" {
		// Branch doesn't exist, create new branch from sourceBranch
		args = append(args, "-b", branchName, worktreePath, sourceBranch)
	} else {
		// Branch doesn't exist and no sourceBranch provided, create from current HEAD
		args = append(args, "-b", branchName, worktreePath)
	}

	return args
}

// buildWorktreeRemoveArgs builds arguments for git worktree remove command
// (pure function, no I/O).
func buildWorktreeRemoveArgs(worktreePath string, force bool) []string {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	return args
}

// CreateWorktree creates new worktree using git CLI (idempotent)
func (c *cliClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	// Validate inputs
	if repoPath == "" {
		return NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}
	if branchName == "" {
		return NewWorktreeError("git.worktree", "branch name cannot be empty", nil)
	}
	if worktreePath == "" {
		return NewWorktreeError("git.worktree", "worktree path cannot be empty", nil)
	}

	// Check if branch already exists
	branchExists, err := c.branchExists(ctx, repoPath, branchName)
	if err != nil {
		return NewWorktreeError("git.worktree", "failed to check if branch exists", err)
	}

	// Build command arguments using pure function
	args := buildWorktreeAddArgs(branchExists, branchName, worktreePath, sourceBranch)

	// Execute command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, args...)
	if err != nil {
		return NewWorktreeError("git.worktree", "failed to create worktree", err)
	}
	if result == nil {
		return NewWorktreeError("git.worktree", "command executor returned nil result for worktree create", nil)
	}

	if result.ExitCode != 0 {
		return NewWorktreeError("git.worktree", "git worktree add failed: "+result.Stderr, nil)
	}

	parentDir := filepath.Dir(worktreePath)
	root, err := os.OpenRoot(parentDir)
	if err != nil {
		return NewWorktreeError("git.worktree", "git worktree add succeeded but parent directory not accessible", err)
	}
	defer root.Close() //nolint:errcheck // read-only filesystem stat cleanup, no actionable error
	if _, err := root.Stat(filepath.Base(worktreePath)); err != nil {
		return NewWorktreeError("git.worktree", "git worktree add succeeded but worktree directory not found", err)
	}

	return nil
}

// DeleteWorktree removes worktree using git CLI (idempotent, no-op if already deleted)
func (c *cliClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	// Validate inputs
	if repoPath == "" {
		return NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}
	if worktreePath == "" {
		return NewWorktreeError("git.worktree", "worktree path cannot be empty", nil)
	}

	// For idempotency, we'll try to delete directly and handle "not found" errors
	// This is more efficient than listing worktrees first

	// Build command arguments using pure function
	args := buildWorktreeRemoveArgs(worktreePath, force)

	// Execute command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, args...)
	if err != nil {
		return NewWorktreeError("git.worktree", "failed to delete worktree", err)
	}
	if result == nil {
		return NewWorktreeError("git.worktree", "command executor returned nil result for worktree delete", nil)
	}

	if result.ExitCode != 0 {
		// Check if worktree was already deleted
		if strings.Contains(result.Stderr, "not found") || strings.Contains(result.Stderr, "does not exist") {
			return nil // No-op if already deleted
		}
		return NewWorktreeError("git.worktree", "git worktree remove failed: "+result.Stderr, nil)
	}

	return nil
}

// ListWorktrees lists all worktrees using git CLI (idempotent)
func (c *cliClient) ListWorktrees(ctx context.Context, repoPath string) ([]core.WorktreeInfo, error) {
	// Validate input
	if repoPath == "" {
		return nil, NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}

	// Execute git worktree list command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, NewWorktreeError("git.worktree", "failed to list worktrees", err)
	}
	if result == nil {
		return nil, NewWorktreeError("git.worktree", "command executor returned nil result for worktree list", nil)
	}

	if result.ExitCode != 0 {
		return nil, NewWorktreeError("git.worktree", "git worktree list failed: "+result.Stderr, nil)
	}

	// Parse output
	return c.parseWorktreeList(result.Stdout)
}

// PruneWorktrees removes stale worktree references
func (c *cliClient) PruneWorktrees(ctx context.Context, repoPath string) error {
	// Validate input
	if repoPath == "" {
		return NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}

	// Execute git worktree prune command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, "worktree", "prune")
	if err != nil {
		return NewWorktreeError("git.worktree", "failed to prune worktrees", err)
	}
	if result == nil {
		return NewWorktreeError("git.worktree", "command executor returned nil result for worktree prune", nil)
	}

	if result.ExitCode != 0 {
		return NewWorktreeError("git.worktree", "git worktree prune failed: "+result.Stderr, nil)
	}

	return nil
}

// DeleteBranch deletes a branch using git CLI (handles worktree-referenced branches)
func (c *cliClient) DeleteBranch(ctx context.Context, repoPath, branchName string) error {
	if repoPath == "" {
		return NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}
	if branchName == "" {
		return NewWorktreeError("git.worktree", "branch name cannot be empty", nil)
	}

	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, "branch", "-D", branchName)
	if err != nil {
		return NewWorktreeError("git.worktree", "failed to delete branch", err)
	}
	if result == nil {
		return NewWorktreeError("git.worktree", "command executor returned nil result for branch delete", nil)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "not found") {
			return nil
		}
		return NewWorktreeError("git.worktree", "git branch -D failed: "+result.Stderr, nil)
	}

	return nil
}

// IsBranchMerged checks if a branch is merged into the current branch
func (c *cliClient) IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	// Validate input
	if repoPath == "" {
		return false, NewWorktreeError("git.worktree", "repository path cannot be empty", nil)
	}
	if branchName == "" {
		return false, NewWorktreeError("git.worktree", "branch name cannot be empty", nil)
	}

	// Execute git branch --merged command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, "branch", "--merged")
	if err != nil {
		return false, NewWorktreeError("git.worktree", "failed to check merged status", err)
	}
	if result == nil {
		return false, NewWorktreeError("git.worktree", "command executor returned nil result for branch merge check", nil)
	}

	if result.ExitCode != 0 {
		return false, NewWorktreeError("git.worktree", "git branch --merged failed: "+result.Stderr, nil)
	}

	// Check if branch name appears in merged branches output
	mergedBranches := strings.SplitSeq(result.Stdout, "\n")
	for branch := range mergedBranches {
		// Remove leading markers: * (current branch), + (branch in another worktree)
		trimmed, _ := strings.CutPrefix(strings.TrimSpace(branch), "*")
		trimmed, _ = strings.CutPrefix(trimmed, "+")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == branchName {
			return true, nil
		}
	}

	return false, nil
}

// parseWorktreeList parses the output of `git worktree list --porcelain`
func (c *cliClient) parseWorktreeList(output string) ([]core.WorktreeInfo, error) {
	var worktrees []core.WorktreeInfo
	var currentWorktree *core.WorktreeInfo

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			// Save previous worktree if exists
			if currentWorktree != nil {
				worktrees = append(worktrees, *currentWorktree)
			}

			// Start new worktree
			absPath, err := filepath.Abs(path)
			if err != nil {
				absPath = path // Use original path if conversion fails
			}

			currentWorktree = &core.WorktreeInfo{
				Path: absPath,
			}
		} else if currentWorktree != nil {
			if commit, ok := strings.CutPrefix(line, "HEAD "); ok {
				currentWorktree.Commit = commit
			} else if branchRef, ok := strings.CutPrefix(line, "branch "); ok {
				// Extract branch name from refs/heads/branch-name
				if name, ok := strings.CutPrefix(branchRef, "refs/heads/"); ok {
					currentWorktree.Branch = name
				} else {
					currentWorktree.Branch = branchRef
				}
				currentWorktree.IsDetached = false
			} else if line == "detached" {
				currentWorktree.IsDetached = true
			}
		}
	}

	// Add last worktree
	if currentWorktree != nil {
		worktrees = append(worktrees, *currentWorktree)
	}

	return worktrees, nil
}

// branchExists checks if a branch exists using git CLI
func (c *cliClient) branchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	// Use git show-ref to check if branch exists
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.defaultTimeout, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	if err != nil {
		// The executor wraps non-zero exit codes as *core.OperationError(Op="git.command").
		// The CommandResult returned alongside err already carries ExitCode / Stdout / Stderr
		// populated by createCommandResult, so we can dispatch on result.ExitCode directly.
		var oe *core.OperationError
		if !errors.As(err, &oe) {
			return false, NewRepoError("git.repository", "failed to check branch existence", err)
		}
	}

	// Exit code 0 means branch exists, exit code 1 means branch doesn't exist
	if result.ExitCode == 0 {
		return true, nil
	} else if result.ExitCode == 1 {
		return false, nil
	}

	// Any other exit code is an error
	return false, NewRepoError("git.repository", fmt.Sprintf("git show-ref exited with unexpected code: %d", result.ExitCode), nil)
}
