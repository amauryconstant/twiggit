# Git-Focused Hybrid Operations & Context Integration Plan

## Overview

This plan implements git-focused hybrid operations while completing the remaining context detection integration points. The approach combines go-git's performance with CLI fallback for critical functionality gaps, and integrates git operations directly into the context detection system for intelligent completion and validation.

**Context**: Foundation, configuration, and context detection layers are 95% complete. This phase completes the remaining 5% and adds comprehensive git operations with intelligent fallback mechanisms.

**Key Challenges**: 
- "go-git limitations: worktree functionality gaps requiring hybrid approach" (technology.md:18)
- Context detection TODO comments requiring git integration (context_resolver.go:97,145,170)
- Missing git-aware caching and validation in context detection

## Architecture Overview

### Hybrid Strategy (from technology.md)

> **go-git usage**: repository discovery, basic git operations, branch management, configuration  
> **CLI fallback**: worktree creation/management, complex merge operations, performance-critical operations

### Implementation Distribution
- **go-git (~70% coverage)**: Repository operations, branch listing, status checks, configuration
- **CLI fallback (~30% coverage)**: Worktree creation/deletion, complex operations, error recovery

## Core Interface Design

### HybridGitClient Interface

```go
// internal/infrastructure/interfaces.go
type HybridGitClient interface {
    // Core worktree operations
    CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error
    DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error
    ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
    
    // Repository operations
    OpenRepository(path string) (GitRepository, error)
    IsRepository(path string) bool
    
    // Branch operations
    ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)
    BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)
    
    // Status and validation
    GetWorktreeStatus(ctx context.Context, worktreePath string) (WorktreeStatus, error)
    ValidateWorktree(ctx context.Context, worktreePath string) error
}

type GitRepository interface {
    Path() string
    Branches() ([]BranchInfo, error)
    Worktrees() ([]WorktreeInfo, error)
    Close() error
}

type WorktreeInfo struct {
    Path     string
    Branch   string
    Commit   string
    Dirty    bool
    Detached bool
}

type BranchInfo struct {
    Name   string
    Remote string
    Head   string
}

type WorktreeStatus struct {
    Clean  bool
    Staged int
    Modified int
    Untracked int
}
```

### Hybrid Error System

```go
// internal/domain/errors.go
type HybridError struct {
    Operation    string
    Implementation string // "go-git" or "cli"
    OriginalErr  error
    Context      map[string]interface{}
}

func (e *HybridError) Error() string {
    return fmt.Sprintf("hybrid git error in %s (%s): %v", 
        e.Operation, e.Implementation, e.OriginalErr)
}

func (e *HybridError) IsFallback() bool {
    return e.Implementation == "cli"
}

// Specific error types for fallback detection
type WorktreeNotSupportedError struct {
    Operation string
    Path      string
}

func (e *WorktreeNotSupportedError) Error() string {
    return fmt.Sprintf("worktree operation not supported by go-git: %s at %s", 
        e.Operation, e.Path)
}
```

## Implementation Components

### 1. GoGitClient Implementation

