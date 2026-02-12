//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
)

func TestContextDetector_Integration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available for integration tests")
	}

	tests := []struct {
		name           string
		setupFunc      func(*testing.T, *domain.Config) string
		expectedType   domain.ContextType
		expectedProj   string
		expectedBranch string
	}{
		{
			name: "real git repository detection",
			setupFunc: func(t *testing.T, config *domain.Config) string {
				t.Helper()
				repoDir := filepath.Join(config.ProjectsDirectory, "test-repo")
				require.NoError(t, os.MkdirAll(repoDir, 0755))

				// Use git to initialize repository
				cmd := exec.Command("git", "init")
				cmd.Dir = repoDir
				require.NoError(t, cmd.Run())

				return repoDir
			},
			expectedType: domain.ContextProject,
			expectedProj: "test-repo",
		},
		{
			name: "real git worktree detection",
			setupFunc: func(t *testing.T, config *domain.Config) string {
				t.Helper()
				// Setup main repository
				mainRepo := filepath.Join(config.ProjectsDirectory, "main-repo")
				require.NoError(t, os.MkdirAll(mainRepo, 0755))

				// Initialize main repo
				cmd := exec.Command("git", "init")
				cmd.Dir = mainRepo
				require.NoError(t, cmd.Run())

				// Configure git user
				cmd = exec.Command("git", "config", "user.email", "test@example.com")
				cmd.Dir = mainRepo
				require.NoError(t, cmd.Run())

				cmd = exec.Command("git", "config", "user.name", "Test User")
				cmd.Dir = mainRepo
				require.NoError(t, cmd.Run())

				// Create initial commit
				cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
				cmd.Dir = mainRepo
				require.NoError(t, cmd.Run())

				// Create worktree
				worktreeDir := filepath.Join(config.WorktreesDirectory, "main-repo", "feature-branch")
				require.NoError(t, os.MkdirAll(filepath.Dir(worktreeDir), 0755))

				cmd = exec.Command("git", "worktree", "add", worktreeDir, "-b", "feature-branch")
				cmd.Dir = mainRepo
				require.NoError(t, cmd.Run())

				return worktreeDir
			},
			expectedType:   domain.ContextWorktree,
			expectedProj:   "main-repo",
			expectedBranch: "feature-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			config := &domain.Config{
				ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
				WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
			}

			detector := infrastructure.NewContextDetector(config)
			testDir := tt.setupFunc(t, config)

			ctx, err := detector.DetectContext(testDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, ctx.Type)
			assert.Equal(t, tt.expectedProj, ctx.ProjectName)
			if tt.expectedBranch != "" {
				assert.Equal(t, tt.expectedBranch, ctx.BranchName)
			}
		})
	}
}

func TestContextResolver_Integration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available for integration tests")
	}

	tempDir := t.TempDir()
	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	// Setup a real git repository with worktree
	mainRepo := filepath.Join(config.ProjectsDirectory, "test-project")
	require.NoError(t, os.MkdirAll(mainRepo, 0755))

	// Initialize main repo
	cmd := exec.Command("git", "init")
	cmd.Dir = mainRepo
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = mainRepo
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = mainRepo
	require.NoError(t, cmd.Run())

	// Create initial commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = mainRepo
	require.NoError(t, cmd.Run())

	// Create worktree
	worktreeDir := filepath.Join(config.WorktreesDirectory, "test-project", "feature-branch")
	require.NoError(t, os.MkdirAll(filepath.Dir(worktreeDir), 0755))

	cmd = exec.Command("git", "worktree", "add", worktreeDir, "-b", "feature-branch")
	cmd.Dir = mainRepo
	require.NoError(t, cmd.Run())

	// Test resolver from project context
	detector := infrastructure.NewContextDetector(config)
	gitService := infrastructure.NewCompositeGitClient(nil, nil) // Mock service for integration test
	resolver := infrastructure.NewContextResolver(config, gitService)

	projectCtx, err := detector.DetectContext(mainRepo)
	require.NoError(t, err)

	// Test resolving "main" from project context
	result, err := resolver.ResolveIdentifier(projectCtx, "main")
	require.NoError(t, err)
	assert.Equal(t, domain.PathTypeProject, result.Type)
	assert.Equal(t, "test-project", result.ProjectName)
	assert.Equal(t, mainRepo, result.ResolvedPath)

	// Test resolving branch name from project context
	result, err = resolver.ResolveIdentifier(projectCtx, "feature-branch")
	require.NoError(t, err)
	assert.Equal(t, domain.PathTypeWorktree, result.Type)
	assert.Equal(t, "test-project", result.ProjectName)
	assert.Equal(t, "feature-branch", result.BranchName)
	assert.Equal(t, worktreeDir, result.ResolvedPath)

	// Test resolver from worktree context
	worktreeCtx, err := detector.DetectContext(worktreeDir)
	require.NoError(t, err)

	// Test resolving "main" from worktree context
	result, err = resolver.ResolveIdentifier(worktreeCtx, "main")
	require.NoError(t, err)
	assert.Equal(t, domain.PathTypeProject, result.Type)
	assert.Equal(t, "test-project", result.ProjectName)
	assert.Equal(t, mainRepo, result.ResolvedPath)
}

