package mocks

import (
	"context"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

// MockWorktreeService is a mock implementation of application.WorktreeService
type MockWorktreeService struct {
	// Configurable functions for testing
	CreateWorktreeFunc    func(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error)
	DeleteWorktreeFunc    func(ctx context.Context, req *domain.DeleteWorktreeRequest) error
	ListWorktreesFunc     func(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error)
	GetWorktreeStatusFunc func(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error)
	ValidateWorktreeFunc  func(ctx context.Context, worktreePath string) error
}

// NewMockWorktreeService creates a new MockWorktreeService with default behavior
func NewMockWorktreeService() *MockWorktreeService {
	return &MockWorktreeService{}
}

// CreateWorktree calls the mock function or returns default
func (m *MockWorktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(ctx, req)
	}
	return nil, nil
}

// DeleteWorktree calls the mock function or returns default
func (m *MockWorktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	if m.DeleteWorktreeFunc != nil {
		return m.DeleteWorktreeFunc(ctx, req)
	}
	return nil
}

// ListWorktrees calls the mock function or returns default
func (m *MockWorktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	if m.ListWorktreesFunc != nil {
		return m.ListWorktreesFunc(ctx, req)
	}
	return nil, nil
}

// GetWorktreeStatus calls the mock function or returns default
func (m *MockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	if m.GetWorktreeStatusFunc != nil {
		return m.GetWorktreeStatusFunc(ctx, worktreePath)
	}
	return nil, nil
}

// ValidateWorktree calls the mock function or returns default
func (m *MockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	if m.ValidateWorktreeFunc != nil {
		return m.ValidateWorktreeFunc(ctx, worktreePath)
	}
	return nil
}

// MockProjectService is a mock implementation of application.ProjectService
type MockProjectService struct {
	// Configurable functions for testing
	DiscoverProjectFunc func(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error)
	ValidateProjectFunc func(ctx context.Context, projectPath string) error
	ListProjectsFunc    func(ctx context.Context) ([]*domain.ProjectInfo, error)
	GetProjectInfoFunc  func(ctx context.Context, projectPath string) (*domain.ProjectInfo, error)
}

// NewMockProjectService creates a new MockProjectService with default behavior
func NewMockProjectService() *MockProjectService {
	return &MockProjectService{}
}

// DiscoverProject calls the mock function or returns default
func (m *MockProjectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	if m.DiscoverProjectFunc != nil {
		return m.DiscoverProjectFunc(ctx, projectName, context)
	}
	return nil, nil
}

// ValidateProject calls the mock function or returns default
func (m *MockProjectService) ValidateProject(ctx context.Context, projectPath string) error {
	if m.ValidateProjectFunc != nil {
		return m.ValidateProjectFunc(ctx, projectPath)
	}
	return nil
}

// ListProjects calls the mock function or returns default
func (m *MockProjectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc(ctx)
	}
	return nil, nil
}

// GetProjectInfo calls the mock function or returns default
func (m *MockProjectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	if m.GetProjectInfoFunc != nil {
		return m.GetProjectInfoFunc(ctx, projectPath)
	}
	return nil, nil
}

// MockNavigationService is a mock implementation of application.NavigationService
type MockNavigationService struct {
	// Configurable functions for testing
	ResolvePathFunc              func(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error)
	ValidatePathFunc             func(ctx context.Context, path string) error
	GetNavigationSuggestionsFunc func(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error)
}

// NewMockNavigationService creates a new MockNavigationService with default behavior
func NewMockNavigationService() *MockNavigationService {
	return &MockNavigationService{}
}

// ResolvePath calls the mock function or returns default
func (m *MockNavigationService) ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	if m.ResolvePathFunc != nil {
		return m.ResolvePathFunc(ctx, req)
	}
	return nil, nil
}

// ValidatePath calls the mock function or returns default
func (m *MockNavigationService) ValidatePath(ctx context.Context, path string) error {
	if m.ValidatePathFunc != nil {
		return m.ValidatePathFunc(ctx, path)
	}
	return nil
}

