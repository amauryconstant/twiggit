# Agent Guidelines for twiggit

## Build/Test Commands

### Testing Commands
- `mise run test` - Run all tests (unit + integration + E2E + race)
- `mise run test:unit` - Run unit tests only
- `mise run test:integration` - Run integration tests only
- `mise run test:e2e` - Run CLI end-to-end tests
- `mise run test:race` - Run tests with race condition detection
- `mise run test:single` - Run single test (usage: `mise run test:single TestName ./pkg/module`)

### Build Commands
- `mise run build:cli` - Build the CLI binary
- `mise run build:e2e` - Build CLI binary for E2E tests
- `mise run build:clean` - Clean build artifacts

### Quality Commands
- `mise run lint:check` - Run linting and formatting checks
- `mise run lint:fix` - Auto-fix linting issues

### Development Commands
- `mise run dev:run` - Test CLI structure
- `mise run dev:tidy` - Clean up go.mod and go.sum

## Development Principles

### Pragmatic Development First

- **Working code over perfect architecture**: Get features working, then refine
- **User value over technical metrics**: Focus on solving real problems for developers
- **Incremental delivery**: Each commit should add visible user value
- **Simple solutions for simple problems**: Avoid over-engineering for personal tools

### Test-Driven Development (TDD)

- Write tests BEFORE implementation
- Use tests as safety net for refactoring, not as bureaucracy
- Focus on integration tests that verify user workflows
- Unit tests for complex algorithms, not simple getters/setters
- E2E tests for CLI commands to ensure user-facing functionality works correctly

### Simplified Domain-Driven Design 

- Use clear domain models that reflect real worktree concepts
- Separate business logic from infrastructure where it makes sense
- Avoid over-abstraction; use direct function calls when appropriate
- Focus on ubiquitous language that matches Git/worktree terminology

### Software Design Principles

- **KISS (Keep It Simple, Stupid)**: Always choose the simplest working solution
- **YAGNI (You Aren't Gonna Need It)**: Only build what's needed for immediate use
- **DRY (Don't Repeat Yourself)**: Extract common patterns, but avoid premature abstraction
- **Solve real problems**: Every line of code should solve an actual user pain point

## Code Style Guidelines

### Go Conventions

- Use Go 1.21+ idioms and standard library patterns
- Follow standard Go project structure: `cmd/`, `internal/`, `test/`
- Error handling: Always check errors, use `fmt.Errorf` with context, avoid panics

### Imports & Dependencies

- Group imports: standard library, third-party, local packages
- Current essential dependencies:
  - **Cobra** (`github.com/spf13/cobra`): CLI framework for command parsing and help
  - **go-git** (`github.com/go-git/go-git/v5`): Native Git operations without system git dependency
  - **Carapace** (`github.com/carapace-sh/carapace`): Shell completion integration
  - **Koanf** (`github.com/knadh/koanf/v2`): Configuration management with YAML and env support
  - **Testify** (`github.com/stretchr/testify`): Testing assertions and mock generation
  - **Ginkgo/Gomega** (`github.com/onsi/ginkgo/v2`, `github.com/onsi/gomega`): E2E testing framework for CLI interaction testing
- Keep go.mod tidy and versioned appropriately
- **Avoid adding dependencies unless they solve a real user problem**

### Naming Conventions

- Functions: `camelCase` for private, `PascalCase` for exported
- Variables: `camelCase`, descriptive names that match domain concepts
- Constants: `PascalCase` or `SCREAMING_SNAKE_CASE`
- Files: `snake_case.go`, match package name when possible

### Types & Interfaces

- Define clear interfaces for external dependencies (Git client, config, etc.)
- Use struct embedding for composition
- Include proper godoc comments for all exported types and functions

### Error Handling

- Create custom error types with `errors.New` or `fmt.Errorf`
- Wrap errors with context: `fmt.Errorf("failed to create worktree: %w", err)`
- Use error checking patterns consistently across the codebase

### Testing Patterns

#### Unit Tests
- Good coverage with mocked dependencies, test behavior not implementation
- Use testify for assertions and mock generation
- Focus on business logic and algorithms
- Tests should be fast, isolated, and repeatable

#### Integration Tests
- Real git repositories in temporary directories
- Test component interactions and workflows
- Use build tags to separate from unit tests
- Skip in short mode with `testing.Short()` check

#### E2E Tests
- Test CLI commands from user perspective using Ginkgo/Gomega
- Build actual binary and execute commands
- Use gexec for process management and output capture
- Test complete user workflows and error scenarios
- Use build tags to separate from other test types

#### General Principles
- Use test doubles (mocks, stubs) to isolate units under test
- Table-driven tests for multiple scenarios with same logic
- Descriptive test names that explain the scenario
- Always cleanup resources with `defer` patterns
