# Standard Go Unit Tests

## Purpose

Define patterns and requirements for Go unit tests using standard testing patterns instead of testify/suite.

## ADDED Requirements

### Requirement: Standard Go testing pattern

All unit tests SHALL use standard Go testing with table-driven patterns instead of testify/suite.

#### Scenario: Test function uses testing.T

- **GIVEN** a unit test file needs to be created or modified
- **WHEN** the test function is defined
- **THEN** it SHALL use `func TestXxx(t *testing.T)` signature

#### Scenario: Subtests use t.Run()

- **GIVEN** a test file contains multiple test cases
- **WHEN** the test is executed
- **THEN** each case SHALL be wrapped in `t.Run(name, func(t *testing.T) { ... })`

### Requirement: Fresh dependencies per subtest

Each subtest SHALL create fresh dependencies via constructor injection or t.TempDir().

#### Scenario: Mocks created per subtest

- **GIVEN** a subtest requires a mock
- **WHEN** the subtest is executed
- **THEN** the mock SHALL be created within the `t.Run()` closure
- **AND** no mocks SHALL be shared between subtests

#### Scenario: TempDir for filesystem isolation

- **GIVEN** a test requires filesystem operations
- **WHEN** the test is executed
- **THEN** it SHALL use `t.TempDir()` for automatic cleanup

### Requirement: Mock assertions via t.Cleanup()

Mock assertions SHALL be registered via t.Cleanup() for automatic verification.

#### Scenario: Cleanup registered at mock creation

- **GIVEN** a mock is created in a test
- **WHEN** the mock is created
- **THEN** `t.Cleanup(func() { mock.AssertExpectations(t) })` SHALL be called immediately after mock creation

#### Scenario: Cleanup runs on test failure

- **WHEN** a test fails or panics
- **THEN** registered cleanup functions SHALL still execute
- **AND** mock assertions SHALL still be verified

### Requirement: Table-driven pattern for 5+ cases

Tests with 5 or more variations SHALL use table-driven pattern.

#### Scenario: Tests slice defines cases

- **GIVEN** a function has 5+ test variations
- **WHEN** the test is written
- **THEN** tests SHALL be defined in a `tests` slice of structs
- **AND** each struct SHALL have a `name` field

#### Scenario: Loop executes test cases

- **GIVEN** table-driven tests are defined
- **WHEN** the test is executed
- **THEN** a `for _, tt := range tests` loop SHALL iterate over cases
- **AND** `t.Run(tt.name, ...)` SHALL execute each case

### Requirement: testify/suite removal

The testify/suite import SHALL be removed after conversion.

#### Scenario: No suite import in test files

- **WHEN** conversion is complete
- **THEN** no test file SHALL import `github.com/stretchr/testify/suite`

#### Scenario: No suite.Suite embedding

- **WHEN** conversion is complete
- **THEN** no test struct SHALL embed `suite.Suite`

#### Scenario: No suite.Run() calls

- **WHEN** conversion is complete
- **THEN** no test SHALL call `suite.Run(t, &suiteStruct)`
