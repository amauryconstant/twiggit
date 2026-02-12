# Go Performance Patterns

Go-specific patterns for caching, slice operations, path validation, and filesystem operations in codebase auditing.

## Cache Patterns

### Patterns

- **Bounded caches**: Use LRU or size-limited caches instead of unbounded maps
- **Cache invalidation**: Implement cleanup mechanisms for cached data
- **Cache keys**: Use stable, unique keys that won't cause collisions

### Common Patterns

```go
// Bounded LRU cache
import "github.com/hashicorp/golang-lru/v2"

type Service struct {
    cache *lru.Cache[string, *Repository] // Bounded cache
}

func NewService(cacheSize int) *Service {
    cache, _ := lru.New[string, *Repository](cacheSize)
    return &Service{cache: cache}
}

func (s *Service) GetRepository(key string) (*Repository, error) {
    if val, ok := s.cache.Get(key); ok {
        return val, nil  // Cache hit
    }
    repo, err := s.loadRepository(key)  // Cache miss
    if err != nil {
        return nil, err
    }
    s.cache.Add(key, repo)
    return repo, nil
}
```

### Anti-Patterns

- **Unbounded maps**: Maps that grow indefinitely without eviction
  ```go
  // BAD: Memory leak
  type GoGitClientImpl struct {
      cache map[string]*git.Repository  // Never cleared, grows forever
  }

  // GOOD: Bounded cache
  type GoGitClientImpl struct {
      cache *lru.Cache[string, *git.Repository]  // Auto-evicts when full
  }
  ```
- **Missing cache invalidation**: Stale data in cache never updated
- **Cache without size limit**: OOM risk as data grows

## Slice Operations

### Patterns

- **Pre-allocate capacity**: Use `make([]T, 0, capacity)` when final size is known
- **Avoid append in loops**: Pre-allocate when iterating known collections
- **Slice growth awareness**: Understand Go's slice growth strategy

### Common Patterns

```go
// GOOD: Pre-allocate capacity
func filterItems(items []Item, predicate func(Item) bool) []Item {
    result := make([]Item, 0, len(items))  // Pre-allocate full capacity
    for _, item := range items {
        if predicate(item) {
            result = append(result, item)  // No reallocation
        }
    }
    return result
}

// GOOD: Estimate capacity
func collectResults(ctx context.Context) ([]string, error) {
    // If you can estimate size, use it
    estimatedSize := estimateResultSize(ctx)
    results := make([]string, 0, estimatedSize)
    // ... populate ...
    return results, nil
}
```

### Anti-Patterns

- **No pre-allocation in loops**: Causes repeated allocations
  ```go
  // BAD: Repeated allocations (O(n²) total)
  func filterSuggestions(suggestions []string, partial string) []string {
      result := make([]string, 0)  // No capacity
      for _, suggestion := range suggestions {
          if strings.HasPrefix(suggestion, partial) {
              result = append(result, suggestion)  // Reallocates
          }
      }
      return result
  }

  // GOOD: Pre-allocate
  func filterSuggestions(suggestions []string, partial string) []string {
      result := make([]string, 0, len(suggestions))  // Full capacity
      for _, suggestion := range suggestions {
          if strings.HasPrefix(suggestion, partial) {
              result = append(result, suggestion)  // No reallocation
          }
      }
      return result
  }
  ```

## Path Validation

### Patterns

- **Clean paths before validation**: Use `filepath.Clean()` to normalize paths
- **Resolve symlinks**: Use `filepath.EvalSymlinks()` before validation
- **Use filepath.Join**: Never concatenate paths with string operations
- **Validate after normalization**: Validate cleaned/resolved paths, not raw input

### Common Patterns

