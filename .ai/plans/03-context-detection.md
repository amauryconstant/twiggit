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

### Context-Aware Identifier Resolution (from design.md lines 179-194)
> **Identifier Resolution Rules**
> - **From Project Context**: `<branch>` WILL resolve to worktree of current project, `<project>` WILL resolve to different project's main directory, `main` WILL resolve to current project's main directory
> - **From Worktree Context**: `<branch>` WILL resolve to different worktree of same project, `main` WILL resolve to main project directory, `<project>` WILL resolve to different project's main directory
> - **From Outside Git Context**: `<project>` WILL resolve to project's main directory, `<project>/<branch>` WILL resolve to cross-project worktree

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
**File**: `internal/domain/context.go`

```go
// ContextDetector detects the current git context
type ContextDetector interface {
    // DetectContext detects the context from the given directory
    DetectContext(dir string) (*Context, error)
}
```

#### 1.3 Define ContextResolver Interface and Types
**File**: `internal/domain/context.go`

```go
// PathType represents the type of resolved path
type PathType int

const (
    PathTypeProject PathType = iota
    PathTypeWorktree
    PathTypeInvalid
)

// ResolutionResult represents the result of identifier resolution
type ResolutionResult struct {
    ResolvedPath   string
    Type           PathType
    ProjectName    string
    BranchName     string
    Explanation    string
}

// ResolutionSuggestion represents a completion suggestion
type ResolutionSuggestion struct {
    Text           string
    Description    string
    Type           PathType
    ProjectName    string
    BranchName     string
}

// ContextResolver resolves target identifiers based on current context
type ContextResolver interface {
    // ResolveIdentifier resolves target identifier based on context
    ResolveIdentifier(ctx *Context, identifier string) (*ResolutionResult, error)
    
    // GetResolutionSuggestions provides completion suggestions
    GetResolutionSuggestions(ctx *Context, partial string) ([]*ResolutionSuggestion, error)
}
```

### Phase 2: Path-Based Detection Implementation

#### 2.1 Implement ContextDetector
**File**: `internal/infrastructure/context_detector.go`

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "twiggit/internal/domain"
)

type contextDetector struct {
    config *domain.Config
}

