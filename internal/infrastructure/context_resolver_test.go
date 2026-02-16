package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

type ContextResolverTestSuite struct {
	suite.Suite
	config *domain.Config
}

func (s *ContextResolverTestSuite) SetupTest() {
	s.config = &domain.Config{
		ProjectsDirectory:  "/home/user/Projects",
		WorktreesDirectory: "/home/user/Worktrees",
	}
}

func TestContextResolver(t *testing.T) {
	suite.Run(t, new(ContextResolverTestSuite))
}

func (s *ContextResolverTestSuite) TestResolveIdentifier() {
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
		s.Run(tt.name, func() {
			resolver := NewContextResolver(s.config, nil)
			result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)

			if tt.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.expectedType, result.Type)
			s.Equal(tt.expectedProj, result.ProjectName)
			if tt.expectedBranch != "" {
				s.Equal(tt.expectedBranch, result.BranchName)
			}
			s.Equal(tt.expectedPath, result.ResolvedPath)
			s.NotEmpty(result.Explanation)
		})
	}
}

func (s *ContextResolverTestSuite) TestGetResolutionSuggestions() {
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
		s.Run(tt.name, func() {
			resolver := NewContextResolver(s.config, nil)
			suggestions, err := resolver.GetResolutionSuggestions(tt.context, tt.partial)

			s.Require().NoError(err)
			s.Len(suggestions, tt.expectedCount)

			if len(tt.expectedTexts) > 0 {
				suggestionTexts := make([]string, len(suggestions))
				for i, suggestion := range suggestions {
					suggestionTexts[i] = suggestion.Text
				}
				s.Equal(tt.expectedTexts, suggestionTexts)
			}
		})
	}
}

func (s *ContextResolverTestSuite) TestWorktreeContextResolution() {
	resolver := NewContextResolver(s.config, nil)

	ctx := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "my-project",
		BranchName:  "current-branch",
		Path:        "/home/user/Worktrees/my-project/current-branch",
	}

	result, err := resolver.ResolveIdentifier(ctx, "other-branch")
	s.Require().NoError(err)

	s.Equal(domain.PathTypeWorktree, result.Type)
	s.Equal("my-project", result.ProjectName)
	s.Equal("other-branch", result.BranchName)
	s.Equal("/home/user/Worktrees/my-project/other-branch", result.ResolvedPath)
	s.Contains(result.Explanation, "my-project")
}

func (s *ContextResolverTestSuite) TestCrossProjectReference() {
	resolver := NewContextResolver(s.config, nil)

	ctx := &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "current-project",
		Path:        "/home/user/Projects/current-project",
	}

	result, err := resolver.ResolveIdentifier(ctx, "other-project/feature-branch")
	s.Require().NoError(err)

	s.Equal(domain.PathTypeWorktree, result.Type)
	s.Equal("other-project", result.ProjectName)
	s.Equal("feature-branch", result.BranchName)
	s.Equal("/home/user/Worktrees/other-project/feature-branch", result.ResolvedPath)
	s.Contains(result.Explanation, "other-project")
}

func (s *ContextResolverTestSuite) TestParseCrossProjectReference() {
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
		s.Run(tt.name, func() {
			project, branch, valid := parseCrossProjectReference(tt.identifier)
			s.Equal(tt.expectedProj, project)
			s.Equal(tt.expectedBranch, branch)
			s.Equal(tt.expectedValid, valid)
		})
	}
}

func (s *ContextResolverTestSuite) TestBuildWorktreePath() {
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
		s.Run(tt.name, func() {
			path := buildWorktreePath(tt.worktreesDir, tt.project, tt.branch)
			s.Equal(tt.expectedPath, path)
		})
	}
}

func (s *ContextResolverTestSuite) TestBuildProjectPath() {
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
		s.Run(tt.name, func() {
			path := buildProjectPath(tt.projectsDir, tt.project)
			s.Equal(tt.expectedPath, path)
		})
	}
}

func (s *ContextResolverTestSuite) TestFilterSuggestions() {
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
		s.Run(tt.name, func() {
			result := filterSuggestions(tt.suggestions, tt.partial)
			s.Equal(tt.expectedResult, result)
		})
	}
}

