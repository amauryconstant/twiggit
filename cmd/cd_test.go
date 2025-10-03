package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
	"twiggit/internal/services"
)

// mockNavigationService for testing
type mockNavigationService struct {
	resolvePathFunc func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error)
}

func (m *mockNavigationService) ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	if m.resolvePathFunc != nil {
		return m.resolvePathFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockNavigationService) ValidatePath(ctx context.Context, path string) error {
	return nil
}

func (m *mockNavigationService) GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

// mockContextService for testing
type mockContextServiceForCD struct {
	getCurrentContextFunc func() (*domain.Context, error)
}

func (m *mockContextServiceForCD) GetCurrentContext() (*domain.Context, error) {
	if m.getCurrentContextFunc != nil {
		return m.getCurrentContextFunc()
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForCD) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForCD) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextServiceForCD) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextServiceForCD) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

func TestCDCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ services.NavigationService = (*mockNavigationService)(nil)
	var _ services.ContextServiceInterface = (*mockContextServiceForCD)(nil)
}

func TestCDCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		setupMocks   func(*mockNavigationService, *mockContextServiceForCD)
		expectError  bool
		errorMessage string
		expectedPath string
	}{
		{
			name: "cd to worktree with branch name",
			args: []string{"feature-branch"},
			setupMocks: func(mockNS *mockNavigationService, mockCS *mockContextServiceForCD) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "test-project",
					}, nil
				}
				mockNS.resolvePathFunc = func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
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
			setupMocks: func(mockNS *mockNavigationService, mockCS *mockContextServiceForCD) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextWorktree,
						ProjectName: "test-project",
						BranchName:  "main",
					}, nil
				}
				mockNS.resolvePathFunc = func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
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
			setupMocks: func(mockNS *mockNavigationService, mockCS *mockContextServiceForCD) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type: domain.ContextOutsideGit,
					}, nil
				}
			},
			expectError:  true,
			errorMessage: "no target specified and no default worktree in context",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockNS := &mockNavigationService{}
			mockCS := &mockContextServiceForCD{}

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
