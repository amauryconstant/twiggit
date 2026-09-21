# Spec Delta: Main Entry Point

## MODIFIED Requirements

### Requirement: Exit code propagation

The system SHALL exit with the exit code returned by the error formatter
(see `cli-error-formatting`). It SHALL NOT convert non-zero errors to a
different non-zero code; the binary SHALL exit with the value returned by
`GetExitCodeForError` unmodified. Under the 3-code dispatch
(`cli-error-formatting` Exit code mapping requirement), `*domain.ValidationError`
maps to `ExitCodeError` (1), cobra usage errors map to `ExitCodeUsage` (2),
and a clean run maps to `ExitCodeSuccess` (0). The "do not override to 1"
prohibition still applies: a non-zero code (`ExitCodeUsage` (2)) SHALL NOT
be converted to `ExitCodeError` (1) at the entry point, and vice versa.

#### Scenario: Specific exit code honored

- **WHEN** the error formatter returns `ExitCodeError` (1) for a
  `domain.ValidationError`
- **THEN** the binary SHALL exit with code 1, not 0

#### Scenario: Cobra usage error → exit 2

- **WHEN** the error formatter returns `ExitCodeUsage` (2) for a cobra
  usage error
- **THEN** the binary SHALL exit with code 2, not 1

#### Scenario: Success → exit 0

- **WHEN** the command returns no error
- **THEN** the binary SHALL exit with code `ExitCodeSuccess` (0)

### Requirement: Configuration loading

The system SHALL load configuration from defaults and
`$HOME/.config/twiggit/config.toml` (XDG), apply environment variable
overrides (`TWIGGIT_*`), and validate before command dispatch. Validation
failures and configuration errors both surface via the error formatter
and exit with `ExitCodeError` (1) under the 3-code dispatch (see
`cli-error-formatting` Exit code mapping requirement).

#### Scenario: Load and validate

- **WHEN** the binary starts
- **THEN** system SHALL load the configuration via `ConfigManager.Load()`
- **AND** SHALL call `Config.Validate()` before dispatch
- **AND** SHALL surface validation failures via the error formatter
  (exit code `ExitCodeError` (1) per the 3-code canonical mapping)
