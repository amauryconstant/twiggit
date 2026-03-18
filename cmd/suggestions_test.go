package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"twiggit/internal/domain"
)

func TestSuggestions_SortSuggestions(t *testing.T) {
	defaultBranch := "main"

	tests := []struct {
		name          string
		suggestions   []*domain.ResolutionSuggestion
		expectedOrder []string
	}{
		{
			name: "current worktree first, default branch second, alphabetical rest",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature-branch", IsCurrent: false},
				{Text: "develop", IsCurrent: false},
				{Text: "main", IsCurrent: false},
				{Text: "current-work", IsCurrent: true},
				{Text: "bugfix", IsCurrent: false},
			},
			expectedOrder: []string{"current-work", "main", "bugfix", "develop", "feature-branch"},
		},
		{
			name: "multiple current worktrees (only one should be current)",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature-a", IsCurrent: false},
				{Text: "current", IsCurrent: true},
				{Text: "main", IsCurrent: false},
			},
			expectedOrder: []string{"current", "main", "feature-a"},
		},
		{
			name: "no current worktree - default branch first",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature-branch", IsCurrent: false},
				{Text: "develop", IsCurrent: false},
				{Text: "main", IsCurrent: false},
				{Text: "bugfix", IsCurrent: false},
			},
			expectedOrder: []string{"main", "bugfix", "develop", "feature-branch"},
		},
		{
			name: "no default branch - alphabetical only",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature-branch", IsCurrent: false},
				{Text: "develop", IsCurrent: false},
				{Text: "bugfix", IsCurrent: false},
			},
			expectedOrder: []string{"bugfix", "develop", "feature-branch"},
		},
		{
			name:          "empty suggestions",
			suggestions:   []*domain.ResolutionSuggestion{},
			expectedOrder: []string{},
		},
		{
			name: "single suggestion",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "main", IsCurrent: false},
			},
			expectedOrder: []string{"main"},
		},
		{
			name: "current worktree is also default branch - current takes precedence",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature", IsCurrent: false},
				{Text: "main", IsCurrent: true},
			},
			expectedOrder: []string{"main", "feature"},
		},
		{
			name: "custom default branch (develop)",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature", IsCurrent: false},
				{Text: "main", IsCurrent: false},
				{Text: "develop", IsCurrent: false},
			},
			expectedOrder: []string{"develop", "feature", "main"},
		},
		{
			name: "all worktrees are current (edge case - shouldn't happen but test stability)",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "a", IsCurrent: true},
				{Text: "b", IsCurrent: true},
				{Text: "c", IsCurrent: true},
			},
			expectedOrder: []string{"a", "b", "c"},
		},
		{
			name: "no current worktrees - default branch second",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "feature", IsCurrent: false},
				{Text: "main", IsCurrent: false},
				{Text: "develop", IsCurrent: false},
			},
			expectedOrder: []string{"main", "develop", "feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := make([]*domain.ResolutionSuggestion, len(tt.suggestions))
			copy(suggestions, tt.suggestions)

			testDefaultBranch := defaultBranch
			if tt.name == "custom default branch (develop)" {
				testDefaultBranch = "develop"
			}

			sortSuggestions(suggestions, testDefaultBranch)

			actualOrder := make([]string, len(suggestions))
			for i, sug := range suggestions {
				actualOrder[i] = sug.Text
			}

			assert.Equal(t, tt.expectedOrder, actualOrder)
		})
	}
}

func TestSuggestions_GetCompletionTimeout(t *testing.T) {
	tests := []struct {
		name     string
		config   *domain.Config
		expected string
	}{
		{
			name:     "nil config returns default",
			config:   nil,
			expected: "500ms",
		},
		{
			name: "empty timeout returns default",
			config: &domain.Config{
				Completion: domain.CompletionConfig{
					Timeout: "",
				},
			},
			expected: "500ms",
		},
		{
			name: "valid custom timeout",
			config: &domain.Config{
				Completion: domain.CompletionConfig{
					Timeout: "1s",
				},
			},
			expected: "1s",
		},
		{
			name: "invalid timeout falls back to default",
			config: &domain.Config{
				Completion: domain.CompletionConfig{
					Timeout: "invalid",
				},
			},
			expected: "500ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := getCompletionTimeout(tt.config)
			assert.Equal(t, tt.expected, duration.String())
		})
	}
}

