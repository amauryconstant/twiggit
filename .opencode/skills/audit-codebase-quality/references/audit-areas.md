# Audit Areas

Configurable audit categories for codebase quality review. Each area defines what to look for, examples, and severity guidelines.

## Extensibility

Add new audit areas by creating sections below. When adding new areas:

1. Define what the area checks for
2. Provide concrete examples of findings
3. Establish severity guidelines
4. Reference language-specific patterns if applicable
5. Consider if the area should use parallel execution or sequential

---

**Note:** See [validation-checklist.md](validation-checklist.md) for finding validation criteria.

---## Package Structure

Check for naming conventions, directory organization, layer separation, and architectural violations.

### What to Look For

- Package naming consistency (singular vs plural)
- Directory structure violations (deep nesting, inappropriate locations)
- Layer boundary violations (e.g., service logic in domain layer)
- Interface placement (application vs domain vs infrastructure)
- Naming conflicts (same name in different packages)
- File naming conventions (snake_case vs camelCase)
- Orphaned or misplaced files

### Examples

- **Package naming**: `services` package (plural) while all others use singular (`application`, `domain`, `infrastructure`)
- **Layer violations**: Service interfaces defined in `domain/interfaces.go` instead of `application/interfaces.go`
- **Interface placement**: `ContextService` interface in wrong layer (domain instead of application)
- **Naming conflicts**: Both `domain.ContextService` (interface) and `services.ContextService` (implementation) exist with same name
- **Test file naming**: `validators_test.go` when source is `validation.go`

### Severity Guidelines

- **CRITICAL**: Architectural violations that break layer separation or design principles
- **HIGH**: Package naming violations that affect consistency across codebase
- **MEDIUM**: File naming violations or minor structural issues
- **LOW**: Stylistic variations that don't affect functionality

**See Finding Validation Checklist above.**

### Language-Specific Notes

For Go projects:
- Package names should be singular (e.g., `service`, not `services`)
- Interfaces in `application/`, implementations in `services/` or `infrastructure/`
- Domain layer should not define service interfaces
- Test files match source: `validation.go` → `validation_test.go`

See [go-patterns.md](go-patterns.md) for Go-specific patterns.

---

## Duplicate Code

Check for repeated logic, similar functions, and consolidation opportunities.

### What to Look For

- Identical or very similar functions in different files
- Duplicate struct or interface definitions
- Repeated error handling patterns
- Similar validation logic
- Copied code blocks
- Configuration or initialization logic repeated

### Examples

- **Duplicate templates**: Shell wrapper templates (bash, zsh, fish) defined in both `domain/shell.go` and `infrastructure/shell_infra.go`
- **Duplicate validation**: `isValidShellType()` function exists in both `domain/shell.go` and `services/shell_service.go`
- **Duplicate file listing**: `ConfigFiles()` and `getConfigFiles()` return nearly identical shell config file lists
- **Duplicate project discovery**: Similar directory scanning and validation logic in `context_resolver.go` and `project_service.go`
- **Repeated error handling**: Same path validation error pattern repeated 5 times in `context_resolver.go`

### Severity Guidelines

- **CRITICAL**: Complete duplication of complex logic (>100 lines)
- **HIGH**: Duplicate functions or templates (>20 lines)
- **MEDIUM**: Similar logic or repeated error patterns (5-20 lines)
- **LOW**: Minor repetitions or similar patterns

**See Finding Validation Checklist above.**

---

## Interface Compliance

Check for implementation completeness, signature mismatches, and unused interfaces.

### What to Look For

- Interface methods not implemented
- Implementation methods with wrong signatures
- Unused or unimplemented interfaces
- Interface methods only used in tests
- Overly broad or overly narrow interfaces

### Examples

- **Dead interfaces**: `GitRepositoryInterface` and `ProjectRepository` defined but never implemented or used
- **Signature mismatches**: `DeleteWorktree` has `force` parameter in interface but `keepBranch` in documentation (if inconsistency exists)
- **Unused methods**: `ContextDetector.InvalidateCacheForRepo()` and `ClearCache()` only used in tests, not production
- **Missing implementations**: Interface defined but no struct satisfies it

### Severity Guidelines

