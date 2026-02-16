package mocks

import (
	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockContextResolver is a mock implementation of domain.ContextResolver
type MockContextResolver struct {
	mock.Mock
}

// NewMockContextResolver creates a new MockContextResolver
func NewMockContextResolver() *MockContextResolver {
	return &MockContextResolver{}
}

// ResolveIdentifier provides a mock function with given fields: ctx, identifier
func (m *MockContextResolver) ResolveIdentifier(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ResolutionResult), args.Error(1)
}

// GetResolutionSuggestions provides a mock function with given fields: ctx, partial, opts
func (m *MockContextResolver) GetResolutionSuggestions(ctx *domain.Context, partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error) {
	args := m.Called(ctx, partial, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ResolutionSuggestion), args.Error(1)
}
