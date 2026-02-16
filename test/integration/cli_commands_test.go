package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		expectedCommands := []string{"list", "create", "delete", "prune", "cd", "init", "version", "completion"}
		for _, expected := range expectedCommands {
			cmd, _, err := rootCmd.Find([]string{expected})
			require.NoError(t, err, "Command '%s' should be registered", expected)
			assert.Equal(t, expected, cmd.Name(), "Command name should match '%s'", expected)
		}

		// Verify total number of commands
		assert.Len(t, rootCmd.Commands(), 9, "Should have exactly 9 subcommands registered")
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
			{"prune help", []string{"prune", "--help"}},
			{"cd help", []string{"cd", "--help"}},
			{"init help", []string{"init", "--help"}},
			{"version help", []string{"version", "--help"}},
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

func TestCreateCommand_WithCdFlag(t *testing.T) {
	t.Run("create with -C flag outputs path only", func(t *testing.T) {
		mockWS := mocks.NewMockWorktreeService()
		mockCS := mocks.NewMockContextService()
		mockPS := mocks.NewMockProjectService()

		mockCS.On("GetCurrentContext").Return(&domain.Context{Type: domain.ContextOutsideGit}, nil)
		mockPS.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
			Name:        "test-project",
			GitRepoPath: "/tmp/test-project",
		}, nil)
		mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.WorktreeInfo{
			Path:   "/tmp/test-project/feature-branch",
			Branch: "feature-branch",
		}, nil)

		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService: mockWS,
				ContextService:  mockCS,
				ProjectService:  mockPS,
			},
		}

		createCmd := cmd.NewCreateCommand(config)
		createCmd.SetArgs([]string{"-C", "test-project/feature-branch"})

		var buf bytes.Buffer
		createCmd.SetOut(&buf)

		err := createCmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Equal(t, "/tmp/test-project/feature-branch\n", output, "Should output path only")
		assert.NotContains(t, output, "Created worktree", "Should not include success message")
	})

	t.Run("create without -C flag outputs success message", func(t *testing.T) {
		mockWS := mocks.NewMockWorktreeService()
		mockCS := mocks.NewMockContextService()
		mockPS := mocks.NewMockProjectService()

		mockCS.On("GetCurrentContext").Return(&domain.Context{Type: domain.ContextOutsideGit}, nil)
		mockPS.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
			Name:        "test-project",
			GitRepoPath: "/tmp/test-project",
		}, nil)
		mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.WorktreeInfo{
			Path:   "/tmp/test-project/feature-branch",
			Branch: "feature-branch",
		}, nil)

		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService: mockWS,
				ContextService:  mockCS,
				ProjectService:  mockPS,
			},
		}

		createCmd := cmd.NewCreateCommand(config)
		createCmd.SetArgs([]string{"test-project/feature-branch"})

		var buf bytes.Buffer
		createCmd.SetOut(&buf)

		err := createCmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Created worktree", "Should include success message")
		assert.NotEqual(t, "/tmp/test-project/feature-branch\n", output, "Should not output path only")
	})
}

func TestDeleteCommand_WithCdFlag(t *testing.T) {
	t.Run("delete with -C from worktree context outputs project path", func(t *testing.T) {
		mockWS := mocks.NewMockWorktreeService()
		mockCS := mocks.NewMockContextService()
		mockNS := mocks.NewMockNavigationService()

		mockCS.On("GetCurrentContext").Return(&domain.Context{
			Type:       domain.ContextWorktree,
			BranchName: "feature-branch",
			Path:       "/tmp/test-project/feature-branch",
		}, nil)
		mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
			ResolvedPath: "/tmp/test-project/feature-branch",
		}, nil)
		mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{IsClean: true}, nil)
		mockWS.On("GetWorktreeByPath", mock.Anything, mock.Anything, mock.Anything).Return(&domain.WorktreeInfo{
			Branch: "feature-branch",
		}, nil)
		mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
		mockNS.On("ResolvePath", mock.Anything, mock.AnythingOfType("*domain.ResolvePathRequest")).Return(&domain.ResolutionResult{
			ResolvedPath: "/tmp/test-project",
		}, nil)

		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService:   mockWS,
				ContextService:    mockCS,
				NavigationService: mockNS,
			},
		}

		deleteCmd := cmd.NewDeleteCommand(config)
		deleteCmd.SetArgs([]string{"-C", "test-project/feature-branch"})

		var buf bytes.Buffer
		deleteCmd.SetOut(&buf)

		err := deleteCmd.Execute()
		require.NoError(t, err)

		output := strings.TrimSpace(buf.String())
		assert.Equal(t, "/tmp/test-project", output, "Should output only project path, not 'Deleted worktree' message")
	})

	t.Run("delete with -C from project context outputs nothing", func(t *testing.T) {
		mockWS := mocks.NewMockWorktreeService()
		mockCS := mocks.NewMockContextService()
		mockNS := mocks.NewMockNavigationService()

		mockCS.On("GetCurrentContext").Return(&domain.Context{
			Type: domain.ContextProject,
			Path: "/tmp/test-project",
		}, nil)
		mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
			ResolvedPath: "/tmp/test-project/feature-branch",
		}, nil)
		mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{IsClean: true}, nil)
		mockWS.On("GetWorktreeByPath", mock.Anything, mock.Anything, mock.Anything).Return(&domain.WorktreeInfo{
			Branch: "feature-branch",
		}, nil)
		mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)

		config := &cmd.CommandConfig{
			Services: &cmd.ServiceContainer{
				WorktreeService:   mockWS,
				ContextService:    mockCS,
				NavigationService: mockNS,
			},
		}

		deleteCmd := cmd.NewDeleteCommand(config)
		deleteCmd.SetArgs([]string{"-C", "test-project/feature-branch"})

		var buf bytes.Buffer
		deleteCmd.SetOut(&buf)

		err := deleteCmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Empty(t, output, "Should output nothing when deleting from project context with -C")
	})
}
