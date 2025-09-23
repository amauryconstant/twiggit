package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorktreeErrorBehavior characterizes current WorktreeError behavior
// These tests ensure we don't break existing behavior during consolidation
func TestWorktreeErrorBehavior(t *testing.T) {
	t.Run("basic error creation and formatting", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidPath, "test message", "/test/path")

		assert.Equal(t, ErrInvalidPath, err.Type)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, "/test/path", err.Path)
		require.NoError(t, err.Cause)
		assert.Empty(t, err.Suggestions)
		assert.Empty(t, err.Code)

		expectedError := "invalid path: test message (path: /test/path)"
		assert.Equal(t, expectedError, err.Error())
	})

	t.Run("error without path", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidBranchName, "branch name empty", "")

		expectedError := "invalid branch name: branch name empty"
		assert.Equal(t, expectedError, err.Error())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := WrapError(ErrGitCommand, "git failed", "/repo", cause)

		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error with suggestions", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidPath, "path invalid", "/bad")
		err.WithSuggestion("Use absolute path")
		err.WithSuggestion("Check path exists")

		assert.Equal(t, []string{"Use absolute path", "Check path exists"}, err.Suggestions)
	})

	t.Run("error with code", func(t *testing.T) {
		err := NewWorktreeError(ErrValidation, "validation failed", "")
		err.WithCode("VALIDATION_ERROR_001")

		assert.Equal(t, "VALIDATION_ERROR_001", err.Code)
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewWorktreeError(ErrWorktreeNotFound, "not found", "/path")

		assert.True(t, IsWorktreeErrorType(err, ErrWorktreeNotFound))
		assert.False(t, IsWorktreeErrorType(err, ErrInvalidPath))

		// Test with non-WorktreeError
		otherErr := errors.New("some other error")
		assert.False(t, IsWorktreeErrorType(otherErr, ErrWorktreeNotFound))
	})

	t.Run("error type string representations", func(t *testing.T) {
		testCases := []struct {
			errType  DomainErrorType
			expected string
		}{
			{ErrNotRepository, "not a git repository"},
			{ErrCurrentDirectory, "current directory operation error"},
			{ErrUncommittedChanges, "uncommitted changes detected"},
			{ErrWorktreeExists, "worktree already exists"},
			{ErrWorktreeNotFound, "worktree not found"},
			{ErrInvalidBranchName, "invalid branch name"},
			{ErrInvalidPath, "invalid path"},
			{ErrPathNotWritable, "path not writable"},
			{ErrGitCommand, "git command failed"},
			{ErrValidation, "validation error"},
			{ErrUnknown, "unknown error"},
		}

		for _, tc := range testCases {
			t.Run(tc.expected, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.errType.String())
			})
		}
	})
}

// TestProjectErrorBehavior characterizes current ProjectError behavior
func TestProjectErrorBehavior(t *testing.T) {
	t.Run("basic error creation and formatting", func(t *testing.T) {
		err := NewProjectError(ErrInvalidProjectName, "name empty", "")

		assert.Equal(t, ErrInvalidProjectName, err.Type)
		assert.Equal(t, "name empty", err.Message)
		assert.Empty(t, err.Path)
		require.NoError(t, err.Cause)
		assert.Empty(t, err.Suggestions)
		assert.Empty(t, err.Code)

		expectedError := "invalid project name: name empty"
		assert.Equal(t, expectedError, err.Error())
	})

	t.Run("error with path", func(t *testing.T) {
		err := NewProjectError(ErrInvalidGitRepoPath, "invalid repo", "/bad/path")

		expectedError := "invalid git repository path: invalid repo (path: /bad/path)"
		assert.Equal(t, expectedError, err.Error())
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("filesystem error")
		err := WrapProjectError(ErrProjectNotFound, "project missing", "/path", cause)

		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewProjectError(ErrProjectAlreadyExists, "duplicate", "")

		assert.True(t, IsProjectErrorType(err, ErrProjectAlreadyExists))
		assert.False(t, IsProjectErrorType(err, ErrInvalidProjectName))
	})

	t.Run("project error type string representations", func(t *testing.T) {
		testCases := []struct {
			errType  DomainErrorType
			expected string
		}{
			{ErrInvalidProjectName, "invalid project name"},
			{ErrInvalidGitRepoPath, "invalid git repository path"},
			{ErrProjectNotFound, "project not found"},
			{ErrProjectAlreadyExists, "project already exists"},
			{ErrProjectValidation, "project validation error"},
		}

		for _, tc := range testCases {
			t.Run(tc.expected, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.errType.String())
			})
		}
	})
}

