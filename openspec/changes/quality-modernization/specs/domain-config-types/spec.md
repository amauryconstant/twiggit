# Spec Delta

## ADDED Requirements

### Requirement: Deprecated ServiceConfig fields are tolerated on load

The `ServiceConfig` struct SHALL continue to expose the fields
`CacheEnabled`, `ConcurrentOps`, `MaxConcurrent`, `Shell.Timeout`,
and `Shell.Wrapper.BackupDir` for backward compatibility with
existing configuration files. The configuration loader SHALL
accept these fields when present, log a `slog.Warn` with the
field name and a "deprecated; ignored" message, and SHALL NOT
use the values for any runtime decision.

#### Scenario: Existing config file with deprecated fields loads without error
- **WHEN** the user's `~/.config/twiggit/config.toml` contains
  `[service]` with `cache_enabled = true` and
  `[shell]` with `timeout = "500ms"`
- **THEN** the configuration loads successfully, both
  `CacheEnabled` and `Shell.Timeout` are parsed into the struct
  for compatibility, and the loader emits a `slog.Warn` with
  the message "service.cache_enabled is deprecated and ignored"

#### Scenario: Default config omits the deprecated fields
- **WHEN** `DefaultConfig()` is called
- **THEN** the returned config does not set `CacheEnabled`,
  `ConcurrentOps`, `MaxConcurrent`, `Shell.Timeout`, or
  `Shell.Wrapper.BackupDir` (zero values)

### Requirement: Tool directive declared for govulncheck

The `go.mod` file SHALL declare `tool` directives for the
development tools that the project invokes via `go tool ...`,
specifically `golang.org/x/vuln/cmd/govulncheck`. Invoking
`go tool govulncheck ./...` SHALL reproduce the exact toolchain
version used by the project without network lookups.

#### Scenario: go tool govulncheck runs without network
- **WHEN** a developer runs `go tool govulncheck ./...` from
  the project root
- **THEN** the tool runs against the locally-resolved module
  graph without contacting any module proxy
