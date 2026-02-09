## Command Structure
- **Root command**: `cmd/root.go` - service container setup
- **Pattern**: Cobra commands → services → infrastructure
- **Error handling**: Centralized in `cmd/error_handler.go`

## Cobra Command Pattern
```go
var cmd = &cobra.Command{
    Use:   "list",
    Short: "List worktrees",
    RunE: func(cmd *cobra.Command, args []string) error {
        worktreeService := di.GetWorktreeService()
        result, err := worktreeService.ListWorktrees(context)
        if err != nil {
            return formatError(err)
        }
        return output(result)
    },
}
```

## Context-Aware Behavior
Commands adapt based on detected context (Project, Worktree, Outside git).
See `internal/infrastructure/AGENTS.md` for detection rules and resolution.

**Command Adaptation:**
- **From project**: List worktrees for current project
- **From worktree**: List worktrees for current project
- **Outside git**: Require explicit project/branch specification

## Testing
- **Unit tests**: None (cmd package tested via E2E only)
- **E2E tests**: `test/e2e/<command>_test.go`
- **Contract tests**: `cmd/contract_test.go` verifies service integration

## Error Handling
Error formatting centralized in `cmd/error_formatter.go`. Error wrapping pattern: `fmt.Errorf("action failed: %w", err)`.

## Command Specifications

### list
Output: Tabular format with branch, last commit, status (clean/dirty)
Flags: `--all` (show all projects, override context)

### create
Required: Project name (inferred), branch name, source branch (default: main)
Flags: `--source <branch>`, `-C / --change-dir`
Behavior: Create worktree, maintain current dir (unless -C)

### delete
Safety checks: Uncommitted changes, current worktree status
Flags: `--keep-branch`, `--force`, `--merged-only`, `-C / --change-dir`
Default behavior: Remove worktree + delete branch

### cd
Output: Absolute path to worktree (for shell wrapper)
Flags: None (target required)
Behavior: Navigation via shell wrapper, escape hatch for builtin cd

### setup-shell
Required: `--shell <bash|zsh|fish>`
Behavior: Generate wrapper, add to shell config, warn about builtin override
