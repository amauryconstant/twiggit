package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (s *ContextTestSuite) SetupTest() {
	// Setup if needed
}

// ContextType String() tests
func (s *ContextTestSuite) TestContextType_String() {
	testCases := []struct {
		name     string
		context  ContextType
		expected string
	}{
		{"unknown context", ContextUnknown, "unknown"},
		{"project context", ContextProject, "project"},
		{"worktree context", ContextWorktree, "worktree"},
		{"outside git context", ContextOutsideGit, "outside-git"},
		{"invalid context", ContextType(999), "unknown"},
		{"negative context", ContextType(-1), "unknown"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := tc.context.String()
			s.Equal(tc.expected, result)
		})
	}
}

// PathType String() tests
func (s *ContextTestSuite) TestPathType_String() {
	testCases := []struct {
		name     string
		pathType PathType
		expected string
	}{
		{"project path", PathTypeProject, "project"},
		{"worktree path", PathTypeWorktree, "worktree"},
		{"invalid path", PathTypeInvalid, "invalid"},
		{"undefined path", PathType(999), "invalid"},
		{"negative path", PathType(-1), "invalid"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := tc.pathType.String()
			s.Equal(tc.expected, result)
		})
	}
}

// Data structure tests
func (s *ContextTestSuite) TestContext_Structure() {
	context := &Context{
		Type:        ContextProject,
		ProjectName: "test-project",
		BranchName:  "main",
		Path:        "/test/path",
		Explanation: "Test explanation",
	}

	s.Equal(ContextProject, context.Type)
	s.Equal("test-project", context.ProjectName)
	s.Equal("main", context.BranchName)
	s.Equal("/test/path", context.Path)
	s.Equal("Test explanation", context.Explanation)
}

func (s *ContextTestSuite) TestResolutionResult_Structure() {
	result := &ResolutionResult{
		ResolvedPath: "/resolved/path",
		Type:         PathTypeWorktree,
		ProjectName:  "test-project",
		BranchName:   "feature-branch",
		Explanation:  "Resolved successfully",
	}

	s.Equal("/resolved/path", result.ResolvedPath)
	s.Equal(PathTypeWorktree, result.Type)
	s.Equal("test-project", result.ProjectName)
	s.Equal("feature-branch", result.BranchName)
	s.Equal("Resolved successfully", result.Explanation)
}

func (s *ContextTestSuite) TestResolutionSuggestion_Structure() {
	suggestion := &ResolutionSuggestion{
		Text:        "feature-branch",
		Description: "Feature branch suggestion",
		Type:        PathTypeWorktree,
		ProjectName: "test-project",
		BranchName:  "feature-branch",
	}

	s.Equal("feature-branch", suggestion.Text)
	s.Equal("Feature branch suggestion", suggestion.Description)
	s.Equal(PathTypeWorktree, suggestion.Type)
	s.Equal("test-project", suggestion.ProjectName)
	s.Equal("feature-branch", suggestion.BranchName)
}

// Edge case tests for zero values
func (s *ContextTestSuite) TestContext_ZeroValues() {
	context := &Context{}

	s.Equal(ContextUnknown, context.Type)
	s.Empty(context.ProjectName)
	s.Empty(context.BranchName)
	s.Empty(context.Path)
	s.Empty(context.Explanation)
}

func (s *ContextTestSuite) TestResolutionResult_ZeroValues() {
	result := &ResolutionResult{}

	s.Empty(result.ResolvedPath)
	s.Equal(PathTypeProject, result.Type) // Zero value should be PathTypeProject (iota = 0)
	s.Empty(result.ProjectName)
	s.Empty(result.BranchName)
	s.Empty(result.Explanation)
}

func (s *ContextTestSuite) TestResolutionSuggestion_ZeroValues() {
	suggestion := &ResolutionSuggestion{}

	s.Empty(suggestion.Text)
	s.Empty(suggestion.Description)
	s.Equal(PathTypeProject, suggestion.Type) // Zero value should be PathTypeProject (iota = 0)
	s.Empty(suggestion.ProjectName)
	s.Empty(suggestion.BranchName)
}

type SuggestionOptionTestSuite struct {
	suite.Suite
}

func TestSuggestionOptionSuite(t *testing.T) {
	suite.Run(t, new(SuggestionOptionTestSuite))
}

func (s *SuggestionOptionTestSuite) TestWithExistingOnly() {
	tests := []struct {
		name     string
		option   SuggestionOption
		expected string
	}{
		{
			name:     "WithExistingOnly option",
			option:   WithExistingOnly(),
			expected: "WithExistingOnly",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.option == nil {
				s.T().Fatal("option should not be nil")
			}
			s.NotNil(tt.option, "option should not be nil")
		})
	}
}
