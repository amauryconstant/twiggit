# Agent Guidelines for twiggit

## Build/Test Commands

- `mise run test:ci` - Run CI test suite (unit + integration + race)
- `mise run test:coverage` - Show coverage report in JSON format
- `mise run test:coverage:core` - Show coverage for core business logic (domain, services)
- `mise run test:coverage:infrastructure` - Show coverage for infrastructure layer (config, git)
- `mise run lint:check` - Run linting and formatting checks
- `mise run lint:fix` - Auto-fix linting issues
- `mise run build:cli` - Build the CLI binary

## Development Principles

### Pragmatic Development First

- **Working code over perfect architecture**: Get features working, then refine
- **User value over technical metrics**: Focus on solving real problems for developers
- **Incremental delivery**: Each commit should add visible user value
- **Simple solutions for simple problems**: Avoid over-engineering for personal tools

### Test-Driven Development (TDD) - Applied Pragmatically

- Write tests BEFORE implementation for critical business logic
- Use tests as safety net for refactoring, not as bureaucracy
- Focus on integration tests that verify user workflows
- Unit tests for complex algorithms, not simple getters/setters

### Domain-Driven Design (DDD) - Simplified

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
- Use essential dependencies: Cobra for CLI, go-git for Git operations, testify for testing
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

- Unit tests: good coverage with mocked dependencies, test behavior not implementation
- Integration tests: Real git repositories in temporary directories
- Use testify for assertions and mock generation
- Tests should be fast, isolated, and repeatable
- Use test doubles (mocks, stubs) to isolate units under test
