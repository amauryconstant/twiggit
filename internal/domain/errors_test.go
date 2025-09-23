package domain

import (
	"errors"
	"fmt"
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
		err := NewWorktreeError(ErrGitCommand, "git failed", "/repo", cause)

		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error with suggestions", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidPath, "path invalid", "/bad")
		err.WithSuggestion("Use absolute path")
		err.WithSuggestion("Check path exists")

		assert.Equal(t, []string{"Use absolute path", "Check path exists"}, err.Suggestions)
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewWorktreeError(ErrWorktreeNotFound, "not found", "/path")

		assert.True(t, IsDomainErrorType(err, ErrWorktreeNotFound))
		assert.False(t, IsDomainErrorType(err, ErrInvalidPath))

		// Test with non-DomainError
		otherErr := errors.New("some other error")
		assert.False(t, IsDomainErrorType(otherErr, ErrWorktreeNotFound))
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
		err := NewProjectError(ErrProjectNotFound, "project missing", "/path", cause)

		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewProjectError(ErrProjectAlreadyExists, "duplicate", "")

		assert.True(t, IsDomainErrorType(err, ErrProjectAlreadyExists))
		assert.False(t, IsDomainErrorType(err, ErrInvalidProjectName))
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
		err := NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "config invalid", cause)

		expectedError := "workspace configuration invalid: config invalid"
		assert.Equal(t, expectedError, err.Error())
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error type checking", func(t *testing.T) {
		err := NewWorkspaceError(ErrWorkspaceProjectNotFound, "project missing")

		assert.True(t, IsDomainErrorType(err, ErrWorkspaceProjectNotFound))
		assert.False(t, IsDomainErrorType(err, ErrWorkspaceInvalidPath))
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

// TestErrorWrappingBehavior tests error wrapping and unwrapping functionality
func TestErrorWrappingBehavior(t *testing.T) {
	t.Run("worktree error with underlying cause", func(t *testing.T) {
		underlying := errors.New("filesystem permission denied")
		err := NewWorktreeError(ErrPathNotWritable, "cannot create worktree", "/protected", underlying)

		assert.Equal(t, underlying, err.Cause)
		assert.Equal(t, underlying, errors.Unwrap(err))
		assert.Contains(t, err.Error(), "path not writable: cannot create worktree")
		assert.Contains(t, err.Error(), "/protected")
	})

	t.Run("project error with underlying cause", func(t *testing.T) {
		underlying := errors.New("network connection failed")
		err := NewProjectError(ErrProjectNotFound, "cannot access project", "/remote", underlying)

		assert.Equal(t, underlying, err.Cause)
		assert.Equal(t, underlying, errors.Unwrap(err))
		assert.Contains(t, err.Error(), "project not found: cannot access project")
		assert.Contains(t, err.Error(), "/remote")
	})

	t.Run("workspace error with underlying cause", func(t *testing.T) {
		underlying := errors.New("yaml parsing error")
		err := NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "config parse failed", underlying)

		assert.Equal(t, underlying, err.Cause)
		assert.Equal(t, underlying, errors.Unwrap(err))
		assert.Contains(t, err.Error(), "workspace configuration invalid: config parse failed")
	})

	t.Run("nested error wrapping", func(t *testing.T) {
		original := errors.New("disk full")
		middle := NewWorktreeError(ErrPathNotWritable, "cannot write file", "/tmp", original)
		top := NewWorktreeError(ErrGitCommand, "git operation failed", "/repo", middle)

		assert.Equal(t, middle, top.Cause)
		assert.Equal(t, middle, errors.Unwrap(top))
		assert.Equal(t, original, errors.Unwrap(middle))
	})

	t.Run("error wrapping with suggestions", func(t *testing.T) {
		underlying := errors.New("branch does not exist")
		err := NewWorktreeError(ErrInvalidBranchName, "invalid branch reference", "/repo", underlying).
			WithSuggestion("Check branch name spelling").
			WithSuggestion("Verify branch exists")

		assert.Equal(t, underlying, err.Cause)
		assert.Len(t, err.Suggestions, 2)
		assert.Contains(t, err.Suggestions[0], "Check branch name spelling")
		assert.Contains(t, err.Suggestions[1], "Verify branch exists")
	})
}