// TestWorkspaceErrorBehavior characterizes current WorkspaceError behavior
func TestWorkspaceErrorBehavior(t *testing.T) {
	t.Run("basic error creation", func(t *testing.T) {
		err := NewWorkspaceError(ErrWorkspaceInvalidPath, "path empty")

		assert.Equal(t, ErrWorkspaceInvalidPath, err.Type)
		assert.Equal(t, "path empty", err.Message)
		require.NoError(t, err.Cause)

		expectedError := "workspace path invalid: path empty"
		assert.Equal(t, expectedError, err.Error())
	})

	t.Run("error with underlying error", func(t *testing.T) {
		cause := errors.New("config error")
		err := WrapWorkspaceError(ErrWorkspaceInvalidConfiguration, "config invalid", cause)

		expectedError := "workspace configuration invalid: config invalid"
		assert.Equal(t, expectedError, err.Error())
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewWorkspaceError(ErrWorkspaceProjectNotFound, "project missing")

		assert.True(t, IsWorkspaceErrorType(err, ErrWorkspaceProjectNotFound))
		assert.False(t, IsWorkspaceErrorType(err, ErrWorkspaceInvalidPath))
	})

	t.Run("workspace error type string representations", func(t *testing.T) {
		testCases := []struct {
			errType  DomainErrorType
			expected string
		}{
			{ErrWorkspaceInvalidPath, "workspace path invalid"},
			{ErrWorkspaceProjectNotFound, "workspace project not found"},
			{ErrWorkspaceProjectAlreadyExists, "workspace project already exists"},
			{ErrWorkspaceWorktreeNotFound, "workspace worktree not found"},
			{ErrWorkspaceInvalidConfiguration, "workspace configuration invalid"},
			{ErrWorkspaceDiscoveryFailed, "workspace discovery failed"},
			{ErrWorkspaceValidationFailed, "workspace validation failed"},
		}

		for _, tc := range testCases {
			t.Run(tc.expected, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.errType.String())
			})
		}
	})
}

// TestValidationResultBehavior characterizes current ValidationResult behavior
func TestValidationResultBehavior(t *testing.T) {
	t.Run("initial validation result state", func(t *testing.T) {
		result := NewValidationResult()

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
		assert.False(t, result.HasErrors())
		assert.Nil(t, result.FirstError())
	})

	t.Run("adding errors makes result invalid", func(t *testing.T) {
		result := NewValidationResult()
		err := NewWorktreeError(ErrInvalidPath, "bad path", "/test")

		result.AddError(err)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
		assert.True(t, result.HasErrors())
		assert.Equal(t, err, result.FirstError())
	})

	t.Run("adding warnings doesn't affect validity", func(t *testing.T) {
		result := NewValidationResult()

		result.AddWarning("this is a warning")
		result.AddWarning("another warning")

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Equal(t, []string{"this is a warning", "another warning"}, result.Warnings)
	})

	t.Run("multiple errors", func(t *testing.T) {
		result := NewValidationResult()
		err1 := NewWorktreeError(ErrInvalidPath, "bad path", "/test")
		err2 := NewWorktreeError(ErrInvalidBranchName, "bad branch", "")

		result.AddError(err1)
		result.AddError(err2)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 2)
		assert.Equal(t, err1, result.FirstError()) // First error returned
	})
}

// TestProjectValidationResultBehavior characterizes current ProjectValidationResult behavior
func TestProjectValidationResultBehavior(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		result := NewProjectValidationResult()

		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("adding project errors", func(t *testing.T) {
		result := NewProjectValidationResult()
		err := NewProjectError(ErrInvalidProjectName, "name empty", "")

		result.AddError(err)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
		assert.Equal(t, err, result.FirstError())
	})
}

// TestWorkspaceValidationResultBehavior characterizes current WorkspaceValidationResult behavior
func TestWorkspaceValidationResultBehavior(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		result := NewWorkspaceValidationResult()

		assert.True(t, result.IsValid())
		assert.Empty(t, result.Errors)
		assert.Equal(t, 0, result.GetErrorCount())
		assert.Nil(t, result.GetFirstError())
		// nolint:testifylint // We expect nil here when there are no errors
		require.Nil(t, result.ToError())
	})

	t.Run("adding workspace errors", func(t *testing.T) {
		result := NewWorkspaceValidationResult()
		err := NewWorkspaceError(ErrWorkspaceInvalidPath, "path invalid")

		result.AddError(err)

		assert.False(t, result.IsValid())
		assert.Equal(t, 1, result.GetErrorCount())
		assert.Equal(t, err, result.Errors[0])
		assert.Equal(t, err, result.GetFirstError())
		assert.Equal(t, err, result.ToError())
	})

	t.Run("merging results", func(t *testing.T) {
		result1 := NewWorkspaceValidationResult()
		err1 := NewWorkspaceError(ErrWorkspaceInvalidPath, "bad path")
		result1.AddError(err1)

		result2 := NewWorkspaceValidationResult()
		err2 := NewWorkspaceError(ErrWorkspaceProjectNotFound, "missing project")
		result2.AddError(err2)

		merged := result1.Merge(result2)

		assert.Equal(t, 2, merged.GetErrorCount())
		assert.Contains(t, merged.Errors, err1)
		assert.Contains(t, merged.Errors, err2)
	})
}
