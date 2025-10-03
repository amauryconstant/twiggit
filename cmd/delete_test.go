package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
	"twiggit/internal/services"
)

// mockWorktreeService for testing
type mockWorktreeServiceForDelete struct {
	deleteWorktreeFunc    func(ctx context.Context, req *domain.DeleteWorktreeRequest) error
	getWorktreeStatusFunc func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error)
}

func (m *mockWorktreeServiceForDelete) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeServiceForDelete) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	if m.deleteWorktreeFunc != nil {
		return m.deleteWorktreeFunc(ctx, req)
	}
	return nil
}

func (m *mockWorktreeServiceForDelete) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeServiceForDelete) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	if m.getWorktreeStatusFunc != nil {
		return m.getWorktreeStatusFunc(ctx, worktreePath)
	}
	return nil, nil
}

func (m *mockWorktreeServiceForDelete) ValidateWorktree(ctx context.Context, worktreePath string) error {
	return nil
}

// mockContextService for testing
type mockContextServiceForDelete struct {
	getCurrentContextFunc func() (*domain.Context, error)
	resolveIdentifierFunc func(identifier string) (*domain.ResolutionResult, error)
}

func (m *mockContextServiceForDelete) GetCurrentContext() (*domain.Context, error) {
	if m.getCurrentContextFunc != nil {
		return m.getCurrentContextFunc()
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForDelete) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForDelete) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	if m.resolveIdentifierFunc != nil {
		return m.resolveIdentifierFunc(identifier)
	}
	return nil, nil
}

func (m *mockContextServiceForDelete) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextServiceForDelete) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

func TestDeleteCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ services.WorktreeService = (*mockWorktreeServiceForDelete)(nil)
	var _ services.ContextServiceInterface = (*mockContextServiceForDelete)(nil)
}

func TestDeleteCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mockWorktreeServiceForDelete, *mockContextServiceForDelete)
		expectError  bool
		errorMessage string
	}{
		{
			name: "delete worktree with safety checks",
			args: []string{"test-project/feature-branch"},
			setupMocks: func(mockWS *mockWorktreeServiceForDelete, mockCS *mockContextServiceForDelete) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.resolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{
						ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
					}, nil
				}
				mockWS.getWorktreeStatusFunc = func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
					return &domain.WorktreeStatus{
						IsClean: true,
					}, nil
				}
				mockWS.deleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "abort on dirty worktree",
			args: []string{"test-project/feature-branch"},
			setupMocks: func(mockWS *mockWorktreeServiceForDelete, mockCS *mockContextServiceForDelete) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.resolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{}, nil
				}
				mockWS.getWorktreeStatusFunc = func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
					return &domain.WorktreeStatus{
						IsClean: false,
					}, nil
				}
			},
			expectError:  true,
			errorMessage: "worktree has uncommitted changes",
		},
		{
			name: "force delete dirty worktree",
			args: []string{"--force", "test-project/feature-branch"},
			setupMocks: func(mockWS *mockWorktreeServiceForDelete, mockCS *mockContextServiceForDelete) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.resolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{}, nil
				}
				mockWS.deleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockWS := &mockWorktreeServiceForDelete{}
			mockCS := &mockContextServiceForDelete{}

			tc.setupMocks(mockWS, mockCS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					WorktreeService: mockWS,
					ContextService:  mockCS,
				},
			}

			cmd := NewDeleteCommand(config)
			cmd.SetArgs(tc.args)

			// Set flags
			for flag, value := range tc.flags {
				cmd.Flags().Set(flag, value)
			}

			err := cmd.Execute()

			// Validate results
			if tc.expectError {
				require.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
