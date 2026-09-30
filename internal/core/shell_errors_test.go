package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinels_Messages(t *testing.T) {
	expected := map[error]string{
		ErrGitRepoNotFound:       "core: git repository not found",
		ErrWorktreeNotFound:      "core: worktree not found",
		ErrProjectNotFound:       "core: project not found",
		ErrResolutionNotFound:    "core: resolution target not found",
		ErrShellAlreadyInstalled: "core: shell wrapper already installed",
		ErrShellNotInstalled:     "core: shell wrapper not installed",
		ErrInvalidShellType:      "core: invalid shell type",
		ErrInferenceFailed:       "core: could not infer shell type",
		ErrDetectionFailed:       "core: shell detection failed",
	}
	for sentinel, msg := range expected {
		assert.Equal(t, msg, sentinel.Error(), "sentinel message mismatch")
	}
}

func TestSentinels_Count(t *testing.T) {
	sentinels := []error{
		ErrGitRepoNotFound,
		ErrWorktreeNotFound,
		ErrProjectNotFound,
		ErrResolutionNotFound,
		ErrShellAlreadyInstalled,
		ErrShellNotInstalled,
		ErrInvalidShellType,
		ErrInferenceFailed,
		ErrDetectionFailed,
	}
	assert.Len(t, sentinels, 9)
}

func TestShellAlreadyInstalledError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("disk full")
	err := NewShellAlreadyInstalledError("bash", "installing wrapper", cause)
	assert.Equal(t, "shell.already_installed", err.Op)
	assert.ErrorIs(t, err, ErrShellAlreadyInstalled)
	assert.ErrorIs(t, err, cause)
	msg := err.Error()
	assert.Contains(t, msg, "shell.already_installed")
	assert.Contains(t, msg, "bash")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellNotInstalledError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("file missing")
	err := NewShellNotInstalledError("zsh", "missing block", cause)
	assert.Equal(t, "shell.not_installed", err.Op)
	assert.ErrorIs(t, err, ErrShellNotInstalled)
	assert.ErrorIs(t, err, cause)
	msg := err.Error()
	assert.Contains(t, msg, "shell.not_installed")
	assert.Contains(t, msg, "zsh")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellInvalidTypeError_Op(t *testing.T) {
	err := NewShellInvalidTypeError("powershell", "unsupported shell", nil)
	assert.Equal(t, "shell.invalid_type", err.Op)
	assert.ErrorIs(t, err, ErrInvalidShellType)
	msg := err.Error()
	assert.Contains(t, msg, "invalid shell type")
	assert.Contains(t, msg, "powershell")
	assert.NotRegexp(t, `\.$`, msg)
	_ = require.NoError
}

func TestShellInferenceError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("path unknown")
	err := NewShellInferenceError("fish", "from /etc/config", cause)
	assert.Equal(t, "shell.inference", err.Op)
	assert.ErrorIs(t, err, ErrInferenceFailed)
	assert.ErrorIs(t, err, cause)
	msg := err.Error()
	assert.Contains(t, msg, "could not infer shell type")
	assert.Contains(t, msg, "fish")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellDetectionError_Op(t *testing.T) {
	err := NewShellDetectionError("SHELL env unset", nil)
	assert.Equal(t, "shell.detection", err.Op)
	assert.ErrorIs(t, err, ErrDetectionFailed)
	msg := err.Error()
	assert.Contains(t, msg, "shell.detection")
	assert.Contains(t, msg, "SHELL env unset")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellWrapperError_Ops(t *testing.T) {
	cause := errors.New("template broken")
	genErr := NewShellWrapperError("bash", "generation", "compose failed", cause)
	assert.Equal(t, "shell.wrapper.generation", genErr.Op)
	assert.Equal(t, cause, genErr.Unwrap())

	instErr := NewShellWrapperError("zsh", "installation", "write failed", nil)
	assert.Equal(t, "shell.wrapper.installation", instErr.Op)
	assert.NoError(t, instErr.Unwrap())
}

func TestShellConfigError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("permission denied")
	err := NewShellConfigError("/home/u/.bashrc", "cannot write", cause)
	assert.Equal(t, "shell.config", err.Op)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "config file error")
	assert.Contains(t, msg, "/home/u/.bashrc")
	assert.NotRegexp(t, `\.$`, msg)
}

// TestShellAlreadyInstalled_Is asserts the sentinel-match contract
// mandated by domain-typed-errors req 8 and design Decision 4:
// errors.Is(opErr, ErrShellAlreadyInstalled) must return true when
// the OperationError carries Op = "shell.already_installed" (whether
// the constructor joined the sentinel into Cause or the caller wraps
// further), and false for unrelated errors.
func TestShellAlreadyInstalled_Is(t *testing.T) {
	wrapped := NewShellAlreadyInstalledError("bash", "installing wrapper", errors.New("disk full"))
	require.ErrorIs(t, wrapped, ErrShellAlreadyInstalled)
	require.Contains(t, wrapped.Error(), "shell wrapper already installed")

	// Layered wrap should still match.
	doubled := fmt.Errorf("install bash wrapper: %w", wrapped)
	require.ErrorIs(t, doubled, ErrShellAlreadyInstalled)

	// Unrelated errors must not match.
	require.NotErrorIs(t, errors.New("nope"), ErrShellAlreadyInstalled)
	require.NotErrorIs(t, NewShellNotInstalledError("bash", "ctx", nil), ErrShellAlreadyInstalled)
}

func TestShellErrors_NoEmoji(t *testing.T) {
	cases := map[string]error{
		"already installed": NewShellAlreadyInstalledError("bash", "ctx", nil),
		"not installed":     NewShellNotInstalledError("bash", "ctx", nil),
		"invalid type":      NewShellInvalidTypeError("powershell", "ctx", nil),
		"inference":         NewShellInferenceError("fish", "ctx", nil),
		"detection":         NewShellDetectionError("ctx", nil),
		"wrapper gen":       NewShellWrapperError("bash", "generation", "ctx", nil),
		"wrapper inst":      NewShellWrapperError("zsh", "installation", "ctx", nil),
		"config":            NewShellConfigError("/p/.bashrc", "ctx", nil),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			msg := err.Error()
			assert.NotContains(t, strings.TrimSpace(msg), "💡", "must not contain emoji")
		})
	}
}
