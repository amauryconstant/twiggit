## Domain Layer
Layer: Business logic, entities, no external dependencies

## Context Types

```go
type ContextType int // ContextUnknown, ContextProject, ContextWorktree, ContextOutsideGit

type Context struct {
    Type        ContextType
    ProjectName string
    BranchName  string  // Only for ContextWorktree
    Path        string
    Explanation string
}

type PathType int // PathTypeUnknown, PathTypeProject, PathTypeWorktree, PathTypeOutside

type ResolutionResult struct {
    TargetPath  string
    TargetType  PathType
    ProjectName string
    BranchName  string
}

type ResolutionSuggestion struct {
    Identifier   string
    DisplayText  string
    ResourceType PathType
}
```

**Detection:** See `internal/infrastructure/AGENTS.md` for rules.

## Core Types

| Type | Fields | Purpose |
|------|--------|---------|
| ProjectInfo | Name, Path, GitRepoPath, Worktrees, Branches, Remotes, DefaultBranch, IsBare, LastModified | Full project data |
| ProjectSummary | Name, Path, GitRepoPath | Lightweight listing |
| WorktreeInfo | Path, Branch, Commit, IsBare, IsDetached, Modified | Worktree details |
| Result[T] | Value, Error | Generic Result/Either pattern |

## Prune Types

```go
type PruneWorktreesRequest struct {
    Context          *Context
    Force            bool
    DeleteBranches   bool
    DryRun           bool
    AllProjects      bool
    SpecificWorktree string  // "project/branch"
}

type PruneWorktreesResult struct {
    DeletedWorktrees       []*PruneWorktreeResult
    SkippedWorktrees       []*PruneWorktreeResult
    ProtectedSkipped       []*PruneWorktreeResult
    UnmergedSkipped        []*PruneWorktreeResult
    CurrentWorktreeSkipped []*PruneWorktreeResult
    TotalDeleted, TotalSkipped, TotalBranchesDeleted int
    NavigationPath string  // Single-worktree prune
}
```

## Validation

```go
type ValidationError struct { /* private fields */ }

func NewValidationError(request, field, value, message string) *ValidationError
func (e *ValidationError) WithSuggestions([]string) *ValidationError  // immutable
func (e *ValidationError) WithContext(string) *ValidationError         // immutable
// Getters: Field(), Value(), Message(), Request(), Suggestions(), Context()
```

## Error Types

| Type | Constructor | IsNotFound() |
|------|-------------|--------------|
| ValidationError | `NewValidationError(request, field, value, message)` | - |
| GitRepositoryError | `NewGitRepositoryError(path, message, cause)` | ✅ |
| GitWorktreeError | `NewGitWorktreeError(worktreePath, branchName, message, cause)` | ✅ |
| GitCommandError | `NewGitCommandError(cmd, args, exitCode, stdout, stderr, msg, cause)` | - |
| ConfigError | `NewConfigError(path, message, cause)` | - |
| ContextDetectionError | `NewContextDetectionError(path, message, cause)` | - |
| ServiceError | `NewServiceError(service, operation, message, cause)` | - |
| WorktreeServiceError | `NewWorktreeServiceError(worktreePath, branchName, op, msg, cause)` | ✅ |
| ProjectServiceError | `NewProjectServiceError(projectName, projectPath, op, msg, cause)` | - |
| NavigationServiceError | `NewNavigationServiceError(target, ctx, op, msg, cause)` | - |
| ShellError | `NewShellError(code, shellType, context)` or `NewShellErrorWithCause(..., cause)` | - |
| ResolutionError | `NewResolutionError(target, ctx, msg, suggestions, cause)` | - |
| ConflictError | `NewConflictError(resource, identifier, operation, message, cause)` | - |

**All error types implement `Unwrap()` for error chain support.**

## Shell Types

```go
type ShellType string // ShellBash, ShellZsh, ShellFish

func IsValidShellType(ShellType) bool
func DetectShellFromEnv() (ShellType, error)  // reads SHELL env
func InferShellTypeFromPath(string) ShellType
```

**Error code:** `ErrShellDetectionFailed = "SHELL_DETECTION_FAILED"` (string constant)

## Suggestion Options

```go
type SuggestionOption func(*suggestionConfig)

func WithExistingOnly() SuggestionOption  // Filter to materialized worktrees only
```

Used by `ContextResolver.GetResolutionSuggestions()` for completion filtering.

## Hook Types

```go
type HookType string
const HookPostCreate HookType = "post-create"

type HookConfig struct {
    PostCreate *HookDefinition `toml:"post-create" koanf:"post-create"`
}

type HookDefinition struct {
    Commands []string `toml:"commands" koanf:"commands"`
}

type HookResult struct {
    HookType HookType
    Executed bool           // Were commands configured?
    Success  bool           // All commands succeeded?
    Failures []HookFailure  // Empty if success
}

type HookFailure struct {
    Command  string
    ExitCode int
    Output   string
}
```

## Config

```go
type Config struct {
    ProjectsDirectory   string
    WorktreesDirectory  string
    CompletionTimeout   time.Duration  // Default: 500ms
}
```
