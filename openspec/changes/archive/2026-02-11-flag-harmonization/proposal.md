## Why

Multiple inconsistencies exist between command flags, their descriptions, and documentation. The `create` command has a zombie `--cd` flag that doesn't work, `delete` uses `--change-dir` while `create` would use `--cd`, and several implemented flags are undocumented. This harms user experience and makes shell wrapper integration incomplete.

## What Changes

- Remove zombie `--cd` string flag from `create` command
- Add `-C, --cd` boolean flag to `create` command (outputs path to stdout for shell wrapper)
- Rename `--change-dir` to `--cd` in `delete` command (keep `-C` short form)
- Add `-f` short form for `--force` flag in `delete` and `init` commands
- Implement navigation logic for `delete -C`: navigate to project root when deleting from worktree context, no change when from project or outside git
- Update shell wrapper templates (bash/zsh/fish) to handle `create -C` and `delete -C` flags
- Update all `Long` descriptions to document all implemented flags
- Update cmd/AGENTS.md to match implemented reality
- Reorder init command flags alphabetically

## Capabilities

### New Capabilities
- `command-flags`: Standardized conventions for command-line flags including naming, short forms, documentation, and shell wrapper integration

### Modified Capabilities
None (spec-level behavior unchanged, only implementation details)

## Impact

- cmd/create.go: Replace zombie flag with `-C, --cd` boolean flag, implement path output behavior
- cmd/delete.go: Rename `--change-dir` to `--cd`, add `-f` for `--force`, implement context-aware navigation
- cmd/init.go: Add `-f` for `--force`, reorder flags alphabetically
- internal/infrastructure/shell/service.go: Update bash/zsh/fish wrapper templates to capture `-C` flag for create/delete
- cmd/AGENTS.md: Update all command specifications to match implemented flags
- cmd/*_test.go: Add unit tests for new `-C` flag behavior and `-f` short form
- test/integration/cli_commands_test.go: Add wrapper integration tests
- test/e2e/create_test.go: Add E2E test for `create -C` path output
- test/e2e/delete_test.go: Add E2E test for `delete -C` navigation
