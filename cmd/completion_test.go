package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompletionShellCommand_AcceptsNoPositional(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	shellCmd := newCompletionShellCommand(root, "zsh")

	require.NotNil(t, shellCmd.Args, "completion subcommand must declare Args validator")

	err := shellCmd.Args(shellCmd, []string{})
	assert.NoError(t, err, "twiggit completion zsh must accept zero positionals")
}

func TestNewCompletionShellCommand_RejectsStrayPositional(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	shellCmd := newCompletionShellCommand(root, "zsh")

	require.NotNil(t, shellCmd.Args, "completion subcommand must declare Args validator")

	err := shellCmd.Args(shellCmd, []string{"extra"})
	assert.Error(t, err, "twiggit completion zsh must reject stray positionals")
}

func TestNewCompletionCommand_HasAllShellsRegistered(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	completion := newCompletionCommand(root)

	expected := []string{
		"bash", "zsh", "fish", "powershell",
		"elvish", "nushell", "oil", "tcsh", "xonsh", "cmd-clink",
	}
	for _, name := range expected {
		matched, _, err := completion.Find([]string{name})
		require.NoError(t, err, "completion subcommand %q lookup must not error", name)
		require.NotNil(t, matched, "completion subcommand %q must be registered", name)
		assert.Equal(t, name, matched.Name(), "completion subcommand %q must resolve exactly", name)
	}
}
