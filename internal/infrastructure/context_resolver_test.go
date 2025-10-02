package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
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

			resolver := NewContextResolver(config, nil)
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

			resolver := NewContextResolver(config, nil)
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

	resolver := NewContextResolver(config, nil)

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

	resolver := NewContextResolver(config, nil)

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

// Pure function tests for extracted functions

func TestParseCrossProjectReference(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		expectedProj   string
		expectedBranch string
		expectedValid  bool
	}{
		{
			name:           "valid cross-project reference",
			identifier:     "project/branch",
			expectedProj:   "project",
			expectedBranch: "branch",
			expectedValid:  true,
		},
		{
			name:           "valid reference with complex names",
			identifier:     "my-project/feature-branch-123",
			expectedProj:   "my-project",
			expectedBranch: "feature-branch-123",
			expectedValid:  true,
		},
		{
			name:           "invalid - missing slash",
			identifier:     "projectbranch",
			expectedProj:   "",
			expectedBranch: "",
			expectedValid:  false,
		},
		{
			name:           "invalid - too many slashes",
			identifier:     "project/branch/extra",
			expectedProj:   "",
			expectedBranch: "",
			expectedValid:  false,
		},
		{
			name:           "invalid - empty project",
			identifier:     "/branch",
			expectedProj:   "",
			expectedBranch: "",
			expectedValid:  false,
		},
		{
			name:           "invalid - empty branch",
			identifier:     "project/",
			expectedProj:   "",
			expectedBranch: "",
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, branch, valid := parseCrossProjectReference(tt.identifier)
			assert.Equal(t, tt.expectedProj, project)
			assert.Equal(t, tt.expectedBranch, branch)
			assert.Equal(t, tt.expectedValid, valid)
		})
	}
}

