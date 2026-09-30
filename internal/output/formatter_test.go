package output_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"twiggit/internal/core"
	"twiggit/internal/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type tabularAdapter struct {
	header []string
	rows   [][]string
}

func (t tabularAdapter) Header() []string { return t.header }
func (t tabularAdapter) Rows() [][]string { return t.rows }

func TestTableFormatter_Tabular(t *testing.T) {
	t.Parallel()

	data := tabularAdapter{
		header: []string{"BRANCH", "PATH"},
		rows:   [][]string{{"feat/foo", "/tmp/feat/foo"}, {"feat/bar", "/tmp/feat/bar"}},
	}
	var buf bytes.Buffer
	require.NoError(t, output.TableFormatter{}.Write(&buf, data))

	out := buf.String()
	assert.Contains(t, out, "BRANCH")
	assert.Contains(t, out, "feat/foo")
	assert.Contains(t, out, "feat/bar")
}

func TestPlainFormatter_TabularTSV(t *testing.T) {
	t.Parallel()

	data := tabularAdapter{
		header: []string{"BRANCH", "PATH"},
		rows:   [][]string{{"feat/foo", "/tmp/feat/foo"}, {"feat/bar", "/tmp/feat/bar"}},
	}
	var buf bytes.Buffer
	require.NoError(t, output.PlainFormatter{}.Write(&buf, data))

	out := buf.String()
	lines := bytes.Split(bytes.TrimRight(buf.Bytes(), "\n"), []byte("\n"))
	require.Len(t, lines, 2, "headerless: no header row emitted")
	for _, line := range lines {
		assert.Contains(t, string(line), "\t", "rows must be tab-separated")
	}
	assert.NotContains(t, out, "BRANCH\n", "header row must not be emitted")
}

func TestPlainFormatter_NonTabularReturnsUsageError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := output.PlainFormatter{}.Write(&buf, "raw string")
	require.Error(t, err)

	var ue *core.UsageError
	require.ErrorAs(t, err, &ue,
		"expected *core.UsageError, got %T", err)
}

func TestNewFormatter_EmptyReturnsNil(t *testing.T) {
	t.Parallel()

	f, err := output.NewFormatter("")
	assert.Nil(t, f)
	assert.NoError(t, err,
		"empty format must return (nil, nil) so commands can defer to their per-command default")
}

func TestNewFormatter_KnownNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		format string
		assert func(t *testing.T, f output.Formatter)
	}{
		{"json", output.FormatJSON, func(t *testing.T, f output.Formatter) {
			t.Helper()
			_, ok := f.(output.JSONFormatter)
			assert.True(t, ok)
		}},
		{"plain", output.FormatPlain, func(t *testing.T, f output.Formatter) {
			t.Helper()
			_, ok := f.(output.PlainFormatter)
			assert.True(t, ok)
		}},
		{"table", output.FormatTable, func(t *testing.T, f output.Formatter) {
			t.Helper()
			_, ok := f.(output.TableFormatter)
			assert.True(t, ok)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := output.NewFormatter(tc.format)
			require.NoError(t, err)
			require.NotNil(t, f)
			tc.assert(t, f)
		})
	}
}

func TestNewFormatter_UnknownReturnsUsageError(t *testing.T) {
	t.Parallel()

	f, err := output.NewFormatter("xml")
	assert.Nil(t, f)
	require.Error(t, err)

	var ue *core.UsageError
	require.ErrorAs(t, err, &ue,
		"unknown format must produce *core.UsageError, got %T", err)
	assert.Contains(t, ue.Message, "xml")
	assert.Contains(t, ue.Message, "'json', 'table', or 'plain'")
}

func TestNewFormatter_TextReturnsUsageError(t *testing.T) {
	t.Parallel()

	f, err := output.NewFormatter("text")
	assert.Nil(t, f)
	require.Error(t, err)

	var ue *core.UsageError
	require.ErrorAs(t, err, &ue,
		"legacy 'text' must NOT alias to plain; must produce *core.UsageError, got %T", err)
	assert.Contains(t, ue.Message, "text")
}
