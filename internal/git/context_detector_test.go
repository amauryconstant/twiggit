package git

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
)

func TestContextDetector_ContextTypeString(t *testing.T) {
	tests := []struct {
		name     string
		context  core.ContextType
		expected string
	}{
		{"unknown", core.ContextUnknown, "unknown"},
		{"project", core.ContextProject, "project"},
		{"worktree", core.ContextWorktree, "worktree"},
		{"outside git", core.ContextOutsideGit, "outside-git"},
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
		pathType core.PathType
		expected string
	}{
		{"project", core.PathTypeProject, "project"},
		{"worktree", core.PathTypeWorktree, "worktree"},
		{"invalid", core.PathTypeInvalid, "invalid"},
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
		expectedType   core.ContextType
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
			expectedType: core.ContextProject,
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
			expectedType:   core.ContextWorktree,
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
			expectedType: core.ContextOutsideGit,
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

			var config *core.Config
			if tc.expectedType == core.ContextWorktree {
				baseDir := dir
				for range 3 {
					baseDir = filepath.Dir(baseDir)
				}
				config = &core.Config{
					WorktreesDirectory: filepath.Join(baseDir, "Worktrees"),
				}
			} else {
				config = &core.Config{
					WorktreesDirectory: filepath.Join(filepath.Dir(dir), "Worktrees"),
				}
			}

			detector, err := NewContextDetector(config)
			require.NoError(t, err)
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

	config := &core.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector, err := NewContextDetector(config)
	require.NoError(t, err)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	assert.Equal(t, core.ContextWorktree, ctx.Type)
	assert.Equal(t, "test-project", ctx.ProjectName)
	assert.Equal(t, "main", ctx.BranchName)
}

func TestContextDetector_ProjectTraversal(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	gitDir := filepath.Join(tempDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	config := &core.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector, err := NewContextDetector(config)
	require.NoError(t, err)
	ctx, err := detector.DetectContext(nestedDir)

	require.NoError(t, err)
	assert.Equal(t, core.ContextProject, ctx.Type)
	assert.Equal(t, filepath.Base(tempDir), ctx.ProjectName)
	assert.Equal(t, tempDir, ctx.Path)
}

func TestContextDetector_InvalidWorktree(t *testing.T) {
	tempDir := t.TempDir()

	worktreeDir := filepath.Join(tempDir, "Worktrees", "test-project", "feature-branch")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	gitDir := filepath.Join(worktreeDir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0755))

	config := &core.Config{
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector, err := NewContextDetector(config)
	require.NoError(t, err)
	ctx, err := detector.DetectContext(worktreeDir)

	require.NoError(t, err)
	assert.NotEqual(t, core.ContextWorktree, ctx.Type)
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

	config := &core.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector, err := NewContextDetector(config)
	require.NoError(t, err)
	require.NotNil(t, detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, core.ContextProject, ctx.Type)
}

func testUnixPaths(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()

	config := &core.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	detector, err := NewContextDetector(config)
	require.NoError(t, err)
	require.NotNil(t, detector)

	projectDir := filepath.Join(config.ProjectsDirectory, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, ".git"), 0755))

	ctx, err := detector.DetectContext(projectDir)
	require.NoError(t, err)
	assert.Equal(t, core.ContextProject, ctx.Type)
}

func TestContextDetectionError(t *testing.T) {
	err := core.NewContextDetectionError("/test/path", "test message", nil)

	assert.Equal(t, "context.detection: test message (entity: /test/path)", err.Error())
	require.NoError(t, err.Unwrap())

	originalErr := assert.AnError
	err = core.NewContextDetectionError("/test/path", "test message", originalErr)

	assert.Equal(t, "context.detection: test message (entity: /test/path): assert.AnError general error for testing", err.Error())
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestNewContextDetector_CacheAllocatorFailure(t *testing.T) {
	allocErr := errors.New("simulated cache allocation failure")
	failingFactory := func(_ int) (*lru.Cache[string, worktreeCacheEntry], error) {
		return nil, allocErr
	}

	detector, err := newContextDetectorWithCacheFactory(
		&core.Config{WorktreesDirectory: "/tmp/wt"},
		failingFactory,
	)
	require.Error(t, err)
	assert.Nil(t, detector)
	assert.ErrorIs(t, err, allocErr)
}

func TestNewContextDetector_CacheFactoryReceivesRequestedSize(t *testing.T) {
	var receivedSize int
	captureFactory := func(size int) (*lru.Cache[string, worktreeCacheEntry], error) {
		receivedSize = size
		return defaultContextDetectorCacheFactory(size)
	}

	detector, err := newContextDetectorWithCacheFactory(
		&core.Config{WorktreesDirectory: "/tmp/wt"},
		captureFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, detector)
	assert.Equal(t, contextDetectorCacheSize, receivedSize)
}

func TestNewContextDetector_ParseTTLFallback(t *testing.T) {
	tests := []struct {
		name     string
		ttl      string
		fallback time.Duration
		want     time.Duration
	}{
		{"empty uses fallback", "", 5 * time.Second, 5 * time.Second},
		{"valid duration parses", "30s", 5 * time.Second, 30 * time.Second},
		{"invalid duration falls back", "not-a-duration", 7 * time.Second, 7 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTTL(tt.ttl, tt.fallback)
			assert.Equal(t, tt.want, got)
		})
	}
}
