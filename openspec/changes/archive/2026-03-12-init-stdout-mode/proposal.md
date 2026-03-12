## Why

The current `init` command writes to a shell config file by default, making `eval`-based activation awkward. Users who want instant activation without modifying files must use `--dry-run` and manually parse around metadata lines. This friction discourages the cleaner eval-based workflow.

Additionally, the current flag structure is complex: `[config-file]` positional arg, `--shell` flag, `--dry-run`, `--check`, and `--force`. This can be simplified by inverting the default behavior.

## What Changes

**BREAKING**: Default behavior changes from file write to stdout output.

- Default mode: Print shell wrapper to stdout (eval-safe, no metadata)
- Add `-i, --install` flag to enable file installation mode
- Add `-c, --config` flag for custom config file path (requires `--install`)
- Positional argument changes from `[config-file]` to `[shell]` (bash|zsh|fish)
- Remove `--dry-run` flag (stdout is the default, making this redundant)
- Remove `--shell` flag (replaced by positional `[shell]` argument)
- Remove `--check` flag (no persistent "installed" state to check when using eval)

## Capabilities

### New Capabilities

None - this is a refactor of existing functionality.

### Modified Capabilities

- `shell-init`: Default stdout output mode, new flag semantics (`--install`, `--config`, positional `[shell]`)

## Impact

**cmd/init.go** - Major rewrite:
- Change positional arg from `[config-file]` to `[shell]`
- Add `-i, --install` and `-c, --config` flags
- Remove `--dry-run`, `--shell`, `--check` flags
- Route to `GenerateWrapper()` (stdout) or `SetupShell()` (--install)

**domain/shell_requests.go** - Minor:
- Remove `DryRun` field from `SetupShellRequest`

**domain/shell_results.go** - Minor:
- Remove `DryRun` field from `SetupShellResult`

**service/shell_service.go** - Minor:
- Remove `DryRun` branch from `SetupShell()` (handled at cmd layer)

**application/interfaces.go** - Optional:
- Consider removing `ValidateInstallation` from `ShellService` interface

**infrastructure/shell_infra.go** - None:
- Existing methods work as-is
