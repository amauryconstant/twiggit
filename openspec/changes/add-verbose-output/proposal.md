## Why

Ad-hoc debug output is scattered throughout the codebase with inconsistent formatting and no user control. Users cannot silence or adjust verbosity levels, and the "DEBUG:" prefix is developer-focused rather than user-friendly. A structured verbose output system provides users with visibility into operations when needed while maintaining clean normal output.

## What Changes

- Add persistent `--verbose` flag (`-v`) that can be specified multiple times for increasing verbosity
- Implement two verbosity levels:
  - `-v`: High-level operation flow (e.g., "Creating worktree for twiggit/test-branch")
  - `-vv`: Detailed parameters and intermediate steps (e.g., "  from branch: main", "  to path: /home/amaury/Worktrees/twiggit/test-branch")
- Create `cmd/util.go` with `logv()` helper function for verbose output
- Add verbose support to all CLI commands (create, delete, list, cd, setup-shell)
- Remove all `fmt.Fprintf(os.Stderr, "DEBUG: ...")` statements from service layer (internal/services/)
- Move relevant verbose output from service layer to command layer
- Format: plain text with indentation for level 2 details, no color, no "DEBUG:" prefix

## Non-Goals

- Structured logging (JSON, timestamps, component tags)
- Color-coded output
- Configuration file support for verbosity settings
- Logging to files
- Debug mode for internal development (use standard Go debugging tools instead)

## Capabilities

### New Capabilities
- `verbose-output`: User-controllable verbosity levels for CLI commands with structured output formatting

### Modified Capabilities
- None (verbose output is a cross-cutting enhancement that doesn't change core capability requirements)

## Impact

- **New files**: `cmd/util.go`
- **Modified files**: `cmd/root.go`, all command files (`create.go`, `delete.go`, `list.go`, `cd.go`, `setup-shell.go`)
- **Modified files**: `internal/services/worktree_service.go` (removing debug statements)
- **No API changes**: No breaking changes to external interfaces or command behavior
- **No new dependencies**: Uses existing Cobra and fmt packages
