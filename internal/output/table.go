package output

import (
	"io"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"twiggit/internal/iostreams"
)

// RenderTable renders headers and rows as a lipgloss table to w.
// The rendered string is followed by a newline.
//
// When ios is non-nil and its color is disabled, the table
// drops its border decoration (boxes / horizontal lines) so the
// output stays terminal-clean for piping. Rows that are shorter
// than the header are right-padded with empty strings; rows
// that are longer than the header are truncated.
func RenderTable(w io.Writer, headers []string, rows [][]string, ios *iostreams.IOStreams) error {
	t := table.New().
		Headers(headers...).
		Rows(normalizeRows(headers, rows)...)
	if ios == nil || !ios.ColorEnabled() {
		t.Border(lipgloss.NormalBorder()).
			BorderTop(false).
			BorderBottom(false).
			BorderLeft(false).
			BorderRight(false).
			BorderHeader(false).
			BorderColumn(false).
			BorderRow(false)
	}
	out := t.Render()
	if out == "" {
		return nil
	}
	if _, err := io.WriteString(w, out); err != nil {
		return err
	}
	if !endsWithNewline(out) {
		_, err := io.WriteString(w, "\n")
		return err
	}
	return nil
}

func endsWithNewline(s string) bool { return len(s) > 0 && s[len(s)-1] == '\n' }

func normalizeRows(headers []string, rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}
	cols := len(headers)
	out := make([][]string, len(rows))
	for i, row := range rows {
		r := make([]string, cols)
		for j := 0; j < cols; j++ {
			if j < len(row) {
				r[j] = row[j]
			}
		}
		out[i] = r
	}
	return out
}
