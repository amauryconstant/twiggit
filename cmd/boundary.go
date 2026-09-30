package cmd

import (
	"log/slog"
	"twiggit/internal/cmdutil"

	"github.com/spf13/cobra"
)

// BoundaryLogger returns the cmd-scoped debug logger: f.Logger()
// with "command" set to cmd.Name(). All boundary debug logging in
// cmd/ should use the result so the singleton logger carries trace
// correlation per the cli-iostreams spec.
//
// runX receives the logger through opts.Logger populated in RunE;
// the boundary logger never escapes the cmd it was created for.
func BoundaryLogger(f *cmdutil.Factory, cmd *cobra.Command) *slog.Logger {
	return f.Logger().With("command", cmd.Name())
}
