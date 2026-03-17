## ADDED Requirements

### Requirement: Automatic resource cleanup
Test helpers SHALL use t.Cleanup() for automatic resource cleanup when tests complete.

#### Scenario: GitTestHelper constructor registers cleanup
- **WHEN** NewGitTestHelper is called
- **THEN** a cleanup function is registered via t.Cleanup()
- **AND** the cleanup function calls the helper's Cleanup method
- **AND** cleanup runs even if the test fails or panics

#### Scenario: Multiple cleanup functions execute in LIFO order
- **WHEN** multiple resources are created in sequence
- **THEN** cleanup functions execute in reverse order of registration
- **AND** the last registered cleanup runs first

#### Scenario: Cleanup runs on test failure
- **WHEN** a test fails or panics
- **THEN** all registered cleanup functions still execute
- **AND** no resources are leaked

### Requirement: Helper function error line reporting
Test helper functions SHALL call t.Helper() to improve error line reporting.

#### Scenario: Helper function marks itself
- **WHEN** a helper function calls t.Helper()
- **THEN** error reports point to the calling test code
- **AND** not to the helper function internals

#### Scenario: Nested helper functions both call t.Helper()
- **WHEN** a helper function calls another helper function
- **THEN** both functions call t.Helper()
- **AND** error reports point to the original test code
