# Spec Delta

## Purpose

Provides a TTY-aware I/O abstraction so commands can be tested without touching `os.Stdout`/`os.Stderr` directly, and so lipgloss rendering can be gated on whether the output is a real terminal or a pipe. Verbose output (`Verbosef`) and debug logging (`Logger`) are distinct channels; see `cli-verbose-output` for the verbose contract.

## ADDED Requirements

### Requirement: IOStreams struct with writers, TTY flags, and color state

The `iostreams.IOStreams` type SHALL be a struct exposing `In io.ReadCloser`, `Out io.Writer`, `ErrOut io.Writer`, `Verbose bool`, `Quiet bool`, `Logger *slog.Logger`, `colorEnabled bool`, `isStdoutTTY bool`, `isStderrTTY bool`, `styles *Styles`. Methods SHALL include `IsStdoutTTY() bool`, `IsStderrTTY() bool`, `IsInteractive() bool`, `Styles() *Styles`, `Verbosef(format string, args ...any)`, and `SetStdoutTTY(bool)`.

#### Scenario: System() binds to process stdio
- **WHEN** `main.go` calls `ios := iostreams.System()`
- **THEN** `ios.Out` equals `os.Stdout`, `ios.ErrOut` equals `os.Stderr`, `ios.In` equals `os.Stdin`, the TTY flags reflect the actual terminal state, and `colorEnabled` is `true` when `NO_COLOR` is unset and stdout is a TTY

#### Scenario: NO_COLOR disables styles
- **WHEN** `NO_COLOR=1` is set and `iostreams.System()` is constructed
- **THEN** `ios.colorEnabled == false`, `ios.Styles()` returns identity-rendering styles, and `ios.IsStdoutTTY() == true`

#### Scenario: TTY detection is false when stdout is piped
- **WHEN** the user runs `twiggit list -o json | jq`
- **THEN** `IsStdoutTTY()` SHALL return `false`, `colorEnabled` SHALL be `false`, the lipgloss styles SHALL be disabled, and the JSON output SHALL be emitted without ANSI escapes

### Requirement: IsInteractive requires both stdout and stdin TTY

`IOStreams.IsInteractive() bool` SHALL return `true` only when both stdout AND stdin are connected to a terminal. When either is piped or redirected, `IsInteractive()` SHALL return `false`. This matches the destruct-operation flow rule from `golang-cli-architecture`: prompt only when interactive; otherwise require `--force` or `--yes`.

#### Scenario: Interactive shell
- **WHEN** stdout is a TTY and stdin is a TTY
- **THEN** `ios.IsInteractive()` returns `true`

#### Scenario: Piped stdin is non-interactive
- **WHEN** stdin is piped (`echo feat/foo | twiggit create`)
- **THEN** `ios.IsInteractive()` returns `false` even if stdout is a TTY

### Requirement: Stdout carries data; stderr carries errors, hints, and verbose output

Data output (tables, JSON, JSON Lines, paths printed for `-C`) SHALL go to `IOStreams.Out`. Errors, hints, and verbose messages SHALL go to `IOStreams.ErrOut`. The `output.Formatter.Write` method receives `ios.Out` and SHALL NOT write to `ios.ErrOut`. The `output.FormatError` function receives `ios.ErrOut` and SHALL NOT write to `ios.Out`.

#### Scenario: List command splits stdout and stderr
- **WHEN** the user runs `twiggit list -o json | jq`
- **THEN** the JSON document lands on `ios.Out` and any error or verbose message lands on `ios.ErrOut`; the pipeline parses cleanly

#### Scenario: Verbose output does not corrupt data output
- **WHEN** the user runs `twiggit list -v -o json`
- **THEN** the JSON document on stdout contains only data; the verbose logs on stderr contain the diagnostic text

### Requirement: Verbosef emits a single newline-terminated line

`IOStreams.Verbosef(format string, args ...any)` SHALL write exactly one line to `ios.ErrOut` terminated by `\n` when `ios.Verbose == true` and SHALL be a no-op otherwise. The line is rendered through `ios.Styles().Dim(...)` so it is unobtrusive in a terminal. No `DEBUG:` or `[VERBOSE]` prefix is prepended. The previous `logv(cmd, level, format, args...)` helper is removed.

#### Scenario: Verbosef emits one line on -v
- **WHEN** `ios.Verbose == true` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** exactly one line "cloning <src>" lands on `ios.ErrOut` with a trailing newline

#### Scenario: Verbosef is silent without -v
- **WHEN** `ios.Verbose == false` and a command calls `ios.Verbosef("cloning %s", src)`
- **THEN** nothing is written to `ios.ErrOut`

### Requirement: Verbose and Logger are distinct channels

`IOStreams.Verbose bool` SHALL be set from the `-v` / `--verbose` flag and SHALL gate `Verbosef` output to stderr. `IOStreams.Logger *slog.Logger` SHALL be set from the Factory's lazy logger and SHALL be used for debug diagnostics, gated by the `TWIGGIT_DEBUG` environment variable. The Factory SHALL construct the logger with level `slog.LevelDebug` when `TWIGGIT_DEBUG` is set, otherwise `slog.LevelWarn`.

#### Scenario: Logger emits only when TWIGGIT_DEBUG is set
- **WHEN** `TWIGGIT_DEBUG=1` is set and a command calls `ios.Logger.Debug("msg", "key", "value")`
- **THEN** the structured log line lands on `ios.ErrOut`

#### Scenario: Logger is silent without TWIGGIT_DEBUG
- **WHEN** `TWIGGIT_DEBUG` is unset and a command calls `ios.Logger.Debug("msg")`
- **THEN** nothing is written to `ios.ErrOut`

### Requirement: Test() returns buffers for command tests

`iostreams.Test()` SHALL return `(*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer)` where the second return is stdin, third is stdout, fourth is stderr. Tests SHALL use it instead of mocking `os.Stdout`. The returned IOStreams SHALL have `Verbose=false`, `Quiet=false`, `colorEnabled=false`, and `Logger=nil`.

#### Scenario: Test buffers capture both streams independently
- **WHEN** a test calls `ios, stdin, stdout, stderr := iostreams.Test()` and runs a command against it
- **THEN** the test asserts on `stdout.String()` and `stderr.String()` independently
