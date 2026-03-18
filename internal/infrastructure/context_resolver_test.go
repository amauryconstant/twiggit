package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func setupContextResolverTest(t *testing.T) *domain.Config {
	t.Helper()
	return &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
	}
}

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
			config := setupContextResolverTest(t)
			resolver := NewContextResolver(config, nil)
			result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)

			if tt.expectError {
				require.Error(t, err)
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
			config := setupContextResolverTest(t)
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
	config := setupContextResolverTest(t)
	resolver := NewContextResolver(config, nil)

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
	config := setupContextResolverTest(t)
	resolver := NewContextResolver(config, nil)

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

func TestContextResolver_ParseCrossProjectReference(t *testing.T) {
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

func TestContextResolver_BuildWorktreePath(t *testing.T) {
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

func TestContextResolver_BuildProjectPath(t *testing.T) {
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

func TestContextResolver_FilterSuggestions(t *testing.T) {
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

func TestContextResolver_ContainsPathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"literal dots", "..", true},
		{"path with dots", "../etc/passwd", true},
		{"lowercase URL-encoded", "%2e%2e", true},
		{"uppercase URL-encoded", "%2E%2E", true},
		{"mixed case URL-encoded", "%2e%2E", true},
		{"mixed case URL-encoded v2", "%2E%2e", true},
		{"double URL-encoded", "%252e%252e", true},
		{"double URL-encoded uppercase", "%252E%252E", true},
		{"URL-encoded with slash", "%2e%2e%2f", true},
		{"valid branch name", "feature-branch", false},
		{"valid project name", "my-project", false},
		{"valid name with slash", "feature/branch", false},
		{"valid name with dots elsewhere", "v1.2.3", false},
		{"empty string", "", false},
		{"single dot", ".", false},
		{"URL-encoded single char", "%2e", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsPathTraversal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextResolver_PathTraversalProtection(t *testing.T) {
	tests := []struct {
		name          string
		context       *domain.Context
		identifier    string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name: "path traversal attack - worktree resolution from project context",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:    "../../../etc/passwd",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in branch name should be rejected",
		},
		{
			name: "path traversal attack - worktree resolution from worktree context",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "test-project",
				BranchName:  "current-branch",
				Path:        "/home/user/Worktrees/test-project/current-branch",
			},
			identifier:    "../../etc/passwd",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in branch name from worktree context should be rejected",
		},
		{
			name: "path traversal attack - project resolution from worktree context",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "../../../etc",
				BranchName:  "current-branch",
				Path:        "/home/user/Worktrees/test-project/current-branch",
			},
			identifier:    "main",
			expectError:   true,
			errorContains: "project name contains path traversal sequences",
			description:   "Path traversal via project name in context should be rejected",
		},
		{
			name: "path traversal attack - project resolution from outside git context",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:    "../../../etc/passwd",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in project name from outside git context should be rejected",
		},
		{
			name: "path traversal attack - cross-project reference",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:    "../../../etc/passwd/../../passwd",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in cross-project reference should be rejected",
		},
		{
			name: "absolute path escape attempt",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:  "/etc/passwd",
			expectError: false,
			description: "Absolute path is treated as invalid format, not an error",
		},
		{
			name: "valid path with special characters should work",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:  "feature/branch-123",
			expectError: false,
			description: "Valid branch names with slashes should be accepted",
		},
		{
			name: "cross-project reference with traversal in project",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:    "../../etc/passwd/branch",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in project part of cross-project reference should be rejected",
		},
		{
			name: "cross-project reference with traversal in branch",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:    "other-project/../../../etc/passwd",
			expectError:   true,
			errorContains: "identifier contains path traversal sequences",
			description:   "Path traversal in branch part of cross-project reference should be rejected",
		},
		{
			name: "normal branch resolution should work",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:  "normal-branch",
			expectError: false,
			description: "Normal branch resolution should succeed",
		},
		{
			name: "normal project resolution should work",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:  "test-project",
			expectError: false,
			description: "Normal project resolution should succeed",
		},
		{
			name: "normal cross-project reference should work",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:  "other-project/feature-branch",
			expectError: false,
			description: "Normal cross-project reference should succeed",
		},
		{
			name: "URL-encoded path traversal lowercase",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:    "%2e%2e",
			expectError:   true,
			errorContains: "path traversal sequences",
			description:   "Lowercase URL-encoded .. should be rejected",
		},
		{
			name: "URL-encoded path traversal uppercase",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:    "%2E%2E",
			expectError:   true,
			errorContains: "path traversal sequences",
			description:   "Uppercase URL-encoded .. should be rejected",
		},
		{
			name: "URL-encoded path traversal mixed case",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:    "%2e%2E",
			expectError:   true,
			errorContains: "path traversal sequences",
			description:   "Mixed case URL-encoded .. should be rejected",
		},
		{
			name: "double URL-encoded path traversal",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
				Path:        "/home/user/Projects/test-project",
			},
			identifier:    "%252e%252e",
			expectError:   true,
			errorContains: "path traversal sequences",
			description:   "Double URL-encoded .. should be rejected",
		},
		{
			name: "URL-encoded path traversal in project name",
			context: &domain.Context{
				Type:        domain.ContextWorktree,
				ProjectName: "%2e%2e",
				BranchName:  "current-branch",
				Path:        "/home/user/Worktrees/test-project/current-branch",
			},
			identifier:    "main",
			expectError:   true,
			errorContains: "project name contains path traversal sequences",
			description:   "URL-encoded path traversal in project name should be rejected",
		},
		{
			name: "URL-encoded path traversal in cross-project reference",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
				Path: "/home/user",
			},
			identifier:    "project/%2e%2e%2f%2e%2e",
			expectError:   true,
			errorContains: "path traversal sequences",
			description:   "URL-encoded traversal in cross-project reference should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := setupContextResolverTest(t)
			resolver := NewContextResolver(config, nil)
			result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)

			if tt.expectError {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.errorContains, tt.description)
				assert.Nil(t, result, tt.description)
			} else {
				require.NoError(t, err, tt.description)
				assert.NotNil(t, result, tt.description)
			}
		})
	}
}

