package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestDeleteCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ interface{} = mocks.NewMockWorktreeService()
	var _ interface{} = mocks.NewMockContextService()
}

func TestDeleteCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mocks.MockWorktreeService, *mocks.MockContextService, *mocks.MockGitService)
		expectError  bool
		errorMessage string
	}{
		{
			name: "delete worktree with safety checks",
			args: []string{"test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.ResolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{
						ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
					}, nil
				}
				mockWS.GetWorktreeStatusFunc = func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
					return &domain.WorktreeStatus{
						IsClean:               true,
						HasUncommittedChanges: false,
					}, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},

		{
			name: "force delete dirty worktree",
			args: []string{"--force", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.ResolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{}, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},

		{
			name: "delete worktree with change-dir flag",
			args: []string{"-C", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Path: "/home/user/Workspaces/test-project"}, nil
				}
				mockCS.ResolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{
						ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
					}, nil
				}
				mockWS.GetWorktreeStatusFunc = func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
					return &domain.WorktreeStatus{
						IsClean: true,
					}, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockWS := mocks.NewMockWorktreeService()
			mockCS := mocks.NewMockContextService()
			mockGS := mocks.NewMockGitService()

			tc.setupMocks(mockWS, mockCS, mockGS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					WorktreeService: mockWS,
					ContextService:  mockCS,
					GitClient:       mockGS,
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
