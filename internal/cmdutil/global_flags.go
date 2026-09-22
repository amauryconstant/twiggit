package cmdutil

import (
	"github.com/spf13/cobra"
)

// GlobalOptions holds the values parsed from the persistent --output /
// --quiet / --verbose flags. AddPersistentFlags binds the cobra flag
// pointers to the fields of the supplied struct so commands can read
// the bound values without re-parsing flag.Value themselves.
//
// GlobalOptions is value-type, not pointer, because cmd/ never mutates
// the struct after construction; it only reads it inside RunE.
type GlobalOptions struct {
	Output  string
	Quiet   bool
	Verbose bool
}

// AddPersistentFlags registers --output, --quiet, and --verbose as
// persistent flags on cmd and binds them to opts. Persistent means
// every subcommand inherits them, matching the cli-verbose-output spec
// ("Verbose flag is global") and the cli-iostreams spec ("Quiet is
// global"). Defaults are the zero values (empty Output, false Quiet /
// Verbose) so a fresh invocation is indistinguishable from no-flags.
//
// Callers wire AddPersistentFlags on the root cobra.Command; subcommand
// flags (force, yes, dry-run, ...) are command-local and stay in the
// cmd/ files.
func AddPersistentFlags(cmd *cobra.Command, opts *GlobalOptions) {
	cmd.PersistentFlags().StringVar(&opts.Output, "output", "", "Output format: table (default), json, jsonl, plain. Empty defaults to table.")
	cmd.PersistentFlags().BoolVar(&opts.Quiet, "quiet", false, "Suppress non-essential output for scripting scenarios.")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Emit verbose diagnostics to stderr.")
}
