//go:build !integration

package integration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amaury/twiggit/internal/domain"
)

// TestErrorFormattingIntegration tests error formatting consistency across the system
func TestErrorFormattingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("worktree error formatting consistency", func(t *testing.T) {
		err := domain.NewWorktreeError(domain.ErrInvalidPath, "path contains spaces", "/bad path").
			WithSuggestion("Use underscores instead of spaces").
			WithSuggestion("Remove spaces from path")

		// Test error structure
		assert.Equal(t, domain.ErrInvalidPath, err.Type)
		assert.Equal(t, "path contains spaces", err.Message)
		assert.Equal(t, "/bad path", err.Path)
		assert.Len(t, err.Suggestions, 2)

		// Test error string formatting
		errorStr := err.Error()
		assert.Contains(t, errorStr, "invalid path: path contains spaces")
		assert.Contains(t, errorStr, "/bad path")
		assert.NotContains(t, errorStr, "VALIDATION_001")  // Code not in string representation
		assert.NotContains(t, errorStr, "Use underscores") // Suggestions not in string representation
	})

	t.Run("project error formatting consistency", func(t *testing.T) {
		err := domain.NewProjectError(domain.ErrProjectNotFound, "project does not exist", "/nonexistent").
			WithSuggestion("Check project path").
			WithSuggestion("Verify project was created")

		// Test error structure
		assert.Equal(t, domain.ErrProjectNotFound, err.Type)
		assert.Equal(t, "project does not exist", err.Message)
		assert.Equal(t, "/nonexistent", err.Path)
		assert.Len(t, err.Suggestions, 2)

		// Test error string formatting
		errorStr := err.Error()
		assert.Contains(t, errorStr, "project not found: project does not exist")
		assert.Contains(t, errorStr, "/nonexistent")
	})

	t.Run("workspace error formatting consistency", func(t *testing.T) {
		err := domain.NewWorkspaceError(domain.ErrWorkspaceInvalidConfiguration, "missing required fields").
			WithSuggestion("Check configuration file").
			WithSuggestion("Add missing required fields")

		// Test error structure
		assert.Equal(t, domain.ErrWorkspaceInvalidConfiguration, err.Type)
		assert.Equal(t, "missing required fields", err.Message)
		assert.Len(t, err.Suggestions, 2)

		// Test error string formatting
		errorStr := err.Error()
		assert.Contains(t, errorStr, "workspace configuration invalid: missing required fields")
		assert.NotContains(t, errorStr, "(path:") // No path for workspace errors
	})
}

// TestErrorWrappingIntegration tests error wrapping behavior across different error types
func TestErrorWrappingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("worktree error wrapping chain", func(t *testing.T) {
		original := errors.New("ENOENT: no such file or directory")
		middle := domain.NewWorktreeError(domain.ErrWorktreeNotFound, "worktree directory missing", "/missing", original)
		top := domain.NewWorktreeError(domain.ErrGitCommand, "git worktree operation failed", "/repo", middle)

		// Test wrapping chain
		assert.Equal(t, middle, top.Cause)
		assert.Equal(t, middle, errors.Unwrap(top))
		assert.Equal(t, original, errors.Unwrap(middle))

		// Test suggestions propagate
		top.WithSuggestion("Check worktree exists").
			WithSuggestion("Verify git repository integrity")

		assert.Len(t, top.Suggestions, 2)
	})

	t.Run("project error with system error", func(t *testing.T) {
		systemErr := errors.New("permission denied: open /protected/project: permission denied")
		projectErr := domain.NewProjectError(domain.ErrProjectNotFound, "cannot access project", "/protected/project", systemErr)

		assert.Equal(t, systemErr, projectErr.Cause)
		assert.Equal(t, systemErr, errors.Unwrap(projectErr))
		assert.Contains(t, projectErr.Error(), "project not found: cannot access project")
		assert.Contains(t, projectErr.Error(), "/protected/project")
	})

	t.Run("workspace error with configuration error", func(t *testing.T) {
		configErr := errors.New("toml: line 2: cannot unmarshal invalid into []string")
		workspaceErr := domain.NewWorkspaceError(domain.ErrWorkspaceInvalidConfiguration, "config parsing failed", configErr)

		assert.Equal(t, configErr, workspaceErr.Cause)
		assert.Equal(t, configErr, errors.Unwrap(workspaceErr))
		assert.Contains(t, workspaceErr.Error(), "workspace configuration invalid: config parsing failed")
	})
}

