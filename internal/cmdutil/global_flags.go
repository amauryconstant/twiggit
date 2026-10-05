package cmdutil

import (
	"twiggit/internal/iostreams"

	"github.com/spf13/cobra"
)

// GlobalOptions holds the values parsed from the persistent --output /
// --quiet / --verbose flags. AddPersistentFlags binds the cobra flag
// pointers to the fields of the supplied struct so commands can read
// the bound values without re-parsing flag.Value themselves.
//
// Verbose is a count (-v → 1, -vv → 2, ...) so callers can branch on
// the level. iostreams.Verbosef only honours level > 0; level 2 is
// reserved for callers that want to add deeper diagnostics in a
// future slice. For now run bodies emit the same Verbosef line on
// any non-zero count so -v and -vv produce the visible user output.
//
// GlobalOptions is shared by reference across every subcommand so the
// bound flag pointers mutate a single backing struct. Subcommands copy
// the pointer into their <Cmd>Options struct so test code can swap
// values after flag parse.
type GlobalOptions struct {
	Output  string
	IsQuiet bool
	Verbose int
}

// IsVerbose reports whether the user supplied at least one -v.
func (g *GlobalOptions) IsVerbose() bool { return g.Verbose > 0 }

// AddPersistentFlags registers --output, --quiet, and --verbose as
// persistent flags on cmd and binds them to opts. Persistent means
// every subcommand inherits them, matching the cli-verbose-output spec
// ("Verbose flag is global") and the cli-iostreams spec ("Quiet is
// global"). Defaults are the zero values (empty Output, false Quiet,
// 0 Verbose) so a fresh invocation is indistinguishable from no-flags.
//
// Callers wire AddPersistentFlags on the root cobra.Command exactly
// once; subcommands reuse the parent flag definitions. Subcommand
// flags (force, yes, dry-run, ...) are command-local and stay in the
// cmd/ files.
func AddPersistentFlags(cmd *cobra.Command, opts *GlobalOptions) {
	cmd.PersistentFlags().StringVar(&opts.Output, "output", "", "Output format: json, table, plain. Empty falls through to the per-command default.")
	cmd.PersistentFlags().BoolVar(&opts.IsQuiet, "quiet", false, "Suppress non-essential output for scripting scenarios.")
	cmd.PersistentFlags().CountVarP(&opts.Verbose, "verbose", "v", "Increase verbosity (can be used multiple times: -v, -vv)")
}

// ApplyToIOS mirrors the GlobalOptions onto the iostreams.IOStreams
// gates so Verbosef / quiet-aware formatters respect the user-supplied
// flags. The root RunE closure calls this after flag parse; tests
// that construct *GlobalOptions directly can call ApplyToIOS without
// flag parse or invoke *iostreams.IOStreams setters.
//
// Idempotent: re-applying the same values is a no-op so the helper is
// safe to call from PersistentPreRunE and individual subcommand RunE.
func (g *GlobalOptions) ApplyToIOS(ios *iostreams.IOStreams) {
	if ios == nil {
		return
	}
	ios.IsQuiet = g.IsQuiet
	ios.IsVerbose = g.IsVerbose()
}
