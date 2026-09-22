package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
)

func TestHandleCLIError_NilError(t *testing.T) {
	assert.Equal(t, ExitCodeSuccess, HandleCLIError(nil))
}

func TestHandleCLIError_ValidationError(t *testing.T) {
	err := core.NewOpValidationError("TestRequest", "field", "value", "invalid field")
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_GitRepositoryError(t *testing.T) {
	err := core.NewGitRepositoryError("/path/to/repo", "failed to open", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_GitWorktreeError(t *testing.T) {
	err := core.NewGitWorktreeError("/path/to/worktree", "feature", "failed to delete", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_WorktreeServiceError(t *testing.T) {
	err := core.NewWorktreeServiceError("/path/to/worktree", "feature", "DeleteWorktree", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_ProjectServiceError(t *testing.T) {
	err := core.NewProjectServiceError("myproject", "/path/to/project", "DiscoverProject", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_NavigationServiceError(t *testing.T) {
	err := core.NewNavigationServiceError("main", "project context", "ResolvePath", "operation failed", errors.New("some error"))
	assert.Equal(t, ExitCodeError, HandleCLIError(err))
}

func TestHandleCLIError_ServiceError(t *testing.T) {
	err := core.NewServiceError("MyService", "DoSomething", "operation failed", errors.New("some error"))
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
		{"validation error", core.NewOpValidationError("R", "f", "v", "m"), ExitCodeError},
		{"git repo error", core.NewGitRepositoryError("/p", "m", nil), ExitCodeError},
		{"git worktree error", core.NewGitWorktreeError("/p", "b", "m", nil), ExitCodeError},
		{"git command error", core.NewGitCommandError("g", nil, 1, "", "", "m", nil), ExitCodeError},
		{"worktree service error", core.NewWorktreeServiceError("/p", "b", "o", "m", nil), ExitCodeError},
		{"project service error", core.NewProjectServiceError("n", "/p", "o", "m", nil), ExitCodeError},
		{"navigation service error", core.NewNavigationServiceError("t", "c", "o", "m", nil), ExitCodeError},
		{"service error", core.NewServiceError("S", "O", "m", nil), ExitCodeError},
		{"conflict error", core.NewConflictError("r", "i", "o", "m", nil), ExitCodeError},
		{"shell already installed", core.NewShellAlreadyInstalledError("bash", "ctx", nil), ExitCodeError},
		{"shell not installed", core.NewShellNotInstalledError("bash", "ctx", nil), ExitCodeError},
		{"shell invalid type", core.NewShellInvalidTypeError("powershell", "ctx", nil), ExitCodeError},
		{"shell inference", core.NewShellInferenceError("bash", "ctx", nil), ExitCodeError},
		{"shell detection", core.NewShellDetectionError("ctx", nil), ExitCodeError},
		{"shell wrapper generation", core.NewShellWrapperError("bash", "generation", "ctx", nil), ExitCodeError},
		{"shell wrapper installation", core.NewShellWrapperError("bash", "installation", "ctx", nil), ExitCodeError},
		{"shell config", core.NewShellConfigError("/p", "ctx", nil), ExitCodeError},
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
		{"usage error with cause", core.NewUsageError("--config requires --install", errors.New("flag: --config"))},
		{"usage error no cause", core.NewUsageError("--force requires --install", nil)},
		{"wrapped usage error", core.UsageWrap(errors.New("raw pflag error"))},
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
		{"err git repo not found", core.NewGitRepositoryError("/p", "m", nil)},
		{"err worktree not found", core.NewGitWorktreeError("/p", "b", "m", nil)},
		{"err project not found", core.NewProjectServiceError("n", "/p", "o", "m", nil)},
		{"err resolution not found", core.NewNavigationServiceError("t", "c", "o", "m", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ErrorCategoryService, CategorizeError(tt.err))
		})
	}
}

func TestCategorizeError_UsageError(t *testing.T) {
	assert.Equal(t, ErrorCategoryCobra, CategorizeError(core.NewUsageError("--foo", nil)))
}

func TestCategorizeError_ValidationError(t *testing.T) {
	assert.Equal(t, ErrorCategoryService, CategorizeError(core.NewOpValidationError("R", "f", "v", "m")))
}

func TestCategorizeError_Generic(t *testing.T) {
	assert.Equal(t, ErrorCategoryGeneric, CategorizeError(errors.New("plain")))
}

func TestIsCobraUsageError_UsageError(t *testing.T) {
	assert.True(t, IsCobraUsageError(core.NewUsageError("--foo", nil)))
	assert.True(t, IsCobraUsageError(core.UsageWrap(errors.New("raw"))))
}

func TestIsCobraUsageError_NonUsageError(t *testing.T) {
	assert.False(t, IsCobraUsageError(errors.New("plain")))
	assert.False(t, IsCobraUsageError(core.NewOpValidationError("R", "f", "v", "m")))
	assert.False(t, IsCobraUsageError(nil))
}

// TestUsageError_IsTerminal pins the terminal-Unwrap contract added when
// the Err field was dropped from UsageError.
func TestUsageError_IsTerminal(t *testing.T) {
	err := core.NewUsageError("--foo", errors.New("parser failure"))
	assert.NoError(t, err.Unwrap(), "UsageError.Unwrap returns nil (terminal)")
}

// TestIsCobraArgumentError_AliasParity pins the legacy-alias contract: any future
// rename of the canonical detector must keep `IsCobraArgumentError` returning the
// same boolean so external callers do not silently diverge.
func TestIsCobraArgumentError_AliasParity(t *testing.T) {
	err := core.NewUsageError("--foo", nil)
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

	var ue *core.UsageError
	is.ErrorAs(err, &ue, "args-validator failure must wrap into *core.UsageError")
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
	err := core.NewOpValidationError("Req", "field", "value", "validation failed")
	is := assert.New(t)
	is.Equal(ExitCodeError, HandleCLIErrorWithCommand(nil, err))
	is.Equal(ExitCodeError, GetExitCodeForError(err))
}

func TestHandleCLIErrorWithCommand_NotFoundReturnsOne(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"git repo not found", core.NewGitRepositoryError("/p", "m", nil)},
		{"worktree not found", core.NewGitWorktreeError("/p", "b", "m", nil)},
		{"project not found", core.NewProjectServiceError("n", "/p", "o", "m", nil)},
		{"resolution not found", core.NewNavigationServiceError("t", "c", "o", "m", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			is.Equal(ExitCodeError, HandleCLIErrorWithCommand(nil, tt.err))
			is.Equal(ExitCodeError, GetExitCodeForError(tt.err))
		})
	}
}
