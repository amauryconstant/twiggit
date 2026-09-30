package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"twiggit/internal/core"
)

// Formatter is the single-method contract all output formats
// satisfy. Implementations MUST write to the supplied writer,
// MUST NOT retain or close it, and SHOULD return an error only
// for unrecoverable I/O failures. The data argument is whatever
// the command wants rendered; each implementation decides what
// shapes it accepts (See: JSONFormatter / TableFormatter
// / PlainFormatter).
type Formatter interface {
	Write(w io.Writer, data any) error
}

// Tabular is the projection contract for `table` and `plain`
// formats. A command supplies an adapter that produces the
// header row and the per-record rows; both formatters consume
// the same interface so the projection lives in one place per
// command.
type Tabular interface {
	Header() []string
	Rows() [][]string
}

// Format names recognized by NewFormatter. Unknown or empty
// values resolve to (nil, error) so the caller can fall through
// to per-command default rendering.
const (
	FormatJSON  = "json"
	FormatTable = "table"
	FormatPlain = "plain"
)

// JSONFormatter encodes data as a single JSON object using
// encoding/json. Slice / map / struct values are emitted as a
// single top-level array / object via the standard encoder
// behavior; primitives are encoded by value.
type JSONFormatter struct{}

// Write encodes data as JSON and writes it (with the encoder's
// trailing newline) to w.
func (JSONFormatter) Write(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// TableFormatter renders a Tabular projection as an aligned table
// via the existing RenderTable helper. The data argument MUST
// implement Tabular; a non-Tabular value produces a
// *core.UsageError so the caller learns to project first.
type TableFormatter struct{}

// Write renders the table to w. Returns *core.UsageError if
// data does not implement Tabular.
func (TableFormatter) Write(w io.Writer, data any) error {
	t, ok := data.(Tabular)
	if !ok {
		return core.NewUsageError(fmt.Sprintf("TableFormatter requires Tabular data, got %T", data), nil)
	}
	return RenderTable(w, t.Header(), t.Rows(), nil)
}

// PlainFormatter emits a headerless TSV rendering of a Tabular
// projection. Each row is written as tab-separated columns
// followed by a newline; the header row is intentionally omitted
// so plain output is grep-friendly and pipe-stable.
type PlainFormatter struct{}

// Write emits headerless TSV. Returns *core.UsageError if data
// does not implement Tabular.
func (PlainFormatter) Write(w io.Writer, data any) error {
	t, ok := data.(Tabular)
	if !ok {
		return core.NewUsageError(fmt.Sprintf("PlainFormatter requires Tabular data, got %T", data), nil)
	}
	for _, row := range t.Rows() {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return fmt.Errorf("write plain: %w", err)
		}
	}
	return nil
}

// NewFormatter resolves a format name to a Formatter and returns
// it alongside an error. The contract:
//
//   - "" → (nil, nil) so the caller can defer to its per-command
//     human default rather than globalising the choice here.
//   - "json" | "table" | "plain" → (formatter, nil).
//   - anything else (including the legacy "text") → (nil,
//     *core.UsageError) so the caller surfaces a Usage: prefix
//     with the recognised vocabulary inline.
//
// Callers should branch on the formatter == nil case to render
// their per-command default shape.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "":
		return nil, nil
	case FormatJSON:
		return JSONFormatter{}, nil
	case FormatTable:
		return TableFormatter{}, nil
	case FormatPlain:
		return PlainFormatter{}, nil
	default:
		return nil, core.NewUsageError(
			fmt.Sprintf("invalid output format '%s': must be 'json', 'table', or 'plain'", format),
			nil,
		)
	}
}
