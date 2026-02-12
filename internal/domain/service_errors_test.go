package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceErrorsTestSuite struct {
	suite.Suite
}

func TestServiceErrors(t *testing.T) {
	suite.Run(t, new(ServiceErrorsTestSuite))
}

func (s *ServiceErrorsTestSuite) TestValidationError_Error_WithoutSuggestions() {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	msg := err.Error()

	s.Contains(msg, "validation failed")
	s.Contains(msg, "CreateWorktree.branch")
	s.Contains(msg, "cannot be empty")
	s.Contains(msg, "value:")
	s.NotContains(msg, "💡")
}

func (s *ServiceErrorsTestSuite) TestValidationError_Error_WithSuggestions() {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty").
		WithSuggestions([]string{
			"Use a valid branch name",
			"Branch names should not be empty",
		})
	msg := err.Error()

	s.Contains(msg, "validation failed")
	s.Contains(msg, "CreateWorktree.branch")
	s.Contains(msg, "cannot be empty")
	s.Contains(msg, "💡 Use a valid branch name")
	s.Contains(msg, "💡 Branch names should not be empty")
}

func (s *ServiceErrorsTestSuite) TestValidationError_Error_WithEmptySuggestions() {
	err := NewValidationError("CreateWorktree", "branch", "", "cannot be empty").
		WithSuggestions([]string{})
	msg := err.Error()

	s.Contains(msg, "validation failed")
	s.NotContains(msg, "💡")
}

func (s *ServiceErrorsTestSuite) TestValidationError_WithSuggestions_Immutability() {
	original := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	suggestions := []string{
		"Use a valid branch name",
		"Branch names should not be empty",
	}

	modified := original.WithSuggestions(suggestions)

	s.NotEqual(original, modified)
	s.Empty(original.Suggestions())
	s.Equal(suggestions, modified.Suggestions())

	s.ModifySuggestionsAndCheck(original, suggestions)
}

func (s *ServiceErrorsTestSuite) ModifySuggestionsAndCheck(original *ValidationError, suggestions []string) {
	s.T().Helper()

	suggestions[0] = "Modified suggestion"
	s.Empty(original.Suggestions())
	s.Contains(suggestions, "Modified suggestion")
}

func (s *ServiceErrorsTestSuite) TestValidationError_WithContext_Immutability() {
	original := NewValidationError("CreateWorktree", "branch", "", "cannot be empty")
	modified := original.WithContext("Additional context information")

	s.NotEqual(original, modified)
	s.Empty(original.Context())
	s.Equal("Additional context information", modified.Context())
}

func (s *ServiceErrorsTestSuite) TestValidationError_WithSuggestionsThenWithContext() {
	err := NewValidationError("CreateWorktree", "branch", "feat/test", "invalid format").
		WithSuggestions([]string{"Use kebab-case for branch names"}).
		WithContext("Branch name validation")

	msg := err.Error()
	s.Contains(msg, "validation failed")
	s.Contains(msg, "💡 Use kebab-case for branch names")
	s.Equal("Branch name validation", err.Context())
	s.Equal([]string{"Use kebab-case for branch names"}, err.Suggestions())
}

func (s *ServiceErrorsTestSuite) TestValidationError_Getters() {
	err := NewValidationError("CreateWorktree", "branch", "test-branch", "invalid format").
		WithSuggestions([]string{"suggestion 1", "suggestion 2"}).
		WithContext("test context")

	s.Equal("branch", err.Field())
	s.Equal("test-branch", err.Value())
	s.Equal("invalid format", err.Message())
	s.Equal("CreateWorktree", err.Request())
	s.Equal([]string{"suggestion 1", "suggestion 2"}, err.Suggestions())
	s.Equal("test context", err.Context())
}

func (s *ServiceErrorsTestSuite) TestWorktreeServiceError_Error_WithBranchName() {
	err := NewWorktreeServiceError("/path/to/worktree", "feature-branch", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	s.Contains(msg, "worktree service operation")
	s.Contains(msg, "CreateWorktree")
	s.Contains(msg, "/path/to/worktree")
	s.Contains(msg, "branch: feature-branch")
	s.Contains(msg, "failed to create")
}

func (s *ServiceErrorsTestSuite) TestWorktreeServiceError_Error_WithoutBranchName() {
	err := NewWorktreeServiceError("/path/to/worktree", "", "DeleteWorktree", "failed to delete", nil)
	msg := err.Error()

	s.Contains(msg, "worktree service operation")
	s.Contains(msg, "DeleteWorktree")
	s.Contains(msg, "/path/to/worktree")
	s.NotContains(msg, "branch:")
	s.Contains(msg, "failed to delete")
}

func (s *ServiceErrorsTestSuite) TestWorktreeServiceError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewWorktreeServiceError("/path", "branch", "operation", "message", cause)
	s.Equal(cause, err.Unwrap())
}