```go
// internal/infrastructure/gogit_client.go
type GoGitClient struct {
    cache     map[string]*git.Repository
    cacheMu   sync.RWMutex
    metrics   *HybridMetrics
}

func NewGoGitClient(metrics *HybridMetrics) *GoGitClient {
    return &GoGitClient{
        cache:   make(map[string]*git.Repository),
        metrics: metrics,
    }
}

func (c *GoGitClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
    c.metrics.RecordOperation("go-git", "create_worktree_attempt")
    
    // Try go-git implementation first
    repo, err := c.openRepository(repoPath)
    if err != nil {
        return &WorktreeNotSupportedError{
            Operation: "create_worktree",
            Path:      repoPath,
        }
    }
    
    // go-git worktree implementation using storage abstractions
    worktree, err := c.createWorktreeUsingStorage(repo, branchName, sourceBranch, worktreePath)
    if err != nil {
        return &WorktreeNotSupportedError{
            Operation: "create_worktree",
            Path:      repoPath,
        }
    }
    
    c.metrics.RecordOperation("go-git", "create_worktree_success")
    return nil
}

func (c *GoGitClient) createWorktreeUsingStorage(repo *git.Repository, branchName, sourceBranch, worktreePath string) error {
    // Implementation using go-git storage abstractions
    // This is the complex part that requires custom implementation
    
    // 1. Get source branch reference
    sourceRef, err := repo.Storer.Reference(plumbing.ReferenceName("refs/heads/" + sourceBranch))
    if err != nil {
        return fmt.Errorf("source branch not found: %w", err)
    }
    
    // 2. Create worktree directory structure
    if err := os.MkdirAll(worktreePath, 0755); err != nil {
        return fmt.Errorf("failed to create worktree directory: %w", err)
    }
    
    // 3. Initialize worktree git directory
    worktreeGitPath := filepath.Join(worktreePath, ".git")
    if err := c.initializeWorktreeGitDir(worktreeGitPath, repoPath, sourceRef); err != nil {
        return fmt.Errorf("failed to initialize worktree git directory: %w", err)
    }
    
    // 4. Checkout files to worktree
    if err := c.checkoutFilesToWorktree(repo, sourceRef.Hash(), worktreePath); err != nil {
        return fmt.Errorf("failed to checkout files: %w", err)
    }
    
    return nil
}

func (c *GoGitClient) initializeWorktreeGitDir(worktreeGitPath, repoPath string, sourceRef *plumbing.Reference) error {
    // Create .git file pointing to main repo
    gitFileContent := fmt.Sprintf("gitdir: %s/.git/worktrees/%s\n", 
        repoPath, filepath.Base(filepath.Dir(worktreeGitPath)))
    
    return os.WriteFile(worktreeGitPath, []byte(gitFileContent), 0644)
}

func (c *GoGitClient) checkoutFilesToWorktree(repo *git.Repository, commitHash plumbing.Hash, worktreePath string) error {
    commit, err := repo.CommitObject(commitHash)
    if err != nil {
        return fmt.Errorf("failed to get commit: %w", err)
    }
    
    tree, err := repo.TreeObject(commit.TreeHash)
    if err != nil {
        return fmt.Errorf("failed to get tree: %w", err)
    }
    
    // Use go-git filesystem operations to checkout files
    return c.checkoutTree(tree, worktreePath)
}
```

### 2. CLIGitClient Implementation

```go
// internal/infrastructure/cli_git_client.go
type CLIGitClient struct {
    executor CommandExecutor
    metrics  *HybridMetrics
}

type CommandExecutor interface {
    Execute(ctx context.Context, workingDir string, args ...string) (*CommandResult, error)
}

type CommandResult struct {
    Stdout   string
    Stderr   string
    ExitCode int
}

func NewCLIGitClient(executor CommandExecutor, metrics *HybridMetrics) *CLIGitClient {
    return &CLIGitClient{
        executor: executor,
        metrics:  metrics,
    }
}

func (c *CLIGitClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
    c.metrics.RecordOperation("cli", "create_worktree_attempt")
    
    args := []string{
        "worktree", "add",
        "-b", branchName,
        worktreePath,
        sourceBranch,
    }
    
    result, err := c.executor.Execute(ctx, repoPath, args...)
    if err != nil {
        return fmt.Errorf("git worktree add failed: %w", err)
    }
    
    if result.ExitCode != 0 {
        return fmt.Errorf("git worktree add failed: %s", result.Stderr)
    }
    
    c.metrics.RecordOperation("cli", "create_worktree_success")
    return nil
}

func (c *CLIGitClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error {
    args := []string{"worktree", "remove"}
    if !keepBranch {
        args = append(args, "--force")
    }
    args = append(args, worktreePath)
    
    result, err := c.executor.Execute(ctx, repoPath, args...)
    if err != nil {
        return fmt.Errorf("git worktree remove failed: %w", err)
    }
    
    if result.ExitCode != 0 {
        return fmt.Errorf("git worktree remove failed: %s", result.Stderr)
    }
    
    // Delete branch if requested
    if !keepBranch {
        branchName := c.extractBranchNameFromWorktree(worktreePath)
        if branchName != "" {
            deleteArgs := []string{"branch", "-D", branchName}
            c.executor.Execute(ctx, repoPath, deleteArgs...)
        }
    }
    
    return nil
}

func (c *CLIGitClient) ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error) {
    result, err := c.executor.Execute(ctx, repoPath, "worktree", "list", "--porcelain")
    if err != nil {
        return nil, fmt.Errorf("git worktree list failed: %w", err)
    }
    
    if result.ExitCode != 0 {
        return nil, fmt.Errorf("git worktree list failed: %s", result.Stderr)
    }
    
    return c.parseWorktreeList(result.Stdout)
}

func (c *CLIGitClient) parseWorktreeList(output string) ([]WorktreeInfo, error) {
    lines := strings.Split(output, "\n")
    var worktrees []WorktreeInfo
    
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        
        parts := strings.Fields(line)
        if len(parts) >= 2 {
            worktreePath := parts[0]
            commitHash := parts[1]
            branchName := ""
            
            // Extract branch name from commit annotation
            if len(parts) >= 3 && strings.HasPrefix(parts[2], "[") {
                branchAnnotation := strings.Trim(parts[2], "[]")
                branchParts := strings.Split(branchAnnotation, "/")
                if len(branchParts) > 1 {
                    branchName = strings.Join(branchParts[1:], "/")
                }
            }
            
            // Check if worktree is dirty
            dirty := strings.Contains(line, "bare")
            
            worktrees = append(worktrees, WorktreeInfo{
                Path:   worktreePath,
                Branch: branchName,
                Commit: commitHash,
                Dirty:  dirty,
            })
        }
    }
    
    return worktrees, nil
}
```

