## Command Structure
- **Root command**: `cmd/root.go` - service container setup
- **Pattern**: Cobra commands → services → infrastructure
- **Error handling**: Centralized in `cmd/error_handler.go` (handling logic) and `cmd/error_formatter.go` (formatting)

## ServiceContainer

Services available to commands (defined in `cmd/root.go`):

```go
type ServiceContainer struct {
    WorktreeService   application.WorktreeService
    ProjectService    application.ProjectService
    NavigationService application.NavigationService
    ContextService    application.ContextService
    ShellService      application.ShellService
}
```

**Note**: GitClient is NOT in ServiceContainer - it's internal to service implementations. Commands use WorktreeService methods (`BranchExists`, `IsBranchMerged`, `GetWorktreeByPath`) instead of accessing GitClient directly.

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
- **Error handling**: `cmd/error_handler.go` - determines error type, formats for CLI output
- **Error formatting**: `cmd/error_formatter.go` - formats error messages for display
- **Pattern**: `fmt.Errorf("action failed: %w", err)` for wrapping

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

### prune
Purpose: Delete merged worktrees for post-merge cleanup
Args: `[project/branch]` (optional, specific worktree to prune)
Flags: `-n, --dry-run`, `-f, --force`, `--delete-branches`, `-a, --all`
Behavior:
- Context-aware: Infers project from current directory (worktree > project > outside git)
- `--dry-run`: Preview what would be deleted without making changes
- `--force`: Bypass uncommitted changes safety check and bulk confirmation
- `--delete-branches`: Also delete corresponding git branches after worktree removal
- `--all`: Prune across all projects (requires confirmation unless --force)
- Protected branches (main, master, develop, staging, production) are never deleted
- Outputs navigation path to stdout for single-worktree prune (for shell wrapper)
Navigation: Single worktree prune outputs project directory path; bulk prune outputs nothing

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

## Shell Completion

Carapace integration provides shell completion for all commands.

**Hidden command:** `twiggit _carapace <shell>` generates completion scripts.

| Shells | bash, zsh, fish, nushell, elvish, powershell, tcsh, oil, xonsh, cmd-clink |
|--------|-----------------------------------------------------------------------------|

**Usage:**
```bash
# Bash
source <(twiggit _carapace bash)
# Zsh
source <(twiggit _carapace zsh)
```

**Implementation:** `cmd/suggestions.go` provides action helpers:
- `actionWorktreeTarget(config, opts...)` - Positional completion with `ActionMultiParts("/")`
- `actionBranches(config)` - Branch completion for `--source` flag
- `.Cache(5s)` - 5-second cache for performance
- `.Timeout(timeout)` - Graceful degradation for slow git ops

**Wiring pattern:**
```go
carapace.Gen(cmd).PositionalCompletion(
    actionWorktreeTarget(config, WithExistingOnly()),  // for delete/prune
)
```
