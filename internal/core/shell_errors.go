package core

import "errors"

// Sentinel errors matched by callers via errors.Is per
// domain-typed-errors req 8. Restored after the prior archive; the
// seven shell-error subtype structs (ShellAlreadyInstalledError,
// etc.) stay removed per design Decision 4.
var (
	ErrShellAlreadyInstalled = errors.New("core: shell wrapper already installed")
	ErrShellNotInstalled     = errors.New("core: shell wrapper not installed")
	ErrInvalidShellType      = errors.New("core: invalid shell type")
	ErrInferenceFailed       = errors.New("core: could not infer shell type")
	ErrDetectionFailed       = errors.New("core: shell detection failed")
)

// Demoted shell error constructors. Each previously concrete shell error
// type (ShellAlreadyInstalledError, ShellNotInstalledError,
// ShellInvalidTypeError, ShellInferenceError, ShellDetectionError,
// ShellWrapperError, ShellConfigError) collapses to OperationError;
// Op names the source. Cause carries the matching sentinel joined
// with the caller's err so errors.Is reaches both the sentinel
// (for dispatch) and the originating error (for diagnostic chains).

// NewShellAlreadyInstalledError demoted.
func NewShellAlreadyInstalledError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.already_installed",
		Entity:  shellType,
		Field:   context,
		Message: "shell wrapper already installed",
		Cause:   errors.Join(ErrShellAlreadyInstalled, err),
	}
}

// NewShellNotInstalledError demoted.
func NewShellNotInstalledError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.not_installed",
		Entity:  shellType,
		Field:   context,
		Message: "shell wrapper not installed",
		Cause:   errors.Join(ErrShellNotInstalled, err),
	}
}

// NewShellInvalidTypeError demoted.
func NewShellInvalidTypeError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.invalid_type",
		Entity:  shellType,
		Field:   context,
		Message: "invalid shell type",
		Cause:   errors.Join(ErrInvalidShellType, err),
	}
}

// NewShellInferenceError demoted.
func NewShellInferenceError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.inference",
		Entity:  shellType,
		Field:   context,
		Message: "could not infer shell type",
		Cause:   errors.Join(ErrInferenceFailed, err),
	}
}

// NewShellDetectionError demoted.
func NewShellDetectionError(context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.detection",
		Field:   context,
		Message: "shell auto-detection failed",
		Cause:   errors.Join(ErrDetectionFailed, err),
	}
}

// NewShellWrapperError demoted. The legacy Op selector (generation /
// installation) is encoded in the consolidated Op value. No sentinel
// join: this path does not have a matching sentinel under req 8.
func NewShellWrapperError(shellType, op, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.wrapper." + op,
		Entity:  shellType,
		Field:   context,
		Message: "wrapper " + op + " failed",
		Cause:   err,
	}
}

// NewShellConfigError demoted. No sentinel join: no matching sentinel
// under req 8.
func NewShellConfigError(path, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.config",
		Entity:  path,
		Field:   context,
		Message: "config file error",
		Cause:   err,
	}
}
