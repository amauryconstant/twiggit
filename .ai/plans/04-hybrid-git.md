# Hybrid Git Operations Implementation Plan

## Overview

This plan implements the sophisticated hybrid git operations system for twiggit, combining go-git's performance with CLI fallback for critical functionality gaps. The hybrid approach addresses go-git's worktree limitations while maintaining cross-platform consistency and performance.

**Context**: Foundation, configuration, and context detection layers are established. This layer provides the core git operations with intelligent fallback mechanisms.

**Key Challenge**: "go-git limitations: worktree functionality gaps requiring hybrid approach" (technology.md:18)

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
    worktreePath := "/test/workspaces/repo/feature-branch"
    
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
    worktreePath := "/test/workspaces/repo/feature-branch"
    
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
            worktreePath := filepath.Join(tempDir, "workspaces", "test-repo", "feature-branch")
            
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
            worktreePath := filepath.Join(tempDir, "workspaces", "test-repo", "complex-branch")
            
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
            worktreePath1 := filepath.Join(tempDir, "workspaces", "test-repo", "feature-1")
            worktreePath2 := filepath.Join(tempDir, "workspaces", "test-repo", "feature-2")
            
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
        worktreePath := filepath.Join(tempDir, "Workspaces", "test-repo", "feature-branch")
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

## Implementation Steps

### Phase 1: Core Interface and Error System
1. **Define interfaces** in `internal/infrastructure/interfaces.go`
2. **Implement error types** in `internal/domain/errors.go`
3. **Create metrics system** in `internal/infrastructure/hybrid_metrics.go`
4. **Write unit tests** for interfaces and error handling

### Phase 2: GoGitClient Implementation
1. **Implement basic repository operations** (open, list branches, status)
2. **Implement custom worktree creation** using storage abstractions
3. **Add caching and performance optimizations**
4. **Write comprehensive unit tests** with mocks

### Phase 3: CLIGitClient Implementation
1. **Implement command executor** with timeout and error handling
2. **Implement all worktree operations** using git CLI
3. **Add output parsing** for worktree list and status
4. **Write unit tests** with command mocking

### Phase 4: HybridClient Integration
1. **Implement fallback logic** with error detection
2. **Add configuration options** for hybrid behavior
3. **Integrate metrics collection**
4. **Write integration tests** for fallback scenarios

### Phase 5: Performance and Optimization
1. **Add caching layer** for repository objects
2. **Implement concurrent operations** where safe
3. **Add performance monitoring** and reporting
4. **Optimize for large repositories**

### Phase 6: Testing and Validation
1. **Run comprehensive test suite** (unit, integration, E2E)
2. **Test with real repositories** of various sizes
3. **Validate fallback behavior** under error conditions
4. **Performance benchmarking** against pure CLI approach

## Configuration Options

```toml
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

1. **Functional Requirements**
   - All worktree operations work with hybrid fallback
   - Consistent behavior across implementations
   - Proper error handling and recovery

2. **Performance Requirements**
   - Operations complete within specified time limits
   - Memory usage remains within bounds
   - Fallback latency is acceptable

3. **Quality Requirements**
   - >80% test coverage for hybrid logic
   - All integration tests pass
   - Performance benchmarks meet targets

4. **Reliability Requirements**
   - Graceful fallback on go-git failures
   - Proper cleanup on errors
   - Consistent error messages

## References

- **Technology Stack**: technology.md lines 15-18 (go-git limitations and hybrid approach)
- **Implementation Requirements**: implementation.md lines 200-224 (error handling patterns)
- **Testing Requirements**: implementation.md lines 1-23 (testing framework usage)
- **Performance Requirements**: implementation.md lines 71-91 (operational performance)

This implementation plan provides a comprehensive approach to building the hybrid git operations system while maintaining consistency with the established architecture and requirements.