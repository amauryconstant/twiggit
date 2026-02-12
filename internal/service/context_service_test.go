package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	fixtures "twiggit/test/fixtures"
	"twiggit/test/mocks"
)

func TestNewContextService(t *testing.T) {
	detector := &mocks.ContextDetectorMock{}
	resolver := &mocks.ContextResolverMock{}
	config := fixtures.NewTestConfig()

	service := NewContextService(detector, resolver, config)

	assert.NotNil(t, service)
}

func TestContextService_GetCurrentContext(t *testing.T) {
	tests := []struct {
		name            string
		setupmock       func(*mocks.ContextDetectorMock, *mocks.ContextResolverMock)
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful context detection",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
			},
			expectedContext: fixtures.NewProjectContext(),
		},
		{
			name: "detect context fails",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &mocks.ContextDetectorMock{}
			resolver := &mocks.ContextResolverMock{}
			config := fixtures.NewTestConfig()

			tt.setupmock(detector, resolver)

			service := NewContextService(detector, resolver, config)
			got, err := service.GetCurrentContext()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContext.Type, got.Type)
				assert.Equal(t, tt.expectedContext.ProjectName, got.ProjectName)
			}

			detector.AssertExpectations(t)
		})
	}
}

func TestContextService_DetectContextFromPath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		setupmock       func(*mocks.ContextDetectorMock, *mocks.ContextResolverMock)
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful detection from valid path",
			path: "/test/path",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", "/test/path").Return(
					fixtures.NewWorktreeContext(), nil)
			},
			expectedContext: fixtures.NewWorktreeContext(),
		},
		{
			name: "detection fails",
			path: "/invalid/path",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", "/invalid/path").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context from path /invalid/path",
		},
		{
			name: "empty path",
			path: "",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", "").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context from path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &mocks.ContextDetectorMock{}
			resolver := &mocks.ContextResolverMock{}
			config := fixtures.NewTestConfig()

			tt.setupmock(detector, resolver)

			service := NewContextService(detector, resolver, config)
			got, err := service.DetectContextFromPath(tt.path)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedContext.Type, got.Type)
				assert.Equal(t, tt.expectedContext.ProjectName, got.ProjectName)
			}

			detector.AssertExpectations(t)
		})
	}
}

func TestContextService_ResolveIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		setupmock      func(*mocks.ContextDetectorMock, *mocks.ContextResolverMock)
		expectedResult *domain.ResolutionResult
		expectedError  string
	}{
		{
			name:       "successful resolution",
			identifier: "main",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "main").Return(
					fixtures.NewProjectResolutionResult(), nil)
			},
			expectedResult: fixtures.NewProjectResolutionResult(),
		},
		{
			name: "get current context fails",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get current context",
		},
		{
			name:       "resolver fails",
			identifier: "invalid-branch",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid-branch").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to resolve identifier 'invalid-branch'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &mocks.ContextDetectorMock{}
			resolver := &mocks.ContextResolverMock{}
			config := fixtures.NewTestConfig()

			tt.setupmock(detector, resolver)

			service := NewContextService(detector, resolver, config)
			got, err := service.ResolveIdentifier(tt.identifier)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Type, got.Type)
				assert.Equal(t, tt.expectedResult.ProjectName, got.ProjectName)
			}

			detector.AssertExpectations(t)
			resolver.AssertExpectations(t)
		})
	}
}

func TestContextService_ResolveIdentifierFromContext(t *testing.T) {
	tests := []struct {
		name           string
		context        *domain.Context
		identifier     string
		setupmock      func(*mocks.ContextDetectorMock, *mocks.ContextResolverMock)
		expectedResult *domain.ResolutionResult
		expectedError  string
	}{
		{
			name: "successful resolution with provided context",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "test-project",
				BranchName:  "current-branch",
			},
			identifier: "other-branch",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "other-branch").Return(
					fixtures.NewWorktreeResolutionResult(), nil)
			},
			expectedResult: fixtures.NewWorktreeResolutionResult(),
		},
		{
			name: "resolver fails",
			context: &domain.Context{
				Type: domain.ContextProject,
			},
			identifier: "invalid",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to resolve identifier 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &mocks.ContextDetectorMock{}
			resolver := &mocks.ContextResolverMock{}
			config := fixtures.NewTestConfig()

			tt.setupmock(detector, resolver)

			service := NewContextService(detector, resolver, config)
			got, err := service.ResolveIdentifierFromContext(tt.context, tt.identifier)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Type, got.Type)
				assert.Equal(t, tt.expectedResult.ProjectName, got.ProjectName)
			}

			resolver.AssertExpectations(t)
		})
	}
}

func TestContextService_GetCompletionSuggestions(t *testing.T) {
	tests := []struct {
		name                string
		partial             string
		setupmock           func(*mocks.ContextDetectorMock, *mocks.ContextResolverMock)
		expectedSuggestions []*domain.ResolutionSuggestion
		expectedError       string
	}{
		{
			name:    "successful suggestions",
			partial: "feat",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "feat").Return(
					fixtures.NewFeatureSuggestions(), nil)
			},
			expectedSuggestions: fixtures.NewFeatureSuggestions(),
		},
		{
			name: "get current context fails",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get current context",
		},
		{
			name:    "suggestions fail",
			partial: "invalid",
			setupmock: func(detector *mocks.ContextDetectorMock, resolver *mocks.ContextResolverMock) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "invalid").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get completion suggestions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &mocks.ContextDetectorMock{}
			resolver := &mocks.ContextResolverMock{}
			config := fixtures.NewTestConfig()

			tt.setupmock(detector, resolver)

			service := NewContextService(detector, resolver, config)
			got, err := service.GetCompletionSuggestions(tt.partial)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, len(tt.expectedSuggestions))
				if len(tt.expectedSuggestions) > 0 {
					assert.Equal(t, tt.expectedSuggestions[0].Text, got[0].Text)
				}
			}

			detector.AssertExpectations(t)
			resolver.AssertExpectations(t)
		})
	}
}
