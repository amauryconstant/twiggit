# Capability: Verbose Output

## Purpose

`--verbose` / `-v` / `-vv` global flag for high-level (level 1) and
detailed (level 2) operation tracing on stderr. Verbose output SHALL
be generated only by the cmd layer, never by the service or domain
layers.

## Requirements

### Requirement: Two verbosity levels

The system SHALL expose `--verbose` (short `-v`) as a global persistent flag on the root command. One `-v` SHALL set the boolean `Verbose` field on `IOStreams` to `true`. The previous two-level scheme (`-v` for high-level flow, `-vv` for parameters) is removed; `-vv` is treated as `-v`.

#### Scenario: Level 1 output

- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** `ios.Verbosef` SHALL emit one line "cloning <src>" on `ios.ErrOut` with a trailing newline

#### Scenario: Level 2 output

- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("from %s", src)` followed by `ios.Verbosef("  to %s", dst)`
- **THEN** both lines SHALL be emitted on `ios.ErrOut` (no level distinction: the previous `-vv` level-2 behavior is collapsed into the single boolean Verbose)

### Requirement: `logv()` helper

The system SHALL provide an `iostreams.IOStreams.Verbosef(format string, args ...any)` helper that all commands use for verbose output. The helper SHALL write exactly one line terminated by `\n` to `ios.ErrOut` when `ios.Verbose == true` and SHALL be a no-op otherwise. The line is rendered through `ios.Styles().Dim(...)`. The previous `logv(cmd, level, format, args...)` helper is removed; commands call `ios.Verbosef` directly. See `cli-iostreams` for the full IOStreams surface.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbosef emits one line on -v

- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** exactly one line "cloning <src>" lands on `ios.ErrOut` with a trailing newline

#### Scenario: Verbosef is silent without -v

- **WHEN** `ios.Verbose == false` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** nothing is written to `ios.ErrOut`

### Requirement: Stderr only

The system SHALL write all verbose output to stderr. Stdout SHALL be reserved for data output (listings, JSON, navigation paths) so `twiggit list -v | jq` works in a pipeline.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbose output does not corrupt data output

- **WHEN** user runs `twiggit list -v -o json`
- **THEN** the JSON document on stdout SHALL contain only data
- **AND** the verbose logs on stderr SHALL contain the diagnostic text

### Requirement: Cmd layer only

The system SHALL emit verbose output only from the cmd layer (`cmd/*.go`). The service, core, and adapter packages SHALL NOT call `ios.Verbosef`, `fmt.Fprintln(os.Stderr, ...)`, or any other stderr-writing helper. A depguard rule SHALL forbid direct `os.Stdout`/`os.Stderr` access in `cmd/` outside `main.go`.

#### Scenario: Service layer silent

- **WHEN** user runs any command with `-v`
- **THEN** verbose output SHALL originate only from `cmd/*.go`
- **AND** no service-layer calls SHALL write to stderr

### Requirement: No debug prefix

The system SHALL NOT prepend `DEBUG:` or `[VERBOSE]` to verbose output. Messages SHALL be plain text with the dim style applied via lipgloss. The previous two-space indentation rule for level-2 details is removed (no level distinction anymore).

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Verbose output omits the DEBUG: prefix

- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** the rendered line SHALL contain "cloning <src>" with no `DEBUG:` or `[VERBOSE]` prefix

### Requirement: Structured debug logger channel

The system SHALL expose a `Logger *slog.Logger` field on the `iostreams.IOStreams` surface, separate from the boolean `Verbose` user-facing channel. The logger SHALL be wired by the composition root only when `TWIGGIT_DEBUG=1` is set; otherwise it SHALL be `nil` (or a discarding logger). When set, the logger SHALL emit `slog.LevelDebug` records to stderr, distinct from the dim-styled `Verbosef` user output. The two channels SHALL NOT interleave: `Verbosef` keeps the dim-style human-readable contract; `Logger.Debug` keeps the structured machine-readable contract.

#### Scenario: Logger nil when TWIGGIT_DEBUG unset

- **WHEN** `TWIGGIT_DEBUG` is unset or empty
- **THEN** `ios.Logger` SHALL be `nil` (or a discarding logger)
- **AND** `Logger.Debug(...)` SHALL be a no-op

#### Scenario: Logger active when TWIGGIT_DEBUG=1

- **WHEN** `TWIGGIT_DEBUG=1` is exported
- **THEN** `ios.Logger` SHALL be a non-nil `*slog.Logger`
- **AND** `Logger.Debug("opening repo", "path", p)` SHALL emit one structured record to stderr

#### Scenario: Logger output stays out of user stream

- **WHEN** `TWIGGIT_DEBUG=1` and `ios.Verbose == true`
- **THEN** `ios.Verbosef("cloning %s", src)` SHALL emit the dim-styled human line
- **AND** `ios.Logger.Debug(...)` SHALL emit the structured record
- **AND** the two outputs SHALL be distinguishable by content shape (structured key=value vs dim plain text)

### Requirement: Logger resource contract

`iostreams.NewLogger(w io.Writer)` SHALL `defer Close()` the underlying writer (when it implements `io.Closer`) immediately after the writer is acquired. Emitted log records SHALL carry lowercase messages without trailing punctuation. The `*slog.Logger` returned SHALL be a singleton reachable from `slog.Default()` after the binary entry point's `SetDefault` call (per `cli-iostreams`).

#### Scenario: Logger resource and lowercase contracts hold

- **WHEN** `main.go` constructs `iostreams.NewLogger(w)` where `w` is an `io.WriteCloser`
- **THEN** the calling `RunE` SHALL `defer logger.Close()` immediately after the constructor returns
- **AND** log records SHALL carry lowercase messages and lowercase attribute keys
