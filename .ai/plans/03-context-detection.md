# Context Detection System Implementation Plan

## Purpose

Implement the core context detection system that automatically detects user context (project, worktree, outside git) to enable context-aware command behavior. This is a key differentiator feature that provides intelligent git worktree management.

## Context from Design Documents

### Context Types (from design.md lines 213-223)
> **Context Detection Rules**
> 1. **Project folder**: `.git/` directory found in current or parent directories
>    - Directory tree WILL be traversed up until finding `.git/` or reaching filesystem root
>    - First `.git/` found WILL be used (closest to current directory)
>    - Project folder SHALL be distinguished from worktree folder by path structure
> 2. **Worktree folder**: Path matches `$HOME/Worktrees/<project>/<branch>/` pattern  
>    - Exact pattern matching SHALL be used with configurable base directories
>    - Alternative worktree detection patterns MAY be supported in future
>    - Worktree SHALL be validated to contain valid git worktree
>    - Worktree folder SHALL be distinguished from project folder by path structure
> 3. **Outside git**: No `.git/` found and not in worktree pattern

### Context Detection Priority (from design.md lines 225-232)
> **Context Detection Priority**
> - Context detection SHALL be performed before identifier resolution
> - Worktree folder detection SHALL take precedence over project folder detection when both patterns match
> - Context type SHALL be determined using the following priority:
>   1. **Worktree folder** (if path matches worktree pattern and contains valid worktree)
>   2. **Project folder** (if `.git/` found and path doesn't match worktree pattern)
>   3. **Outside git** (neither condition met)
> - Context detection results SHALL be cached for performance during single command execution

### Implementation Requirements (from implementation.md lines 118-127)
> **Core Components**
> - **ContextDetector**: WILL detect current user context (project, worktree, outside git)
> - **ContextResolver**: WILL resolve target identifiers based on current context
> - **Context Types**: WILL include `ContextProject`, `ContextWorktree`, `ContextOutsideGit`, `ContextUnknown`

### Path Detection Requirements (from implementation.md line 247)
> NO regex patterns for path detection - use Go's filepath package for cross-platform compatibility

## Implementation Plan

### Phase 1: Core Interface and Types

#### 1.1 Define Context Types
**File**: `internal/domain/context.go`

```go
// ContextType represents the type of git context
type ContextType int

const (
    ContextUnknown ContextType = iota
    ContextProject
    ContextWorktree
    ContextOutsideGit
)

// String returns the string representation of ContextType
func (c ContextType) String() string {
    switch c {
    case ContextProject:
        return "project"
    case ContextWorktree:
        return "worktree"
    case ContextOutsideGit:
        return "outside-git"
    default:
        return "unknown"
    }
}

// Context represents the detected git context
type Context struct {
    Type        ContextType
    ProjectName string
    BranchName  string // Only for ContextWorktree
    Path        string // Absolute path to context root
    Explanation string // Human-readable explanation of detection
}
```

#### 1.2 Define ContextDetector Interface
**File**: `internal/infrastructure/interfaces.go`

```go
// ContextDetector detects the current git context
type ContextDetector interface {
    // DetectContext detects the context from the given directory
    DetectContext(dir string) (*Context, error)
}
```

#### 1.3 Define Configuration Integration
**File**: `internal/domain/config.go` (extend existing)

```go
// ContextDetectionConfig holds configuration for context detection
type ContextDetectionConfig struct {
    ProjectsDir  string
    WorktreesDir string
}

// GetContextDetectionConfig returns context detection configuration
func (c *Config) GetContextDetectionConfig() *ContextDetectionConfig {
    return &ContextDetectionConfig{
        ProjectsDir:  c.GetString("projects_dir"),
        WorktreesDir: c.GetString("worktrees_dir"),
    }
}
```

### Phase 2: Path-Based Detection Implementation

#### 2.1 Implement ContextDetector
**File**: `internal/infrastructure/context_detector.go`

```go
type contextDetector struct {
    config *config.ContextDetectionConfig
}

func NewContextDetector(cfg *config.ContextDetectionConfig) ContextDetector {
    return &contextDetector{config: cfg}
}

func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
    // Normalize path to absolute
    absDir, err := filepath.Abs(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to get absolute path: %w", err)
    }

    // Priority 1: Check worktree pattern first
    if ctx := cd.detectWorktreeContext(absDir); ctx != nil {
        return ctx, nil
    }

    // Priority 2: Check project context
    if ctx := cd.detectProjectContext(absDir); ctx != nil {
        return ctx, nil
    }

    // Priority 3: Outside git context
    return &Context{
        Type:        ContextOutsideGit,
        Path:        absDir,
        Explanation: "Not in a git repository or worktree",
    }, nil
}
```

#### 2.2 Implement Worktree Detection
**File**: `internal/infrastructure/context_detector.go` (continued)

```go
func (cd *contextDetector) detectWorktreeContext(dir string) *Context {
    // Normalize worktree directory
    worktreeDir := filepath.Clean(cd.config.WorktreesDir)
    
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
    
    // Validate that we have exactly 2 parts (project/branch)
    if len(parts) > 2 {
        // Check if we're in a subdirectory of the worktree
        // Still consider this a worktree context
        branchName = parts[1]
    }
    
    // Validate this is a valid git worktree
    if !cd.isValidGitWorktree(dir) {
        return nil
    }
    
    return &Context{
        Type:        ContextWorktree,
        ProjectName: projectName,
        BranchName:  branchName,
        Path:        dir,
        Explanation: fmt.Sprintf("In worktree for project '%s' on branch '%s'", projectName, branchName),
    }
}
```

#### 2.3 Implement Project Detection
**File**: `internal/infrastructure/context_detector.go` (continued)

```go
func (cd *contextDetector) detectProjectContext(dir string) *Context {
    currentDir := dir
    
    // Traverse up directory tree looking for .git
    for {
        gitPath := filepath.Join(currentDir, ".git")
        if _, err := os.Stat(gitPath); err == nil {
            // Found .git directory
            projectName := cd.extractProjectName(currentDir)
            
            return &Context{
                Type:        ContextProject,
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
```

#### 2.4 Helper Methods
**File**: `internal/infrastructure/context_detector.go` (continued)

```go
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
```

### Phase 3: Cross-Platform Compatibility

#### 3.1 Path Normalization Utilities
**File**: `internal/infrastructure/path_utils.go`

```go
// NormalizePath normalizes a path for cross-platform compatibility
func NormalizePath(path string) (string, error) {
    // Clean the path
    cleaned := filepath.Clean(path)
    
    // Convert to absolute path
    abs, err := filepath.Abs(cleaned)
    if err != nil {
        return "", fmt.Errorf("failed to normalize path: %w", err)
    }
    
    // Resolve symlinks
    resolved, err := filepath.EvalSymlinks(abs)
    if err != nil {
        // If symlink resolution fails, use absolute path
        return abs, nil
    }
    
    return resolved, nil
}

// IsPathUnder checks if target is under base directory
func IsPathUnder(base, target string) (bool, error) {
    rel, err := filepath.Rel(base, target)
    if err != nil {
        return false, err
    }
    
    // Check if relative path starts with ".."
    return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
```

#### 3.2 Symlink Handling
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
    // Normalize path and resolve symlinks
    normalizedDir, err := path_utils.NormalizePath(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to normalize directory: %w", err)
    }

    // Cache the normalized path for performance
    cd.cachedPath = normalizedDir
    
    // Continue with detection logic...
}
```

### Phase 4: Performance Optimization

#### 4.1 Context Caching
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
type contextDetector struct {
    config    *config.ContextDetectionConfig
    cache     map[string]*Context
    cacheMu   sync.RWMutex
}

func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
    // Check cache first
    cd.cacheMu.RLock()
    if cached, exists := cd.cache[dir]; exists {
        cd.cacheMu.RUnlock()
        return cached, nil
    }
    cd.cacheMu.RUnlock()

    // Perform detection
    ctx, err := cd.detectContextInternal(dir)
    if err != nil {
        return nil, err
    }

    // Cache result
    cd.cacheMu.Lock()
    if cd.cache == nil {
        cd.cache = make(map[string]*Context)
    }
    cd.cache[dir] = ctx
    cd.cacheMu.Unlock()

    return ctx, nil
}
```

