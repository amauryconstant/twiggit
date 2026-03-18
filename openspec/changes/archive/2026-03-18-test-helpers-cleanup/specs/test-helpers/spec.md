## ADDED Requirements

### Requirement: Automatic resource cleanup
Test helpers SHALL use appropriate cleanup mechanisms for automatic resource cleanup when tests complete.

#### Scenario: RepoTestHelper constructor registers cleanup
- **GIVEN** a test requiring RepoTestHelper
- **WHEN** NewRepoTestHelper is called
- **THEN** a cleanup function is registered via t.Cleanup() to call the helper's Cleanup() method
- **AND** the cleanup function removes all created repositories
- **AND** cleanup runs even if the test fails or panics

#### Scenario: GitTestHelper uses t.TempDir for automatic cleanup
- **GIVEN** a test requiring GitTestHelper
- **WHEN** NewGitTestHelper is called
- **THEN** the helper uses t.TempDir() for the base directory
- **AND** the testing package automatically cleans up the temp directory when the test completes
- **AND** cleanup runs even if the test fails or panics

#### Scenario: Multiple cleanup functions execute in LIFO order
- **GIVEN** a test with multiple resources that register cleanup
- **WHEN** multiple resources are created in sequence
- **THEN** cleanup functions execute in reverse order of registration (LIFO)
- **AND** the last registered cleanup runs first

### Requirement: Helper function error line reporting
Test helper functions SHALL call t.Helper() to improve error line reporting.

#### Scenario: Helper function marks itself
- **GIVEN** a test that calls a helper function
- **WHEN** a helper function calls t.Helper()
- **THEN** error reports point to the calling test code
- **AND** not to the helper function internals

#### Scenario: Nested helper functions both call t.Helper()
- **GIVEN** a test that calls a helper function
- **WHEN** a helper function calls another helper function
- **THEN** both functions call t.Helper()
- **AND** error reports point to the original test code
