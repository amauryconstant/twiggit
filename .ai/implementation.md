# Twiggit Implementation Requirements

## Testing Requirements

### Test Coverage
- >80% test coverage SHOULD be achieved for business logic
- Happy paths and critical business logic SHALL have 100% coverage
- Unit, integration, and E2E tests SHALL be included
- External dependencies SHALL be mocked in unit tests OR tested via integration tests
- Real git repositories SHALL be used in integration tests
- CLI behavior SHALL be validated through E2E tests (cmd/ package uses E2E tests only)
- Tests that fail in CI environment SHALL NOT be committed

### Testing Framework Usage
- Testify SHALL be used for unit testing with mock generation
- Testify SHALL be used for integration testing with structured test organization
- Ginkgo/Gomega SHALL be used for E2E testing with BDD patterns and CLI interaction testing
- Gexec SHALL be used for E2E testing with real binary execution
- Test organization SHALL follow Go testing best practices
- All tests SHALL be runnable with standard Go test commands
- Alternative testing frameworks MAY be used for specific test scenarios
- Parallel test execution MAY be implemented for performance optimization

### Test Data Management
- Test fixtures SHALL be created in temporary directories
- Test repositories SHALL be cleaned up after test execution
- Test data SHALL be isolated between test cases
- Real-world scenarios SHALL be represented in test cases

### Context Detection Testing Requirements
- Context detection logic SHALL be tested with unit tests
- Project folder detection SHALL be tested with mocked .git/ directories
- Worktree folder detection SHALL be tested with pattern matching
- Edge cases SHALL be tested: nested directories, multiple .git dirs, broken repos
- Context detection SHALL be tested with integration tests using real git repositories
- Context behavior SHALL be validated for all three contexts: project, worktree, outside git

### Shell Integration Testing Requirements
- Shell type validation SHALL be tested for bash, zsh, and fish
- Shell wrapper function generation SHALL be tested with explicit shell types
- Escape hatch behavior SHALL be tested with builtin cd functionality
- Shell configuration file modification SHALL be tested with temporary files
- Directory navigation output SHALL be tested for correct path format
- Shell wrapper installation SHALL be tested with integration tests

## Quality Requirements

### Code Quality
- All golangci-lint checks SHALL pass
- Go standard project structure SHALL be followed
- Business logic SHALL be separated from infrastructure
- Clear interfaces SHALL be defined for external dependencies
- Comprehensive godoc comments SHALL be included
- Compiler warnings or lint errors SHALL NOT be ignored
- Additional static analysis tools MAY be integrated beyond golangci-lint
- Custom linting rules MAY be added for project-specific requirements

### Code Organization
- Business logic SHALL be placed in internal/domain/
- Infrastructure implementations SHALL be placed in internal/infrastructure/
- CLI commands SHALL be placed in cmd/
- Interfaces SHALL be defined in internal/infrastructure/interfaces.go
- Standard Go project structure SHALL be followed
- Package boundaries SHALL be respected

### Documentation
- All exported functions and types SHALL have godoc comments
- Complex algorithms SHALL be documented with usage examples
- Configuration options SHALL be documented with examples
- Error types SHALL be documented with recovery suggestions

## Performance Requirements

### Startup Performance
- Startup time SHALL be <100ms on modern hardware
- Initial configuration loading SHALL be optimized for speed
- Dependency initialization SHALL be minimized
- No blocking operations SHALL occur during startup

### Operational Performance
- List operations SHALL complete in <500ms for <100 worktrees
- Create operations SHALL complete in <2s for typical repositories
- Delete operations SHALL complete in <1s for typical repositories
- Memory usage SHALL remain <50MB during operations
- Main thread SHALL NOT be blocked for long-running operations

### Resource Management
- File handles SHALL be properly closed after use
- Git repositories SHALL be properly closed after operations
- Memory SHALL be efficiently managed for large repositories
- Concurrent operations SHALL be properly synchronized

## Security Requirements

### Input Validation
- All user inputs SHALL be validated to prevent injection attacks
- Path traversal attacks SHALL be prevented through proper validation
- Git branch names SHALL be validated against git naming rules
- Project names SHALL be validated against repository naming rules

### File System Security
- Proper file permissions SHALL be used for directory operations
- Sensitive configuration files SHALL have restricted permissions
- Temporary files SHALL be created in secure locations
- File operations SHALL handle permission errors gracefully

### Information Security
- Sensitive information SHALL NOT be logged
- Error messages SHALL NOT expose sensitive paths or tokens
- Configuration values SHALL NOT be displayed in output
- Network operations SHALL use secure protocols when required

### Error Security
- Errors SHALL be handled gracefully without panicking
- Error messages SHALL provide actionable information without exposing system details
- Error codes SHALL follow POSIX standards
- Error handling SHALL not leak sensitive information

## Context System Implementation

### Core Components
- **ContextDetector**: WILL detect current user context (project, worktree, outside git)
- **ContextResolver**: WILL resolve target identifiers based on current context
- **Context Types**: WILL include `ContextProject`, `ContextWorktree`, `ContextOutsideGit`, `ContextUnknown`

### Implementation Files
- `internal/domain/context.go` - WILL contain core context system implementation
- `internal/domain/context_test.go` - WILL contain comprehensive test suite
- `cmd/cd.go` - WILL contain context-aware cd command implementation

## Code Organization Requirements