func setupMinimalTestRepo(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(filepath.Join(path, ".git"), 0755)
	require.NoError(t, err)

	gitDir := filepath.Join(path, ".git", "refs", "heads")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(path, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(gitDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0644)
	require.NoError(t, err)
}

func TestContextResolver_DescriptionFormats(t *testing.T) {
	t.Run("main suggestion has project root directory description", func(t *testing.T) {
		config := setupContextResolverTest(t)
		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "test-project",
			Path:        "/home/user/Projects/test-project",
		}

		resolver := NewContextResolver(config, nil)
		suggestions, err := resolver.GetResolutionSuggestions(ctx, "main")

		require.NoError(t, err)
		require.Len(t, suggestions, 1)
		assert.Equal(t, "main", suggestions[0].Text)
		assert.Equal(t, "Project root directory", suggestions[0].Description)
	})

	t.Run("worktree suggestion has worktree for branch description", func(t *testing.T) {
		config := setupContextResolverTest(t)
		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "test-project",
			Path:        "/home/user/Projects/test-project",
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/home/user/Projects/test-project").
			Return([]domain.WorktreeInfo{
				{Branch: "feature-1", Path: "/home/user/Worktrees/test-project/feature-1"},
			}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, "/home/user/Projects/test-project").
			Return([]domain.BranchInfo{}, nil)

		resolver := NewContextResolver(config, mockGitService)
		suggestions, err := resolver.GetResolutionSuggestions(ctx, "feature-1")

		require.NoError(t, err)
		require.Len(t, suggestions, 1)
		assert.Equal(t, "feature-1", suggestions[0].Text)
		assert.Equal(t, "Worktree for branch feature-1", suggestions[0].Description)
	})

	t.Run("branch without worktree has create worktree description", func(t *testing.T) {
		config := setupContextResolverTest(t)
		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "test-project",
			Path:        "/home/user/Projects/test-project",
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/home/user/Projects/test-project").
			Return([]domain.WorktreeInfo{}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, "/home/user/Projects/test-project").
			Return([]domain.BranchInfo{{Name: "develop"}}, nil)

		resolver := NewContextResolver(config, mockGitService)
		suggestions, err := resolver.GetResolutionSuggestions(ctx, "develop")

		require.NoError(t, err)
		require.Len(t, suggestions, 1)
		assert.Equal(t, "develop", suggestions[0].Text)
		assert.Equal(t, "Branch develop (create worktree)", suggestions[0].Description)
	})

	t.Run("project suggestion has project directory description", func(t *testing.T) {
		tempDir := t.TempDir()
		projectsDir := tempDir

		project1Path := filepath.Join(projectsDir, "project1")
		require.NoError(t, os.MkdirAll(filepath.Join(project1Path, ".git"), 0755))
		setupMinimalTestRepo(t, project1Path)

		config := &domain.Config{
			ProjectsDirectory:  projectsDir,
			WorktreesDirectory: filepath.Join(tempDir, "worktrees"),
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockGoGitClient.On("ValidateRepository", project1Path).Return(nil)

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type: domain.ContextOutsideGit,
			Path: tempDir,
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "proj")

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(suggestions), 1)

		var project1 *domain.ResolutionSuggestion
		for _, sug := range suggestions {
			if sug.Text == "project1" {
				project1 = sug
				break
			}
		}
		require.NotNil(t, project1)
		assert.Equal(t, "Project directory", project1.Description)
	})
}

