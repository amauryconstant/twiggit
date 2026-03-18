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

**Exit codes:**

| Code | Constant | Meaning | Use case |
|------|----------|---------|----------|
| 0 | ExitCodeSuccess | Success | Normal completion |
| 1 | ExitCodeError | General error | Unclassified errors, panics |
| 2 | ExitCodeUsage | Usage error | Invalid command syntax |
| 3 | ExitCodeConfig | Configuration error | Invalid or missing config |
| 4 | ExitCodeGit | Git operation error | Git command failures |
| 5 | ExitCodeValidation | Validation error | Input validation failures |
| 6 | ExitCodeNotFound | Resource not found | Project/worktree not found |

**Error formatting:**
- Messages are user-friendly without internal operation names
- Includes actionable hints for recovery (e.g., "Use 'twiggit list' to see available worktrees")
- Error categories map to specific exit codes via `GetExitCodeForError()`

**Debug mode:**
- Set `TWIGGIT_DEBUG=1` to enable internal details and stack traces
- Shows full error context including operation names when set
- Stack traces displayed for panics only in debug mode

**Panic recovery:**
- Unexpected panics caught in main.go via defer/recover
- Displays "Internal error: <panic value>" to stderr
- Exits with code 1

**Implementation:**
- `cmd/error_handler.go` - error categorization and exit code mapping
- `cmd/error_formatter.go` - user-friendly message formatting with hints
- Pattern: `fmt.Errorf("action failed: %w", err)` for wrapping

### Explicit Error Formatter Pattern

The error formatter uses an explicit strategy pattern with `errors.As()` matching instead of reflection.

**Core types:**
```go
type matcherFunc func(error) bool      // Checks if error matches a type
type formatterFunc func(error) string   // Formats error into user-friendly message
```

**Matcher functions** (use `errors.As()` for type matching):
- `isValidationError(err)` - matches `domain.ValidationError`
- `isWorktreeError(err)` - matches `domain.WorktreeServiceError`
- `isProjectError(err)` - matches `domain.ProjectServiceError`
- `isServiceError(err)` - matches `domain.ServiceError`

**Registration pattern** (in `NewErrorFormatterWithOptions`):
```go
formatter.register(isValidationError, formatValidationError)
formatter.register(isWorktreeError, formatWorktreeError)
formatter.register(isProjectError, formatProjectError)
formatter.register(isServiceError, formatServiceError)
```

**Important:** Registration order determines priority - more specific errors first.

**Adding a new error type:**
1. Create matcher function in `cmd/error_formatter.go`:
   ```go
   func isMyCustomError(err error) bool {
       var target *domain.MyCustomError
       return errors.As(err, &target)
   }
   ```
2. Create formatter function:
   ```go
   func formatMyCustomError(err error) string {
       customErr := func() *domain.MyCustomError {
           target := &domain.MyCustomError{}
           _ = errors.As(err, &target)
           return target
       }()
       return fmt.Sprintf("Error: %s\n", customErr.Error())
   }
   ```
3. Register in `NewErrorFormatterWithOptions` before generic `formatServiceError`

## Command Specifications

### list
Alias: `ls` (Unix-style shortcut)
Output: Tabular format with branch, last commit, status (clean/dirty) or JSON for scripting
Flags:
- `--all/-a` (show all projects, override context)
- `--output/-o <format>`: Output format: `text` (default) or `json`
- JSON output structure: `{"worktrees": [{"branch": "...", "path": "...", "status": "clean|modified|detached"}]}`
- JSON output uses stdout for data, stderr for errors/verbose messages

### create
Required: Project name (inferred), branch name, source branch (default: main)
Flags: `--source <branch>`, `-C, --cd`
Behavior: Create worktree, execute post-create hooks if `.twiggit.toml` configured, display hook failure warnings
Output: Worktree info + hook warnings (if any)

### delete
Alias: `rm` (Unix-style shortcut)
Safety checks: Uncommitted changes, current worktree status
Flags: `-f, --force`, `--merged-only`, `-C, --cd`
Default behavior: Remove worktree + delete branch
Navigation: With -C from worktree context, outputs project root path; from project or outside git, outputs nothing

### cd
Output: Absolute path to worktree (for shell wrapper)
Flags: None (target required)
Behavior: Navigation via shell wrapper, escape hatch for builtin cd

### init
Default: Print shell wrapper to stdout (eval-safe, no metadata)
Optional: `[shell]` (bash|zsh|fish, auto-detected from $SHELL if omitted)
Flags: `-i, --install` (file mode), `-c, --config <path>` (requires --install), `-f, --force` (requires --install)
Behavior:
  - Default (no flags): Print wrapper to stdout for eval-based activation
  - With `--install`: Write wrapper to shell config file
Usage: `eval "$(twiggit init)"` | `twiggit init bash` | `twiggit init --install` | `twiggit init zsh --install -c ~/.zshrc`

### prune
Purpose: Delete merged worktrees for post-merge cleanup
Args: `[project/branch]` (optional, specific worktree to prune)
Flags: `-n, --dry-run`, `-f, --force`, `-y, --yes`, `--delete-branches`, `-a, --all`
Behavior:
- Context-aware: Infers project from current directory (worktree > project > outside git)
- `--dry-run`: Preview what would be deleted without making changes
- `--force`: Bypass uncommitted changes safety check and bulk confirmation
- `--yes/-y`: Auto-confirm prompts (keeps safety checks, distinct from --force)
- `--delete-branches`: Also delete corresponding git branches after worktree removal
- `--all`: Prune across all projects (requires confirmation unless --yes or --force)
- Protected branches (main, master, develop, staging, production) are never deleted
- Progress reporting: Bulk operations (`--all` or no specific target) report progress to stderr
- Outputs navigation path to stdout for single-worktree prune (for shell wrapper)
- Progress is suppressed in quiet mode
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

## Quiet Mode

Global `--quiet/-q` flag suppresses non-essential output for scripting scenarios. Available on all commands.

**Behavior:**
- Suppresses success messages (e.g., "Created worktree...")
- Suppresses hint messages
- Preserves error output to stderr
- Preserves essential output (paths for `-C` mode)
- Suppressed by `--verbose` flag (verbose wins over quiet)

**Implementation:**
- Use `isQuiet(cmd)` helper from `cmd/util.go` to check quiet flag
- Check before outputting success/hint messages
- `ProgressReporter` automatically respects quiet mode

**Use case:** Cleaner automation scripts where only errors matter
```bash
# Script example - only care about failures
if ! twiggit list --quiet; then
  echo "Error listing worktrees"
  exit 1
fi
```

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

**Enhanced Features:**
| Feature | Description | Config |
|---------|-------------|--------|
| Fuzzy matching | Case-insensitive subsequence matching for partial input | `navigation.fuzzy_matching` |
| Smart sorting | Current worktree → default branch → alphabetical | Automatic |
| Enhanced descriptions | Remote tracking info, relative dates (e.g., "2 days ago") | Automatic |
| Status indicators | ⚠ prefix for dirty current worktree only | Automatic |
| Exclusion patterns | Filter noisy branches/projects via glob patterns | `completion.exclude_branches`, `completion.exclude_projects` |
| Progressive completion | Auto "/" suffix on project suggestions | Automatic |
| Cross-project | Complete branches from other projects via `project/branch` syntax | Automatic |

**Configuration:**
```toml
[completion]
exclude_branches = ["dependabot/*", "renovate/*"]
exclude_projects = ["archive/*"]
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

**Shell Plugins:** See `contrib/bash/`, `contrib/zsh/`, `contrib/fish/` for ready-to-use integrations with navigation.
