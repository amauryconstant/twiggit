package cmd

import (
	"errors"
	"strings"
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter()
	assert.NotNil(t, formatter)
	assert.False(t, formatter.quiet)
}

func TestNewErrorFormatterWithOptions(t *testing.T) {
	tests := []struct {
		name  string
		quiet bool
		want  bool
	}{
		{"quiet mode disabled", false, false},
		{"quiet mode enabled", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatterWithOptions(tt.quiet)
			assert.NotNil(t, formatter)
			assert.Equal(t, tt.want, formatter.quiet)
		})
	}
}

func TestErrorFormatter_FormatValidationError(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required")
	validationErr.Suggestions = []string{"Provide a valid branch name"}

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Provide a valid branch name")
}

func TestErrorFormatter_FormatValidationErrorWithContext(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "ProjectName", "", "project name required when not in project context")
	validationErr.Suggestions = []string{"Current directory: /home/user/random-dir"}

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project name required when not in project context")
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Current directory: /home/user/random-dir")
}

func TestErrorFormatter_FormatValidationErrorMultipleSuggestions(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required")
	validationErr.Suggestions = []string{
		"Provide a valid branch name",
		"Branch names should follow git naming conventions",
	}

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.Contains(t, output, "Hint: Provide a valid branch name")
	assert.Contains(t, output, "Hint: Branch names should follow git naming conventions")
}

func TestErrorFormatter_FormatShellSubtypes(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		contain string
	}{
		{"ShellAlreadyInstalledError", core.NewShellAlreadyInstalledError("bash", "ctx", nil), "shell wrapper already installed"},
		{"ShellNotInstalledError", core.NewShellNotInstalledError("bash", "ctx", nil), "shell wrapper not installed"},
		{"ShellInvalidTypeError", core.NewShellInvalidTypeError("powershell", "ctx", nil), "invalid shell type"},
		{"ShellInferenceError", core.NewShellInferenceError("fish", "ctx", nil), "could not infer shell type"},
		{"ShellDetectionError", core.NewShellDetectionError("ctx", nil), "shell auto-detection failed"},
		{"ShellWrapperError installation", core.NewShellWrapperError("bash", "installation", "ctx", nil), "wrapper installation failed"},
		{"ShellWrapperError generation", core.NewShellWrapperError("bash", "generation", "ctx", nil), "wrapper generation failed"},
		{"ShellConfigError", core.NewShellConfigError("/home/u/.bashrc", "ctx", nil), "config file error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewErrorFormatter().Format(tt.err)
			assert.Contains(t, output, "Error:")
			assert.Contains(t, output, tt.contain)
		})
	}
}

func TestErrorFormatter_FormatWorktreeServiceErrorWithHint(t *testing.T) {
	err := core.NewWorktreeServiceError("/path/to/worktree", "feature-branch", "ResolvePath", "operation failed", core.ErrWorktreeNotFound)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "Hint: Use 'twiggit list' to see available worktrees")
}

func TestErrorFormatter_FormatProjectServiceErrorWithHint(t *testing.T) {
	err := core.NewProjectServiceError("nonexistent-project", "", "DiscoverProject", "project not found", core.ErrProjectNotFound)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project not found")
	assert.Contains(t, output, "Hint: Use 'twiggit list --all' to see available projects")
}

func TestErrorFormatter_FormatGitRepositoryErrorWithHint(t *testing.T) {
	err := core.NewGitRepositoryError("/path/to/repo", "failed to open", core.ErrGitRepoNotFound)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "Hint: Verify the repository path")
}

func TestErrorFormatter_FormatNavigationServiceErrorWithHint(t *testing.T) {
	err := core.NewNavigationServiceError("feature", "ctx", "ResolvePath", "target not found", core.ErrResolutionNotFound)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "Hint: Use 'twiggit list' to see available navigation targets")
}

func TestErrorFormatter_FormatGenericServiceError(t *testing.T) {
	genericErr := core.NewServiceError("ContextService", "GetCurrentContext", "failed to detect context", nil)
	output := NewErrorFormatter().Format(genericErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "failed to detect context")
}

