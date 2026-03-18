package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ContextType String() tests
func TestContextType_String(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			result := tc.context.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// PathType String() tests
func TestPathType_String(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			result := tc.pathType.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Data structure tests
func TestContext_Structure(t *testing.T) {
	context := &Context{
		Type:        ContextProject,
		ProjectName: "test-project",
		BranchName:  "main",
		Path:        "/test/path",
		Explanation: "Test explanation",
	}

	assert.Equal(t, ContextProject, context.Type)
	assert.Equal(t, "test-project", context.ProjectName)
	assert.Equal(t, "main", context.BranchName)
	assert.Equal(t, "/test/path", context.Path)
	assert.Equal(t, "Test explanation", context.Explanation)
}

func TestResolutionResult_Structure(t *testing.T) {
	result := &ResolutionResult{
		ResolvedPath: "/resolved/path",
		Type:         PathTypeWorktree,
		ProjectName:  "test-project",
		BranchName:   "feature-branch",
		Explanation:  "Resolved successfully",
	}

	assert.Equal(t, "/resolved/path", result.ResolvedPath)
	assert.Equal(t, PathTypeWorktree, result.Type)
	assert.Equal(t, "test-project", result.ProjectName)
	assert.Equal(t, "feature-branch", result.BranchName)
	assert.Equal(t, "Resolved successfully", result.Explanation)
}

func TestResolutionSuggestion_Structure(t *testing.T) {
	suggestion := &ResolutionSuggestion{
		Text:        "feature-branch",
		Description: "Feature branch suggestion",
		Type:        PathTypeWorktree,
		ProjectName: "test-project",
		BranchName:  "feature-branch",
	}

	assert.Equal(t, "feature-branch", suggestion.Text)
	assert.Equal(t, "Feature branch suggestion", suggestion.Description)
	assert.Equal(t, PathTypeWorktree, suggestion.Type)
	assert.Equal(t, "test-project", suggestion.ProjectName)
	assert.Equal(t, "feature-branch", suggestion.BranchName)
}

// Edge case tests for zero values
func TestContext_ZeroValues(t *testing.T) {
	context := &Context{}

	assert.Equal(t, ContextUnknown, context.Type)
	assert.Empty(t, context.ProjectName)
	assert.Empty(t, context.BranchName)
	assert.Empty(t, context.Path)
	assert.Empty(t, context.Explanation)
}

func TestResolutionResult_ZeroValues(t *testing.T) {
	result := &ResolutionResult{}

	assert.Empty(t, result.ResolvedPath)
	assert.Equal(t, PathTypeProject, result.Type) // Zero value should be PathTypeProject (iota = 0)
	assert.Empty(t, result.ProjectName)
	assert.Empty(t, result.BranchName)
	assert.Empty(t, result.Explanation)
}

func TestResolutionSuggestion_ZeroValues(t *testing.T) {
	suggestion := &ResolutionSuggestion{}

	assert.Empty(t, suggestion.Text)
	assert.Empty(t, suggestion.Description)
	assert.Equal(t, PathTypeProject, suggestion.Type) // Zero value should be PathTypeProject (iota = 0)
	assert.Empty(t, suggestion.ProjectName)
	assert.Empty(t, suggestion.BranchName)
}

func TestWithExistingOnly(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			if tt.option == nil {
				t.Fatal("option should not be nil")
			}
			assert.NotNil(t, tt.option, "option should not be nil")
		})
	}
}
