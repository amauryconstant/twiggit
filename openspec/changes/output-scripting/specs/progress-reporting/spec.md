## Purpose

The progress-reporting capability provides feedback during long-running bulk operations, showing users what is being processed without requiring external progress bar libraries.

---

## Requirements

### Requirement: Progress output during bulk operations
The system SHALL display progress messages during bulk operations like `prune --all` to indicate what is being processed.

#### Scenario: Progress during bulk prune
- **WHEN** user runs `prune --all` with multiple worktrees to process
- **THEN** progress messages are displayed for each worktree being processed
- **AND** messages indicate which worktree is being evaluated or deleted

#### Scenario: Progress shows current item
- **WHEN** progress is reported during bulk operation
- **THEN** message includes current item being processed
- **AND** message format is human-readable

### Requirement: Progress output goes to stderr
The system SHALL write all progress messages to stderr to separate them from command output on stdout.

#### Scenario: Progress does not mix with stdout
- **WHEN** user runs `prune --all` and pipes stdout to a file
- **THEN** progress messages go to stderr
- **AND** stdout contains only command result data

#### Scenario: Progress visible in terminal
- **WHEN** user runs `prune --all` in an interactive terminal
- **THEN** progress messages are visible
- **AND** messages do not interfere with final output

### Requirement: ProgressReporter provides simple interface
The system SHALL provide a ProgressReporter struct for consistent progress output across commands.

#### Scenario: ProgressReporter construction
- **WHEN** creating a ProgressReporter
- **THEN** it accepts quiet mode flag and output writer
- **AND** respects quiet mode to suppress output

#### Scenario: ProgressReporter report method
- **WHEN** Report() method is called with format and args
- **THEN** formatted message is written to output writer
- **AND** message is terminated with newline

#### Scenario: ProgressReporter with quiet mode
- **WHEN** ProgressReporter is created with quiet=true
- **THEN** all Report() calls produce no output

### Requirement: Progress suppressed in quiet mode
The system SHALL suppress progress output when `--quiet` flag is set.

#### Scenario: Quiet mode suppresses progress
- **WHEN** user runs `prune --all --quiet`
- **THEN** no progress messages are displayed
- **AND** operation completes silently (except errors)

### Requirement: No external dependencies for progress
The system SHALL implement progress reporting using only Go standard library.

#### Scenario: Progress uses stdlib only
- **WHEN** progress reporting is implemented
- **THEN** no external progress bar libraries are imported
- **AND** implementation uses fmt and io packages only