// TestErrorFormattingAndPresentation tests how errors are formatted for display
func TestErrorFormattingAndPresentation(t *testing.T) {
	t.Run("worktree error with all fields", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidPath, "path contains invalid characters", "/bad/path").
			WithSuggestion("Use only alphanumeric characters").
			WithSuggestion("Avoid special characters")

		errorStr := err.Error()
		assert.Contains(t, errorStr, "invalid path: path contains invalid characters")
		assert.Contains(t, errorStr, "/bad/path")
		// Note: Suggestions are not included in Error() string, but are available for programmatic use
	})

	t.Run("project error without path", func(t *testing.T) {
		err := NewProjectError(ErrInvalidProjectName, "project name cannot be empty", "")

		errorStr := err.Error()
		assert.Contains(t, errorStr, "invalid project name: project name cannot be empty")
		assert.NotContains(t, errorStr, "(path:")
	})

	t.Run("workspace error with minimal information", func(t *testing.T) {
		err := NewWorkspaceError(ErrWorkspaceProjectNotFound, "project not found in workspace")

		errorStr := err.Error()
		assert.Contains(t, errorStr, "workspace project not found: project not found in workspace")
		assert.NotContains(t, errorStr, "(path:")
	})

	t.Run("error with cause chain formatting", func(t *testing.T) {
		cause := errors.New("ENOENT: no such file or directory")
		err := NewWorktreeError(ErrWorktreeNotFound, "worktree directory missing", "/missing", cause)

		errorStr := err.Error()
		assert.Contains(t, errorStr, "worktree not found: worktree directory missing")
		assert.Contains(t, errorStr, "/missing")
		// The underlying cause is available via errors.Unwrap() but not in the Error() string
	})

	t.Run("validation result error formatting", func(t *testing.T) {
		result := NewValidationResult()
		err1 := NewWorktreeError(ErrInvalidPath, "bad path", "/test1")
		err2 := NewWorktreeError(ErrInvalidBranchName, "bad branch", "")

		result.AddError(err1)
		result.AddError(err2)

		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 2)
		assert.Equal(t, err1, result.FirstError())

		// Test ToError() method
		combinedErr := result.ToError()
		require.Error(t, combinedErr)
		assert.Contains(t, combinedErr.Error(), "invalid path")
	})
}

// TestErrorSuggestions tests suggestion functionality
func TestErrorSuggestions(t *testing.T) {
	t.Run("adding multiple suggestions", func(t *testing.T) {
		err := NewWorktreeError(ErrInvalidPath, "path invalid", "/bad")
		err.WithSuggestion("Use absolute path")
		err.WithSuggestion("Check path exists")
		err.WithSuggestion("Verify permissions")

		assert.Len(t, err.Suggestions, 3)
		assert.Equal(t, "Use absolute path", err.Suggestions[0])
		assert.Equal(t, "Check path exists", err.Suggestions[1])
		assert.Equal(t, "Verify permissions", err.Suggestions[2])
	})

	t.Run("suggestions with error codes", func(t *testing.T) {
		err := NewProjectError(ErrInvalidProjectName, "name invalid", "").
			WithSuggestion("Use lowercase letters only").
			WithSuggestion("Avoid special characters")

		assert.Len(t, err.Suggestions, 2)
		assert.Contains(t, err.Suggestions[0], "lowercase letters")
		assert.Contains(t, err.Suggestions[1], "special characters")
	})

	t.Run("suggestions with wrapped errors", func(t *testing.T) {
		cause := errors.New("git branch not found")
		err := NewWorktreeError(ErrInvalidBranchName, "branch reference invalid", "/repo", cause).
			WithSuggestion("Check git branch exists").
			WithSuggestion("Use 'git branch' to list branches")

		assert.Equal(t, cause, err.Cause)
		assert.Len(t, err.Suggestions, 2)
		assert.Contains(t, err.Suggestions[0], "git branch exists")
		assert.Contains(t, err.Suggestions[1], "git branch")
	})
}

// TestErrorTypeChecking tests type checking functionality
func TestErrorTypeChecking(t *testing.T) {
	t.Run("worktree error type checking", func(t *testing.T) {
		err := NewWorktreeError(ErrWorktreeNotFound, "missing", "/path")

		assert.True(t, IsDomainErrorType(err, ErrWorktreeNotFound))
		assert.False(t, IsDomainErrorType(err, ErrInvalidPath))
		assert.False(t, IsDomainErrorType(err, ErrProjectNotFound))
	})

	t.Run("project error type checking", func(t *testing.T) {
		err := NewProjectError(ErrProjectAlreadyExists, "duplicate", "")

		assert.True(t, IsDomainErrorType(err, ErrProjectAlreadyExists))
		assert.False(t, IsDomainErrorType(err, ErrInvalidProjectName))
		assert.False(t, IsDomainErrorType(err, ErrWorktreeExists))
	})

	t.Run("workspace error type checking", func(t *testing.T) {
		err := NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "config bad")

		assert.True(t, IsDomainErrorType(err, ErrWorkspaceInvalidConfiguration))
		assert.False(t, IsDomainErrorType(err, ErrWorkspaceInvalidPath))
		assert.False(t, IsDomainErrorType(err, ErrGitCommand))
	})

	t.Run("type checking with non-domain errors", func(t *testing.T) {
		standardErr := errors.New("standard error")
		customErr := fmt.Errorf("custom error: %w", standardErr)

		assert.False(t, IsDomainErrorType(standardErr, ErrWorktreeNotFound))
		assert.False(t, IsDomainErrorType(customErr, ErrProjectNotFound))
		assert.False(t, IsDomainErrorType(standardErr, ErrWorkspaceInvalidPath))
	})

	t.Run("type checking with nil errors", func(t *testing.T) {
		assert.False(t, IsDomainErrorType(nil, ErrWorktreeNotFound))
		assert.False(t, IsDomainErrorType(nil, ErrProjectNotFound))
		assert.False(t, IsDomainErrorType(nil, ErrWorkspaceInvalidPath))
	})
}