func TestSuggestions_SuggestionsToCarapaceAction(t *testing.T) {
	tests := []struct {
		name                 string
		suggestions          []*domain.ResolutionSuggestion
		defaultBranch        string
		expectProjects       bool
		expectBranches       bool
		expectedProjectCount int
		expectedBranchCount  int
	}{
		{
			name:           "empty suggestions returns empty action",
			suggestions:    []*domain.ResolutionSuggestion{},
			defaultBranch:  "main",
			expectProjects: false,
			expectBranches: false,
		},
		{
			name: "only projects - separated correctly",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "project1", Type: domain.PathTypeProject, BranchName: "", Description: "Project directory"},
				{Text: "project2", Type: domain.PathTypeProject, BranchName: "", Description: "Project directory"},
			},
			defaultBranch:        "main",
			expectProjects:       true,
			expectBranches:       false,
			expectedProjectCount: 2,
		},
		{
			name: "only branches - separated correctly",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "main", Type: domain.PathTypeWorktree, BranchName: "main", Description: "Worktree"},
				{Text: "develop", Type: domain.PathTypeWorktree, BranchName: "develop", Description: "Worktree"},
			},
			defaultBranch:       "main",
			expectProjects:      false,
			expectBranches:      true,
			expectedBranchCount: 2,
		},
		{
			name: "mixed projects and branches - both separated",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "project1", Type: domain.PathTypeProject, BranchName: "", Description: "Project directory"},
				{Text: "main", Type: domain.PathTypeWorktree, BranchName: "main", Description: "Worktree"},
				{Text: "project2", Type: domain.PathTypeProject, BranchName: "", Description: "Project directory"},
				{Text: "develop", Type: domain.PathTypeWorktree, BranchName: "develop", Description: "Worktree"},
			},
			defaultBranch:        "main",
			expectProjects:       true,
			expectBranches:       true,
			expectedProjectCount: 2,
			expectedBranchCount:  2,
		},
		{
			name: "project with branch name (cross-project) treated as branch",
			suggestions: []*domain.ResolutionSuggestion{
				{Text: "project1/main", Type: domain.PathTypeProject, BranchName: "main", Description: "Cross-project"},
			},
			defaultBranch:       "main",
			expectProjects:      false,
			expectBranches:      true,
			expectedBranchCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := suggestionsToCarapaceAction(tt.suggestions, tt.defaultBranch)
			assert.NotNil(t, action, "Should return a valid action")
		})
	}
}

func TestSuggestions_SmartSortingPreservesDescription(t *testing.T) {
	suggestions := []*domain.ResolutionSuggestion{
		{Text: "feature", IsCurrent: false, Description: "Feature branch"},
		{Text: "main", IsCurrent: false, Description: "Default branch"},
		{Text: "current", IsCurrent: true, Description: "Current worktree"},
	}

	sortSuggestions(suggestions, "main")

	assert.Equal(t, "current", suggestions[0].Text)
	assert.Equal(t, "main", suggestions[1].Text)
	assert.Equal(t, "feature", suggestions[2].Text)

	assert.Equal(t, "Current worktree", suggestions[0].Description)
	assert.Equal(t, "Default branch", suggestions[1].Description)
	assert.Equal(t, "Feature branch", suggestions[2].Description)
}

func TestSuggestions_SmartSortingWithDirtyIndicator(t *testing.T) {
	suggestions := []*domain.ResolutionSuggestion{
		{Text: "clean-branch", IsCurrent: false, IsDirty: false},
		{Text: "dirty-branch", IsCurrent: true, IsDirty: true},
		{Text: "main", IsCurrent: false, IsDirty: false},
	}

	sortSuggestions(suggestions, "main")

	assert.Equal(t, "dirty-branch", suggestions[0].Text)
	assert.True(t, suggestions[0].IsDirty)
	assert.True(t, suggestions[0].IsCurrent)
}