- **CRITICAL**: Critical interface in wrong layer causing architectural violation
- **HIGH**: Unused interfaces (>30 lines of dead code)
- **MEDIUM**: Signature mismatches or missing implementations
- **LOW**: Unused methods on otherwise valid interfaces

**See Finding Validation Checklist above.**

---

## Test Patterns

Check for test organization, coverage gaps, and mock usage consistency.

### What to Look For

- Test file organization and naming
- Mock patterns (inline vs centralized)
- Coverage gaps (files without tests)
- Test pattern consistency (suite-based vs table-driven)
- Mock duplication
- TDD artifacts or outdated comments
- Critical files without comprehensive tests

### Examples

- **Inline mocks**: `mockShellIntegration` defined in `shell_service_test.go` instead of centralized `test/mocks/cmd_mocks.go`
- **Mock duplication**: `MockWorktreeService`, `MockProjectService`, `MockNavigationService` exist in both `test/mocks/cmd_mocks.go` and `test/mocks/services/`
- **MockCLIClient duplication**: Inline functional mock in `cli_client.go:324-377` AND centralized Testify mock in `test/mocks/git_service_mock.go:85-126` - creates maintenance burden
  ```go
  // BAD: Inline mock in cli_client.go
  type MockCLIClient struct {
      CreateWorktreeFunc func(...)
      DeleteWorktreeFunc func(...)
      ListWorktreesFunc func(...)
      // ...
  }
  
  // GOOD: Use centralized mock from test/mocks/
  // Remove inline version, use test/mocks.MockCLIClient consistently
  ```
- **Critical files missing tests**: Files >100 lines of production code with no test coverage
  - `git_client.go` (136 lines) - contains CompositeGitClient with routing logic for 13 methods
  - `error_handler.go` (137 lines) - contains critical CLI error handling functions:
    - `HandleCLIError()` - main error handling
    - `GetExitCodeForError()` - error to exit code mapping
    - `CategorizeError()` - error categorization with reflection
    - `IsCobraArgumentError()` - pattern matching for argument errors
- **Coverage gaps**: Error type files (`errors.go`, `service_errors.go`, `shell_errors.go`) lack tests
- **Pattern mixing**: `cli_client_test.go` uses both suite-based and table-driven tests inconsistently
- **Naming inconsistency**: `ContextResolverMock` should be `MockContextResolver` for consistency

### Severity Guidelines

- **CRITICAL**: All application service mocks duplicated (causes maintenance burden)
- **HIGH**: Critical files missing tests (error constructors, validation logic)
- **MEDIUM**: Inline mocks that should be centralized, test pattern inconsistency
- **LOW**: TDD comments or minor naming variations

**See Finding Validation Checklist above.**

---

## Documentation Accuracy

Check AGENTS.md and other documentation files against actual code implementation.

### What to Look For

- Missing fields in struct documentation
- Incorrect method signatures
- Wrong parameter types or names
- Undocumented types (error types, result types, domain types)
- Missing methods in interface documentation
- Outdated patterns (e.g., referencing removed fields or deprecated approaches)
- Misleading information (non-existent features)
- Missing error type documentation
- Parameter name inconsistencies in documentation

### Examples

- **Missing error type documentation**: Major error types completely undocumented
  - `GitCommandError` (errors.go:119-171) - Git CLI command execution failures with detailed command info, exit codes, stdout/stderr
  - `ServiceError` (service_errors.go:8-32) - General service operation error type
  - `ShellError` (shell_errors.go:37-87) - Shell service errors with context
  - `ResolutionError` (service_errors.go:199-231) - Path resolution errors with suggestions
  - `ConflictError` (service_errors.go:233-259) - Operation conflict errors
- **Incorrect interface method signatures**: Parameter names don't match implementation
  ```go
  // DOCUMENTATION (WRONG):
  type CLIClient interface {
      DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error
  }
  
  // ACTUAL CODE (CORRECT):
  type CLIClient interface {
      DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error
  }
  ```
  - Issue: `keepBranch` vs `force` parameter name mismatch
- **Missing interface methods**: Documentation incomplete for GoGitClient
  ```go
  // DOCUMENTATION MISSING:
  GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error)
  ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error)
  GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error)
  ```
