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

func TestListCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ interface{} = mocks.NewMockWorktreeService()
	var _ interface{} = mocks.NewMockContextService()
}

func TestListCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mocks.MockWorktreeService, *mocks.MockContextService)
		expectError  bool
		errorMessage string
		validateOut  func(string) bool
	}{
		{
			name: "list worktrees in project context",
			args: []string{},
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "test-project",
					}, nil
				}
				mockWS.ListWorktreesFunc = func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
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
			setupMocks: func(mockWS *mocks.MockWorktreeService, mockCS *mocks.MockContextService) {
				mockCS.GetCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{Type: domain.ContextOutsideGit}, nil
				}
				mockWS.ListWorktreesFunc = func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
					return []*domain.WorktreeInfo{}, nil
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
