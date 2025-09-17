package types

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// MaxBranchNameLength is the maximum allowed length for branch names
	MaxBranchNameLength = 250
	// MinBranchNameLength is the minimum allowed length for branch names
	MinBranchNameLength = 1
)

var (
	// validBranchNameRegex matches valid git branch names
	// Based on git-check-ref-format rules
	validBranchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

	// invalidBranchChars contains characters not allowed in branch names
	invalidBranchChars = []string{" ", "\t", "\n", "\r", "~", "^", ":", "?", "*", "[", "\\", "..", "//"}

	// maxPathLength is the maximum allowed filesystem path length
	maxPathLength = 4096
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid    bool
	Errors   []*WorktreeError
	Warnings []string
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(err *WorktreeError) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, err)
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// FirstError returns the first validation error, or nil if none
func (vr *ValidationResult) FirstError() error {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return nil
}

// NewValidationResult creates a new ValidationResult
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   make([]*WorktreeError, 0),
		Warnings: make([]string, 0),
	}
}

// ValidateBranchName validates a Git branch name according to git naming rules
func ValidateBranchName(branchName string) *ValidationResult {
	result := NewValidationResult()

	if branchName == "" {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name cannot be empty",
			"",
		).WithSuggestion("Provide a valid branch name"))
		return result
	}

	// Check length
	if len(branchName) < MinBranchNameLength {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			fmt.Sprintf("branch name too short (minimum %d characters)", MinBranchNameLength),
			"",
		).WithSuggestion("Use a longer branch name"))
	}

	if len(branchName) > MaxBranchNameLength {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			fmt.Sprintf("branch name too long (maximum %d characters)", MaxBranchNameLength),
			"",
		).WithSuggestion("Use a shorter branch name"))
	}

	// Check for invalid characters
	for _, invalidChar := range invalidBranchChars {
		if strings.Contains(branchName, invalidChar) {
			result.AddError(NewWorktreeError(
				ErrInvalidBranchName,
				fmt.Sprintf("branch name contains invalid character: '%s'", invalidChar),
				"",
			).WithSuggestion("Remove invalid characters from branch name"))
		}
	}

	// Check for reserved names
	if strings.HasPrefix(branchName, "-") {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name cannot start with a hyphen",
			"",
		).WithSuggestion("Choose a branch name that doesn't start with '-'"))
	}

	if strings.HasSuffix(branchName, ".lock") {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name cannot end with '.lock'",
			"",
		).WithSuggestion("Choose a branch name that doesn't end with '.lock'"))
	}

	// Check UTF-8 validity
	if !utf8.ValidString(branchName) {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name contains invalid UTF-8 characters",
			"",
		).WithSuggestion("Use only valid UTF-8 characters in branch name"))
	}

	// Check against regex pattern (if no other errors)
	if result.Valid && !validBranchNameRegex.MatchString(branchName) {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name format is invalid",
			"",
		).WithSuggestion("Use alphanumeric characters, dots, slashes, and hyphens only"))
	}

	return result
}

// ValidatePath validates a filesystem path for worktree operations
func ValidatePath(path string) *ValidationResult {
	result := NewValidationResult()

	if path == "" {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path cannot be empty",
			"",
		).WithSuggestion("Provide a valid filesystem path"))
		return result
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path must be absolute",
			path,
		).WithSuggestion("Use an absolute path starting with '/'"))
	}

	// Check path length (filesystem dependent, but 4096 is a safe limit)
	if len(path) > maxPathLength {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			fmt.Sprintf("path too long (maximum %d characters)", maxPathLength),
			path,
		).WithSuggestion("Use a shorter path"))
	}

	// Check for invalid characters in path
	if strings.Contains(path, "\x00") {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path contains null character",
			path,
		).WithSuggestion("Remove null characters from path"))
	}

	// Check UTF-8 validity
	if !utf8.ValidString(path) {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path contains invalid UTF-8 characters",
			path,
		).WithSuggestion("Use only valid UTF-8 characters in path"))
	}

	return result
}

// ValidatePathWritable checks if a path is writable (parent directory must exist and be writable)
func ValidatePathWritable(path string) *ValidationResult {
	result := NewValidationResult()

	// First validate the path format
	pathResult := ValidatePath(path)
	if !pathResult.Valid {
		result.Errors = append(result.Errors, pathResult.Errors...)
		result.Valid = false
		return result
	}

	// Check if path already exists
	if _, err := os.Stat(path); err == nil {
		result.AddError(NewWorktreeError(
			ErrPathNotWritable,
			"path already exists",
			path,
		).WithSuggestion("Choose a different path that doesn't already exist"))
		return result
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(path)
	parentInfo, err := os.Stat(parentDir)
	if os.IsNotExist(err) {
		result.AddError(NewWorktreeError(
			ErrPathNotWritable,
			"parent directory does not exist",
			path,
		).WithSuggestion(fmt.Sprintf("Create the parent directory: %s", parentDir)))
		return result
	}

	if err != nil {
		result.AddError(WrapError(
			ErrPathNotWritable,
			"cannot access parent directory",
			path,
			err,
		).WithSuggestion("Ensure the parent directory is accessible"))
		return result
	}

	if !parentInfo.IsDir() {
		result.AddError(NewWorktreeError(
			ErrPathNotWritable,
			"parent path is not a directory",
			path,
		).WithSuggestion("Ensure the parent path is a directory"))
		return result
	}

	// Test writability by attempting to create a temporary file
	tempFile := filepath.Join(parentDir, ".twiggit-write-test")
	file, err := os.Create(tempFile)
	if err != nil {
		result.AddError(WrapError(
			ErrPathNotWritable,
			"parent directory is not writable",
			path,
			err,
		).WithSuggestion("Ensure you have write permissions to the parent directory"))
		return result
	}

	// Clean up the temporary file
	_ = file.Close()
	_ = os.Remove(tempFile)

	return result
}

// ValidateWorktreeCreation performs comprehensive validation for worktree creation
func ValidateWorktreeCreation(branchName, targetPath string) *ValidationResult {
	result := NewValidationResult()

	// Validate branch name
	branchResult := ValidateBranchName(branchName)
	result.Errors = append(result.Errors, branchResult.Errors...)
	result.Warnings = append(result.Warnings, branchResult.Warnings...)

	// Validate target path
	pathResult := ValidatePathWritable(targetPath)
	result.Errors = append(result.Errors, pathResult.Errors...)
	result.Warnings = append(result.Warnings, pathResult.Warnings...)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}
