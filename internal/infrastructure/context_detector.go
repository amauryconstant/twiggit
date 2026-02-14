package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"twiggit/internal/domain"
)

type worktreeCacheEntry struct {
	valid     bool
	expiresAt time.Time
}

type contextDetector struct {
	config *domain.Config
	cache  map[string]worktreeCacheEntry
	mu     sync.RWMutex
	ttl    time.Duration
}

// NewContextDetector creates a new context detector
func NewContextDetector(cfg *domain.Config) domain.ContextDetector {
	ttl := parseTTL(cfg.ContextDetection.CacheTTL, 5*time.Second)
	return &contextDetector{
		config: cfg,
		cache:  make(map[string]worktreeCacheEntry),
		ttl:    ttl,
	}
}

func parseTTL(ttlStr string, defaultTTL time.Duration) time.Duration {
	if ttlStr == "" {
		return defaultTTL
	}
	if d, err := time.ParseDuration(ttlStr); err == nil {
		return d
	}
	return defaultTTL
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

	// Normalize path and resolve symlinks
	normalizedDir, err := NormalizePath(dir)
	if err != nil {
		return nil, domain.NewContextDetectionError(dir, "failed to normalize directory", err)
	}

	// Perform detection
	ctx := cd.detectContextInternal(normalizedDir)
	if ctx == nil {
		return nil, domain.NewContextDetectionError(normalizedDir, "failed to detect context for directory", nil)
	}

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
	gitDir := FindGitDirByTraversal(dir)
	if gitDir != nil {
		projectName := cd.extractProjectName(*gitDir)

		return &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: projectName,
			Path:        *gitDir,
			Explanation: fmt.Sprintf("In project directory '%s'", projectName),
		}
	}

	return nil
}

func (cd *contextDetector) isValidGitWorktree(dir string) bool {
	now := time.Now()

	cd.mu.RLock()
	if entry, ok := cd.cache[dir]; ok && entry.expiresAt.After(now) {
		cd.mu.RUnlock()
		return entry.valid
	}
	cd.mu.RUnlock()

	valid := cd.checkValidGitWorktree(dir)

	cd.mu.Lock()
	cd.cache[dir] = worktreeCacheEntry{
		valid:     valid,
		expiresAt: now.Add(cd.ttl),
	}
	cd.mu.Unlock()

	return valid
}

func (cd *contextDetector) checkValidGitWorktree(dir string) bool {
	gitPath := filepath.Join(dir, ".git")

	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}

	if !info.Mode().IsRegular() {
		return false
	}

	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "gitdir:")
}

func (cd *contextDetector) extractProjectName(dir string) string {
	// Extract project name from directory path
	// Use the directory name as project name
	return filepath.Base(dir)
}
