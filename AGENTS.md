# Agent Guidelines for twiggit

## Essential Knowledge (Every Prompt)

### Project Context

- **Purpose**: Pragmatic git worktree management tool with focus on rebase workflows
- **Architecture**: Go CLI with domain-driven design, clear separation of concerns
- **Core Commands**: `list`, `create`, `delete`, `cd`, `setup-shell`
- **Key Technologies**: Cobra, go-git, Koanf, Testify, Ginkgo/Gomega, Carapace

### Critical Specification Keywords

| Keyword        | Meaning                  | When to Use                                    |
| -------------- | ------------------------ | ---------------------------------------------- |
| **SHALL**      | Mandatory requirement    | Critical functionality, security constraints   |
| **SHALL NOT**  | Absolute prohibition     | Security constraints, architectural boundaries |
| **SHOULD**     | Recommended practice     | Best practices, optimization recommendations   |
| **SHOULD NOT** | Discouraged practice     | Anti-patterns, performance pitfalls            |
| **WILL**       | System fact or guarantee | System behavior declarations                   |
| **WILL NOT**   | System absence guarantee | System non-occurrence declarations             |
| **MAY**        | Optional feature         | Extensibility points, future enhancements      |
| **MAY NOT**    | Optional restriction     | Configurable constraints                       |

### Essential Commands

- `mise run test` - Run all tests
- `mise run lint:fix` - Run linting and formatting checks with automatic fixes
- `mise run dev:run` - Run the application in development mode
- `mise tasks` - List all available tasks

### Core Development Principles

- **Working code over perfect architecture**: Get features working, then refine
- **User value over technical metrics**: Focus on solving real problems for developers
- **Write tests BEFORE implementation**: Use tests as safety net for refactoring
- **KISS, YAGNI, DRY**: Keep it simple, build what's needed, avoid repetition

## When to Consult Additional Documentation

### For Feature Implementation

**Consult**: [`.ai/design.md`](.ai/design.md) (commands section) + [`.ai/implementation.md`](.ai/implementation.md) (testing section)
**When**: Implementing new commands, modifying existing functionality, adding CLI features
**Focus**: Command specifications, behavior requirements, testing frameworks, quality standards

### For Technology Decisions

**Consult**: [`.ai/technology.md`](.ai/technology.md)
**When**: Choosing libraries, architectural decisions, integration patterns, dependency management
**Focus**: Technology stack rationale, constraints, integration patterns, decision framework

### For Testing Requirements

**Consult**: [`.ai/testing.md`](.ai/testing.md)
**When**: Writing tests, ensuring test coverage, setting up test frameworks, test data management
**Focus**: Testing philosophy, patterns, framework usage, coverage requirements, quality standards

### For Documentation Design Guidance

**Consult**: [`.ai/documentation-design.md`](.ai/documentation-design.md)
**When**: Understanding documentation architecture, keyword usage, file responsibilities, maintenance procedures
**Focus**: Documentation system design, keyword definitions, quality assurance, update procedures

### For Code Style Details

**Consult**: [`.ai/code-style-guide.md`](.ai/code-style-guide.md)
**When**: Writing Go code, naming conventions, error handling
**Focus**: Concrete examples, patterns, and anti-patterns


