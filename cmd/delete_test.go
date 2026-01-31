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
			name: "abort on dirty worktree",
			args: []string{"test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{}, nil
				}
				mockCS.ResolveIdentifierFunc = func(identifier string) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{}, nil
				}
				mockWS.GetWorktreeStatusFunc = func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
					return &domain.WorktreeStatus{
						IsClean:               false,
						HasUncommittedChanges: true,
					}, nil
				}
			},
			expectError:  true,
			errorMessage: "worktree has uncommitted changes",
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
			name: "delete worktree with keep-branch flag",
			args: []string{"--keep-branch", "test-project/feature-branch"},
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
						IsClean: true,
					}, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					assert.True(t, req.KeepBranch, "KeepBranch should be true")
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "delete worktree without keep-branch flag",
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
						IsClean: true,
					}, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					assert.False(t, req.KeepBranch, "KeepBranch should be false")
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "delete worktree with merged-only flag and merged branch",
			args: []string{"--merged-only", "test-project/feature-branch"},
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
						IsClean: true,
					}, nil
				}
				mockGS.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
					return []domain.WorktreeInfo{
						{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"},
					}, nil
				}
				mockGS.IsBranchMergedFunc = func(ctx context.Context, repoPath, branchName string) (bool, error) {
					return true, nil
				}
				mockWS.DeleteWorktreeFunc = func(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "delete worktree with merged-only flag and unmerged branch",
			args: []string{"--merged-only", "test-project/feature-branch"},
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
						IsClean: true,
					}, nil
				}
				mockGS.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
					return []domain.WorktreeInfo{
						{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"},
					}, nil
				}
				mockGS.IsBranchMergedFunc = func(ctx context.Context, repoPath, branchName string) (bool, error) {
					return false, nil
				}
			},
			expectError:  true,
			errorMessage: "is not merged",
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
