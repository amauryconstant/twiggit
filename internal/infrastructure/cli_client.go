package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/domain"
)

// parseWorktreeLine parses a single line from git worktree list output
func parseWorktreeLine(line string) *domain.WorktreeInfo {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if strings.HasPrefix(line, "worktree ") {
		path := strings.TrimPrefix(line, "worktree ")
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path // Use original path if conversion fails
		}
		return &domain.WorktreeInfo{
			Path: absPath,
		}
	}

	return nil
}

// buildWorktreeAddArgs builds arguments for git worktree add command
func buildWorktreeAddArgs(branchExists bool, branchName, worktreePath, sourceBranch string) []string {
	var args []string
	args = append(args, "worktree", "add")

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
func buildWorktreeRemoveArgs(worktreePath string, force bool) []string {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)
	return args
}

// CLIClientImpl implements CLIClient using git CLI commands
type CLIClientImpl struct {
	executor CommandExecutor
	timeout  time.Duration
}

// NewCLIClient creates a new CLIClient implementation
func NewCLIClient(executor CommandExecutor, timeoutSeconds ...int) *CLIClientImpl {
	defaultTimeout := 30 * time.Second
	if len(timeoutSeconds) > 0 {
		defaultTimeout = time.Duration(timeoutSeconds[0]) * time.Second
	}

	return &CLIClientImpl{
		executor: executor,
		timeout:  defaultTimeout,
	}
}

// CreateWorktree creates new worktree using git CLI (idempotent)
func (c *CLIClientImpl) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	// Validate inputs
	if repoPath == "" {
		return domain.NewGitWorktreeError(worktreePath, branchName, "repository path cannot be empty", nil)
	}
	if branchName == "" {
		return domain.NewGitWorktreeError(worktreePath, branchName, "branch name cannot be empty", nil)
	}
	if worktreePath == "" {
		return domain.NewGitWorktreeError(worktreePath, branchName, "worktree path cannot be empty", nil)
	}

	// Check if branch already exists
	branchExists, err := c.branchExists(ctx, repoPath, branchName)
	if err != nil {
		return domain.NewGitWorktreeError(worktreePath, branchName, "failed to check if branch exists", err)
	}

	// Build command arguments using pure function
	args := buildWorktreeAddArgs(branchExists, branchName, worktreePath, sourceBranch)

	// Execute command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, args...)
	if err != nil {
		return domain.NewGitWorktreeError(worktreePath, branchName, "failed to create worktree", err)
	}

	if result.ExitCode != 0 {
		return domain.NewGitWorktreeError(worktreePath, branchName,
			"git worktree add failed: "+result.Stderr, nil)
	}

	// Check if worktree was actually created
	if _, err := os.Stat(worktreePath); err != nil {
		return domain.NewGitWorktreeError(worktreePath, branchName,
			"git worktree add succeeded but worktree directory not found: "+err.Error(), nil)
	}

	return nil
}

// DeleteWorktree removes worktree using git CLI (idempotent, no-op if already deleted)
func (c *CLIClientImpl) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	// Validate inputs
	if repoPath == "" {
		return domain.NewGitWorktreeError(worktreePath, "", "repository path cannot be empty", nil)
	}
	if worktreePath == "" {
		return domain.NewGitWorktreeError(worktreePath, "", "worktree path cannot be empty", nil)
	}

	// For idempotency, we'll try to delete directly and handle "not found" errors
	// This is more efficient than listing worktrees first

	// Build command arguments using pure function
	args := buildWorktreeRemoveArgs(worktreePath, force)

	// Execute command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, args...)
	if err != nil {
		return domain.NewGitWorktreeError(worktreePath, "", "failed to delete worktree", err)
	}

	if result.ExitCode != 0 {
		// Check if worktree was already deleted
		if strings.Contains(result.Stderr, "not found") || strings.Contains(result.Stderr, "does not exist") {
			return nil // No-op if already deleted
		}
		return domain.NewGitWorktreeError(worktreePath, "",
			"git worktree remove failed: "+result.Stderr, nil)
	}

	return nil
}

// ListWorktrees lists all worktrees using git CLI (idempotent)
func (c *CLIClientImpl) ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
	// Validate input
	if repoPath == "" {
		return nil, domain.NewGitWorktreeError("", "", "repository path cannot be empty", nil)
	}

	// Execute git worktree list command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, domain.NewGitWorktreeError("", "", "failed to list worktrees", err)
	}

	if result.ExitCode != 0 {
		return nil, domain.NewGitWorktreeError("", "",
			"git worktree list failed: "+result.Stderr, nil)
	}

	// Parse output
	return c.parseWorktreeList(result.Stdout)
}

