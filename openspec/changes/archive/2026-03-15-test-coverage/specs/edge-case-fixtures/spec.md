## ADDED Requirements

### Requirement: Corrupted repository handling
The system SHALL gracefully handle operations on repositories with corrupted git objects.

#### Scenario: List worktrees on corrupted repository
- **WHEN** user runs list command on repository with corrupted .git/objects
- **THEN** system returns descriptive error
- **AND** system does not panic

#### Scenario: Detect context in corrupted repository
- **WHEN** context detection runs in corrupted repository
- **THEN** system identifies as git repository context
- **AND** handles read errors gracefully

### Requirement: Bare repository handling
The system SHALL gracefully handle operations on bare git repositories.

#### Scenario: List worktrees on bare repository
- **WHEN** user runs list command on bare repository
- **THEN** system returns appropriate error or empty list
- **AND** system does not panic

#### Scenario: Create worktree on bare repository
- **WHEN** user attempts to create worktree on bare repository
- **THEN** system returns descriptive error
- **AND** explains bare repository limitation

### Requirement: Submodule repository handling
The system SHALL handle repositories containing git submodules correctly.

#### Scenario: List worktrees with submodules
- **WHEN** user runs list command on repository with submodules
- **THEN** system lists main repository worktrees
- **AND** submodules are handled appropriately

#### Scenario: Context detection with submodules
- **WHEN** context detection runs in repository with submodules
- **THEN** system correctly identifies main repository context
- **AND** does not confuse submodule for main project

### Requirement: Detached HEAD handling
The system SHALL gracefully handle repositories in detached HEAD state.

#### Scenario: List worktrees in detached HEAD
- **WHEN** user runs list command while in detached HEAD state
- **THEN** system lists worktrees correctly
- **AND** detached HEAD is reported appropriately

#### Scenario: Context detection in detached HEAD
- **WHEN** context detection runs in detached HEAD state
- **THEN** system correctly identifies context
- **AND** branch name may be reported as detached or HEAD
