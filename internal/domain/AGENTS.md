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

## Testing
- **Unit tests**: Table-driven tests with Testify
- **Mocking**: Inline mocks (keep tests self-contained)
- **Focus**: Business logic, validation rules, edge cases
