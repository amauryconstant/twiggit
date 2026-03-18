package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	fixtures "twiggit/test/fixtures"
	"twiggit/test/mocks"
)

func TestNewContextService(t *testing.T) {
	detector := mocks.NewMockContextDetector()
	resolver := mocks.NewMockContextResolver()
	config := fixtures.NewTestConfig()

	service := NewContextService(detector, resolver, config)
	if service == nil {
		t.Error("expected service to be non-nil")
	}
}

func TestContextService_GetCurrentContext(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*mocks.MockContextDetector, *mocks.MockContextResolver)
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful context detection",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "feat", []domain.SuggestionOption(nil)).Return(
					[]*domain.ResolutionSuggestion{
						{
							Text:        "feature-branch",
							Description: "Feature branch",
						},
					}, nil)
			},
			expectedContext: fixtures.NewProjectContext(),
		},
		{
			name: "detect context fails",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, errors.New("detection failed"))
			},
			expectedError: "failed to detect context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mocks.NewMockContextDetector()
			resolver := mocks.NewMockContextResolver()
			config := fixtures.NewTestConfig()
			service := NewContextService(detector, resolver, config)

			tt.setupMock(detector, resolver)

			t.Cleanup(func() {
				detector.AssertExpectations(t)
			})

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
		})
	}
}

func TestContextService_DetectContextFromPath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		setupMock       func(*mocks.MockContextDetector)
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful detection from valid path",
			path: "/test/path",
			setupMock: func(detector *mocks.MockContextDetector) {
				detector.On("DetectContext", "/test/path").Return(
					fixtures.NewWorktreeContext(), nil)
			},
			expectedContext: fixtures.NewWorktreeContext(),
		},
		{
			name: "detection fails",
			path: "/invalid/path",
			setupMock: func(detector *mocks.MockContextDetector) {
				detector.On("DetectContext", "/invalid/path").Return(
					nil, errors.New("detection failed"))
			},
			expectedError: "failed to detect context from path /invalid/path",
		},
		{
			name: "empty path",
			path: "",
			setupMock: func(detector *mocks.MockContextDetector) {
				detector.On("DetectContext", "").Return(
					nil, errors.New("detection failed"))
			},
			expectedError: "failed to detect context from path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mocks.NewMockContextDetector()
			resolver := mocks.NewMockContextResolver()
			config := fixtures.NewTestConfig()
			service := NewContextService(detector, resolver, config)

			tt.setupMock(detector)

			t.Cleanup(func() {
				detector.AssertExpectations(t)
			})

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
		})
	}
}

func TestContextService_ResolveIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		setupMock      func(*mocks.MockContextDetector, *mocks.MockContextResolver)
		expectedResult *domain.ResolutionResult
		expectedError  string
	}{
		{
			name:       "successful resolution",
			identifier: "main",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "main").Return(
					fixtures.NewProjectResolutionResult(), nil)
			},
			expectedResult: fixtures.NewProjectResolutionResult(),
		},
		{
			name: "get current context fails",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, errors.New("detection failed"))
			},
			expectedError: "failed to get current context",
		},
		{
			name:       "resolver fails",
			identifier: "invalid-branch",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid-branch").Return(
					nil, errors.New("resolution failed"))
			},
			expectedError: "failed to resolve identifier 'invalid-branch'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mocks.NewMockContextDetector()
			resolver := mocks.NewMockContextResolver()
			config := fixtures.NewTestConfig()
			service := NewContextService(detector, resolver, config)

			tt.setupMock(detector, resolver)

			t.Cleanup(func() {
				detector.AssertExpectations(t)
				resolver.AssertExpectations(t)
			})

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
		})
	}
}

func TestContextService_ResolveIdentifierFromContext(t *testing.T) {
	tests := []struct {
		name           string
		context        *domain.Context
		identifier     string
		setupMock      func(*mocks.MockContextDetector, *mocks.MockContextResolver)
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
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
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
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid").Return(
					nil, errors.New("resolution failed"))
			},
			expectedError: "failed to resolve identifier 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mocks.NewMockContextDetector()
			resolver := mocks.NewMockContextResolver()
			config := fixtures.NewTestConfig()
			service := NewContextService(detector, resolver, config)

			tt.setupMock(detector, resolver)

			t.Cleanup(func() {
				resolver.AssertExpectations(t)
			})

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
		})
	}
}

func TestContextService_GetCompletionSuggestions(t *testing.T) {
	tests := []struct {
		name                string
		partial             string
		setupMock           func(*mocks.MockContextDetector, *mocks.MockContextResolver)
		expectedSuggestions []*domain.ResolutionSuggestion
		expectedError       string
	}{
		{
			name:    "successful suggestions",
			partial: "feat",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "feat", []domain.SuggestionOption(nil)).Return(
					fixtures.NewFeatureSuggestions(), nil)
			},
			expectedSuggestions: fixtures.NewFeatureSuggestions(),
		},
		{
			name: "get current context fails",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, errors.New("detection failed"))
			},
			expectedError: "failed to get current context",
		},
		{
			name:    "suggestions_fail",
			partial: "invalid",
			setupMock: func(detector *mocks.MockContextDetector, resolver *mocks.MockContextResolver) {
				detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "invalid", []domain.SuggestionOption(nil)).Return(
					nil, errors.New("suggestions failed"))
			},
			expectedError: "failed to get completion suggestions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := mocks.NewMockContextDetector()
			resolver := mocks.NewMockContextResolver()
			config := fixtures.NewTestConfig()
			service := NewContextService(detector, resolver, config)

			tt.setupMock(detector, resolver)

			t.Cleanup(func() {
				detector.AssertExpectations(t)
				resolver.AssertExpectations(t)
			})

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
		})
	}
}
