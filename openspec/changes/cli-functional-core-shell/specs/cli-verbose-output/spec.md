# Spec Delta

## MODIFIED Requirements

### Requirement: Two verbosity levels

The system SHALL expose `--verbose` (short `-v`) as a global persistent flag on the root command. One `-v` SHALL set the boolean `Verbose` field on `IOStreams` to `true`. The previous two-level scheme (`-v` for high-level flow, `-vv` for parameters) is removed; `-vv` is treated as `-v`.

#### Scenario: Level 1 output
- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** `ios.Verbosef` SHALL emit one line "cloning <src>" on `ios.ErrOut` with a trailing newline

#### Scenario: Level 2 output
- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("from %s", src)` followed by `ios.Verbosef("  to %s", dst)`
- **THEN** both lines SHALL be emitted on `ios.ErrOut` (no level distinction: the previous `-vv` level-2 behavior is collapsed into the single boolean Verbose)

### Requirement: `logv()` helper

The system SHALL provide an `iostreams.IOStreams.Verbosef(format string, args ...any)` helper that all commands use for verbose output. The helper SHALL write exactly one line terminated by `\n` to `ios.ErrOut` when `ios.Verbose == true` and SHALL be a no-op otherwise. The line is rendered through `ios.Styles().Dim(...)`. The previous `logv(cmd, level, format, args...)` helper is removed; commands call `ios.Verbosef` directly. See `cli-iostreams` for the full IOStreams surface.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbosef emits one line on -v
- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** exactly one line "cloning <src>" lands on `ios.ErrOut` with a trailing newline

#### Scenario: Verbosef is silent without -v
- **WHEN** `ios.Verbose == false` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** nothing is written to `ios.ErrOut`

### Requirement: Stderr only

The system SHALL write all verbose output to stderr. Stdout SHALL be reserved for data output (listings, JSON, navigation paths) so `twiggit list -v | jq` works in a pipeline.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbose output does not corrupt data output
- **WHEN** user runs `twiggit list -v -o json`
- **THEN** the JSON document on stdout SHALL contain only data
- **AND** the verbose logs on stderr SHALL contain the diagnostic text

### Requirement: Cmd layer only

The system SHALL emit verbose output only from the cmd layer (`cmd/*.go`). The service, core, and adapter packages SHALL NOT call `ios.Verbosef`, `fmt.Fprintln(os.Stderr, ...)`, or any other stderr-writing helper. A depguard rule SHALL forbid direct `os.Stdout`/`os.Stderr` access in `cmd/` outside `main.go`.

#### Scenario: Service layer silent
- **WHEN** user runs any command with `-v`
- **THEN** verbose output SHALL originate only from `cmd/*.go`
- **AND** no service-layer calls SHALL write to stderr

### Requirement: No debug prefix

The system SHALL NOT prepend `DEBUG:` or `[VERBOSE]` to verbose output. Messages SHALL be plain text with the dim style applied via lipgloss. The previous two-space indentation rule for level-2 details is removed (no level distinction anymore).

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbose output omits the DEBUG: prefix
- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** the rendered line SHALL contain "cloning <src>" with no `DEBUG:` or `[VERBOSE]` prefix