func TestBuildWorktreePath(t *testing.T) {
	tests := []struct {
		name         string
		worktreesDir string
		project      string
		branch       string
		expectedPath string
	}{
		{
			name:         "basic worktree path",
			worktreesDir: "/home/user/Worktrees",
			project:      "my-project",
			branch:       "feature-branch",
			expectedPath: "/home/user/Worktrees/my-project/feature-branch",
		},
		{
			name:         "complex project and branch names",
			worktreesDir: "/opt/worktrees",
			project:      "complex-project-name",
			branch:       "feature/branch-123",
			expectedPath: "/opt/worktrees/complex-project-name/feature/branch-123",
		},
		{
			name:         "relative worktrees directory",
			worktreesDir: "./worktrees",
			project:      "test-project",
			branch:       "main",
			expectedPath: "worktrees/test-project/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := buildWorktreePath(tt.worktreesDir, tt.project, tt.branch)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestBuildProjectPath(t *testing.T) {
	tests := []struct {
		name         string
		projectsDir  string
		project      string
		expectedPath string
	}{
		{
			name:         "basic project path",
			projectsDir:  "/home/user/Projects",
			project:      "my-project",
			expectedPath: "/home/user/Projects/my-project",
		},
		{
			name:         "complex project name",
			projectsDir:  "/opt/projects",
			project:      "complex-project-name",
			expectedPath: "/opt/projects/complex-project-name",
		},
		{
			name:         "relative projects directory",
			projectsDir:  "./projects",
			project:      "test-project",
			expectedPath: "projects/test-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := buildProjectPath(tt.projectsDir, tt.project)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestFilterSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		suggestions    []string
		partial        string
		expectedResult []string
	}{
		{
			name:           "empty partial returns all suggestions",
			suggestions:    []string{"main", "feature-branch", "bugfix"},
			partial:        "",
			expectedResult: []string{"main", "feature-branch", "bugfix"},
		},
		{
			name:           "partial matches some suggestions",
			suggestions:    []string{"main", "feature-branch", "bugfix", "maintenance"},
			partial:        "fe",
			expectedResult: []string{"feature-branch"},
		},
		{
			name:           "partial matches multiple suggestions",
			suggestions:    []string{"main", "feature-branch", "feature-api", "bugfix"},
			partial:        "feature",
			expectedResult: []string{"feature-branch", "feature-api"},
		},
		{
			name:           "no matches returns empty",
			suggestions:    []string{"main", "feature-branch", "bugfix"},
			partial:        "xyz",
			expectedResult: []string{},
		},
		{
			name:           "case sensitive matching",
			suggestions:    []string{"Main", "feature-branch", "BUGFIX"},
			partial:        "m",
			expectedResult: []string{},
		},
		{
			name:           "empty suggestions list",
			suggestions:    []string{},
			partial:        "test",
			expectedResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterSuggestions(tt.suggestions, tt.partial)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestContextResolver_getOutsideGitContextSuggestions(t *testing.T) {
	tests := []struct {
		name          string
		projectsDir   string
		setupProjects func(dir string) error
		partial       string
		expectedCount int
		expectedTexts []string
	}{
		{
			name:          "no projects directory configured",
			projectsDir:   "",
			partial:       "test",
			expectedCount: 0,
			expectedTexts: []string{},
		},
		{
			name:          "projects directory does not exist",
			projectsDir:   "/nonexistent/directory",
			partial:       "test",
			expectedCount: 0,
			expectedTexts: []string{},
		},
		{
			name:        "projects directory with git repositories",
			projectsDir: t.TempDir(),
			setupProjects: func(dir string) error {
				// Create project directories
				if err := os.MkdirAll(filepath.Join(dir, "project1"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "project2"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "not-a-repo"), 0755); err != nil {
					return err
				}

				// Initialize git repositories in project1 and project2
				setupTestRepo(t, filepath.Join(dir, "project1"))
				setupTestRepo(t, filepath.Join(dir, "project2"))
				return nil
			},
			partial:       "proj",
			expectedCount: 2,
			expectedTexts: []string{"project1", "project2"},
		},
		{
			name:        "partial matching with 'test'",
			projectsDir: t.TempDir(),
			setupProjects: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "test-project"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "other-project"), 0755); err != nil {
					return err
				}

				setupTestRepo(t, filepath.Join(dir, "test-project"))
				setupTestRepo(t, filepath.Join(dir, "other-project"))
				return nil
			},
			partial:       "test",
			expectedCount: 1,
			expectedTexts: []string{"test-project"},
		},
		{
			name:        "empty partial matches all",
			projectsDir: t.TempDir(),
			setupProjects: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "project1"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "project2"), 0755); err != nil {
					return err
				}

				setupTestRepo(t, filepath.Join(dir, "project1"))
				setupTestRepo(t, filepath.Join(dir, "project2"))
				return nil
			},
			partial:       "",
			expectedCount: 2,
			expectedTexts: []string{"project1", "project2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup projects directory
			if tt.setupProjects != nil {
				if err := tt.setupProjects(tt.projectsDir); err != nil {
					t.Fatalf("Failed to setup test projects: %v", err)
				}
			}

			config := &domain.Config{
				ProjectsDirectory:  tt.projectsDir,
				WorktreesDirectory: "/home/user/Worktrees",
			}

			// Create mock git service
			mockGitService := mocks.NewMockGitService()
			mockGitService.ValidateRepositoryFunc = func(path string) error {
				return nil // Assume all directories are valid repos for testing
			}

			resolver := NewContextResolver(config, mockGitService)
			suggestions := resolver.(*contextResolver).getOutsideGitContextSuggestions(tt.partial)

			assert.Len(t, suggestions, tt.expectedCount, "Expected %d suggestions", tt.expectedCount)

			if len(tt.expectedTexts) > 0 {
				suggestionTexts := make([]string, len(suggestions))
				for i, suggestion := range suggestions {
					suggestionTexts[i] = suggestion.Text
				}
				assert.Equal(t, tt.expectedTexts, suggestionTexts, "Suggestion texts don't match")
			}

			// Verify suggestion properties
			for _, suggestion := range suggestions {
				assert.NotEmpty(t, suggestion.Text, "Suggestion text should not be empty")
				assert.Equal(t, domain.PathTypeProject, suggestion.Type, "Suggestion type should be project")
				assert.Equal(t, suggestion.Text, suggestion.ProjectName, "Project name should match text")
				assert.Equal(t, "Project directory", suggestion.Description, "Description should be consistent")
			}
		})
	}
}