func TestErrorFormatter_FormatResolutionError(t *testing.T) {
	resErr := core.NewResolutionError("target", "/path/to/project", "target not found", nil, nil)
	output := NewErrorFormatter().Format(resErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "target not found")
}

func TestErrorFormatter_FormatGitCommandError(t *testing.T) {
	err := core.NewGitCommandError("git", []string{"status"}, 1, "", "error output", "command failed", nil)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "command failed")
}

func TestErrorFormatter_FormatConflictError(t *testing.T) {
	err := core.NewConflictError("worktree", "feature", "CreateWorktree", "already exists", nil)
	output := NewErrorFormatter().Format(err)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "already exists")
}

func TestErrorFormatter_QuietModeSuppressesHints(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required")
	validationErr.Suggestions = []string{"Provide a valid branch name"}

	formatter := NewErrorFormatterWithOptions(true)
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.NotContains(t, output, "Hint:", "quiet mode should suppress hints")
}

func TestErrorFormatter_QuietModePreservesErrorMessage(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required")
	validationErr.Suggestions = []string{"Provide a valid branch name"}

	formatter := NewErrorFormatterWithOptions(true)
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "branch name is required")
}

func TestErrorFormatter_FormatGenericError(t *testing.T) {
	genericErr := errors.New("something went wrong")
	output := NewErrorFormatter().Format(genericErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "something went wrong")
}

func TestErrorFormatter_RegistrationCount(t *testing.T) {
	formatter := NewErrorFormatter()
	// 3 formatters: ValidationError, NotFoundError, OperationError.
	assert.Len(t, formatter.matchers, 3)
}

func TestErrorFormatter_SpecificityOrdering(t *testing.T) {
	formatter := NewErrorFormatter()
	// ValidationError is terminal; the wrapping OperationError wins.
	ve := core.NewOpValidationError("R", "f", "v", "m")
	wErr := core.NewWorktreeServiceError("/path", "b", "o", "msg", ve)
	output := formatter.Format(wErr)
	assert.Contains(t, output, "Hint: Use 'twiggit list' to see available worktrees")
}

func TestErrorFormatter_NotFoundHintDiscrimination(t *testing.T) {
	tests := []struct {
		name string
		err  error
		hint string
	}{
		{"project not found", core.NewProjectServiceError("p", "/p", "o", "m", core.ErrProjectNotFound), "Use 'twiggit list --all' to see available projects"},
		{"worktree not found", core.NewWorktreeServiceError("/p", "b", "o", "m", core.ErrWorktreeNotFound), "Use 'twiggit list' to see available worktrees"},
		{"resolution not found", core.NewNavigationServiceError("t", "c", "o", "m", core.ErrResolutionNotFound), "Use 'twiggit list' to see available navigation targets"},
		{"git repo not found", core.NewGitRepositoryError("/p", "m", core.ErrGitRepoNotFound), "Verify the repository path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewErrorFormatter().Format(tt.err)
			assert.Contains(t, output, tt.hint)
		})
	}
}

func TestAsType(t *testing.T) {
	t.Run("matches wrapped OperationError", func(t *testing.T) {
		wrapped := core.NewGitRepositoryError("/p", "m", errors.New("io"))
		err := error(wrapped)
		got, ok := asType[*core.OperationError](err)
		assert.True(t, ok)
		assert.Equal(t, wrapped, got)
	})
	t.Run("returns false on non-matching type", func(t *testing.T) {
		err := errors.New("plain")
		_, ok := asType[*core.OperationError](err)
		assert.False(t, ok)
	})
}

func TestFormatValidationError(t *testing.T) {
	validationErr := core.NewOpValidationError("CreateWorktreeRequest", "branch", "feature-1", "branch name cannot be empty")
	validationErr.Suggestions = []string{"Specify a branch name"}

	output := formatValidationError(validationErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "Hint:")
}

func TestFormatGenericError(t *testing.T) {
	formatter := NewErrorFormatter()
	genericErr := errors.New("something went wrong")

	output := formatter.formatGenericError(genericErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "something went wrong")
}
