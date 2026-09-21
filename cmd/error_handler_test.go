package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestGetExitCodeForError_UsageError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"usage error with cause", domain.NewUsageError("--config requires --install", errors.New("flag: --config"))},
		{"usage error no cause", domain.NewUsageError("--force requires --install", nil)},
		{"wrapped usage error", domain.UsageWrap(errors.New("raw pflag error"))},
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

func TestCategorizeError_UsageError(t *testing.T) {
	assert.Equal(t, ErrorCategoryCobra, CategorizeError(domain.NewUsageError("--foo", nil)))
}

func TestCategorizeError_ValidationError(t *testing.T) {
	assert.Equal(t, ErrorCategoryService, CategorizeError(domain.NewValidationError("R", "f", "v", "m")))
}

func TestCategorizeError_Generic(t *testing.T) {
	assert.Equal(t, ErrorCategoryGeneric, CategorizeError(errors.New("plain")))
}

func TestIsCobraUsageError_UsageError(t *testing.T) {
	assert.True(t, IsCobraUsageError(domain.NewUsageError("--foo", nil)))
	assert.True(t, IsCobraUsageError(domain.UsageWrap(errors.New("raw"))))
}

func TestIsCobraUsageError_NonUsageError(t *testing.T) {
	assert.False(t, IsCobraUsageError(errors.New("plain")))
	assert.False(t, IsCobraUsageError(domain.NewValidationError("R", "f", "v", "m")))
	assert.False(t, IsCobraUsageError(nil))
}

func TestUsageError_SentinelParticipation(t *testing.T) {
	err := domain.NewUsageError("--foo", nil)
	assert.ErrorIs(t, err, domain.ErrUsageFlag)
	assert.NotErrorIs(t, err, domain.ErrWorktreeNotFound)
}

func TestUsageError_Unwrap(t *testing.T) {
	cause := errors.New("parser failure")
	err := domain.NewUsageError("--foo", cause)
	assert.Equal(t, cause, err.Unwrap())
	assert.ErrorIs(t, err, domain.ErrUsageFlag)
	assert.ErrorIs(t, err, cause)
}

// TestIsCobraArgumentError_AliasParity pins the legacy-alias contract: any future
// rename of the canonical detector must keep `IsCobraArgumentError` returning the
// same boolean so external callers do not silently diverge.
func TestIsCobraArgumentError_AliasParity(t *testing.T) {
	err := domain.NewUsageError("--foo", nil)
	assert.Equal(t, IsCobraUsageError(err), IsCobraArgumentError(err))
}

// TestWrapArgsValidator_WrapsAsUsageError pins the wrap contract used by every
// command's Args: field. Cobra's Args: validator failures reach Execute()
// unwrapped; without the helper, IsCobraUsageError returns false and the error
// dispatches to ExitCodeError (1) instead of ExitCodeUsage (2).
func TestWrapArgsValidator_WrapsAsUsageError(t *testing.T) {
	must := require.New(t)
	is := assert.New(t)

	wrapped := wrapArgsValidator(cobra.ExactArgs(1))
	cmd := &cobra.Command{Use: "x", Args: wrapped}
	cmd.SetArgs([]string{})

	err := wrapped(cmd, nil)
	must.Error(err)

	var ue *domain.UsageError
	is.ErrorAs(err, &ue, "args-validator failure must wrap into *domain.UsageError")
	is.True(IsCobraUsageError(err), "IsCobraUsageError must match the wrapped error")
	is.Equal(ExitCodeUsage, GetExitCodeForError(err), "exit code must be 2")
}

func TestWrapArgsValidator_NilOnPass(t *testing.T) {
	is := assert.New(t)
	wrapped := wrapArgsValidator(cobra.NoArgs)
	cmd := &cobra.Command{Use: "x", Args: wrapped}
	is.NoError(wrapped(cmd, nil))
	is.NoError(wrapped(cmd, []string{}))
}

func TestHandleCLIErrorWithCommand_ValidationErrorReturnsOne(t *testing.T) {
	err := domain.NewValidationError("Req", "field", "value", "validation failed")
	is := assert.New(t)
	is.Equal(ExitCodeError, HandleCLIErrorWithCommand(nil, err))
	is.Equal(ExitCodeError, GetExitCodeForError(err))
}

func TestHandleCLIErrorWithCommand_NotFoundReturnsOne(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"git repo not found", domain.NewGitRepositoryError("/p", "m", nil)},
		{"worktree not found", domain.NewGitWorktreeError("/p", "b", "m", nil)},
		{"project not found", domain.NewProjectServiceError("n", "/p", "o", "m", nil)},
		{"resolution not found", domain.NewNavigationServiceError("t", "c", "o", "m", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			is.Equal(ExitCodeError, HandleCLIErrorWithCommand(nil, tt.err))
			is.Equal(ExitCodeError, GetExitCodeForError(tt.err))
		})
	}
}
