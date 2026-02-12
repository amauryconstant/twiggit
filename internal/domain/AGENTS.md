## Domain Layer
Layer: Business logic, entities, no external dependencies

## Context Types

```go
type ContextType int

const (
    ContextUnknown ContextType = iota
    ContextProject
    ContextWorktree
    ContextOutsideGit
)

type Context struct {
    Type        ContextType
    ProjectName string
    BranchName  string // Only for ContextWorktree
    Path        string
    Explanation string
}
```

**Context detection**: See `internal/infrastructure/AGENTS.md` for rules and resolution.

## Domain Model Pattern

```go
type Project struct {
    name      string
    path      string
    worktrees []*Worktree
}

func NewProject(name, path string) (*Project, error) {
    if name == "" {
        return nil, &ValidationError{Field: "name", Message: "cannot be empty"}
    }
    return &Project{name: name, path: path, worktrees: []*Worktree{}}, nil
}
// Immutable: getters only, no setters
```

## Validation Pipeline

Functional validation in `validation.go`:
- Composable validation rules
- Returns `ValidationError` with field + message
- Used in service layer before business logic

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}
```

## Error Handling

### Domain Layer Error Types

- **ValidationError**: Used for all validation failures across the domain layer
  - Constructor validation (NewProject, NewWorktree, etc.)
  - Config validation (Config.Validate)
  - Shell validation (shell type inference)
  - Pattern: `domain.NewValidationError("RequestType", "field", value, "message")`

- **GitRepositoryError**: Git repository operation errors
  - Repository access failures
  - Branch listing failures
  - Status retrieval failures
  - Pattern: `domain.NewGitRepositoryError(path, "message", cause)`

- **GitWorktreeError**: Git worktree operation errors
  - Worktree creation/deletion failures
  - Worktree listing failures
  - Pattern: `domain.NewGitWorktreeError(worktreePath, branchName, "message", cause)`

- **GitCommandError**: Git command execution errors (infrastructure layer, used as cause)
  - Failed git CLI commands
  - Non-zero exit codes
  - Includes command, args, exit code, stdout, stderr
  - Pattern: `domain.NewGitCommandError(command, args, exitCode, stdout, stderr, message, cause)`

- **ContextDetectionError**: Context detection failures
  - Directory access issues
  - Invalid paths
  - Pattern: `domain.NewContextDetectionError(path, "message", cause)`

- **ServiceError**: General service operation errors
  - Generic service failures
  - Includes service name, operation name, message
  - Pattern: `domain.NewServiceError(service, operation, message, cause)`

- **ShellError**: Shell service errors
  - Shell type validation failures
  - Config file issues
  - Wrapper generation/installation failures
  - Pattern: `domain.NewShellError(code, shellType, context)`

- **ResolutionError**: Path resolution errors
  - Failed identifier resolution
  - Invalid target paths
  - Includes optional suggestions
  - Pattern: `domain.NewResolutionError(target, context, message, suggestions, cause)`

- **ConflictError**: Operation conflict errors
  - Resource conflicts during operations
  - Includes resource type, identifier, operation
  - Pattern: `domain.NewConflictError(resource, identifier, operation, message, cause)`

### Error Wrapping Rules

1. **Domain constructors MUST return ValidationError for validation failures**
   - Use `domain.NewValidationError()` instead of `errors.New()`
   - Leverage `WithSuggestions()` for actionable guidance

2. **All domain error types implement Unwrap() for error chain support**
   - Use `%w` verb when wrapping errors
   - Enables `errors.As()` and `errors.Is()` for error type checking

3. **Error message format consistency**
   - Use lowercase "failed to" consistently
   - Include relevant context (path, branch, operation)
   - Keep messages concise but informative

## Error Types

```go
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
```

## Shell Detection

**Function**: `DetectShellFromEnv()` reads SHELL environment variable
**Returns**: ShellType (bash/zsh/fish) or error
**Error**: `ErrShellDetectionFailed` when SHELL not set or unsupported
**Pattern**: Case-insensitive path parsing (e.g., `/bin/BASH`, `/usr/local/Zsh/bin/zsh`)

## Shell Domain Model

```go
type Shell interface {
    Type() ShellType
    Path() string
    Version() string
}

type ShellType string

const (
    ShellBash ShellType = "bash"
    ShellZsh ShellType = "zsh"
    ShellFish ShellType = "fish"
)
```

**Design Rationale**: Shell domain model is minimal - only Type, Path, and Version. Wrapper template generation and config file detection are infrastructure concerns (see `internal/infrastructure/AGENTS.md`).

**Validation**: Use `IsValidShellType(shellType)` to validate shell types (exported function).

## Testing
- **Unit tests**: Testify suites with table-driven tests
- **Mocking**: Centralized mocks in test/mocks/
- **Focus**: Business logic, validation rules, edge cases
