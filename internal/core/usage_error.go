package core

// UsageError marks an invocation-level usage failure (invalid flag
// combination, missing required flag value, malformed argument) that
// the cmd layer maps to ExitCodeUsage (2).
//
// The legacy Err field is dropped per the cli-error-formatting spec:
// UsageError is terminal, no chain walking. Unwrap returns nil
// (typed-nil-safe per golang-safety).
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string { return e.Message }

// Unwrap returns nil: UsageError is terminal.
func (e *UsageError) Unwrap() error { return nil }

// NewUsageError constructs a UsageError from a message and optional
// underlying parser error. The error's text is appended to the message
// so callers retain context without a chain hop.
func NewUsageError(message string, err error) *UsageError {
	if err == nil {
		return &UsageError{Message: message}
	}
	return &UsageError{Message: message + ": " + err.Error()}
}

// UsageWrap wraps any error as a UsageError; nil returns nil. Used at
// the cmd boundary where cobra / pflag error types are translated into
// the core usage contract.
func UsageWrap(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{Message: err.Error()}
}
