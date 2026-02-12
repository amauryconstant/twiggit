package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		setupMocks   func(*mocks.MockWorktreeService, *mocks.MockContextService, *mocks.MockGitService, *mocks.MockNavigationService)
		expectError  bool
		errorMessage string
		validateOut  func(string) bool
	}{
		{
			name: "delete worktree with safety checks",
			args: []string{"test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
					ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
				}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"}}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{
					IsClean:               true,
					HasUncommittedChanges: false,
				}, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
			},
			expectError: false,
		},

		{
			name: "force delete dirty worktree",
			args: []string{"--force", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
			},
			expectError: false,
		},

		{
			name: "delete with -C flag from worktree context outputs project path",
			args: []string{"-C", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{
					Type:       domain.ContextWorktree,
					BranchName: "feature-branch",
					Path:       "/home/user/Worktrees/test-project/feature-branch",
				}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
					ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
				}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"}}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{
					IsClean: true,
				}, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
				mockNS.On("ResolvePath", mock.Anything, mock.AnythingOfType("*domain.ResolvePathRequest")).Return(&domain.ResolutionResult{
					ResolvedPath: "/home/user/Projects/test-project",
				}, nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return strings.TrimSpace(output) == "/home/user/Projects/test-project"
			},
		},

		{
			name: "delete with -C flag from project context outputs nothing",
			args: []string{"-C", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{
					Type: domain.ContextProject,
					Path: "/home/user/Projects/test-project",
				}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
					ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
				}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"}}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{
					IsClean: true,
				}, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return output == ""
			},
		},

		{
			name: "delete with -C flag from outside git context outputs nothing",
			args: []string{"-C", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{
					Type: domain.ContextOutsideGit,
					Path: "/home/user",
				}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{
					ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
				}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{{Path: "/home/user/Worktrees/test-project/feature-branch", Branch: "feature-branch"}}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("GetWorktreeStatus", mock.Anything, mock.AnythingOfType("string")).Return(&domain.WorktreeStatus{
					IsClean: true,
				}, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return output == ""
			},
		},

		{
			name: "delete with -f short form flag works correctly",
			args: []string{"-f", "test-project/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockGS *mocks.MockGitService, mockNS *mocks.MockNavigationService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{}, nil)
				mockCS.On("ResolveIdentifier", mock.AnythingOfType("string")).Return(&domain.ResolutionResult{}, nil)
				mockGS.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{}, nil)
				mockGS.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
				mockWS.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("*domain.DeleteWorktreeRequest")).Return(nil)
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
			mockNS := mocks.NewMockNavigationService()

			tc.setupMocks(mockWS, mockCS, mockGS, mockNS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					WorktreeService:   mockWS,
					ContextService:    mockCS,
					GitClient:         mockGS,
					NavigationService: mockNS,
				},
			}

			cmd := NewDeleteCommand(config)
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