func (s *ServiceErrorsTestSuite) TestProjectServiceError_Error_WithProjectName() {
	err := NewProjectServiceError("test-project", "/path/to/project", "DiscoverProject", "not a git repository", nil)
	msg := err.Error()

	s.Contains(msg, "project service operation")
	s.Contains(msg, "DiscoverProject")
	s.Contains(msg, "test-project")
	s.Contains(msg, "not a git repository")
}

func (s *ServiceErrorsTestSuite) TestProjectServiceError_Error_WithoutProjectName() {
	err := NewProjectServiceError("", "/path/to/project", "ValidateProject", "invalid path", nil)
	msg := err.Error()

	s.Contains(msg, "project service operation")
	s.Contains(msg, "ValidateProject")
	s.Contains(msg, "/path/to/project")
	s.NotContains(msg, "project '")
	s.Contains(msg, "invalid path")
}

func (s *ServiceErrorsTestSuite) TestProjectServiceError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewProjectServiceError("project", "/path", "operation", "message", cause)
	s.Equal(cause, err.Unwrap())
}

func (s *ServiceErrorsTestSuite) TestNavigationServiceError_Error() {
	err := NewNavigationServiceError("feature-branch", "project-root", "Navigate", "worktree not found", nil)
	msg := err.Error()

	s.Contains(msg, "navigation service operation")
	s.Contains(msg, "Navigate")
	s.Contains(msg, "feature-branch")
	s.Contains(msg, "context: project-root")
	s.Contains(msg, "worktree not found")
}

func (s *ServiceErrorsTestSuite) TestNavigationServiceError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewNavigationServiceError("target", "context", "operation", "message", cause)
	s.Equal(cause, err.Unwrap())
}

func (s *ServiceErrorsTestSuite) TestResolutionError_Error_WithSuggestions() {
	suggestions := []string{
		"Check if the target exists",
		"Verify the context is correct",
	}
	err := NewResolutionError("invalid-target", "project-root", "target not found", suggestions, nil)
	msg := err.Error()

	s.Contains(msg, "resolution failed")
	s.Contains(msg, "invalid-target")
	s.Contains(msg, "project-root")
	s.Contains(msg, "target not found")
	s.Contains(msg, "suggestions:")
	s.Contains(msg, "Check if the target exists")
}

func (s *ServiceErrorsTestSuite) TestResolutionError_Error_WithoutSuggestions() {
	err := NewResolutionError("invalid-target", "project-root", "target not found", nil, nil)
	msg := err.Error()

	s.Contains(msg, "resolution failed")
	s.Contains(msg, "invalid-target")
	s.Contains(msg, "project-root")
	s.Contains(msg, "target not found")
	s.NotContains(msg, "suggestions:")
}

func (s *ServiceErrorsTestSuite) TestResolutionError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewResolutionError("target", "context", "message", nil, cause)
	s.Equal(cause, err.Unwrap())
}

func (s *ServiceErrorsTestSuite) TestConflictError_Error() {
	err := NewConflictError("worktree", "feature-branch", "CreateWorktree", "worktree already exists", nil)
	msg := err.Error()

	s.Contains(msg, "conflict during")
	s.Contains(msg, "CreateWorktree")
	s.Contains(msg, "worktree")
	s.Contains(msg, "feature-branch")
	s.Contains(msg, "worktree already exists")
}

func (s *ServiceErrorsTestSuite) TestConflictError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewConflictError("resource", "identifier", "operation", "message", cause)
	s.Equal(cause, err.Unwrap())
}

func (s *ServiceErrorsTestSuite) TestServiceError_Error() {
	err := NewServiceError("WorktreeService", "CreateWorktree", "failed to create", nil)
	msg := err.Error()

	s.Contains(msg, "WorktreeService")
	s.Contains(msg, "CreateWorktree")
	s.Contains(msg, "failed")
}

func (s *ServiceErrorsTestSuite) TestServiceError_Unwrap() {
	cause := NewValidationError("request", "field", "value", "error")
	err := NewServiceError("Service", "Operation", "message", cause)
	s.Equal(cause, err.Unwrap())
}
