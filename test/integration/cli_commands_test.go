package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/cmd"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestRootCommand_Integration(t *testing.T) {
	t.Run("all commands registered and accessible", func(t *testing.T) {
		// Create a minimal config for testing
		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService:   mocks.NewMockWorktreeService(),
				ProjectService:    mocks.NewMockProjectService(),
				NavigationService: mocks.NewMockNavigationService(),
				ContextService:    mocks.NewMockContextService(),
				ShellService:      mocks.NewMockShellService(),
			},
			Config: &domain.Config{},
		}

		// Create root command
		rootCmd := cmd.NewRootCommand(config)

		// Verify root command properties
		assert.Equal(t, "twiggit", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)

		// Verify all subcommands are registered
		expectedCommands := []string{"list", "create", "delete", "cd", "setup-shell"}
		for _, expected := range expectedCommands {
			cmd, _, err := rootCmd.Find([]string{expected})
			require.NoError(t, err, "Command '%s' should be registered", expected)
			assert.Equal(t, expected, cmd.Name(), "Command name should match '%s'", expected)
		}

		// Verify total number of commands
		assert.Len(t, rootCmd.Commands(), 5, "Should have exactly 5 subcommands registered")
	})

	t.Run("command help accessibility", func(t *testing.T) {
		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService:   mocks.NewMockWorktreeService(),
				ProjectService:    mocks.NewMockProjectService(),
				NavigationService: mocks.NewMockNavigationService(),
				ContextService:    mocks.NewMockContextService(),
				ShellService:      mocks.NewMockShellService(),
			},
			Config: &domain.Config{},
		}

		rootCmd := cmd.NewRootCommand(config)

		// Test that help works for all commands
		testCases := []struct {
			name string
			args []string
		}{
			{"root help", []string{"--help"}},
			{"list help", []string{"list", "--help"}},
			{"create help", []string{"create", "--help"}},
			{"delete help", []string{"delete", "--help"}},
			{"cd help", []string{"cd", "--help"}},
			{"setup-shell help", []string{"setup-shell", "--help"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				rootCmd.SetArgs(tc.args)
				err := rootCmd.Execute()
				assert.NoError(t, err, "Help should be accessible for %s", tc.name)
			})
		}
	})

	t.Run("invalid command handling", func(t *testing.T) {
		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService:   mocks.NewMockWorktreeService(),
				ProjectService:    mocks.NewMockProjectService(),
				NavigationService: mocks.NewMockNavigationService(),
				ContextService:    mocks.NewMockContextService(),
				ShellService:      mocks.NewMockShellService(),
			},
			Config: &domain.Config{},
		}

		rootCmd := cmd.NewRootCommand(config)
		rootCmd.SetArgs([]string{"invalid-command"})

		err := rootCmd.Execute()
		require.Error(t, err, "Should return error for invalid command")
		assert.Contains(t, err.Error(), "unknown command", "Error should mention unknown command")
	})
}
