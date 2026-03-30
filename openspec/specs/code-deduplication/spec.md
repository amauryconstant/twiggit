# Code Deduplication

## Purpose

Eliminate duplicate code patterns by extracting shared logic into reusable helper functions.

## Requirements

### Requirement: Shell auto-detection logic SHALL be shared

The shell auto-detection logic SHALL exist in a single shared function rather than duplicated across SetupShell and ValidateInstallation.

#### Scenario: SetupShell and ValidateInstallation use same auto-detection
- **WHEN** both SetupShell and ValidateInstallation run
- **THEN** same auto-detection logic is used for finding shell configuration
- **AND** no code duplication exists between the two functions

#### Scenario: Auto-detection returns consistent results
- **WHEN** auto-detection logic is called
- **THEN** same shell type is detected regardless of which function calls it

### Requirement: Navigation target resolution SHALL be shared

The logic for finding "main" project path used in delete and prune commands SHALL be extracted to a shared utility function.

#### Scenario: Delete and prune resolve main project path identically
- **WHEN** both delete and prune commands need to find main project path
- **THEN** same resolution logic is used
- **AND** no duplicate implementation exists

### Requirement: Path validation logic SHALL be shared

Path validation logic duplicated between resolveWorktreePath and resolveCrossProjectReference SHALL be extracted to a shared function.

#### Scenario: Path validation consistency
- **WHEN** resolving paths in different contexts
- **THEN** same validation rules are applied
- **AND** duplicate validation code is eliminated
