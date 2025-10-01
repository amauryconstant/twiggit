package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestContextResolver_ResolveIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		context        *domain.Context
		identifier     string
		expectedType   domain.PathType
		expectedProj   string
		expectedBranch string
		expectedPath   string
		expectError    bool
	}{
		{
			name: "project context - main to project root",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:   "main",
			expectedType: domain.PathTypeProject,
			expectedProj: "test-project",
			expectedPath: "/home/user/Projects/test-project",
		},
		{
			name: "project context - branch to worktree",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:     "feature-branch",
			expectedType:   domain.PathTypeWorktree,
			expectedProj:   "test-project",
			expectedBranch: "feature-branch",
			expectedPath:   "/home/user/Worktrees/test-project/feature-branch",
		},
		{
			name: "worktree context - main to project root",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "test-project",
				BranchName:  "current-branch",
				Path:        "/home/user/Worktrees/test-project/current-branch",
			},
			identifier:   "main",
			expectedType: domain.PathTypeProject,
			expectedProj: "test-project",
			expectedPath: "/home/user/Projects/test-project",
		},
		{
			name: "outside git context - project to project directory",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:   "test-project",
			expectedType: domain.PathTypeProject,
			expectedProj: "test-project",
			expectedPath: "/home/user/Projects/test-project",
		},
		{
			name: "cross-project reference",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:     "other-project/feature-branch",
			expectedType:   domain.PathTypeWorktree,
			expectedProj:   "other-project",
			expectedBranch: "feature-branch",
			expectedPath:   "/home/user/Worktrees/other-project/feature-branch",
		},
		{
			name: "empty identifier",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:  "",
			expectError: true,
		},
		{
			name: "invalid cross-project reference",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:   "project/branch/extra",
			expectedType: domain.PathTypeInvalid,
			expectedPath: "",
		},
		{
			name: "unknown context",
			context: &domain.Context{
				Type: domain.ContextUnknown,
				Path: "/home/user",
			},
			identifier:   "test",
			expectedType: domain.PathTypeInvalid,
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &domain.Config{
				ProjectsDirectory:  "/home/user/Projects",
				WorktreesDirectory: "/home/user/Worktrees",
			}

			resolver := NewContextResolver(config)
			result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, result.Type)
			assert.Equal(t, tt.expectedProj, result.ProjectName)
			if tt.expectedBranch != "" {
				assert.Equal(t, tt.expectedBranch, result.BranchName)
			}
			assert.Equal(t, tt.expectedPath, result.ResolvedPath)
			assert.NotEmpty(t, result.Explanation)
		})
	}
}

func TestContextResolver_GetResolutionSuggestions(t *testing.T) {
	tests := []struct {
		name          string
		context       *domain.Context
		partial       string
		expectedCount int
		expectedTexts []string
	}{
		{
			name: "project context - partial 'm'",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			partial:       "m",
			expectedCount: 1,
			expectedTexts: []string{"main"},
		},
		{
			name: "worktree context - partial 'main'",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "test-project",
				BranchName:  "current-branch",
				Path:        "/home/user/Worktrees/test-project/current-branch",
			},
			partial:       "main",
			expectedCount: 1,
			expectedTexts: []string{"main"},
		},
		{
			name: "outside git context - partial 'test'",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			partial:       "test",
			expectedCount: 0,
			expectedTexts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &domain.Config{
				ProjectsDirectory:  "/home/user/Projects",
				WorktreesDirectory: "/home/user/Worktrees",
			}

			resolver := NewContextResolver(config)
			suggestions, err := resolver.GetResolutionSuggestions(tt.context, tt.partial)

			require.NoError(t, err)
			assert.Len(t, suggestions, tt.expectedCount)

			if len(tt.expectedTexts) > 0 {
				suggestionTexts := make([]string, len(suggestions))
				for i, suggestion := range suggestions {
					suggestionTexts[i] = suggestion.Text
				}
				assert.Equal(t, tt.expectedTexts, suggestionTexts)
			}
		})
	}
}

func TestContextResolver_WorktreeContextResolution(t *testing.T) {
	config := &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
	}

	resolver := NewContextResolver(config)

	// Test worktree context resolving to different branch
	ctx := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "my-project",
		BranchName:  "current-branch",
		Path:        "/home/user/Worktrees/my-project/current-branch",
	}

	result, err := resolver.ResolveIdentifier(ctx, "other-branch")
	require.NoError(t, err)

	assert.Equal(t, domain.PathTypeWorktree, result.Type)
	assert.Equal(t, "my-project", result.ProjectName)
	assert.Equal(t, "other-branch", result.BranchName)
	assert.Equal(t, "/home/user/Worktrees/my-project/other-branch", result.ResolvedPath)
	assert.Contains(t, result.Explanation, "my-project")
}

func TestContextResolver_CrossProjectReference(t *testing.T) {
	config := &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
	}

	resolver := NewContextResolver(config)

	// Test from project context
	ctx := &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "current-project",
		Path:        "/home/user/Projects/current-project",
	}

	result, err := resolver.ResolveIdentifier(ctx, "other-project/feature-branch")
	require.NoError(t, err)

	assert.Equal(t, domain.PathTypeWorktree, result.Type)
	assert.Equal(t, "other-project", result.ProjectName)
	assert.Equal(t, "feature-branch", result.BranchName)
	assert.Equal(t, "/home/user/Worktrees/other-project/feature-branch", result.ResolvedPath)
	assert.Contains(t, result.Explanation, "other-project")
}
