package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"twiggit/internal/domain"
)

func TestHandleCLIError_NilError(t *testing.T) {
	assert.Equal(t, ExitCodeSuccess, HandleCLIError(nil))
}

func TestHandleCLIError_ValidationError(t *testing.T) {
	err := domain.NewValidationError("TestRequest", "field", "value", "invalid field")
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_GitRepositoryError(t *testing.T) {
	err := domain.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_GitWorktreeError(t *testing.T) {
	err := domain.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_WorktreeServiceError(t *testing.T) {
	err := domain.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_ProjectServiceError(t *testing.T) {
	err := domain.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_NavigationServiceError(t *testing.T) {
	err := domain.NewNavigationServiceError("main", "project context", "ResolvePath", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_ServiceError(t *testing.T) {
	err := domain.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_GenericError(t *testing.T) {
	err := errors.New("generic error")
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestGetExitCodeForError_ThreeCodeDispatch(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ExitCode
	}{
		{"nil", nil, ExitCodeSuccess},
		{"validation error", domain.NewValidationError("R", "f", "v", "m"), ExitCodeError},
		{"git repo error", domain.NewGitRepositoryError("/p", "m", nil), ExitCodeError},
		{"git worktree error", domain.NewGitWorktreeError("/p", "b", "m", nil), ExitCodeError},
		{"git command error", domain.NewGitCommandError("g", nil, 1, "", "", "m", nil), ExitCodeError},
		{"worktree service error", domain.NewWorktreeServiceError("/p", "b", "o", "m", nil), ExitCodeError},
		{"project service error", domain.NewProjectServiceError("n", "/p", "o", "m", nil), ExitCodeError},
		{"navigation service error", domain.NewNavigationServiceError("t", "c", "o", "m", nil), ExitCodeError},
		{"service error", domain.NewServiceError("S", "O", "m", nil), ExitCodeError},
		{"conflict error", domain.NewConflictError("r", "i", "o", "m", nil), ExitCodeError},
		{"shell already installed", domain.NewShellAlreadyInstalledError("bash", "ctx", nil), ExitCodeError},
		{"shell not installed", domain.NewShellNotInstalledError("bash", "ctx", nil), ExitCodeError},
		{"shell invalid type", domain.NewShellInvalidTypeError("powershell", "ctx", nil), ExitCodeError},
		{"shell inference", domain.NewShellInferenceError("bash", "ctx", nil), ExitCodeError},
		{"shell detection", domain.NewShellDetectionError("ctx", nil), ExitCodeError},
		{"shell wrapper generation", domain.NewShellWrapperError("bash", "generation", "ctx", nil), ExitCodeError},
		{"shell wrapper installation", domain.NewShellWrapperError("bash", "installation", "ctx", nil), ExitCodeError},
		{"shell config", domain.NewShellConfigError("/p", "ctx", nil), ExitCodeError},
		{"generic error", errors.New("plain"), ExitCodeError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetExitCodeForError(tt.err))
		})
	}
}

func TestGetExitCodeForError_CobraUsageError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"pflag value required", &pflag.ValueRequiredError{}},
		{"pflag invalid value", &pflag.InvalidValueError{}},
		{"pflag invalid syntax", &pflag.InvalidSyntaxError{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ExitCodeUsage, GetExitCodeForError(tt.err))
		})
	}
}

func TestCategorizeError_SentinelNotFoundDispatch(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"err git repo not found", domain.NewGitRepositoryError("/p", "m", nil)},
		{"err worktree not found", domain.NewGitWorktreeError("/p", "b", "m", nil)},
		{"err project not found", domain.NewProjectServiceError("n", "/p", "o", "m", nil)},
		{"err resolution not found", domain.NewNavigationServiceError("t", "c", "o", "m", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ErrorCategoryService, CategorizeError(tt.err))
		})
	}
}

func TestCategorizeError_CobraUsageError(t *testing.T) {
	assert.Equal(t, ErrorCategoryCobra, CategorizeError(&pflag.ValueRequiredError{}))
}

func TestCategorizeError_ValidationError(t *testing.T) {
	assert.Equal(t, ErrorCategoryService, CategorizeError(domain.NewValidationError("R", "f", "v", "m")))
}

func TestCategorizeError_Generic(t *testing.T) {
	assert.Equal(t, ErrorCategoryGeneric, CategorizeError(errors.New("plain")))
}

func TestIsCobraUsageError_PflagErrors(t *testing.T) {
	assert.True(t, IsCobraUsageError(&pflag.ValueRequiredError{}))
	assert.True(t, IsCobraUsageError(&pflag.InvalidValueError{}))
	assert.True(t, IsCobraUsageError(&pflag.InvalidSyntaxError{}))
}

func TestIsCobraUsageError_NonCobraError(t *testing.T) {
	assert.False(t, IsCobraUsageError(errors.New("plain")))
	assert.False(t, IsCobraUsageError(domain.NewValidationError("R", "f", "v", "m")))
	assert.False(t, IsCobraUsageError(nil))
}

func TestIsCobraArgumentError_Alias(t *testing.T) {
	assert.Equal(t, IsCobraUsageError(&pflag.ValueRequiredError{}), IsCobraArgumentError(&pflag.ValueRequiredError{}))
}
