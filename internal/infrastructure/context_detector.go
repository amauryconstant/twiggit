package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"twiggit/internal/domain"
)

type contextDetector struct {
	config  *domain.Config
	cache   map[string]*domain.Context
	cacheMu sync.RWMutex
}

// NewContextDetector creates a new context detector
func NewContextDetector(cfg *domain.Config) domain.ContextDetector {
	return &contextDetector{config: cfg}
}

func (cd *contextDetector) DetectContext(dir string) (*domain.Context, error) {
	// Validate input directory
	if dir == "" {
		return nil, domain.NewContextDetectionError("", "empty directory path", nil)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil, domain.NewContextDetectionError(dir, "directory does not exist", err)
		}
		return nil, domain.NewContextDetectionError(dir, "cannot access directory", err)
	}

	// Normalize path and resolve symlinks first for consistent cache keys
	normalizedDir, err := NormalizePath(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize directory: %w", err)
	}

	// Check cache first using normalized path as key
	cd.cacheMu.RLock()
	if cached, exists := cd.cache[normalizedDir]; exists {
		cd.cacheMu.RUnlock()
		return cached, nil
	}
	cd.cacheMu.RUnlock()

	// Perform detection
	ctx := cd.detectContextInternal(normalizedDir)
	if ctx == nil {
		return nil, fmt.Errorf("failed to detect context for directory: %s", normalizedDir)
	}

	// Cache result using normalized path as key
	cd.cacheMu.Lock()
	if cd.cache == nil {
		cd.cache = make(map[string]*domain.Context)
	}
	cd.cache[normalizedDir] = ctx
	cd.cacheMu.Unlock()

	return ctx, nil
}

func (cd *contextDetector) detectContextInternal(dir string) *domain.Context {
	// Priority 1: Check worktree pattern first
	if ctx := cd.detectWorktreeContext(dir); ctx != nil {
		return ctx
	}

	// Priority 2: Check project context
	if ctx := cd.detectProjectContext(dir); ctx != nil {
		return ctx
	}

	// Priority 3: Outside git context
	return &domain.Context{
		Type:        domain.ContextOutsideGit,
		Path:        dir,
		Explanation: "Not in a git repository or worktree",
	}
}

func (cd *contextDetector) detectWorktreeContext(dir string) *domain.Context {
	// Normalize worktree directory
	worktreeDir := filepath.Clean(cd.config.WorktreesDirectory)

	// Quick check: if not under worktrees dir, exit early
	if !strings.HasPrefix(dir, worktreeDir+string(filepath.Separator)) {
		return nil
	}

	// Check if current directory is under worktree directory
	relPath, err := filepath.Rel(worktreeDir, dir)
	if err != nil {
		return nil // Not under worktree directory
	}

	// Split relative path to extract project and branch
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		return nil // Not in project/branch structure
	}

	projectName := parts[0]
	branchName := parts[1]

	// Construct worktree root path for validation
	worktreeRoot := filepath.Join(worktreeDir, projectName, branchName)

	// Validate the worktree root (not current dir) has .git file
	if !cd.isValidGitWorktree(worktreeRoot) {
		return nil
	}

	return &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: projectName,
		BranchName:  branchName,
		Path:        dir,
		Explanation: fmt.Sprintf("In worktree for project '%s' on branch '%s'", projectName, branchName),
	}
}

func (cd *contextDetector) detectProjectContext(dir string) *domain.Context {
	currentDir := dir

	// Traverse up directory tree looking for .git
	for {
		gitPath := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Found .git directory
			projectName := cd.extractProjectName(currentDir)

			return &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: projectName,
				Path:        currentDir,
				Explanation: fmt.Sprintf("In project directory '%s'", projectName),
			}
		}

		// Move to parent directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached filesystem root
			break
		}
		currentDir = parent
	}

	return nil
}

func (cd *contextDetector) isValidGitWorktree(dir string) bool {
	gitPath := filepath.Join(dir, ".git")

	// Check if .git exists and is a file (worktree indicator)
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}

	if !info.Mode().IsRegular() {
		return false // Should be a regular file for worktrees
	}

	// Read .git file to verify it's a worktree
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}

	// Worktree .git files contain: "gitdir: <path>"
	return strings.Contains(string(content), "gitdir:")
}

func (cd *contextDetector) extractProjectName(dir string) string {
	// Extract project name from directory path
	// Use the directory name as project name
	return filepath.Base(dir)
}

// InvalidateCacheForRepo removes all cache entries related to a repository path
func (cd *contextDetector) InvalidateCacheForRepo(repoPath string) {
	// Normalize the repo path first
	normalizedRepoPath, err := NormalizePath(repoPath)
	if err != nil {
		return // If we can't normalize, skip invalidation
	}

	cd.cacheMu.Lock()
	defer cd.cacheMu.Unlock()

	// Remove cache entries that are under the repository path
	keysToDelete := make([]string, 0)
	for cacheKey := range cd.cache {
		// Check if the cached path is under the repository
		if isPathUnder, err := IsPathUnder(normalizedRepoPath, cacheKey); err == nil && isPathUnder {
			keysToDelete = append(keysToDelete, cacheKey)
		}
	}

	// Delete the identified keys
	for _, key := range keysToDelete {
		delete(cd.cache, key)
	}
}

// ClearCache empties the entire context cache
func (cd *contextDetector) ClearCache() {
	cd.cacheMu.Lock()
	defer cd.cacheMu.Unlock()

	cd.cache = make(map[string]*domain.Context)
}
