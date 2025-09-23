package services

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
)

// WorktreeCreator handles worktree creation operations
type WorktreeCreator struct {
	gitClient  infrastructure.GitClient
	validation *ValidationService
	mise       infrastructure.MiseIntegration
}

// NewWorktreeCreator creates a new WorktreeCreator instance
func NewWorktreeCreator(
	gitClient infrastructure.GitClient,
	validation *ValidationService,
	mise infrastructure.MiseIntegration,
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
		return validationResult.FirstError()
	}

	isRepo, err := wc.gitClient.IsGitRepository(ctx, project)
	if err != nil {
		return domain.NewWorktreeError(
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
		return domain.NewWorktreeError(
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

// CreateWithFallback creates a worktree with fallback path resolution
// This method provides enhanced error recovery when primary path resolution fails
func (wc *WorktreeCreator) CreateWithFallback(ctx context.Context, project, branch, targetPath string) error {
	// Try primary creation first
	err := wc.Create(ctx, project, branch, targetPath)
	if err == nil {
		return nil
	}

	// Check if it's a path-related error that might benefit from fallback
	if domain.IsDomainErrorType(err, domain.ErrInvalidPath) ||
		domain.IsDomainErrorType(err, domain.ErrPathNotWritable) {
		// Try alternative path resolution
		fallbackPath, fallbackErr := wc.resolvePathWithFallback(project, branch, targetPath)
		if fallbackErr != nil {
			// Return original error if fallback also fails
			return err
		}

		// Try creation with fallback path
		return wc.Create(ctx, project, branch, fallbackPath)
	}

	// Return original error for non-path-related issues
	return err
}

// resolvePathWithFallback provides alternative path resolution when primary method fails
func (wc *WorktreeCreator) resolvePathWithFallback(project, branch, originalPath string) (string, error) {
	// Extract project name for alternative path construction
	projectName := filepath.Base(project)

	// Try alternative path patterns
	alternativePaths := []string{
		// Try with different branch name sanitization
		filepath.Join(filepath.Dir(originalPath), strings.ReplaceAll(branch, "/", "-")),
		// Try with project name prefix
		filepath.Join(filepath.Dir(originalPath), projectName+"-"+branch),
		// Try simple branch name without path
		filepath.Join(filepath.Dir(originalPath), filepath.Base(branch)),
	}

	for _, altPath := range alternativePaths {
		// Check if alternative path is valid and doesn't exist
		if wc.validation.pathExists(altPath) {
			continue // Path already exists, skip
		}

		// Check if parent directory is writable
		parentDir := filepath.Dir(altPath)
		if wc.validation.pathExists(parentDir) && wc.validation.pathWritable(altPath) {
			return altPath, nil
		}
	}

	return "", domain.NewWorktreeError(
		domain.ErrInvalidPath,
		"unable to resolve valid worktree path with fallback",
		originalPath,
	).WithSuggestion("Check workspace configuration and permissions")
}
