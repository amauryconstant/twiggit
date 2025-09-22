package services

import (
	"path/filepath"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
)

// ValidationService handles validation operations that require infrastructure access
type ValidationService struct {
	infraService infrastructure.InfrastructureService
}

// NewValidationService creates a new ValidationService instance
func NewValidationService(infraService infrastructure.InfrastructureService) *ValidationService {
	return &ValidationService{
		infraService: infraService,
	}
}

// ValidatePathWritable checks if a path is writable using the infrastructure service
func (s *ValidationService) ValidatePathWritable(path string) *domain.ValidationResult {
	result := domain.NewValidationResult()

	// First validate the path format
	pathResult := domain.ValidatePath(path)
	if !pathResult.Valid {
		result.Errors = append(result.Errors, pathResult.Errors...)
		result.Valid = false
		return result
	}

	// Check if path already exists
	if s.infraService.PathExists(path) {
		result.AddError(domain.NewWorktreeError(
			domain.ErrPathNotWritable,
			"path already exists",
			path,
		).WithSuggestion("Choose a different path that doesn't already exist"))
		return result
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(path)
	if !s.infraService.PathExists(parentDir) {
		result.AddError(domain.NewWorktreeError(
			domain.ErrPathNotWritable,
			"parent directory does not exist",
			path,
		).WithSuggestion("Create the parent directory: " + parentDir))
		return result
	}

	if !s.infraService.PathWritable(path) {
		result.AddError(domain.NewWorktreeError(
			domain.ErrPathNotWritable,
			"parent directory is not writable",
			path,
		).WithSuggestion("Ensure you have write permissions to the parent directory"))
		return result
	}

	return result
}

// ValidateWorktreeCreation performs comprehensive validation for worktree creation
func (s *ValidationService) ValidateWorktreeCreation(branchName, targetPath string) *domain.ValidationResult {
	result := domain.NewValidationResult()

	// Validate branch name (domain validation only)
	branchResult := domain.ValidateBranchName(branchName)
	result.Errors = append(result.Errors, branchResult.Errors...)
	result.Warnings = append(result.Warnings, branchResult.Warnings...)

	// Only validate path if branch name is valid (to avoid unnecessary infrastructure calls)
	if branchResult.Valid {
		// Validate target path (includes infrastructure checks)
		pathResult := s.ValidatePathWritable(targetPath)
		result.Errors = append(result.Errors, pathResult.Errors...)
		result.Warnings = append(result.Warnings, pathResult.Warnings...)
	}

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}