### 3. HybridClient with Fallback Logic

```go
// internal/infrastructure/hybrid_client.go
type HybridClient struct {
    goGit *GoGitClient
    cliGit *CLIGitClient
    metrics *HybridMetrics
    config HybridConfig
}

type HybridConfig struct {
    PreferCLI        bool
    FallbackTimeout  time.Duration
    RetryAttempts    int
    CacheEnabled     bool
}

func NewHybridClient(goGit *GoGitClient, cliGit *CLIGitClient, metrics *HybridMetrics, config HybridConfig) *HybridClient {
    return &HybridClient{
        goGit:  goGit,
        cliGit: cliGit,
        metrics: metrics,
        config: config,
    }
}

func (h *HybridClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
    // Try go-git first unless CLI is preferred
    if !h.config.PreferCLI {
        err := h.goGit.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)
        if err == nil {
            h.metrics.RecordFallback("create_worktree", false)
            return nil
        }
        
        // Check if error is fallback-eligible
        if h.isFallbackEligible(err) {
            h.metrics.RecordFallback("create_worktree", true)
            return h.cliGit.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)
        }
        
        return err
    }
    
    // CLI preferred path
    return h.cliGit.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)
}

func (h *HybridClient) isFallbackEligible(err error) bool {
    var worktreeErr *WorktreeNotSupportedError
    if errors.As(err, &worktreeErr) {
        return true
    }
    
    // Check for other go-git limitations
    if strings.Contains(err.Error(), "worktree") || 
       strings.Contains(err.Error(), "not supported") ||
       strings.Contains(err.Error(), "reference not found") {
        return true
    }
    
    return false
}

func (h *HybridClient) ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error) {
    // List worktrees - try go-git first, fallback to CLI
    worktrees, err := h.goGit.ListWorktrees(ctx, repoPath)
    if err == nil {
        h.metrics.RecordFallback("list_worktrees", false)
        return worktrees, nil
    }
    
    if h.isFallbackEligible(err) {
        h.metrics.RecordFallback("list_worktrees", true)
        return h.cliGit.ListWorktrees(ctx, repoPath)
    }
    
    return nil, err
}

func (h *HybridClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error {
    // Delete operations typically require CLI for proper cleanup
    if h.config.PreferCLI {
        return h.cliGit.DeleteWorktree(ctx, repoPath, worktreePath, keepBranch)
    }
    
    // Try go-git first
    err := h.goGit.DeleteWorktree(ctx, repoPath, worktreePath, keepBranch)
    if err == nil {
        h.metrics.RecordFallback("delete_worktree", false)
        return nil
    }
    
    if h.isFallbackEligible(err) {
        h.metrics.RecordFallback("delete_worktree", true)
        return h.cliGit.DeleteWorktree(ctx, repoPath, worktreePath, keepBranch)
    }
    
    return err
}
```

### 4. Performance Monitoring and Metrics

```go
// internal/infrastructure/hybrid_metrics.go
type HybridMetrics struct {
    operations map[string]*OperationMetrics
    mu         sync.RWMutex
}

type OperationMetrics struct {
    GoGitSuccess   int64
    GoGitFailure   int64
    CLISuccess     int64
    CLIFailure     int64
    FallbackCount  int64
    TotalLatency   time.Duration
}

func NewHybridMetrics() *HybridMetrics {
    return &HybridMetrics{
        operations: make(map[string]*OperationMetrics),
    }
}

func (m *HybridMetrics) RecordOperation(implementation, operation string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.operations[operation] == nil {
        m.operations[operation] = &OperationMetrics{}
    }
    
    metrics := m.operations[operation]
    switch implementation {
    case "go-git":
        metrics.GoGitSuccess++
    case "cli":
        metrics.CLISuccess++
    }
}

func (m *HybridMetrics) RecordFallback(operation string, triggered bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.operations[operation] == nil {
        m.operations[operation] = &OperationMetrics{}
    }
    
    if triggered {
        m.operations[operation].FallbackCount++
    }
}

func (m *HybridMetrics) GetReport() map[string]OperationMetrics {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    report := make(map[string]OperationMetrics)
    for op, metrics := range m.operations {
        report[op] = *metrics
    }
    
    return report
}
```

