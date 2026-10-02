# Capability: Quiet Mode

## Purpose

Global `--quiet` flag (long form only; no `-q` short flag is bound
per `internal/cmdutil/global_flags.go:46`) that suppresses non-essential
output for cleaner scripting, while preserving errors and essential
data (navigation paths, JSON output).

## Requirements

### Requirement: Suppress success and hint messages

The system SHALL suppress success messages, hint lines, and progress
output when `--quiet` is set. Emitted errors SHALL be lowercase with no
trailing punctuation. The quiet-mode output SHALL route data through `cmd.OutOrStdout()` and errors through `cmd.ErrOrStderr()` (never direct `os.Stdout`/`os.Stderr`).

#### Scenario: Quiet list output

- **WHEN** user runs `twiggit list --quiet`
- **THEN** system SHALL emit only the listing data on stdout
- **AND** SHALL NOT emit success or hint messages on stderr
- **AND** errors SHALL still be printed to stderr

### Requirement: Preserve essential output

The system SHALL preserve essential output: navigation paths printed by
`-C` mode and JSON data on stdout under `--quiet`.

#### Scenario: Quiet create with `-C`

- **WHEN** user runs `twiggit create -C myproject/feature --quiet`
- **AND** creation succeeds
- **THEN** system SHALL print the worktree path to stdout
- **AND** SHALL NOT print success message to stderr

### Requirement: Verbose overrides quiet

The system SHALL allow `--verbose` to bypass quiet-mode suppression
when both flags are set. See `cli-command-options-pattern`.

#### Scenario: Verbose wins

- **WHEN** user runs `twiggit list --quiet --verbose`
- **THEN** verbose output SHALL appear on stderr
- **AND** quiet suppression SHALL be bypassed

### Requirement: Progress reporter respects quiet

The system SHALL suppress progress reporter output when `--quiet` is
set; the reporter SHALL be constructed with the quiet flag.

#### Scenario: Bulk prune quiet

- **WHEN** user runs `twiggit prune --all --quiet`
- **THEN** progress messages SHALL NOT appear on stderr
- **AND** summary messages SHALL also be suppressed
