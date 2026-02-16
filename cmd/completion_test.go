package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type CompletionTestSuite struct {
	suite.Suite
	config *CommandConfig
}

func TestCompletion(t *testing.T) {
	suite.Run(t, new(CompletionTestSuite))
}

func (s *CompletionTestSuite) SetupTest() {
	s.config = &CommandConfig{
		Services: &ServiceContainer{},
		Config:   &domain.Config{},
	}
}

func (s *CompletionTestSuite) TestGetCompletionTimeout() {
	tests := []struct {
		name     string
		config   *domain.Config
		expected time.Duration
	}{
		{
			name:     "nil config returns default",
			config:   nil,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "empty timeout returns default",
			config:   &domain.Config{},
			expected: 500 * time.Millisecond,
		},
		{
			name: "valid timeout from config",
			config: &domain.Config{
				Completion: domain.CompletionConfig{Timeout: "1s"},
			},
			expected: time.Second,
		},
		{
			name: "invalid timeout returns default",
			config: &domain.Config{
				Completion: domain.CompletionConfig{Timeout: "invalid"},
			},
			expected: 500 * time.Millisecond,
		},
		{
			name: "custom timeout 250ms",
			config: &domain.Config{
				Completion: domain.CompletionConfig{Timeout: "250ms"},
			},
			expected: 250 * time.Millisecond,
		},
		{
			name: "custom timeout 2s",
			config: &domain.Config{
				Completion: domain.CompletionConfig{Timeout: "2s"},
			},
			expected: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := getCompletionTimeout(tt.config)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *CompletionTestSuite) TestSuggestionsToCarapaceAction_EmptySuggestions() {
	action := suggestionsToCarapaceAction([]*domain.ResolutionSuggestion{})

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestSuggestionsToCarapaceAction_NilSuggestions() {
	action := suggestionsToCarapaceAction(nil)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestSuggestionsToCarapaceAction_WithSuggestions() {
	suggestions := []*domain.ResolutionSuggestion{
		{Text: "main", Description: "Project root directory"},
		{Text: "feature-1", Description: "Worktree for branch feature-1"},
		{Text: "develop", Description: "Branch develop (create worktree)"},
	}

	action := suggestionsToCarapaceAction(suggestions)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestSuggestionsToCarapaceAction_SingleSuggestion() {
	suggestions := []*domain.ResolutionSuggestion{
		{Text: "main", Description: "Project root directory", Type: domain.PathTypeProject, ProjectName: "test-project"},
	}

	action := suggestionsToCarapaceAction(suggestions)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestActionWorktreeTarget_ReturnsAction() {
	action := actionWorktreeTarget(s.config)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestActionBranches_ReturnsAction() {
	action := actionBranches(s.config)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestActionBranchesForProject_ReturnsAction() {
	action := actionBranchesForProject("myproject", s.config)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestActionWorktreeTarget_WithExistingOnly() {
	action := actionWorktreeTarget(s.config, domain.WithExistingOnly())

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestActionWorktreeTarget_WithNilConfig() {
	nilConfig := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   nil,
	}

	action := actionWorktreeTarget(nilConfig)

	s.NotNil(action)
}

func (s *CompletionTestSuite) TestGetCompletionTimeout_NilConfigSafety() {
	result := getCompletionTimeout(nil)

	s.Equal(500*time.Millisecond, result)
}
