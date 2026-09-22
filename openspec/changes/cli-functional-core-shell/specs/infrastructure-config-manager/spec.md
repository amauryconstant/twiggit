# Spec Delta

## MODIFIED Requirements

### Requirement: Load priority

The system SHALL load configuration in this order, each layer overriding the previous:

1. `core.DefaultConfig()` (built-in defaults)
2. `$XDG_CONFIG_HOME/twiggit/config.toml` (XDG; falls back to `$HOME/.config/twiggit/config.toml`)
3. `TWIGGIT_*` environment variables (see `domain-config-types`)
4. CLI flags (handled in the cmd layer)

The previous `domain.DefaultConfig()` is removed; `core.DefaultConfig()` is the canonical defaults function.

#### Scenario: Layered load
- **WHEN** config file sets `DefaultSourceBranch = "develop"`
- **AND** no env override is present
- **THEN** the loaded `Config.DefaultSourceBranch` SHALL be `"develop"`
- **AND** unset fields SHALL fall back to defaults

#### Scenario: Env overrides file
- **WHEN** config file sets `DefaultSourceBranch = "develop"`
- **AND** `TWIGGIT_DEFAULT_SOURCE_BRANCH = "trunk"` is exported
- **THEN** the loaded value SHALL be `"trunk"`

### Requirement: XDG config path

The system SHALL resolve the config file path as `$HOME/.config/twiggit/config.toml`. When `$XDG_CONFIG_HOME` is set, the system SHALL use `$XDG_CONFIG_HOME/twiggit/config.toml` instead.

#### Scenario: Default XDG path
- **WHEN** `$XDG_CONFIG_HOME` is unset
- **THEN** config file SHALL be resolved at `$HOME/.config/twiggit/config.toml`

#### Scenario: Custom XDG path
- **WHEN** `$XDG_CONFIG_HOME = /custom/cfg`
- **THEN** config file SHALL be resolved at `/custom/cfg/twiggit/config.toml`

### Requirement: Path expansion

The system SHALL expand `$VAR`, `${VAR}`, and `~` in the following fields only:

- `Config.ProjectsDirectory`
- `Config.WorktreesDirectory`
- `Config.Shell.Wrapper.BackupDir`

Other fields SHALL NOT be expanded.

#### Scenario: Tilde expansion
- **WHEN** config sets `worktrees_directory = "~/Worktrees"`
- **THEN** the loaded value SHALL be `/home/<user>/Worktrees`

#### Scenario: Env-var expansion
- **WHEN** `$WORKTREE_BASE` is exported and config sets `worktrees_directory = "$WORKTREE_BASE"`
- **THEN** the loaded value SHALL be the expanded path

#### Scenario: Unrelated field not expanded
- **WHEN** config sets `default_source_branch = "$MAIN"` (not a path field)
- **THEN** the loaded value SHALL remain the literal `"$MAIN"`
- **AND** system SHALL NOT error

### Requirement: Malformed expansion ignored

When an expansion target references an undefined env var (e.g. `$UNDEFINED`), the system SHALL leave the literal substring in place rather than fail the load. Failures SHALL be loud (logged) only when the resulting path is unusable.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: TOML parse errors

When the config file exists but is not valid TOML, the system SHALL return a `*core.OperationError` wrapping the koanf parse error with `Op = "config.load"`. The previous `domain.ConfigError` (exit code 3) is removed; parse failures exit 1 via `ExitError` because `*core.OperationError` dispatches via `ExitCodeFor`.

#### Scenario: Bad TOML
- **WHEN** config file contains malformed TOML
- **THEN** system SHALL return a `*core.OperationError` with `Op = "config.load"` and the file path and parser message
- **AND** SHALL exit with code 1 (per the 3-code canonical mapping)

### Requirement: Missing config file is OK

When the config file does not exist, the system SHALL fall back to `core.DefaultConfig()` and SHALL NOT error.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Config immutability after load

`Config()` SHALL return the loaded `*core.Config`. The returned pointer SHALL be treated as immutable by consumers; modifications SHALL require a reload. `Config.ProtectedBranches()` SHALL return a defensive copy via `slices.Clone` so callers cannot mutate the underlying slice.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: ProtectedBranches returns a defensive copy
- **WHEN** the caller appends to the slice returned by `cfg.ProtectedBranches()`
- **THEN** a subsequent call to `cfg.ProtectedBranches()` SHALL return the original slice without the appended element

### Requirement: Service injection

`Config` SHALL be injected via the `cmdutil.Factory.Config func() (*core.Config, error)` lazy function field into all commands that need config access. No global config singleton. The previous `ConfigManager` is removed; the lazy function replaces the per-service constructor-injection pattern. The loader respects `NO_COLOR=1` when setting lipgloss style state via the Factory's lazy Logger.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: NO_COLOR respected
- **WHEN** `NO_COLOR=1` is set
- **THEN** the Factory's lazy Logger SHALL NOT enable colored output
