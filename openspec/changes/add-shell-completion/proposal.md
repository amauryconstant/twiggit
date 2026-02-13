## Why

Users cannot tab-complete project or branch names when using twiggit commands. This forces manual typing of identifiers, reducing efficiency and increasing errors. Shell completion is a standard expectation for CLI tools and significantly improves the user experience.

## What Changes

- Add Carapace-based shell completion for `cd`, `create`, and `delete` commands
- Add completion for `--source` flag on `create` command
- Support progressive cross-project completion (`project/` → branches of that project) via `ActionMultiParts`
- Add existing-worktree-only filter for `delete` command (only show worktrees that exist)
- Shell completion scripts generated via Carapace's hidden `_carapace` command (10+ shells supported)
- Consolidate `internal/version/` into `cmd/version.go` (simplify package structure)

## Capabilities

### New Capabilities

- `shell-completion`: Tab completion for twiggit CLI commands via Carapace framework, including context-aware suggestions for projects, branches, and worktrees

### Modified Capabilities

- `path-resolution`: Extend completion suggestion system to support existing-worktree-only filtering

## Impact

- Adds Carapace dependency (`github.com/carapace-sh/carapace`) to go.mod
- Creates new `cmd/completion.go` for completion action helpers
- Modifies `internal/domain/context.go` to add `SuggestionOption` type
- Extends `internal/infrastructure/context_resolver.go` with `WithExistingOnly()` filter
- Updates `cmd/cd.go`, `cmd/create.go`, `cmd/delete.go` with Carapace wiring
- Deletes `internal/version/` package (consolidated into cmd)
- Updates ldflags in `.mise/config.toml` and `.goreleaser.yml` for version variables
