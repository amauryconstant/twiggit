# Spec Delta

## ADDED Requirements

### Requirement: Config exposes [rebase] section

The merged config SHALL expose a `[rebase]` section typed as
`core.RebaseConfig` with two fields: `FetchOnAll bool` (default
`false`) and `ConflictPolicy string` (default `"stop"`). The loader
SHALL honour the same koanf merge order (defaults → file → env → flags)
documented in the merge-order requirement.

#### Scenario: Config defaults populate [rebase]

- **WHEN** no config file, env var, or flag overrides `[rebase]`
- **THEN** the loaded config SHALL set `Rebase.FetchOnAll = false` and
  `Rebase.ConflictPolicy = "stop"`

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

- **WHEN** the file sets `fetch_on_all = false` and
  `TWIGGIT_REBASE__FETCH_ON_ALL=true` is set in the environment
- **THEN** the loaded config SHALL set `Rebase.FetchOnAll = true`

### Requirement: Config exposes [sync] section

The merged config SHALL expose a `[sync]` section typed as
`core.SyncConfig` with three fields: `DefaultRemote string` (default
`"origin"`), `PruneRemoteRefs bool` (default `true`), and
`RebaseAfterSync bool` (default `false`).

#### Scenario: Config defaults populate [sync]

- **WHEN** no config file, env var, or flag overrides `[sync]`
- **THEN** the loaded config SHALL set `Sync.DefaultRemote = "origin"`,
  `Sync.PruneRemoteRefs = true`, `Sync.RebaseAfterSync = false`

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