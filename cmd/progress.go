package cmd

import (
	"fmt"
	"io"
	"twiggit/internal/iostreams"
)

// ProgressReporter provides progress feedback for bulk operations.
// It honours the iostreams.Quiet gate rather than reading the flag
// directly so the constructor composes cleanly with the test-only
// iostreams.Test() helper.
type ProgressReporter struct {
	isQuiet bool
	out     io.Writer
}

// NewProgressReporter creates a new progress reporter; pass ios to
// inherit its Quiet flag and ErrOut writer.
func NewProgressReporter(ios *iostreams.IOStreams) *ProgressReporter {
	if ios == nil {
		return &ProgressReporter{isQuiet: false, out: io.Discard}
	}
	return &ProgressReporter{isQuiet: ios.Quiet, out: ios.ErrOut}
}

// Report outputs a progress message if not in quiet mode
func (p *ProgressReporter) Report(format string, args ...any) {
	if p.isQuiet {
		return
	}
	_, _ = fmt.Fprintf(writeOrIgnore(p.out), format+"\n", args...)
}
