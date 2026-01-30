# Agent Guidelines for twiggit

## 🚀 Quick Start

### Project Overview

- **Purpose**: Pragmatic git worktree management tool with focus on rebase workflows
- **Architecture**: Go CLI with domain-driven design, clear separation of concerns
- **Core Commands**: `list`, `create`, `delete`, `cd`, `setup-shell`
- **Key Technologies**: Cobra, go-git, Koanf, Testify, Ginkgo/Gomega, Carapace

### Essential Commands

```bash
# Development
mise run test          # Run all tests (quick)
mise run test:full      # Run all tests (unit, integration, e2e, race)
mise run test:coverage # Run tests with coverage
mise run lint:fix      # Run linting and formatting checks
mise run format        # Format Go code with gofmt
mise run check         # Run all validation checks
mise run dev:run       # Run in development mode
mise run build         # Build CLI binary to bin/twiggit (with version injection)
mise run build:local   # Build and install to $HOME/.local/bin

# Release
mise run release:validate  # Validate release prerequisites
mise run release:tag patch  # Create and push patch version tag
mise run release:dry-run    # Test GoReleaser without publishing

# Tool Management
mise run tools:check   # Check for tool updates
mise run tools:update   # Update tool versions in .mise/config.toml

mise tasks             # List all available tasks
```

---

## 🐳 Docker + CI Infrastructure

### Overview

- **CI Platform**: GitLab CI with multi-stage Docker images
- **Image**: `registry.gitlab.com/amoconst/twiggit/ci:latest`
- **Base**: `golang:1.25.5-alpine3.23` with pre-installed mise tools
- **Cache Strategy**: Content-addressed image caching
- **Release Automation**: GoReleaser with GitLab integration

### Pipeline Stages

1. **build-ci**: Builds Docker image when `Dockerfile.ci` or `.mise/config.toml` changes
2. **setup**: Downloads and caches Go modules
3. **validate**: Runs lint, tests, and GoReleaser dry-run
4. **distribute**: Creates GitLab releases on tags, mirrors to GitHub

### Image Caching

Docker images are tagged with `{mise-version}-{SHA256(Dockerfile.ci+config)}`, ensuring:
- Efficient cache utilization
- Automatic rebuilds on config changes
- No manual version management needed

---

## 📦 Release Workflow

### Process

1. **Validate**: Run `mise run release:validate` to check:
   - Working directory is clean
   - On main branch
   - GoReleaser configuration is valid

2. **Tag**: Run `mise run release:tag <patch|minor|major>` to:
   - Calculate next version
   - Create annotated tag (e.g., `v0.5.0`)
   - Push tag to remote
   - Trigger CI release job

3. **Release**: GoReleaser automatically:
   - Builds multi-platform binaries (linux/darwin/windows, amd64/arm64)
   - Creates GitLab release with assets
   - Generates checksums and SBOMs
   - Generates changelog from commit messages
   - Publishes all artifacts

### Multi-Platform Support

| OS      | Architecture | Binary Name                |
| ------- | ------------ | --------------------------- |
| Linux   | amd64        | `twiggit_v0.5.0_linux_amd64.tar.gz`   |
| Linux   | arm64        | `twiggit_v0.5.0_linux_arm64.tar.gz`   |
| macOS   | amd64        | `twiggit_v0.5.0_darwin_amd64.tar.gz`  |
| macOS   | arm64        | `twiggit_v0.5.0_darwin_arm64.tar.gz`  |
| Windows | amd64        | `twiggit_v0.5.0_windows_amd64.zip`   |
| Windows | arm64        | `twiggit_v0.5.0_windows_arm64.zip`   |

### Installation

Users can install using the provided script:
```bash
curl -fsSL https://gitlab.com/amoconst/twiggit/-/raw/main/install.sh | bash
```

Or download manually from: https://gitlab.com/amoconst/twiggit/-/releases

---

---

## 🎯 Core Development Principles

- **Test-driven Development**: Write tests in the RED phase, Implement the bare minimum to pass the tests in the GREEN phase, Refactor the code to improve its quality
- **Working code over perfect architecture**: Get features working, then refine
- **User value over technical metrics**: Focus on solving real problems for developers
- **KISS, YAGNI, DRY**: Keep it simple, build what's needed, avoid repetition

---

## 🎨 Code Quality & Tooling

### Pre-Commit Hooks

 twiggit uses pre-commit hooks to enforce code quality standards:

**Enabled Hooks:**
- Trailing whitespace removal
- End-of-file fixing
- YAML/TOML/JSON validation
- Merge conflict detection
- Mixed line ending normalization

**Setup:**
```bash
# Pre-commit is automatically installed via mise
mise install
pre-commit install
```

**Manual Execution:**
```bash
pre-commit run --all-files
```

**Skip Hooks:**
```bash
git commit -m "message" --no-verify
```

---

---

## 🔤 Critical Specification Keywords

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

---

## 📚 Documentation Reference Guide

### 🛠️ Feature Implementation

**Consult**: [`.ai/design.md`](.ai/design.md) (commands section) + [`.ai/implementation.md`](.ai/implementation.md) (testing section)  
**When**: Implementing new commands, modifying existing functionality, adding CLI features  
**Focus**: Command specifications, behavior requirements, testing frameworks, quality standards

### 💻 Technology Decisions

**Consult**: [`.ai/technology.md`](.ai/technology.md)  
**When**: Choosing libraries, architectural decisions, integration patterns, dependency management  
**Focus**: Technology stack rationale, constraints, integration patterns, decision framework

### 🧪 Testing Requirements

**Consult**: [`.ai/testing.md`](.ai/testing.md)  
**When**: Writing tests, ensuring test coverage, setting up test frameworks, test data management  
**Focus**: Testing philosophy, patterns, framework usage, coverage requirements, quality standards

### 📖 Documentation Design

**Consult**: [`.ai/documentation-design.md`](.ai/documentation-design.md)  
**When**: Understanding documentation architecture, keyword usage, file responsibilities, maintenance procedures  
**Focus**: Documentation system design, keyword definitions, quality assurance, update procedures

### ✍️ Code Style

**Consult**: [`.ai/code-style-guide.md`](.ai/code-style-guide.md)  
**When**: Writing Go code, naming conventions, error handling  
**Focus**: Concrete examples, patterns, and anti-patterns
