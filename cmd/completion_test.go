package cmd

import (
	"testing"
	"twiggit/internal/core"

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

func TestNewCompletionCommand_AcceptsZeroPositionals(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	completion := newCompletionCommand(root)

	require.NotNil(t, completion.Args, "completion parent must declare Args validator")

	err := completion.Args(completion, []string{})
	assert.NoError(t, err, "twiggit completion must accept zero positionals to show help")
}

func TestNewCompletionCommand_AcceptsEachSupportedShell(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	completion := newCompletionCommand(root)

	shells := []string{
		"bash", "zsh", "fish", "powershell",
		"elvish", "nushell", "oil", "tcsh", "xonsh", "cmd-clink",
	}
	for _, s := range shells {
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			err := completion.Args(completion, []string{s})
			assert.NoError(t, err, "shell %q must be accepted by completion parent", s)
		})
	}
}

func TestNewCompletionCommand_RejectsUnsupportedShell(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "twiggit"}
	completion := newCompletionCommand(root)

	err := completion.Args(completion, []string{"ksh"})
	require.Error(t, err, "twiggit completion ksh must fail validation")

	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve, "rejection must surface as *core.ValidationError")

	assert.Equal(t, "shell", ve.Field)
	assert.Equal(t, "ksh", ve.Value)
	assert.Contains(t, ve.Message, "unsupported")
	assert.Contains(t, ve.Message, "ksh")
}
