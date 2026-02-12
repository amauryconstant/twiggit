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