func NewContextDetector(cfg *domain.Config) ContextDetector {
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
    worktreeDir := filepath.Clean(cd.config.WorktreesDirectory)
    
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

### Phase 3: ContextResolver Implementation

#### 3.1 Implement ContextResolver
**File**: `internal/infrastructure/context_resolver.go`

```go
import (
    "fmt"
    "path/filepath"
    "strings"
    
    "twiggit/internal/domain"
)

type contextResolver struct {
    config *domain.Config
}

func NewContextResolver(cfg *domain.Config) domain.ContextResolver {
    return &contextResolver{config: cfg}
}

func (cr *contextResolver) ResolveIdentifier(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
    // Handle empty identifier
    if identifier == "" {
        return nil, domain.NewContextDetectionError("", "empty identifier", nil)
    }
    
    switch ctx.Type {
    case domain.ContextProject:
        return cr.resolveFromProjectContext(ctx, identifier)
    case domain.ContextWorktree:
        return cr.resolveFromWorktreeContext(ctx, identifier)
    case domain.ContextOutsideGit:
        return cr.resolveFromOutsideGitContext(ctx, identifier)
    default:
        return &domain.ResolutionResult{
            Type:        domain.PathTypeInvalid,
            Explanation: fmt.Sprintf("Cannot resolve identifier '%s' from unknown context", identifier),
        }, nil
    }
}

func (cr *contextResolver) GetResolutionSuggestions(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
    var suggestions []*domain.ResolutionSuggestion
    
    switch ctx.Type {
    case domain.ContextProject:
        suggestions = append(suggestions, cr.getProjectContextSuggestions(ctx, partial)...)
    case domain.ContextWorktree:
        suggestions = append(suggestions, cr.getWorktreeContextSuggestions(ctx, partial)...)
    case domain.ContextOutsideGit:
        suggestions = append(suggestions, cr.getOutsideGitContextSuggestions(ctx, partial)...)
    }
    
    return suggestions, nil
}
```

#### 3.2 Project Context Resolution
**File**: `internal/infrastructure/context_resolver.go` (continued)

```go
func (cr *contextResolver) resolveFromProjectContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
    // Handle special case: "main" resolves to project root
    if identifier == "main" {
        return &domain.ResolutionResult{
            ResolvedPath: ctx.Path,
            Type:         domain.PathTypeProject,
            ProjectName:  ctx.ProjectName,
            Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
        }, nil
    }
    
    // Check if identifier contains "/" (cross-project reference)
    if strings.Contains(identifier, "/") {
        return cr.resolveCrossProjectReference(identifier)
    }
    
    // Resolve as branch name (worktree of current project)
    worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)
    
    return &domain.ResolutionResult{
        ResolvedPath: worktreePath,
        Type:         domain.PathTypeWorktree,
        ProjectName:  ctx.ProjectName,
        BranchName:   identifier,
        Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
    }, nil
}

func (cr *contextResolver) getProjectContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
    var suggestions []*domain.ResolutionSuggestion
    
    // Always suggest "main" for project context
    if strings.HasPrefix("main", partial) {
        suggestions = append(suggestions, &domain.ResolutionSuggestion{
            Text:        "main",
            Description: "Project root directory",
            Type:        domain.PathTypeProject,
            ProjectName: ctx.ProjectName,
        })
    }
    
    // TODO: Add actual worktree discovery when git operations are available
    // For now, provide basic branch name suggestions
    
    return suggestions
}
```

#### 3.3 Worktree Context Resolution
**File**: `internal/infrastructure/context_resolver.go` (continued)

```go
func (cr *contextResolver) resolveFromWorktreeContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
    // Handle special case: "main" resolves to project root
    if identifier == "main" {
        projectPath := filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
        return &domain.ResolutionResult{
            ResolvedPath: projectPath,
            Type:         domain.PathTypeProject,
            ProjectName:  ctx.ProjectName,
            Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
        }, nil
    }
    
    // Check if identifier contains "/" (cross-project reference)
    if strings.Contains(identifier, "/") {
        return cr.resolveCrossProjectReference(identifier)
    }
    
    // Resolve as different worktree of same project
    worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)
    
    return &domain.ResolutionResult{
        ResolvedPath: worktreePath,
        Type:         domain.PathTypeWorktree,
        ProjectName:  ctx.ProjectName,
        BranchName:   identifier,
        Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
    }, nil
}

func (cr *contextResolver) getWorktreeContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
    var suggestions []*domain.ResolutionSuggestion
    
    // Always suggest "main" for worktree context
    if strings.HasPrefix("main", partial) {
        suggestions = append(suggestions, &domain.ResolutionSuggestion{
            Text:        "main",
            Description: "Project root directory",
            Type:        domain.PathTypeProject,
            ProjectName: ctx.ProjectName,
        })
    }
    
    // TODO: Add actual worktree discovery when git operations are available
    
    return suggestions
}
```

#### 3.4 Outside Git Context Resolution
**File**: `internal/infrastructure/context_resolver.go` (continued)

```go
func (cr *contextResolver) resolveFromOutsideGitContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
    // Check if identifier contains "/" (project/branch format)
    if strings.Contains(identifier, "/") {
        return cr.resolveCrossProjectReference(identifier)
    }
    
    // Resolve as project name
    projectPath := filepath.Join(cr.config.ProjectsDirectory, identifier)
    
    return &domain.ResolutionResult{
        ResolvedPath: projectPath,
        Type:         domain.PathTypeProject,
        ProjectName:  identifier,
        Explanation:  fmt.Sprintf("Resolved '%s' to project directory", identifier),
    }, nil
}

func (cr *contextResolver) getOutsideGitContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
    var suggestions []*domain.ResolutionSuggestion
    
    // TODO: Add actual project discovery when git operations are available
    // For now, provide basic suggestions
    
    return suggestions
}
```

#### 3.5 Cross-Project Reference Resolution
**File**: `internal/infrastructure/context_resolver.go` (continued)

```go
func (cr *contextResolver) resolveCrossProjectReference(identifier string) (*domain.ResolutionResult, error) {
    parts := strings.Split(identifier, "/")
    if len(parts) != 2 {
        return &domain.ResolutionResult{
            Type:        domain.PathTypeInvalid,
            Explanation: fmt.Sprintf("Invalid cross-project reference format: '%s'. Expected: project/branch", identifier),
        }, nil
    }
    
    projectName := parts[0]
    branchName := parts[1]
    
    // Resolve to worktree of specified project
    worktreePath := filepath.Join(cr.config.WorktreesDirectory, projectName, branchName)
    
    return &domain.ResolutionResult{
        ResolvedPath: worktreePath,
        Type:         domain.PathTypeWorktree,
        ProjectName:  projectName,
        BranchName:   branchName,
        Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, projectName),
    }, nil
}
```

### Phase 4: Cross-Platform Compatibility

#### 4.1 Path Normalization Utilities
**File**: `internal/infrastructure/pathutils.go`

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

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

// isPathUnder checks if target is under base directory
func isPathUnder(base, target string) (bool, error) {
    rel, err := filepath.Rel(base, target)
    if err != nil {
        return false, err
    }
    
    // Check if relative path starts with ".."
    return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
```

#### 4.2 Symlink Handling
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
    // Normalize path and resolve symlinks
    normalizedDir, err := pathutils.NormalizePath(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to normalize directory: %w", err)
    }

    // Continue with detection logic...
}
```

### Phase 5: Performance Optimization

#### 5.1 Context Caching
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
import (
    "sync"
    
    "twiggit/internal/domain"
)

type contextDetector struct {
    config    *domain.Config
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

#### 5.2 Early Exit Optimizations
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
func (cd *contextDetector) detectWorktreeContext(dir string) *Context {
    // Quick check: if not under worktrees dir, exit early
    if !strings.HasPrefix(dir, cd.config.WorktreesDirectory+string(filepath.Separator)) {
        return nil
    }
    
    // Continue with full detection...
}
```

### Phase 6: Comprehensive Testing

#### 6.1 Unit Tests Structure
**File**: `internal/domain/context_test.go`

```go
import (
    "os"
    "path/filepath"
    "runtime"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "twiggit/internal/domain"
)

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
            
config := &domain.Config{
    WorktreesDirectory: filepath.Join(os.Getenv("HOME"), "Worktrees"),
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

#### 6.2 Integration Tests with Real Git
**File**: `internal/infrastructure/context_detector_integration_test.go`

```go
import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "twiggit/internal/domain"
)

func TestContextDetector_Integration(t *testing.T) {
    tests := []struct {
        name        string
        setupFunc   func(*testing.T, *domain.Config) string
        expectedType ContextType
        expectedProj string
        expectedBranch string
    }{
        {
            name: "real git repository detection",
            setupFunc: func(t *testing.T, config *domain.Config) string {
                repoDir := filepath.Join(config.ProjectsDirectory, "test-repo")
                require.NoError(t, os.MkdirAll(repoDir, 0755))
                
                // Use git to initialize repository
                cmd := exec.Command("git", "init")
                cmd.Dir = repoDir
                require.NoError(t, cmd.Run())
                
                return repoDir
            },
            expectedType: ContextProject,
            expectedProj: "test-repo",
        },
        {
            name: "real git worktree detection",
            setupFunc: func(t *testing.T, config *domain.Config) string {
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
            expectedType:   ContextWorktree,
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
            
            detector := NewContextDetector(config)
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
```

#### 6.3 Cross-Platform Tests
**File**: `internal/infrastructure/context_detector_platform_test.go`

```go
import (
    "runtime"
    "testing"
    
    "github.com/stretchr/testify/require"
    
    "twiggit/internal/domain"
)

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
    config := &domain.Config{
        ProjectsDirectory:  `C:\Users\Test\Projects`,
        WorktreesDirectory: `C:\Users\Test\Worktrees`,
    }
    
    detector := NewContextDetector(config)
    require.NotNil(t, detector)
    
    // Add Windows-specific path tests here
}

func testUnixPaths(t *testing.T) {
    // Test Unix path handling, symlinks, etc.
    config := &domain.Config{
        ProjectsDirectory:  "/home/test/Projects",
        WorktreesDirectory: "/home/test/Worktrees",
    }
    
    detector := NewContextDetector(config)
    require.NotNil(t, detector)
    
    // Add Unix-specific path tests here
}
```

#### 6.4 ContextResolver Tests
**File**: `internal/domain/context_resolver_test.go`

```go
import (
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "twiggit/internal/domain"
)

func TestContextResolver_ResolveIdentifier(t *testing.T) {
    tests := []struct {
        name           string
        context        *domain.Context
        identifier     string
        expectedType   domain.PathType
        expectedProj   string
        expectedBranch string
        expectedPath   string
        expectError    bool
    }{
        {
            name: "project context - main to project root",
            context: &domain.Context{
                Type:        domain.ContextProject,
                ProjectName: "test-project",
                Path:        "/home/user/Projects/test-project",
            },
            identifier:     "main",
            expectedType:   domain.PathTypeProject,
            expectedProj:   "test-project",
            expectedPath:   "/home/user/Projects/test-project",
        },
        {
            name: "project context - branch to worktree",
            context: &domain.Context{
                Type:        domain.ContextProject,
                ProjectName: "test-project",
                Path:        "/home/user/Projects/test-project",
            },
            identifier:     "feature-branch",
            expectedType:   domain.PathTypeWorktree,
            expectedProj:   "test-project",
            expectedBranch: "feature-branch",
            expectedPath:   "/home/user/Worktrees/test-project/feature-branch",
        },
        {
            name: "worktree context - main to project root",
            context: &domain.Context{
                Type:        domain.ContextWorktree,
                ProjectName: "test-project",
                BranchName:  "current-branch",
                Path:        "/home/user/Worktrees/test-project/current-branch",
            },
            identifier:     "main",
            expectedType:   domain.PathTypeProject,
            expectedProj:   "test-project",
            expectedPath:   "/home/user/Projects/test-project",
        },
        {
            name: "outside git context - project to project directory",
            context: &domain.Context{
                Type: domain.ContextOutsideGit,
                Path: "/home/user",
            },
            identifier:     "test-project",
            expectedType:   domain.PathTypeProject,
            expectedProj:   "test-project",
            expectedPath:   "/home/user/Projects/test-project",
        },
        {
            name: "cross-project reference",
            context: &domain.Context{
                Type: domain.ContextOutsideGit,
                Path: "/home/user",
            },
            identifier:     "other-project/feature-branch",
            expectedType:   domain.PathTypeWorktree,
            expectedProj:   "other-project",
            expectedBranch: "feature-branch",
            expectedPath:   "/home/user/Worktrees/other-project/feature-branch",
        },
        {
            name: "empty identifier",
            context: &domain.Context{
                Type: domain.ContextOutsideGit,
                Path: "/home/user",
            },
            identifier:  "",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &domain.Config{
                ProjectsDirectory:  "/home/user/Projects",
                WorktreesDirectory: "/home/user/Worktrees",
            }
            
            resolver := NewContextResolver(config)
            result, err := resolver.ResolveIdentifier(tt.context, tt.identifier)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expectedType, result.Type)
            assert.Equal(t, tt.expectedProj, result.ProjectName)
            if tt.expectedBranch != "" {
                assert.Equal(t, tt.expectedBranch, result.BranchName)
            }
            assert.Equal(t, tt.expectedPath, result.ResolvedPath)
            assert.NotEmpty(t, result.Explanation)
        })
    }
}

func TestContextResolver_GetResolutionSuggestions(t *testing.T) {
    tests := []struct {
        name           string
        context        *domain.Context
        partial        string
        expectedCount  int
        expectedTexts  []string
    }{
        {
            name: "project context - partial 'm'",
            context: &domain.Context{
                Type:        domain.ContextProject,
                ProjectName: "test-project",
                Path:        "/home/user/Projects/test-project",
            },
            partial:       "m",
            expectedCount: 1,
            expectedTexts: []string{"main"},
        },
        {
            name: "worktree context - partial 'main'",
            context: &domain.Context{
                Type:        domain.ContextWorktree,
                ProjectName: "test-project",
                BranchName:  "current-branch",
                Path:        "/home/user/Worktrees/test-project/current-branch",
            },
            partial:       "main",
            expectedCount: 1,
            expectedTexts: []string{"main"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := &domain.Config{
                ProjectsDirectory:  "/home/user/Projects",
                WorktreesDirectory: "/home/user/Worktrees",
            }
            
            resolver := NewContextResolver(config)
            suggestions, err := resolver.GetResolutionSuggestions(tt.context, tt.partial)
            
            require.NoError(t, err)
            assert.Len(t, suggestions, tt.expectedCount)
            
            if len(tt.expectedTexts) > 0 {
                suggestionTexts := make([]string, len(suggestions))
                for i, suggestion := range suggestions {
                    suggestionTexts[i] = suggestion.Text
                }
                assert.Equal(t, tt.expectedTexts, suggestionTexts)
            }
        })
    }
}
```

### Phase 7: Error Handling and Edge Cases

#### 7.1 Error Types
**File**: `internal/domain/errors.go`

```go
import (
    "fmt"
)

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

#### 7.2 Edge Case Handling
**File**: `internal/infrastructure/context_detector.go` (enhanced)

```go
import (
    "os"
    
    "twiggit/internal/domain"
)

func (cd *contextDetector) DetectContext(dir string) (*Context, error) {
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
    
    // Continue with detection...
}
```

### Phase 8: Integration with Commands

#### 8.1 Context Service
**File**: `internal/domain/context_service.go`

```go
import (
    "fmt"
    "os"
    
    "twiggit/internal/domain"
)

// ContextService provides context-aware operations
type ContextService struct {
    detector  ContextDetector
    resolver  ContextResolver
    config    *domain.Config
}

func NewContextService(detector ContextDetector, resolver ContextResolver, cfg *domain.Config) *ContextService {
    return &ContextService{
        detector: detector,
        resolver: resolver,
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

// ResolveIdentifier resolves identifier based on current context
func (cs *ContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
    ctx, err := cs.GetCurrentContext()
    if err != nil {
        return nil, fmt.Errorf("failed to get current context: %w", err)
    }
    
    return cs.resolver.ResolveIdentifier(ctx, identifier)
}

// ResolveIdentifierFromContext resolves identifier based on specified context
func (cs *ContextService) ResolveIdentifierFromContext(ctx *Context, identifier string) (*domain.ResolutionResult, error) {
    return cs.resolver.ResolveIdentifier(ctx, identifier)
}

// GetCompletionSuggestions provides completion suggestions based on current context
func (cs *ContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
    ctx, err := cs.GetCurrentContext()
    if err != nil {
        return nil, fmt.Errorf("failed to get current context: %w", err)
    }
    
    return cs.resolver.GetResolutionSuggestions(ctx, partial)
}
```

## Implementation Checklist

### Core Implementation
- [ ] Define ContextType enum and Context struct
- [ ] Implement ContextDetector interface in domain
- [ ] Define ContextResolver interface and supporting types
- [ ] Create contextDetector implementation with *domain.Config
- [ ] Create contextResolver implementation with context-aware resolution
- [ ] Implement worktree pattern detection using WorktreesDirectory
- [ ] Implement project detection with .git traversal
- [ ] Add path normalization utilities in pathutils package
- [ ] Implement symlink handling
- [ ] Add context caching for performance (Phase 5)

### Cross-Platform Support
- [ ] Use filepath package for all path operations
- [ ] Test on Windows, macOS, and Linux
- [ ] Handle platform-specific edge cases
- [ ] Implement proper path separator handling

### Testing
- [ ] Unit tests for all detection methods using Testify
- [ ] Unit tests for ContextResolver identifier resolution
- [ ] Integration tests with real git repositories using Testify
- [ ] Cross-platform compatibility tests
- [ ] Context-aware resolution tests for all context types
- [ ] Edge case and error handling tests
- [ ] Performance benchmarks
- [ ] Add import sections to all test files

### Error Handling
- [ ] Define context-specific error types in domain/errors.go
- [ ] Handle permission errors gracefully
- [ ] Provide actionable error messages
- [ ] Handle broken git repositories

### Integration
- [ ] Create ContextService for command integration
- [ ] Integrate ContextResolver with ContextService
- [ ] Integrate with existing configuration system
- [ ] Add context detection to CLI commands
- [ ] Add identifier resolution to CLI commands
- [ ] Update command behavior based on context

## Success Criteria

1. **Correct Context Detection**: System correctly identifies all three context types in various scenarios
2. **Identifier Resolution**: Correctly resolves all identifier formats from all contexts
3. **Context-Aware Suggestions**: Provides appropriate completion suggestions based on context
4. **Cross-Platform Compatibility**: Works consistently on Windows, macOS, and Linux
5. **Performance**: Context detection and resolution completes in <50ms for typical scenarios
6. **Test Coverage**: >90% test coverage for context detection and resolution logic
7. **Error Handling**: Graceful handling of edge cases with clear error messages
8. **Integration**: Seamless integration with existing command structure

## Dependencies

- Go standard library: `path/filepath`, `os`, `strings`, `fmt`
- Configuration system (already implemented)
- Testing framework: Testify for unit and integration tests, Ginkgo/Gomega for E2E tests

## Timeline

- **Phase 1-2**: Core interface and detection implementation (2-3 days)
- **Phase 3**: ContextResolver implementation (2-3 days)
- **Phase 4**: Cross-platform support and optimization (1-2 days)
- **Phase 5**: Performance optimization (1-2 days)
- **Phase 6**: Comprehensive testing (2-3 days)
- **Phase 7**: Error handling and edge cases (1-2 days)
- **Phase 8**: Integration with commands (1-2 days)

**Total Estimated Time**: 10-17 days