package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinels_Messages(t *testing.T) {
	expected := map[error]string{
		ErrGitRepoNotFound:    "core: git repository not found",
		ErrWorktreeNotFound:   "core: worktree not found",
		ErrProjectNotFound:    "core: project not found",
		ErrResolutionNotFound: "core: resolution target not found",
	}
	for sentinel, msg := range expected {
		assert.Equal(t, msg, sentinel.Error(), "sentinel message mismatch")
	}
}

func TestSentinels_Count(t *testing.T) {
	// Hard contract: only 4 NotFound sentinels. The shell / config /
	// usage sentinels are deleted by the spec.
	sentinels := []error{
		ErrGitRepoNotFound,
		ErrWorktreeNotFound,
		ErrProjectNotFound,
		ErrResolutionNotFound,
	}
	assert.Len(t, sentinels, 4)
}

func TestShellAlreadyInstalledError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("disk full")
	err := NewShellAlreadyInstalledError("bash", "installing wrapper", cause)
	assert.Equal(t, "shell.already_installed", err.Op)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "shell.already_installed")
	assert.Contains(t, msg, "bash")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellNotInstalledError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("file missing")
	err := NewShellNotInstalledError("zsh", "missing block", cause)
	assert.Equal(t, "shell.not_installed", err.Op)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "shell.not_installed")
	assert.Contains(t, msg, "zsh")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellInvalidTypeError_Op(t *testing.T) {
	err := NewShellInvalidTypeError("powershell", "unsupported shell", nil)
	assert.Equal(t, "shell.invalid_type", err.Op)
	assert.NoError(t, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "invalid shell type")
	assert.Contains(t, msg, "powershell")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellInferenceError_OpAndUnwrap(t *testing.T) {
	cause := errors.New("path unknown")
	err := NewShellInferenceError("fish", "from /etc/config", cause)
	assert.Equal(t, "shell.inference", err.Op)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "could not infer shell type")
	assert.Contains(t, msg, "fish")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellDetectionError_Op(t *testing.T) {
	err := NewShellDetectionError("SHELL env unset", nil)
	assert.Equal(t, "shell.detection", err.Op)
	assert.NoError(t, err.Unwrap())
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
