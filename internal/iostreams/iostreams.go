package iostreams

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"golang.org/x/term"
)

var (
	loggerMu    sync.Mutex
	loggerCache = map[io.Writer]*slog.Logger{}
)

// IOStreams abstracts terminal I/O so commands never touch
// os.Stdout / os.Stderr directly. The struct carries In / Out /
// ErrOut streams, TTY detection state, the NO_COLOR-aware
// colorEnabled flag, a Quiet gate for --quiet, a Verbose gate
// for --verbose, and a Logger channel for TWIGGIT_DEBUG.
//
// System() and Test() constructors cover production and unit-test
// paths; Test() forces non-TTY / no-color / in-memory buffers.
type IOStreams struct {
	In           io.ReadCloser
	Out          io.Writer
	ErrOut       io.Writer
	colorEnabled bool
	isStdoutTTY  bool
	isStderrTTY  bool
	isStdinTTY   bool
	Quiet        bool
	Verbose      bool
	Logger       *slog.Logger
	styles       *Styles
}

// System constructs production-grade IOStreams backed by real
// stdin / stdout / stderr file descriptors. TTY state is detected
// via golang.org/x/term; colorEnabled is on only when stdout is
// a TTY and NO_COLOR is unset.
func System() *IOStreams {
	stdoutFD := int(os.Stdout.Fd()) // #nosec G115 -- file descriptors bounded by OS limit
	stderrFD := int(os.Stderr.Fd()) // #nosec G115 -- file descriptors bounded by OS limit
	stdinFD := int(os.Stdin.Fd())   // #nosec G115 -- file descriptors bounded by OS limit
	stdoutTTY := term.IsTerminal(stdoutFD)
	stderrTTY := term.IsTerminal(stderrFD)
	stdinTTY := term.IsTerminal(stdinFD)
	colorEnabled := stdoutTTY && os.Getenv("NO_COLOR") == ""
	return &IOStreams{
		In:           os.Stdin,
		Out:          os.Stdout,
		ErrOut:       os.Stderr,
		colorEnabled: colorEnabled,
		isStdoutTTY:  stdoutTTY,
		isStderrTTY:  stderrTTY,
		isStdinTTY:   stdinTTY,
		Logger:       NewLogger(os.Stderr),
		styles:       NewStyles(colorEnabled),
	}
}

// Test constructs deterministic, in-memory IOStreams for unit
// tests. Streams are non-TTY, color output is off, Quiet is off.
// The returned *bytes.Buffer values back In / Out / ErrOut so
// callers can read what commands wrote.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:           io.NopCloser(in),
		Out:          out,
		ErrOut:       errOut,
		colorEnabled: false,
		isStdoutTTY:  false,
		isStderrTTY:  false,
		isStdinTTY:   false,
		Logger:       NewLogger(errOut),
		styles:       NewStyles(false),
	}, in, out, errOut
}

// IsStdoutTTY reports whether the stdout stream is a terminal.
func (s *IOStreams) IsStdoutTTY() bool { return s.isStdoutTTY }

// IsStderrTTY reports whether the stderr stream is a terminal.
func (s *IOStreams) IsStderrTTY() bool { return s.isStderrTTY }

// IsStdinTTY reports whether the stdin stream is a terminal.
func (s *IOStreams) IsStdinTTY() bool { return s.isStdinTTY }

// ColorEnabled reports whether styled output should be rendered.
func (s *IOStreams) ColorEnabled() bool { return s.colorEnabled }

// IsInteractive reports whether the user is in an interactive
// session. Per spec 3.3: color on AND stdin + stdout TTY.
func (s *IOStreams) IsInteractive() bool {
	return s.colorEnabled && s.isStdinTTY && s.isStdoutTTY
}

// Verbosef writes a dim-styled line to ErrOut when Verbose is
// set. Explicit trailing newline is always appended, matching
// spec 3.3.
func (s *IOStreams) Verbosef(format string, args ...any) {
	if !s.Verbose {
		return
	}
	line := s.styles.Dim(fmt.Sprintf(format, args...))
	_, _ = fmt.Fprintln(s.ErrOut, line)
}

// SetColorEnabled forces the color gate on or off. Test-only
// helper; production paths read env at System() time. Re-creates
// the underlying Styles set so styling flips immediately.
func (s *IOStreams) SetColorEnabled(enabled bool) {
	s.colorEnabled = enabled
	s.styles = NewStyles(enabled)
}

// Styles returns the active styling set. Styles().Error() and
// friends are identity when color is off and lipgloss-rendered
// when color is on.
func (s *IOStreams) Styles() *Styles { return s.styles }

// NewLogger returns the slog logger channel for TWIGGIT_DEBUG:
// LevelDebug when TWIGGIT_DEBUG is non-empty at construction time,
// else LevelWarn. Output goes through the supplied writer so the
// test buffer path (iostreams.Test) captures debug output for
// assertions and the production path (iostreams.System) routes
// through os.Stderr.
//
// Caches the constructed *slog.Logger per writer so subsequent
// calls with the same writer return the same pointer — the
// singleton guarantee mandated by cli-iostreams (Factory.Logger,
// IOStreams.Logger, and slog.Default share one instance).
func NewLogger(w io.Writer) *slog.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if l, ok := loggerCache[w]; ok {
		return l
	}
	level := slog.LevelWarn
	if os.Getenv("TWIGGIT_DEBUG") != "" {
		level = slog.LevelDebug
	}
	l := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
	loggerCache[w] = l
	return l
}
