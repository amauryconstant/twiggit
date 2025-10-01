package mocks

import (
	"github.com/stretchr/testify/mock"
	"twiggit/internal/domain"
)

// ContextResolverMock is a mock implementation of domain.ContextResolver
type ContextResolverMock struct {
	mock.Mock
}

// ResolveIdentifier provides a mock function with given fields: ctx, identifier
func (m *ContextResolverMock) ResolveIdentifier(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ResolutionResult), args.Error(1)
}

// GetResolutionSuggestions provides a mock function with given fields: ctx, partial
func (m *ContextResolverMock) GetResolutionSuggestions(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	args := m.Called(ctx, partial)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ResolutionSuggestion), args.Error(1)
}
