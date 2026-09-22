package iostreams_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/iostreams"
)

func TestTest_ReturnsNonTTYStreams(t *testing.T) {
	t.Parallel()

	ios, in, out, errOut := iostreams.Test()

	require.NotNil(t, ios)
	assert.False(t, ios.IsStdoutTTY())
	assert.False(t, ios.IsStderrTTY())
	assert.False(t, ios.IsStdinTTY())
	assert.False(t, ios.ColorEnabled())
	assert.False(t, ios.IsInteractive())
	assert.False(t, ios.Quiet)
	assert.NotNil(t, ios.Logger)
	assert.Same(t, ios.Out, out)
	assert.NotNil(t, in)
	assert.NotNil(t, errOut)
}

func TestTest_BuffersAreBackingStreams(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	ios, in, out, errOut := iostreams.Test()

	_, err := io.WriteString(ios.Out, "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", out.String())

	_, err = io.WriteString(ios.ErrOut, "warn")
	require.NoError(t, err)
	assert.Equal(t, "warn", errOut.String())

	_, err = io.WriteString(in, "from-input")
	require.NoError(t, err)
	_ = ctx
}

func TestVerbosef_NoOpWhenVerboseFalse(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = false

	ios.Verbosef("should not appear %d", 42)

	assert.Empty(t, errOut.String())
}

func TestVerbosef_WritesDimLineWhenVerboseTrue(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	ios.Verbosef("step %d of %d", 2, 5)

	// colorEnabled false → Dim is identity, so raw line written.
	assert.Equal(t, "step 2 of 5\n", errOut.String())
}

func TestVerbosef_AppendsTrailingNewline(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	ios.Verbosef("no-newline")

	assert.True(t, bytes.HasSuffix(errOut.Bytes(), []byte("\n")))
}

func TestColorEnabled_ToggleViaAccessor(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	assert.False(t, ios.ColorEnabled())

	ios.SetColorEnabled(true)
	assert.True(t, ios.ColorEnabled())
	assert.NotEmpty(t, ios.Styles().Error("boom"))

	ios.SetColorEnabled(false)
	assert.False(t, ios.ColorEnabled())
	assert.Equal(t, "boom", ios.Styles().Error("boom"))
}

func TestNewLogger_DefaultLevelIsWarn(t *testing.T) {
	t.Setenv("TWIGGIT_DEBUG", "")
	logger := iostreams.NewLogger()
	require.NotNil(t, logger)
	assert.True(t, logger.Enabled(t.Context(), slog.LevelWarn))
	assert.False(t, logger.Enabled(t.Context(), slog.LevelInfo))
}

func TestNewLogger_DebugLevelWhenEnvSet(t *testing.T) {
	t.Setenv("TWIGGIT_DEBUG", "1")
	logger := iostreams.NewLogger()
	require.NotNil(t, logger)
	assert.True(t, logger.Enabled(t.Context(), slog.LevelDebug))
	assert.True(t, logger.Enabled(t.Context(), slog.LevelInfo))
}
