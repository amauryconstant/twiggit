package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"twiggit/internal/domain"
)

func TestHandleCLIError_ValidationError(t *testing.T) {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeValidation, exitCode)
}

func TestHandleCLIError_GitRepositoryError(t *testing.T) {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeGit, exitCode)
}

func TestHandleCLIError_GitWorktreeError(t *testing.T) {
	err := domain.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeGit, exitCode)
}

func TestHandleCLIError_WorktreeServiceError(t *testing.T) {
	err := domain.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestHandleCLIError_ProjectServiceError(t *testing.T) {
	err := domain.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestHandleCLIError_NavigationServiceError(t *testing.T) {
	err := domain.NewNavigationServiceError("main", "project context", "ResolvePath", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestHandleCLIError_ServiceError(t *testing.T) {
	err := domain.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestHandleCLIError_GenericError(t *testing.T) {
	err := errors.New("generic error")

	exitCode := HandleCLIError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestGetExitCodeForError_ValidationError(t *testing.T) {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	exitCode := GetExitCodeForError(err)

	assert.Equal(t, ExitCodeValidation, exitCode)
}

func TestGetExitCodeForError_GitError(t *testing.T) {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	exitCode := GetExitCodeForError(err)

	assert.Equal(t, ExitCodeGit, exitCode)
}

func TestGetExitCodeForError_UnknownError(t *testing.T) {
	err := errors.New("unknown error")

	exitCode := GetExitCodeForError(err)

	assert.Equal(t, ExitCodeError, exitCode)
}

func TestCategorizeError_ValidationError(t *testing.T) {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryValidation, category)
}

func TestCategorizeError_WorktreeServiceError(t *testing.T) {
	err := domain.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryService, category)
}

func TestCategorizeError_ProjectServiceError(t *testing.T) {
	err := domain.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryService, category)
}

func TestCategorizeError_ServiceError(t *testing.T) {
	err := domain.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryService, category)
}

func TestCategorizeError_GitRepositoryError(t *testing.T) {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryGit, category)
}

func TestCategorizeError_GitWorktreeError(t *testing.T) {
	err := domain.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryGit, category)
}

func TestCategorizeError_GitCommandError(t *testing.T) {
	err := domain.NewGitCommandError("git", []string{"status"}, 1, "", "error output", "command failed", errors.New("some error"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryGit, category)
}

func TestCategorizeError_ConfigError(t *testing.T) {
	err := domain.NewConfigError("/path/to/config.toml", "failed to parse config file", nil)

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryConfig, category)
}

func TestCategorizeError_ConfigErrorWithCause(t *testing.T) {
	err := domain.NewConfigError("/path/to/config.toml", "validation failed", errors.New("invalid field"))

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryConfig, category)
}

func TestCategorizeError_ValidationErrorString(t *testing.T) {
	err := errors.New("validation failed: field is required")

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryValidation, category)
}

func TestCategorizeError_InvalidString(t *testing.T) {
	err := errors.New("invalid input")

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryValidation, category)
}

func TestCategorizeError_Generic(t *testing.T) {
	err := errors.New("some random error")

	category := CategorizeError(err)

	assert.Equal(t, ErrorCategoryGeneric, category)
}

func TestIsCobraArgumentError_AcceptsPattern(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := IsCobraArgumentError(tt.err)
			assert.Equal(t, tt.match, result)
		})
	}
}

func TestIsCobraArgumentError_CobraError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()

	if err != nil {
		t.Run("Cobra command error", func(t *testing.T) {
			result := IsCobraArgumentError(err)
			assert.True(t, result)
		})
	}
}
