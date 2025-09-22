package services

import (
	"context"
	"fmt"

	"github.com/amaury/twiggit/internal/domain"
)

// WorktreeCreator handles worktree creation operations
type WorktreeCreator struct {
	gitClient  domain.GitClient
	validation *ValidationService
	mise       domain.MiseIntegration
}

// NewWorktreeCreator creates a new WorktreeCreator instance
func NewWorktreeCreator(
	gitClient domain.GitClient,
	validation *ValidationService,
	mise domain.MiseIntegration,
) *WorktreeCreator {
	return &WorktreeCreator{
		gitClient:  gitClient,
		validation: validation,
		mise:       mise,
	}
}

// Create creates a new worktree from specified branch with comprehensive validation
func (wc *WorktreeCreator) Create(ctx context.Context, project, branch, targetPath string) error {
	if project == "" {
		return domain.NewWorktreeError(
			domain.ErrValidation,
			"project path cannot be empty",
			"",
		).WithSuggestion("Provide a valid project path")
	}

	validationResult := wc.validation.ValidateWorktreeCreation(branch, targetPath)
	if !validationResult.Valid {
		return fmt.Errorf("validation failed: %w", validationResult.FirstError())
	}

	isRepo, err := wc.gitClient.IsGitRepository(ctx, project)
	if err != nil {
		return domain.WrapError(
			domain.ErrNotRepository,
			"failed to validate project repository",
			project,
			err,
		).WithSuggestion("Ensure that project path exists and is accessible")
	}

	if !isRepo {
		return domain.NewWorktreeError(
			domain.ErrNotRepository,
			"project is not a git repository",
			project,
		).WithSuggestion("Initialize a git repository in the project directory")
	}

	branchExists := wc.gitClient.BranchExists(ctx, project, branch)

	err = wc.gitClient.CreateWorktree(ctx, project, branch, targetPath)
	if err != nil {
		return domain.WrapError(
			domain.ErrGitCommand,
			"failed to create worktree",
			targetPath,
			err,
		).WithSuggestion("Check that the branch exists or can be created")
	}

	if err := wc.mise.SetupWorktree(project, targetPath); err != nil {
		// Don't fail the entire operation if mise setup fails
		// In a real implementation, we might log this error
		// For now, we continue silently
		_ = err
	}

	if !branchExists {
		// Branch was created as part of worktree creation
		// Could add logging or hooks here
		_ = branchExists
	}

	return nil
}
