# Contributing to Twiggit

Thank you for your interest in contributing to Twiggit! This guide will help you get started.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites

- **Go 1.25+** (see `go.mod` for exact version)
- **mise** for task automation (optional but recommended)
- **pre-commit** for git hooks

### Initial Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://gitlab.com/YOUR_USERNAME/twiggit.git
   cd twiggit
   ```

2. **Install development tools:**
   ```bash
   # Using mise (recommended)
   mise install

   # Or manually install required tools
   go mod download
   ```

3. **Set up pre-commit hooks:**
   ```bash
   pre-commit install
   ```

4. **Verify your setup:**
   ```bash
   mise run check
   ```

## Project Structure

```
twiggit/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── application/        # Interface definitions
│   ├── domain/            # Domain entities and errors
│   ├── infrastructure/    # Git client, config, shell
│   └── service/           # Business logic implementations
├── test/
│   ├── mocks/             # Generated mock implementations
│   ├── integration/       # Integration tests
│   ├── e2e/               # End-to-end CLI tests
│   └── helpers/           # Test utilities
├── contrib/               # Shell plugin files
└── openspec/              # OpenSpec change management
```

For detailed architecture and conventions, see [AGENTS.md](AGENTS.md).

## Testing

### Quick Tests

```bash
mise run test          # Unit tests only
mise run test:full     # All tests (unit, integration, e2e, race)
```

### Specific Test Types

```bash
mise run test:unit         # Unit tests only
mise run test:integration  # Integration tests
mise run test:e2e          # End-to-end CLI tests
mise run test:race         # Race condition detection
```

### Writing Tests

- **Unit tests**: Place alongside source files (`*_test.go`)
- **Integration tests**: Use testify/suite in `test/integration/`
- **E2E tests**: Use Ginkgo/Gomega in `test/e2e/`

See [test/AGENTS.md](test/AGENTS.md) for detailed testing conventions.

## Code Style

### Formatting

We use standard Go tooling:

```bash
mise run format    # Format code
mise run lint:fix  # Auto-fix linting issues
```

### Linting

```bash
mise run lint:check   # Check linting
mise run lint:fix     # Auto-fix issues
```

### Pre-commit Hooks

Pre-commit hooks run automatically on commit:

- `gofmt` - Code formatting
- `govet` - Static analysis
- `golangci-lint` - Comprehensive linting

To run all hooks manually:
```bash
pre-commit run --all-files
```

### Code Conventions

- Follow standard Go idioms and [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Write tests for new functionality

## Pull Request Process

### Before Submitting

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes and test:**
   ```bash
   mise run check    # Run all validation
   ```

3. **Commit your changes:**
   ```bash
   git add .
   git commit -m "type: brief description"
   ```

   Follow conventional commit format:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation
   - `test:` - Tests
   - `refactor:` - Code refactoring
   - `chore:` - Maintenance tasks

4. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

### Submitting the PR

1. Open a merge request on GitLab
2. Fill in the merge request template
3. Link any related issues
4. Wait for review

### Review Process

- All PRs require at least one review
- CI must pass before merging
- Address review feedback promptly
- Squash commits before merging (if requested)

## Getting Help

- **Documentation**: Check [AGENTS.md](AGENTS.md) for project-specific details
- **Issues**: Open an issue on GitLab for bugs or feature requests
- **Questions**: Start a discussion on GitLab

Thank you for contributing!
