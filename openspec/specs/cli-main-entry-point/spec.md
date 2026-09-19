# Capability: Main Entry Point

## Purpose

The `main.go` entry point that constructs the `ServiceContainer`,
loads configuration, dispatches the cobra command tree, recovers from
panics, and exits with the code returned by the formatter.

## Requirements

### Requirement: Bootstrap service container

The system SHALL construct the `ServiceContainer` (WorktreeService,
ProjectService, NavigationService, ContextService, ShellService) at
startup and pass it through cobra commands.

#### Scenario: Container construction

- **WHEN** the binary starts
- **THEN** system SHALL construct all five services via constructor
  injection (no globals, no `init()`)
- **AND** SHALL attach them to the root cobra command for child access

### Requirement: Configuration loading

The system SHALL load configuration from defaults and
`$HOME/.config/twiggit/config.toml` (XDG), apply environment variable
overrides (`TWIGGIT_*`), and validate before command dispatch.

#### Scenario: Load and validate

- **WHEN** the binary starts
- **THEN** system SHALL load the configuration via `ConfigManager.Load()`
- **AND** SHALL call `Config.Validate()` before dispatch
- **AND** SHALL surface validation failures via the error formatter
  (exit code 3 for config errors, 5 for validation errors)

### Requirement: Panic recovery

The system SHALL recover from any panic in the command tree, write
"Internal error: <panic value>" to stderr, and exit with code 1. With
`TWIGGIT_DEBUG=1`, the system SHALL additionally print the full stack
trace. See `cli-error-formatting`.

#### Scenario: Recover from panic

- **WHEN** a panic occurs during command execution
- **THEN** system SHALL write "Internal error: <panic value>" to stderr
- **AND** SHALL exit with code 1

#### Scenario: Debug stack

- **WHEN** `TWIGGIT_DEBUG=1` is set and a panic occurs
- **THEN** system SHALL print the stack trace in addition to the panic value

### Requirement: Exit code propagation

The system SHALL exit with the exit code returned by the error formatter
(see `cli-error-formatting`). It SHALL NOT convert non-zero errors into
exit code 1 if a more specific mapping exists.

#### Scenario: Specific exit code honored

- **WHEN** the error formatter returns `ExitCodeValidation` (5)
- **THEN** the binary SHALL exit with code 5, not 1
