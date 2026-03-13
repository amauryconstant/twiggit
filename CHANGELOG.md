# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Shell plugins for zsh, bash, and fish integration in `contrib/` directory
- Post-create hooks for automatic worktree setup (e.g., `mise trust`, `npm install`)
- Prune command for cleaning up merged worktrees
- Carapace-based shell completion with dynamic branch/project suggestions
- Configurable completion timeout
- Verbose output flags (`-v`, `--verbose`) across CLI commands
- Shell auto-detection in `init` command
- Config file path support and wrapper block management in shell init
- `list --all` flag to list worktrees across all projects
- Path traversal protection for security
- Pre-commit hooks with worktree automation
- Interactive installation script with shell completion setup
- OpenSpec workflow for structured development process

### Changed

- Refactored `init` command to output shell wrapper to stdout by default
- Renamed `setup-shell` to `init` command
- Standardized CLI flags across `create`, `delete`, and `init` commands
- Removed emoji formatting in error messages for better readability
- Consolidated error handling with typed errors and user-friendly messages
- Migrated configuration format from YAML to TOML
- Simplified architecture following KISS and YAGNI principles
- Replaced CLI-based git operations with native go-git library

### Fixed

- Completion command to properly delegate to Carapace
- Duplicate brackets in shell wrapper template
- GitLab CI pipeline to prevent duplicate runs
- golangci-lint configuration for E2E build tags
- Error handling for wrapped errors

### Security

- Path traversal protection against URL encoding and symlinks
- Security warning for post-create hooks in untrusted repositories

## [0.1.0] - Initial Release

### Added

- Core worktree management commands: `create`, `delete`, `list`, `cd`, `prune`
- Context-aware operation (auto-detects current project/worktree)
- Shell integration for directory navigation (`twiggit cd`)
- Comprehensive test coverage (unit, integration, E2E)
- Docker support and GitLab CI/CD pipeline
- GitHub mirroring capability

[Unreleased]: https://gitlab.com/amoconst/twiggit/-/compare/v0.1.0...main
[0.1.0]: https://gitlab.com/amoconst/twiggit/-/tags/v0.1.0
