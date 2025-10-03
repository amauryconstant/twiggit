package services

import (
	"context"
	"fmt"

	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockNavigationService provides functional setup builders for navigation service testing
type MockNavigationService struct {
	mock.Mock
}

// NewMockNavigationService creates a new MockNavigationService instance
func NewMockNavigationService() *MockNavigationService {
	return &MockNavigationService{}
}

// ResolvePath mocks resolving a path
func (m *MockNavigationService) ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ResolutionResult), args.Error(1)
}

// ValidatePath mocks validating a path
func (m *MockNavigationService) ValidatePath(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// GetNavigationSuggestions mocks getting navigation suggestions
func (m *MockNavigationService) GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	args := m.Called(ctx, context, partial)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ResolutionSuggestion), args.Error(1)
}

// Functional setup methods for fluent interface
func (m *MockNavigationService) SetupResolvePathSuccess(target string) *MockNavigationSetup {
	setup := &MockNavigationSetup{mock: m}
	return setup.ForResolvePath(target).WillSucceed()
}

func (m *MockNavigationService) SetupResolvePathError(target, errorMsg string) *MockNavigationSetup {
	setup := &MockNavigationSetup{mock: m}
	return setup.ForResolvePath(target).WillFail(errorMsg)
}

// MockNavigationSetup provides functional setup builder for navigation operations
type MockNavigationSetup struct {
	mock      *MockNavigationService
	operation string
	params    map[string]interface{}
}

func (s *MockNavigationSetup) ForResolvePath(target string) *MockNavigationSetup {
	s.operation = "ResolvePath"
	s.params = map[string]interface{}{
		"target": target,
	}
	return s
}

func (s *MockNavigationSetup) WithPath(path string) *MockNavigationSetup {
	s.params["path"] = path
	return s
}

func (s *MockNavigationSetup) WithType(pathType domain.PathType) *MockNavigationSetup {
	s.params["type"] = pathType
	return s
}

func (s *MockNavigationSetup) WillSucceed() *MockNavigationSetup {
	// Set default values if not provided
	path, hasPath := s.params["path"]
	if !hasPath {
		path = "/default/path"
	}
	pathType, hasType := s.params["type"]
	if !hasType {
		pathType = domain.PathTypeWorktree
	}

	result := &domain.ResolutionResult{
		ResolvedPath: path.(string),
		Type:         pathType.(domain.PathType),
		Explanation:  "Successfully resolved target",
	}

	s.mock.On("ResolvePath", mock.Anything, mock.MatchedBy(func(req *domain.ResolvePathRequest) bool {
		return req.Target == s.params["target"]
	})).Return(result, nil)

	return s
}

func (s *MockNavigationSetup) WillFail(errorMsg string) *MockNavigationSetup {
	s.mock.On("ResolvePath", mock.Anything, mock.MatchedBy(func(req *domain.ResolvePathRequest) bool {
		return req.Target == s.params["target"]
	})).Return(nil, fmt.Errorf("%s", errorMsg))

	return s
}
