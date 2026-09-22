package mocks

import (
	"twiggit/internal/application"
	"twiggit/internal/core"

	"github.com/stretchr/testify/mock"
)

var _ application.ContextResolver = (*MockContextResolver)(nil)

// MockContextResolver is a mock implementation of application.ContextResolver
type MockContextResolver struct {
	mock.Mock
}

// NewMockContextResolver creates a new MockContextResolver
func NewMockContextResolver() *MockContextResolver {
	return &MockContextResolver{}
}

// ResolveIdentifier provides a mock function with given fields: ctx, identifier
func (m *MockContextResolver) ResolveIdentifier(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ResolutionResult), args.Error(1)
}

// GetResolutionSuggestions provides a mock function with given fields: ctx, partial, opts
func (m *MockContextResolver) GetResolutionSuggestions(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
	args := m.Called(ctx, partial, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ResolutionSuggestion), args.Error(1)
}
