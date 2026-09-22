package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// Formatter is the single-method contract all output formats
// satisfy. Implementations MUST write to the supplied writer,
// MUST NOT retain or close it, and SHOULD return an error only
// for unrecoverable I/O failures. The data argument is whatever
// the command wants rendered; each implementation decides what
// shapes it accepts (See: JSONFormatter / TableFormatter
// / PlainFormatter / JSONLinesFormatter).
type Formatter interface {
	Write(w io.Writer, data any) error
}

// Format names recognized by NewFormatter. Empty / unknown
// values yield a nil Formatter so the caller can fall through to
// human-readable rendering.
const (
	FormatJSON   = "json"
	FormatJSONL  = "jsonl"
	FormatTable  = "table"
	FormatPlain  = "plain"
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
	return enc.Encode(data)
}

// JSONLinesFormatter encodes one JSON object per line, no
// surrounding brackets. Slices render element-by-element; any
// other shape encodes the value once.
type JSONLinesFormatter struct{}

// Write streams one JSON-encoded value per line for slices
// (excluding []byte, which is treated as a primitive to avoid
// base64 noise); non-slices encode as a single line.
func (JSONLinesFormatter) Write(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice || v.Type().Elem().Kind() == reflect.Uint8 {
		return enc.Encode(data)
	}
	for i := range v.Len() {
		if err := enc.Encode(v.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

// TableFormatter renders tabular data with the supplied headers.
// Data MUST be [][]string where each inner slice has the same
// length as headers.
type TableFormatter struct {
	Headers []string
}

// Write renders the table to w. Returns an error if data is not
// [][]string.
func (f TableFormatter) Write(w io.Writer, data any) error {
	rows, ok := data.([][]string)
	if !ok {
		return fmt.Errorf("TableFormatter: data must be [][]string, got %T", data)
	}
	return RenderTable(w, f.Headers, rows, nil)
}

// PlainFormatter writes one item per line via fmt.Fprintln.
// Slices render element-by-element; everything else prints as a
// single line via fmt.Sprint. No decoration.
type PlainFormatter struct{}

// Write emits one printable line per slice element (or one
// line for non-slices).
func (PlainFormatter) Write(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		_, err := fmt.Fprintln(w, fmt.Sprint(data))
		return err
	}
	for i := range v.Len() {
		if _, err := fmt.Fprintln(w, fmt.Sprint(v.Index(i).Interface())); err != nil {
			return err
		}
	}
	return nil
}

// NewFormatter resolves a format name to a Formatter. Returns
// nil when format is empty or unrecognized so callers can fall
// through to human-readable rendering.
func NewFormatter(format string) Formatter {
	switch format {
	case FormatJSON:
		return JSONFormatter{}
	case FormatJSONL:
		return JSONLinesFormatter{}
	case FormatTable:
		return TableFormatter{}
	case FormatPlain:
		return PlainFormatter{}
	default:
		return nil
	}
}
