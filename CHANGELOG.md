# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.11.0] - 2026-03-30


### Changed

- Centralized all interface definitions in application package (consolidated from domain/ and infrastructure/)
- Service layer now depends on interfaces from application/ following dependency inversion principle

### Added

- compile-time interface satisfaction checks to all implementations
- doc.go package documentation files
- govulncheck to pre-commit for vulnerability scanning
- depguard linter with domain package isolation rules

## [0.10.0] - 2026-03-18

### Added

- Golden file testing infrastructure for snapshot testing of list and error output
- Explicit error formatter strategy pattern with `errors.As()` matching
- Automatic cleanup registration for `RepoTestHelper` using `t.Cleanup()`
- Environment variable expansion in config paths (`$HOME` and `~` support)

### Changed

- Converted all layer tests (service, infrastructure, domain) from testify/suite to standard Go testing with `t.Run()` and `t.Cleanup()`
- Error formatter refactored to explicit strategy pattern with ordered matcher-formatter slice
- Build configuration updated to use `internal/version` package
- Go bumped to 1.26.1

### Fixed

- Golden file comparison normalization with TrimSpace on both actual and expected output
- Test helpers properly respect `HOME` env var for test isolation

## [0.9.2] - 2026-03-16

### Added

- Environment variable expansion in config paths (`ProjectsDirectory`, `WorktreesDirectory`, `BackupDir`)
- Support for `~` (tilde) expansion in config paths

## [0.9.1] - 2026-03-16

### Changed

- Go upgraded to 1.26.1

### Fixed

- E2E tests to respect HOME and use explicit config paths
- Path validation to ensure config paths stay under home directory

## [0.9.0] - 2026-03-16

### Added

- Command aliases: `ls` for `list`, `rm` for `delete`
- Short flags: `-a` for `list --all`, `-y/--yes` for `prune` auto-confirmation
- JSON output via `--output/-o` flag on `list` command
- Quiet mode with global `--quiet/-q` flag
- Progress reporting for `prune` command
- Shell completion enhancements: fuzzy matching, smart sorting, status indicators
- Granular exit codes: ExitCodeConfig (3), ExitCodeGit (4), ExitCodeValidation (5), ExitCodeNotFound (6)
- Panic recovery in main.go with user-friendly error messages
- Help text improvements with examples sections

### Changed

- Error messages simplified - removed internal operation names
- Mise tasks refactored to use explicit `run` arrays

### Fixed

- E2E test failures from output-scripting change
- Test isolation issues with HOME env var
- Cross-project branch completion

## [0.8.1] - 2026-03-12

### Changed

- `init` command default behavior changed from file installation to stdout output
  - Default mode: Print shell wrapper to stdout (eval-safe)
  - `-i, --install` flag enables file installation mode
  - `-c, --config` flag for custom config file path

## [0.8.0] - 2026-03-12

### Added

- Shell plugins for zsh (Oh My Zsh, antidote, zinit, znap with lazy-load completions)
- Shell plugins for bash (standalone and Bash-It integration)
- Shell plugins for fish (conf.d and Oh My Fish integration)
- Post-create hook execution for worktree setup
- HookRunner infrastructure to execute commands from `.twiggit.toml`
- Graceful interrupt handling for OpenSpec autonomous workflow

### Changed

- Pre-commit hooks refactored to use direct tool invocation
- GitLab CI pipeline optimized to prevent duplicate runs
- Validation jobs now run only on MRs and tags

### Fixed

- Completion command delegation to Carapace
- Duplicate brackets in shell wrapper template
- golangci-lint configuration for E2E build tags

## [0.7.0] - 2026-02-17

### Added

- Carapace integration for shell completion via `_carapace` command
- Configurable completion timeout (500ms default)
- LRU cache (25 repos) for git repositories
- TTL-based cache for git worktree validation
- ProjectSummary type for lightweight project enumeration
- `ListProjectSummaries` method on ProjectService

### Changed

- Dependencies updated: go-git v5.16.5, koanf v2.3.2, ginkgo v2.28.1, gomega v1.39.1
- Removed visible completion command (use `_carapace` instead)
- Version package consolidated into cmd layer

### Fixed

- Branch description tagging loop index bug
- Parent directory comparison edge case in IsPathUnder
- Error detection using typed ConfigError instead of string matching

## [0.6.0] - 2026-02-13

### Added

- Prune command for merged worktree cleanup with `--dry-run`, `--force`, `--delete-branches`, `--all` flags
- Protected branch protection (main, master, develop, staging, production)
- Cross-project pruning with confirmation prompt
- Navigation path output for shell integration
- OpenSpec extended skills for artifact management

## [0.5.6] - 2026-02-12

### Added

- Comprehensive codebase quality audit skill with modular patterns
- Depth levels: Minimal/Standard/Detailed/Maximum report granularity

### Changed

- Codebase quality refactored for modularity and flexibility

## [0.5.5] - 2026-02-12

### Added

- Standardized mock pattern with testify/mock (`.On()`/`.Return()`)
- Error handler tests for CLI error handling
- Git client tests for CompositeGitClient routing