func TestContextResolver_WithExistingOnlyFilter(t *testing.T) {
	_, err := os.Stat("/tmp/nonexistent")
	require.True(t, os.IsNotExist(err))
}

func TestContextResolver_FuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected bool
	}{
		{"empty pattern matches anything", "", "anything", true},
		{"empty pattern matches empty", "", "", true},
		{"exact match", "main", "main", true},
		{"prefix match", "ma", "main", true},
		{"subsequence match", "mn", "main", true},
		{"non-matching pattern", "xyz", "main", false},
		{"case insensitive - lowercase pattern", "main", "MAIN", true},
		{"case insensitive - uppercase pattern", "MAIN", "main", true},
		{"case insensitive - mixed case", "MaIn", "mAiN", true},
		{"case insensitive subsequence", "mn", "MAIN", true},
		{"fuzzy - f1 matches feature-1", "f1", "feature-1", true},
		{"fuzzy - fb matches feature-branch", "fb", "feature-branch", true},
		{"fuzzy - ftb matches feature-test-branch", "ftb", "feature-test-branch", true},
		{"fuzzy - fbn does not match feature", "fbn", "feature", false},
		{"fuzzy - dv matches develop", "dv", "develop", true},
		{"fuzzy - dvp matches develop", "dvp", "develop", true},
		{"pattern longer than text", "longpattern", "short", false},
		{"single char match", "m", "main", true},
		{"single char no match", "x", "main", false},
		{"unicode characters", "f", "fëätürë", true},
		{"bugfix pattern", "bf", "bugfix-123", true},
		{"feature pattern", "feat", "feature/new-thing", true},
		{"hotfix pattern", "hf", "hotfix-urgent", true},
		{"release pattern", "rel", "release/v1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fuzzyMatch(tt.pattern, tt.text)
			assert.Equal(t, tt.expected, result, "fuzzyMatch(%q, %q)", tt.pattern, tt.text)
		})
	}
}

