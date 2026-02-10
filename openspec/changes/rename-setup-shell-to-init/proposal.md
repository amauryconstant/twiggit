## Why

The current `setup-shell` command uses a flag-based API (`--shell=bash`) that couples the command to specific shell types and auto-detects config file location. This limits flexibility for custom installations and makes the API less intuitive. The refactoring to `init` with positional config file argument simplifies the API, enables custom config file locations, and improves maintainability with proper block delimiters for wrapper replacement.

## What Changes

- **BREAKING**: Rename `setup-shell` command to `init`
- **BREAKING**: Change API from flag-based (`--shell=bash`) to positional argument (`init <config-file>`)
- **BREAKING**: Remove `--shell` flag (shell type inferred from config file path, optional override)
- Add block delimiters (`### BEGIN/END TWIGGIT WRAPPER`) to wrapper templates for proper replacement
- Add shell type inference from config file path (`.bash*` → bash, `.zsh*` → zsh, contains "fish" → fish)
- Update install.sh to detect shell, prompt user, and call `init` as orchestrator
- Shell service removes old wrapper blocks when `--force` is used

## Capabilities

### New Capabilities
- `shell-init`: Command for installing shell wrapper to specified config file with inferred shell type

### Modified Capabilities
- `shell-integration`: API changes from `setup-shell --shell=<type>` to `init <config-file>` with inferred shell type; validation now uses block delimiters instead of partial markers

## Impact

**Affected code:**
- `cmd/setup-shell.go` → deleted
- `cmd/setup-shell_test.go` → deleted
- `cmd/init.go` → created
- `cmd/root.go` → command registration updated
- `internal/domain/shell.go` → add block delimiters and inference function
- `internal/domain/shell_errors.go` → add inference error constant
- `internal/infrastructure/shell/service.go` → update InstallWrapper signature and add block handling
- `internal/services/shell_service.go` → update SetupShell with inference logic
- `test/e2e/setup_shell_test.go` → deleted
- `test/e2e/init_test.go` → created
- `test/integration/cli_commands_test.go` → update expected commands

**Affected APIs:**
- `setup-shell` command → removed
- `init` command → added with different API
- `ShellInfrastructure.InstallWrapper()` → signature changed to include explicit config file and force flag

**Affected dependencies:**
- install.sh → updated to orchestrate `init` command