- **Missing CLIClient methods**: Documentation missing `PruneWorktrees` and `IsBranchMerged` methods
- **Missing fields**: Documentation shows `Force bool` field in `CreateWorktreeRequest` (missing in docs)
- **Incorrect types**: Docs reference `ProjectRepository` but actual uses `application.ProjectService`
- **Wrong package names**: Docs show `service_requests.CreateWorktreeRequest` but actual is `domain.CreateWorktreeRequest`
- **Missing context**: Shell interface methods (`Type()`, `Path()`, `Version()`) completely undocumented

### Severity Guidelines

- **CRITICAL**: Documentation completely missing for critical types (error types, result types)
- **HIGH**: Incorrect signatures or types affecting architectural understanding
- **MEDIUM**: Missing fields or methods (documentation is incomplete but usable)
- **LOW**: Typos or minor inaccuracies

**See Finding Validation Checklist above.**

---

## Import Consistency

Check for import ordering, circular dependencies, and dependency health.

### What to Look For

- Import ordering violations (stdlib, third-party, internal)
- Circular dependencies between packages
- Unused imports
- Import aliases (used appropriately)
- External dependency management
- Layer dependency violations (e.g., services importing domain only, not application)

### Examples

- **Circular dependencies**: A imports B, B imports C, C imports A (rare but critical)
- **Import ordering**: Files mixing stdlib and third-party imports randomly
- **Unused imports**: Import statement but package never referenced
- **Wrong layer dependencies**: Application layer importing infrastructure directly (should go through services)

### Severity Guidelines

- **CRITICAL**: Circular dependencies breaking compilation or design
- **HIGH**: Layer violations or dependency injection issues
- **MEDIUM**: Import ordering or unused imports
- **LOW**: Inconsistent alias usage

**See Finding Validation Checklist above.**

### Language-Specific Notes

For Go projects:
- Order: stdlib (alphabetical), third-party (alphabetical), internal (alphabetical)
- No circular dependencies (verified via `go build`)
- Minimal external dependencies
- Clean layer separation: `services` → `application`, `domain`, `infrastructure`

### Layer Dependency Violations

Check for violations of clean architecture layering principles where code imports from inappropriate layers.

### What to Look For

- Command layer (cmd/) importing infrastructure directly
- Application layer importing infrastructure directly
- Infrastructure layer importing services
- Domain layer depending on any other internal layer
- Circumventing application services in cmd layer

### Examples

- **cmd importing infrastructure directly**:
  ```go
  // cmd/root.go:27 - BAD: cmd depends on infrastructure
  import (
      "myapp/internal/infrastructure"
  )
  
  type ServiceContainer struct {
      GitClient infrastructure.GitClient  // Should go through application layer
  }
  
  // cmd/create.go:76 - BAD: Direct infrastructure call
  config.Services.GitClient.BranchExists(ctx, project.Path, source)
  
  // CORRECT: Use application service
  config.Services.WorktreeService.BranchExists(ctx, project.Path, source)
  ```
- **cmd/importing infrastructure bypasses application layer**: Commands should only use application services
  ```go
  // cmd/delete.go:48 - BAD: Direct infrastructure access
  worktrees, err := config.Services.GitClient.ListWorktrees(ctx, project.Path)
  
  // CORRECT: Go through application service
  worktreeInfo, err := config.Services.WorktreeService.GetWorktreeStatus(ctx, req)
  ```

### Severity Guidelines

- **HIGH**: cmd layer importing infrastructure directly (architectural violation)
- **MEDIUM**: Application layer importing infrastructure when service layer exists
- **LOW**: Minor layering issues that don't affect architecture

See [go-patterns.md](go-patterns.md) for Go-specific patterns.

---

## Error Handling

Check for error wrapping patterns, error type usage, and message consistency.

### What to Look For

- Mixed error wrapping styles (fmt.Errorf vs domain error types)
- Inconsistent error type selection
- Error message format inconsistency (casing, prefixes)
- Loss of error context (not using %w)
- Error chain breaks (converting errors to nil returns)
- String-based error detection (strings.Contains instead of errors.As)
- Panic in production code
- Inconsistent boundary error handling

### Examples

