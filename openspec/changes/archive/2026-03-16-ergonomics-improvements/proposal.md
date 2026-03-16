## Why

Daily usage friction slows down workflows. Users type `twiggit list` and `twiggit delete` frequently but lack common Unix-style shortcuts (`ls`, `rm`). Confirmation prompts in `prune` require manual input even when the user is certain. Help text in `create` and `init` duplicates flag descriptions already shown by Cobra. These small frictions compound over time.

## What Changes

- Add alias `ls` for `list` command (Unix convention)
- Add alias `rm` for `delete` command (Unix convention)
- Add `--yes/-y` flag to `prune` command for auto-confirmation of prompts (distinct from `--force`)
- Add `-a` as short form for `--all` flag in `list` command
- Remove duplicate flag descriptions from Long descriptions in `create.go` and `init.go`
- Add `Examples:` sections to `list`, `delete`, and `create` commands

## Capabilities

### New Capabilities

- `command-aliases`: Command aliases for common Unix-style shortcuts (`ls`, `rm`)

### Modified Capabilities

- `command-flags`: Adding `-a` short flag for list --all, adding `--yes/-y` flag for prune auto-confirmation

## Impact

- **cmd/list.go**: Add `Aliases` field, add `-a` short flag, add Examples section
- **cmd/delete.go**: Add `Aliases` field, add Examples section
- **cmd/prune.go**: Add `--yes/-y` flag, modify confirmation logic to skip when flag set
- **cmd/create.go**: Remove duplicate flag descriptions from Long, add Examples section
- **cmd/init.go**: Remove duplicate flag descriptions from Long
- **test/e2e/**: New E2E tests for aliases and --yes flag behavior
