## Domain Layer
Layer: Business logic, entities, no external dependencies

**Note:** NO interfaces are defined in this package. All contracts are in `application/`. Domain only contains types and error types.

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
    Text        string
    Description string
    Type        PathType
    ProjectName string
    BranchName  string
    IsCurrent   bool  // Current worktree (sorting priority)
    IsDirty     bool  // Worktree has uncommitted changes
}
```

**Detection:** See `internal/infrastructure/AGENTS.md` for rules.

## Core Types

| Type | Fields | Purpose |
|------|--------|---------|
| ProjectInfo | Name, Path, GitRepoPath, Worktrees, Branches, Remotes, DefaultBranch, IsBare, LastModified | Full project data |
| ProjectSummary | Name, Path, GitRepoPath | Lightweight listing |
| WorktreeInfo | Path, Branch, Commit, IsDetached, Modified | Worktree details |
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

| Type | Constructor | Sentinel / `errors.Is` target |
|------|-------------|-------------------------------|
| ValidationError | `NewValidationError(request, field, value, message)` | — (terminal `Unwrap() = nil`) |
| GitRepositoryError | `NewGitRepositoryError(path, message, err)` | `ErrGitRepoNotFound` |
| GitWorktreeError | `NewGitWorktreeError(worktreePath, branchName, message, err)` | `ErrWorktreeNotFound` |
| GitCommandError | `NewGitCommandError(cmd, args, exitCode, stdout, stderr, msg, err)` | — |
| ConfigError | `NewConfigError(path, message, err)` | — |
| ContextDetectionError | `NewContextDetectionError(path, message, err)` | — |
| ServiceError | `NewServiceError(service, operation, message, err)` | — |
| WorktreeServiceError | `NewWorktreeServiceError(worktreePath, branchName, op, msg, err)` | `ErrWorktreeNotFound` |
| ProjectServiceError | `NewProjectServiceError(projectName, projectPath, op, msg, err)` | `ErrProjectNotFound` |
| NavigationServiceError | `NewNavigationServiceError(target, ctx, op, msg, err)` | `ErrResolutionNotFound` |
| ResolutionError | `NewResolutionError(target, ctx, msg, suggestions, err)` | `ErrResolutionNotFound` |
| ConflictError | `NewConflictError(resource, identifier, operation, message, err)` | — |
| ShellAlreadyInstalledError | `NewShellAlreadyInstalledError(shellType, context, err)` | `ErrShellAlreadyInstalled` |
| ShellNotInstalledError | `NewShellNotInstalledError(shellType, context, err)` | `ErrShellNotInstalled` |
| ShellInvalidTypeError | `NewShellInvalidTypeError(shellType, context, err)` | `ErrInvalidShellType` |
| ShellInferenceError | `NewShellInferenceError(shellType, context, err)` | `ErrShellInferenceFailed` |
| ShellDetectionError | `NewShellDetectionError(context, err)` | `ErrShellDetectionFailed` |
| ShellWrapperError | `NewShellWrapperError(shellType, op, context, err)` | `ErrWrapperGeneration` / `ErrWrapperInstallation` |
| ShellConfigError | `NewShellConfigError(path, context, err)` | `ErrConfigFileNotFound` |
| UsageError | `NewUsageError(message, err)` / `UsageWrap(err)` | `ErrUsageFlag` |

**All wrapper types implement `Unwrap() error` returning the `Err` field.**
**Wrapper types with a sentinel implement `Is(target error) bool` matching only that sentinel.**

> **Removed:** the substring-based `IsNotFound() bool` methods on
> `GitRepositoryError`, `GitWorktreeError`, and `WorktreeServiceError` are
> gone. Identify NotFound conditions via
> `errors.Is(err, domain.ErrXNotFound)` instead.
> `ShellError` (single struct with `Code string`) was replaced by the
> seven concrete shell subtypes listed above; the `code` field is gone.
> The legacy `cause` parameter is now `err`; all constructors take
> `err error` as the last argument.

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
```

Concrete options (e.g., `WithExistingOnly`) are defined in the
`infrastructure` package, where the private `suggestionConfig` type
lives. Used by `ContextResolver.GetResolutionSuggestions()` for
completion filtering.

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
