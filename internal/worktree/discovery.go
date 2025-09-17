package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/pkg/types"
)

const (
	defaultConcurrency = 4
	maxConcurrency     = 16
	cacheExpiryTime    = 5 * time.Minute
)

// DiscoveryService handles the discovery and analysis of worktrees and projects
type DiscoveryService struct {
	gitClient   types.GitClient
	concurrency int
	mu          sync.RWMutex
	cache       map[string]*discoveryResult
}

type discoveryResult struct {
	worktree  *domain.Worktree
	timestamp time.Time
}

// NewDiscoveryService creates a new DiscoveryService instance
func NewDiscoveryService(gitClient types.GitClient) *DiscoveryService {
	return &DiscoveryService{
		gitClient:   gitClient,
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

// DiscoverWorktrees discovers all worktrees in a workspace directory using concurrent processing
func (ds *DiscoveryService) DiscoverWorktrees(workspacePath string) ([]*domain.Worktree, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	// Check if workspace path exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace path does not exist: %s", workspacePath)
	}

	// Find all potential worktree directories
	paths, err := ds.findPotentialWorktreePaths(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	if len(paths) == 0 {
		return []*domain.Worktree{}, nil
	}

	// Process paths concurrently
	return ds.analyzePathsConcurrently(paths)
}

// AnalyzeWorktree analyzes a single worktree path and returns detailed information
func (ds *DiscoveryService) AnalyzeWorktree(path string) (*domain.Worktree, error) {
	if path == "" {
		return nil, fmt.Errorf("worktree path cannot be empty")
	}

	// Check cache first
	if cached := ds.getCachedResult(path); cached != nil {
		return cached, nil
	}

	// Get worktree status from git client
	status, err := ds.gitClient.GetWorktreeStatus(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree status for %s: %w", path, err)
	}

	// Convert to domain model
	worktree, err := ds.convertToWorktree(status)
	if err != nil {
		return nil, fmt.Errorf("failed to convert worktree info: %w", err)
	}

	// Cache the result
	ds.cacheResult(path, worktree)

	return worktree, nil
}

// DiscoverProjects finds all Git repositories (projects) in the workspace directory
func (ds *DiscoveryService) DiscoverProjects(workspacePath string) ([]*domain.Project, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	// Check if workspace path exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace path does not exist: %s", workspacePath)
	}

	// Find all directories in workspace
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	var projects []*domain.Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(workspacePath, entry.Name())

		// Check if it's a main git repository (not a worktree)
		isMainRepo, err := ds.gitClient.IsMainRepository(projectPath)
		if err != nil {
			// Log error but continue with other directories
			continue
		}

		if !isMainRepo {
			continue
		}

		// Create project and discover its worktrees
		project, err := domain.NewProject(entry.Name(), projectPath)
		if err != nil {
			continue
		}

		// Get worktrees for this project
		worktreeInfos, err := ds.gitClient.ListWorktrees(projectPath)
		if err != nil {
			continue
		}

		// Convert and add worktrees to project
		for _, wtInfo := range worktreeInfos {
			worktree, err := ds.convertToWorktree(wtInfo)
			if err != nil {
				continue
			}
			if err := project.AddWorktree(worktree); err != nil {
				continue
			}
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// findPotentialWorktreePaths scans the workspace for directories that might be worktrees
func (ds *DiscoveryService) findPotentialWorktreePaths(workspacePath string) ([]string, error) {
	var paths []string

	// First, find all project directories (git repositories)
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(workspacePath, entry.Name())

		// Check if it's a git repository
		isRepo, err := ds.gitClient.IsGitRepository(projectPath)
		if err != nil {
			continue // Skip on error
		}

		if isRepo {
			// Get all worktrees for this project
			worktreeInfos, err := ds.gitClient.ListWorktrees(projectPath)
			if err != nil {
				continue // Skip on error
			}

			// Add all worktree paths
			for _, wtInfo := range worktreeInfos {
				paths = append(paths, wtInfo.Path)
			}
		}
	}

	return paths, nil
}

// analyzePathsConcurrently processes multiple paths concurrently using worker pools
func (ds *DiscoveryService) analyzePathsConcurrently(paths []string) ([]*domain.Worktree, error) {
	pathsChan := make(chan string, len(paths))
	resultsChan := make(chan *domain.Worktree, len(paths))
	errorsChan := make(chan error, len(paths))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < ds.concurrency; i++ {
		wg.Add(1)
		go ds.workerAnalyze(pathsChan, resultsChan, errorsChan, &wg)
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

	// Collect results
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

	// Return error if too many failures (more than half failed)
	if len(errors) > 0 && len(errors) > len(paths)/2 {
		return nil, fmt.Errorf("too many failures during discovery (%d/%d failed): %v", len(errors), len(paths), errors[0])
	}

	return worktrees, nil
}

// workerAnalyze is a worker function for concurrent path analysis
func (ds *DiscoveryService) workerAnalyze(paths <-chan string, results chan<- *domain.Worktree, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range paths {
		worktree, err := ds.AnalyzeWorktree(path)
		if err != nil {
			errors <- err
		} else {
			results <- worktree
		}
	}
}

// convertToWorktree converts a types.WorktreeInfo to domain.Worktree
func (ds *DiscoveryService) convertToWorktree(info *types.WorktreeInfo) (*domain.Worktree, error) {
	worktree, err := domain.NewWorktree(info.Path, info.Branch)
	if err != nil {
		return nil, err
	}

	// Set additional properties
	if err := worktree.SetCommit(info.Commit); err != nil {
		return nil, fmt.Errorf("failed to set commit: %w", err)
	}

	if info.Clean {
		if err := worktree.UpdateStatus(domain.StatusClean); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
	} else {
		if err := worktree.UpdateStatus(domain.StatusDirty); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
	}

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
