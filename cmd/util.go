package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// isQuiet checks if quiet mode is enabled
func isQuiet(cmd *cobra.Command) bool {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}

func logv(cmd *cobra.Command, level int, format string, args ...interface{}) {
	// Verbose wins over quiet (mutual exclusion)
	verbosity, _ := cmd.Flags().GetCount("verbose")

	if verbosity < level {
		return
	}

	prefix := ""
	if level > 1 {
		prefix = "  "
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s%s\n", prefix, msg)
}

// ProgressReporter provides progress feedback for bulk operations
type ProgressReporter struct {
	quiet bool // Suppress progress output in quiet mode
	out   io.Writer
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(quiet bool, out io.Writer) *ProgressReporter {
	return &ProgressReporter{
		quiet: quiet,
		out:   out,
	}
}

// Report outputs a progress message if not in quiet mode
func (p *ProgressReporter) Report(format string, args ...interface{}) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(p.out, format+"\n", args...)
}

// ReportProgress outputs progress for bulk operations
func (p *ProgressReporter) ReportProgress(current, total int, item string) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(p.out, "[%d/%d] Processing %s\n", current, total, item)
}