### 5. Command Executor Implementation

```go
// internal/infrastructure/command_executor.go
type RealCommandExecutor struct {
    timeout time.Duration
}

func NewRealCommandExecutor(timeout time.Duration) *RealCommandExecutor {
    return &RealCommandExecutor{
        timeout: timeout,
    }
}

func (e *RealCommandExecutor) Execute(ctx context.Context, workingDir string, args ...string) (*CommandResult, error) {
    ctx, cancel := context.WithTimeout(ctx, e.timeout)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "git", args...)
    cmd.Dir = workingDir
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err := cmd.Run()
    
    result := &CommandResult{
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        ExitCode: cmd.ProcessState.ExitCode(),
    }
    
    if err != nil && ctx.Err() == context.DeadlineExceeded {
        return result, fmt.Errorf("command timed out after %v", e.timeout)
    }
    
    return result, err
}
```

## Testing Strategy

### Unit Tests (Testify)

```go
// internal/infrastructure/hybrid_client_test.go
func TestHybridClient_CreateWorktree_GoGitSuccess(t *testing.T) {
    mockGoGit := &MockGoGitClient{}
    mockCLI := &MockCLIGitClient{}
    metrics := NewHybridMetrics()
    
    client := NewHybridClient(mockGoGit, mockCLI, metrics, HybridConfig{})
    
    repoPath := "/test/repo"
    branchName := "feature-branch"
    sourceBranch := "main"
    worktreePath := "/test/worktrees/repo/feature-branch"
    
    mockGoGit.On("CreateWorktree", mock.Anything, repoPath, branchName, sourceBranch, worktreePath).Return(nil)
    
    err := client.CreateWorktree(context.Background(), repoPath, branchName, sourceBranch, worktreePath)
    
    assert.NoError(t, err)
    mockGoGit.AssertExpectations(t)
    mockCLI.AssertNotCalled(t, "CreateWorktree")
    
    // Verify metrics
    report := metrics.GetReport()
    assert.Equal(t, int64(1), report["create_worktree"].GoGitSuccess)
    assert.Equal(t, int64(0), report["create_worktree"].FallbackCount)
}

func TestHybridClient_CreateWorktree_FallbackToCLI(t *testing.T) {
    mockGoGit := &MockGoGitClient{}
    mockCLI := &MockCLIGitClient{}
    metrics := NewHybridMetrics()
    
    client := NewHybridClient(mockGoGit, mockCLI, metrics, HybridConfig{})
    
    repoPath := "/test/repo"
    branchName := "feature-branch"
    sourceBranch := "main"
    worktreePath := "/test/worktrees/repo/feature-branch"
    
    worktreeErr := &WorktreeNotSupportedError{Operation: "create_worktree", Path: repoPath}
    mockGoGit.On("CreateWorktree", mock.Anything, repoPath, branchName, sourceBranch, worktreePath).Return(worktreeErr)
    mockCLI.On("CreateWorktree", mock.Anything, repoPath, branchName, sourceBranch, worktreePath).Return(nil)
    
    err := client.CreateWorktree(context.Background(), repoPath, branchName, sourceBranch, worktreePath)
    
    assert.NoError(t, err)
    mockGoGit.AssertExpectations(t)
    mockCLI.AssertExpectations(t)
    
    // Verify metrics
    report := metrics.GetReport()
    assert.Equal(t, int64(1), report["create_worktree"].FallbackCount)
    assert.Equal(t, int64(1), report["create_worktree"].CLISuccess)
}
```

### Integration Tests (Ginkgo/Gomega)

