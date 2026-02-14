package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

// ErrorHandlerTestSuite tests error handler functions
type ErrorHandlerTestSuite struct {
	suite.Suite
}

func TestErrorHandlerSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlerTestSuite))
}

// Test HandleCLIError with various error types
func (s *ErrorHandlerTestSuite) TestHandleCLIError_ValidationError() {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_GitRepositoryError() {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_GitWorktreeError() {
	err := domain.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_WorktreeServiceError() {
	err := domain.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_ProjectServiceError() {
	err := domain.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_NavigationServiceError() {
	err := domain.NewNavigationServiceError("main", "project context", "ResolvePath", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_ServiceError() {
	err := domain.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestHandleCLIError_GenericError() {
	err := errors.New("generic error")

	exitCode := HandleCLIError(err)

	s.Equal(ExitCodeError, exitCode)
}

// Test GetExitCodeForError mapping
func (s *ErrorHandlerTestSuite) TestGetExitCodeForError_ValidationError() {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	exitCode := GetExitCodeForError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestGetExitCodeForError_GitError() {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	exitCode := GetExitCodeForError(err)

	s.Equal(ExitCodeError, exitCode)
}

func (s *ErrorHandlerTestSuite) TestGetExitCodeForError_UnknownError() {
	err := errors.New("unknown error")

	exitCode := GetExitCodeForError(err)

	s.Equal(ExitCodeError, exitCode)
}

// Test CategorizeError logic
func (s *ErrorHandlerTestSuite) TestCategorizeError_ValidationError() {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	category := CategorizeError(err)

	s.Equal(ErrorCategoryValidation, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_WorktreeServiceError() {
	err := domain.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryService, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_ProjectServiceError() {
	err := domain.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryService, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_ServiceError() {
	err := domain.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryService, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_GitRepositoryError() {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryGit, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_GitWorktreeError() {
	err := domain.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryGit, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_GitCommandError() {
	err := domain.NewGitCommandError("git", []string{"status"}, 1, "", "error output", "command failed", errors.New("some error"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryGit, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_ConfigError() {
	err := domain.NewConfigError("/path/to/config.toml", "failed to parse config file", nil)

	category := CategorizeError(err)

	s.Equal(ErrorCategoryConfig, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_ConfigErrorWithCause() {
	err := domain.NewConfigError("/path/to/config.toml", "validation failed", errors.New("invalid field"))

	category := CategorizeError(err)

	s.Equal(ErrorCategoryConfig, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_ValidationErrorString() {
	err := errors.New("validation failed: field is required")

	category := CategorizeError(err)

	s.Equal(ErrorCategoryValidation, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_InvalidString() {
	err := errors.New("invalid input")

	category := CategorizeError(err)

	s.Equal(ErrorCategoryValidation, category)
}

func (s *ErrorHandlerTestSuite) TestCategorizeError_Generic() {
	err := errors.New("some random error")

	category := CategorizeError(err)

	s.Equal(ErrorCategoryGeneric, category)
}

// Test IsCobraArgumentError pattern matching
func (s *ErrorHandlerTestSuite) TestIsCobraArgumentError_AcceptsPattern() {
	tests := []struct {
		name  string
		err   error
		match bool
	}{
		{
			name:  "accepts pattern",
			err:   errors.New("accepts 1 arg(s), received 2"),
			match: true,
		},
		{
			name:  "requires pattern",
			err:   errors.New("requires at least 1 arg(s), only received 0"),
			match: true,
		},
		{
			name:  "received pattern",
			err:   errors.New("received 2 args, expected 1"),
			match: true,
		},
		{
			name:  "unknown shorthand flag",
			err:   errors.New("unknown shorthand flag: 'x' in -x"),
			match: true,
		},
		{
			name:  "unknown flag",
			err:   errors.New("unknown flag: --unknown-flag"),
			match: true,
		},
		{
			name:  "flag needs an argument",
			err:   errors.New("flag needs an argument: 'b' in -b"),
			match: true,
		},
		{
			name:  "required flag",
			err:   errors.New("required flag(s) \"branch\" not set"),
			match: true,
		},
		{
			name:  "non-Cobra error",
			err:   errors.New("some other error message"),
			match: false,
		},
		{
			name:  "empty error",
			err:   errors.New(""),
			match: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := IsCobraArgumentError(tt.err)
			s.Equal(tt.match, result)
		})
	}
}

func (s *ErrorHandlerTestSuite) TestIsCobraArgumentError_CobraError() {
	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()

	if err != nil {
		s.Run("Cobra command error", func() {
			result := IsCobraArgumentError(err)
			s.True(result)
		})
	}
}
