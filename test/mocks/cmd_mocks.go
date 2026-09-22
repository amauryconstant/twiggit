package mocks

import (
	"context"

	"twiggit/internal/core"

	"github.com/stretchr/testify/mock"
)

// MockWorktreeService is a mock implementation of application.WorktreeService
type MockWorktreeService struct {
	mock.Mock
}

// NewMockWorktreeService creates a new MockWorktreeService
func NewMockWorktreeService() *MockWorktreeService {
	return &MockWorktreeService{}
}

// CreateWorktree mocks creating a worktree
func (m *MockWorktreeService) CreateWorktree(ctx context.Context, req *core.CreateWorktreeRequest) (*core.CreateWorktreeResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.CreateWorktreeResult), args.Error(1)
}

// DeleteWorktree mocks deleting a worktree
func (m *MockWorktreeService) DeleteWorktree(ctx context.Context, req *core.DeleteWorktreeRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// ListWorktrees mocks listing worktrees
func (m *MockWorktreeService) ListWorktrees(ctx context.Context, req *core.ListWorktreesRequest) ([]*core.WorktreeInfo, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.WorktreeInfo), args.Error(1)
}

// GetWorktreeStatus mocks getting worktree status
func (m *MockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*core.WorktreeStatus, error) {
	args := m.Called(ctx, worktreePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.WorktreeStatus), args.Error(1)
}

// ValidateWorktree mocks validating a worktree
func (m *MockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	args := m.Called(ctx, worktreePath)
	return args.Error(0)
}

// PruneMergedWorktrees mocks pruning merged worktrees
func (m *MockWorktreeService) PruneMergedWorktrees(ctx context.Context, req *core.PruneWorktreesRequest) (*core.PruneWorktreesResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.PruneWorktreesResult), args.Error(1)
}

// BranchExists mocks checking if a branch exists
func (m *MockWorktreeService) BranchExists(ctx context.Context, projectPath, branchName string) (bool, error) {
	args := m.Called(ctx, projectPath, branchName)
	return args.Bool(0), args.Error(1)
}

// IsBranchMerged mocks checking if a branch is merged
func (m *MockWorktreeService) IsBranchMerged(ctx context.Context, worktreePath, branchName string) (bool, error) {
	args := m.Called(ctx, worktreePath, branchName)
	return args.Bool(0), args.Error(1)
}

// GetWorktreeByPath mocks getting a worktree by path
func (m *MockWorktreeService) GetWorktreeByPath(ctx context.Context, projectPath, worktreePath string) (*core.WorktreeInfo, error) {
	args := m.Called(ctx, projectPath, worktreePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.WorktreeInfo), args.Error(1)
}

// MockProjectService is a mock implementation of application.ProjectService
type MockProjectService struct {
	mock.Mock
}

// NewMockProjectService creates a new MockProjectService
func NewMockProjectService() *MockProjectService {
	return &MockProjectService{}
}

// DiscoverProject mocks discovering a project
func (m *MockProjectService) DiscoverProject(ctx context.Context, projectName string, context *core.Context) (*core.ProjectInfo, error) {
	args := m.Called(ctx, projectName, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ProjectInfo), args.Error(1)
}

// ValidateProject mocks validating a project
func (m *MockProjectService) ValidateProject(ctx context.Context, projectPath string) error {
	args := m.Called(ctx, projectPath)
	return args.Error(0)
}

// ListProjects mocks listing projects
func (m *MockProjectService) ListProjects(ctx context.Context) ([]*core.ProjectInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ProjectInfo), args.Error(1)
}

// ListProjectSummaries mocks listing project summaries
func (m *MockProjectService) ListProjectSummaries(ctx context.Context) ([]*core.ProjectSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ProjectSummary), args.Error(1)
}

// GetProjectInfo mocks getting project info
func (m *MockProjectService) GetProjectInfo(ctx context.Context, projectPath string) (*core.ProjectInfo, error) {
	args := m.Called(ctx, projectPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ProjectInfo), args.Error(1)
}

// MockNavigationService is a mock implementation of application.NavigationService
type MockNavigationService struct {
	mock.Mock
}

// NewMockNavigationService creates a new MockNavigationService
func NewMockNavigationService() *MockNavigationService {
	return &MockNavigationService{}
}

// ResolvePath mocks resolving a path
func (m *MockNavigationService) ResolvePath(ctx context.Context, req *core.ResolvePathRequest) (*core.ResolutionResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ResolutionResult), args.Error(1)
}

// ValidatePath mocks validating a path
func (m *MockNavigationService) ValidatePath(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// GetNavigationSuggestions mocks getting navigation suggestions
func (m *MockNavigationService) GetNavigationSuggestions(ctx context.Context, context *core.Context, partial string) ([]*core.ResolutionSuggestion, error) {
	args := m.Called(ctx, context, partial)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ResolutionSuggestion), args.Error(1)
}

// MockContextService is a mock implementation of application.ContextService
type MockContextService struct {
	mock.Mock
}

// NewMockContextService creates a new MockContextService
func NewMockContextService() *MockContextService {
	return &MockContextService{}
}

// GetCurrentContext mocks getting current context
func (m *MockContextService) GetCurrentContext() (*core.Context, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Context), args.Error(1)
}

// DetectContextFromPath mocks detecting context from path
func (m *MockContextService) DetectContextFromPath(path string) (*core.Context, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Context), args.Error(1)
}

// ResolveIdentifier mocks resolving an identifier
func (m *MockContextService) ResolveIdentifier(identifier string) (*core.ResolutionResult, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ResolutionResult), args.Error(1)
}

// ResolveIdentifierFromContext mocks resolving identifier from context
func (m *MockContextService) ResolveIdentifierFromContext(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ResolutionResult), args.Error(1)
}

// GetCompletionSuggestions mocks getting completion suggestions
func (m *MockContextService) GetCompletionSuggestions(partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
	args := m.Called(variadicArgs(partial, opts)...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ResolutionSuggestion), args.Error(1)
}

// GetCompletionSuggestionsFromContext mocks getting completion suggestions from context
func (m *MockContextService) GetCompletionSuggestionsFromContext(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
	partialArgs := variadicArgs(partial, opts)
	args := m.Called(append([]any{ctx}, partialArgs...)...)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ResolutionSuggestion), args.Error(1)
}

// MockShellService is a mock implementation of application.ShellService
type MockShellService struct {
	mock.Mock
}

// NewMockShellService creates a new MockShellService
func NewMockShellService() *MockShellService {
	return &MockShellService{}
}

// SetupShell mocks setting up shell
func (m *MockShellService) SetupShell(ctx context.Context, req *core.SetupShellRequest) (*core.SetupShellResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.SetupShellResult), args.Error(1)
}

// ValidateInstallation mocks validating installation
func (m *MockShellService) ValidateInstallation(ctx context.Context, req *core.ValidateInstallationRequest) (*core.ValidateInstallationResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ValidateInstallationResult), args.Error(1)
}

// GenerateWrapper mocks generating wrapper
func (m *MockShellService) GenerateWrapper(ctx context.Context, req *core.GenerateWrapperRequest) (*core.GenerateWrapperResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.GenerateWrapperResult), args.Error(1)
}
