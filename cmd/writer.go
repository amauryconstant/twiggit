package cmd

import "io"

// ignoreWriter wraps an io.Writer so its Write always reports success.
//
// Per the swallowed-error policy (modernization sweep, task 16.5),
// terminal writes to user-visible streams are intentionally best-effort:
// a broken pipe or closed TTY cannot be meaningfully recovered from
// inside a CLI, and logging the error via slog.Error would re-target
// the same broken stream. Use writeOrIgnore to wrap the destination so
// the call site does not need a `_, _ =` lint pattern.
type ignoreWriter struct{ io.Writer }

// Write delegates to the wrapped writer and discards the error.
func (w ignoreWriter) Write(p []byte) (int, error) {
	n, _ := w.Writer.Write(p)
	return n, nil
}

// writeOrIgnore returns an io.Writer whose Write never errors.
// See ignoreWriter docs.
func writeOrIgnore(w io.Writer) io.Writer { return ignoreWriter{w} }