func TestContextDetector_CacheNormalizationWithSymlinks_Integration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available for integration tests")
	}

	tempDir := t.TempDir()
	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	// Setup a real git repository
	repoDir := filepath.Join(config.ProjectsDirectory, "symlink-test")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Create a symlink to the repository
	symlinkPath := filepath.Join(tempDir, "symlink-repo")
	require.NoError(t, os.Symlink(repoDir, symlinkPath))

	detector := infrastructure.NewContextDetector(config)

	// Detect context using original path
	ctx1, err := detector.DetectContext(repoDir)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx1.Type)
	assert.Equal(t, "symlink-test", ctx1.ProjectName)

	// Detect context using symlink - should hit cache due to normalized path
	ctx2, err := detector.DetectContext(symlinkPath)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx2.Type)
	assert.Equal(t, "symlink-test", ctx2.ProjectName)

	// Both contexts should be the same object (from cache)
	assert.Same(t, ctx1, ctx2)
	assert.Equal(t, ctx1.Path, ctx2.Path)

	// Invalidate cache using original path
	detector.InvalidateCacheForRepo(repoDir)

	// Detect again using symlink - should create new object
	ctx3, err := detector.DetectContext(symlinkPath)
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx3.Type)
	assert.NotSame(t, ctx1, ctx3) // Different object from cache
}

func TestContextService_Integration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available for integration tests")
	}

	tempDir := t.TempDir()
	config := &domain.Config{
		ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
		WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
	}

	// Setup a real git repository
	repoDir := filepath.Join(config.ProjectsDirectory, "service-test")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run())

	// Create context service
	detector := infrastructure.NewContextDetector(config)

	// Create real git service for integration testing
	executor := infrastructure.NewDefaultCommandExecutor(30 * time.Second)
	goGitClient := infrastructure.NewGoGitClient(true)
	cliClient := infrastructure.NewCLIClient(executor, 30)
	gitService := infrastructure.NewCompositeGitClient(goGitClient, cliClient)

	resolver := infrastructure.NewContextResolver(config, gitService)
	contextService := service.NewContextService(detector, resolver, config)

	// Change to the repository directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()
	require.NoError(t, os.Chdir(repoDir))

	// Test getting current context
	ctx, err := contextService.GetCurrentContext()
	require.NoError(t, err)
	assert.Equal(t, domain.ContextProject, ctx.Type)
	assert.Equal(t, "service-test", ctx.ProjectName)

	// Test resolving identifier from current context
	result, err := contextService.ResolveIdentifier("main")
	require.NoError(t, err)
	assert.Equal(t, domain.PathTypeProject, result.Type)
	assert.Equal(t, "service-test", result.ProjectName)
	assert.Equal(t, repoDir, result.ResolvedPath)

	// Test getting completion suggestions
	suggestions, err := contextService.GetCompletionSuggestions("m")
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
	// Check that "main" is among the suggestions
	foundMain := false
	for _, suggestion := range suggestions {
		if suggestion.Text == "main" {
			foundMain = true
			break
		}
	}
	assert.True(t, foundMain, "Expected 'main' to be among completion suggestions")
}
