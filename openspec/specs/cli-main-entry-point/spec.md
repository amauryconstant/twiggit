# cli-main-entry-point Specification

## Purpose

Binary entry point that loads configuration from defaults, XDG, and
environment overrides, validates it before dispatch, invokes the cobra
command tree, and propagates the exit code returned by the cmd-side
error formatter (`cli-error-formatting`) without remapping non-zero
values.

## Requirements

### Requirement: Exit code propagation

The system SHALL exit with the exit code returned by `cmdutil.ExitCodeFor(err)` (see `cli-error-formatting`). It SHALL NOT convert non-zero errors to a different non-zero code; the binary SHALL exit with the value returned by `ExitCodeFor` unmodified. Under the 3-code dispatch, `*core.ValidationError` and `*core.NotFoundError` map to `ExitError` (1), `*core.UsageError` maps to `ExitUsage` (2), and a clean run maps to `ExitOK` (0). Signal cancellation via `signal.NotifyContext` (SIGINT/SIGTERM) bypasses `ExitCodeFor` and exits 130/143 respectively.

#### Scenario: Specific exit code honored

- **WHEN** the error formatter returns `ExitError` (1) for a `core.ValidationError`
- **THEN** the binary SHALL exit with code 1, not 0

#### Scenario: Cobra usage error → exit 2

- **WHEN** the error formatter returns `ExitUsage` (2) for a cobra usage error
- **THEN** the binary SHALL exit with code 2, not 1

#### Scenario: Success → exit 0

- **WHEN** the command returns no error
- **THEN** the binary SHALL exit with code `ExitOK` (0)

#### Scenario: SIGINT exits 130

- **WHEN** the user sends SIGINT to the binary during a long-running command
- **THEN** `main` detects cancellation via `ctx.Err() != nil` and calls `os.Exit(130)`

### Requirement: Configuration loading

The system SHALL load configuration from `core.DefaultConfig()` defaults, the resolved XDG config file (`$XDG_CONFIG_HOME/twiggit/config.toml` or `$HOME/.config/twiggit/config.toml`), and `TWIGGIT_*` environment variables, via the `cmdutil.Factory.Config func() (*core.Config, error)` lazy function field. CLI flags override configuration values at the cmd layer via cobra; `main.go` SHALL NOT bind flags or read configuration directly. See `infrastructure-config-manager` for the loading order and `domain-config-types` for the type surface.

#### Scenario: Load and validate

- **WHEN** the binary starts
- **THEN** the Factory's lazy `Config()` field SHALL be assigned a `func() (*core.Config, error)` loader
- **AND** the loader SHALL NOT execute until a command calls `f.Config()`
- **AND** configuration validation failures SHALL surface via `output.FormatError` and exit with `ExitError` (1)

### Requirement: Thin composition root (main.go ≤ 50 lines)

`main.go` SHALL act as a thin composition root: it wires the cobra root command, the `cmdutil.Factory`, the `iostreams.IOStreams`, the signal-aware `context.Context`, the debug `*slog.Logger` (when `TWIGGIT_DEBUG=1`), and delegates execution to the cmd tree. The file SHALL NOT bind flags, parse configuration eagerly, or contain business logic. The target size is ≤ 50 lines of Go (excluding imports); exceeding this limit is a smell pointing at logic that belongs in `internal/cmdutil`, `internal/iostreams`, or a subcommand file. No additional helper package SHALL be introduced solely to shrink `main.go` — the file's responsibility is composition, not delegation.

#### Scenario: main.go size budget

- **WHEN** `main.go` is reviewed
- **THEN** it SHALL contain ≤ 50 lines of Go (excluding the import block)
- **AND** it SHALL NOT bind cobra flags or read configuration values

#### Scenario: Composition root wires dependencies only

- **WHEN** the binary starts
- **THEN** `main.go` constructs the `cmdutil.Factory`, the `IOStreams`, the `*slog.Logger` (gated on `TWIGGIT_DEBUG`), and the `context.Context` (from `signal.NotifyContext`)
- **AND** every other concern (config loading, output formatting, error dispatch) SHALL live outside `main.go`