func TestContextResolver_MatchesExclusionPatterns(t *testing.T) {
	tests := []struct {
		name        string
		nameToCheck string
		patterns    []string
		expected    bool
	}{
		{"no patterns", "main", []string{}, false},
		{"exact match", "main", []string{"main"}, true},
		{"no match", "develop", []string{"main"}, false},
		{"glob wildcard prefix", "dependabot-123", []string{"dependabot-*"}, true},
		{"glob wildcard suffix", "test-branch", []string{"*-branch"}, true},
		{"glob multiple patterns - first matches", "renovate-456", []string{"renovate-*", "dependabot-*"}, true},
		{"glob multiple patterns - second matches", "dependabot-789", []string{"renovate-*", "dependabot-*"}, true},
		{"glob multiple patterns - none match", "feature-branch", []string{"renovate-*", "dependabot-*"}, false},
		{"glob question mark", "test1", []string{"test?"}, true},
		{"glob question mark - no match", "test12", []string{"test?"}, false},
		{"glob character class", "v1.0", []string{"v[0-9].*"}, true},
		{"archive pattern", "archive-old-project", []string{"archive/*"}, false},
		{"empty name", "", []string{"*"}, true},
		{"pattern with invalid glob syntax", "branch", []string{"["}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesExclusionPatterns(tt.nameToCheck, tt.patterns)
			assert.Equal(t, tt.expected, result, "matchesExclusionPatterns(%q, %v)", tt.nameToCheck, tt.patterns)
		})
	}
}

func TestContextResolver_ProjectSuggestionsFromProjectContext(t *testing.T) {
	t.Run("includes other projects when in project context", func(t *testing.T) {
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		require.NoError(t, os.MkdirAll(projectsDir, 0755))

		project1Path := filepath.Join(projectsDir, "project1")
		project2Path := filepath.Join(projectsDir, "project2")
		require.NoError(t, os.MkdirAll(filepath.Join(project1Path, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(project2Path, ".git"), 0755))
		setupMinimalTestRepo(t, project1Path)
		setupMinimalTestRepo(t, project2Path)

		config := &domain.Config{
			ProjectsDirectory:  projectsDir,
			WorktreesDirectory: filepath.Join(tempDir, "worktrees"),
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockGoGitClient.On("ValidateRepository", project1Path).Return(nil)
		mockGitService.MockGoGitClient.On("ValidateRepository", project2Path).Return(nil)
		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, project1Path).
			Return([]domain.WorktreeInfo{}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, project1Path).
			Return([]domain.BranchInfo{}, nil)

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "project1",
			Path:        project1Path,
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "proj")
		require.NoError(t, err)

		var projectNames []string
		for _, sug := range suggestions {
			if sug.Type == domain.PathTypeProject && sug.BranchName == "" {
				projectNames = append(projectNames, sug.Text)
			}
		}
		assert.Contains(t, projectNames, "project2", "Should suggest other projects")
		assert.NotContains(t, projectNames, "project1", "Should not suggest current project")
	})

	t.Run("respects exclusion patterns for projects", func(t *testing.T) {
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		require.NoError(t, os.MkdirAll(projectsDir, 0755))

		project1Path := filepath.Join(projectsDir, "project1")
		archivePath := filepath.Join(projectsDir, "archive-old")
		require.NoError(t, os.MkdirAll(filepath.Join(project1Path, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(archivePath, ".git"), 0755))
		setupMinimalTestRepo(t, project1Path)
		setupMinimalTestRepo(t, archivePath)

		config := &domain.Config{
			ProjectsDirectory:  projectsDir,
			WorktreesDirectory: filepath.Join(tempDir, "worktrees"),
			Completion: domain.CompletionConfig{
				ExcludeProjects: []string{"archive-*"},
			},
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockGoGitClient.On("ValidateRepository", project1Path).Return(nil)
		mockGitService.MockGoGitClient.On("ValidateRepository", archivePath).Return(nil)
		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, project1Path).
			Return([]domain.WorktreeInfo{}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, project1Path).
			Return([]domain.BranchInfo{}, nil)

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "project1",
			Path:        project1Path,
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "")
		require.NoError(t, err)

		var projectNames []string
		for _, sug := range suggestions {
			if sug.Type == domain.PathTypeProject && sug.BranchName == "" {
				projectNames = append(projectNames, sug.Text)
			}
		}
		assert.NotContains(t, projectNames, "archive-old", "Should exclude projects matching pattern")
	})
}