#### 4.2 Early Exit Optimizations
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
func (cd *contextDetector) detectWorktreeContext(dir string) *Context {
    // Quick check: if not under worktrees dir, exit early
    if !strings.HasPrefix(dir, cd.config.WorktreesDir+string(filepath.Separator)) {
        return nil
    }
    
    // Continue with full detection...
}
```

### Phase 5: Comprehensive Testing

#### 5.1 Unit Tests Structure
**File**: `internal/domain/context_test.go`

```go
func TestContextDetector_DetectContext(t *testing.T) {
    tests := []struct {
        name           string
        setupFunc      func(*testing.T) string
        expectedType   ContextType
        expectedProj   string
        expectedBranch string
        expectError    bool
    }{
        {
            name: "project context with .git directory",
            setupFunc: func(t *testing.T) string {
                dir := t.TempDir()
                require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0755))
                return dir
            },
            expectedType: ContextProject,
            expectedProj: "temp",
        },
        {
            name: "worktree context in worktree pattern",
            setupFunc: func(t *testing.T) string {
                worktreeDir := filepath.Join(os.Getenv("HOME"), "Worktrees", "test-project", "feature-branch")
                require.NoError(t, os.MkdirAll(worktreeDir, 0755))
                
                // Create worktree .git file
                gitFile := filepath.Join(worktreeDir, ".git")
                require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /path/to/git/dir"), 0644))
                
                return worktreeDir
            },
            expectedType:   ContextWorktree,
            expectedProj:   "test-project",
            expectedBranch: "feature-branch",
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := tt.setupFunc(t)
            
            config := &config.ContextDetectionConfig{
                WorktreesDir: filepath.Join(os.Getenv("HOME"), "Worktrees"),
            }
            
            detector := NewContextDetector(config)
            ctx, err := detector.DetectContext(dir)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expectedType, ctx.Type)
            assert.Equal(t, tt.expectedProj, ctx.ProjectName)
            assert.Equal(t, tt.expectedBranch, ctx.BranchName)
        })
    }
}
```

#### 5.2 Integration Tests with Real Git
**File**: `internal/infrastructure/context_detector_integration_test.go`

```go
func TestContextDetector_Integration(t *testing.T) {
    // Use Ginkgo/Gomega for BDD-style integration tests
    Describe("Context Detection with Real Git", func() {
        var (
            tempDir    string
            detector   ContextDetector
            config     *config.ContextDetectionConfig
        )
        
        BeforeEach(func() {
            tempDir = t.TempDir()
            config = &config.ContextDetectionConfig{
                ProjectsDir:  filepath.Join(tempDir, "Projects"),
                WorktreesDir: filepath.Join(tempDir, "Worktrees"),
            }
            detector = NewContextDetector(config)
        })
        
        Context("when in a real git repository", func() {
            It("should detect project context", func() {
                // Initialize real git repository
                repoDir := filepath.Join(config.ProjectsDir, "test-repo")
                err := os.MkdirAll(repoDir, 0755)
                Expect(err).ToNot(HaveOccurred())
                
                // Use git to initialize repository
                cmd := exec.Command("git", "init")
                cmd.Dir = repoDir
                err = cmd.Run()
                Expect(err).ToNot(HaveOccurred())
                
                ctx, err := detector.DetectContext(repoDir)
                Expect(err).ToNot(HaveOccurred())
                Expect(ctx.Type).To(Equal(ContextProject))
                Expect(ctx.ProjectName).To(Equal("test-repo"))
            })
        })
        
        Context("when in a git worktree", func() {
            It("should detect worktree context", func() {
                // Setup main repository
                mainRepo := filepath.Join(config.ProjectsDir, "main-repo")
                err := os.MkdirAll(mainRepo, 0755)
                Expect(err).ToNot(HaveOccurred())
                
                // Initialize main repo
                cmd := exec.Command("git", "init")
                cmd.Dir = mainRepo
                err = cmd.Run()
                Expect(err).ToNot(HaveOccurred())
                
                // Create initial commit
                cmd = exec.Command("git", "config", "user.email", "test@example.com")
                cmd.Dir = mainRepo
                err = cmd.Run()
                Expect(err).ToNot(HaveOccurred())
                
                cmd = exec.Command("git", "config", "user.name", "Test User")
                cmd.Dir = mainRepo
                err = cmd.Run()
                Expect(err).ToNot(HaveOccurred())
                
                // Create worktree
                worktreeDir := filepath.Join(config.WorktreesDir, "main-repo", "feature-branch")
                err = os.MkdirAll(filepath.Dir(worktreeDir), 0755)
                Expect(err).ToNot(HaveOccurred())
                
                cmd = exec.Command("git", "worktree", "add", worktreeDir, "-b", "feature-branch")
                cmd.Dir = mainRepo
                err = cmd.Run()
                Expect(err).ToNot(HaveOccurred())
                
                ctx, err := detector.DetectContext(worktreeDir)
                Expect(err).ToNot(HaveOccurred())
                Expect(ctx.Type).To(Equal(ContextWorktree))
                Expect(ctx.ProjectName).To(Equal("main-repo"))
                Expect(ctx.BranchName).To(Equal("feature-branch"))
            })
        })
    })
}
```

#### 5.3 Cross-Platform Tests
**File**: `internal/infrastructure/context_detector_platform_test.go`

```go
func TestContextDetector_CrossPlatform(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Run("windows paths", func(t *testing.T) {
            // Test Windows-specific path handling
            testWindowsPaths(t)
        })
    } else {
        t.Run("unix paths", func(t *testing.T) {
            // Test Unix-specific path handling
            testUnixPaths(t)
        })
    }
}

