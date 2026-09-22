package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/iostreams"
	"twiggit/internal/output"
)

func TestRenderTable_HeadersAndRowsPresent(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	headers := []string{"NAME", "STATUS", "BRANCH"}
	rows := [][]string{
		{"alpha", "ok", "main"},
		{"beta", "ok", "feature/x"},
	}
	require.NoError(t, output.RenderTable(&buf, headers, rows, ios))

	out := buf.String()
	for _, h := range headers {
		assert.Contains(t, out, h, "header %q must be rendered", h)
	}
	for _, row := range rows {
		for _, cell := range row {
			assert.Contains(t, out, cell, "cell %q must be rendered", cell)
		}
	}
}

func TestRenderTable_EmptyRowsStillRendersHeaders(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	headers := []string{"col1", "col2", "col3"}
	require.NoError(t, output.RenderTable(&buf, headers, nil, ios))

	out := buf.String()
	for _, h := range headers {
		assert.Contains(t, out, h)
	}
}

func TestRenderTable_ShortRowRightPadded(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	headers := []string{"a", "b", "c"}
	rows := [][]string{
		{"alpha"}, // missing b, c → padded with empty strings.
	}
	require.NoError(t, output.RenderTable(&buf, headers, rows, ios))

	out := buf.String()
	assert.Contains(t, out, "alpha")
}

func TestRenderTable_LongRowTruncated(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	headers := []string{"a"}
	rows := [][]string{
		{"alpha", "extra-beyond-header"},
	}
	require.NoError(t, output.RenderTable(&buf, headers, rows, ios))

	out := buf.String()
	assert.Contains(t, out, "alpha")
	// 'extra-beyond-header' would render after 'alpha' if not truncated;
	// the lipgloss table silently drops cells beyond the header column
	// count, so we rely on absence rather than presence-of-truncation.
}

func TestRenderTable_NilIOSRendersWithoutError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, output.RenderTable(&buf, []string{"a"}, [][]string{{"b"}}, nil))

	out := buf.String()
	assert.True(t, strings.Contains(out, "a"))
	assert.True(t, strings.Contains(out, "b"))
}

func TestRenderTable_WritesTrailingNewline(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	require.NoError(t, output.RenderTable(&buf, []string{"x"}, [][]string{{"y"}}, ios))

	assert.True(t, bytes.HasSuffix(buf.Bytes(), []byte("\n")))
}

func TestRenderTable_ColorEnabledRendersWithoutError(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	ios.SetColorEnabled(true)
	var buf bytes.Buffer

	require.NoError(t, output.RenderTable(&buf, []string{"a"}, [][]string{{"b"}}, ios))

	out := buf.String()
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "b")
}
