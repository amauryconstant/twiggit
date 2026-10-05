package core

import "errors"

// Shell-error constructors. The five package-level sentinels
// (ErrShellAlreadyInstalled, ErrShellNotInstalled, ErrInvalidShellType,
// ErrInferenceFailed, ErrDetectionFailed) live in sentinels.go — single
// source of truth for the errors.Is membership catalog. Each
// constructor joins the sentinel into Cause so dispatch reaches the
// sentinel AND the originating error in one walk.

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
