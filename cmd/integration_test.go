package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

// Simple stub implementations for integration testing
type stubWorktreeService struct{}

func (s *stubWorktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	return nil, nil
}
func (s *stubWorktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	return nil
}
func (s *stubWorktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	return nil, nil
}
func (s *stubWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	return nil, nil
}
func (s *stubWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	return nil
}

type stubProjectService struct{}

func (s *stubProjectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	return nil, nil
}
func (s *stubProjectService) ValidateProject(ctx context.Context, projectPath string) error {
	return nil
}
func (s *stubProjectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	return nil, nil
}
func (s *stubProjectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	return nil, nil
}

type stubNavigationService struct{}

func (s *stubNavigationService) ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	return nil, nil
}
func (s *stubNavigationService) ValidatePath(ctx context.Context, path string) error { return nil }
func (s *stubNavigationService) GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

type stubContextService struct{}

func (s *stubContextService) GetCurrentContext() (*domain.Context, error) {
	return &domain.Context{}, nil
}

func (s *stubContextService) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{}, nil
}

func (s *stubContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	return &domain.ResolutionResult{}, nil
}

func (s *stubContextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return &domain.ResolutionResult{}, nil
}

func (s *stubContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

func TestRootCommand_Integration(t *testing.T) {
	t.Run("all commands registered and accessible", func(t *testing.T) {
		// Create a minimal config for testing
		config := &CommandConfig{
			Services: &ServiceContainer{
				WorktreeService:   &stubWorktreeService{},
				ProjectService:    &stubProjectService{},
				NavigationService: &stubNavigationService{},
				ContextService:    &stubContextService{},
			},
			Config: &domain.Config{},
		}

		// Create root command
		rootCmd := NewRootCommand(config)

		// Verify root command properties
		assert.Equal(t, "twiggit", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)

		// Verify all subcommands are registered
		expectedCommands := []string{"list", "create", "delete", "cd"}
		for _, expected := range expectedCommands {
			cmd, _, err := rootCmd.Find([]string{expected})
			require.NoError(t, err, "Command '%s' should be registered", expected)
			assert.Equal(t, expected, cmd.Name(), "Command name should match '%s'", expected)
		}

		// Verify total number of commands
		assert.Len(t, rootCmd.Commands(), 4, "Should have exactly 4 subcommands registered")
	})

	t.Run("command help accessibility", func(t *testing.T) {
		config := &CommandConfig{
			Services: &ServiceContainer{
				WorktreeService:   &stubWorktreeService{},
				ProjectService:    &stubProjectService{},
				NavigationService: &stubNavigationService{},
				ContextService:    &stubContextService{},
			},
			Config: &domain.Config{},
		}

		rootCmd := NewRootCommand(config)

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
		config := &CommandConfig{
			Services: &ServiceContainer{
				WorktreeService:   &stubWorktreeService{},
				ProjectService:    &stubProjectService{},
				NavigationService: &stubNavigationService{},
				ContextService:    &stubContextService{},
			},
			Config: &domain.Config{},
		}

		rootCmd := NewRootCommand(config)
		rootCmd.SetArgs([]string{"invalid-command"})

		err := rootCmd.Execute()
		require.Error(t, err, "Should return error for invalid command")
		assert.Contains(t, err.Error(), "unknown command", "Error should mention unknown command")
	})
}
