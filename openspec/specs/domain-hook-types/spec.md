# Capability: Hook Types

## Purpose

Domain types for the post-create hook subsystem. The runner that
executes them lives in `infrastructure-hook-runner`; the cmd display
contract lives in `cli-worktree-hooks`.

## Requirements

### Requirement: `HookType` and constants

The system SHALL provide a `HookType` string type and a
`HookPostCreate` constant:

```go
type HookType string
const HookPostCreate HookType = "post-create"
```

`HookPostCreate` is the only currently supported hook type.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `HookConfig` and `HookDefinition`

The system SHALL provide `HookConfig` and `HookDefinition` with the
following shapes:

```go
type HookConfig struct {
    PostCreate *HookDefinition `toml:"post-create" koanf:"post-create"`
}

type HookDefinition struct {
    Commands []string `toml:"commands" koanf:"commands"`
}
```



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `HookResult`

The system SHALL provide a `HookResult` struct with the following
shape:

```go
type HookResult struct {
    HookType HookType
    Executed bool          // Were any commands configured?
    Success  bool          // Did all commands exit 0?
    Failures []HookFailure // empty when Success
}
```



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `HookFailure`

```go
type HookFailure struct {
    Command  string
    ExitCode int
    Output   string
}
```

`Output` SHALL capture combined stdout+stderr for the failing command.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `HookRunRequest`

The system SHALL provide an application-layer `HookRunRequest` with
the following shape:

```go
type HookRunRequest struct {
    HookType       HookType
    WorktreePath   string
    ProjectName    string
    BranchName     string
    SourceBranch   string
    MainRepoPath   string
    ConfigFilePath string
}
```

This struct lives in the application layer
(`application.HookRunRequest`); the domain type surface is the four
structs above. The list of env vars the runner injects into the
process is owned by `infrastructure-hook-runner`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
