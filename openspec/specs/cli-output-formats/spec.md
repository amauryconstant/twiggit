# Capability: Output Formats

## Purpose

Text and JSON output for scripting. The `--output` flag selects the
format; text is default, JSON is structured. Data goes to stdout;
errors and progress go to stderr.

## Requirements

### Requirement: Text output (default)

The system SHALL emit tabular text by default for commands that produce
listings (`list`, etc.).

#### Scenario: Default text output

- **WHEN** user runs `twiggit list` without `--output`
- **THEN** system SHALL emit a human-readable table on stdout
- **AND** errors and progress SHALL go to stderr

### Requirement: JSON output via `--output`

The system SHALL accept `--output=json` / `-o json` to emit structured
JSON suitable for parsing by `jq` and other tools.

#### Scenario: JSON list output

- **WHEN** user runs `twiggit list -o json`
- **THEN** system SHALL emit JSON with shape:
  `{"worktrees": [{"branch": "...", "path": "...", "status": "clean|modified|detached"}]}`
- **AND** SHALL exit with status 0

#### Scenario: Unknown format rejected

- **WHEN** user passes `--output=xml` (unsupported)
- **THEN** system SHALL return error naming supported formats
  (`text`, `json`)
- **AND** SHALL exit with code 2 (usage error)

### Requirement: Stream separation

The system SHALL send all data output to stdout and all diagnostic
output (errors, progress, verbose messages) to stderr so that
`twiggit list -o json | jq` works in a pipeline.

#### Scenario: Pipeable JSON

- **WHEN** user runs `twiggit list -o json | jq .`
- **THEN** `jq` SHALL receive only the JSON document on stdin
- **AND** progress messages SHALL NOT corrupt the JSON stream
