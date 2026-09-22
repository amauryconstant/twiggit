package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/output"
)

type sample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONFormatter_WritesValidJSON(t *testing.T) {
	t.Parallel()

	data := sample{Name: "ada", Age: 36}
	var buf bytes.Buffer
	require.NoError(t, output.JSONFormatter{}.Write(&buf, data))

	var got sample
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got))
	assert.Equal(t, data, got)
}

func TestJSONFormatter_SliceYieldsJSONArray(t *testing.T) {
	t.Parallel()

	data := []sample{{Name: "a"}, {Name: "b"}}
	var buf bytes.Buffer
	require.NoError(t, output.JSONFormatter{}.Write(&buf, data))

	var got []sample
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got))
	assert.Equal(t, data, got)
}

func TestJSONLinesFormatter_OneObjectPerLine(t *testing.T) {
	t.Parallel()

	data := []sample{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	var buf bytes.Buffer
	require.NoError(t, output.JSONLinesFormatter{}.Write(&buf, data))

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 3, "each slice element must produce exactly one line")

	for i, line := range lines {
		var got sample
		require.NoError(t, json.Unmarshal([]byte(line), &got), "line %d invalid", i)
		assert.Equal(t, data[i], got)
	}
}

func TestJSONLinesFormatter_SingleValueYieldsOneLine(t *testing.T) {
	t.Parallel()

	data := sample{Name: "solo", Age: 1}
	var buf bytes.Buffer
	require.NoError(t, output.JSONLinesFormatter{}.Write(&buf, data))

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 1)
}

func TestJSONLinesFormatter_ByteSliceYieldsSingleLine(t *testing.T) {
	t.Parallel()

	// []byte is treated as primitive (one line, not element-by-element)
	// to avoid base64-streaming aliasing that would look like multiple
	// records when piped through jq/grep.
	var buf bytes.Buffer
	require.NoError(t, output.JSONLinesFormatter{}.Write(&buf, []byte("hi")))

	trimmed := strings.TrimRight(buf.String(), "\n")
	lines := strings.Split(trimmed, "\n")
	require.Len(t, lines, 1, "[]byte must produce exactly one line, got %d", len(lines))
	assert.Equal(t, `"aGk="`, lines[0], "encoding/json base64-encodes []byte")
}

func TestPlainFormatter_OneLinePerElement(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, output.PlainFormatter{}.Write(&buf, []string{"a", "b", "c"}))

	assert.Equal(t, "a\nb\nc\n", buf.String())
}

func TestPlainFormatter_SingleValue(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, output.PlainFormatter{}.Write(&buf, "hello"))

	assert.Equal(t, "hello\n", buf.String())
}

func TestTableFormatter_RejectsWrongType(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := output.TableFormatter{Headers: []string{"a"}}.Write(&buf, "not-a-table")
	require.Error(t, err)
}

func TestTableFormatter_EmptyRowsDoNotPanic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, output.TableFormatter{
		Headers: []string{"a", "b"},
	}.Write(&buf, [][]string{}))
}

func TestNewFormatter_KnownNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		format string
		assert func(t *testing.T, f output.Formatter)
	}{
		{"json", output.FormatJSON, func(t *testing.T, f output.Formatter) { _, ok := f.(output.JSONFormatter); assert.True(t, ok) }},
		{"jsonl", output.FormatJSONL, func(t *testing.T, f output.Formatter) { _, ok := f.(output.JSONLinesFormatter); assert.True(t, ok) }},
		{"plain", output.FormatPlain, func(t *testing.T, f output.Formatter) { _, ok := f.(output.PlainFormatter); assert.True(t, ok) }},
		{"table", output.FormatTable, func(t *testing.T, f output.Formatter) {
			_, ok := f.(output.TableFormatter)
			assert.True(t, ok)
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := output.NewFormatter(tc.format)
			require.NotNil(t, f)
			tc.assert(t, f)
		})
	}
}

func TestNewFormatter_EmptyOrUnknownReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, output.NewFormatter(""))
	assert.Nil(t, output.NewFormatter("yaml"))
}
