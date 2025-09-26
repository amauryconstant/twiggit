# Twiggit Technology Stack

## Core Technology Stack

### Go 1.21+
**Rationale**: Fast compilation, single binary deployment, excellent cross-platform support
**WILL Enable**: No runtime dependencies, strong concurrency model, excellent standard library
**Constrains**: Static typing requires more verbose code, compilation step required

### Cobra (github.com/spf13/cobra)
**Rationale**: De facto standard for Go CLI applications
**WILL Enable**: Built-in help generation, flag parsing, subcommand structure, shell completion
**Constrains**: Specific command organization patterns, Cobra-specific idioms

### go-git (github.com/go-git/go-git/v5)
**Rationale**: Pure Go implementation, no system git dependency, extensible architecture
**WILL Enable**: Cross-platform consistency, portable git operations, custom worktree implementation
**Constrains**: Worktree functionality not natively supported, requires custom implementation

**go-git References**:
- **[Compatibility Matrix](https://github.com/go-git/go-git/blob/main/COMPATIBILITY.md)** - Complete feature support status and limitations
- **[Extension Guide](https://github.com/go-git/go-git/blob/main/EXTENDING.md)** - How to extend go-git functionality for custom implementations

### Koanf (github.com/knadh/koanf/v2)
**Rationale**: Modern configuration library with TOML support
**WILL Enable**: TOML parsing, environment variable overrides, simple API
**Constrains**: Specific configuration loading patterns, Koanf-specific data structures

### Testify (github.com/stretchr/testify)
**Rationale**: Enhanced testing framework for Go
**WILL Enable**: Mock generation, assertion helpers, test suite organization
**Constrains**: Testify-specific patterns, additional dependency

### Ginkgo/Gomega (github.com/onsi/ginkgo/v2, github.com/onsi/gomega)
**Rationale**: BDD-style testing for complex workflows
**WILL Enable**: Readable test specifications, rich matchers, parallel execution
**Constrains**: Ginkgo-specific test structure, learning curve

### Gexec + Gomega
**Rationale**: Process management for CLI testing
**WILL Enable**: Real binary execution, output capture, process control
**Constrains**: Additional testing complexity, process management overhead

### Carapace (github.com/carapace-sh/carapace)
**Rationale**: Advanced shell completion system
**WILL Enable**: Multi-shell support, dynamic completions, rich actions
**Constrains**: Carapace-specific completion patterns, shell detection requirements

### Standard Go os/filepath
**Rationale**: Built-in cross-platform path handling
**WILL Enable**: Platform-agnostic path operations, safety
**Constrains**: Standard library limitations, no advanced path features

### Standard Go os package
**Rationale**: Direct file system operations
**WILL Enable**: Directory creation, removal, permission handling
**Constrains**: Platform-specific behavior, manual error handling

### Standard Go fmt package
**Rationale**: Simple, reliable text formatting
**WILL Enable**: String formatting, tabular output
**Constrains**: Basic formatting only, no advanced templating

### Mise (github.com/jdx/mise)
**Rationale**: Modern task runner with environment management
**WILL Enable**: Task definitions, environment variables, version management
**Constrains**: Mise-specific configuration, additional tool dependency

### golangci-lint
**Rationale**: Comprehensive Go linting and formatting
**WILL Enable**: Code quality, consistency, best practices
**Constrains**: Configuration complexity, potential false positives

## Technology Integration Patterns

### Git Operations Integration
go-git storage and filesystem abstractions SHALL be used for custom worktree implementation. Shell commands SHALL NOT be used for git operations.

### Configuration Loading Pattern
Koanf SHALL load configuration in priority order: defaults → config file → environment variables → command flags.

### Shell Integration Strategy
Carapace SHALL generate completion scripts. Directory navigation SHALL use shell wrapper functions that intercept command output.

## Technology-Specific Constraints

### go-git Limitations
- Worktree functionality SHALL be implemented using storage abstractions
- Native worktree commands SHALL NOT be available
- Custom implementation SHALL maintain git repository integrity

### Cobra Requirements
- Commands SHALL follow Cobra's command/flag pattern
- Help generation SHALL use Cobra's built-in system
- Shell completion SHALL integrate with Cobra's completion framework

### Configuration Constraints
- TOML SHALL be the only supported configuration format
- Configuration SHALL follow XDG Base Directory specification
- Environment variables SHALL use TWIGGIT_ prefix

### Testing Framework Constraints
- Testify SHALL be used for unit testing with mocks
- Ginkgo/Gomega SHALL be used for integration testing
- Gexec SHALL be used for E2E testing with real binary execution

### Shell Integration Constraints
- Carapace SHALL be used for shell completion generation
- Shell wrapper functions SHALL be generated for directory navigation
- Shell detection SHALL be performed automatically during setup

## Decision Framework

### Technology Selection Criteria
1. **Minimal Dependencies**: Standard library SHOULD be preferred where possible
2. **Cross-Platform**: Technologies MUST work on Linux, macOS, Windows
3. **Performance**: Fast startup and execution SHALL be prioritized for CLI tool
4. **Maintainability**: Active projects with clear documentation SHOULD be selected
5. **Community**: Strong Go ecosystem support SHOULD be required

### Future-Proofing Considerations
- **Modularity**: Components SHALL be easy to replace if needed
- **Standards**: Go community best practices SHOULD be followed
- **Extensibility**: Design SHALL accommodate future feature additions
- **Compatibility**: Backward compatibility SHOULD be maintained where possible

### Optional Features
- **Alternative Configuration Formats**: Additional configuration formats MAY be supported in future releases
- **Plugin System**: A plugin system MAY be implemented to extend functionality
- **Remote Repository Management**: Remote repository operations MAY be added beyond local worktree management
- **Advanced Shell Integration**: Additional shell-specific features MAY be implemented for enhanced user experience

## Summary

A modern Go technology stack focused on simplicity, portability, and extensibility. Key choices include go-git for git operations (with custom worktree implementation), Cobra for CLI framework, Koanf for configuration, and Carapace for shell completion. All technology decisions align with pragmatic development principles and maintainability.