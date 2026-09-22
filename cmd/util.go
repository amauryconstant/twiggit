package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"twiggit/internal/core"
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

// wrapArgsValidator wraps a cobra.PositionalArgs validator so any error it
// returns is converted to *core.UsageError. This routes args-shape failures
// (e.g., cobra.ExactArgs(1) on `create` with 0 args) through the same exit-code
// dispatch path as flag-parse errors, so they yield ExitCodeUsage (2) instead
// of the default ExitCodeError (1). Cobra v1.10 does not expose an
// args-equivalent of SetFlagErrorFunc, so we wrap at each command's Args field.
func wrapArgsValidator(v cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := v(cmd, args); err != nil {
			return core.UsageWrap(err)
		}
		return nil
	}
}

// resolveNavigationTarget resolves a navigation target with context-aware defaults
// If target is empty, uses context-aware defaults:
//   - From worktree: current branch
//   - From project: "main"
//   - From outside git: error
func resolveNavigationTarget(ctx context.Context, config *CommandConfig, target string) (*core.Context, *core.ResolutionResult, error) {
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return nil, nil, fmt.Errorf("context detection failed: %w", err)
	}

	// If no target specified, use context-aware default
	if target == "" {
		switch currentCtx.Type {
		case core.ContextWorktree:
			target = currentCtx.BranchName
		case core.ContextProject:
			target = "main"
		default:
			return nil, nil, errors.New("no target specified and no default worktree in context")
		}
	}

	// Resolve path
	req := &core.ResolvePathRequest{
		Target:  target,
		Context: currentCtx,
	}

	result, err := config.Services.NavigationService.ResolvePath(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve path for %s: %w", target, err)
	}

	return currentCtx, result, nil
}