- **Mixed wrapping**: Infrastructure layer sometimes returns `fmt.Errorf()` and sometimes `domain.NewGitRepositoryError()`
- **Plain errors**: Domain constructors using `errors.New()` instead of `ValidationError`
- **Context loss**: `context_detector.go` returns plain error without wrapping underlying cause
- **Message casing**: Some errors use "failed to", others use "Failed to"
- **Chain breaks**: Errors converted to nil returns, then checked separately
- **String-based error detection** (CRITICAL): Using `strings.Contains(err.Error(), "...")` instead of `errors.As()`
  ```go
  // BAD: Fragile string-based detection - breaks if error message changes
  errStr := err.Error()
  if strings.Contains(errStr, "worktree not found") ||
      strings.Contains(errStr, "invalid git repository") ||
      strings.Contains(errStr, "repository does not exist") ||
      strings.Contains(errStr, "no such file or directory") {
      return fmt.Errorf("worktree not found: %s", worktreePath)
  }
  
  // GOOD: Type-based error detection
  var worktreeErr *domain.WorktreeServiceError
  var gitErr *domain.GitRepositoryError
  if errors.As(err, &worktreeErr) {
      return fmt.Errorf("worktree not found: %s", worktreePath)
  } else if errors.As(err, &gitErr) {
      return fmt.Errorf("git repository error: %w", gitErr)
  }
  ```
- **Layer-specific error type violations**: Infrastructure layer using `fmt.Errorf` instead of domain types
  ```go
  // BAD: Infrastructure returns plain fmt.Errorf
  return fmt.Errorf("failed to validate %s path: %w", targetType, err)
  
  // GOOD: Infrastructure returns domain error type
  return domain.NewContextDetectionError(target, "path validation failed", err)
  
  // BAD: Domain constructor uses fmt.Errorf
  if !IsValidShellType(shellType) {
      return nil, fmt.Errorf("unsupported shell type: %s", shellType)
  }
  
  // GOOD: Domain constructor uses ValidationError
  if !IsValidShellType(shellType) {
      return nil, domain.NewValidationError("NewShell", "shellType", string(shellType), "unsupported shell type").
          WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
  }
  ```

### Severity Guidelines

- **CRITICAL**: Panic in production code, complete loss of error context, string-based error detection (strings.Contains instead of errors.As)
- **HIGH**: Layer-specific error type violations (infrastructure using fmt.Errorf), mixed wrapping styles in critical paths
- **MEDIUM**: Mixed wrapping styles, error chain breaks, message casing inconsistencies
- **LOW**: Minor format variations or minor inconsistencies

**See Finding Validation Checklist above.**

---

## Mock Centralization

Check for inline vs centralized mocks, duplicate implementations, and missing mocks.

### What to Look For

- Inline mock definitions in test files
- Duplicate mock implementations for same interface
- Missing centralized mocks for infrastructure interfaces
- Mock naming inconsistencies
- Mocks with incomplete method implementations

### Examples

- **Inline mocks**: `mockShellIntegration` in `shell_service_test.go`, `mockCommandExecutor` in `command_executor_test.go`
- **Duplicate mocks**: `MockWorktreeService`, `MockProjectService`, `MockNavigationService` exist in both `test/mocks/cmd_mocks.go` and `test/mocks/services/`
- **Missing mocks**: No `MockShellInfrastructure` in centralized mocks
- **Naming inconsistency**: `ContextResolverMock` should be `MockContextResolver` for consistent naming pattern

### Severity Guidelines

- **CRITICAL**: All application service mocks duplicated (maintenance burden)
- **HIGH**: Inline mocks for commonly used interfaces
- **MEDIUM**: Missing centralized mocks, naming inconsistencies
- **LOW**: Mocks for niche or rarely-used interfaces

**See Finding Validation Checklist above.**

---
 
## Cross-References Between Audit Areas

When documenting or implementing audit areas, reference related areas to help users understand relationships:

### Common Cross-References

- **Error handling** ↔ **Import consistency**: Mixed error wrapping often causes import issues
- **Duplicate code** ↔ **Interface compliance**: Duplicate interfaces may not implement all methods
- **Test patterns** ↔ **Mock centralization**: Test mock usage issues relate to mock organization
- **Documentation accuracy** ↔ **Package structure**: Incorrect documentation often reflects structural issues

