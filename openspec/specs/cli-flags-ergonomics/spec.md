# Capability: Flags Ergonomics

## Purpose

Cross-cutting ergonomic improvements spanning multiple subcommands that
do not fit a single user-visible flag: source-branch defaulting, prune
preview-then-confirm flow, and short-flag coverage. Conventions for
short flags and aliases live in `cli-command-options-pattern`; this spec
captures the behavior that lives in the cmd layer across commands.

## Requirements

### Requirement: Source branch defaults to configured principal

The system SHALL default `--source` to `Config.DefaultSourceBranch`
when the user does not pass `--source`. This applies to `create`,
`delete`, and `prune` operations that need a base branch. See
`cli-command-options-pattern` for the flag convention; this requirement
is the cross-cutting behavioral contract.

#### Scenario: Default to main

- **WHEN** `Config.DefaultSourceBranch = "main"` (default)
- **AND** user runs `twiggit create myproject/feature` without `--source`
- **THEN** system SHALL create the worktree from `main`

#### Scenario: Custom default branch

- **WHEN** user sets `DefaultSourceBranch = "develop"` in config
- **AND** runs `twiggit create myproject/feature` without `--source`
- **THEN** system SHALL create from `develop`

### Requirement: Prune preview-then-confirm

The system SHALL run a dry-run preview to stderr before bulk prune
operations when the user has not passed `--force`, `--yes`, or
`--dry-run`, and SHALL prompt for confirmation on stdin. See `cli-prune`
for the full prune behavior.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Short-flag coverage

Every common operation flag SHALL have a short alias. The full mapping
is owned by `cli-command-options-pattern`; this requirement restates
that ergonomic short forms SHALL be available on the four high-traffic
flags: `-f`, `-y`, `-a`, `-n`.

#### Scenario: Short flags cover all ergonomic cases

- **WHEN** user runs any of `twiggit prune -f`, `-y`, `-a`, `-n`
- **THEN** each short flag SHALL be honored identically to its long form

### Requirement: Help examples

Every subcommand's `--help` output SHALL include at least one example
invocation showing the most common usage. Examples SHALL be realistic
strings (not abstract `<value>` placeholders) when possible.

#### Scenario: Prune help shows examples

- **WHEN** user runs `twiggit prune --help`
- **THEN** output SHALL include at least one example block
  covering: default prune, `--all`, and `--delete-branches`

### Requirement: Progress reporting via stderr

The system SHALL report progress for bulk operations (prune `--all`,
etc.) to stderr only, SHALL suppress progress under `--quiet`, and
SHALL NOT print progress to stdout (stdout is reserved for data and
`-C` navigation paths). See `cli-quiet-mode`.

#### Scenario: Bulk progress visible

- **WHEN** user runs `twiggit prune --all` without `--quiet`
- **THEN** "Pruning merged worktrees..." SHALL appear on stderr

#### Scenario: Bulk progress suppressed

- **WHEN** user runs `twiggit prune --all --quiet`
- **THEN** progress messages SHALL NOT appear on stderr
