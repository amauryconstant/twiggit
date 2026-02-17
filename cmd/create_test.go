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

func TestCreateCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ interface{} = mocks.NewMockWorktreeService()
	var _ interface{} = mocks.NewMockProjectService()
	var _ interface{} = mocks.NewMockContextService()
}

func TestCreateCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mocks.MockWorktreeService, *mocks.MockContextService, *mocks.MockProjectService)
		expectError  bool
		errorMessage string
		validateOut  func(string) bool
	}{
		{
			name:  "create worktree with project/branch",
			args:  []string{"test-project/feature-branch"},
			flags: map[string]string{"source": "main"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{Type: domain.ContextOutsideGit}, nil)
				mockPS.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
					Name:        "test-project",
					GitRepoPath: "/home/user/Projects/test-project",
				}, nil)
				mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.CreateWorktreeResult{
					Worktree: &domain.WorktreeInfo{
						Path:   "/home/user/Worktrees/test-project/feature-branch",
						Branch: "feature-branch",
					},
				}, nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return strings.Contains(output, "Created worktree")
			},
		},
		{
			name: "infer project from context",
			args: []string{"feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "current-project",
				}, nil)
				mockPS.On("DiscoverProject", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
					Name:        "current-project",
					GitRepoPath: "/home/user/Projects/current-project",
				}, nil)
				mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.CreateWorktreeResult{
					Worktree: &domain.WorktreeInfo{},
				}, nil)
			},
			expectError: false,
		},
		{
			name:  "create worktree with -C flag outputs path only",
			args:  []string{"test-project/feature-branch"},
			flags: map[string]string{"cd": "true"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{Type: domain.ContextOutsideGit}, nil)
				mockPS.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
					Name:        "test-project",
					GitRepoPath: "/home/user/Projects/test-project",
				}, nil)
				mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.CreateWorktreeResult{
					Worktree: &domain.WorktreeInfo{
						Path:   "/home/user/Worktrees/test-project/feature-branch",
						Branch: "feature-branch",
					},
				}, nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return output == "/home/user/Worktrees/test-project/feature-branch\n" &&
					!strings.Contains(output, "Created worktree")
			},
		},
		{
			name:  "create worktree without -C flag outputs success message",
			args:  []string{"test-project/feature-branch"},
			flags: map[string]string{"source": "main"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.On("GetCurrentContext").Return(&domain.Context{Type: domain.ContextOutsideGit}, nil)
				mockPS.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*domain.Context")).Return(&domain.ProjectInfo{
					Name:        "test-project",
					GitRepoPath: "/home/user/Projects/test-project",
				}, nil)
				mockWS.On("BranchExists", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				mockWS.On("CreateWorktree", mock.Anything, mock.AnythingOfType("*domain.CreateWorktreeRequest")).Return(&domain.CreateWorktreeResult{
					Worktree: &domain.WorktreeInfo{
						Path:   "/home/user/Worktrees/test-project/feature-branch",
						Branch: "feature-branch",
					},
				}, nil)
			},
			expectError: false,
			validateOut: func(output string) bool {
				return strings.Contains(output, "Created worktree") &&
					!strings.HasPrefix(output, "/")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWS := mocks.NewMockWorktreeService()
			mockCS := mocks.NewMockContextService()
			mockPS := mocks.NewMockProjectService()

			tc.setupMocks(mockWS, mockCS, mockPS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					WorktreeService: mockWS,
					ContextService:  mockCS,
					ProjectService:  mockPS,
				},
			}

			cmd := NewCreateCommand(config)
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
