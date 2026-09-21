package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error_WithoutSuggestions(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	msg := err.Error()

	assert.Contains(t, msg, "validation failed")
	assert.Contains(t, msg, "CreateWorktree.branch")
	assert.Contains(t, msg, "cannot be empty")
	assert.Contains(t, msg, "value:")
	assert.NotContains(t, msg, "💡")
}

func TestValidationError_Suggestions(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty").
		WithSuggestions([]string{
			"Use a valid branch name",
			"Branch names should not be empty",
		})

	assert.Equal(t, []string{"Use a valid branch name", "Branch names should not be empty"}, err.Suggestions())

	msg := err.Error()
	assert.Contains(t, msg, "validation failed")
	assert.Contains(t, msg, "CreateWorktree.branch")
	assert.Contains(t, msg, "cannot be empty")
	assert.NotContains(t, msg, "💡 Use a valid branch name")
}

func TestValidationError_Error_WithEmptySuggestions(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty").
		WithSuggestions([]string{})
	msg := err.Error()

	assert.Contains(t, msg, "validation failed")
	assert.NotContains(t, msg, "💡")
}

func TestValidationError_WithSuggestions_Immutability(t *testing.T) {
	original := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	suggestions := []string{
		"Use a valid branch name",
		"Branch names should not be empty",
	}

	modified := original.WithSuggestions(suggestions)

	assert.NotEqual(t, original, modified)
	assert.Empty(t, original.Suggestions())
	assert.Equal(t, suggestions, modified.Suggestions())

	ModifySuggestionsAndCheck(t, original, suggestions)
}

func ModifySuggestionsAndCheck(t *testing.T, original *ValidationError, suggestions []string) {
	t.Helper()

	suggestions[0] = "Modified suggestion"
	assert.Empty(t, original.Suggestions())
	assert.Contains(t, suggestions, "Modified suggestion")
}

func TestValidationError_WithContext_Immutability(t *testing.T) {
	original := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	modified := original.WithContext("Additional context information")

	assert.NotEqual(t, original, modified)
	assert.Empty(t, original.Context())
	assert.Equal(t, "Additional context information", modified.Context())
}

func TestValidationError_WithSuggestionsThenWithContext(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "feat/test", "invalid format").
		WithSuggestions([]string{"Use kebab-case for branch names"}).
		WithContext("Branch name validation")

	assert.Equal(t, []string{"Use kebab-case for branch names"}, err.Suggestions())
	assert.Equal(t, "Branch name validation", err.Context())
	msg := err.Error()
	assert.Contains(t, msg, "validation failed")
}

func TestValidationError_Getters(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "test-branch", "invalid format").
		WithSuggestions([]string{"suggestion 1", "suggestion 2"}).
		WithContext("test context")

	assert.Equal(t, "branch", err.Field())
	assert.Equal(t, "test-branch", err.Value())
	assert.Equal(t, "invalid format", err.Message())
	assert.Equal(t, "CreateWorktree", err.Request())
	assert.Equal(t, []string{"suggestion 1", "suggestion 2"}, err.Suggestions())
	assert.Equal(t, "test context", err.Context())
}

func TestValidationError_Unwrap(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "feat/x", "invalid")
	assert.NoError(t, err.Unwrap())
}

func TestValidationError_NoTrailingPunctuation(t *testing.T) {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	msg := err.Error()
	assert.NotRegexp(t, `\.$`, msg)
	assert.NotContains(t, msg, "💡")
}

func TestWorktreeServiceError_Error_WithBranchName(t *testing.T) {
	err := NewWorktreeServiceError("/path/to/worktree", "feature-branch", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to create")
	assert.Contains(t, msg, "/path/to/worktree")
	assert.Contains(t, msg, "branch: feature-branch")
	assert.NotContains(t, msg, "worktree service operation")
	assert.NotContains(t, msg, "CreateWorktree")
}

func TestWorktreeServiceError_Error_WithoutBranchName(t *testing.T) {
	err := NewWorktreeServiceError("/path/to/worktree", "", "DeleteWorktree", "failed to delete", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to delete")
	assert.Contains(t, msg, "/path/to/worktree")
	assert.NotContains(t, msg, "branch:")
	assert.NotContains(t, msg, "worktree service operation")
	assert.NotContains(t, msg, "DeleteWorktree")
}

func TestWorktreeServiceError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewWorktreeServiceError("/path", "branch", "operation", "message", cause)
	assert.Equal(t, cause, err.Unwrap())
}

func TestWorktreeServiceError_SentinelWalk(t *testing.T) {
	t.Run("ErrWorktreeNotFound walks through", func(t *testing.T) {
		err := NewWorktreeServiceError("/path", "branch", "op", "msg", ErrWorktreeNotFound)
		assert.ErrorIs(t, err, ErrWorktreeNotFound)
		assert.NotErrorIs(t, err, ErrProjectNotFound)
	})
	t.Run("matches without populated Err", func(t *testing.T) {
		err := NewWorktreeServiceError("/path", "branch", "op", "msg", nil)
		assert.ErrorIs(t, err, ErrWorktreeNotFound)
	})
}

func TestProjectServiceError_Error_WithProjectName(t *testing.T) {
	err := NewProjectServiceError("test-project", "/path/to/project", "DiscoverProject", "not a git repository", nil)
	msg := err.Error()

	assert.Contains(t, msg, "not a git repository")
	assert.Contains(t, msg, "test-project")
	assert.NotContains(t, msg, "project service operation")
	assert.NotContains(t, msg, "DiscoverProject")
}

func TestProjectServiceError_Error_WithoutProjectName(t *testing.T) {
	err := NewProjectServiceError("", "/path/to/project", "ValidateProject", "invalid path", nil)
	msg := err.Error()

	assert.Contains(t, msg, "invalid path")
	assert.Contains(t, msg, "/path/to/project")
	assert.NotContains(t, msg, "project '")
	assert.NotContains(t, msg, "project service operation")
	assert.NotContains(t, msg, "ValidateProject")
}

func TestProjectServiceError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewProjectServiceError("project", "/path", "operation", "message", cause)
	assert.Equal(t, cause, err.Unwrap())
}

