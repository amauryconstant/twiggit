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

### Requirement: Flag Registration Pattern

Commands SHALL use appropriate flag registration patterns based on flag scope and access pattern.

#### Scenario: Command-specific flags

- **WHEN** a flag is specific to a single command
- **THEN** flag SHALL be registered using `*Var` or `*VarP` pattern (e.g., `BoolVar`, `BoolVarP`, `StringVar`)
- **THEN** flag SHALL be bound to a local variable at registration time
- **THEN** flag value SHALL be accessed directly via the bound variable in RunE handler
- **THEN** flag value SHALL be passed as a function parameter if needed in helper functions

- **WHEN** a flag has a short form
- **THEN** flag SHALL use `*VarP` variant (e.g., `BoolVarP(&force, "force", "f", ...)`)
- **WHEN** a flag has no short form
- **THEN** flag SHALL use `*Var` variant (e.g., `BoolVar(&check, "check", false, ...)`)
- **THEN** flag SHALL NOT use `*VarP` with empty string for short form

#### Scenario: Global persistent flags

- **WHEN** a flag is a global persistent flag (defined in root command)
- **THEN** flag SHALL be registered using `PersistentFlags()`
- **WHEN** flag value is accessed from utility functions without direct access to command context
- **THEN** flag value SHALL be retrieved using `Get*()` methods (e.g., `GetCount()`, `GetBool()`, `GetString()`)
- **THEN** flag value retrieval SHALL use string-based flag name lookup

#### Scenario: Type safety and compile-time checking

- **WHEN** using `*Var` or `*VarP` pattern for command-specific flags
- **THEN** flag type SHALL be enforced at compile time through variable type
- **WHEN** using `Get*()` methods for global flags
- **THEN** flag type SHALL be validated at runtime through Cobra's flag API

#### Scenario: Code consistency across commands

- **WHEN** multiple commands implement similar flag functionality
- **THEN** all commands SHALL use consistent registration pattern (`*Var`/`*VarP`)
- **WHEN** command-specific flags are used within command execution
- **THEN** pattern SHALL NOT mix `*Var` registration with `Get*()` retrieval in the same command

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
