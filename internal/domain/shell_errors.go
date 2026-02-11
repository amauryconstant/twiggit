package domain

import "fmt"

// Shell error types
const (
	// ErrInvalidShellType indicates an unsupported shell type
	ErrInvalidShellType = "INVALID_SHELL_TYPE"

	// ErrShellNotInstalled indicates shell wrapper is not installed
	ErrShellNotInstalled = "SHELL_NOT_INSTALLED"

	// ErrShellAlreadyInstalled indicates shell wrapper is already installed
	ErrShellAlreadyInstalled = "SHELL_ALREADY_INSTALLED"

	// ErrConfigFileNotFound indicates shell config file was not found
	ErrConfigFileNotFound = "CONFIG_FILE_NOT_FOUND"

	// ErrConfigFileNotWritable indicates shell config file is not writable
	ErrConfigFileNotWritable = "CONFIG_FILE_NOT_WRITABLE"

	// ErrWrapperGeneration indicates wrapper generation failed
	ErrWrapperGeneration = "WRAPPER_GENERATION_FAILED"

	// ErrWrapperInstallation indicates wrapper installation failed
	ErrWrapperInstallation = "WRAPPER_INSTALLATION_FAILED"

	// ErrInferenceFailed indicates shell type inference from path failed
	ErrInferenceFailed = "INFERENCE_FAILED"

	// ErrShellDetectionFailed indicates automatic shell detection failed
	ErrShellDetectionFailed = "SHELL_DETECTION_FAILED"
)

// ShellError represents a shell service error with context
type ShellError struct {
	Code      string
	ShellType string
	Context   string
	Cause     error
}

// Error implements the error interface
func (e *ShellError) Error() string {
	if e.Context != "" {
		if e.Cause != nil {
			return fmt.Sprintf("%s: %s", e.Context, e.Cause.Error())
		}
		if e.ShellType != "" {
			return fmt.Sprintf("%s: %s", e.Context, e.ShellType)
		}
		return e.Context
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.ShellType != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.ShellType)
	}
	return fmt.Sprintf("shell error [%s]", e.Code)
}

// Unwrap returns the underlying cause
func (e *ShellError) Unwrap() error {
	return e.Cause
}

// NewShellError creates a new shell error
func NewShellError(code, shellType, context string) *ShellError {
	return &ShellError{
		Code:      code,
		ShellType: shellType,
		Context:   context,
	}
}

// NewShellErrorWithCause creates a new shell error with an underlying cause
func NewShellErrorWithCause(code, shellType, context string, cause error) *ShellError {
	return &ShellError{
		Code:      code,
		ShellType: shellType,
		Context:   context,
		Cause:     cause,
	}
}
