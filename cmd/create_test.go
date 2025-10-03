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

// mockWorktreeService for testing
type mockWorktreeServiceForCreate struct {
	createWorktreeFunc func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error)
}

func (m *mockWorktreeServiceForCreate) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	if m.createWorktreeFunc != nil {
		return m.createWorktreeFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockWorktreeServiceForCreate) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	return nil
}

func (m *mockWorktreeServiceForCreate) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeServiceForCreate) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	return nil, nil
}

func (m *mockWorktreeServiceForCreate) ValidateWorktree(ctx context.Context, worktreePath string) error {
	return nil
}

// mockProjectService for testing
type mockProjectServiceForCreate struct {
	discoverProjectFunc func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error)
}

func (m *mockProjectServiceForCreate) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	if m.discoverProjectFunc != nil {
		return m.discoverProjectFunc(ctx, projectName, context)
	}
	return nil, nil
}

func (m *mockProjectServiceForCreate) ValidateProject(ctx context.Context, projectPath string) error {
	return nil
}

func (m *mockProjectServiceForCreate) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	return nil, nil
}

func (m *mockProjectServiceForCreate) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	return nil, nil
}

// mockContextService for testing
type mockContextServiceForCreate struct {
	getCurrentContextFunc func() (*domain.Context, error)
}

func (m *mockContextServiceForCreate) GetCurrentContext() (*domain.Context, error) {
	if m.getCurrentContextFunc != nil {
		return m.getCurrentContextFunc()
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForCreate) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

func (m *mockContextServiceForCreate) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextServiceForCreate) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockContextServiceForCreate) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}

func TestCreateCommand_MockInterfaces(t *testing.T) {
	// Interface compliance checks
	var _ services.WorktreeService = (*mockWorktreeServiceForCreate)(nil)
	var _ services.ProjectService = (*mockProjectServiceForCreate)(nil)
	var _ services.ContextServiceInterface = (*mockContextServiceForCreate)(nil)
}

func TestCreateCommand_Execute(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		flags        map[string]string
		setupMocks   func(*mockWorktreeServiceForCreate, *mockContextServiceForCreate, *mockProjectServiceForCreate)
		expectError  bool
		errorMessage string
		validateOut  func(string) bool
	}{
		{
			name:  "create worktree with project/branch",
			args:  []string{"test-project/feature-branch"},
			flags: map[string]string{"source": "main"},
			setupMocks: func(mockWS *mockWorktreeServiceForCreate, mockCS *mockContextServiceForCreate, mockPS *mockProjectServiceForCreate) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type: domain.ContextOutsideGit,
					}, nil
				}
				mockPS.discoverProjectFunc = func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
					return &domain.ProjectInfo{
						Name:        "test-project",
						GitRepoPath: "/home/user/Projects/test-project",
					}, nil
				}
				mockWS.createWorktreeFunc = func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
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
			setupMocks: func(mockWS *mockWorktreeServiceForCreate, mockCS *mockContextServiceForCreate, mockPS *mockProjectServiceForCreate) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type:        domain.ContextProject,
						ProjectName: "current-project",
					}, nil
				}
				mockPS.discoverProjectFunc = func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
					return &domain.ProjectInfo{
						Name:        "current-project",
						GitRepoPath: "/home/user/Projects/current-project",
					}, nil
				}
				mockWS.createWorktreeFunc = func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
					return &domain.WorktreeInfo{}, nil
				}
			},
			expectError: false,
		},
		{
			name: "invalid project/branch format",
			args: []string{"invalid-format"},
			setupMocks: func(mockWS *mockWorktreeServiceForCreate, mockCS *mockContextServiceForCreate, mockPS *mockProjectServiceForCreate) {
				mockCS.getCurrentContextFunc = func() (*domain.Context, error) {
					return &domain.Context{
						Type: domain.ContextOutsideGit,
					}, nil
				}
			},
			expectError:  true,
			errorMessage: "cannot infer project: not in a project context and no project specified",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockWS := &mockWorktreeServiceForCreate{}
			mockCS := &mockContextServiceForCreate{}
			mockPS := &mockProjectServiceForCreate{}

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
					assert.True(t, tc.validateOut(buf.String()))
				}
			}
		})
	}
}
