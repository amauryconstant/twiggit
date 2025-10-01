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

func TestContextType_String(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.context.String())
		})
	}
}

func TestPathType_String(t *testing.T) {
	tests := []struct {
		name     string
		pathType domain.PathType
		expected string
	}{
		{"project", domain.PathTypeProject, "project"},
		{"worktree", domain.PathTypeWorktree, "worktree"},
		{"invalid", domain.PathTypeInvalid, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pathType.String())
		})
	}
}

func TestContextDetector_DetectContext(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*testing.T) string
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
				worktreesDir := filepath.Join(tempDir, "Worktrees")
				worktreeDir := filepath.Join(worktreesDir, "test-project", "feature-branch")
				require.NoError(t, os.MkdirAll(worktreeDir, 0755))

				// Create worktree .git file
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupFunc(t)

			// For worktree tests, we need to set up the config properly
			var config *domain.Config
			if tt.expectedType == domain.ContextWorktree {
				// Extract the base temp directory and construct worktrees path
				baseDir := dir
				for i := 0; i < 3; i++ { // Go up 3 levels: feature-branch -> test-project -> Worktrees -> tempDir
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

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, ctx.Type)
			if tt.expectedProj != "" {
				assert.Equal(t, tt.expectedProj, ctx.ProjectName)
			}
			assert.Equal(t, tt.expectedBranch, ctx.BranchName)
			assert.NotEmpty(t, ctx.Explanation)
		})
	}
}

func TestContextDetector_WorktreePriority(t *testing.T) {
	tempDir := t.TempDir()

	// Create a directory that matches both project and worktree patterns
	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "main")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	// Create worktree .git file
	gitFile := filepath.Join(worktreeDir, ".git")
	require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	// Worktree detection should take priority
	assert.Equal(t, domain.ContextWorktree, ctx.Type)
	assert.Equal(t, "test-project", ctx.ProjectName)
	assert.Equal(t, "main", ctx.BranchName)
}

func TestContextDetector_ProjectTraversal(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	// Create .git at the root
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

	// Create worktree directory structure but without proper .git file
	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	// Create .git as directory instead of file
	gitDir := filepath.Join(worktreeDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	// Should not detect as worktree since .git is not a file with gitdir:
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

	// Test Windows-specific path handling
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

	// Test Unix path handling
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

	// Test with cause
	originalErr := assert.AnError
	err = domain.NewContextDetectionError("/test/path", "test message", originalErr)

	assert.Equal(t, "context detection failed for /test/path: test message", err.Error())
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestContextDetector_CacheKeyNormalization(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tempDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)

	// Test that different paths to the same directory use the same cache key
	// Use absolute path and a path with ./ prefix that should resolve to the same location
	absolutePath := projectDir
	withDotPrefix := filepath.Join(tempDir, "./test-project")

	// Detect context using absolute path
	ctx1, err := detector.DetectContext(absolutePath)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1.Type)

	// Detect context using path with ./ prefix - should hit cache
	ctx2, err := detector.DetectContext(withDotPrefix)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2.Type)

	// Both contexts should be the same object (from cache)
	assert.Same(t, ctx1, ctx2)
}

func TestContextDetector_CacheKeyNormalizationWithSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests skipped on Windows")
	}

	tempDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tempDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	// Create a symlink to the project directory
	symlinkPath := filepath.Join(tempDir, "symlink-project")
	require.NoError(t, os.Symlink(projectDir, symlinkPath))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)

	// Detect context using original path
	ctx1, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1.Type)

	// Detect context using symlink - should hit cache
	ctx2, err := detector.DetectContext(symlinkPath)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2.Type)

	// Both contexts should be the same object (from cache due to normalized path)
	assert.Same(t, ctx1, ctx2)
	assert.Equal(t, ctx1.Path, ctx2.Path)
}

func TestContextDetector_InvalidateCacheForRepo(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tempDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	// Create a subdirectory
	subDir := filepath.Join(projectDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)

	// Detect context for both directories
	ctx1, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1.Type)

	ctx2, err := detector.DetectContext(subDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2.Type)

	// Verify both are cached (second call should return same object)
	ctx1Cached, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Same(t, ctx1, ctx1Cached)

	ctx2Cached, err := detector.DetectContext(subDir)
	require.NoError(t, err)
	assert.Same(t, ctx2, ctx2Cached)

	// Invalidate cache for the repository
	detector.InvalidateCacheForRepo(projectDir)

	// Detect again - should create new objects (cache was cleared)
	ctx1New, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1New.Type)
	assert.NotSame(t, ctx1, ctx1New) // Different object from cache

	ctx2New, err := detector.DetectContext(subDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2New.Type)
	assert.NotSame(t, ctx2, ctx2New) // Different object from cache
}

func TestContextDetector_InvalidateCacheForRepoWithPathVariations(t *testing.T) {
	tempDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tempDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	// Create a subdirectory
	subDir := filepath.Join(projectDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	config := &domain.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector := NewContextDetector(config)

	// Detect context for both directories to populate cache
	ctx1, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1.Type)

	ctx2, err := detector.DetectContext(subDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2.Type)

	// Verify both are cached
	ctx1Cached, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Same(t, ctx1, ctx1Cached)

	// Invalidate cache using different path format (with ./ prefix)
	detector.InvalidateCacheForRepo(filepath.Join(tempDir, "./test-project"))

	// Detect again - should create new objects (cache was cleared)
	ctx1New, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1New.Type)
	assert.NotSame(t, ctx1, ctx1New) // Different object from cache

	ctx2New, err := detector.DetectContext(subDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2New.Type)
	assert.NotSame(t, ctx2, ctx2New) // Different object from cache
}
