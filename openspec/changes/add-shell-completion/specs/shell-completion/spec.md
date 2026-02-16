## ADDED Requirements

### Requirement: Shell Completion Generation

The system SHALL provide shell completion scripts via Carapace's hidden `_carapace` command, supporting bash, zsh, fish, powershell, nushell, elvish, and other shells.

#### Scenario: Generate bash completion script

- **WHEN** user runs `twiggit _carapace bash` (or sources output via shell hook)
- **THEN** system SHALL output bash completion script to stdout
- **AND** script SHALL be sourceable in bash shell

#### Scenario: Generate zsh completion script

- **WHEN** user runs `twiggit _carapace zsh`
- **THEN** system SHALL output zsh completion script to stdout
- **AND** script SHALL be sourceable in zsh shell

#### Scenario: Generate fish completion script

- **WHEN** user runs `twiggit _carapace fish`
- **THEN** system SHALL output fish completion script to stdout
- **AND** script SHALL be sourceable in fish shell

#### Scenario: Generate completion for additional shells

- **WHEN** user runs `twiggit _carapace <shell>` where shell is nushell, elvish, powershell, tcsh, oil, xonsh, or cmd-clink
- **THEN** system SHALL output appropriate completion script to stdout
- **AND** script SHALL be sourceable in the target shell

### Requirement: Command Argument Completion

The system SHALL provide context-aware tab completion for command arguments via Carapace's `PositionalCompletion`.

#### Scenario: Complete cd command from project context

- **WHEN** user presses tab after `twiggit cd ` from within a project directory
- **THEN** system SHALL suggest branch names for current project
- **AND** system SHALL include "main" as a suggestion
- **AND** suggestions SHALL include descriptions (e.g., "Project root directory", "Worktree for branch X")

#### Scenario: Complete cd command from worktree context

- **WHEN** user presses tab after `twiggit cd ` from within a worktree
- **THEN** system SHALL suggest sibling worktree branch names
- **AND** system SHALL NOT include "main" (already at project root)

#### Scenario: Complete cd command from outside git context

- **WHEN** user presses tab after `twiggit cd ` from outside any git repository
- **THEN** system SHALL suggest project names from configured projects directory

#### Scenario: Complete create command argument

- **WHEN** user presses tab after `twiggit create ` from within a project directory
- **THEN** system SHALL suggest branch names (indicating worktree creation)
- **AND** system SHALL NOT include branches that already have worktrees (unless showing all)

#### Scenario: Complete delete command argument

- **WHEN** user presses tab after `twiggit delete ` from within a project directory
- **THEN** system SHALL suggest ONLY existing worktrees
- **AND** system SHALL NOT suggest branches without materialized worktrees

### Requirement: Flag Completion

The system SHALL provide tab completion for flag values via Carapace's `FlagCompletion`.

#### Scenario: Complete --source flag

- **WHEN** user presses tab after `twiggit create feature-1 --source ` from within a project
- **THEN** system SHALL suggest branch names available as source branches
- **AND** suggestions SHALL include all local branches in the repository

### Requirement: Prune Command Completion

The system SHALL provide tab completion for prune command arguments via Carapace, filtering to existing worktrees only.

#### Scenario: Complete prune command argument from project context

- **WHEN** user presses tab after `twiggit prune ` from within a project directory
- **THEN** system SHALL suggest ONLY existing worktrees for the current project
- **AND** system SHALL NOT suggest branches without materialized worktrees
- **AND** system SHALL NOT include the project root ("main") as a suggestion

#### Scenario: Complete prune command argument with cross-project prefix

- **WHEN** user presses tab after `twiggit prune myproject/` from outside any project
- **THEN** system SHALL suggest existing worktrees for the specified project
- **AND** suggestions SHALL be formatted as `<project>/<branch>`

### Requirement: Completion Descriptions

The system SHALL include helpful descriptions with completion suggestions.

#### Scenario: Show description for branch suggestions

- **WHEN** completion suggests a branch name
- **THEN** description SHALL indicate whether it is a worktree or unmaterialized branch
- **AND** description format for worktree SHALL be "Worktree for branch <branch>"
- **AND** description format for unmaterialized branch SHALL be "Branch <branch> (create worktree)"

#### Scenario: Show description for main suggestion

- **WHEN** completion suggests "main"
- **THEN** description SHALL be "Project root directory"

#### Scenario: Show description for project suggestions

- **WHEN** completion suggests a project name
- **THEN** description SHALL be "Project directory"

### Requirement: Completion Performance

The system SHALL cache completion results for acceptable performance.

#### Scenario: Cache branch suggestions

- **WHEN** completion fetches branch names from git
- **THEN** system SHALL cache results for 5 seconds
- **AND** subsequent tab presses within cache window SHALL use cached data
- **AND** cache misses SHALL trigger fresh git operations with 500ms timeout protection

#### Scenario: Cache hit avoids timeout

- **WHEN** user requests completion within 5 seconds of previous completion
- **THEN** system SHALL return cached results immediately
- **AND** system SHALL not perform git operations
- **AND** 500ms timeout SHALL not apply to cache hits

#### Scenario: Timeout protection for slow operations

- **WHEN** git operations take longer than 500ms during completion
- **THEN** system SHALL gracefully degrade by returning empty suggestions
- **AND** system SHALL not return partial results that might be incomplete
- **AND** user experience SHALL not be degraded by slow git operations
