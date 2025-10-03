package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
	"twiggit/internal/services"
)

// mockWorktreeService for testing
type mockWorktreeService struct {
	listWorktreesFunc func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error)
}

func (m *mockWorktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	return nil
}

func (m *mockWorktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	if m.listWorktreesFunc != nil {
		return m.listWorktreesFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	return nil, nil
}

func (m *mockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	return nil
}

// mockContextService for testing - we need to import the services package for the interface
type mockContextService struct {
	getCurrentContextFunc func() (*domain.Context, error)
}

func (m *mockContextService) GetCurrentContext() (*domain.Context, error) {
	if m.getCurrentContextFunc != nil {
		return m.getCurrentContextFunc()
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextService) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

func TestListCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ services.WorktreeService = (*mockWorktreeService)(nil)
	var _ services.ContextServiceInterface = (*mockContextService)(nil)
}

func TestListCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mockWorktreeService, *mockContextService)
		expectError  bool
		errorMessage string
		validateOut  func(string) bool
	}{
		{
			name: "list worktrees in project context",
			args: []string{},
			setupMocks: func(mockWS *mockWorktreeService, mockCS *mockContextService) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "test-project",
					}, nil
				}
				mockWS.listWorktreesFunc = func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
					return []*domain.WorktreeInfo{
						{Path: "/home/user/Worktrees/test-project/main", Branch: "main"},
						{Path: "/home/user/Worktrees/test-project/feature", Branch: "feature"},
					}, nil
				}
			},
			expectError: false,
			validateOut: func(output string) bool {
				return strings.Contains(output, "main") && strings.Contains(output, "feature")
			},
		},
		{
			name: "list all worktrees with --all flag",
			args: []string{"--all"},
			setupMocks: func(mockWS *mockWorktreeService, mockCS *mockContextService) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type: domain.ContextOutsideGit,
					}, nil
				}
				mockWS.listWorktreesFunc = func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
					return []*domain.WorktreeInfo{}, nil
				}
			},
			expectError: false,
		},
		{
			name: "context detection failure",
			args: []string{},
			setupMocks: func(mockWS *mockWorktreeService, mockCS *mockContextService) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return nil, errors.New("detection failed")
				}
			},
			expectError:  true,
			errorMessage: "context detection failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockWS := &mockWorktreeService{}
			mockCS := &mockContextService{}

			tc.setupMocks(mockWS, mockCS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					WorktreeService: mockWS,
					ContextService:  mockCS,
				},
			}

			cmd := NewListCommand(config)
			cmd.SetArgs(tc.args)

			// Set flags
			for flag, value := range tc.flags {
				cmd.Flags().Set(flag, value)
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			err := cmd.Execute()

			// Validate results
			if tc.expectError {
				require.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				if tc.validateOut != nil {
					output := buf.String()
					result := tc.validateOut(output)
					if !result {
						t.Logf("Validation failed. Output: %q", output)
					}
					assert.True(t, result)
				}
			}
		})
	}
}