func TestProjectServiceError_SentinelWalk(t *testing.T) {
	err := NewProjectServiceError("p", "/path", "op", "msg", nil)
	assert.ErrorIs(t, err, ErrProjectNotFound)
	assert.NotErrorIs(t, err, ErrWorktreeNotFound)
}

func TestNavigationServiceError_Error(t *testing.T) {
	err := NewNavigationServiceError("feature-branch", "project-root", "Navigate", "worktree not found", nil)
	msg := err.Error()

	assert.Contains(t, msg, "worktree not found")
	assert.Contains(t, msg, "feature-branch")
	assert.Contains(t, msg, "context: project-root")
	assert.NotContains(t, msg, "navigation service operation")
	assert.NotContains(t, msg, "Navigate")
}

func TestNavigationServiceError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewNavigationServiceError("target", "context", "operation", "message", cause)
	assert.Equal(t, cause, err.Unwrap())
}

func TestNavigationServiceError_SentinelWalk(t *testing.T) {
	err := NewNavigationServiceError("t", "ctx", "op", "msg", nil)
	assert.ErrorIs(t, err, ErrResolutionNotFound)
	assert.NotErrorIs(t, err, ErrWorktreeNotFound)
}

func TestResolutionError_Error_WithSuggestions(t *testing.T) {
	suggestions := []string{
		"Check if target exists",
		"Verify context is correct",
	}
	err := NewResolutionError("invalid-target", "project-root", "target not found", suggestions, nil)
	msg := err.Error()

	assert.Contains(t, msg, "resolution failed")
	assert.Contains(t, msg, "invalid-target")
	assert.Contains(t, msg, "project-root")
	assert.Contains(t, msg, "target not found")
	assert.Contains(t, msg, "suggestions:")
	assert.Contains(t, msg, "Check if target exists")
}

func TestResolutionError_Error_WithoutSuggestions(t *testing.T) {
	err := NewResolutionError("invalid-target", "project-root", "target not found", nil, nil)
	msg := err.Error()

	assert.Contains(t, msg, "resolution failed")
	assert.Contains(t, msg, "invalid-target")
	assert.Contains(t, msg, "project-root")
	assert.Contains(t, msg, "target not found")
	assert.NotContains(t, msg, "suggestions:")
}

func TestResolutionError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewResolutionError("target", "context", "message", nil, cause)
	assert.Equal(t, cause, err.Unwrap())
}

func TestResolutionError_SentinelWalk(t *testing.T) {
	err := NewResolutionError("t", "ctx", "msg", nil, nil)
	assert.ErrorIs(t, err, ErrResolutionNotFound)
	assert.NotErrorIs(t, err, ErrWorktreeNotFound)
}

func TestConflictError_Error(t *testing.T) {
	err := NewConflictError("worktree", "feature-branch", "CreateWorktree", "worktree already exists", nil)
	msg := err.Error()

	assert.Contains(t, msg, "conflict during")
	assert.Contains(t, msg, "CreateWorktree")
	assert.Contains(t, msg, "worktree")
	assert.Contains(t, msg, "feature-branch")
	assert.Contains(t, msg, "worktree already exists")
}

func TestConflictError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewConflictError("resource", "identifier", "operation", "message", cause)
	assert.Equal(t, cause, err.Unwrap())
}

func TestServiceError_Error(t *testing.T) {
	err := NewServiceError("WorktreeService", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to create")
	assert.NotContains(t, msg, "WorktreeService")
	assert.NotContains(t, msg, "CreateWorktree")
}

func TestServiceError_Unwrap(t *testing.T) {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewServiceError("Service", "Operation", "message", cause)
	assert.Equal(t, cause, err.Unwrap())
}
