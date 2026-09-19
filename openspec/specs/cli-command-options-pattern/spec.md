# Capability: Command Options and Aliases

## Purpose

Naming, short-flag, and alias conventions for every cobra flag and
subcommand in twiggit, plus ergonomic flags that span multiple
subcommands. All flags SHALL follow a uniform set of conventions so
that users can predict `--help` output and tab completion behavior.

## Requirements

### Requirement: `--quiet` / `-q` global flag

The system SHALL expose `--quiet` (short `-q`) as a global persistent
flag on the root command. When set, the system SHALL suppress
non-essential output (success messages, hints, progress) but SHALL
preserve errors on stderr and any essential output (paths printed for
`-C` mode) on stdout. See `cli-quiet-mode`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `--verbose` / `-v` / `-vv` levels

The system SHALL expose `--verbose` (short `-v`) as a global persistent
flag. One `-v` SHALL set verbosity level 1 (high-level flow); `-vv` SHALL
set level 2 (parameters and intermediate steps). Verbose output SHALL
go to stderr only. See `cli-verbose-output`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Mutual exclusion of `--quiet` and `--verbose`

The system SHALL allow both `--quiet` and `--verbose` to be set;
verbose SHALL win over quiet when both are active.

#### Scenario: Quiet wins over verbose missing

- **WHEN** only `--quiet` is set
- **THEN** non-essential output SHALL be suppressed

#### Scenario: Verbose wins

- **WHEN** both `--quiet` and `--verbose` are set
- **THEN** verbose output SHALL appear on stderr
- **AND** quiet suppression SHALL be bypassed

### Requirement: Standard short flags

The system SHALL use the following short flags consistently across all
subcommands where the concept applies:

| Long | Short |
|---|---|
| `--force` | `-f` |
| `--yes` | `-y` |
| `--all` | `-a` |
| `--dry-run` | `-n` |
| `--output` | `-o` |
| `--cd` | `-C` |
| `--config` | `-c` |
| `--install` | `-i` |
| `--delete-branches` | `-d` |
| `--source` | (none; long-only) |
| `--merged-only` | `-m` |

#### Scenario: Short flag accepted

- **WHEN** user runs a subcommand with a recognized short flag
- **THEN** system SHALL honor it identically to the long form

### Requirement: Subcommand aliases

The system SHALL expose Unix-style aliases:

| Subcommand | Alias |
|---|---|
| `list` | `ls` |
| `delete` | `rm` |

#### Scenario: `ls` is `list`

- **WHEN** user runs `twiggit ls`
- **THEN** system SHALL behave identically to `twiggit list`

#### Scenario: `rm` is `delete`

- **WHEN** user runs `twiggit rm myproject/feature`
- **THEN** system SHALL behave identically to
  `twiggit delete myproject/feature`

### Requirement: Context-aware argument inference

The system SHALL infer project names from CWD when the user is inside a
project or worktree, using `application.ContextService`. When CWD is
outside any git context, an explicit project argument SHALL be
required.

#### Scenario: Infer from project context

- **WHEN** user runs `twiggit create feature` from inside a project
- **THEN** system SHALL infer the project from CWD
- **AND** SHALL create `~/Worktrees/<project>/feature`

#### Scenario: Outside git requires explicit project

- **WHEN** user runs `twiggit create feature` from outside any git context
- **THEN** system SHALL return a usage error suggesting `myproject/feature` form

### Requirement: Default source branch

The system SHALL default `--source` to `Config.DefaultSourceBranch`
(`main` by default) when the user does not specify one.

#### Scenario: Source defaults to main

- **WHEN** user runs `twiggit create myproject/feature` without `--source`
- **AND** `DefaultSourceBranch = "main"`
- **THEN** system SHALL create the worktree from `main`

#### Scenario: Custom default branch

- **WHEN** user sets `DefaultSourceBranch = "develop"`
- **AND** runs `twiggit create myproject/feature` without `--source`
- **THEN** system SHALL create the worktree from `develop`

### Requirement: Path output for `-C` flag

When the user passes `-C`/`--cd` to a subcommand, the system SHALL
print the absolute path of the target directory to stdout on success,
so the shell wrapper can navigate the user.

#### Scenario: `-C` after create

- **WHEN** user runs `twiggit create -C myproject/feature`
- **AND** creation succeeds
- **THEN** system SHALL print the new worktree path to stdout

### Requirement: Position shell argument vs `--shell=` flag

The `init` subcommand SHALL accept a positional `shell` argument and
SHALL NOT accept a `--shell=` flag. The `--install`, `--config`, and
`--force` flags on `init` are documented in `cli-init`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
