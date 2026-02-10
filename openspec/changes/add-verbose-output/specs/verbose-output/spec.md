## ADDED Requirements

### Requirement: Verbose flag controls output verbosity
The system SHALL provide a persistent `--verbose` flag that accepts multiple occurrences to control output verbosity levels across all CLI commands.

#### Scenario: No verbose flag specified
- **WHEN** user runs any command without `--verbose` or `-v` flag
- **THEN** only normal command output is displayed
- **AND** no verbose messages are shown

#### Scenario: Single verbose flag (-v)
- **WHEN** user runs any command with `-v` or `--verbose`
- **THEN** normal command output is displayed
- **AND** level 1 verbose messages (high-level operation flow) are displayed
- **AND** level 2 verbose messages (detailed parameters) are NOT displayed

#### Scenario: Double verbose flag (-vv)
- **WHEN** user runs any command with `-vv` or `--verbose --verbose`
- **THEN** normal command output is displayed
- **AND** level 1 verbose messages (high-level operation flow) are displayed
- **AND** level 2 verbose messages (detailed parameters) are displayed

### Requirement: Level 1 verbose output describes high-level operations
The system SHALL provide level 1 verbose messages that describe the high-level flow of operations in user-friendly language.

#### Scenario: Create command level 1 output
- **WHEN** user runs `twiggit create test-branch -v`
- **THEN** output includes "Creating worktree for <project>/<branch>" message
- **AND** message is displayed before the operation completes

#### Scenario: Delete command level 1 output
- **WHEN** user runs `twiggit delete test-branch -v`
- **THEN** output includes "Deleting worktree at <path>" message
- **AND** message is displayed before the operation completes

#### Scenario: List command level 1 output
- **WHEN** user runs `twiggit list -v`
- **THEN** output includes "Listing worktrees for <project>" message
- **AND** message is displayed before the operation completes

#### Scenario: Cd command level 1 output
- **WHEN** user runs `twiggit cd <identifier> -v`
- **THEN** output includes "Navigating to worktree" message
- **AND** message is displayed before the operation completes

#### Scenario: Setup-shell command level 1 output
- **WHEN** user runs `twiggit setup-shell -v`
- **THEN** output includes "Setting up shell wrapper" message
- **AND** message is displayed before the operation completes

### Requirement: Level 2 verbose output shows detailed parameters
The system SHALL provide level 2 verbose messages that display detailed parameters and intermediate steps with two-space indentation.

#### Scenario: Create command level 2 output
- **WHEN** user runs `twiggit create test-branch -vv`
- **THEN** output includes indented level 2 messages showing:
  - "  from branch: <source-branch>"
  - "  to path: <worktree-path>"
  - "  in repo: <repo-path>"
  - "  creating parent directory: <parent-dir>"

#### Scenario: Delete command level 2 output
- **WHEN** user runs `twiggit delete test-branch -vv`
- **THEN** output includes indented level 2 messages showing:
  - "  project: <project-name>"
  - "  branch: <branch-name>"
  - "  force: <true|false>"

#### Scenario: List command level 2 output
- **WHEN** user runs `twiggit list -vv`
- **THEN** output includes indented level 2 messages showing:
  - "  project: <project-name>"
  - "  repository: <repo-path>"
  - "  including main worktree: <true|false>"

#### Scenario: Cd command level 2 output
- **WHEN** user runs `twiggit cd <identifier> -vv`
- **THEN** output includes indented level 2 messages showing:
  - "  worktree path: <worktree-path>"
  - "  resolved project: <project-name>"

#### Scenario: Setup-shell command level 2 output
- **WHEN** user runs `twiggit setup-shell -vv`
- **THEN** output includes indented level 2 messages showing:
  - "  shell type: <bash|zsh|fish>"
  - "  config file path: <config-file-path>"

### Requirement: Verbose output uses plain text format
The system SHALL format all verbose output as plain text with no color, no prefix, and no formatting characters.

#### Scenario: Verbose messages have no color
- **WHEN** user runs any command with verbose flag
- **THEN** verbose messages are displayed in plain text
- **AND** no ANSI color codes are included

#### Scenario: Verbose messages have no prefix
- **WHEN** user runs any command with verbose flag
- **THEN** verbose messages do NOT include "DEBUG:" prefix
- **AND** verbose messages do NOT include "[VERBOSE]" prefix
- **AND** verbose messages start directly with the message text

#### Scenario: Verbose output is pipe-friendly
- **WHEN** user runs any command with verbose flag and pipes to another command
- **THEN** normal output goes to stdout
- **AND** verbose output goes to stderr
- **AND** downstream command receives only normal output on stdin

### Requirement: Verbose output only appears in command layer
The system SHALL only generate verbose output from the command layer (cmd/*.go), never from the service layer (internal/services/*.go).

#### Scenario: Service layer produces no verbose output
- **WHEN** service layer code executes any operation
- **THEN** no verbose output is generated by service layer
- **AND** service layer does not call logv() or similar functions
- **AND** service layer does not write to stdout or stderr

#### Scenario: Command layer generates all verbose output
- **WHEN** command layer calls service methods
- **THEN** command layer generates appropriate verbose messages before/after service calls
- **AND** verbose output timing reflects the operation flow

### Requirement: Log helper function provides consistent interface
The system SHALL provide a `logv()` helper function in cmd/util.go that handles all verbose output logic.

#### Scenario: logv checks verbose level
- **WHEN** logv() is called with level parameter
- **THEN** function checks the verbose flag count from Cobra command
- **AND** function outputs message only if verbose count >= level parameter

#### Scenario: logv applies indentation
- **WHEN** logv() is called with level parameter > 1
- **THEN** function prepends two spaces ("  ") to the message
- **AND** function does NOT indent messages for level 1

#### Scenario: logv writes to stderr
- **WHEN** logv() is called with any message
- **THEN** function writes to command's ErrOrStderr()
- **AND** function writes with newline termination