func testWindowsPaths(t *testing.T) {
    // Test Windows path separators, drive letters, etc.
}

func testUnixPaths(t *testing.T) {
    // Test Unix path handling, symlinks, etc.
}
```

### Phase 6: Error Handling and Edge Cases

#### 6.1 Error Types
**File**: `internal/domain/errors.go`

```go
// ContextDetectionError represents context detection errors
type ContextDetectionError struct {
    Path    string
    Cause   error
    Message string
}

func (e *ContextDetectionError) Error() string {
    return fmt.Sprintf("context detection failed for %s: %s", e.Path, e.Message)
}

func (e *ContextDetectionError) Unwrap() error {
    return e.Cause
}

// NewContextDetectionError creates a new context detection error
func NewContextDetectionError(path, message string, cause error) *ContextDetectionError {
    return &ContextDetectionError{
        Path:    path,
        Cause:   cause,
        Message: message,
    }
}
```

#### 6.2 Edge Case Handling
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
    // Validate input directory
    if dir == "" {
        return nil, NewContextDetectionError("", "empty directory path", nil)
    }
    
    // Check if directory exists
    if _, err := os.Stat(dir); err != nil {
        if os.IsNotExist(err) {
            return nil, NewContextDetectionError(dir, "directory does not exist", err)
        }
        return nil, NewContextDetectionError(dir, "cannot access directory", err)
    }
    
    // Continue with detection...
}
```

