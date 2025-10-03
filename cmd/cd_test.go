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

func TestCDCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ interface{} = mocks.NewMockNavigationService()
	var _ interface{} = mocks.NewMockContextService()
}

func TestCDCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		setupMocks   func(*mocks.MockNavigationService, *mocks.MockContextService)
		expectError  bool
		errorMessage string
		expectedPath string
	}{
		{
			name: "cd to worktree with branch name",
			args: []string{"feature-branch"},
			setupMocks: func(mockNS *mocks.MockNavigationService, mockCS *mocks.MockContextService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "test-project",
					}, nil
				}
				mockNS.ResolvePathFunc = func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{
						ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
					}, nil
				}
			},
			expectError:  false,
			expectedPath: "/home/user/Worktrees/test-project/feature-branch",
		},
		{
			name: "cd to default worktree",
			args: []string{},
			setupMocks: func(mockNS *mocks.MockNavigationService, mockCS *mocks.MockContextService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextWorktree,
						ProjectName: "test-project",
						BranchName:  "main",
					}, nil
				}
				mockNS.ResolvePathFunc = func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
					return &domain.ResolutionResult{
						ResolvedPath: "/home/user/Worktrees/test-project/main",
					}, nil
				}
			},
			expectError:  false,
			expectedPath: "/home/user/Worktrees/test-project/main",
		},
		{
			name: "no target and no default",
			args: []string{},
			setupMocks: func(mockNS *mocks.MockNavigationService, mockCS *mocks.MockContextService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Type: domain.ContextOutsideGit}, nil
				}
			},
			expectError:  true,
			errorMessage: "no target specified and no default worktree in context",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockNS := mocks.NewMockNavigationService()
			mockCS := mocks.NewMockContextService()

			tc.setupMocks(mockNS, mockCS)

			config := &CommandConfig{
				Services: &ServiceContainer{
					NavigationService: mockNS,
					ContextService:    mockCS,
				},
			}

			cmd := NewCDCommand(config)
			cmd.SetArgs(tc.args)

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
				if tc.expectedPath != "" {
					assert.Equal(t, tc.expectedPath, strings.TrimSpace(buf.String()))
				}
			}
		})
	}
}
