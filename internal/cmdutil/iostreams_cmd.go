package cmdutil

import (
	"context"
	"twiggit/internal/iostreams"

	"github.com/spf13/cobra"
)

// iosContextKey is the unexported context.Context key under which a
// cobra command tree stashes its *iostreams.IOStreams. The
// IOStreamsFromCmd / SetIOStreams helpers read and write through this
// key; key type is a private struct so no external package can
// collide.
//
// Stashing on the cobra command's context lets helper functions that
// only see *cobra.Command recover the IOS without needing the Factory
// reference. Subcommand RunE closures usually hold a Factory-derived
// pointer and prefer Factory.IOStreams directly; the context lookup
// exists for the error-handler path where the cmd is the only handle.
type iosContextKey struct{}

// SetIOStreams attaches ios to cmd's context. main.go calls this
// once when building the root command; subcommands inherit the same
// context through cobra's PersistentPreRunE / RunE plumbing.
//
// Returns the receiver for fluent composition with NewRootCommand.
//
// Exported so cmd/root.go's PersistentPreRunE can stash the
// Factory-supplied IOStreams on cmd before the first subcommand runs.
func SetIOStreams(cmd *cobra.Command, ios *iostreams.IOStreams) *cobra.Command {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, iosContextKey{}, ios))
	return cmd
}

// IOStreamsFromCmd returns the *iostreams.IOStreams previously stashed
// on cmd via SetIOStreams. Falls back to iostreams.System() when no
// value is present so callers in the error-handler path always get a
// writable IOS even before the cobra tree is fully wired.
func IOStreamsFromCmd(cmd *cobra.Command) *iostreams.IOStreams {
	if cmd == nil {
		return iostreams.System()
	}
	if v := cmd.Context(); v != nil {
		if ios, ok := v.Value(iosContextKey{}).(*iostreams.IOStreams); ok && ios != nil {
			return ios
		}
	}
	return iostreams.System()
}