```go
// GOOD: Clean and resolve before validation
func validatePathUnder(base, target string) error {
    // Resolve symlinks first
    resolvedBase, err := filepath.EvalSymlinks(base)
    if err != nil {
        return err
    }

    resolvedTarget, err := filepath.EvalSymlinks(target)
    if err != nil {
        return err
    }

    // Clean both paths
    resolvedBase = filepath.Clean(resolvedBase)
    resolvedTarget = filepath.Clean(resolvedTarget)

    // Check if target is under base
    if !strings.HasPrefix(resolvedTarget, resolvedBase) {
        return fmt.Errorf("target path outside base directory")
    }

    return nil
}

// GOOD: Detect traversal via cleaning
func containsPathTraversal(s string) bool {
    cleaned := filepath.Clean(s)
    // If cleaning changed the path, traversal was detected
    if cleaned != s {
        return true
    }
    // Also check for URL-encoded traversal
    return strings.Contains(s, "%2e%2e") || strings.Contains(s, "%2E%2E")
}
```

### Anti-Patterns

- **String-based traversal detection**: Using `strings.Contains(s, "..")` is insufficient
  ```go
  // BAD: Incomplete pattern matching
  func containsPathTraversal(s string) bool {
      return strings.Contains(s, "..") || strings.Contains(s, string(filepath.Separator)+".")
  }
  // Fails to catch: ..//, %2e%2e, etc.

  // GOOD: Clean and compare
  func containsPathTraversal(s string) bool {
      cleaned := filepath.Clean(s)
      return cleaned != s || strings.Contains(s, "%2e%2e") || strings.Contains(s, "%2E%2E")
  }
  ```

- **Not resolving symlinks**: Validation can be bypassed with malicious symlinks
  ```go
  // BAD: Symlink not resolved
  if under, err := IsPathUnder(base, target); err == nil && under {
      // Attacker can create symlink inside base pointing outside
  }

  // GOOD: Resolve symlinks first
  resolvedBase, _ := filepath.EvalSymlinks(base)
  resolvedTarget, _ := filepath.EvalSymlinks(target)
  if under, err := IsPathUnder(resolvedBase, resolvedTarget); err == nil && under {
      // Symlinks resolved, validation accurate
  }
  ```

- **Path concatenation**: Using string operations instead of `filepath.Join`
  ```go
  // BAD: String concatenation
  path := base + "/" + filename  // Wrong separator on Windows

  // GOOD: filepath.Join
  path := filepath.Join(base, filename)  // Cross-platform
  ```

## Filesystem Operations

### Patterns

- **Avoid N+1 queries**: Batch filesystem operations when possible
- **Cache directory traversal results**: Avoid repeated os.Stat calls
- **Use filepath operations**: Cross-platform path handling
- **Single I/O operations**: Parse files instead of traversing directories

### Common Patterns

```go
// GOOD: Parse file directly instead of traversal
func findMainRepoFromWorktree(worktreePath string) string {
    // Read .git file once
    gitFilePath := filepath.Join(worktreePath, ".git")
    content, err := os.ReadFile(gitFilePath)
    if err != nil {
        return worktreePath
    }

    // Parse to find main repo
    contentStr := strings.TrimSpace(string(content))
    if strings.HasPrefix(contentStr, "gitdir:") {
        gitdirPath := strings.TrimPrefix(contentStr, "gitdir:")
        gitdirPath = strings.TrimSpace(gitdirPath)

        // Extract repo path from gitdir path
        if strings.Contains(gitdirPath, "/.git/worktrees/") {
            parts := strings.Split(gitdirPath, "/.git/worktrees/")
            if len(parts) > 0 {
                return filepath.Dir(filepath.Dir(parts[0]))
            }
        }
    }

    return worktreePath
}

// GOOD: Cache repository root detection
type ContextDetector struct {
    config        *domain.Config
    repoRootCache map[string]string  // path → repo root
}

func (cd *ContextDetector) DetectProjectContext(dir string) *domain.Context {
    absDir, _ := filepath.Abs(dir)

    // Check cache first
    if repoRoot, exists := cd.repoRootCache[absDir]; exists {
        return &domain.Context{
            Type:        domain.ContextProject,
            ProjectName: filepath.Base(repoRoot),
            Path:        repoRoot,
        }
    }

    // Traverse filesystem
    currentDir := dir
    for {
        gitPath := filepath.Join(currentDir, ".git")
        if _, err := os.Stat(gitPath); err == nil {
            // Cache the result
            cd.repoRootCache[absDir] = currentDir
            return &domain.Context{
                Type:        domain.ContextProject,
                ProjectName: filepath.Base(currentDir),
                Path:        currentDir,
            }
        }

        parent := filepath.Dir(currentDir)
        if parent == currentDir {
            break
        }
        currentDir = parent
    }

    return nil
}

// GOOD: Use file parsing instead of repeated I/O
func findProjectByWorktreePath(worktreePath, worktreesDir string) string {
    // Parse path directly - single operation
    relPath, err := filepath.Rel(worktreesDir, worktreePath)
    if err != nil {
        return ""
    }

    parts := strings.Split(relPath, string(filepath.Separator))
    if len(parts) == 0 {
        return ""
    }

    // Project name is first component
    return parts[0]
}
```

