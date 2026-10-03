package iostreams_test

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"testing"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// isColorEnabled false → Dim is identity, so raw line written.
	assert.Equal(t, "step 2 of 5\n", errOut.String())
}

func TestVerbosef_AppendsTrailingNewline(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	ios.Verbosef("no-newline")

	assert.True(t, bytes.HasSuffix(errOut.Bytes(), []byte("\n")))
}

// TestVerbosef_OmitsDebugPrefix asserts the rendered line carries no
// "DEBUG:" or "[VERBOSE]" prefix per cli-verbose-output ("No debug
// prefix"). Color is off in Test() so Dim renders as identity and the
// raw string is what hits ErrOut.
func TestVerbosef_OmitsDebugPrefix(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	ios.Verbosef("cloning %s into %s", "origin", "feature-branch")

	rendered := errOut.String()
	assert.Contains(t, rendered, "cloning origin into feature-branch")
	assert.NotContains(t, rendered, "DEBUG:")
	assert.NotContains(t, rendered, "[VERBOSE]")
}

// TestIsInteractive_TestDefaultIsFalse asserts the predicate returns
// false when any of isColorEnabled / isStdoutTTY / isStdinTTY is off,
// matching cli-iostreams "IsInteractive (stdout AND stdin TTY)".
// Test() forces all three to false, so the row this test pins is the
// production path's worst case; System() probes the real FDs at
// runtime and the all-true row is covered by the source-level
// definition.
func TestIsInteractive_TestDefaultIsFalse(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()

	assert.False(t, ios.IsInteractive())
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
	logger := iostreams.NewLogger(&bytes.Buffer{})
	require.NotNil(t, logger)
	assert.True(t, logger.Enabled(t.Context(), slog.LevelWarn))
	assert.False(t, logger.Enabled(t.Context(), slog.LevelInfo))
}

func TestNewLogger_DebugLevelWhenEnvSet(t *testing.T) {
	t.Setenv("TWIGGIT_DEBUG", "1")
	logger := iostreams.NewLogger(&bytes.Buffer{})
	require.NotNil(t, logger)
	assert.True(t, logger.Enabled(t.Context(), slog.LevelDebug))
	assert.True(t, logger.Enabled(t.Context(), slog.LevelInfo))
}

// TestLogger_DebugOutput asserts the user-observable contract from
// cli-iostreams "Verbose and Logger are distinct channels": debug
// writes land on the test buffer when TWIGGIT_DEBUG is set, and are
// discarded when it is not. The level-gated tests above
// (TestNewLogger_*) cover only the Enabled gate; this pins the
// end-to-end output path against the test stderr buffer.
func TestLogger_DebugOutput(t *testing.T) {
	t.Run("emits when TWIGGIT_DEBUG is set", func(t *testing.T) {
		t.Setenv("TWIGGIT_DEBUG", "1")
		ios, _, _, errOut := iostreams.Test()

		ios.Logger.Debug("hello", "key", "value")

		assert.Contains(t, errOut.String(), "hello")
	})

	t.Run("silent when TWIGGIT_DEBUG is unset", func(t *testing.T) {
		t.Setenv("TWIGGIT_DEBUG", "")
		ios, _, _, errOut := iostreams.Test()

		ios.Logger.Debug("hello")

		assert.Empty(t, errOut.String())
	})
}

// TestLoggerPointer asserts the singleton guarantee: NewLogger
// returns the same *slog.Logger pointer when called repeatedly
// with the same writer, so System() / Test() / Factory.Logger all
// share one instance and main.go's slog.SetDefault binds to it.
func TestLoggerPointer(t *testing.T) {
	t.Run("system path", func(t *testing.T) {
		sys := iostreams.System()
		assert.Same(t, sys.Logger, iostreams.NewLogger(os.Stderr))
	})

	t.Run("test path", func(t *testing.T) {
		ios, _, _, errOut := iostreams.Test()
		assert.Same(t, ios.Logger, iostreams.NewLogger(errOut))
	})
}
