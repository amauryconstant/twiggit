// Path: internal/iostreams/iostreams.go
package iostreams

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/term"
)

// IOStreams is the only door to the terminal. Commands write data to Out and
// everything else to ErrOut, and ask it about TTY state, so tests swap in
// buffers and exercise the non-interactive path by default.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	Verbose bool         // set by root PersistentPreRunE from -v/--verbose
	Logger  *slog.Logger // set by root PersistentPreRunE from --debug/MYAPP_DEBUG

	stdoutTTY bool
	stdinTTY  bool
	stderrTTY bool
	styles    *Styles // for Out
	errStyles *Styles // for ErrOut: stderr can be a terminal while stdout is piped, or the reverse
}

func System() *IOStreams {
	s := &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		Logger: slog.New(slog.DiscardHandler),
		// Detected once here; SetTTY leaves it alone, so tests keep stderr colorless.
		stderrTTY: term.IsTerminal(int(os.Stderr.Fd())),
	}
	s.SetTTY(term.IsTerminal(int(os.Stdout.Fd())), term.IsTerminal(int(os.Stdin.Fd())))
	return s
}

// Test returns non-TTY, colorless streams backed by buffers (in, out, errOut).
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	s := &IOStreams{
		In:        io.NopCloser(in),
		Out:       out,
		ErrOut:    errOut,
		Logger:    slog.New(slog.DiscardHandler),
		styles:    NewStyles(false),
		errStyles: NewStyles(false),
	}
	return s, in, out, errOut
}

// SetTTY records terminal state and rebuilds styles. System calls it with the
// real state; tests call it to exercise the interactive path.
func (s *IOStreams) SetTTY(stdout, stdin bool) {
	s.stdoutTTY, s.stdinTTY = stdout, stdin
	color := os.Getenv("NO_COLOR") == ""
	s.styles = NewStyles(stdout && color)
	s.errStyles = NewStyles(s.stderrTTY && color)
}

func (s *IOStreams) IsStdoutTTY() bool   { return s.stdoutTTY }
func (s *IOStreams) IsInteractive() bool { return s.stdoutTTY && s.stdinTTY }
func (s *IOStreams) Styles() *Styles     { return s.styles }    // for writes to Out
func (s *IOStreams) ErrStyles() *Styles  { return s.errStyles } // for writes to ErrOut

// Verbosef writes a dim progress line to stderr when -v is set.
func (s *IOStreams) Verbosef(format string, args ...any) {
	if s.Verbose {
		fmt.Fprintln(s.ErrOut, s.errStyles.Dim(fmt.Sprintf(format, args...)))
	}
}
