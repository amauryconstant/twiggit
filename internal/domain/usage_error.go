package domain

import "fmt"

// UsageError marks an invocation-level usage failure (invalid flag
// combination, missing required flag value, malformed argument) that
// the cmd layer maps to ExitCodeUsage (2). It carries the underlying
// parser-side error so the wrapped chain stays inspectable through
// errors.Is / errors.As.
type UsageError struct {
	Message string
	Err     error
}

func (e *UsageError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

func (e *UsageError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *UsageError) Is(target error) bool {
	return target == ErrUsageFlag
}

// NewUsageError constructs a UsageError with the given message and an
// optional underlying parser error.
func NewUsageError(message string, err error) *UsageError {
	return &UsageError{Message: message, Err: err}
}

// UsageWrap wraps any error as a UsageError; nil returns nil. Use at
// the cmd boundary where cobra/pflag error types are translated into
// the domain-level usage contract.
func UsageWrap(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{Err: err}
}