// PruneWorktrees removes stale worktree references
func (c *CLIClientImpl) PruneWorktrees(ctx context.Context, repoPath string) error {
	// Validate input
	if repoPath == "" {
		return domain.NewGitWorktreeError("", "", "repository path cannot be empty", nil)
	}

	// Execute git worktree prune command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, "worktree", "prune")
	if err != nil {
		return domain.NewGitWorktreeError("", "", "failed to prune worktrees", err)
	}

	if result.ExitCode != 0 {
		return domain.NewGitWorktreeError("", "",
			"git worktree prune failed: "+result.Stderr, nil)
	}

	return nil
}

// IsBranchMerged checks if a branch is merged into the current branch
func (c *CLIClientImpl) IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	// Validate input
	if repoPath == "" {
		return false, domain.NewGitWorktreeError("", branchName, "repository path cannot be empty", nil)
	}
	if branchName == "" {
		return false, domain.NewGitWorktreeError("", "", "branch name cannot be empty", nil)
	}

	// Execute git branch --merged command
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, "branch", "--merged")
	if err != nil {
		return false, domain.NewGitWorktreeError("", branchName, "failed to check merged status", err)
	}

	if result.ExitCode != 0 {
		return false, domain.NewGitWorktreeError("", branchName,
			"git branch --merged failed: "+result.Stderr, nil)
	}

	// Check if branch name appears in merged branches output
	mergedBranches := strings.Split(result.Stdout, "\n")
	for _, branch := range mergedBranches {
		trimmed := strings.TrimSpace(branch)
		// Remove leading asterisk if present (indicates current branch)
		trimmed = strings.TrimPrefix(trimmed, "*")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == branchName {
			return true, nil
		}
	}

	return false, nil
}

// parseWorktreeList parses the output of `git worktree list --porcelain`
func (c *CLIClientImpl) parseWorktreeList(output string) ([]domain.WorktreeInfo, error) {
	var worktrees []domain.WorktreeInfo
	var currentWorktree *domain.WorktreeInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			// Save previous worktree if exists
			if currentWorktree != nil {
				worktrees = append(worktrees, *currentWorktree)
			}

			// Start new worktree
			path := strings.TrimPrefix(line, "worktree ")
			absPath, err := filepath.Abs(path)
			if err != nil {
				absPath = path // Use original path if conversion fails
			}

			currentWorktree = &domain.WorktreeInfo{
				Path: absPath,
			}
		} else if currentWorktree != nil {
			if strings.HasPrefix(line, "HEAD ") {
				commit := strings.TrimPrefix(line, "HEAD ")
				currentWorktree.Commit = commit
			} else if strings.HasPrefix(line, "branch ") {
				branchRef := strings.TrimPrefix(line, "branch ")
				// Extract branch name from refs/heads/branch-name
				if strings.HasPrefix(branchRef, "refs/heads/") {
					currentWorktree.Branch = strings.TrimPrefix(branchRef, "refs/heads/")
				} else {
					currentWorktree.Branch = branchRef
				}
				currentWorktree.IsDetached = false
			} else if strings.HasPrefix(line, "detached") {
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
func (c *CLIClientImpl) branchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	// Use git show-ref to check if branch exists
	result, err := c.executor.ExecuteWithTimeout(ctx, repoPath, "git", c.timeout, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	if err != nil {
		// Check if it's a GitCommandError (which happens when exit code is non-zero)
		gitCmdErr := &domain.GitCommandError{}
		if errors.As(err, &gitCmdErr) {
			// Use the exit code from the GitCommandError
			result = &CommandResult{
				ExitCode: gitCmdErr.ExitCode,
				Stdout:   gitCmdErr.Stdout,
				Stderr:   gitCmdErr.Stderr,
			}
		} else {
			return false, domain.NewGitRepositoryError(repoPath, "failed to check branch existence", err)
		}
	}

	// Exit code 0 means branch exists, exit code 1 means branch doesn't exist
	if result.ExitCode == 0 {
		return true, nil
	} else if result.ExitCode == 1 {
		return false, nil
	}

	// Any other exit code is an error
	return false, domain.NewGitRepositoryError(repoPath, fmt.Sprintf("git show-ref exited with unexpected code: %d", result.ExitCode), nil)
}
