package core

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error_WithoutSuggestions(t *testing.T) {
	err := &ValidationError{
		Op:      "CreateWorktree",
		Field:   "branch",
		Value:   "",
		Message: "cannot be empty",
	}
	msg := err.Error()

	assert.Contains(t, msg, "validation failed")
	assert.Contains(t, msg, "CreateWorktree")
	assert.Contains(t, msg, "branch")
	assert.Contains(t, msg, "cannot be empty")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestValidationError_Suggestions(t *testing.T) {
	err := &ValidationError{
		Op:          "CreateWorktree",
		Field:       "branch",
		Message:     "cannot be empty",
		Suggestions: []string{"Use a valid branch name", "Branch names should not be empty"},
	}

	assert.True(t, slices.Equal([]string{"Use a valid branch name", "Branch names should not be empty"}, err.Suggestions))

	msg := err.Error()
	assert.Contains(t, msg, "validation failed")
	assert.Contains(t, msg, "CreateWorktree")
	assert.Contains(t, msg, "cannot be empty")
}

func TestValidationError_Getters(t *testing.T) {
	err := &ValidationError{
		Op:          "CreateWorktree",
		Field:       "branch",
		Value:       "test-branch",
		Message:     "invalid format",
		Suggestions: []string{"suggestion 1", "suggestion 2"},
	}

	assert.Equal(t, "CreateWorktree", err.Op)
	assert.Equal(t, "branch", err.Field)
	assert.Equal(t, "test-branch", err.Value)
	assert.Equal(t, "invalid format", err.Message)
	assert.True(t, slices.Equal([]string{"suggestion 1", "suggestion 2"}, err.Suggestions))
}

func TestValidationError_Unwrap(t *testing.T) {
	err := &ValidationError{
		Op:      "CreateWorktree",
		Field:   "branch",
		Message: "invalid",
	}
	assert.NoError(t, err.Unwrap())
}

func TestValidationError_NoTrailingPunctuation(t *testing.T) {
	err := &ValidationError{
		Op:      "CreateWorktree",
		Field:   "branch",
		Message: "cannot be empty",
	}
	msg := err.Error()
	assert.NotRegexp(t, `\.$`, msg)
}

func TestValidationError_SentinelWalk(t *testing.T) {
	t.Run("worktree.service walks to ErrWorktreeNotFound", func(t *testing.T) {
		err := NewWorktreeServiceError("/path", "branch", "op", "msg", nil)
		assert.ErrorIs(t, err, ErrWorktreeNotFound)
		assert.NotErrorIs(t, err, ErrProjectNotFound)
	})
	t.Run("project.service walks to ErrProjectNotFound", func(t *testing.T) {
		err := NewProjectServiceError("p", "/path", "op", "msg", nil)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
	t.Run("navigation.service walks to ErrResolutionNotFound", func(t *testing.T) {
		err := NewNavigationServiceError("t", "ctx", "op", "msg", nil)
		assert.ErrorIs(t, err, ErrResolutionNotFound)
	})
}

func TestWorktreeServiceError_Error_WithBranchName(t *testing.T) {
	err := NewWorktreeServiceError("/path/to/worktree", "feature-branch", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to create")
	assert.Contains(t, msg, "/path/to/worktree")
	assert.Contains(t, msg, "branch:feature-branch")
}

func TestWorktreeServiceError_Error_WithoutBranchName(t *testing.T) {
	err := NewWorktreeServiceError("/path/to/worktree", "", "DeleteWorktree", "failed to delete", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to delete")
	assert.Contains(t, msg, "/path/to/worktree")
}

func TestProjectServiceError_Error_WithProjectName(t *testing.T) {
	err := NewProjectServiceError("test-project", "/path/to/project", "DiscoverProject", "not a git repository", nil)
	msg := err.Error()

	assert.Contains(t, msg, "not a git repository")
	assert.Contains(t, msg, "test-project")
}

func TestNavigationServiceError_Error(t *testing.T) {
	err := NewNavigationServiceError("feature-branch", "project-root", "Navigate", "worktree not found", nil)
	msg := err.Error()

	assert.Contains(t, msg, "worktree not found")
	assert.Contains(t, msg, "feature-branch")
	assert.Contains(t, msg, "context:project-root")
}

func TestResolutionError_Error_WithSuggestions(t *testing.T) {
	suggestions := []string{
		"Check if target exists",
		"Verify context is correct",
	}
	err := NewResolutionError("invalid-target", "project-root", "target not found", suggestions, nil)
	msg := err.Error()

	assert.Contains(t, msg, "invalid-target")
	assert.Contains(t, msg, "project-root")
	assert.Contains(t, msg, "target not found")
	assert.Equal(t, suggestions, err.Suggestions)
}

func TestResolutionError_Error_WithoutSuggestions(t *testing.T) {
	err := NewResolutionError("invalid-target", "project-root", "target not found", nil, nil)
	msg := err.Error()

	assert.Contains(t, msg, "invalid-target")
	assert.Contains(t, msg, "project-root")
	assert.Contains(t, msg, "target not found")
}

func TestConflictError_Error(t *testing.T) {
	err := NewConflictError("worktree", "feature-branch", "CreateWorktree", "worktree already exists", nil)
	msg := err.Error()

	assert.Contains(t, msg, "worktree")
	assert.Contains(t, msg, "feature-branch")
	assert.Contains(t, msg, "worktree already exists")
}

func TestServiceError_Error(t *testing.T) {
	err := NewServiceError("WorktreeService", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	assert.Contains(t, msg, "failed to create")
}
