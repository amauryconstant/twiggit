## Command-Flags Spec

This specification establishes conventions and requirements for command-line flags in the twiggit CLI, including naming, documentation, consistency, and shell wrapper integration.

### Requirement: Flag Naming Conventions

The system SHALL use consistent flag naming patterns across all commands.

#### Scenario: Short form flag usage

- **WHEN** a command has a commonly used flag (e.g., --force)
- **THEN** the flag SHALL have a short form (e.g., -f)

- **WHEN** a command has a flag that enables machine-readable output for shell wrapper integration
- **THEN** the flag SHALL use `-C` as the short form

- **WHEN** multiple commands share the same flag functionality (e.g., --force)
- **THEN** the short form SHALL be consistent across all commands

#### Scenario: Long form flag naming

- **WHEN** a flag name is a commonly used Unix convention
- **THEN** the long form SHALL match common Unix naming (e.g., --cd for "change directory")

- **WHEN** similar functionality exists across multiple commands (e.g., changing directories)
- **THEN** the flag name SHALL be consistent across all commands

- **WHEN** a flag is specific to one command
- **THEN** the flag name SHALL clearly indicate its purpose and context

#### Scenario: Flag type consistency

- **WHEN** a flag accepts a boolean value
- **THEN** the flag registration SHALL use BoolVar or BoolVarP

- **WHEN** a flag accepts a string or number value
- **THEN** the flag registration SHALL use StringVar, IntVar, etc.

### Requirement: Flag Documentation Requirements

All command flags SHALL be documented in command help text and AGENTS.md.

#### Scenario: Long description documentation

- **WHEN** a command defines flags
- **THEN** the Long description SHALL list all implemented flags

- **WHEN** a flag has a default value
- **THEN** the Long description SHALL document the default

- **WHEN** a flag modifies command behavior in a significant way
- **THEN** the Long description SHALL explain the behavior change

#### Scenario: AGENTS.md documentation

- **WHEN** AGENTS.md documents a command
- **THEN** the Flags section SHALL list all implemented flags

- **WHEN** a flag has a short form
- **THEN** AGENTS.md SHALL document it as `-X, --long-form`

- **WHEN** AGENTS.md documents flag behavior
- **THEN** the documented behavior SHALL match actual implementation

### Requirement: Shell Wrapper Integration

Commands that support shell wrapper integration SHALL use consistent output patterns.

#### Scenario: Wrapper integration flags

- **WHEN** a command outputs a path for shell wrapper integration
- **THEN** the command SHALL support a `-C, --cd` flag

- **WHEN** the `-C` flag is set
- **THEN** the command SHALL output only the absolute path to stdout

- **WHEN** the `-C` flag is NOT set
- **THEN** the command SHALL output human-friendly success messages

#### Scenario: Context-aware navigation

- **WHEN** a command with `-C` flag is executed from a worktree context and needs to navigate away
- **THEN** the command SHALL output the project root path

- **WHEN** a command with `-C` flag is executed from a project context
- **THEN** the command SHALL output nothing (no navigation needed)

- **WHEN** a command with `-C` flag is executed from outside git context
- **THEN** the command SHALL output nothing (no sensible navigation target)

### Requirement: Output Format Conventions

Commands SHALL follow consistent output format patterns.

#### Scenario: Default output format

- **WHEN** a command completes successfully without wrapper integration flags
- **THEN** the command SHALL output a descriptive success message

- **WHEN** a command creates a resource
- **THEN** the output SHALL include the resource identifier and location

- **WHEN** a command deletes a resource
- **THEN** the output SHALL confirm the deletion and resource path

#### Scenario: Machine-readable output format

- **WHEN** a wrapper integration flag (e.g., `-C`) is set
- **THEN** the command SHALL output only the path to stdout

- **WHEN** outputting a path for shell wrapper
- **THEN** the output SHALL be an absolute path

- **WHEN** outputting a path
- **THEN** the path SHALL be followed by a newline character

### Requirement: Backward Compatibility

Changes to flags SHALL maintain backward compatibility where possible.

#### Scenario: Existing flag behavior preservation

- **WHEN** a command is executed without new flags
- **THEN** the behavior SHALL remain unchanged from previous versions

- **WHEN** a user script uses the old form of a renamed flag
- **THEN** the short form (if unchanged) SHALL continue to work

- **WHEN** breaking changes to flags are necessary
- **THEN** the change SHALL be clearly documented as **BREAKING**
