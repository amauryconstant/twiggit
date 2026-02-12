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
