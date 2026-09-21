# Spec Delta

## Purpose

Coordinates SIGINT and SIGTERM cancellation through the cobra command
tree so long-running operations (bulk prune, init wrapper install,
worktree mutation) exit cleanly and the cmd layer can replace
`context.Background()` with a cancellation-aware `cmd.Context()`.

## ADDED Requirements

### Requirement: Signal-aware command context

The main entry point SHALL wrap the cobra command tree with
`signal.NotifyContext` so that SIGINT and SIGTERM cancel the
context returned by `cmd.Context()`. Every leaf command SHALL call
`cmd.Context()` instead of `context.Background()` to receive the
cancellation signal.

#### Scenario: SIGINT during a long-running prune
- **WHEN** the user runs `twiggit prune --all --yes` and presses
  Ctrl-C (SIGINT) while git operations are in flight
- **THEN** the cobra command's `ctx` is cancelled, in-flight
  service calls observe the cancellation, and the process exits
  cleanly with a non-zero exit code reflecting the cancellation

#### Scenario: SIGTERM from a process manager
- **WHEN** an external process manager sends SIGTERM to the
  running twiggit process during `twiggit create`
- **THEN** the cobra command's `ctx` is cancelled and the worktree
  state reflects either a fully-completed create or no create at
  all (atomic, never partially created)

### Requirement: TTY-gated interactive prompts

The prune command's interactive confirmation prompt
(`confirmBulkPrune`) SHALL gate on `isatty(os.Stdin)` before
reading from stdin. When stdin is not a TTY (a pipe, redirection,
or CI environment) AND the user did not pass `--yes` or `--force`,
the command SHALL exit with exit code 2 (usage error) and a
message directing the user to re-run with `--yes` or `--force`.

#### Scenario: Piped stdin without --yes
- **WHEN** the user runs `echo y | twiggit prune --all` (stdin is
  a pipe, not a TTY)
- **THEN** the command exits with exit code 2 and prints a message
  stating "interactive prompts require a TTY; re-run with --yes
  or --force"

#### Scenario: Interactive TTY with confirmation
- **WHEN** the user runs `twiggit prune --all` from an interactive
  shell (stdin is a TTY) and types `y`
- **THEN** the prune proceeds and writes the result to stdout/stderr
  per the existing `cli-prune` contract

### Requirement: Init wrapper install under signal

The shell wrapper install path (`runInitInstall`) SHALL observe
the signal context while writing the wrapper to the shell config
file. A SIGINT during the write SHALL leave the shell config file
either fully unchanged or fully written; never partially written.

#### Scenario: SIGINT mid-write during init wrapper install
- **WHEN** the user runs `twiggit init bash --install` and presses
  Ctrl-C while the wrapper is being appended to `~/.bashrc`
- **THEN** the wrapper write is either complete or not started;
  `~/.bashrc` is never left with a partial wrapper block