```go
// internal/infrastructure/hybrid_integration_test.go
var _ = Describe("Hybrid Git Operations", func() {
    var (
        tempDir    string
        repoPath   string
        client     *HybridClient
        metrics    *HybridMetrics
    )
    
    BeforeEach(func() {
        var err error
        tempDir, err = os.MkdirTemp("", "twiggit-hybrid-test")
        Expect(err).NotTo(HaveOccurred())
        
        repoPath = filepath.Join(tempDir, "test-repo")
        
        // Initialize test repository
        err = exec.Command("git", "init", repoPath).Run()
        Expect(err).NotTo(HaveOccurred())
        
        // Create initial commit
        testFile := filepath.Join(repoPath, "test.txt")
        err = os.WriteFile(testFile, []byte("test content"), 0644)
        Expect(err).NotTo(HaveOccurred())
        
        err = exec.Command("git", "-C", repoPath, "add", "test.txt").Run()
        Expect(err).NotTo(HaveOccurred())
        
        err = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit").Run()
        Expect(err).NotTo(HaveOccurred())
        
        // Setup hybrid client
        goGitClient := NewGoGitClient(NewHybridMetrics())
        cliClient := NewCLIGitClient(NewRealCommandExecutor(30*time.Second), NewHybridMetrics())
        metrics = NewHybridMetrics()
        client = NewHybridClient(goGitClient, cliClient, metrics, HybridConfig{})
    })
    
    AfterEach(func() {
        os.RemoveAll(tempDir)
    })
    
    Context("when creating worktrees", func() {
        It("should create worktree using go-git when possible", func() {
            worktreePath := filepath.Join(tempDir, "worktrees", "test-repo", "feature-branch")
            
            err := client.CreateWorktree(context.Background(), repoPath, "feature-branch", "main", worktreePath)
            Expect(err).NotTo(HaveOccurred())
            
            // Verify worktree exists
            _, err = os.Stat(filepath.Join(worktreePath, "test.txt"))
            Expect(err).NotTo(HaveOccurred())
            
            // Verify git status
            worktrees, err := client.ListWorktrees(context.Background(), repoPath)
            Expect(err).NotTo(HaveOccurred())
            Expect(len(worktrees)).To(Equal(2)) // main + feature-branch
        })
        
        It("should fallback to CLI when go-git fails", func() {
            // Test with complex scenario that triggers fallback
            worktreePath := filepath.Join(tempDir, "worktrees", "test-repo", "complex-branch")
            
            err := client.CreateWorktree(context.Background(), repoPath, "complex-branch", "main", worktreePath)
            Expect(err).NotTo(HaveOccurred())
            
            // Verify fallback was triggered
            report := metrics.GetReport()
            Expect(report["create_worktree"].FallbackCount).To(BeNumerically(">", 0))
        })
    })
    
    Context("when listing worktrees", func() {
        It("should list all worktrees correctly", func() {
            // Create multiple worktrees
            worktreePath1 := filepath.Join(tempDir, "worktrees", "test-repo", "feature-1")
            worktreePath2 := filepath.Join(tempDir, "worktrees", "test-repo", "feature-2")
            
            err := client.CreateWorktree(context.Background(), repoPath, "feature-1", "main", worktreePath1)
            Expect(err).NotTo(HaveOccurred())
            
            err = client.CreateWorktree(context.Background(), repoPath, "feature-2", "main", worktreePath2)
            Expect(err).NotTo(HaveOccurred())
            
            worktrees, err := client.ListWorktrees(context.Background(), repoPath)
            Expect(err).NotTo(HaveOccurred())
            Expect(len(worktrees)).To(Equal(3)) // main + 2 features
            
            // Verify worktree details
            branchNames := []string{}
            for _, wt := range worktrees {
                branchNames = append(branchNames, wt.Branch)
            }
            Expect(branchNames).To(ContainElements("main", "feature-1", "feature-2"))
        })
    })
})
```

### E2E Tests (Gexec)

```go
// cmd/hybrid_e2e_test.go
var _ = Describe("Hybrid Git Operations E2E", func() {
    var (
        tempDir  string
        repoPath string
        binPath  string
    )
    
    BeforeEach(func() {
        var err error
        tempDir, err = os.MkdirTemp("", "twiggit-e2e-hybrid")
        Expect(err).NotTo(HaveOccurred())
        
        repoPath = filepath.Join(tempDir, "test-repo")
        
        // Build binary for testing
        binPath = filepath.Join(tempDir, "twiggit")
        session := gexec.Start(gexec.Command("go", "build", "-o", binPath, "."), GinkgoWriter, GinkgoWriter)
        Eventually(session).Should(gexec.Exit(0))
        
        // Initialize test repository
        err = exec.Command("git", "init", repoPath).Run()
        Expect(err).NotTo(HaveOccurred())
        
        testFile := filepath.Join(repoPath, "test.txt")
        err = os.WriteFile(testFile, []byte("test content"), 0644)
        Expect(err).NotTo(HaveOccurred())
        
        err = exec.Command("git", "-C", repoPath, "add", "test.txt").Run()
        Expect(err).NotTo(HaveOccurred())
        
        err = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit").Run()
        Expect(err).NotTo(HaveOccurred())
    })
    
    AfterEach(func() {
        os.RemoveAll(tempDir)
    })
    
    It("should create and manage worktrees end-to-end", func() {
        // Test create command
        session := gexec.Start(gexec.Command(binPath, "create", "test-repo", "feature-branch"), GinkgoWriter, GinkgoWriter)
        Eventually(session).Should(gexec.Exit(0))
        
        // Verify worktree was created
        worktreePath := filepath.Join(tempDir, "Worktrees", "test-repo", "feature-branch")
        Expect(worktreePath).To(BeADirectory())
        
        // Test list command
        session = gexec.Start(gexec.Command(binPath, "list"), GinkgoWriter, GinkgoWriter)
        Eventually(session).Should(gexec.Exit(0))
        Expect(string(session.Out.Contents())).To(ContainSubstring("feature-branch"))
        
        // Test delete command
        session = gexec.Start(gexec.Command(binPath, "delete", "test-repo", "feature-branch"), GinkgoWriter, GinkgoWriter)
        Eventually(session).Should(gexec.Exit(0))
        
        // Verify worktree was deleted
        Expect(worktreePath).NotTo(BeADirectory())
    })
})
```

