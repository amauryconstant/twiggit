## Purpose

The quiet-mode capability suppresses non-essential output for clean scripting, reducing noise in logs and automation pipelines while preserving critical information like errors and essential data output.

---

## ADDED Requirements

### Requirement: Quiet flag suppresses non-essential output
The system SHALL provide a global `--quiet/-q` flag that suppresses success messages and hints while preserving errors and essential data.

#### Scenario: Quiet mode suppresses success messages
- **WHEN** user runs any command with `--quiet`
- **THEN** success messages are not displayed
- **AND** command exit code indicates success (0)

#### Scenario: Quiet mode suppresses hint messages
- **WHEN** user runs any command with `--quiet`
- **THEN** hint messages (usage tips, suggestions) are not displayed

#### Scenario: Quiet mode preserves errors
- **WHEN** user runs any command with `--quiet` and an error occurs
- **THEN** error message is displayed to stderr
- **AND** command exit code indicates failure (non-zero)

#### Scenario: Quiet mode preserves essential output
- **WHEN** user runs `cd <branch> -C --quiet`
- **THEN** worktree path is output to stdout
- **AND** no success messages are output

### Requirement: Quiet and verbose are mutually exclusive
The system SHALL handle the conflict between `--quiet` and `--verbose` flags with verbose taking priority.

#### Scenario: Both quiet and verbose specified
- **WHEN** user runs any command with both `--quiet` and `--verbose`
- **THEN** verbose output is displayed (verbose wins)
- **AND** quiet flag is effectively ignored

#### Scenario: Quiet with single verbose
- **WHEN** user runs any command with `--quiet -v`
- **THEN** level 1 verbose output is displayed
- **AND** success messages are still suppressed (per verbose behavior)

### Requirement: Quiet mode is a global flag
The system SHALL define `--quiet/-q` as a global persistent flag on the root command.

#### Scenario: Quiet flag available on all commands
- **WHEN** user runs any twiggit command
- **THEN** `--quiet/-q` flag is available

#### Scenario: Quiet flag position
- **WHEN** user runs `twiggit --quiet list`
- **THEN** quiet mode is enabled for the list command
- **AND** flag can appear before or after subcommand

### Requirement: Quiet mode suppresses progress output
The system SHALL suppress progress reporting when quiet mode is enabled.

#### Scenario: Quiet mode with bulk prune
- **WHEN** user runs `prune --all --quiet`
- **THEN** progress messages are not displayed
- **AND** only final summary (if any) is suppressed
- **AND** errors are still displayed if they occur
