## ADDED Requirements

### Requirement: Command Aliases

The system SHALL provide Unix-style aliases for frequently used commands to reduce typing friction.

#### Scenario: List command alias

- **WHEN** user executes `twiggit ls`
- **THEN** the system SHALL behave identically to `twiggit list`

#### Scenario: List alias in help text

- **WHEN** user executes `twiggit list --help`
- **THEN** the help text SHALL display "Aliases: ls"

#### Scenario: Delete command alias

- **WHEN** user executes `twiggit rm <target>`
- **THEN** the system SHALL behave identically to `twiggit delete <target>`

#### Scenario: Delete alias in help text

- **WHEN** user executes `twiggit delete --help`
- **THEN** the help text SHALL display "Aliases: rm"

#### Scenario: Tab completion for aliases

- **WHEN** user invokes shell completion
- **THEN** aliases `ls` and `rm` SHALL appear as completion options
