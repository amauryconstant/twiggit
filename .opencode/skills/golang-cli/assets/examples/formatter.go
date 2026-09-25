// Path: internal/output/formatter.go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/you/myapp/internal/core"
)

// Formatter renders a result for --output. It takes an io.Writer, not
// IOStreams, so it stays free of terminal concerns: pass opts.IO.Out.
type Formatter interface {
	Write(w io.Writer, data any) error
}

// Tabular is implemented by results that render as columns (table, plain).
type Tabular interface {
	Header() []string
	Rows() [][]string
}

// NewFormatter maps an --output value to a Formatter. Unknown values are usage
// errors, so commands resolve the formatter before doing any work.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return JSONFormatter{}, nil
	case "table":
		return TableFormatter{}, nil
	case "plain":
		return PlainFormatter{}, nil
	default:
		return nil, &core.UsageError{Message: fmt.Sprintf("unknown --output %q: want json, table, or plain", format)}
	}
}

type JSONFormatter struct{}

func (JSONFormatter) Write(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// TableFormatter aligns columns with a header, for humans.
type TableFormatter struct{}

func (TableFormatter) Write(w io.Writer, data any) error {
	t, ok := data.(Tabular)
	if !ok {
		return fmt.Errorf("table output unsupported for %T", data)
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(t.Header(), "\t"))
	for _, row := range t.Rows() {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// PlainFormatter writes headerless TSV: stable input for cut, awk, and grep.
type PlainFormatter struct{}

func (PlainFormatter) Write(w io.Writer, data any) error {
	t, ok := data.(Tabular)
	if !ok {
		return fmt.Errorf("plain output unsupported for %T", data)
	}
	for _, row := range t.Rows() {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return nil
}