// GetNavigationSuggestions calls the mock function or returns default
func (m *MockNavigationService) GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	if m.GetNavigationSuggestionsFunc != nil {
		return m.GetNavigationSuggestionsFunc(ctx, context, partial)
	}
	return nil, nil
}

// MockContextService is a mock implementation of application.ContextService
type MockContextService struct {
	application.ContextService

	// Configurable functions for testing
	GetCurrentContextFunc                   func() (*domain.Context, error)
	DetectContextFromPathFunc               func(path string) (*domain.Context, error)
	ResolveIdentifierFunc                   func(identifier string) (*domain.ResolutionResult, error)
	ResolveIdentifierFromContextFunc        func(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error)
	GetCompletionSuggestionsFunc            func(partial string) ([]*domain.ResolutionSuggestion, error)
	GetCompletionSuggestionsFromContextFunc func(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error)
}

// NewMockContextService creates a new MockContextService with default behavior
func NewMockContextService() *MockContextService {
	return &MockContextService{}
}

// GetCurrentContext calls the mock function or returns default
func (m *MockContextService) GetCurrentContext() (*domain.Context, error) {
	if m.GetCurrentContextFunc != nil {
		return m.GetCurrentContextFunc()
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

// DetectContextFromPath calls the mock function or returns default
func (m *MockContextService) DetectContextFromPath(path string) (*domain.Context, error) {
	if m.DetectContextFromPathFunc != nil {
		return m.DetectContextFromPathFunc(path)
	}
	return &domain.Context{Type: domain.ContextOutsideGit}, nil
}

// ResolveIdentifier calls the mock function or returns default
func (m *MockContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	if m.ResolveIdentifierFunc != nil {
		return m.ResolveIdentifierFunc(identifier)
	}
	return &domain.ResolutionResult{}, nil
}

// ResolveIdentifierFromContext calls the mock function or returns default
func (m *MockContextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	if m.ResolveIdentifierFromContextFunc != nil {
		return m.ResolveIdentifierFromContextFunc(ctx, identifier)
	}
	return &domain.ResolutionResult{}, nil
}

// GetCompletionSuggestions calls the mock function or returns default
func (m *MockContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	if m.GetCompletionSuggestionsFunc != nil {
		return m.GetCompletionSuggestionsFunc(partial)
	}
	return nil, nil
}

// GetCompletionSuggestionsFromContext calls the mock function or returns default
func (m *MockContextService) GetCompletionSuggestionsFromContext(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	if m.GetCompletionSuggestionsFromContextFunc != nil {
		return m.GetCompletionSuggestionsFromContextFunc(ctx, partial)
	}
	return nil, nil
}

// MockShellService is a mock implementation of application.ShellService
type MockShellService struct {
	// Configurable functions for testing
	SetupShellFunc           func(ctx context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error)
	ValidateInstallationFunc func(ctx context.Context, req *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error)
	GenerateWrapperFunc      func(ctx context.Context, req *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error)
}

// NewMockShellService creates a new MockShellService with default behavior
func NewMockShellService() *MockShellService {
	return &MockShellService{}
}

// SetupShell calls the mock function or returns default
func (m *MockShellService) SetupShell(ctx context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error) {
	if m.SetupShellFunc != nil {
		return m.SetupShellFunc(ctx, req)
	}
	return &domain.SetupShellResult{}, nil
}

// ValidateInstallation calls the mock function or returns default
func (m *MockShellService) ValidateInstallation(ctx context.Context, req *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error) {
	if m.ValidateInstallationFunc != nil {
		return m.ValidateInstallationFunc(ctx, req)
	}
	return &domain.ValidateInstallationResult{}, nil
}

// GenerateWrapper calls the mock function or returns default
func (m *MockShellService) GenerateWrapper(ctx context.Context, req *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error) {
	if m.GenerateWrapperFunc != nil {
		return m.GenerateWrapperFunc(ctx, req)
	}
	return &domain.GenerateWrapperResult{}, nil
}
