package core

// Demoted shell error constructors. Each previously concrete shell error
// type (ShellAlreadyInstalledError, ShellNotInstalledError,
// ShellInvalidTypeError, ShellInferenceError, ShellDetectionError,
// ShellWrapperError, ShellConfigError) collapses to OperationError;
// Op names the source. The legacy sentinels (ErrShellAlreadyInstalled,
// etc.) are deleted — callers that previously used errors.Is must now
// check oe.Op via errors.As(err, &oe) && oe.Op == "shell.x".

// NewShellAlreadyInstalledError demoted.
func NewShellAlreadyInstalledError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.already_installed",
		Entity:  shellType,
		Field:   context,
		Message: "shell wrapper already installed",
		Cause:   err,
	}
}

// NewShellNotInstalledError demoted.
func NewShellNotInstalledError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.not_installed",
		Entity:  shellType,
		Field:   context,
		Message: "shell wrapper not installed",
		Cause:   err,
	}
}

// NewShellInvalidTypeError demoted.
func NewShellInvalidTypeError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.invalid_type",
		Entity:  shellType,
		Field:   context,
		Message: "invalid shell type",
		Cause:   err,
	}
}

// NewShellInferenceError demoted.
func NewShellInferenceError(shellType, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.inference",
		Entity:  shellType,
		Field:   context,
		Message: "could not infer shell type",
		Cause:   err,
	}
}

// NewShellDetectionError demoted.
func NewShellDetectionError(context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.detection",
		Field:   context,
		Message: "shell detection failed",
		Cause:   err,
	}
}

// NewShellWrapperError demoted. The legacy Op selector (generation /
// installation) is encoded in the consolidated Op value.
func NewShellWrapperError(shellType, op, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.wrapper." + op,
		Entity:  shellType,
		Field:   context,
		Message: "wrapper " + op + " failed",
		Cause:   err,
	}
}

// NewShellConfigError demoted.
func NewShellConfigError(path, context string, err error) *OperationError {
	return &OperationError{
		Op:      "shell.config",
		Entity:  path,
		Field:   context,
		Message: "config file error",
		Cause:   err,
	}
}
