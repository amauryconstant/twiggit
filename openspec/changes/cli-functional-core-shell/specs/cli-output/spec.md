# Spec Delta

## Purpose

Defines the `--output` flag formatter interface for `json`, `table`, and `plain` rendering so every list-style command shares one formatting seam and one registry.

## ADDED Requirements

### Requirement: --output accepts json, table, plain values

The system SHALL accept `--output=<value>` (short `-o <value>`) on every list-style command. The accepted values SHALL be exactly `json`, `table`, and `plain`.

#### Scenario: --output json emits structured JSON
- **WHEN** the user runs `twiggit list -o json`
- **THEN** the output is a single JSON document suitable for `jq`

#### Scenario: --output table emits aligned columns
- **WHEN** the user runs `twiggit list -o table`
- **THEN** the output is an aligned table with headers and a row per worktree

#### Scenario: --output plain emits raw human-readable rendering
- **WHEN** the user runs `twiggit list -o plain`
- **THEN** the output is unbordered human-readable rendering without table borders or JSON quoting

#### Scenario: Unknown --output value is rejected
- **WHEN** the user runs `twiggit list -o xml`
- **THEN** the system SHALL emit a `core.UsageError` and exit 2

### Requirement: Formatter writes to a writer (not IOStreams)

The `output.Formatter` interface SHALL accept an `io.Writer` for every method. Formatters SHALL NOT touch `iostreams.IOStreams` directly so they remain unit-testable without TTY detection.

#### Scenario: Formatter tests use *bytes.Buffer
- **WHEN** a test invokes `formatter.FormatJSON(&buf, &data)`
- **THEN** the test reads `buf.String()` to verify the output without configuring IOStreams

### Requirement: Default (no flag) falls through to human-readable rendering

When `--output` is not supplied, the system SHALL use `table` for list commands and `plain` for everything else.

#### Scenario: list without --output renders as table
- **WHEN** the user runs `twiggit list` without `--output`
- **THEN** the output SHALL be a table (same as `--output table`)

#### Scenario: create without --output renders as plain
- **WHEN** the user runs `twiggit create feat/foo` without `--output`
- **THEN** the output SHALL be a plain human-readable success message
