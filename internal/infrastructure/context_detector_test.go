package infrastructure

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestContextDetector_ContextTypeString(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.context.String())
		})
	}
}

func TestContextDetector_PathTypeString(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.pathType.String())
		})
	}
}

func TestContextDetector_DetectContext(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) string
		expectedType   domain.ContextType
		expectedProj   string
		expectedBranch string
		expectError    bool
	}{
		{
			name: "project context with .git directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0755))
				return dir
			},
			expectedType: domain.ContextProject,
		},
		{
			name: "worktree context in worktree pattern",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				tempDir := t.TempDir()
				worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
				require.NoError(t, os.MkdirAll(worktreeDir, 0755))

				gitFile := filepath.Join(worktreeDir, ".git")
				require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))

				return worktreeDir
			},
			expectedType:   domain.ContextWorktree,
			expectedProj:   "test-project",
			expectedBranch: "feature-branch",
		},
		{
			name: "outside git context",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				return dir
			},
			expectedType: domain.ContextOutsideGit,
		},
		{
			name: "empty directory path",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			expectError: true,
		},
		{
			name: "nonexistent directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()
				return "/nonexistent/directory"
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.setupFunc(t)

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
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedType, ctx.Type)
			if tc.expectedProj != "" {
				assert.Equal(t, tc.expectedProj, ctx.ProjectName)
			}
			assert.Equal(t, tc.expectedBranch, ctx.BranchName)
			assert.NotEmpty(t, ctx.Explanation)
		})
	}
}

func TestContextDetector_WorktreePriority(t *testing.T) {
	tempDir := t.TempDir()

	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "main")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	gitFile := filepath.Join(worktreeDir, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	assert.Equal(t, domain.ContextWorktree, ctx.Type)
	assert.Equal(t, "test-project", ctx.ProjectName)
	assert.Equal(t, "main", ctx.BranchName)
}

func TestContextDetector_ProjectTraversal(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(nestedDir)

	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx.Type)
	assert.Equal(t, filepath.Base(tempDir), ctx.ProjectName)
	assert.Equal(t, tempDir, ctx.Path)
}

func TestContextDetector_InvalidWorktree(t *testing.T) {
	tempDir := t.TempDir()

	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	gitDir := filepath.Join(worktreeDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	assert.NotEqual(t, domain.ContextWorktree, ctx.Type)
}

func TestContextDetector_CrossPlatform(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Run("windows paths", func(t *testing.T) {
			testWindowsPaths(t)
		})
	} else {
		t.Run("unix paths", func(t *testing.T) {
			testUnixPaths(t)
		})
	}
}

func testWindowsPaths(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()

	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	require.NotNil(t, detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx.Type)
}

func testUnixPaths(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()

	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	require.NotNil(t, detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx.Type)
}

func TestContextDetectionError(t *testing.T) {
	err := domain.NewContextDetectionError("/test/path", "test message", nil)

	assert.Equal(t, "context detection failed for /test/path: test message", err.Error())
	require.NoError(t, err.Unwrap())

	originalErr := assert.AnError
	err = domain.NewContextDetectionError("/test/path", "test message", originalErr)

	assert.Equal(t, "context detection failed for /test/path: test message", err.Error())
	assert.Equal(t, originalErr, err.Unwrap())
}
