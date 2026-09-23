# Spec Delta

## Purpose

Defines the Factory composition pattern with lazy initialization for CLI command dependencies, giving each command a single injection point for config, git client, IOStreams, and output formatters without an interface explosion.

## ADDED Requirements

### Requirement: Factory exposes lazy function fields

The `cmdutil.Factory` type SHALL expose one lazy `func() T` field per dependency (`Config`, `GitClient`, `IOStreams`, `Formatter`, `FormatError`). The fields SHALL be assigned at construction time and invoked at first use.

#### Scenario: Lazy field is nil before first call
- **WHEN** a command reads `f.Config` before invoking it
- **THEN** the field returns the zero value and the command MUST call the function form (`f.Config()`) to initialize it

#### Scenario: Lazy field caches its result across calls
- **WHEN** a command calls `f.Config()` twice during one binary invocation
- **THEN** the second call SHALL return the same `*core.Config` pointer as the first call

### Requirement: Config is cached via sync.OnceValue

`Factory.Config` SHALL use `sync.OnceValue(func() (*core.Config, error))` to guarantee the config file is read and parsed exactly once per binary invocation, even when multiple commands access it.

#### Scenario: Config file is read at most once
- **WHEN** three commands each call `f.Config()` in sequence
- **THEN** the underlying koanf loader SHALL read `$XDG_CONFIG_HOME/twiggit/config.toml` exactly once

#### Scenario: Config failure surfaces on first access
- **WHEN** the config file is invalid TOML
- **THEN** the first call to `f.Config()` SHALL return a `*core.OperationError` wrapping the parse error; subsequent calls SHALL return the same error without re-reading the file

### Requirement: Factory is the only place where infrastructure is instantiated

Every `os/exec`, `git.PlainOpen`, file read, and XDG path lookup SHALL happen inside a Factory lazy function. `cmd/*.go` files SHALL NOT construct infrastructure types directly.

#### Scenario: cmd/ files do not import os/exec directly
- **WHEN** `golangci-lint run` enforces the Tier 2 depguard
- **THEN** `cmd/*.go` SHALL NOT import `os/exec` directly; the `CommandExecutor` is the only path

#### Scenario: Factory is the wiring seam for tests
- **WHEN** a test invokes `NewCmdList(testFactory, runF)` with a Factory whose fields are pre-set to test doubles
- **THEN** the command runs against the test doubles without spawning real git subprocesses