### Anti-Patterns

- **N+1 filesystem queries**: Repeated os.Stat calls in loops
  ```go
  // BAD: O(n*m) - stat called for each project/worktree
  func findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
      projects, err := s.projectService.ListProjects(ctx)  // O(n)
      for _, project := range projects {
          worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)  // O(m) each!
          // Find matching worktree
      }
  }

  // GOOD: Parse path directly - O(1)
  func findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
      relPath, _ := filepath.Rel(s.config.WorktreesDirectory, worktreePath)
      parts := strings.Split(relPath, string(filepath.Separator))
      projectName := parts[0]
      return s.projectService.DiscoverProject(ctx, projectName, nil)
  }
  ```

- **Repeated directory traversal**: No caching of filesystem operations
  ```go
  // BAD: Traverse filesystem every time
  func detectContext(dir string) *domain.Context {
      currentDir := dir
      for {
          if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
              return &domain.Context{Path: currentDir}
          }
          currentDir = filepath.Dir(currentDir)
          if currentDir == filepath.Dir(currentDir) { break }
      }
  }

  // GOOD: Cache repository roots
  type ContextDetector struct {
      repoRootCache map[string]string
  }

  func (cd *ContextDetector) DetectContext(dir string) *domain.Context {
      absDir, _ := filepath.Abs(dir)
      if repoRoot, exists := cd.repoRootCache[absDir]; exists {
          return &domain.Context{Path: repoRoot}  // Cache hit
      }
      // ... traverse and cache ...
  }
  ```

- **Double I/O operations**: Calling same method twice for same data
  ```go
  // BAD: ListWorktrees called twice
  worktrees, _ := gitService.ListWorktrees(ctx, repoPath)  // First call
  if len(worktrees) > 0 { ... }
  worktrees, _ = gitService.ListWorktrees(ctx, repoPath)  // Second call - wasteful!
  for _, wt := range worktrees { ... }

  // GOOD: Cache result or call once
  worktrees, _ := gitService.ListWorktrees(ctx, repoPath)
  if len(worktrees) > 0 {
      // Use cached worktrees variable
      worktreeMap := make(map[string]bool, len(worktrees))
      for _, wt := range worktrees {
          worktreeMap[wt.Branch] = true
      }
  }
  ```

## Audit-Specific Patterns

### For Performance Audits

- Check for unbounded caches or maps (memory leaks)
- Identify N+1 queries (nested loops with repeated operations)
- Look for missing slice pre-allocation
- Verify filesystem operations are cached
- Check for repeated I/O operations (same method called multiple times)
- Look for O(n²) algorithms where O(n) exists
- Verify path validation uses cleaning and symlink resolution
- Check for double method calls in hot paths
