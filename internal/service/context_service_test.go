package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	fixtures "twiggit/test/fixtures"
	"twiggit/test/mocks"
)

type ContextServiceTestSuite struct {
	suite.Suite
	detector *mocks.MockContextDetector
	resolver *mocks.MockContextResolver
	service  application.ContextService
	config   *domain.Config
}

func (s *ContextServiceTestSuite) SetupTest() {
	s.detector = mocks.NewMockContextDetector()
	s.resolver = mocks.NewMockContextResolver()
	s.config = fixtures.NewTestConfig()
	s.service = NewContextService(s.detector, s.resolver, s.config)
}

func TestContextService(t *testing.T) {
	suite.Run(t, new(ContextServiceTestSuite))
}

func (s *ContextServiceTestSuite) TestNewContextService() {
	detector := mocks.NewMockContextDetector()
	resolver := mocks.NewMockContextResolver()
	config := fixtures.NewTestConfig()

	service := NewContextService(detector, resolver, config)
	s.NotNil(service)
}

func (s *ContextServiceTestSuite) TestGetCurrentContext() {
	tests := []struct {
		name            string
		setupMock       func()
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful context detection",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				s.resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "feat", []domain.SuggestionOption(nil)).Return(
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
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.detector.ExpectedCalls = nil
			s.resolver.ExpectedCalls = nil
			s.SetupTest()

			tt.setupMock()

			got, err := s.service.GetCurrentContext()

			if tt.expectedError != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedError)
				s.Nil(got)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedContext.Type, got.Type)
				s.Equal(tt.expectedContext.ProjectName, got.ProjectName)
			}

			s.detector.AssertExpectations(s.T())
		})
	}
}

func (s *ContextServiceTestSuite) TestDetectContextFromPath() {
	tests := []struct {
		name            string
		path            string
		setupMock       func()
		expectedContext *domain.Context
		expectedError   string
	}{
		{
			name: "successful detection from valid path",
			path: "/test/path",
			setupMock: func() {
				s.detector.On("DetectContext", "/test/path").Return(
					fixtures.NewWorktreeContext(), nil)
			},
			expectedContext: fixtures.NewWorktreeContext(),
		},
		{
			name: "detection fails",
			path: "/invalid/path",
			setupMock: func() {
				s.detector.On("DetectContext", "/invalid/path").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context from path /invalid/path",
		},
		{
			name: "empty path",
			path: "",
			setupMock: func() {
				s.detector.On("DetectContext", "").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to detect context from path",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.detector.ExpectedCalls = nil
			s.resolver.ExpectedCalls = nil
			s.SetupTest()

			tt.setupMock()

			got, err := s.service.DetectContextFromPath(tt.path)

			if tt.expectedError != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedError)
				s.Nil(got)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedContext.Type, got.Type)
				s.Equal(tt.expectedContext.ProjectName, got.ProjectName)
			}

			s.detector.AssertExpectations(s.T())
		})
	}
}

func (s *ContextServiceTestSuite) TestResolveIdentifier() {
	tests := []struct {
		name           string
		identifier     string
		setupMock      func()
		expectedResult *domain.ResolutionResult
		expectedError  string
	}{
		{
			name:       "successful resolution",
			identifier: "main",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				s.resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "main").Return(
					fixtures.NewProjectResolutionResult(), nil)
			},
			expectedResult: fixtures.NewProjectResolutionResult(),
		},
		{
			name: "get current context fails",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get current context",
		},
		{
			name:       "resolver fails",
			identifier: "invalid-branch",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				s.resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid-branch").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to resolve identifier 'invalid-branch'",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.detector.ExpectedCalls = nil
			s.resolver.ExpectedCalls = nil
			s.SetupTest()

			tt.setupMock()

			got, err := s.service.ResolveIdentifier(tt.identifier)

			if tt.expectedError != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedError)
				s.Nil(got)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedResult.Type, got.Type)
				s.Equal(tt.expectedResult.ProjectName, got.ProjectName)
			}

			s.detector.AssertExpectations(s.T())
			s.resolver.AssertExpectations(s.T())
		})
	}
}

func (s *ContextServiceTestSuite) TestResolveIdentifierFromContext() {
	tests := []struct {
		name           string
		context        *domain.Context
		identifier     string
		setupMock      func()
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
			setupMock: func() {
				s.resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "other-branch").Return(
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
			setupMock: func() {
				s.resolver.On("ResolveIdentifier", mock.AnythingOfType("*domain.Context"), "invalid").Return(
					nil, assert.AnError)
			},
			expectedError: "failed to resolve identifier 'invalid'",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.detector.ExpectedCalls = nil
			s.resolver.ExpectedCalls = nil
			s.SetupTest()

			tt.setupMock()

			got, err := s.service.ResolveIdentifierFromContext(tt.context, tt.identifier)

			if tt.expectedError != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedError)
				s.Nil(got)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedResult.Type, got.Type)
				s.Equal(tt.expectedResult.ProjectName, got.ProjectName)
			}

			s.resolver.AssertExpectations(s.T())
		})
	}
}

func (s *ContextServiceTestSuite) TestGetCompletionSuggestions() {
	tests := []struct {
		name                string
		partial             string
		setupMock           func()
		expectedSuggestions []*domain.ResolutionSuggestion
		expectedError       string
	}{
		{
			name:    "successful suggestions",
			partial: "feat",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				s.resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "feat", []domain.SuggestionOption(nil)).Return(
					fixtures.NewFeatureSuggestions(), nil)
			},
			expectedSuggestions: fixtures.NewFeatureSuggestions(),
		},
		{
			name: "get current context fails",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get current context",
		},
		{
			name:    "suggestions_fail",
			partial: "invalid",
			setupMock: func() {
				s.detector.On("DetectContext", mock.AnythingOfType("string")).Return(
					fixtures.NewProjectContext(), nil)
				s.resolver.On("GetResolutionSuggestions", mock.AnythingOfType("*domain.Context"), "invalid", []domain.SuggestionOption(nil)).Return(
					nil, assert.AnError)
			},
			expectedError: "failed to get completion suggestions",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.detector.ExpectedCalls = nil
			s.resolver.ExpectedCalls = nil
			s.SetupTest()

			tt.setupMock()

			got, err := s.service.GetCompletionSuggestions(tt.partial)

			if tt.expectedError != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.expectedError)
				s.Nil(got)
			} else {
				s.Require().NoError(err)
				s.Len(got, len(tt.expectedSuggestions))
				if len(tt.expectedSuggestions) > 0 {
					s.Equal(tt.expectedSuggestions[0].Text, got[0].Text)
				}
			}

			s.detector.AssertExpectations(s.T())
			s.resolver.AssertExpectations(s.T())
		})
	}
}
