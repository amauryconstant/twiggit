# Capability: Output Progress

## Purpose

Defines the progress reporter used by bulk operations (`prune --all`, future bulk delete). Lives in `cmd/suggestions.go` and follows the `cli-quiet-mode` suppression contract. The reporter SHALL write to `cmd.ErrOrStderr()` (never `os.Stderr` directly). The reporter's internal state SHALL be private and not exposed via getters.

## Requirements

### Requirement: ProgressReporter construction

`NewProgressReporter(quiet bool, ios *iostreams.IOStreams) *ProgressReporter` SHALL return a reporter that writes to `ios.ErrOut` when `quiet == false`. When `quiet == true`, every reporter method SHALL be a no-op so callers do not need to gate calls at the command site.

#### Scenario: Quiet reporter is silent

- **WHEN** `NewProgressReporter(true, ios)` is constructed
- **AND** the caller invokes `reporter.Report("step 1", 1, total)`
- **THEN** nothing SHALL be written to `ios.ErrOut`

### Requirement: Report emits one line per call

`(*ProgressReporter).Report(message string, current int, total int)` SHALL emit exactly one line to `ios.ErrOut` of the form `<message> (<current>/<total>)` terminated by `\n`. The rendered line SHALL be lowercase without trailing punctuation.

#### Scenario: Report renders one line

- **WHEN** `Report("pruning worktrees", 3, 10)` is invoked on a non-quiet reporter
- **THEN** `ios.ErrOut` SHALL contain exactly `pruning worktrees (3/10)\n`
- **AND** no other stream SHALL receive output

### Requirement: ReportStart and ReportComplete bracket a phase

`(*ProgressReporter).ReportStart(message string)` SHALL emit one line `<message>...` (no trailing newline-count). `(*ProgressReporter).ReportComplete(message string)` SHALL emit one line `<message>` complete`. These bracket a long-running phase. Per the `golang-cli` "single-handling-rule" rule, the reporter SHALL NOT log structured diagnostics; that channel is reserved for `iostreams.Logger`.

#### Scenario: Bracket phase

- **WHEN** a bulk prune begins with `ReportStart("Pruning merged worktrees")`
- **AND** ends with `ReportComplete("Prune complete")`
- **THEN** `ios.ErrOut` SHALL contain:
  ```
  Pruning merged worktrees...
  Prune complete
  ```
  (with trailing newlines after each line)

### Requirement: ProgressReporter does not allocate per-call closures

`(*ProgressReporter)` SHALL hold its `ios *iostreams.IOStreams` and `quiet bool` fields directly; no per-call closures or `slog` handler allocations. This keeps the bulk-prune path allocation-light across thousands of worktrees.

#### Scenario: Bulk-prune path allocation-light

- **WHEN** the bulk prune processes 1000 worktrees with `quiet == false`
- **THEN** `reporter.Report` SHALL write one line per worktree to `ios.ErrOut`
- **AND** the per-call allocation profile SHALL be O(1) — no per-worktree closures, goroutines, or `slog` handlers