### Audit Area Relationships

| Area | Related To | Rationale |
|-------|-----------|-----------|
| package-structure | Interface compliance, Duplicate code, Test patterns | Layer structure affects interface placement and code organization |
| duplicate-code | Interface compliance, Documentation accuracy | Duplicated code may not match documented interfaces |
| interface-compliance | Package structure, Mock centralization | Interface placement affects implementation patterns |
| test-patterns | Mock centralization, Import consistency | Test organization relates to mock and import patterns |
| documentation-accuracy | Package structure, Interface compliance | Documentation should match code structure |
| import-consistency | Package structure, Error handling | Import layering affects error propagation |
| error-handling | Package structure, Documentation accuracy, Test patterns | Error handling depends on architectural decisions |
| mock-centralization | Test patterns, Interface compliance | Mock usage is part of test organization |

### Guidelines

- When adding a new audit area, review existing areas and add cross-references
- When documenting findings, check if related findings exist in other areas and reference them
- Cross-references help users understand the interconnected nature of code quality issues

### Examples

- **Example 1**: When documenting `import-consistency`, add cross-reference to `package-structure` since poor import organization often reflects layer violations
- **Example 2**: When documenting `duplicate-code`, add cross-reference to `interface-compliance` since duplicated code often means interface design issues
- **Example 3**: When documenting `error-handling`, add cross-reference to `documentation-accuracy` since missing error types cause inconsistent error handling

### Severity Guidelines

- **CRITICAL**: Cross-references that affect architectural integrity
- **HIGH**: Cross-references that impact maintainability
- **MEDIUM**: Cross-references that improve understanding
- **LOW**: Optional cross-references for additional context

---

## Security


Check for common security vulnerabilities, secret handling, and input validation.

### What to Look For

- Hardcoded secrets or API keys
- SQL injection vulnerabilities
- Command injection in shell/command execution
- Path traversal vulnerabilities
- Missing input validation
- Insecure random number generation
- Unencrypted sensitive data storage
- Weak path traversal detection (incomplete pattern matching)
- Symlink-based path traversal bypasses
- Branch name sanitization in path construction

### Examples

- **Hardcoded secrets**: AWS keys, database passwords in source code
- **SQL injection**: String concatenation in SQL queries instead of parameterized queries
- **Command injection**: User input directly passed to shell/exec without sanitization
- **Weak path traversal detection**: `strings.Contains(s, "..")` fails to catch URL-encoded sequences like `..//` or `%2e%2e`
  ```go
  // BAD: Incomplete pattern matching
  func containsPathTraversal(s string) bool {
      return strings.Contains(s, "..") || strings.Contains(s, string(filepath.Separator)+".")
  }
  
  // GOOD: Clean path and detect traversal
  func containsPathTraversal(s string) bool {
      cleaned := filepath.Clean(s)
      if cleaned != s {
          return true  // Traversal detected
      }
      return strings.Contains(s, "..") || strings.Contains(s, "%2e%2e") || strings.Contains(s, "%2E%2E")
  }
  ```
- **Symlink bypass vulnerability**: `validatePathUnder` doesn't resolve symlinks before validation
  ```go
  // BAD: Symlink not resolved
  if under, err := IsPathUnder(base, target); err == nil && under { ... }
  
  // GOOD: Resolve symlinks first
  resolvedBase, _ := filepath.EvalSymlinks(base)
  resolvedTarget, _ := filepath.EvalSymlinks(target)
  if under, err := IsPathUnder(resolvedBase, resolvedTarget); err == nil && under { ... }
  ```
- **Missing project name validation**: Cross-project reference `<project>/<branch>` only validates branch component
  ```go
  // BAD: Project name not validated
  parts := strings.SplitN(spec, "/", 2)
  return parts[0], parts[1], nil  // parts[0] used without validation
  
  // GOOD: Validate both components
  parts := strings.SplitN(spec, "/", 2)
  if validation := domain.ValidateProjectName(parts[0]); validation.IsError() {
      return "", "", validation.Error
  }
  return parts[0], parts[1], nil
  ```