func TestContextResolver_ProjectSuggestionsFromWorktreeContext(t *testing.T) {
	t.Run("includes other projects when in worktree context", func(t *testing.T) {
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		worktreesDir := filepath.Join(tempDir, "worktrees")
		require.NoError(t, os.MkdirAll(projectsDir, 0755))
		require.NoError(t, os.MkdirAll(worktreesDir, 0755))

		project1Path := filepath.Join(projectsDir, "project1")
		project2Path := filepath.Join(projectsDir, "project2")
		require.NoError(t, os.MkdirAll(filepath.Join(project1Path, ".git"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join(project2Path, ".git"), 0755))
		setupMinimalTestRepo(t, project1Path)
		setupMinimalTestRepo(t, project2Path)

		worktreePath := filepath.Join(worktreesDir, "project1", "feature-branch")
		require.NoError(t, os.MkdirAll(worktreePath, 0755))

		config := &domain.Config{
			ProjectsDirectory:  projectsDir,
			WorktreesDirectory: worktreesDir,
		}

		mockGitService := mocks.NewMockGitService()
		mockGitService.MockGoGitClient.On("ValidateRepository", project1Path).Return(nil)
		mockGitService.MockGoGitClient.On("ValidateRepository", project2Path).Return(nil)
		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, project1Path).
			Return([]domain.WorktreeInfo{
				{Branch: "feature-branch", Path: worktreePath},
			}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, project1Path).
			Return([]domain.BranchInfo{}, nil)

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type:        domain.ContextWorktree,
			ProjectName: "project1",
			BranchName:  "feature-branch",
			Path:        worktreePath,
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "proj")
		require.NoError(t, err)

		var projectNames []string
		for _, sug := range suggestions {
			if sug.Type == domain.PathTypeProject && sug.BranchName == "" {
				projectNames = append(projectNames, sug.Text)
			}
		}
		assert.Contains(t, projectNames, "project2", "Should suggest other projects")
		assert.NotContains(t, projectNames, "project1", "Should not suggest current project")
	})
}

func TestContextResolver_ExclusionPatternFiltering(t *testing.T) {
	t.Run("excludes branches matching patterns", func(t *testing.T) {
		config := &domain.Config{
			ProjectsDirectory:  "/home/user/Projects",
			WorktreesDirectory: "/home/user/Worktrees",
			Completion: domain.CompletionConfig{
				ExcludeBranches: []string{"dependabot/*", "renovate/*"},
			},
		}

		mockGitService := mocks.NewMockGitService()
		projectPath := "/home/user/Projects/my-project"
		worktreePath := "/home/user/Worktrees/my-project/feature-branch"

		mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, projectPath).
			Return([]domain.WorktreeInfo{
				{Branch: "feature-branch", Path: worktreePath},
				{Branch: "other-branch", Path: "/home/user/Worktrees/my-project/other-branch"},
			}, nil)
		mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, projectPath).
			Return([]domain.BranchInfo{
				{Name: "main"},
				{Name: "dependabot/npm-123"},
				{Name: "renovate/docker-456"},
			}, nil)

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "my-project",
			Path:        "/home/user/Projects/my-project",
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "")
		require.NoError(t, err)

		allNames := make([]string, 0)
		for _, sug := range suggestions {
			allNames = append(allNames, sug.Text)
		}

		assert.Contains(t, allNames, "main", "Should include main branch")
		assert.Contains(t, allNames, "feature-branch", "Should include feature-branch worktree")
		assert.Contains(t, allNames, "other-branch", "Should include other-branch worktree")
		assert.NotContains(t, allNames, "dependabot/npm-123", "Should exclude dependabot branches")
		assert.NotContains(t, allNames, "renovate/docker-456", "Should exclude renovate branches")
	})

	t.Run("excludes projects matching patterns", func(t *testing.T) {
		tempDir := t.TempDir()
		projectsDir := filepath.Join(tempDir, "projects")
		require.NoError(t, os.MkdirAll(projectsDir, 0755))

		for _, name := range []string{"active-project", "archived-old", "test-project"} {
			projectPath := filepath.Join(projectsDir, name)
			require.NoError(t, os.MkdirAll(filepath.Join(projectPath, ".git"), 0755))
			setupMinimalTestRepo(t, projectPath)
		}

		config := &domain.Config{
			ProjectsDirectory:  projectsDir,
			WorktreesDirectory: filepath.Join(tempDir, "worktrees"),
			Completion: domain.CompletionConfig{
				ExcludeProjects: []string{"archived-*"},
			},
		}

		mockGitService := mocks.NewMockGitService()
		for _, name := range []string{"active-project", "archived-old", "test-project"} {
			projectPath := filepath.Join(projectsDir, name)
			mockGitService.MockGoGitClient.On("ValidateRepository", projectPath).Return(nil)
		}

		resolver := NewContextResolver(config, mockGitService)

		ctx := &domain.Context{
			Type: domain.ContextOutsideGit,
			Path: tempDir,
		}

		suggestions, err := resolver.GetResolutionSuggestions(ctx, "")
		require.NoError(t, err)

		projectNames := make([]string, 0)
		for _, sug := range suggestions {
			projectNames = append(projectNames, sug.Text)
		}

		assert.Contains(t, projectNames, "active-project")
		assert.Contains(t, projectNames, "test-project")
		assert.NotContains(t, projectNames, "archived-old", "Should exclude archived projects")
	})
}

