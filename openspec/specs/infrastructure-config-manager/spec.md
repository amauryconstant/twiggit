# Capability: Configuration Manager

## Purpose

Load configuration from defaults, the XDG config file, and env-var
overrides. Expand `$VAR`, `${VAR}`, and `~` in path fields.

## Requirements

### Requirement: Load priority

The system SHALL load configuration in this order, each layer
overriding the previous:

1. `domain.DefaultConfig()` (built-in defaults)
2. `$HOME/.config/twiggit/config.toml` (XDG)
3. `TWIGGIT_*` environment variables (see `domain-config-types`)
4. CLI flags (handled in cmd layer)

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

The system SHALL resolve the config file path as
`$HOME/.config/twiggit/config.toml`. When `$XDG_CONFIG_HOME` is set,
the system SHALL use `$XDG_CONFIG_HOME/twiggit/config.toml` instead.

#### Scenario: Default XDG path

- **WHEN** `$XDG_CONFIG_HOME` is unset
- **THEN** config file SHALL be resolved at `$HOME/.config/twiggit/config.toml`

#### Scenario: Custom XDG path

- **WHEN** `$XDG_CONFIG_HOME = /custom/cfg`
- **THEN** config file SHALL be resolved at `/custom/cfg/twiggit/config.toml`

### Requirement: Path expansion

The system SHALL expand `$VAR`, `${VAR}`, and `~` in the following
fields only:

- `Config.ProjectsDirectory`
- `Config.WorktreesDirectory`
- `Config.Shell.Wrapper.BackupDir`

Other fields SHALL NOT be expanded.

#### Scenario: Tilde expansion

- **WHEN** config sets `worktrees_directory = "~/Worktrees"`
- **THEN** the loaded value SHALL be `/home/<user>/Worktrees`

#### Scenario: Env-var expansion

- **WHEN** `$WORKTREE_BASE` is exported and config sets
  `worktrees_directory = "$WORKTREE_BASE"`
- **THEN** the loaded value SHALL be the expanded path

#### Scenario: Unrelated field not expanded

- **WHEN** config sets `default_source_branch = "$MAIN"` (not a path
  field)
- **THEN** the loaded value SHALL remain the literal `"$MAIN"`
- **AND** system SHALL NOT error

### Requirement: Malformed expansion ignored

When an expansion target references an undefined env var (e.g.
`$UNDEFINED`), the system SHALL leave the literal substring in place
rather than fail the load. Failures SHALL be loud (logged) only when
the resulting path is unusable.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: TOML parse errors

When the config file exists but is not valid TOML, the system SHALL
return a `domain.ConfigError` (exit code 3) naming the file path and
the parse error message.

#### Scenario: Bad TOML

- **WHEN** config file contains malformed TOML
- **THEN** system SHALL return `ConfigError` with file path and parser message
- **AND** SHALL exit with code 3

### Requirement: Missing config file is OK

When the config file does not exist, the system SHALL fall back to
`DefaultConfig()` and SHALL NOT error.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Config immutability after load

`GetConfig()` SHALL return the loaded `*Config`. The returned pointer
SHALL be treated as immutable by consumers; modifications SHALL
require a reload.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Service injection

`ConfigManager` SHALL be injected via constructor into all services
that need config access. No global config singleton.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