## Implementation Summary

### Direct Implementation Items (Context Detection Completion)
These are the remaining 5% of context detection that need immediate implementation:

1. **Resolution Suggestions** (`context_resolver.go:97,145,170`)
   - Replace TODO comments with filesystem-based discovery
   - Scan worktrees directory for existing branches
   - Add project discovery from ProjectsDirectory

2. **Enhanced Caching** (`context_detector.go:38-66`)
   - Add TTL-based expiration
   - Fix cache key normalization (use normalized paths)
   - Add cache invalidation methods

3. **Error Types** (`domain/errors.go`)
   - Add git-specific error types
   - Implement error context preservation
   - Add validation error types

4. **Configuration** (`domain/config.go`)
   - Add context detection configuration options
   - Add git operation settings
   - Add caching parameters

### Git-Focused Components (Hybrid Operations)
These are the core git operations that constitute 70% of this phase:

1. **HybridGitClient Interface** - Core abstraction for git operations
2. **GoGitClient Implementation** - 70% coverage (repository ops, branch listing)
3. **CLIGitClient Implementation** - 30% coverage (worktree operations)
4. **HybridClient Integration** - Intelligent fallback logic
5. **Performance Monitoring** - Metrics and optimization

## Implementation Strategy

### Part A: Context Detection Completion (Week 1) - Direct Implementation

#### A1. Complete Resolution Suggestions with Filesystem Discovery
**Target**: Replace TODO comments in `context_resolver.go:97,145,170`

```go
// Add to internal/infrastructure/context_resolver.go
func (cr *contextResolver) getProjectContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
    var suggestions []*domain.ResolutionSuggestion
    
    // Always suggest "main"
    if strings.HasPrefix("main", partial) {
        suggestions = append(suggestions, &domain.ResolutionSuggestion{
            Text:        "main",
            Description: "Project root directory",
            Type:        domain.PathTypeProject,
            ProjectName: ctx.ProjectName,
        })
    }
    
    // Scan worktrees directory for existing branches
    worktreeProjectDir := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName)
    if entries, err := os.ReadDir(worktreeProjectDir); err == nil {
        for _, entry := range entries {
            if entry.IsDir() && strings.HasPrefix(entry.Name(), partial) {
                suggestions = append(suggestions, &domain.ResolutionSuggestion{
                    Text:        entry.Name(),
                    Description: fmt.Sprintf("Worktree for branch '%s'", entry.Name()),
                    Type:        domain.PathTypeWorktree,
                    ProjectName: ctx.ProjectName,
                    BranchName:  entry.Name(),
                })
            }
        }
    }
    
    return suggestions
}
```

#### A2. Enhanced Context Caching with TTL
**Target**: Improve `context_detector.go:38-66` with TTL and normalization

```go
// Add to internal/infrastructure/context_detector.go
type CacheEntry struct {
    Context   *domain.Context
    ExpiresAt time.Time
}

func (cd *contextDetector) DetectContext(dir string) (*domain.Context, error) {
    // Normalize path first for consistent cache keys
    normalizedDir, err := NormalizePath(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to normalize directory: %w", err)
    }
    
    // Check cache with TTL
    cd.cacheMu.RLock()
    if cached, exists := cd.cache[normalizedDir]; exists && time.Now().Before(cached.ExpiresAt) {
        cd.cacheMu.RUnlock()
        return cached.Context, nil
    }
    cd.cacheMu.RUnlock()
    
    // Continue with detection...
}
```

#### A3. Git-Specific Error Types
**Target**: Extend `domain/errors.go` with git operation errors