func TestContextResolver_FuzzyMatchingEnabled(t *testing.T) {
	config := &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
		Navigation: domain.NavigationConfig{
			FuzzyMatching: true,
		},
	}

	projectPath := "/home/user/Projects/my-project"
	worktreePath := "/home/user/Worktrees/my-project/feature-branch"

	mockGitService := mocks.NewMockGitService()
	mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, projectPath).
		Return([]domain.WorktreeInfo{
			{Branch: "feature-branch", Path: worktreePath},
			{Branch: "other-branch", Path: "/home/user/Worktrees/my-project/other-branch"},
		}, nil)
	mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, projectPath).
		Return([]domain.BranchInfo{
			{Name: "feature-123"},
			{Name: "main"},
		}, nil)

	resolver := NewContextResolver(config, mockGitService)

	ctx := &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "my-project",
		Path:        "/home/user/Projects/my-project",
	}

	suggestions, err := resolver.GetResolutionSuggestions(ctx, "f1")
	require.NoError(t, err)

	branchNames := make([]string, 0)
	for _, sug := range suggestions {
		branchNames = append(branchNames, sug.Text)
	}

	assert.Contains(t, branchNames, "feature-123", "Fuzzy match 'f1' should match 'feature-123'")
}

func TestContextResolver_WorktreeStatusFields(t *testing.T) {
	config := &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
	}

	mockGitService := mocks.NewMockGitService()
	worktreePath := "/home/user/Worktrees/my-project/feature-branch"
	projectPath := "/home/user/Projects/my-project"

	mockGitService.MockCLIClient.On("ListWorktrees", mock.Anything, projectPath).
		Return([]domain.WorktreeInfo{
			{Branch: "feature-branch", Path: worktreePath},
			{Branch: "other-branch", Path: "/home/user/Worktrees/my-project/other-branch"},
		}, nil)
	mockGitService.MockGoGitClient.On("ListBranches", mock.Anything, projectPath).
		Return([]domain.BranchInfo{}, nil)
	mockGitService.MockGoGitClient.On("GetRepositoryStatus", mock.Anything, worktreePath).
		Return(domain.RepositoryStatus{IsClean: false}, nil)

	resolver := NewContextResolver(config, mockGitService)

	ctx := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "my-project",
		BranchName:  "feature-branch",
		Path:        worktreePath,
	}

	suggestions, err := resolver.GetResolutionSuggestions(ctx, "")
	require.NoError(t, err)

	for _, sug := range suggestions {
		if sug.Text == "feature-branch" {
			assert.True(t, sug.IsCurrent, "Current worktree should have IsCurrent=true")
			assert.True(t, sug.IsDirty, "Dirty worktree should have IsDirty=true")
			assert.Contains(t, sug.Description, "⚠", "Dirty worktree description should have warning indicator")
		}
		if sug.Text == "other-branch" {
			assert.False(t, sug.IsCurrent, "Other worktree should have IsCurrent=false")
			assert.False(t, sug.IsDirty, "Other worktree should not have IsDirty set (performance optimization)")
		}
	}
}
