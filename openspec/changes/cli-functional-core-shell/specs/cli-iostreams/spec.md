# Spec Delta

## Purpose

Provides a TTY-aware I/O abstraction so commands can be tested without touching `os.Stdout`/`os.Stderr` directly, and so lipgloss rendering can be gated on whether the output is a real terminal or a pipe.

## ADDED Requirements

### Requirement: IOStreams exposes stdout, stderr, and TTY detection

The `iostreams.IOStreams` interface SHALL expose `Stdout() io.Writer`, `Stderr() io.Writer`, `In() io.Reader`, `IsStdoutTTY() bool`, and `IsStderrTTY() bool`. The concrete `System()` constructor SHALL bind these to `os.Stdout`, `os.Stderr`, `os.Stdin`, and `isatty` detection.

#### Scenario: System() binds to process stdio
- **WHEN** `main.go` calls `ios := iostreams.System()`
- **THEN** `ios.Stdout()` returns `os.Stdout`, `ios.Stderr()` returns `os.Stderr`, `ios.In()` returns `os.Stdin`, and the TTY flags reflect the actual terminal state

#### Scenario: TTY detection is false when stdout is piped
- **WHEN** the user runs `twiggit list -o json | jq`
- **THEN** `IsStdoutTTY()` SHALL return `false`, the lipgloss styles SHALL be disabled, and the JSON output SHALL be emitted without ANSI escapes

### Requirement: Stdout carries data; stderr carries errors and verbose output

Data output (tables, JSON, paths printed for `-C`) SHALL go to `IOStreams.Stdout`. Errors, hints, and verbose messages SHALL go to `IOStreams.Stderr`.

#### Scenario: List command splits stdout and stderr
- **WHEN** the user runs `twiggit list -o json | jq`
- **THEN** the JSON document lands on stdout and any error / verbose message lands on stderr; the pipeline parses cleanly

#### Scenario: Verbose output does not corrupt data output
- **WHEN** the user runs `twiggit list -vv -o json`
- **THEN** the JSON document on stdout contains only data; the verbose logs on stderr contain the diagnostic text

### Requirement: Test() helper returns buffers for command tests

`iostreams.Test()` SHALL return an `IOStreams` whose `Stdout()`, `Stderr()`, and `In()` are `*bytes.Buffer`. Tests SHALL use it instead of mocking `os.Stdout`.

#### Scenario: Test buffers capture both streams independently
- **WHEN** a test calls `ios := iostreams.Test()` and runs a command against it
- **THEN** the test asserts on `ios.Stdout().(*bytes.Buffer).String()` and `ios.Stderr().(*bytes.Buffer).String()` independently