### Phase 7: Integration with Commands

#### 7.1 Context Service
**File**: `internal/domain/context_service.go`

```go
// ContextService provides context-aware operations
type ContextService struct {
    detector ContextDetector
    config   *config.Config
}

func NewContextService(detector ContextDetector, cfg *config.Config) *ContextService {
    return &ContextService{
        detector: detector,
        config:   cfg,
    }
}

// GetCurrentContext detects context from current working directory
func (cs *ContextService) GetCurrentContext() (*Context, error) {
    wd, err := os.Getwd()
    if err != nil {
        return nil, fmt.Errorf("failed to get working directory: %w", err)
    }
    
    return cs.detector.DetectContext(wd)
}

// DetectContextFromPath detects context from specified path
func (cs *ContextService) DetectContextFromPath(path string) (*Context, error) {
    return cs.detector.DetectContext(path)
}
```

## Implementation Checklist

### Core Implementation
- [ ] Define ContextType enum and Context struct
- [ ] Implement ContextDetector interface
- [ ] Create contextDetector implementation
- [ ] Implement worktree pattern detection
- [ ] Implement project detection with .git traversal
- [ ] Add path normalization utilities
- [ ] Implement symlink handling
- [ ] Add context caching for performance

### Cross-Platform Support
- [ ] Use filepath package for all path operations
- [ ] Test on Windows, macOS, and Linux
- [ ] Handle platform-specific edge cases
- [ ] Implement proper path separator handling

### Testing
- [ ] Unit tests for all detection methods
- [ ] Integration tests with real git repositories
- [ ] Cross-platform compatibility tests
- [ ] Edge case and error handling tests
- [ ] Performance benchmarks

### Error Handling
- [ ] Define context-specific error types
- [ ] Handle permission errors gracefully
- [ ] Provide actionable error messages
- [ ] Handle broken git repositories

### Integration
- [ ] Create ContextService for command integration
- [ ] Integrate with configuration system
- [ ] Add context detection to CLI commands
- [ ] Update command behavior based on context

## Success Criteria

1. **Correct Context Detection**: System correctly identifies all three context types in various scenarios
2. **Cross-Platform Compatibility**: Works consistently on Windows, macOS, and Linux
3. **Performance**: Context detection completes in <50ms for typical scenarios
4. **Test Coverage**: >90% test coverage for context detection logic
5. **Error Handling**: Graceful handling of edge cases with clear error messages
6. **Integration**: Seamless integration with existing command structure

## Dependencies

- Go standard library: `path/filepath`, `os`, `strings`
- Configuration system (already implemented)
- Testing framework: Testify for unit tests, Ginkgo/Gomega for integration tests

## Timeline

- **Phase 1-2**: Core interface and detection implementation (2-3 days)
- **Phase 3-4**: Cross-platform support and optimization (1-2 days)
- **Phase 5**: Comprehensive testing (2-3 days)
- **Phase 6-7**: Error handling and integration (1-2 days)

**Total Estimated Time**: 6-10 days