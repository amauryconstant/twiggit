package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellError_Error_WithContextAndCause(t *testing.T) {
	cause := errors.New("file not found")
	err := NewShellErrorWithCause(ErrConfigFileNotFound, "bash", "wrapper installation", cause)
	msg := err.Error()

	assert.Contains(t, msg, "wrapper installation")
	assert.Contains(t, msg, "file not found")
}

func TestShellError_Error_WithContextAndShellType(t *testing.T) {
	err := NewShellError(ErrInvalidShellType, "invalid", "shell detection")
	msg := err.Error()

	assert.Contains(t, msg, "shell detection")
	assert.Contains(t, msg, "invalid")
}

func TestShellError_Error_OnlyContext(t *testing.T) {
	err := NewShellError(ErrInferenceFailed, "", "could not infer shell type")
	msg := err.Error()

	assert.Equal(t, "could not infer shell type", msg)
}

func TestShellError_Error_OnlyCause(t *testing.T) {
	cause := errors.New("permission denied")
	err := NewShellErrorWithCause(ErrConfigFileNotWritable, "", "", cause)
	msg := err.Error()

	assert.Equal(t, "permission denied", msg)
}

func TestShellError_Error_OnlyShellType(t *testing.T) {
	err := NewShellError(ErrInvalidShellType, "fish", "")
	msg := err.Error()

	assert.Contains(t, msg, ErrInvalidShellType)
	assert.Contains(t, msg, "fish")
	assert.Equal(t, ErrInvalidShellType+": fish", msg)
}

func TestShellError_Error_OnlyCode(t *testing.T) {
	err := NewShellError(ErrWrapperGeneration, "", "")
	msg := err.Error()

	assert.Contains(t, msg, ErrWrapperGeneration)
	assert.Equal(t, fmt.Sprintf("shell error [%s]", ErrWrapperGeneration), msg)
}

func TestShellError_Unwrap(t *testing.T) {
	t.Run("with cause returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewShellErrorWithCause(ErrConfigFileNotFound, "bash", "context", cause)
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("without cause returns nil", func(t *testing.T) {
		err := NewShellError(ErrInvalidShellType, "bash", "context")
		assert.NoError(t, err.Unwrap())
	})
}

func TestShellError_ConstantValues(t *testing.T) {
	assert.Equal(t, "INVALID_SHELL_TYPE", ErrInvalidShellType)
	assert.Equal(t, "SHELL_NOT_INSTALLED", ErrShellNotInstalled)
	assert.Equal(t, "SHELL_ALREADY_INSTALLED", ErrShellAlreadyInstalled)
	assert.Equal(t, "CONFIG_FILE_NOT_FOUND", ErrConfigFileNotFound)
	assert.Equal(t, "CONFIG_FILE_NOT_WRITABLE", ErrConfigFileNotWritable)
	assert.Equal(t, "WRAPPER_GENERATION_FAILED", ErrWrapperGeneration)
	assert.Equal(t, "WRAPPER_INSTALLATION_FAILED", ErrWrapperInstallation)
	assert.Equal(t, "INFERENCE_FAILED", ErrInferenceFailed)
	assert.Equal(t, "SHELL_DETECTION_FAILED", ErrShellDetectionFailed)
}

func TestShellError_ComplexScenarios(t *testing.T) {
	t.Run("full error with all fields", func(t *testing.T) {
		cause := errors.New("disk full")
		err := NewShellErrorWithCause(ErrWrapperInstallation, "zsh", "installing wrapper", cause)
		msg := err.Error()

		assert.Contains(t, msg, "installing wrapper")
		assert.Contains(t, msg, "disk full")
		assert.Equal(t, "zsh", err.ShellType)
		assert.Equal(t, ErrWrapperInstallation, err.Code)
		assert.Equal(t, "installing wrapper", err.Context)
	})

	t.Run("error with wrapped ShellError", func(t *testing.T) {
		innerCause := errors.New("permission denied")
		innerErr := NewShellErrorWithCause(ErrConfigFileNotWritable, "bash", "writing config", innerCause)
		outerCause := innerErr
		outerErr := NewShellErrorWithCause(ErrWrapperInstallation, "bash", "wrapper installation", outerCause)
		msg := outerErr.Error()

		assert.Contains(t, msg, "wrapper installation")
		assert.Contains(t, msg, "writing config")
	})
}

func TestShellError_NewShellError_CreatesWithoutCause(t *testing.T) {
	err := NewShellError(ErrInvalidShellType, "bash", "shell detection")

	assert.Equal(t, ErrInvalidShellType, err.Code)
	assert.Equal(t, "bash", err.ShellType)
	assert.Equal(t, "shell detection", err.Context)
	assert.NoError(t, err.Cause)
}

func TestShellError_NewShellErrorWithCause_CreatesWithCause(t *testing.T) {
	cause := errors.New("test error")
	err := NewShellErrorWithCause(ErrConfigFileNotFound, "zsh", "file lookup", cause)

	assert.Equal(t, ErrConfigFileNotFound, err.Code)
	assert.Equal(t, "zsh", err.ShellType)
	assert.Equal(t, "file lookup", err.Context)
	assert.Equal(t, cause, err.Cause)
}

func TestShellError_Error_AllBranches(t *testing.T) {
	testCases := []struct {
		name        string
		err         *ShellError
		contains    []string
		notContains []string
		expected    string
	}{
		{
			name:     "context + cause",
			err:      NewShellErrorWithCause("CODE", "bash", "context", errors.New("cause msg")),
			contains: []string{"context", "cause msg"},
			expected: "context: cause msg",
		},
		{
			name:     "context + shellType (no cause)",
			err:      NewShellError("CODE", "bash", "context"),
			contains: []string{"context", "bash"},
			expected: "context: bash",
		},
		{
			name:     "context only",
			err:      NewShellError("CODE", "", "context"),
			contains: []string{"context"},
			expected: "context",
		},
		{
			name:     "cause only (no context)",
			err:      NewShellErrorWithCause("CODE", "", "", errors.New("cause msg")),
			contains: []string{"cause msg"},
			expected: "cause msg",
		},
		{
			name:     "shellType only (no context)",
			err:      NewShellError("CODE", "bash", ""),
			contains: []string{"CODE", "bash"},
			expected: "CODE: bash",
		},
		{
			name:     "code only",
			err:      NewShellError("CODE", "", ""),
			contains: []string{"CODE"},
			expected: "shell error [CODE]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.err.Error()
			assert.Equal(t, tc.expected, msg)

			for _, contain := range tc.contains {
				assert.Contains(t, msg, contain)
			}
			for _, notContain := range tc.notContains {
				assert.NotContains(t, msg, notContain)
			}
		})
	}
}
