// Package output owns the Respond concern of twiggit: multi-format
// rendering for command results and structured error formatting.
//
// The Formatter interface is single-method (Write(io.Writer, any) error) so
// new formats (json, table, plain) compose without coupling to command
// wiring. The Tabular projection is consumed by table and plain formatters;
// JSONFormatter encodes any shape. FormatError dispatches on the core.*
// error hierarchy via errors.As and renders per-type hint text.
package output
