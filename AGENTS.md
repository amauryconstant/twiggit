# Agent Guidelines for twiggit

## Build/Test Commands
- `go run main.go --help` - Test CLI structure
- `go test -race -coverprofile=coverage.out ./...` - Run all tests with coverage
- `go test -run TestFunctionName ./pkg/module` - Run single test
- `go test -tags=integration ./test/integration/...` - Run integration tests
- `golangci-lint run` - Run linting and formatting checks

## Development Principles

### Test-Driven Development (TDD)
- Write tests BEFORE implementation code (Red-Green-Refactor cycle)
- All new features must start with failing tests
- Maintain high test coverage (>95%) as a safety net for refactoring
- Use tests as living documentation for expected behavior
- Refactor confidently with test coverage as validation

### Domain-Driven Design (DDD) Inspiration
- Structure code around business domains and core concepts
- Define clear domain models and bounded contexts
- Use ubiquitous language that matches business terminology
- Separate domain logic from infrastructure concerns
- Design aggregates and entities that reflect real-world concepts

### Software Design Principles
- **DRY (Don't Repeat Yourself)**: Extract common functionality into reusable components
- **YAGNI (You Aren't Gonna Need It)**: Only implement what's currently needed, avoid over-engineering
- **KISS (Keep It Simple, Stupid)**: Choose the simplest solution that works, avoid unnecessary complexity

### Functional Programming Ideas
- Prefer pure functions without side effects where possible
- Use immutability for data structures to prevent unexpected mutations
- Leverage higher-order functions for common operations (map, filter, reduce patterns)
- Design composable functions that can be easily combined
- Use function composition to build complex behavior from simple parts

## Code Style Guidelines

### Go Conventions
- Use Go 1.21+ idioms and standard library patterns
- Follow standard Go project structure: `cmd/`, `internal/`, `pkg/`, `test/`
- Error handling: Always check errors, use `fmt.Errorf` with context, avoid panics

### Imports & Dependencies
- Group imports: standard library, third-party, local packages
- Use required dependencies: Cobra, Bubble Tea, Lip Gloss, Viper, Testify
- Keep go.mod tidy and versioned appropriately

### Naming Conventions
- Functions: `camelCase` for private, `PascalCase` for exported
- Variables: `camelCase`, descriptive names
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
- **TDD Workflow**: Write failing test first, then implement minimal code to pass
- Unit tests: >95% coverage with mocked dependencies, test behavior not implementation
- Integration tests: Real git repositories in temporary directories
- Use testify for assertions and mock generation
- Table-driven tests for multiple scenarios
- Tests should be fast, isolated, and repeatable
- Use test doubles (mocks, stubs) to isolate units under test

### UI/UX Patterns
- Use Bubble Tea for rich terminal interfaces
- Lip Gloss for consistent styling and colors
- Implement proper keyboard navigation and accessibility
- Show progress indicators for long-running operations

### Performance
- Use goroutines for concurrent git operations where safe
- Implement efficient worktree discovery algorithms
- Target: <100ms discovery for 100 worktrees, <50ms per worktree status check