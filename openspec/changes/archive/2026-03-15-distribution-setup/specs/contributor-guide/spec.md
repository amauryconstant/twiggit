## ADDED Requirements

### Requirement: CONTRIBUTING.md exists at repository root
A `CONTRIBUTING.md` file SHALL exist at the repository root with contributor documentation.

#### Scenario: File location
- **WHEN** a potential contributor views the repository
- **THEN** `CONTRIBUTING.md` SHALL be visible at the root level

### Requirement: Development setup documented
The contributor guide SHALL document the development environment setup including Go version and tooling.

#### Scenario: Prerequisites listed
- **WHEN** a developer reads the guide
- **THEN** they SHALL find requirements for Go 1.21+ and mise

### Requirement: Test commands documented
The contributor guide SHALL document all test commands.

#### Scenario: Test commands visible
- **WHEN** a developer reads the guide
- **THEN** they SHALL find commands for `mise run test`, `mise run test:full`, and `mise run check`

### Requirement: Code style documented
The contributor guide SHALL document code style requirements including linting.

#### Scenario: Lint commands visible
- **WHEN** a developer reads the guide
- **THEN** they SHALL find commands for `mise run lint:fix` and reference to golangci-lint

### Requirement: Pull request process documented
The contributor guide SHALL document the pull request process.

#### Scenario: PR workflow visible
- **WHEN** a developer reads the guide
- **THEN** they SHALL understand the PR submission and review process

### Requirement: Commit message conventions documented
The contributor guide SHALL document commit message conventions.

#### Scenario: Commit style visible
- **WHEN** a developer reads the guide
- **THEN** they SHALL understand expected commit message format