- **Weak branch name sanitization**: `filepath.Base()` and `filepath.Clean()` insufficient
  ```go
  // BAD: Weak sanitization
  safeBranchName := filepath.Base(branchName)
  safeBranchName = filepath.Clean(safeBranchName)
  worktreePath := filepath.Join(s.config.WorktreesDirectory, projectName, safeBranchName)
  
  // GOOD: Validate first, then sanitize
  if validation := domain.ValidateBranchName(branchName); validation.IsError() {
      return "", validation.Error
  }
  safeBranchName := filepath.Base(branchName)
  safeBranchName = filepath.Clean(safeBranchName)
  worktreePath := filepath.Join(s.config.WorktreesDirectory, projectName, safeBranchName)
  if err := infrastructure.ValidatePathUnder(s.config.WorktreesDirectory, worktreePath, "worktree", "worktrees"); err != nil {
      return "", err
  }
  ```

### Severity Guidelines

- **CRITICAL**: Hardcoded secrets, injection vulnerabilities, credential leaks, path traversal bypasses, symlink vulnerabilities
- **HIGH**: Missing input validation on user-controlled data, weak path traversal detection, unvalidated cross-project references
- **MEDIUM**: Weak cryptography or insecure defaults, insufficient sanitization
- **LOW**: Security best practice violations (no immediate vulnerability)

**See Finding Validation Checklist above.**

---

## Performance

Check for common performance anti-patterns, inefficient algorithms, and resource leaks.

### What to Look For

- N+1 queries in loops (N+1 problem)
- Inefficient string operations in loops
- Unnecessary memory allocations
- Missing database indexes (inferred)
- Resource leaks (unclosed files, connections)
- Synchronous operations that could be async
- Inefficient data structures for common operations
- Memory leaks from unbounded caches/maps
- Directory traversal without caching
- Repeated filesystem I/O operations
- O(n²) slice operations (missing pre-allocation)

### Examples

- **Memory leak - unbounded cache growth**: Repository cache grows indefinitely
  ```go
  // BAD: Unbounded map - memory leak
  type GoGitClientImpl struct {
      cache        map[string]*git.Repository // Grows forever
      cacheEnabled bool
  }
  
  // GOOD: Bounded LRU cache
  import "github.com/hashicorp/golang-lru/v2"
  
  type GoGitClientImpl struct {
      cache        *lru.Cache[string, *git.Repository] // Bounded cache
      cacheEnabled bool
  }
  
  func NewGoGitClient(cacheSize int, cacheEnabled ...bool) *GoGitClientImpl {
      cache, _ := lru.New[string, *git.Repository](cacheSize)
      return &GoGitClientImpl{cache: cache, cacheEnabled: true}
  }
  ```
- **N+1 filesystem query**: Lists all projects, then calls `ListWorktrees` for each
  ```go
  // BAD: O(n*m) complexity
  func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
      projects, err := s.projectService.ListProjects(ctx) // O(n)
      for _, project := range projects {
          worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath) // O(m) per project
          // Find match
      }
  }
  
  // GOOD: Parse path directly - O(1)
  func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
      relPath, _ := filepath.Rel(s.config.WorktreesDirectory, worktreePath)
      parts := strings.Split(relPath, string(filepath.Separator))
      projectName := parts[0]
      return s.projectService.DiscoverProject(ctx, projectName, nil)
  }
  ```
- **Double I/O operation**: Calling same method twice for same data
  ```go
  // BAD: ListWorktrees called twice
  worktrees, _ := cr.gitService.ListWorktrees(ctx, ctx.Path) // First call
  if len(worktrees) > 0 { ... }
  worktrees, _ = cr.gitService.ListWorktrees(ctx, ctx.Path) // Second call!
  for _, wt := range worktrees { ... }
  
  // GOOD: Cache result or call once
  worktrees, _ := cr.gitService.ListWorktrees(ctx, ctx.Path)
  if len(worktrees) > 0 {
      worktreeMap := make(map[string]bool)
      for _, wt := range worktrees {
          worktreeMap[wt.Branch] = true
      }
      // Use cached map instead of second call
  }
  ```
