# Interface Centralization - Specification Reference

## Overview

This change is a **pure refactoring** with no new capabilities and no modifications to existing capability requirements. All interface definitions are consolidated into `internal/application/interfaces.go` without changing any behavior.

## No Capability Changes

This change does not introduce new capabilities or modify existing ones. It only reorganizes where interface contracts are defined within the codebase.

### No New Capabilities

This change does not add new user-facing or system-facing capabilities.

### No Modified Capabilities

This change does not modify requirements for any existing capabilities. No delta specs are required.

## Interface Map

The following interfaces are moved from their current locations to `internal/application/interfaces.go`:

| Interface | From | To | Purpose |
|-----------|------|----|---------|
| `ConfigManager` | `domain/config.go` | `application/interfaces.go` | Configuration loading contract |
| `ContextDetector` | `domain/context.go` | `application/interfaces.go` | Context detection contract |
| `ContextResolver` | `domain/context.go` | `application/interfaces.go` | Context resolution contract |
| `GitClient` | `infrastructure/interfaces.go` | `application/interfaces.go` | Unified git operations contract |
| `GoGitClient` | `infrastructure/interfaces.go` | `application/interfaces.go` | go-git operations contract |
| `CLIClient` | `infrastructure/interfaces.go` | `application/interfaces.go` | CLI git operations contract |
| `HookRunner` | `infrastructure/interfaces.go` | `application/interfaces.go` | Hook execution contract |
| `ShellInfrastructure` | `infrastructure/interfaces.go` | `application/interfaces.go` | Shell integration contract |

## Verification

Since no capability behavior changes, verification focuses on:

1. **Compilation**: All packages compile without import errors
2. **Tests Pass**: `mise run test` completes successfully
3. **Linting Passes**: `mise run check` completes successfully
4. **Interface Compliance**: Compile-time checks (`var _ Interface = (*impl)(nil)`) confirm implementations
