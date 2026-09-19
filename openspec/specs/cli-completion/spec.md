# Capability: Shell Completion

## Purpose

Carapace-driven completion for twiggit subcommands and flags, with
configurable per-call timeout, 5s caching, fuzzy matching, smart sort,
and project/branch exclusion lists. Long-form completion scripts are
exposed via the `twiggit completion <shell>` subcommand; positional
project/branch completion is provided via Carapace action callbacks.

## Requirements

### Requirement: Long-form completion via `twiggit completion <shell>`

The system SHALL expose a `completion` subcommand that emits the Carapace
completion script for one of the supported shells.

#### Scenario: Emit bash completion

- **WHEN** user runs `twiggit completion bash`
- **THEN** system SHALL print the Carapace bash completion script to stdout
- **AND** exit with status 0

#### Scenario: Emit zsh completion

- **WHEN** user runs `twiggit completion zsh`
- **THEN** system SHALL print the Carapace zsh completion script to stdout

#### Scenario: Emit fish completion

- **WHEN** user runs `twiggit completion fish`
- **THEN** system SHALL print the Carapace fish completion script to stdout

#### Scenario: Emit unsupported shell

- **WHEN** user runs `twiggit completion ksh` (unsupported)
- **THEN** system SHALL return an error naming unsupported shell
- **AND** exit with non-zero status

### Requirement: Configurable completion timeout

The system SHALL read `CompletionConfig.Timeout` (Go duration string)
for per-completion-call timeout. If unset or unparseable, the system
SHALL fall back to the 500ms default. Slow git operations that exceed
the timeout SHALL gracefully degrade to an empty suggestion list rather
than block the user's shell.

#### Scenario: Custom timeout from config

- **WHEN** `config.Completion.Timeout = "300ms"`
- **THEN** completion callbacks SHALL enforce a 300ms per-call budget
- **AND** SHALL return empty suggestions on timeout (not error)

#### Scenario: Default 500ms timeout

- **WHEN** `config.Completion.Timeout` is empty or unparseable
- **THEN** system SHALL use 500ms per call

#### Scenario: Graceful timeout degradation

- **WHEN** a completion callback exceeds the timeout
- **THEN** system SHALL return an empty ActionValues()
- **AND** SHALL NOT propagate the error to the shell

### Requirement: Cached completion actions

The system SHALL cache completion action results for 5 seconds using
Carapace's `.Cache(5s)` to absorb rapid repeated tab presses without
re-running git operations.

#### Scenario: Cache hit

- **WHEN** user hits tab twice within 5 seconds for the same input
- **THEN** second call SHALL use the cached result
- **AND** SHALL NOT re-execute the underlying git operation

#### Scenario: Cache miss after expiry

- **WHEN** user hits tab after the 5-second TTL expires
- **THEN** system SHALL re-execute the underlying git operation

### Requirement: Fuzzy match with exclusion

The system SHALL perform case-insensitive subsequence fuzzy matching
for partial inputs, gated by `NavigationConfig.FuzzyMatching`, and
SHALL exclude branches matching any glob in
`CompletionConfig.ExcludeBranches` and projects matching any glob in
`CompletionConfig.ExcludeProjects`.

#### Scenario: Fuzzy match enabled

- **WHEN** `NavigationConfig.FuzzyMatching = true`
- **AND** user types `twiggit cd f1<TAB>` from a project with branch `feature-1`
- **THEN** `feature-1` SHALL appear in suggestions

#### Scenario: Fuzzy match disabled

- **WHEN** `NavigationConfig.FuzzyMatching = false` (default)
- **AND** user types `twiggit cd f1<TAB>` from a project with branch `feature-1`
- **THEN** `feature-1` SHALL NOT appear unless the prefix matches

#### Scenario: Branch exclusion

- **WHEN** `CompletionConfig.ExcludeBranches = ["dependabot/*"]`
- **THEN** suggestions SHALL NOT include any branch matching `dependabot/*`

#### Scenario: Project exclusion

- **WHEN** `CompletionConfig.ExcludeProjects = ["archive/*"]`
- **THEN** suggestions SHALL NOT include any project matching `archive/*`

### Requirement: Smart sort order

The system SHALL sort completion suggestions with the priority:
current worktree first, default source branch second, remaining
alphabetically.

#### Scenario: Current worktree pinned first

- **WHEN** suggestions include the user's current worktree plus others
- **THEN** current worktree SHALL be first in the list

#### Scenario: Default branch second

- **WHEN** suggestions include the default source branch (`main` by default)
- **AND** current worktree is not among suggestions
- **THEN** default branch SHALL appear before other branches

### Requirement: Progressive project/branch completion

The system SHALL use Carapace's `ActionMultiParts("/")` to support
progressive completion of `project/branch` targets, with projects
emitting a trailing `/` suffix.

#### Scenario: Project suggestion with suffix

- **WHEN** user types `twiggit cd my<TAB>` from outside any git context
- **AND** a project `myapp` exists
- **THEN** `myapp/` SHALL appear as a suggestion (with trailing slash)

#### Scenario: Branch completion for selected project

- **WHEN** user types `twiggit cd myapp/fe<TAB>`
- **THEN** suggestions SHALL be limited to branches of project `myapp`
- **AND** `myapp/` prefix SHALL NOT appear in branch suggestions

#### Scenario: No project suggestions when scoped

- **WHEN** user is completing the branch segment (second segment typed)
- **THEN** project suggestions SHALL be filtered out

### Requirement: Existing-only filter

The system SHALL support `WithExistingOnly()` (from `domain`) which
filters completion suggestions to materialized worktrees, excluding
remote branches and stale entries.

#### Scenario: Existing-only on delete

- **WHEN** `delete` and `prune` use `WithExistingOnly()`
- **THEN** only materialized worktrees SHALL be offered
- **AND** branches without worktrees SHALL NOT be offered

### Requirement: Shell auto-detection for init

When `twiggit init` auto-detects the shell from `$SHELL`, the system
SHALL accept any path containing `bash`, `zsh`, or `fish` substring
and SHALL map to the corresponding `domain.ShellType`. See
`infrastructure-shell-detect` for the canonical detection rules.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Hidden Carapace entry point

The system SHALL expose `twiggit _carapace <shell>` as a hidden command
for direct Carapace snippet generation, used by the `init` stdout mode
to embed completion in shell configs.

#### Scenario: Carapace snippet

- **WHEN** user runs `twiggit _carapace zsh`
- **THEN** system SHALL emit Carapace's zsh snippet to stdout
- **AND** exit with status 0
