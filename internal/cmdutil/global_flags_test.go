package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
)

// TestAddPersistentFlags_AttachesAllThreeFlags confirms each global
// flag is registered on the cobra.Command after the helper runs.
func TestAddPersistentFlags_AttachesAllThreeFlags(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	opts := &cmdutil.GlobalOptions{}

	cmdutil.AddPersistentFlags(cmd, opts)

	for _, name := range []string{"output", "quiet", "verbose"} {
		flag := cmd.PersistentFlags().Lookup(name)
		require.NotNil(t, flag, "missing persistent flag %q", name)
	}
}

// TestAddPersistentFlags_DefaultsAreZeroValues confirms the defaults
// match the GlobalOptions zero values, so the no-flags invocation is
// indistinguishable from a fresh struct.
func TestAddPersistentFlags_DefaultsAreZeroValues(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	opts := &cmdutil.GlobalOptions{}
	cmdutil.AddPersistentFlags(cmd, opts)

	assert.Empty(t, opts.Output)
	assert.False(t, opts.Quiet)
	assert.Equal(t, 0, opts.Verbose)
	assert.False(t, opts.IsVerbose())
}

// TestAddPersistentFlags_BindsValuesToOpts confirms the cobra binding
// writes parsed flag values back into the GlobalOptions struct so
// cmd/ can read opts.Quiet instead of re-parsing Flag.Lookup.
func TestAddPersistentFlags_BindsValuesToOpts(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	opts := &cmdutil.GlobalOptions{}
	cmdutil.AddPersistentFlags(cmd, opts)

	require.NoError(t, cmd.PersistentFlags().Set("output", "json"))
	require.NoError(t, cmd.PersistentFlags().Set("quiet", "true"))
	require.NoError(t, cmd.PersistentFlags().Set("verbose", "1"))

	assert.Equal(t, "json", opts.Output)
	assert.True(t, opts.Quiet)
	assert.Equal(t, 1, opts.Verbose)
	assert.True(t, opts.IsVerbose())
}

// TestAddPersistentFlags_QuietAndVerboseArePersistent confirms the
// flags persist to subcommands, matching the cli-verbose-output and
// cli-iostreams specs ("Verbose is global", "Quiet is global").
func TestAddPersistentFlags_QuietAndVerboseArePersistent(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "root"}
	opts := &cmdutil.GlobalOptions{}
	cmdutil.AddPersistentFlags(root, opts)

	child := &cobra.Command{Use: "child"}
	root.AddCommand(child)

	// cobra exposes inherited persistent flags via InheritedFlags();
	// PersistentFlags() returns only the cmd's own set.
	assert.NotNil(t, child.InheritedFlags().Lookup("quiet"))
	assert.NotNil(t, child.InheritedFlags().Lookup("verbose"))
}

// TestAddPersistentFlags_VerboseCountCollapse confirms the previous
// two-level (-v / -vv) scheme is collapsed: any non-zero count
// reports IsVerbose() == true, per cli-verbose-output "Level 2
// output" and "the previous two-level scheme is removed".
func TestAddPersistentFlags_VerboseCountCollapse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  bool
	}{
		{"zero is silent", 0, false},
		{"one -v is verbose", 1, true},
		{"two -vv is verbose", 2, true},
		{"three is verbose", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)

			opts := &cmdutil.GlobalOptions{Verbose: tt.count}
			is.Equal(tt.want, opts.IsVerbose())
		})
	}
}
