package domain

import (
	"strings"
)

type shellErrorBase struct {
	ShellType string
	Context   string
	Err       error
}

func (b shellErrorBase) errorMessage(prefix string) string {
	parts := make([]string, 0, 4)
	if prefix != "" {
		parts = append(parts, prefix)
	}
	if b.ShellType != "" {
		parts = append(parts, b.ShellType)
	}
	if b.Context != "" {
		parts = append(parts, b.Context)
	}
	msg := strings.Join(parts, " ")
	if b.Err != nil {
		msg = msg + ": " + b.Err.Error()
	}
	return msg
}

func (b shellErrorBase) Unwrap() error { return b.Err }

// ShellAlreadyInstalledError reports that a shell wrapper is already installed.
type ShellAlreadyInstalledError struct {
	shellErrorBase
}

func (e *ShellAlreadyInstalledError) Error() string {
	return e.errorMessage("shell wrapper already installed for")
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellAlreadyInstalledError) Is(target error) bool {
	return target == ErrShellAlreadyInstalled
}

// NewShellAlreadyInstalledError constructs a ShellAlreadyInstalledError with the given fields.
func NewShellAlreadyInstalledError(shellType, context string, err error) *ShellAlreadyInstalledError {
	return &ShellAlreadyInstalledError{shellErrorBase: shellErrorBase{ShellType: shellType, Context: context, Err: err}}
}

// ShellNotInstalledError reports that a shell wrapper is not installed.
type ShellNotInstalledError struct {
	shellErrorBase
}

func (e *ShellNotInstalledError) Error() string {
	return e.errorMessage("shell wrapper not installed for")
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellNotInstalledError) Is(target error) bool {
	return target == ErrShellNotInstalled
}

// NewShellNotInstalledError constructs a ShellNotInstalledError with the given fields.
func NewShellNotInstalledError(shellType, context string, err error) *ShellNotInstalledError {
	return &ShellNotInstalledError{shellErrorBase: shellErrorBase{ShellType: shellType, Context: context, Err: err}}
}

// ShellInvalidTypeError reports an unsupported shell type.
type ShellInvalidTypeError struct {
	shellErrorBase
}

func (e *ShellInvalidTypeError) Error() string {
	return e.errorMessage("invalid shell type")
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellInvalidTypeError) Is(target error) bool {
	return target == ErrInvalidShellType
}

// NewShellInvalidTypeError constructs a ShellInvalidTypeError with the given fields.
func NewShellInvalidTypeError(shellType, context string, err error) *ShellInvalidTypeError {
	return &ShellInvalidTypeError{shellErrorBase: shellErrorBase{ShellType: shellType, Context: context, Err: err}}
}

// ShellInferenceError reports that inferring the shell type failed.
type ShellInferenceError struct {
	shellErrorBase
}

func (e *ShellInferenceError) Error() string {
	return e.errorMessage("could not infer shell type from")
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellInferenceError) Is(target error) bool {
	return target == ErrShellInferenceFailed
}

// NewShellInferenceError constructs a ShellInferenceError with the given fields.
func NewShellInferenceError(shellType, context string, err error) *ShellInferenceError {
	return &ShellInferenceError{shellErrorBase: shellErrorBase{ShellType: shellType, Context: context, Err: err}}
}

// ShellDetectionError reports that automatic shell detection failed.
type ShellDetectionError struct {
	shellErrorBase
}

func (e *ShellDetectionError) Error() string {
	return e.errorMessage("shell detection failed")
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellDetectionError) Is(target error) bool {
	return target == ErrShellDetectionFailed
}

// NewShellDetectionError constructs a ShellDetectionError with the given fields.
func NewShellDetectionError(context string, err error) *ShellDetectionError {
	return &ShellDetectionError{shellErrorBase: shellErrorBase{Context: context, Err: err}}
}

// ShellWrapperError reports a wrapper operation failure; op selects
// between generation and installation sentinels.
type ShellWrapperError struct {
	shellErrorBase
	Op string
}

func (e *ShellWrapperError) Error() string {
	prefix := "wrapper " + e.Op + " failed for"
	return e.errorMessage(prefix)
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellWrapperError) Is(target error) bool {
	switch e.Op {
	case "generation":
		return target == ErrWrapperGeneration
	case "installation":
		return target == ErrWrapperInstallation
	}
	return false
}

// NewShellWrapperError constructs a ShellWrapperError with the given fields.
func NewShellWrapperError(shellType, op, context string, err error) *ShellWrapperError {
	return &ShellWrapperError{shellErrorBase: shellErrorBase{ShellType: shellType, Context: context, Err: err}, Op: op}
}

// ShellConfigError reports a shell configuration file problem.
type ShellConfigError struct {
	shellErrorBase
	Path string
}

func (e *ShellConfigError) Error() string {
	prefix := "config file error"
	if e.Path != "" {
		prefix = "config file error at " + e.Path
	}
	return e.errorMessage(prefix)
}

// Is reports whether the wrapped sentinel matches.
func (e *ShellConfigError) Is(target error) bool {
	return target == ErrConfigFileNotFound
}

// NewShellConfigError constructs a ShellConfigError with the given fields.
func NewShellConfigError(path, context string, err error) *ShellConfigError {
	return &ShellConfigError{shellErrorBase: shellErrorBase{Context: context, Err: err}, Path: path}
}