```go
// Add to internal/domain/errors.go
type GitOperationError struct {
    Operation string
    Path      string
    Cause     error
}

func (e *GitOperationError) Error() string {
    return fmt.Sprintf("git operation '%s' failed at %s: %v", e.Operation, e.Path, e.Cause)
}

type WorktreeValidationError struct {
    WorktreePath string
    Reason       string
}

func (e *WorktreeValidationError) Error() string {
    return fmt.Sprintf("worktree validation failed for %s: %s", e.WorktreePath, e.Reason)
}
```

#### A4. Configuration Extensions
**Target**: Add context detection configuration to `domain/config.go`

```go
// Add to internal/domain/config.go
type ContextDetectionConfig struct {
    CacheTTL                time.Duration `toml:"cache_ttl"`
    MaxCacheEntries         int           `toml:"max_cache_entries"`
    GitOperationTimeout     time.Duration `toml:"git_operation_timeout"`
    EnableGitValidation     bool          `toml:"enable_git_validation"`
    FallbackToFilesystem    bool          `toml:"fallback_to_filesystem"`
}

type Config struct {
    // Existing fields...
    ContextDetection ContextDetectionConfig `toml:"context_detection"`
}
```

### Part B: Git Interface Foundation (Week 1-2)

#### B1. Git Context Provider Interface
**Target**: Define git integration interfaces in `domain/git_interfaces.go`

```go
// internal/domain/git_interfaces.go
type GitContextProvider interface {
    ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)
    ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
    ValidateWorktree(ctx context.Context, worktreePath string) error
    GetRepositoryStatus(ctx context.Context, repoPath string) (RepositoryStatus, error)
}

type BranchInfo struct {
    Name   string
    Remote string
    Head   string
}

type WorktreeInfo struct {
    Path     string
    Branch   string
    Commit   string
    Dirty    bool
    Detached bool
}

type RepositoryStatus struct {
    Clean     bool
    Staged    int
    Modified  int
    Untracked int
}
```

#### B2. Extend ContextService with Git Integration
**Target**: Add git methods to `service/context_service.go`

```go
// Add to internal/service/context_service.go
type ContextService struct {
    detector  domain.ContextDetector
    resolver  domain.ContextResolver
    config    *domain.Config
    gitProvider GitContextProvider  // New field
}

func (cs *ContextService) GetGitProvider() GitContextProvider {
    return cs.gitProvider
}

func (cs *ContextService) ListAvailableBranches(projectName string) ([]BranchInfo, error) {
    projectPath := filepath.Join(cs.config.ProjectsDirectory, projectName)
    return cs.gitProvider.ListBranches(context.Background(), projectPath)
}

func (cs *ContextService) ListExistingWorktrees(projectName string) ([]WorktreeInfo, error) {
    projectPath := filepath.Join(cs.config.ProjectsDirectory, projectName)
    return cs.gitProvider.ListWorktrees(context.Background(), projectPath)
}
```

### Part C: Hybrid Git Implementation (Week 2-4)

#### C1. Core Interface and Error System
1. **Define HybridGitClient interface** in `internal/infrastructure/interfaces.go`
2. **Implement hybrid error types** in `internal/domain/errors.go`
3. **Create metrics system** in `internal/infrastructure/hybrid_metrics.go`
4. **Write unit tests** for interfaces and error handling

#### C2. GoGitClient Implementation (70% coverage)
1. **Implement basic repository operations** (open, list branches, status)
2. **Implement custom worktree creation** using storage abstractions
3. **Add caching and performance optimizations**
4. **Write comprehensive unit tests** with mocks

#### C3. CLIGitClient Implementation (30% coverage)
1. **Implement command executor** with timeout and error handling
2. **Implement all worktree operations** using git CLI
3. **Add output parsing** for worktree list and status
4. **Write unit tests** with command mocking

#### C4. HybridClient Integration
1. **Implement fallback logic** with error detection
2. **Add configuration options** for hybrid behavior
3. **Integrate metrics collection**
4. **Write integration tests** for fallback scenarios

### Part D: Integration & Performance (Week 4-5)

#### D1. Complete Context-Git Integration
1. **Replace all TODO comments** with actual git discovery
2. **Integrate GitContextProvider into ContextResolver**
3. **Add git-aware validation to ContextDetector**
4. **Implement cross-project reference validation**

#### D2. Performance and Optimization
1. **Add git-aware caching layer** with TTL
2. **Implement concurrent operations** where safe
3. **Add performance monitoring** and reporting
4. **Optimize for large repositories**

#### D3. Testing and Validation
1. **Run comprehensive test suite** (unit, integration, E2E)
2. **Test with real repositories** of various sizes
3. **Validate fallback behavior** under error conditions
4. **Performance benchmarking** against pure CLI approach

## Configuration Options