// TestErrorSuggestionIntegration tests suggestion functionality across error types
func TestErrorSuggestionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("worktree error suggestions", func(t *testing.T) {
		err := domain.NewWorktreeError(domain.ErrInvalidBranchName, "branch name contains invalid characters", "").
			WithSuggestion("Use only alphanumeric characters, hyphens, and underscores").
			WithSuggestion("Avoid spaces and special characters").
			WithSuggestion("Start branch name with a letter")

		assert.Len(t, err.Suggestions, 3)
		assert.Contains(t, err.Suggestions[0], "alphanumeric characters")
		assert.Contains(t, err.Suggestions[1], "spaces and special characters")
		assert.Contains(t, err.Suggestions[2], "Start branch name")
	})

	t.Run("project error suggestions", func(t *testing.T) {
		err := domain.NewProjectError(domain.ErrInvalidGitRepoPath, "not a git repository", "/not/repo").
			WithSuggestion("Initialize git repository with 'git init'").
			WithSuggestion("Check if path is correct").
			WithSuggestion("Verify directory contains .git folder")

		assert.Len(t, err.Suggestions, 3)
		assert.Contains(t, err.Suggestions[0], "git init")
		assert.Contains(t, err.Suggestions[1], "path is correct")
		assert.Contains(t, err.Suggestions[2], ".git folder")
	})

	t.Run("workspace error suggestions", func(t *testing.T) {
		err := domain.NewWorkspaceError(domain.ErrWorkspaceProjectNotFound, "project not found in workspace configuration").
			WithSuggestion("Add project to workspace configuration").
			WithSuggestion("Check project name spelling").
			WithSuggestion("Verify project path is correct")

		assert.Len(t, err.Suggestions, 3)
		assert.Contains(t, err.Suggestions[0], "workspace configuration")
		assert.Contains(t, err.Suggestions[1], "spelling")
		assert.Contains(t, err.Suggestions[2], "project path")
	})
}

// TestValidationResultIntegration tests validation result functionality
func TestValidationResultIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("worktree validation result with multiple errors", func(t *testing.T) {
		result := domain.NewValidationResult()

		err1 := domain.NewWorktreeError(domain.ErrInvalidPath, "path invalid", "/bad").
			WithSuggestion("Use valid path")
		err2 := domain.NewWorktreeError(domain.ErrInvalidBranchName, "branch invalid", "").
			WithSuggestion("Use valid branch name")

		result.AddError(err1)
		result.AddError(err2)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 2)
		assert.Equal(t, err1, result.FirstError())

		combinedErr := result.ToError()
		require.Error(t, combinedErr)
		assert.Contains(t, combinedErr.Error(), "invalid path")
	})

	t.Run("project validation result", func(t *testing.T) {
		result := domain.NewProjectValidationResult()

		err := domain.NewProjectError(domain.ErrInvalidProjectName, "name empty", "").
			WithSuggestion("Provide project name")

		result.AddError(err)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.FirstError())
	})

	t.Run("workspace validation result merging", func(t *testing.T) {
		result1 := domain.NewWorkspaceValidationResult()
		err1 := domain.NewWorkspaceError(domain.ErrWorkspaceInvalidPath, "path invalid")
		result1.AddError(err1)

		result2 := domain.NewWorkspaceValidationResult()
		err2 := domain.NewWorkspaceError(domain.ErrWorkspaceProjectNotFound, "project missing")
		result2.AddError(err2)

		merged := result1.Merge(result2)

		assert.Equal(t, 2, merged.GetErrorCount())
		assert.Contains(t, merged.Errors, err1)
		assert.Contains(t, merged.Errors, err2)
		assert.Equal(t, err1, merged.GetFirstError())
	})
}
