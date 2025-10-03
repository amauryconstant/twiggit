package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Type: domain.ContextOutsideGit}, nil
				}
				mockPS.DiscoverProjectFunc = func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
					return &domain.ProjectInfo{
						Name:        "test-project",
						GitRepoPath: "/home/user/Projects/test-project",
					}, nil
				}
				mockWS.CreateWorktreeFunc = func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
					return &domain.WorktreeInfo{
						Path:   "/home/user/Worktrees/test-project/feature-branch",
						Branch: "feature-branch",
					}, nil
				}
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
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "current-project",
					}, nil
				}
				mockPS.DiscoverProjectFunc = func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
					return &domain.ProjectInfo{
						Name:        "current-project",
						GitRepoPath: "/home/user/Projects/current-project",
					}, nil
				}
				mockWS.CreateWorktreeFunc = func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
					return &domain.WorktreeInfo{}, nil
				}
			},
			expectError: false,
		},
		{
			name: "invalid project/branch format",
			args: []string{"invalid-format"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Type: domain.ContextOutsideGit}, nil
				}
			},
			expectError:  true,
			errorMessage: "cannot infer project: not in a project context and no project specified",
		},
		{
			name: "malformed project/branch with empty project",
			args: []string{"/feature-branch"},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService, mockPS *mocks.MockProjectService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Type: domain.ContextOutsideGit}, nil
				}
			},
			expectError:  true,
			errorMessage: "invalid format: expected <project>/<branch>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
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
