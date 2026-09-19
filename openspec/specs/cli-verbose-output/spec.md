# Capability: Verbose Output

## Purpose

`--verbose` / `-v` / `-vv` global flag for high-level (level 1) and
detailed (level 2) operation tracing on stderr. Verbose output SHALL
be generated only by the cmd layer, never by the service or domain
layers.

## Requirements

### Requirement: Two verbosity levels

The system SHALL map `-v` to level 1 (high-level operation flow) and
`-vv` to level 2 (parameters and intermediate steps).

#### Scenario: Level 1 output

- **WHEN** user runs `twiggit create myproject/feature -v`
- **THEN** system SHALL emit high-level messages like "Creating worktree for myproject/feature" on stderr

#### Scenario: Level 2 output

- **WHEN** user runs `twiggit create myproject/feature -vv`
- **THEN** system SHALL emit level 1 messages
- **AND** SHALL additionally emit parameter messages like
  "  from branch: main", "  to path: /home/.../myproject/feature"

### Requirement: `logv()` helper

The system SHALL provide a `logv(cmd, level, format, args...)` helper
in `cmd/util.go` that all commands use for verbose output. The helper
SHALL no-op when `cmd` has no `verbose` count or the level is too low.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Stderr only

The system SHALL write all verbose output to stderr. Stdout SHALL be
reserved for data output (listings, JSON, navigation paths) so
`twiggit list -v | jq` works in a pipeline.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Cmd layer only

The system SHALL emit verbose output only from the cmd layer
(`cmd/*.go`). The service and domain layers SHALL NOT call `logv()`,
`fmt.Fprintln(os.Stderr, ...)`, or any other verbose-output helper.

#### Scenario: Service layer silent

- **WHEN** user runs any command with `-vv`
- **THEN** verbose output SHALL originate only from `cmd/*.go`
- **AND** no service-layer calls SHALL write to stderr

### Requirement: No debug prefix

The system SHALL NOT prepend `DEBUG:` or `[VERBOSE]` to verbose output.
Messages SHALL be plain text with optional `  ` (two-space) indentation
for level 2 details.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