- **Directory traversal without caching**: O(n) filesystem calls every time
  ```go
  // BAD: Traverse filesystem every time
  func (cd *contextDetector) detectProjectContext(dir string) *domain.Context {
      currentDir := dir
      for {
          gitPath := filepath.Join(currentDir, ".git")
          if _, err := os.Stat(gitPath); err == nil { ... }
          parent := filepath.Dir(currentDir)
          if parent == currentDir { break }
          currentDir = parent
      }
  }
  
  // GOOD: Cache repository root paths
  type contextDetector struct {
      config        *domain.Config
      repoRootCache map[string]string // Cache: path → repo root
  }
  
  func (cd *contextDetector) detectProjectContext(dir string) *domain.Context {
      absDir, _ := filepath.Abs(dir)
      if repoRoot, exists := cd.repoRootCache[absDir]; exists {
          return &domain.Context{Type: domain.ContextProject, Path: repoRoot, ...}
      }
      // ... traversal logic ...
      if ctx != nil {
          cd.repoRootCache[absDir] = ctx.Path
      }
      return ctx
  }
  ```
- **O(n²) slice appends**: No pre-allocation causing repeated reallocations
  ```go
  // BAD: Repeated allocations
  func filterSuggestions(suggestions []string, partial string) []string {
      result := make([]string, 0) // No capacity
      for _, suggestion := range suggestions {
          if strings.HasPrefix(suggestion, partial) {
              result = append(result, suggestion) // Reallocates many times
          }
      }
      return result
  }
  
  // GOOD: Pre-allocate capacity
  func filterSuggestions(suggestions []string, partial string) []string {
      result := make([]string, 0, len(suggestions)) // Full capacity
      for _, suggestion := range suggestions {
          if strings.HasPrefix(suggestion, partial) {
              result = append(result, suggestion) // No reallocation
          }
      }
      return result
  }
  ```
- **Repeated os.Stat calls**: Multiple filesystem I/O in traversal
  ```go
  // BAD: Stat in loop
  func findMainRepoFromWorktree(worktreePath string) string {
      current := worktreePath
      for {
          if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
              if info, _ := os.Stat(filepath.Join(current, ".git")); info.IsDir() {
                  return current
              }
          }
          parent := filepath.Dir(current)
          if parent == current { break }
          current = parent
      }
  }
  
  // GOOD: Parse .git file directly (single I/O)
  func findMainRepoFromWorktree(worktreePath string) string {
      gitFilePath := filepath.Join(worktreePath, ".git")
      if content, err := os.ReadFile(gitFilePath); err == nil {
          contentStr := strings.TrimSpace(string(content))
          if strings.HasPrefix(contentStr, "gitdir:") {
              gitdirPath := strings.TrimPrefix(contentStr, "gitdir:")
              gitdirPath = strings.TrimSpace(gitdirPath)
              if strings.Contains(gitdirPath, "/.git/worktrees/") {
                  parts := strings.Split(gitdirPath, "/.git/worktrees/")
                  return filepath.Dir(filepath.Dir(parts[0]))
              }
          }
      }
      return worktreePath
  }
  ```

### Severity Guidelines

- **CRITICAL**: Resource leaks causing gradual degradation (memory leaks, unbounded growth)
- **HIGH**: O(n²) algorithm where O(n) exists and affects common operations, N+1 queries in critical paths
- **MEDIUM**: Inefficient operations in hot paths, missing caching for repeated operations
- **LOW**: Suboptimal data structure choice or minor performance issues

**See Finding Validation Checklist above.**

---

## Dead Code

Check for unused code, unreachable code, and commented-out production code.

### What to Look For

- Unused functions or methods
- Unreachable code after return statements
- Commented-out production code (large blocks)
- Unused imports
- Unused variables or constants
- Unreferenced files or entire modules

### Examples

- **Unused interfaces**: `GitRepositoryInterface` and `ProjectRepository` defined but never used
- **Unreachable code**: Code after `return` or infinite loop prevention
- **Commented code**: Large blocks of functional code commented out with `//` or `/* */`
- **Dead functions**: Helper functions defined but never called

### Severity Guidelines

- **CRITICAL**: Dead code in critical paths affecting functionality
- **HIGH**: Entire unused files or modules (>50 lines)
- **MEDIUM**: Unused functions or unreachable code blocks (10-50 lines)
- **LOW**: Unused variables, small helper functions, or commented debug code

**See Finding Validation Checklist above.**
