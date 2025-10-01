package domain

import (
	"fmt"
)

// ContextDetectionError represents context detection errors
type ContextDetectionError struct {
	Path    string
	Cause   error
	Message string
}

func (e *ContextDetectionError) Error() string {
	return fmt.Sprintf("context detection failed for %s: %s", e.Path, e.Message)
}

func (e *ContextDetectionError) Unwrap() error {
	return e.Cause
}

// NewContextDetectionError creates a new context detection error
func NewContextDetectionError(path, message string, cause error) *ContextDetectionError {
	return &ContextDetectionError{
		Path:    path,
		Cause:   cause,
		Message: message,
	}
}
