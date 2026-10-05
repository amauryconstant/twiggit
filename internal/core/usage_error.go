package core

// UsageError marks an invocation-level usage failure (invalid flag
// combination, missing required flag value, malformed argument) that
// the cmd layer maps to ExitCodeUsage (2).
//
// The Err field is intentionally not exposed (no public field per the
// cli-error-formatting spec); the hidden cause chain is preserved
// via Unwrap() so errors.Is walks through to the originating
// parser error.
type UsageError struct {
	Message string
	cause   error
}

// Error returns the constructed usage message. The Message field
// already carries the underlying parser error context appended at
// construction time, so no chain walking is required here.
func (e *UsageError) Error() string { return e.Message }

// Unwrap returns the hidden cause so errors.Is / errors.As walk
// through to the originating parser error. Returns nil when no
// cause was supplied at construction time.
func (e *UsageError) Unwrap() error { return e.cause }

// NewUsageError constructs a UsageError from a message and optional
// underlying parser error. The error's text is appended to the message
// so callers retain context without a chain hop; the cause is also
// stored on the hidden field so errors.Is reaches the parser error.
func NewUsageError(message string, err error) *UsageError {
	if err == nil {
		return &UsageError{Message: message}
	}
	return &UsageError{Message: message + ": " + err.Error(), cause: err}
}

// UsageWrap wraps any error as a UsageError; nil returns nil. Used at
// the cmd boundary where cobra / pflag error types are translated into
// the core usage contract. The original error is preserved as the
// hidden cause so errors.Is / errors.As can walk to it.
func UsageWrap(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{Message: err.Error(), cause: err}
}
