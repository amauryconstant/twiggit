// Package services contains business logic for twiggit operations
package services

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
)

const (
	defaultConcurrency = 4
	maxConcurrency     = 16
	cacheExpiryTime    = 5 * time.Minute
	failureThreshold   = 0.5 // 50% failure rate threshold
)

// DiscoveryService handles the discovery and analysis of worktrees and projects
type DiscoveryService struct {
	gitClient   infrastructure.GitClient
	config      *config.Config
	fileSystem  fs.FS
	concurrency int
	mu          sync.RWMutex
	cache       map[string]*discoveryResult
}

type discoveryResult struct {
	worktree  *domain.Worktree
	timestamp time.Time
}

// NewDiscoveryService creates a new DiscoveryService instance
func NewDiscoveryService(gitClient infrastructure.GitClient, config *config.Config, fileSystem fs.FS) *DiscoveryService {
	return &DiscoveryService{
		gitClient:   gitClient,
		config:      config,
		fileSystem:  fileSystem,
		concurrency: defaultConcurrency,
		cache:       make(map[string]*discoveryResult),
	}
}

// SetConcurrency sets the number of concurrent workers for discovery operations
func (ds *DiscoveryService) SetConcurrency(workers int) {
	if workers > 0 && workers <= maxConcurrency {
		ds.concurrency = workers
	}
}

// pathExists checks if a path exists on the filesystem
func (ds *DiscoveryService) pathExists(path string) bool {
	// Convert to absolute path first to ensure we check the correct location
	absolutePath := ds.convertToAbsolutePath(path)
	_, err := os.Stat(absolutePath)
	return err == nil
}

// validatePath validates that a path is not empty and exists
func (ds *DiscoveryService) validatePath(path, pathType string) error {
	if path == "" {
		return fmt.Errorf("%s path cannot be empty", pathType)
	}

	if !ds.pathExists(path) {
		return fmt.Errorf("%s path does not exist: %s", pathType, path)
	}

	return nil
}

// isGitRepositorySafe safely checks if a path is a git repository, returning false on errors
func (ds *DiscoveryService) isGitRepositorySafe(ctx context.Context, path string) bool {
	isRepo, err := ds.gitClient.IsGitRepository(ctx, path)
	return err == nil && isRepo
}

// isBareRepositorySafe safely checks if a path is a bare repository, returning false on errors
func (ds *DiscoveryService) isBareRepositorySafe(ctx context.Context, path string) bool {
	isBare, err := ds.gitClient.IsBareRepository(ctx, path)
	return err == nil && isBare
}

// DiscoverWorktrees discovers all worktrees in a workspaces directory using concurrent processing
func (ds *DiscoveryService) DiscoverWorktrees(ctx context.Context, workspacesPath string) ([]*domain.Worktree, error) {
	// Check if workspaces path exists, return empty list if it doesn't
	if !ds.pathExists(workspacesPath) {
		return []*domain.Worktree{}, nil
	}

	if err := ds.validatePath(workspacesPath, "workspaces"); err != nil {
		return nil, err
	}

	// Find all potential worktree directories in the workspaces directory
	paths, err := ds.findWorktreePathsInWorkspaces(ctx, workspacesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workspaces: %w", err)
	}

	if len(paths) == 0 {
		return []*domain.Worktree{}, nil
	}

	// Process paths concurrently
	return ds.analyzePathsConcurrently(ctx, paths)
}

// AnalyzeWorktree analyzes a single worktree path and returns detailed information
func (ds *DiscoveryService) AnalyzeWorktree(ctx context.Context, path string) (*domain.Worktree, error) {
	if path == "" {
		return nil, errors.New("worktree path cannot be empty")
	}

	// Check cache first
	if cached := ds.getCachedResult(path); cached != nil {
		return cached, nil
	}

	// Convert relative path to absolute path for Git client
	absolutePath := ds.convertToAbsolutePath(path)

	// Get worktree status from git client
	status, err := ds.gitClient.GetWorktreeStatus(ctx, absolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree status for %s: %w", path, err)
	}

	// Convert to domain model
	worktree, err := ds.convertToWorktree(status)
	if err != nil {
		return nil, fmt.Errorf("failed to convert worktree info: %w", err)
	}

	// Cache result
	ds.cacheResult(path, worktree)

	return worktree, nil
}

