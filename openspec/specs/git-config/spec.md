# Capability: Git Config

## Purpose

Documents config loading migrated from `internal/infrastructure/config_manager.go` to `internal/config/manager.go`, preserving the precedence and XDG-resolution contract while moving the package boundary.

## Requirements

### Requirement: Config loads via koanf from $XDG_CONFIG_HOME/twiggit/config.toml

The `internal/config/manager.go` loader SHALL use koanf to read the TOML config file. The file path SHALL resolve as `$XDG_CONFIG_HOME/twiggit/config.toml` when `XDG_CONFIG_HOME` is set, otherwise `$HOME/.config/twiggit/config.toml`.

#### Scenario: XDG_CONFIG_HOME overrides default

- **WHEN** `XDG_CONFIG_HOME=/custom/cfg` and the file `/custom/cfg/twiggit/config.toml` exists
- **THEN** the loader SHALL read from `/custom/cfg/twiggit/config.toml`

#### Scenario: Default config path used when XDG unset

- **WHEN** `XDG_CONFIG_HOME` is unset and the file `$HOME/.config/twiggit/config.toml` exists
- **THEN** the loader SHALL read from `$HOME/.config/twiggit/config.toml`

### Requirement: Merge order: defaults → file → env (TWIGGIT_*) → flags

The loader SHALL merge configuration in this exact order, each layer overriding the previous:
1. `core.DefaultConfig()` (built-in defaults)
2. The TOML file at the resolved XDG path
3. `TWIGGIT_*` environment variables
4. CLI flags (handled in the cmd layer, not in `internal/config/`)

#### Scenario: Environment overrides file

- **WHEN** the file sets `ProjectsDirectory = "$HOME/Projects"` and `TWIGGIT_PROJECTS_DIRECTORY` is set in the environment
- **THEN** the loaded config SHALL reflect the environment value

#### Scenario: File overrides defaults

- **WHEN** the file sets `WorktreesDirectory = "$HOME/Worktrees"` and the default is `$HOME/Worktrees`
- **THEN** the loaded config SHALL equal the file value (no-op, identical)

### Requirement: Missing config file falls back to defaults (not an error)

When the resolved config file does not exist, the loader SHALL fall back to `core.DefaultConfig()` and SHALL NOT return an error.

#### Scenario: No config file is not an error

- **WHEN** `$XDG_CONFIG_HOME/twiggit/config.toml` does not exist
- **THEN** `Config()` returns `(&core.DefaultConfig(), nil)` with no error

#### Scenario: Invalid TOML is an error

- **WHEN** the file exists but is not valid TOML
- **THEN** the loader SHALL return a `*core.OperationError` wrapping the koanf parse error

### Requirement: ProtectedBranches returns slices.Clone (defensive copy)

`Config.ProtectedBranches()` SHALL return a defensive copy via `slices.Clone` so callers cannot mutate the underlying slice.

#### Scenario: Mutating the returned slice does not affect the loaded config

- **WHEN** the caller appends to the slice returned by `cfg.ProtectedBranches()`
- **THEN** a subsequent call to `cfg.ProtectedBranches()` SHALL return the original slice without the appended element

### Requirement: Config error contract

Config-loading errors SHALL be lowercase without trailing punctuation (e.g., `core: parse failure at /path/to/config.toml`). Sentinel matching for not-found config SHALL use `errors.Is(err, core.ErrConfigNotFound)` (or the equivalent `core.NewConfigError` walk) — never `strings.Contains(err.Error(), "config")` or similar substring matching.

#### Scenario: Config error lowercase and sentinel-match contracts hold

- **WHEN** the loader fails to parse `config.toml`
- **THEN** the wrapped error's message SHALL be lowercase without trailing period
- **AND** the loader SHALL return a `*core.OperationError` walking to the typed-walk

### Requirement: Config exposes [rebase] section

The merged config SHALL expose a `[rebase]` section typed as `core.RebaseConfig` with two fields: `FetchOnAll bool` (default `false`) and `ConflictPolicy string` (default `"stop"`). The loader SHALL honour the same koanf merge order (defaults → file → env → flags) documented in the merge-order requirement.

#### Scenario: Config defaults populate [rebase]

- **WHEN** no config file, env var, or flag overrides `[rebase]`
- **THEN** the loaded config SHALL set `Rebase.FetchOnAll = false` and `Rebase.ConflictPolicy = "stop"`

#### Scenario: TOML [rebase] block overrides defaults

- **WHEN** the config file contains:
  ```toml
  [rebase]
  fetch_on_all = true
  conflict_policy = "stop"
  ```
- **THEN** the loaded config SHALL set `Rebase.FetchOnAll = true`
- **AND** the loader SHALL NOT return an error

#### Scenario: TWIGGIT_REBASE__FETCH_ON_ALL overrides file

- **WHEN** the file sets `fetch_on_all = false` and `TWIGGIT_REBASE__FETCH_ON_ALL=true` is set in the environment
- **THEN** the loaded config SHALL set `Rebase.FetchOnAll = true`

### Requirement: Config exposes [sync] section

The merged config SHALL expose a `[sync]` section typed as `core.SyncConfig` with three fields: `DefaultRemote string` (default `"origin"`), `PruneRemoteRefs bool` (default `true`), and `RebaseAfterSync bool` (default `false`).

#### Scenario: Config defaults populate [sync]

- **WHEN** no config file, env var, or flag overrides `[sync]`
- **THEN** the loaded config SHALL set `Sync.DefaultRemote = "origin"`, `Sync.PruneRemoteRefs = true`, `Sync.RebaseAfterSync = false`

#### Scenario: TOML [sync] block overrides defaults

- **WHEN** the config file contains:
  ```toml
  [sync]
  default_remote = "upstream"
  prune_remote_refs = false
  rebase_after_sync = true
  ```
- **THEN** the loaded config SHALL reflect all three values
- **AND** the loader SHALL NOT return an error