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
Flags: `--source <branch>`, `-C, --cd`
Behavior: Create worktree, maintain current dir (unless -C outputs path to stdout)

### delete
Safety checks: Uncommitted changes, current worktree status
Flags: `-f, --force`, `--merged-only`, `-C, --cd`
Default behavior: Remove worktree + delete branch
Navigation: With -C from worktree context, outputs project root path; from project or outside git, outputs nothing

### cd
Output: Absolute path to worktree (for shell wrapper)
Flags: None (target required)
Behavior: Navigation via shell wrapper, escape hatch for builtin cd

### init
Optional: `[config-file]` (auto-detected if omitted)
Flags: `--check`, `-f, --force`, `--dry-run`, `--shell <bash|zsh|fish>` (alphabetical)
Behavior: Auto-detects shell/config file from SHELL env var, generates wrapper, adds to shell config
Usage: `twiggit init` | `twiggit init ~/.bashrc` | `twiggit init --shell=zsh`

## Verbose Output

Commands use `logv()` helper function for verbose output. See `cmd/util.go`.

**Verbosity levels:**
- `-v`: Level 1 - High-level operation flow
- `-vv`: Level 2 - Detailed parameters and intermediate steps
- No flag: Normal output only

**Output format:**
- Plain text, no color, no "DEBUG:" or "[VERBOSE]" prefixes
- Level 2 details indented with "  " prefix
- All verbose output goes to stderr, normal output to stdout

**logv() usage:**
```go
import "twiggit/cmd"

logv(cmd, 1, "Creating worktree for %s/%s", project, branch)
logv(cmd, 2, "  from branch: %s", source)
logv(cmd, 2, "  to path: %s", path)
```

**When to use level 1:**
- Major operation announcements (Creating, Deleting, Listing)
- High-level flow messages

**When to use level 2:**
- Detailed parameters (paths, branches, flags)
- Intermediate steps
- Configuration values

**Constraints:**
- Verbose output SHALL only appear in command layer (`cmd/*.go`)
- SHALL NOT add verbose output to service layer
- SHALL use user-focused language, not developer-focused
- SHALL NOT use "DEBUG:" prefix