// DiscoverProjects finds all Git repositories (projects) in the projects directory
func (ds *DiscoveryService) DiscoverProjects(ctx context.Context, projectsPath string) ([]*domain.Project, error) {
	// Check if projects path exists, return empty list if it doesn't
	if !ds.pathExists(projectsPath) {
		return []*domain.Project{}, nil
	}

	if err := ds.validatePath(projectsPath, "projects"); err != nil {
		return nil, err
	}

	// Find all directories in projects directory
	absoluteProjectsPath := ds.convertToAbsolutePath(projectsPath)
	entries, err := os.ReadDir(absoluteProjectsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	//nolint:prealloc // Number of valid projects is unpredictable due to filtering
	var projects []*domain.Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsPath, entry.Name())

		// Convert relative path to absolute path for Git client
		absoluteProjectPath := ds.convertToAbsolutePath(projectPath)

		// Check if it's a main git repository (not a worktree)
		isMainRepo, err := ds.gitClient.IsMainRepository(ctx, absoluteProjectPath)
		if err != nil {
			// Log the error but continue processing other projects
			// This prevents one failing project from blocking all discovery
			continue
		}
		if !isMainRepo {
			continue
		}

		// Create project with absolute path
		project, err := domain.NewProject(entry.Name(), absoluteProjectPath)
		if err != nil {
			continue
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// DiscoverProjectsWithFallback discovers projects with fallback mechanisms
// This method provides enhanced error recovery when primary discovery fails
func (ds *DiscoveryService) DiscoverProjectsWithFallback(ctx context.Context, projectsPath string) ([]*domain.Project, error) {
	projects, err := ds.DiscoverProjects(ctx, projectsPath)
	if err != nil {
		// Try fallback detection method
		projects, fallbackErr := ds.fallbackProjectDiscovery(ctx, projectsPath)
		if fallbackErr != nil {
			return nil, domain.NewWorkspaceError(
				domain.ErrWorkspaceDiscoveryFailed,
				"failed to discover projects with fallback",
				err,
			).WithSuggestion("Check if projects directory exists and is accessible")
		}
		return projects, nil
	}
	return projects, nil
}

// fallbackProjectDiscovery provides alternative project discovery when primary method fails
func (ds *DiscoveryService) fallbackProjectDiscovery(ctx context.Context, projectsPath string) ([]*domain.Project, error) {
	// Check if path exists at all
	if !ds.pathExists(projectsPath) {
		return []*domain.Project{}, nil
	}

	// Try to read directory with basic error handling
	absolutePath := ds.convertToAbsolutePath(projectsPath)
	entries, err := os.ReadDir(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("fallback discovery failed: %w", err)
	}

	//nolint:prealloc // Number of valid projects is unpredictable
	var projects []*domain.Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsPath, entry.Name())
		absoluteProjectPath := ds.convertToAbsolutePath(projectPath)

		// Simple git repository check without complex validation
		isRepo, err := ds.gitClient.IsGitRepository(ctx, absoluteProjectPath)
		if err != nil || !isRepo {
			continue
		}

		// Create project with basic validation
		project, err := domain.NewProject(entry.Name(), absoluteProjectPath)
		if err != nil {
			continue
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// findWorktreePathsInWorkspaces scans the workspaces directory for directories that might be worktrees
func (ds *DiscoveryService) findWorktreePathsInWorkspaces(ctx context.Context, workspacesPath string) ([]string, error) {
	var paths []string

	// Read the workspaces directory to find project subdirectories
	absoluteWorkspacesPath := ds.convertToAbsolutePath(workspacesPath)
	entries, err := os.ReadDir(absoluteWorkspacesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspaces directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(workspacesPath, entry.Name())

		// First, check if the project directory itself is a git repository (main worktree)
		absoluteProjectDir := ds.convertToAbsolutePath(projectDir)
		if ds.isGitRepositorySafe(ctx, absoluteProjectDir) && !ds.isBareRepositorySafe(ctx, absoluteProjectDir) {
			paths = append(paths, projectDir) // Keep relative path for consistency
		}

		// Then, look for worktree directories within each project directory
		worktreeEntries, err := os.ReadDir(absoluteProjectDir)
		if err != nil {
			continue // Skip on error
		}

		for _, worktreeEntry := range worktreeEntries {
			if !worktreeEntry.IsDir() {
				continue
			}

			worktreePath := filepath.Join(projectDir, worktreeEntry.Name())

			// Check if it's a git repository (worktree)
			absoluteWorktreePath := ds.convertToAbsolutePath(worktreePath)
			if ds.isGitRepositorySafe(ctx, absoluteWorktreePath) && !ds.isBareRepositorySafe(ctx, absoluteWorktreePath) {
				paths = append(paths, worktreePath) // Keep relative path for consistency
			}
		}
	}

	return paths, nil
}

// analyzePathsConcurrently processes multiple paths concurrently using worker pools
func (ds *DiscoveryService) analyzePathsConcurrently(ctx context.Context, paths []string) ([]*domain.Worktree, error) {
	pathsChan := make(chan string, len(paths))
	resultsChan := make(chan *domain.Worktree, len(paths))
	errorsChan := make(chan error, len(paths))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < ds.concurrency; i++ {
		wg.Add(1)
		go ds.workerAnalyze(ctx, pathsChan, resultsChan, errorsChan, &wg)
	}

	// Send paths to workers
	for _, path := range paths {
		pathsChan <- path
	}
	close(pathsChan)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	worktrees, errors := ds.collectResults(resultsChan, errorsChan, len(paths))

	// Return error if too many failures
	if len(errors) > 0 && float64(len(errors))/float64(len(paths)) > failureThreshold {
		return nil, fmt.Errorf("too many failures during discovery (%d/%d failed): %w", len(errors), len(paths), errors[0])
	}

	return worktrees, nil
}

// workerAnalyze is a worker function for concurrent path analysis
func (ds *DiscoveryService) workerAnalyze(ctx context.Context, paths <-chan string, results chan<- *domain.Worktree, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range paths {
		worktree, err := ds.AnalyzeWorktree(ctx, path)
		if err != nil {
			errors <- err
		} else {
			results <- worktree
		}
	}
}

// collectResults collects results and errors from channels until both are closed
func (ds *DiscoveryService) collectResults(resultsChan <-chan *domain.Worktree, errorsChan <-chan error, _ int) ([]*domain.Worktree, []error) {
	var worktrees []*domain.Worktree
	var errors []error

	for {
		select {
		case worktree, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else if worktree != nil {
				worktrees = append(worktrees, worktree)
			}
		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else if err != nil {
				errors = append(errors, err)
			}
		}

		if resultsChan == nil && errorsChan == nil {
			break
		}
	}

	return worktrees, errors
}

// convertToWorktree converts a domain.WorktreeInfo to domain.Worktree
func (ds *DiscoveryService) convertToWorktree(info *domain.WorktreeInfo) (*domain.Worktree, error) {
	worktree, err := domain.NewWorktree(info.Path, info.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w", err)
	}

	// Set additional properties
	if err := worktree.SetCommit(info.Commit); err != nil {
		return nil, fmt.Errorf("failed to set commit: %w", err)
	}

	// Set status without using UpdateStatus to avoid overriding LastUpdated
	if info.Clean {
		worktree.Status = domain.StatusClean
	} else {
		worktree.Status = domain.StatusDirty
	}

	// Set the LastUpdated to the commit timestamp instead of current time
	worktree.LastUpdated = info.CommitTime

	return worktree, nil
}

// getCachedResult retrieves a cached discovery result
func (ds *DiscoveryService) getCachedResult(path string) *domain.Worktree {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result, exists := ds.cache[path]
	if !exists {
		return nil
	}

	// Check if cache is stale
	if time.Since(result.timestamp) > cacheExpiryTime {
		return nil
	}

	return result.worktree
}

// cacheResult caches a discovery result
func (ds *DiscoveryService) cacheResult(path string, worktree *domain.Worktree) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.cache[path] = &discoveryResult{
		worktree:  worktree,
		timestamp: time.Now(),
	}
}

// ClearCache clears the discovery cache
func (ds *DiscoveryService) ClearCache() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.cache = make(map[string]*discoveryResult)
}

// convertToAbsolutePath converts a relative FileSystem path to an absolute path for Git client
func (ds *DiscoveryService) convertToAbsolutePath(relativePath string) string {
	// If the path is already absolute, return it as-is
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	// If the path is ".", use the WorkspacesPath
	if relativePath == "." {
		if ds.config.WorkspacesPath != "" {
			return ds.config.WorkspacesPath
		}
		// Fallback to ProjectsPath if Workspace is not set
		if ds.config.ProjectsPath != "" {
			return ds.config.ProjectsPath
		}
		return relativePath
	}

	// If the path starts with "Projects", use the configured ProjectsPath
	if strings.HasPrefix(relativePath, "Projects") {
		// Remove "Projects" prefix and join with ProjectsPath
		rest := strings.TrimPrefix(relativePath, "Projects")
		if rest == "" {
			return ds.config.ProjectsPath
		}
		return filepath.Join(ds.config.ProjectsPath, rest)
	}

	// If the path starts with "Workspaces", use the configured WorkspacesPath
	if strings.HasPrefix(relativePath, "Workspaces") {
		// Remove "Workspaces" prefix and join with WorkspacesPath
		rest := strings.TrimPrefix(relativePath, "Workspaces")
		if rest == "" {
			return ds.config.WorkspacesPath
		}
		return filepath.Join(ds.config.WorkspacesPath, rest)
	}

	// For other cases (like relative paths within the workspace),
	// join with the WorkspacesPath
	if ds.config.WorkspacesPath != "" {
		return filepath.Join(ds.config.WorkspacesPath, relativePath)
	}

	// Fallback: assume it's already absolute or handle as needed
	return relativePath
}
