package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"twiggit/internal/application"
	"twiggit/internal/core"
)

const contextDetectorCacheSize = 256

type worktreeCacheEntry struct {
	valid     bool
	expiresAt time.Time
}

type contextDetectorCacheFactory func(size int) (*lru.Cache[string, worktreeCacheEntry], error)

func defaultContextDetectorCacheFactory(size int) (*lru.Cache[string, worktreeCacheEntry], error) {
	cache, err := lru.New[string, worktreeCacheEntry](size)
	if err != nil {
		return nil, fmt.Errorf("create context detector LRU cache: %w", err)
	}
	return cache, nil
}

type contextDetector struct {
	config *core.Config
	cache  *lru.Cache[string, worktreeCacheEntry]
	mu     sync.RWMutex
	ttl    time.Duration
}

func NewContextDetector(cfg *core.Config) (application.ContextDetector, error) {
	return newContextDetectorWithCacheFactory(cfg, defaultContextDetectorCacheFactory)
}

func newContextDetectorWithCacheFactory(cfg *core.Config, factory contextDetectorCacheFactory) (application.ContextDetector, error) {
	ttl := parseTTL(cfg.ContextDetection.CacheTTL, 5*time.Second)
	cache, err := factory(contextDetectorCacheSize)
	if err != nil {
		return nil, err
	}
	return &contextDetector{
		config: cfg,
		cache:  cache,
		ttl:    ttl,
	}, nil
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

func (cd *contextDetector) DetectContext(dir string) (*core.Context, error) {
	// Validate input directory
	if dir == "" {
		return nil, core.NewContextDetectionError("", "empty directory path", nil)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, core.NewContextDetectionError(dir, "directory does not exist", err)
		}
		return nil, core.NewContextDetectionError(dir, "cannot access directory", err)
	}

	// Normalize path and resolve symlinks
	normalizedDir, err := core.NormalizePath(dir)
	if err != nil {
		return nil, core.NewContextDetectionError(dir, "failed to normalize directory", err)
	}

	// Perform detection
	ctx := cd.detectContextInternal(normalizedDir)
	if ctx == nil {
		return nil, core.NewContextDetectionError(normalizedDir, "failed to detect context for directory", nil)
	}

	return ctx, nil
}

func (cd *contextDetector) detectContextInternal(dir string) *core.Context {
	// Priority 1: Check worktree pattern first
	if ctx := cd.detectWorktreeContext(dir); ctx != nil {
		return ctx
	}

	// Priority 2: Check project context
	if ctx := cd.detectProjectContext(dir); ctx != nil {
		return ctx
	}

	// Priority 3: Outside git context
	return &core.Context{
		Type:        core.ContextOutsideGit,
		Path:        dir,
		Explanation: "Not in a git repository or worktree",
	}
}

func (cd *contextDetector) detectWorktreeContext(dir string) *core.Context {
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

	return &core.Context{
		Type:        core.ContextWorktree,
		ProjectName: projectName,
		BranchName:  branchName,
		Path:        dir,
		Explanation: fmt.Sprintf("In worktree for project '%s' on branch '%s'", projectName, branchName),
	}
}

func (cd *contextDetector) detectProjectContext(dir string) *core.Context {
	gitDir, ok := core.FindGitDirByTraversal(dir)
	if ok {
		projectName := cd.extractProjectName(gitDir)

		return &core.Context{
			Type:        core.ContextProject,
			ProjectName: projectName,
			Path:        gitDir,
			Explanation: fmt.Sprintf("In project directory '%s'", projectName),
		}
	}

	return nil
}

func (cd *contextDetector) isValidGitWorktree(dir string) bool {
	now := time.Now()

	cd.mu.RLock()
	if entry, ok := cd.cache.Get(dir); ok && entry.expiresAt.After(now) {
		cd.mu.RUnlock()
		return entry.valid
	}
	cd.mu.RUnlock()

	valid := cd.checkValidGitWorktree(dir)

	cd.mu.Lock()
	cd.cache.Add(dir, worktreeCacheEntry{
		valid:     valid,
		expiresAt: now.Add(cd.ttl),
	})
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

	content, err := os.ReadFile(gitPath) // #nosec G304 -- gitPath is always .git file in known worktree location
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
