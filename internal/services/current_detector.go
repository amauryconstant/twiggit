package services

import (
	"context"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
)

// CurrentDirectoryDetector handles detection of whether current directory is a worktree
type CurrentDirectoryDetector struct {
	gitClient infrastructure.GitClient
}

// NewCurrentDirectoryDetector creates a new CurrentDirectoryDetector instance
func NewCurrentDirectoryDetector(gitClient infrastructure.GitClient) *CurrentDirectoryDetector {
	return &CurrentDirectoryDetector{
		gitClient: gitClient,
	}
}

// Detect detects if the current directory is a worktree and returns worktree info
func (cd *CurrentDirectoryDetector) Detect(ctx context.Context, currentDir string) (*domain.WorktreeInfo, error) {
	if currentDir == "" {
		return nil, domain.NewWorktreeError(
			domain.ErrValidation,
			"current directory path cannot be empty",
			"",
		).WithSuggestion("Provide a valid current directory path")
	}

	repoRoot, err := cd.gitClient.GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		return nil, domain.WrapError(
			domain.ErrNotRepository,
			"failed to get repository root",
			currentDir,
			err,
		).WithSuggestion("Ensure the current directory is within a git repository")
	}

	worktrees, err := cd.gitClient.ListWorktrees(ctx, repoRoot)
	if err != nil {
		return nil, domain.WrapError(
			domain.ErrGitCommand,
			"failed to list worktrees",
			repoRoot,
			err,
		).WithSuggestion("Check that the repository is valid and accessible")
	}

	for _, worktree := range worktrees {
		if worktree.Path == currentDir && worktree.Path != repoRoot {
			return worktree, nil
		}
	}

	return nil, nil
}