```toml
# config.toml additions for context detection
[context_detection]
cache_ttl = "5m"
max_cache_entries = 1000
git_operation_timeout = "30s"
enable_git_validation = true
fallback_to_filesystem = true

[context_detection.git]
prefer_cli_for_validation = false
worktree_discovery_timeout = "10s"
branch_cache_ttl = "15m"

# config.toml additions for hybrid system
[hybrid]
prefer_cli = false
fallback_timeout = "30s"
retry_attempts = 3
cache_enabled = true
cache_ttl = "5m"

[hybrid.operations]
create_worktree = "go-git"  # "go-git" or "cli"
delete_worktree = "cli"
list_worktrees = "go-git"
```

## Performance Considerations

### Caching Strategy
- Repository objects cached with TTL
- Worktree lists cached for short duration
- Branch information cached longer term

### Concurrent Operations
- Read operations (list, status) can be concurrent
- Write operations (create, delete) require serialization
- Fallback operations isolated to prevent race conditions

### Memory Management
- Repository objects properly closed
- Command output limited in size
- Cache size bounded with LRU eviction

## Error Handling Patterns

### Consistent Error Types
```go
// All implementations return these consistent error types
var (
    ErrRepositoryNotFound = errors.New("git repository not found")
    ErrWorktreeExists = errors.New("worktree already exists")
    ErrWorktreeNotFound = errors.New("worktree not found")
    ErrBranchNotFound = errors.New("branch not found")
    ErrInvalidRepository = errors.New("invalid git repository")
)
```

### Error Context Preservation
```go
func wrapHybridError(operation, implementation string, err error) error {
    return &HybridError{
        Operation:     operation,
        Implementation: implementation,
        OriginalErr:   err,
        Context:       make(map[string]interface{}),
    }
}
```

## Success Criteria

### Context Detection Completion
- [ ] All TODO comments resolved with git integration (context_resolver.go:97,145,170)
- [ ] TTL-based caching implemented with proper normalization
- [ ] Git-specific error types added to domain/errors.go
- [ ] Configuration extensions complete and validated
- [ ] Filesystem-based discovery working as fallback

### Git Operations Implementation
- [ ] HybridGitClient with intelligent fallback logic working
- [ ] 70% go-git coverage achieved (repository ops, branch listing)
- [ ] 30% CLI fallback coverage achieved (worktree operations)
- [ ] Performance metrics collected and reported
- [ ] GitContextProvider interface fully implemented

### Integration Quality
- [ ] ContextService fully integrated with git operations
- [ ] Resolution suggestions powered by actual git discovery
- [ ] Cross-project reference validation implemented
- [ ] Git-aware validation added to ContextDetector
- [ ] Cross-platform compatibility maintained

### Performance and Quality
- [ ] >90% test coverage for context detection and git operations
- [ ] All integration tests pass with real git repositories
- [ ] Performance benchmarks meet targets (<50ms for context detection)
- [ ] Memory usage remains within bounds with proper caching
- [ ] Fallback latency is acceptable (<100ms additional overhead)

### Reliability and User Experience
- [ ] Graceful fallback on go-git failures with proper error messages
- [ ] Proper cleanup on errors and resource management
- [ ] Consistent error messages across all implementations
- [ ] Completion suggestions provide real value with git discovery
- [ ] Configuration options provide appropriate flexibility

## Timeline

**Total Duration**: 5 weeks
- **Week 1**: Complete context detection + git interface foundation
- **Week 2-3**: Hybrid git client implementation  
- **Week 4**: Integration and performance optimization
- **Week 5**: Testing, documentation, and validation

## Dependencies

### Context Detection Dependencies
- Current context detection implementation (95% complete)
- Configuration system (already implemented)
- Testing framework (Testify, Ginkgo/Gomega already available)

### Git Operations Dependencies
- go-git library for core git operations
- Git CLI availability for fallback operations
- Command execution infrastructure
- Performance monitoring system

### Integration Dependencies
- Completed context detection system
- Hybrid git client implementation
- Configuration extensions
- Testing infrastructure

## References

- **Technology Stack**: technology.md lines 15-18 (go-git limitations and hybrid approach)
- **Implementation Requirements**: implementation.md lines 200-224 (error handling patterns)
- **Testing Requirements**: implementation.md lines 1-23 (testing framework usage)
- **Performance Requirements**: implementation.md lines 71-91 (operational performance)
- **Context Detection**: 03-context-detection.md (TODO comments and integration points)
- **Current Implementation**: context_resolver.go:97,145,170 (missing git integration)

This amended plan focuses Phase 04 specifically on git-related components while ensuring the context detection system is completed with direct implementation of the missing pieces. The hybrid git operations remain the core focus, but now with proper integration into the existing context detection foundation.