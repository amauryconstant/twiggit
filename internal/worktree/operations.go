package worktree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/mise"
	"github.com/amaury/twiggit/pkg/types"
)

// OperationsService handles worktree creation, removal, and management operations
type OperationsService struct {
	gitClient types.GitClient
	discovery *DiscoveryService
	config    *config.Config
	mise      *mise.MiseIntegration
}

// NewOperationsService creates a new OperationsService instance
func NewOperationsService(gitClient types.GitClient, discovery *DiscoveryService, config *config.Config) *OperationsService {
	return &OperationsService{
		gitClient: gitClient,
		discovery: discovery,
		config:    config,
		mise:      mise.NewMiseIntegration(),
	}
}

// Create creates a new worktree from the specified branch with comprehensive validation
func (ops *OperationsService) Create(project, branch, targetPath string) error {
	// Validate inputs
	if project == "" {
		return types.NewWorktreeError(
			types.ErrValidation,
			"project path cannot be empty",
			"",
		).WithSuggestion("Provide a valid project path")
	}

	// Validate branch name and target path
	validationResult := types.ValidateWorktreeCreation(branch, targetPath)
	if !validationResult.Valid {
		return validationResult.FirstError()
	}

	// Check if project is a valid git repository
	isRepo, err := ops.gitClient.IsGitRepository(project)
	if err != nil {
		return types.WrapError(
			types.ErrNotRepository,
			"failed to validate project repository",
			project,
			err,
		).WithSuggestion("Ensure the project path exists and is accessible")
	}

	if !isRepo {
		return types.NewWorktreeError(
			types.ErrNotRepository,
			"project is not a git repository",
			project,
		).WithSuggestion("Initialize a git repository in the project directory")
	}

	// Check if branch exists (for logging purposes later)
	branchExists := ops.gitClient.BranchExists(project, branch)

	// Create the worktree
	err = ops.gitClient.CreateWorktree(project, branch, targetPath)
	if err != nil {
		return types.WrapError(
			types.ErrGitCommand,
			"failed to create worktree",
			targetPath,
			err,
		).WithSuggestion("Check that the branch exists or can be created")
	}

	// Setup mise development environment if available
	if err := ops.mise.SetupWorktree(project, targetPath); err != nil {
		// Don't fail the entire operation if mise setup fails
		// In a real implementation, we might log this error
		// For now, we continue silently
		_ = err
	}

	// Log creation for potential integrations
	if !branchExists {
		// Branch was created as part of worktree creation
		// Could add logging or hooks here
		_ = branchExists
	}

	return nil
}

// Remove removes a worktree with safety checks
func (ops *OperationsService) Remove(worktreePath string, force bool) error {
	// Basic validation
	if worktreePath == "" {
		return types.NewWorktreeError(
			types.ErrValidation,
			"worktree path cannot be empty",
			"",
		).WithSuggestion("Provide a valid worktree path")
	}

	// Validate removal safety if not forced
	if !force {
		if err := ops.ValidateRemoval(worktreePath); err != nil {
			return err
		}
	}

	// Get repository root to perform removal
	repoRoot, err := ops.gitClient.GetRepositoryRoot(worktreePath)
	if err != nil {
		return types.WrapError(
			types.ErrWorktreeNotFound,
			"failed to find repository root for worktree",
			worktreePath,
			err,
		).WithSuggestion("Ensure the worktree path is valid and accessible")
	}

	// Remove the worktree
	err = ops.gitClient.RemoveWorktree(repoRoot, worktreePath, force)
	if err != nil {
		return types.WrapError(
			types.ErrGitCommand,
			"failed to remove worktree",
			worktreePath,
			err,
		).WithSuggestion("Try using --force flag if the worktree has uncommitted changes")
	}

	return nil
}

// GetCurrent returns information about the current worktree (if any)
func (ops *OperationsService) GetCurrent() (*domain.Worktree, error) {
	// Get current working directory
	currentDir, err := ops.getCurrentWorkingDirectory()
	if err != nil {
		return nil, types.WrapError(
			types.ErrValidation,
			"failed to get current directory",
			"",
			err,
		).WithSuggestion("Ensure you have permission to access the current directory")
	}

	// Analyze the current directory as a worktree
	worktree, err := ops.discovery.AnalyzeWorktree(currentDir)
	if err != nil {
		return nil, types.WrapError(
			types.ErrWorktreeNotFound,
			"current directory is not a git worktree",
			currentDir,
			err,
		).WithSuggestion("Navigate to a git worktree directory")
	}

	return worktree, nil
}

// ValidateRemoval performs safety checks before worktree removal
func (ops *OperationsService) ValidateRemoval(worktreePath string) error {
	// Check if trying to remove current directory
	if ops.isCurrentDirectory(worktreePath) {
		return types.NewWorktreeError(
			types.ErrCurrentDirectory,
			"cannot remove current working directory",
			worktreePath,
		).WithSuggestion("Change to a different directory before removing this worktree")
	}

	// Check if worktree exists and get status
	_, err := ops.gitClient.GetWorktreeStatus(worktreePath)
	if err != nil {
		return types.WrapError(
			types.ErrWorktreeNotFound,
			"failed to get worktree status",
			worktreePath,
			err,
		).WithSuggestion("Ensure the worktree path exists and is valid")
	}

	// Check for uncommitted changes
	hasChanges := ops.gitClient.HasUncommittedChanges(worktreePath)
	if hasChanges {
		return types.NewWorktreeError(
			types.ErrUncommittedChanges,
			"worktree has uncommitted changes",
			worktreePath,
		).WithSuggestion("Commit or stash changes, or use --force flag to remove anyway")
	}

	return nil
}

// isCurrentDirectory checks if the given path is the current working directory
func (ops *OperationsService) isCurrentDirectory(worktreePath string) bool {
	currentDir, err := ops.getCurrentWorkingDirectory()
	if err != nil {
		return false
	}

	// Resolve both paths to absolute paths for comparison
	absWorktreePath, err1 := filepath.Abs(worktreePath)
	absCurrentDir, err2 := filepath.Abs(currentDir)

	return err1 == nil && err2 == nil && absWorktreePath == absCurrentDir
}

// getCurrentWorkingDirectory gets the current working directory with error handling
func (ops *OperationsService) getCurrentWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}