func (s *ContextResolverTestSuite) TestContainsPathTraversal() {
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
		s.Run(tt.name, func() {
			result := containsPathTraversal(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ContextResolverTestSuite) TestPathTraversalProtection() {
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
		s.Run(tt.name, func() {
			resolver := NewContextResolver(s.config, nil)
			result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)

			if tt.expectError {
				s.Require().Error(err, tt.description)
				s.Contains(err.Error(), tt.errorContains, tt.description)
				s.Nil(result, tt.description)
			} else {
				s.Require().NoError(err, tt.description)
				s.NotNil(result, tt.description)
			}
		})
	}
}

func (s *ContextResolverTestSuite) TestGetOutsideGitContextSuggestions() {
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
			projectsDir: s.T().TempDir(),
			setupProjects: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "project1"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "project2"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "not-a-repo"), 0755); err != nil {
					return err
				}

				s.setupTestRepo(filepath.Join(dir, "project1"))
				s.setupTestRepo(filepath.Join(dir, "project2"))
				return nil
			},
			partial:       "proj",
			expectedCount: 2,
			expectedTexts: []string{"project1", "project2"},
		},
		{
			name:        "partial matching with 'test'",
			projectsDir: s.T().TempDir(),
			setupProjects: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "test-project"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "other-project"), 0755); err != nil {
					return err
				}

				s.setupTestRepo(filepath.Join(dir, "test-project"))
				s.setupTestRepo(filepath.Join(dir, "other-project"))
				return nil
			},
			partial:       "test",
			expectedCount: 1,
			expectedTexts: []string{"test-project"},
		},
		{
			name:        "empty partial matches all",
			projectsDir: s.T().TempDir(),
			setupProjects: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "project1"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, "project2"), 0755); err != nil {
					return err
				}

				s.setupTestRepo(filepath.Join(dir, "project1"))
				s.setupTestRepo(filepath.Join(dir, "project2"))
				return nil
			},
			partial:       "",
			expectedCount: 2,
			expectedTexts: []string{"project1", "project2"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupProjects != nil {
				if err := tt.setupProjects(tt.projectsDir); err != nil {
					s.T().Fatalf("Failed to setup test projects: %v", err)
				}
			}

			config := &domain.Config{
				ProjectsDirectory:  tt.projectsDir,
				WorktreesDirectory: "/home/user/Worktrees",
			}

			mockGitService := mocks.NewMockGitService()
			mockGitService.MockGoGitClient.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)

			resolver := NewContextResolver(config, mockGitService)
			suggestions := resolver.(*contextResolver).getOutsideGitContextSuggestions(tt.partial)

			s.Len(suggestions, tt.expectedCount, "Expected %d suggestions", tt.expectedCount)

			if len(tt.expectedTexts) > 0 {
				suggestionTexts := make([]string, len(suggestions))
				for i, suggestion := range suggestions {
					suggestionTexts[i] = suggestion.Text
				}
				s.Equal(tt.expectedTexts, suggestionTexts, "Suggestion texts don't match")
			}

			for _, suggestion := range suggestions {
				s.NotEmpty(suggestion.Text, "Suggestion text should not be empty")
				s.Equal(domain.PathTypeProject, suggestion.Type, "Suggestion type should be project")
				s.Equal(suggestion.Text, suggestion.ProjectName, "Project name should match text")
				s.Equal("Project directory", suggestion.Description, "Description should be consistent")
			}
		})
	}
}

func (s *ContextResolverTestSuite) setupTestRepo(path string) {
	s.T().Helper()

	err := os.MkdirAll(filepath.Join(path, ".git"), 0755)
	s.Require().NoError(err)

	gitDir := filepath.Join(path, ".git", "refs", "heads")
	err = os.MkdirAll(gitDir, 0755)
	s.Require().NoError(err)

	err = os.WriteFile(filepath.Join(path, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	s.Require().NoError(err)

	err = os.WriteFile(filepath.Join(gitDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0644)
	s.Require().NoError(err)
}

func (s *ContextResolverTestSuite) TestWithExistingOnlyFilter() {
	_, err := os.Stat("/tmp/nonexistent")
	s.Require().True(os.IsNotExist(err))
}
