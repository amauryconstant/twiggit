package domain

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinels_Messages(t *testing.T) {
	expected := map[error]string{
		ErrGitRepoNotFound:       "domain: git repository not found",
		ErrWorktreeNotFound:      "domain: worktree not found",
		ErrProjectNotFound:       "domain: project not found",
		ErrResolutionNotFound:    "domain: resolution target not found",
		ErrShellAlreadyInstalled: "domain: shell wrapper already installed",
		ErrShellNotInstalled:     "domain: shell wrapper not installed",
		ErrInvalidShellType:      "domain: invalid shell type",
		ErrShellInferenceFailed:  "domain: could not infer shell type",
		ErrShellDetectionFailed:  "domain: shell detection failed",
		ErrWrapperGeneration:     "domain: wrapper generation failed",
		ErrWrapperInstallation:   "domain: wrapper installation failed",
		ErrConfigFileNotFound:    "domain: config file not found",
	}
	for sentinel, msg := range expected {
		assert.Equal(t, msg, sentinel.Error(), "sentinel message mismatch")
	}
}

func TestShellAlreadyInstalledError_IsAndUnwrap(t *testing.T) {
	cause := errors.New("disk full")
	err := NewShellAlreadyInstalledError("bash", "installing wrapper", cause)
	assert.ErrorIs(t, err, ErrShellAlreadyInstalled)
	assert.NotErrorIs(t, err, ErrShellNotInstalled)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "shell wrapper already installed")
	assert.Contains(t, msg, "bash")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellNotInstalledError_IsAndUnwrap(t *testing.T) {
	cause := errors.New("file missing")
	err := NewShellNotInstalledError("zsh", "missing block", cause)
	assert.ErrorIs(t, err, ErrShellNotInstalled)
	assert.NotErrorIs(t, err, ErrShellAlreadyInstalled)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "shell wrapper not installed")
	assert.Contains(t, msg, "zsh")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellInvalidTypeError_IsAndUnwrap(t *testing.T) {
	err := NewShellInvalidTypeError("powershell", "unsupported shell", nil)
	assert.ErrorIs(t, err, ErrInvalidShellType)
	assert.NotErrorIs(t, err, ErrShellNotInstalled)
	require.NoError(t, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "invalid shell type")
	assert.Contains(t, msg, "powershell")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellInferenceError_IsAndUnwrap(t *testing.T) {
	cause := errors.New("path unknown")
	err := NewShellInferenceError("fish", "from /etc/config", cause)
	assert.ErrorIs(t, err, ErrShellInferenceFailed)
	assert.NotErrorIs(t, err, ErrShellDetectionFailed)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "could not infer shell type")
	assert.Contains(t, msg, "fish")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellDetectionError_IsAndUnwrap(t *testing.T) {
	err := NewShellDetectionError("SHELL env unset", nil)
	assert.ErrorIs(t, err, ErrShellDetectionFailed)
	assert.NotErrorIs(t, err, ErrShellInferenceFailed)
	require.NoError(t, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "shell detection failed")
	assert.Contains(t, msg, "SHELL env unset")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellWrapperError_GenerationAndInstallation(t *testing.T) {
	cause := errors.New("template broken")
	genErr := NewShellWrapperError("bash", "generation", "compose failed", cause)
	assert.ErrorIs(t, genErr, ErrWrapperGeneration)
	assert.NotErrorIs(t, genErr, ErrWrapperInstallation)
	assert.Equal(t, cause, genErr.Unwrap())

	instErr := NewShellWrapperError("zsh", "installation", "write failed", nil)
	assert.ErrorIs(t, instErr, ErrWrapperInstallation)
	assert.NotErrorIs(t, instErr, ErrWrapperGeneration)
	require.NoError(t, instErr.Unwrap())

	unknown := NewShellWrapperError("fish", "other", "", nil)
	assert.NotErrorIs(t, unknown, ErrWrapperGeneration)
	assert.NotErrorIs(t, unknown, ErrWrapperInstallation)
}

func TestShellConfigError_IsAndUnwrap(t *testing.T) {
	cause := errors.New("permission denied")
	err := NewShellConfigError("/home/u/.bashrc", "cannot write", cause)
	assert.ErrorIs(t, err, ErrConfigFileNotFound)
	assert.NotErrorIs(t, err, ErrShellNotInstalled)
	assert.Equal(t, cause, err.Unwrap())
	msg := err.Error()
	assert.Contains(t, msg, "config file error")
	assert.Contains(t, msg, "/home/u/.bashrc")
	assert.NotRegexp(t, `\.$`, msg)
}

func TestShellErrors_NoEmojiOrSentinelCode(t *testing.T) {
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
			require.NotContains(t, strings.TrimSpace(msg), "💡", "must not contain emoji")
			for _, code := range []string{"SHELL_", "INVALID_", "CONFIG_", "WRAPPER_", "INFERENCE_"} {
				assert.NotContains(t, strings.ToUpper(msg), code, "must not expose sentinel code")
			}
		})
	}
}

// TestNoShellErrorInterface guards against reintroduction of a single
// `ShellError` interface. The domain-typed-errors spec mandates seven concrete
// subtypes sharing a private `shellErrorBase` struct; an interface would undo
// the explicit strategy-pattern dispatch in cmd/error_formatter.go and break
// the Is(target) bool matchers. Static source scan via go/parser — no
// reflection on runtime symbols required.
func TestNoShellErrorInterface(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "shell_errors.go", nil, 0)
	require.NoError(t, err)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Name.Name != "ShellError" {
				continue
			}
			if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
				t.Errorf("type ShellError interface must not exist in internal/domain; " +
					"use the seven concrete subtypes (ShellAlreadyInstalledError, etc.)")
			}
		}
	}
}
