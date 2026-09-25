// Path: internal/core/errors.go
//
// Semantic error types. They carry meaning and next steps, never exit codes or
// formatting: cmdutil.ExitCodeFor and output.FormatError map them at the edge.
package core

import "fmt"

// ValidationError: a domain value broke a rule (bad branch name, protected
// branch). A runtime failure, so exit 1.
type ValidationError struct {
	Field       string
	Value       string
	Message     string
	Suggestions []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Message)
}

type NotFoundError struct {
	Entity string
	Name   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Entity, e.Name)
}

// OperationError answers what happened (Message), why (Cause), and what to do
// next (Suggestions).
type OperationError struct {
	Op          string
	Message     string
	Cause       error
	Suggestions []string
}

func (e *OperationError) Error() string { return e.Message }
func (e *OperationError) Unwrap() error { return e.Cause }

// UsageError: the invocation itself was wrong (unknown flag or command, wrong
// argument count, missing input in a non-interactive run). The only type that
// exits 2.
type UsageError struct {
	Message string
	Cause   error
}

func (e *UsageError) Error() string { return e.Message }
func (e *UsageError) Unwrap() error { return e.Cause }
