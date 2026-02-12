package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

type NavigationServiceTestSuite struct {
	suite.Suite
	service application.NavigationService
	config  *domain.Config
}

func (s *NavigationServiceTestSuite) SetupTest() {
	s.config = domain.DefaultConfig()
	s.service = s.setupTestNavigationService()
}

func (s *NavigationServiceTestSuite) setupTestNavigationService() application.NavigationService {
	projectService := mocks.NewMockProjectService()
	contextService := mocks.NewMockContextService()
	return NewNavigationService(projectService, contextService, s.config)
}

func TestNavigationService(t *testing.T) {
	suite.Run(t, new(NavigationServiceTestSuite))
}

func (s *NavigationServiceTestSuite) TestResolvePath() {
	tests := []struct {
		name         string
		request      *domain.ResolvePathRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid branch resolution from project context",
			request: &domain.ResolvePathRequest{
				Target: "feature-branch",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError: false,
		},
		{
			name: "empty target",
			request: &domain.ResolvePathRequest{
				Target: "",
				Context: &domain.Context{
					Type: domain.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "validation failed for ResolvePathRequest.target: cannot be empty",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.ResolvePath(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *NavigationServiceTestSuite) TestValidatePath() {
	tests := []struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid path validation",
			path:        ".",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorMessage: "validation failed for ValidatePath.path: cannot be empty",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.service.ValidatePath(context.Background(), tc.path)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *NavigationServiceTestSuite) TestGetNavigationSuggestions() {
	tests := []struct {
		name        string
		context     *domain.Context
		partial     string
		expectError bool
	}{
		{
			name: "valid suggestions from project context",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
			},
			partial:     "feat",
			expectError: false,
		},
		{
			name: "suggestions from outside context",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
			},
			partial:     "test",
			expectError: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			projectService := mocks.NewMockProjectService()
			contextService := mocks.NewMockContextService()

			contextService.GetCompletionSuggestionsFromContextFunc = func(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
				return []*domain.ResolutionSuggestion{
					{
						Text:        "feature-branch",
						Description: "Feature branch",
						Type:        domain.PathTypeWorktree,
						ProjectName: "test-project",
						BranchName:  "feature-branch",
					},
				}, nil
			}

			contextService.GetCompletionSuggestionsFunc = func(partial string) ([]*domain.ResolutionSuggestion, error) {
				return []*domain.ResolutionSuggestion{
					{
						Text:        "test-project",
						Description: "Test project",
						Type:        domain.PathTypeProject,
						ProjectName: "test-project",
					},
				}, nil
			}

			service := NewNavigationService(projectService, contextService, s.config)
			result, err := service.GetNavigationSuggestions(context.Background(), tc.context, tc.partial)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}
