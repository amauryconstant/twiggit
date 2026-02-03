package helpers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WorktreeTestHelper provides git worktree operations using git CLI
// Note: We use direct git CLI commands because go-git doesn't fully support worktree operations
type WorktreeTestHelper struct {
	timeout time.Duration
}

// NewWorktreeTestHelper creates a new WorktreeTestHelper
func NewWorktreeTestHelper() *WorktreeTestHelper {
	return &WorktreeTestHelper{
		timeout: 30 * time.Second,
	}
}

// WithTimeout sets the timeout for git operations
func (h *WorktreeTestHelper) WithTimeout(timeout time.Duration) *WorktreeTestHelper {
	h.timeout = timeout
	return h
}

// CreateWorktree creates a new git worktree from the repository
func (h *WorktreeTestHelper) CreateWorktree(repoPath, worktreePath, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, worktreePath)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CreateWorktreeFromSource creates a worktree from a specific source branch
func (h *WorktreeTestHelper) CreateWorktreeFromSource(repoPath, worktreePath, branch, sourceBranch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, worktreePath, sourceBranch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree from source: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CheckoutExistingBranch creates a worktree for an existing branch
func (h *WorktreeTestHelper) CheckoutExistingBranch(repoPath, worktreePath, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", worktreePath, branch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout existing branch as worktree: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RemoveWorktree removes a git worktree
func (h *WorktreeTestHelper) RemoveWorktree(worktreePath string, force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	repoPath := findRepoPath(worktreePath)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Worktree may already be deleted, consider this idempotent
		if strings.Contains(string(output), "not found") || strings.Contains(string(output), "does not exist") {
			return nil
		}
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ListWorktrees returns all worktrees for the repository
func (h *WorktreeTestHelper) ListWorktrees(repoPath string) ([]WorktreeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w\nOutput: %s", err, string(output))
	}

	return h.parseWorktreeList(string(output))
}

// WorktreeInfo represents information about a worktree
type WorktreeInfo struct {
	Path       string
	Commit     string
	Branch     string
	IsDetached bool
}

// parseWorktreeList parses the output of `git worktree list --porcelain`
func (h *WorktreeTestHelper) parseWorktreeList(output string) ([]WorktreeInfo, error) {
	var worktrees []WorktreeInfo
	var currentWorktree *WorktreeInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			if currentWorktree != nil {
				worktrees = append(worktrees, *currentWorktree)
			}

			path := strings.TrimPrefix(line, "worktree ")
			absPath, err := filepath.Abs(path)
			if err != nil {
				absPath = path
			}

			currentWorktree = &WorktreeInfo{
				Path: absPath,
			}
		} else if currentWorktree != nil {
			if strings.HasPrefix(line, "HEAD ") {
				currentWorktree.Commit = strings.TrimPrefix(line, "HEAD ")
			} else if strings.HasPrefix(line, "branch ") {
				branchRef := strings.TrimPrefix(line, "branch ")
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

	if currentWorktree != nil {
		worktrees = append(worktrees, *currentWorktree)
	}

	return worktrees, nil
}

// findRepoPath finds the git repository path from a worktree path
func findRepoPath(worktreePath string) string {
	path := worktreePath
	for {
		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Check if it's a file (worktree) or directory (main repo)
			if info, err := os.Stat(gitPath); err == nil {
				if !info.IsDir() {
					// This is a worktree, read the .git file to find the main repo
					if content, err := os.ReadFile(gitPath); err == nil {
						lines := strings.Split(string(content), "\n")
						for _, line := range lines {
							if strings.HasPrefix(line, "gitdir:") {
								gitdir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
								// The .git file contains path like: gitdir: /path/to/main/.git/worktrees/branch
								// We need to extract the main repo path
								if strings.Contains(gitdir, "/.git/worktrees/") {
									parts := strings.Split(gitdir, "/.git/worktrees/")
									if len(parts) > 0 {
										return parts[0]
									}
								}
							}
						}
					}
				}
			}
			return path
		}

		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	return worktreePath
}
