# Agent Guidelines for twiggit

## Build/Test Commands
- `mise run dev:run` - Test CLI structure
- `mise run test:all` - Run all tests with coverage
- `mise run test:single TestName ./pkg/module` - Run single test
- `mise run test:coverage` - Show test coverage report (requires existing coverage.out)
- `go test -tags=integration ./test/integration/...` - Run integration tests
- `mise run lint:check` - Run linting and formatting checks
- `mise run lint:fix` - Auto-fix linting issues
- `mise run lint:testify` - Run testifylint specifically for test code analysis
- `mise run lint:testify:fix` - Auto-fix testifylint issues in test code
- `mise run build:cli` - Build the CLI binary
- `mise run build:clean` - Clean build artifacts
- `mise run dev:tidy` - Clean up go.mod and go.sum

## Development Principles

### Test-Driven Development (TDD)
- Write tests BEFORE implementation code (Red-Green-Refactor cycle)
- All new features must start with failing tests
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

- Unit tests: good coverage with mocked dependencies, test behavior not implementation
- Integration tests: Real git repositories in temporary directories
- Use testify for assertions and mock generation
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