### Changed

- Error handling standardized across domain/infrastructure/service layers with ValidationError types
- Test patterns unified to use Testify suites
- CLI flags standardized (`-C` for `--cd`, `-f` for `--force`)
- Infrastructure package flattened
- GoReleaser speed optimizations

### Removed

- Unused caching layer in ContextDetector
- Unused interfaces (GitRepositoryInterface, ProjectRepository)

## [0.5.4] - 2026-02-11

### Changed

- GoReleaser: Disabled Go module proxy

## [0.5.3] - 2026-02-11

### Added

- HTML coverage reports (`coverage.html`)
- XML coverage reports in Cobertura format (`coverage.xml`)

### Changed

- Gitignore updated for coverage file patterns

## [0.5.2] - 2026-02-11

### Added

- Shell auto-detection in `init` command from SHELL environment variable
- `init` command renamed from `setup-shell`
- `--check` flag to validate wrapper installation
- Verbose output with `-v` and `-vv` flags
- Interactive shell completions for bash, zsh, fish

### Changed

- `setup-shell` renamed to `init` command

## [0.5.1] - 2026-02-09

### Added

- Location-specific AGENTS.md documentation (cmd/, internal/, test/)
- OpenSpec workflow with specifications and change tracking
- Pre-commit hooks with worktree automation

### Changed

- Removed centralized `.ai/` directory in favor of co-located AGENTS.md files

## [0.4.0] - 2026-02-09

### Added

- Shell integration system with `setup-shell` command (bash, zsh, fish support)
- Context-aware navigation system with project/worktree detection
- `default_source_branch` configuration option
- `--source` flag for `create` command
- Unified error handling with structured suggestions

### Changed

- Configuration format migrated from YAML to TOML
- CLI command `switch` renamed to `cd`
- Shell integration enhanced with zfunctions support for zsh

### Fixed

- Shell wrapper functions properly preserve exit codes

### Security

- Path traversal protection against directory escape attacks

## [0.3.0] - 2026-02-09

### Changed

- Domain layer refactored to pure business logic
- Architecture properly separates domain and infrastructure layers
- Filesystem abstraction with dependency injection
- Infrastructure abstraction layer with mockable interfaces

## [0.2.0] - 2026-02-09

### Added

- Dependency injection container in `internal/infrastructure/deps.go`
- CLI installation task

### Changed

- All CLI commands refactored to accept dependencies
- Architecture refactored to use dependency injection pattern
- Native git command execution moved to service layer

## [0.1.13] - 2026-02-09

### Changed

- Docker support and modernized CI/CD pipeline
- Docker Buildx for multi-platform builds

### Fixed

- GitLab CI pipeline to prevent duplicate runs

## [0.1.7] - 2026-01-21

### Added

- Delete command for worktree removal
- Improved create command UX with automatic branch creation

## [0.1.3] - 2026-01-19

### Added

- Unified status and list commands
- Bare repository filtering to prevent discovery failures

## [0.1.0] - 2025-09-18 - Initial Release

### Added

- Core worktree management commands: `create`, `delete`, `list`, `cd`, `prune`
- Context-aware operation (auto-detects current project/worktree)
- Shell integration for directory navigation (`twiggit cd`)
- Comprehensive test coverage (unit, integration, E2E)
- Docker support and GitLab CI/CD pipeline
- GitHub mirroring capability

[0.10.1]: https://gitlab.com/amoconst/twiggit/-/compare/v0.10.0...v0.10.1
[0.10.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.9.2...v0.10.0
[0.9.2]: https://gitlab.com/amoconst/twiggit/-/compare/v0.9.1...v0.9.2
[0.9.1]: https://gitlab.com/amoconst/twiggit/-/compare/v0.9.0...v0.9.1
[0.9.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.8.1...v0.9.0
[0.8.1]: https://gitlab.com/amoconst/twiggit/-/compare/v0.8.0...v0.8.1
[0.8.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.7.0...v0.8.0
[0.7.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.6.0...v0.7.0
[0.6.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.6...v0.6.0
[0.5.6]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.5...v0.5.6
[0.5.5]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.4...v0.5.5
[0.5.4]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.3...v0.5.4
[0.5.3]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.2...v0.5.3
[0.5.2]: https://gitlab.com/amoconst/twiggit/-/compare/v0.5.1...v0.5.2
[0.5.1]: https://gitlab.com/amoconst/twiggit/-/compare/v0.4.0...v0.5.1
[0.4.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.3.0...v0.4.0
[0.3.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.2.0...v0.3.0
[0.2.0]: https://gitlab.com/amoconst/twiggit/-/compare/v0.1.13...v0.2.0
[0.1.13]: https://gitlab.com/amoconst/twiggit/-/compare/v0.1.7...v0.1.13
[0.1.7]: https://gitlab.com/amoconst/twiggit/-/compare/v0.1.3...v0.1.7
[0.1.3]: https://gitlab.com/amoconst/twiggit/-/compare/v0.1.0...v0.1.3
[0.1.0]: https://gitlab.com/amoconst/twiggit/-/tags/v0.1.0