### Package Structure
- Business logic SHALL be placed in internal/domain/
- Infrastructure implementations SHALL be placed in internal/infrastructure/
- CLI commands SHALL be placed in cmd/
- Test files SHALL be colocated with implementation files
- Shared utilities SHALL be placed in appropriate internal packages

### Interface Design
- Interfaces SHALL be defined in internal/infrastructure/interfaces.go
- Interfaces SHALL be small and focused on single responsibilities
- Implementation details SHALL NOT be exposed through interfaces
- Interface methods SHALL follow Go naming conventions

### Error Handling
- Custom error types SHALL be defined for domain-specific errors
- Errors SHALL be wrapped with context using fmt.Errorf
- Error messages SHALL be consistent and actionable
- Error handling SHALL follow Go best practices

## Configuration Management

### Configuration Implementation
- **Location**: XDG Base Directory specification SHALL be followed for config folders (`$HOME/.config/twiggit/config.toml`)
- **Format**: TOML format SHALL be supported exclusively
- **Purpose**: Default settings SHALL be overridden
- **Priority**: defaults → config file → environment variables → command flags

### Configuration Options
```toml
# File: $HOME/.config/twiggit/config.toml
# Configuration file for twiggit - overrides default settings

# Directory paths
projects_dir = "/custom/path/to/projects"
worktrees_dir = "/custom/path/to/worktrees"

# Default behavior
default_source_branch = "main"
```

### Settings
- **Directory paths**: Defaults for projects and worktrees directories SHALL be overridden
- **Default source branch**: Default `main` branch for create command SHALL be overridden

## Command Structure

The command structure SHALL be organized as follows:

```go
cmd/
├── create.go          // Create command implementation
├── delete.go          // Delete command implementation  
├── list.go            // List command implementation
├── cd.go              // CD (change directory) command implementation
├── setup-shell.go     // Shell setup command implementation
└── root.go            // Root command and CLI setup
```

## CLI Implementation

### CLI Features
- Commands SHALL follow Cobra's command/flag pattern
- Help generation SHALL use Cobra's built-in system
- Shell completion SHALL integrate with Cobra's completion framework
- --version flag SHALL display version information
- Semantic versioning SHALL be followed
- Shell completion support MAY be provided for bash, zsh, and fish
- Default POSIX output behavior SHALL be used

## Error Handling Standards

### Error Categories
- Configuration errors SHALL be clearly identified and explained
- Git operation errors SHALL include specific git error context
- File system errors SHALL include path and permission details
- Validation errors SHALL indicate what failed and how to fix
- Network errors SHALL include connection and timeout information

### Error Recovery
- Non-zero exit codes SHALL be used for all error conditions
- Specific, actionable error messages SHALL be provided
- Interactive error recovery SHALL NOT be implemented
- All inputs SHALL be validated before execution
- POSIX-compliant exit codes SHALL be used (0=success, 1=general error, 2=misuse)

### Error Presentation
- Error messages SHALL be written to stderr
- Error messages SHALL be consistent in format and style
- Error messages SHALL include suggested actions when possible
- Error codes SHALL be documented and consistent
- Error handling SHALL not expose sensitive system information

### Specific Error Messages
- **Invalid git repository**: `error: not a git repository (or any parent up to mount point /)`
- **Network errors**: `error: unable to access 'https://github.com/user/repo.git': Failed to connect to github.com port 443`
- **Permission issues**: `error: permission denied: '/home/user/Worktrees/project/branch'`
- **Ambiguous contexts**: `error: ambiguous context - please specify project and worktree explicitly`
- **Missing directories**: `error: projects directory does not exist: /home/user/Projects`

## Dependency Management

### Dependency Selection
- Dependencies SHALL NOT be added without explicit approval
- Exact versions specified in go.mod SHALL be used
- Standard library SHALL be preferred over third-party packages
- Dependencies SHALL be regularly updated for security patches
- go.mod SHALL be kept tidy and versioned appropriately

### Version Management
- Go modules SHALL be used for dependency management
- Semantic versioning SHALL be followed for dependencies
- Breaking changes SHALL be carefully evaluated before adoption
- Dependency updates SHALL be tested thoroughly

### Security Updates
- Dependencies SHALL be regularly scanned for vulnerabilities
- Security patches SHALL be applied promptly
- Deprecated dependencies SHALL be replaced with alternatives
- Dependency licenses SHALL be compatible with project requirements

## Build and Deployment Requirements

### Build Process
- Standard Go build tools SHALL be used for compilation
- Cross-compilation SHALL be supported for multiple platforms
- Build artifacts SHALL be versioned and reproducible
- Build scripts SHALL be idempotent and reliable

### Development Environment
- Mise SHALL be used for task running and environment management
- Development tasks SHALL be defined in mise configuration
- Environment variables SHALL be managed through mise
- Development tools SHALL be versioned and consistent

### Continuous Integration
- All tests SHALL pass in CI environment
- Code quality checks SHALL be enforced in CI
- Security scans SHALL be run in CI
- Build artifacts SHALL be generated and tested in CI

## Summary

Implementation requirements SHALL ensure high-quality, secure, and maintainable code. Testing SHALL prioritize end-user functionality while maintaining comprehensive coverage. Performance SHALL be optimized for CLI responsiveness. Security SHALL be enforced through proper validation and error handling. Code organization SHALL follow Go best practices with clear separation of concerns.