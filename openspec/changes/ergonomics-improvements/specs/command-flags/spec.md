## ADDED Requirements

### Requirement: Auto-Confirmation Flag

Commands with interactive confirmation prompts SHALL support a `--yes/-y` flag to auto-confirm prompts while preserving safety checks.

#### Scenario: Prune with --yes flag

- **WHEN** user executes `twiggit prune --all --yes`
- **THEN** the system SHALL skip the bulk prune confirmation prompt
- **AND** the system SHALL still enforce safety checks (uncommitted changes, protected branches)

#### Scenario: Prune with --yes short flag

- **WHEN** user executes `twiggit prune -a -y`
- **THEN** the system SHALL behave identically to `twiggit prune --all --yes`

#### Scenario: --yes without --force preserves safety

- **WHEN** user executes `twiggit prune --yes` on a worktree with uncommitted changes
- **THEN** the system SHALL refuse to delete the worktree
- **AND** the error message SHALL indicate uncommitted changes

#### Scenario: --yes combined with --force

- **WHEN** user executes `twiggit prune --yes --force`
- **THEN** the system SHALL skip confirmation prompts
- **AND** the system SHALL bypass uncommitted changes safety check

### Requirement: Short Flag for List All

The `list` command's `--all` flag SHALL have a short form `-a` for common Unix convention.

#### Scenario: List all with short flag

- **WHEN** user executes `twiggit list -a`
- **THEN** the system SHALL behave identically to `twiggit list --all`

#### Scenario: List all flag documentation

- **WHEN** user executes `twiggit list --help`
- **THEN** the help text SHALL show `-a, --all` for the flag

### Requirement: Help Text Without Duplication

Command Long descriptions SHALL NOT duplicate flag information that Cobra displays in the Flags section.

#### Scenario: Create command help

- **WHEN** user executes `twiggit create --help`
- **THEN** the Long description SHALL NOT include a "Flags:" section listing flags
- **AND** the Long description SHALL include an "Examples:" section

#### Scenario: Init command help

- **WHEN** user executes `twiggit init --help`
- **THEN** the Long description SHALL NOT include a "Flags:" section listing flags

#### Scenario: List command help

- **WHEN** user executes `twiggit list --help`
- **THEN** the Long description SHALL include an "Examples:" section with practical usage

#### Scenario: Delete command help

- **WHEN** user executes `twiggit delete --help`
- **THEN** the Long description SHALL include an "Examples:" section with practical usage
