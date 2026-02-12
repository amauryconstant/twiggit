package infrastructure

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type ContextDetectorTestSuite struct {
	suite.Suite
	config *domain.Config
}

func (s *ContextDetectorTestSuite) SetupTest() {
	tempDir := s.T().TempDir()
	s.config = &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}
}

func TestContextDetector(t *testing.T) {
	suite.Run(t, new(ContextDetectorTestSuite))
}

func (s *ContextDetectorTestSuite) TestContextTypeString() {
	tests := []struct {
		name     string
		context  domain.ContextType
		expected string
	}{
		{"unknown", domain.ContextUnknown, "unknown"},
		{"project", domain.ContextProject, "project"},
		{"worktree", domain.ContextWorktree, "worktree"},
		{"outside git", domain.ContextOutsideGit, "outside-git"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.expected, tc.context.String())
		})
	}
}

func (s *ContextDetectorTestSuite) TestPathTypeString() {
	tests := []struct {
		name     string
		pathType domain.PathType
		expected string
	}{
		{"project", domain.PathTypeProject, "project"},
		{"worktree", domain.PathTypeWorktree, "worktree"},
		{"invalid", domain.PathTypeInvalid, "invalid"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.expected, tc.pathType.String())
		})
	}
}

func (s *ContextDetectorTestSuite) TestDetectContext() {
	tests := []struct {
		name           string
		setupFunc      func() string
		expectedType   domain.ContextType
		expectedProj   string
		expectedBranch string
		expectError    bool
	}{
		{
			name: "project context with .git directory",
			setupFunc: func() string {
				dir := s.T().TempDir()
				s.Require().NoError(os.Mkdir(filepath.Join(dir, ".git"), 0755))
				return dir
			},
			expectedType: domain.ContextProject,
		},
		{
			name: "worktree context in worktree pattern",
			setupFunc: func() string {
				tempDir := s.T().TempDir()
				worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
				s.Require().NoError(os.MkdirAll(worktreeDir, 0755))

				gitFile := filepath.Join(worktreeDir, ".git")
				s.Require().NoError(os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))

				return worktreeDir
			},
			expectedType:   domain.ContextWorktree,
			expectedProj:   "test-project",
			expectedBranch: "feature-branch",
		},
		{
			name: "outside git context",
			setupFunc: func() string {
				dir := s.T().TempDir()
				return dir
			},
			expectedType: domain.ContextOutsideGit,
		},
		{
			name: "empty directory path",
			setupFunc: func() string {
				return ""
			},
			expectError: true,
		},
		{
			name: "nonexistent directory",
			setupFunc: func() string {
				return "/nonexistent/directory"
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			dir := tc.setupFunc()

			var config *domain.Config
			if tc.expectedType == domain.ContextWorktree {
				baseDir := dir
				for i := 0; i < 3; i++ {
					baseDir = filepath.Dir(baseDir)
				}
				config = &domain.Config{
					WorktreesDirectory: filepath.Join(baseDir, "Worktrees"),
				}
			} else {
				config = &domain.Config{
					WorktreesDirectory: filepath.Join(filepath.Dir(dir), "Worktrees"),
				}
			}

			detector := NewContextDetector(config)
			ctx, err := detector.DetectContext(dir)

			if tc.expectError {
				s.Error(err)
				return
			}

			s.Require().NoError(err)
			s.Equal(tc.expectedType, ctx.Type)
			if tc.expectedProj != "" {
				s.Equal(tc.expectedProj, ctx.ProjectName)
			}
			s.Equal(tc.expectedBranch, ctx.BranchName)
			s.NotEmpty(ctx.Explanation)
		})
	}
}

func (s *ContextDetectorTestSuite) TestWorktreePriority() {
	tempDir := s.T().TempDir()

	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "main")
	s.Require().NoError(os.MkdirAll(worktreeDir, 0755))

	gitFile := filepath.Join(worktreeDir, ".git")
	s.Require().NoError(os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	s.Require().NoError(err)
	s.Equal(domain.ContextWorktree, ctx.Type)
	s.Equal("test-project", ctx.ProjectName)
	s.Equal("main", ctx.BranchName)
}

func (s *ContextDetectorTestSuite) TestProjectTraversal() {
	tempDir := s.T().TempDir()

	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	s.Require().NoError(os.MkdirAll(nestedDir, 0755))

	gitDir := filepath.Join(tempDir, ".git")
	s.Require().NoError(os.Mkdir(gitDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(nestedDir)

	s.Require().NoError(err)
	s.Equal(domain.ContextProject, ctx.Type)
	s.Equal(filepath.Base(tempDir), ctx.ProjectName)
	s.Equal(tempDir, ctx.Path)
}

func (s *ContextDetectorTestSuite) TestInvalidWorktree() {
	tempDir := s.T().TempDir()

	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
	s.Require().NoError(os.MkdirAll(worktreeDir, 0755))

	gitDir := filepath.Join(worktreeDir, ".git")
	s.Require().NoError(os.Mkdir(gitDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	s.Require().NoError(err)
	s.NotEqual(domain.ContextWorktree, ctx.Type)
}

func (s *ContextDetectorTestSuite) TestCrossPlatform() {
	if runtime.GOOS == "windows" {
		s.Run("windows paths", func() {
			s.testWindowsPaths()
		})
	} else {
		s.Run("unix paths", func() {
			s.testUnixPaths()
		})
	}
}

func (s *ContextDetectorTestSuite) testWindowsPaths() {
	s.T().Helper()

	tempDir := s.T().TempDir()

	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	s.Require().NotNil(detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	s.Require().NoError(os.MkdirAll(projectDir, 0755))
	s.Require().NoError(os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	s.Require().NoError(err)
	s.Equal(domain.ContextProject, ctx.Type)
}

func (s *ContextDetectorTestSuite) testUnixPaths() {
	s.T().Helper()

	tempDir := s.T().TempDir()

	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	s.Require().NotNil(detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	s.Require().NoError(os.MkdirAll(projectDir, 0755))
	s.Require().NoError(os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	s.Require().NoError(err)
	s.Equal(domain.ContextProject, ctx.Type)
}

func (s *ContextDetectorTestSuite) TestContextDetectionError() {
	err := domain.NewContextDetectionError("/test/path", "test message", nil)

	s.Equal("context detection failed for /test/path: test message", err.Error())
	s.Require().NoError(err.Unwrap())

	originalErr := assert.AnError
	err = domain.NewContextDetectionError("/test/path", "test message", originalErr)

	s.Equal("context detection failed for /test/path: test message", err.Error())
	s.Equal(originalErr, err.Unwrap())
}
