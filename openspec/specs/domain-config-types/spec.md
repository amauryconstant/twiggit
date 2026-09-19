# Capability: Configuration Types

## Purpose

The full configuration knob surface (`domain.Config` and its sub-types)
loaded from `$HOME/.config/twiggit/config.toml`, plus defaults, env-var
overrides, and validation. This spec owns the type definitions; loading
and path expansion live in `infrastructure-config-manager`.

## Requirements

### Requirement: `Config` root struct

The system SHALL provide a `Config` root struct with the following shape:

```go
type Config struct {
    ProjectsDirectory   string                  // absolute path
    WorktreesDirectory  string                  // absolute path
    DefaultSourceBranch string                  // default "main"

    ContextDetection ContextDetectionConfig
    Git              GitConfig
    Services         ServiceConfig
    Validation       ValidationConfig
    Navigation       NavigationConfig
    Shell            ShellConfig
    Completion       CompletionConfig
}
```



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ContextDetectionConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `CacheTTL` | `cache_ttl` | `"5m"` | TTL for cached detection results |
| `GitOperationTimeout` | `git_operation_timeout` | `"30s"` | Per-call git op timeout during detection |
| `EnableGitValidation` | `enable_git_validation` | `true` | Run `git rev-parse` to confirm `.git` |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `GitConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `CLITimeout` | `cli_timeout` | `30` | Seconds before CLI git calls time out |
| `CacheEnabled` | `cache_enabled` | `true` | Enable GoGitClient LRU cache |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ServiceConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `CacheEnabled` | `cache_enabled` | `true` | Service-layer result caching |
| `CacheTTL` | `cache_ttl` | `5m` | Service cache TTL |
| `ConcurrentOps` | `concurrent_operations` | `true` | Allow concurrent per-project ops |
| `MaxConcurrent` | `max_concurrent` | `4` | Max concurrent ops (used when ConcurrentOps=true) |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ValidationConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `StrictBranchNames` | `strict_branch_names` | `true` | Enforce strict branch-name regex |
| `RequireCleanWorktree` | `require_clean_worktree` | `true` | Block destructive ops on dirty worktrees |
| `AllowForceDelete` | `allow_force_delete` | `false` | Allow `--force` on delete/prune |
| `ProtectedBranches` | `protected_branches` | `[main, master, develop, staging, production]` | Never delete these branches |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `NavigationConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `EnableSuggestions` | `enable_suggestions` | `true` | Enable completion suggestions |
| `MaxSuggestions` | `max_suggestions` | `10` | Cap on returned suggestions |
| `FuzzyMatching` | `fuzzy_matching` | `false` | Case-insensitive subsequence matching |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ShellConfig` and `ShellWrapperConfig`

The system SHALL provide `ShellConfig` and `ShellWrapperConfig` with
the following shapes:

```go
type ShellConfig struct {
    Wrapper     ShellWrapperConfig
    Enabled     bool  // default true
    Timeout     int   // seconds for shell operations; default 30
    HookTimeout int   // seconds for hook execution; default 30
}

type ShellWrapperConfig struct {
    Enabled      bool   // default true
    AutoDetect   bool   // default true
    DefaultShell string // default "bash"
    BackupEnabled bool  // default true
    BackupDir    string // default "~/.config/twiggit/backups"
}
```



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `CompletionConfig`

| Field | TOML key | Default | Purpose |
|---|---|---|---|
| `Timeout` | `timeout` | `"500ms"` | Per-completion-call timeout |
| `ExcludeBranches` | `exclude_branches` | `[]` | Glob patterns for branches to hide |
| `ExcludeProjects` | `exclude_projects` | `[]` | Glob patterns for projects to hide |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Defaults

`DefaultConfig()` SHALL return a `*Config` populated with the defaults
above, expanding `$HOME` to the user's home directory
(`~/Projects`, `~/Worktrees`).



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Validation

`Config.Validate()` SHALL return a `domain.ValidationError` (exit code
5) when:

- `ProjectsDirectory` is not absolute
- `WorktreesDirectory` is not absolute
- `DefaultSourceBranch` is empty

The error SHALL include the offending field name and the failing value.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Path expansion

The configuration loader SHALL expand `$VAR`, `${VAR}`, and `~` in
path fields: `ProjectsDirectory`, `WorktreesDirectory`,
`Shell.Wrapper.BackupDir`. Other fields SHALL NOT be expanded. See
`infrastructure-config-manager`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Environment overrides

Env vars of the form `TWIGGIT_<UPPER_SNAKE_FIELD>` SHALL override
configuration-file values at load time. The override SHALL be applied
after file loading but before validation.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
