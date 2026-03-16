## Why

Config paths containing `$HOME`, `${HOME}`, or `~` fail validation because koanf loads them as literal strings. Users cannot use shell-style environment variable expansion in their config files, making configs less portable across environments.

## What Changes

- Add environment variable expansion for path fields in config
- Support `$VAR`, `${VAR}`, and `~` syntax
- Expand paths after TOML loading but before validation

## Capabilities

### New Capabilities

- `config-path-expansion`: Expand environment variables and tilde in config path fields during loading

### Modified Capabilities

- `path-resolution`: Path resolution will receive pre-expanded paths from config (no spec change needed - implementation detail)

## Impact

- `internal/infrastructure/config_manager.go`: Add `expandConfigPaths()` and `normalizeConfigPaths()` functions
- `internal/infrastructure/config_manager_test.go`: Add unit tests for expansion
- Affected config fields: `ProjectsDirectory`, `WorktreesDirectory`, `Shell.Wrapper.BackupDir`
