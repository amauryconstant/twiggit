package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_HelpCommand(t *testing.T) {
	// Test that main command can be executed with --help
	os.Args = []string{"twiggit", "--help"}

	// This should not panic and should work
	// We can't easily test the actual output without more complex setup
	// But we can ensure the command structure is valid
	assert.NotPanics(t, func() {
		// Just ensure we can import and the root command exists
		rootCmd := getRootCommand()
		require.NotNil(t, rootCmd)
		assert.Equal(t, "twiggit", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "Git worktree management")
	})
}

func TestMain_SubcommandStructure(t *testing.T) {
	rootCmd := getRootCommand()
	require.NotNil(t, rootCmd)

	// Check that all expected subcommands exist
	expectedCommands := []string{"create", "switch", "status", "cleanup", "list"}

	commandNames := make([]string, len(rootCmd.Commands()))
	for i, cmd := range rootCmd.Commands() {
		commandNames[i] = cmd.Name()
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected, "Missing subcommand: %s", expected)
	}
}

func TestMain_GlobalFlags(t *testing.T) {
	rootCmd := getRootCommand()
	require.NotNil(t, rootCmd)

	// Check that global flags exist
	workspaceFlag := rootCmd.PersistentFlags().Lookup("workspace")
	assert.NotNil(t, workspaceFlag, "Missing --workspace flag")

	projectFlag := rootCmd.PersistentFlags().Lookup("project")
	assert.NotNil(t, projectFlag, "Missing --project flag")

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag, "Missing --verbose flag")

	quietFlag := rootCmd.PersistentFlags().Lookup("quiet")
	assert.NotNil(t, quietFlag, "Missing --quiet flag")
}
