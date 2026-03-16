## Why

Users cannot easily integrate twiggit into automation scripts because output is human-oriented text only. This makes parsing unreliable and adds noise to logs. Additionally, bulk operations like `prune --all` provide no progress feedback during execution.

**Dependency:** This change depends on "error-clarity" being implemented first for consistent error and output patterns.

## What Changes

- Add `--output/-o` flag with `text` (default) and `json` formats for machine-readable output
- Add global `--quiet/-q` flag to suppress non-essential messages for cleaner scripting
- Add progress indicators during bulk operations (`prune --all`)
- Progress output goes to stderr, essential data to stdout
- Quiet mode suppresses success/hint messages but preserves errors and essential output

## Capabilities

### New Capabilities

- `output-formats`: Defines JSON and text output formats with formatter interface for commands
- `quiet-mode`: Controls suppression of non-essential output for scripting scenarios
- `progress-reporting`: Provides progress feedback during long-running bulk operations

### Modified Capabilities

- `command-flags`: Adds `--output/-o` flag (per-command) and `--quiet/-q` flag (global) to flag conventions
- `verbose-output`: Adds mutual exclusion rule between `--quiet` and `--verbose` (verbose wins)

## Impact

- `cmd/*.go`: Add output formatter interface, quiet mode checks, progress reporter
- `cmd/util.go`: Extend with `OutputFormatter` interface and `ProgressReporter`
- `cmd/root.go`: Add global `--quiet/-q` flag
- `cmd/list.go`: Implement JSON output for worktrees
- `cmd/prune.go`: Add progress reporting for bulk operations
- `internal/domain/git_types.go`: Ensure WorktreeInfo is JSON-serializable
