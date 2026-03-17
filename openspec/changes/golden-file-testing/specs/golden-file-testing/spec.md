## ADDED Requirements

### Requirement: Golden file comparison function
The test/helpers package SHALL provide a CompareGolden function that compares actual output against a golden file.

#### Scenario: Compare matching output
- **WHEN** CompareGolden is called with actual output matching golden file content
- **THEN** the test passes
- **AND** no error is reported

#### Scenario: Compare mismatching output
- **WHEN** CompareGolden is called with actual output not matching golden file content
- **THEN** the test fails
- **AND** a diff is displayed showing the difference

#### Scenario: Golden file not found
- **WHEN** CompareGolden is called with non-existent golden file
- **THEN** the test fails
- **AND** an error message indicates the missing file

### Requirement: UPDATE_GOLDEN environment variable
The CompareGolden function SHALL support an UPDATE_GOLDEN environment variable to update golden files instead of failing.

#### Scenario: Update golden file
- **WHEN** UPDATE_GOLDEN=true is set
- **AND** CompareGolden is called with mismatching output
- **THEN** the golden file is updated with actual output
- **AND** the test passes

#### Scenario: Create new golden file
- **WHEN** UPDATE_GOLDEN=true is set
- **AND** CompareGolden is called with non-existent golden file
- **THEN** the golden file is created with actual output
- **AND** the test passes

### Requirement: Golden file path resolution
Golden files SHALL be resolved relative to the test/golden/ directory with .golden extension.

#### Scenario: Resolve golden file path
- **WHEN** CompareGolden is called with goldenFile "list/basic-text"
- **THEN** the function resolves to test/golden/list/basic-text.golden

### Requirement: Priority coverage for CLI output
Golden file tests SHALL cover list command output (text and JSON formats), and error formatting.

#### Scenario: List command text output
- **WHEN** list command is executed with default output
- **THEN** output is verified against golden file

#### Scenario: List command JSON output
- **WHEN** list command is executed with --output json
- **THEN** output is verified against golden file

#### Scenario: Error formatting
- **WHEN** commands produce validation, service, or not-found errors
- **THEN** error output is verified against golden files

### Requirement: Mise tasks for golden testing
The project SHALL provide mise tasks for running and updating golden file tests.

#### Scenario: Run golden tests
- **WHEN** mise run test:golden is executed
- **THEN** all golden file tests are run
- **AND** tests fail on mismatch

#### Scenario: Update golden files
- **WHEN** mise run test:golden:update is executed
- **THEN** all golden file tests are run with UPDATE_GOLDEN=true
- **AND** golden files are updated
