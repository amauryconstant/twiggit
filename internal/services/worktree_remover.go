package services

import (
	"context"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
)

// WorktreeRemover handles worktree removal operations
type WorktreeRemover struct {
	gitClient infrastructure.GitClient
}

// NewWorktreeRemover creates a new WorktreeRemover instance
func NewWorktreeRemover(gitClient infrastructure.GitClient) *WorktreeRemover {
	return &WorktreeRemover{
		gitClient: gitClient,
	}
}

// Remove removes a worktree with safety checks
func (wr *WorktreeRemover) Remove(ctx context.Context, worktreePath string, force bool) error {
	if worktreePath == "" {
		return domain.NewWorktreeError(
			domain.ErrValidation,
			"worktree path cannot be empty",
			"",
		).WithSuggestion("Provide a valid worktree path")
	}

	repoRoot, err := wr.gitClient.GetRepositoryRoot(ctx, worktreePath)
	if err != nil {
		return domain.NewWorktreeError(
			domain.ErrNotRepository,
			"failed to get repository root",
			worktreePath,
			err,
		).WithSuggestion("Ensure the worktree path is a valid git worktree")
	}

	if !force {
		hasChanges := wr.gitClient.HasUncommittedChanges(ctx, worktreePath)
		if hasChanges {
			return domain.NewWorktreeError(
				domain.ErrUncommittedChanges,
				"cannot remove worktree with uncommitted changes",
				worktreePath,
			).WithSuggestion("Commit or stash your changes first, or use --force to override")
		}
	}

	err = wr.gitClient.RemoveWorktree(ctx, repoRoot, worktreePath, force)
	if err != nil {
		return domain.NewWorktreeError(
			domain.ErrGitCommand,
			"failed to remove worktree",
			worktreePath,
			err,
		).WithSuggestion("Check that the worktree exists and is not currently in use")
	}

	return nil
}
