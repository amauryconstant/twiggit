## ADDED Requirements

### Requirement: Project Discovery

The system SHALL discover and validate git projects by name or from current context, supporting both projects directory lookup and contextual inference.

#### Scenario: Discover project by name from projects directory

- **WHEN** user specifies a project name explicitly (e.g., "myproject")
- **AND** project exists in the configured projects directory (default: `~/Projects/`)
- **THEN** system SHALL return project information
- **AND** project path SHALL be set to the projects directory location
- **AND** project git repository path SHALL be identified
- **AND** project SHALL include list of worktrees, branches, and remotes

#### Scenario: Discover project from context when in project directory

- **WHEN** user does not specify a project name
- **AND** system detects current context as `ContextProject`
- **THEN** system SHALL use the detected project from context
- **AND** project information SHALL be returned for the current directory's project
- **AND** project name SHALL match the directory name containing `.git/`

#### Scenario: Discover project from context when in worktree

- **WHEN** user does not specify a project name
- **AND** system detects current context as `ContextWorktree`
- **THEN** system SHALL resolve the main repository from the worktree
- **AND** project information SHALL be returned for the main repository (not the worktree)
- **AND** project path SHALL point to the main project directory in `~/Projects/`

#### Scenario: Return error for project name required outside git context

- **WHEN** user does not specify a project name
- **AND** system detects current context as `ContextOutsideGit`
- **THEN** system SHALL return error indicating project name is required
- **AND** error message SHALL specify that project name is needed when outside git context
- **AND** suggestions SHALL be provided for how to specify the project

#### Scenario: Return error for nonexistent project

- **WHEN** user specifies a project name that does not exist in the projects directory
- **THEN** system SHALL return error indicating project not found
- **AND** error SHALL include the specified project name
- **AND** system SHALL suggest verifying the project name and directory

#### Scenario: List all available projects

- **WHEN** system needs to enumerate all projects
- **THEN** system SHALL scan the configured projects directory
- **AND** only directories containing valid git repositories SHALL be included
- **AND** system SHALL return list of project information for each valid project
- **AND** empty list SHALL be returned if no projects exist

#### Scenario: Skip non-git directories in projects directory

- **WHEN** scanning projects directory for projects
- **AND** a directory does not contain a valid git repository
- **THEN** system SHALL skip that directory
- **AND** it SHALL NOT be included in the projects list
- **AND** system SHALL continue scanning other directories

#### Scenario: Validate project repository

- **WHEN** system validates a project path
- **AND** path points to a valid git repository
- **THEN** validation SHALL succeed
- **AND** system SHALL proceed with project operations

#### Scenario: Return error for invalid project repository

- **WHEN** system validates a project path
- **AND** path does not point to a valid git repository
- **THEN** validation SHALL fail
- **AND** error SHALL indicate "invalid git repository"
- **AND** underlying git validation error SHALL be included

#### Scenario: Case-insensitive project name matching

- **WHEN** user specifies a project name with different case than the actual directory
- **THEN** system SHALL perform case-insensitive lookup
- **AND** project SHALL be found if case-insensitive match exists
- **AND** actual project directory name SHALL be used in returned information

#### Scenario: Resolve main repository from worktree

- **WHEN** discovering project from a worktree context
- **THEN** system SHALL identify the parent main repository
- **AND** main repository SHALL be located in the projects directory (not worktrees)
- **AND** worktree SHALL map back to its parent project

#### Scenario: Get detailed project information

- **WHEN** system retrieves project information
- **THEN** result SHALL include project name, path, and git repository path
- **AND** result SHALL include list of all worktrees for the project
- **AND** result SHALL include list of all branches
- **AND** result SHALL include list of all remotes
- **AND** result SHALL include the default branch name
- **AND** result SHALL indicate whether repository is bare
