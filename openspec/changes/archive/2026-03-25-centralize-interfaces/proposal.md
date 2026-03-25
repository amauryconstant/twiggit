## Why

Interfaces are currently scattered across three packages (`domain/`, `application/`, `infrastructure/`) instead of being centralized in `application/`. This violates the target architecture's interface location convention and creates ambiguity about where contracts should be defined. The `service` layer depends directly on `infrastructure` package interfaces, blurring the architectural boundary between application services and infrastructure implementations.

## What Changes

1. **(Section 1)** Add all 8 interfaces to `internal/application/interfaces.go`
2. **(Section 2)** Remove interface definitions from `domain/config.go` and `domain/context.go`
3. **(Section 3)** Update service layer import paths from `infrastructure/` to `application/`
4. **(Section 4)** Add compile-time interface checks to infrastructure implementations; remove interface definitions from `infrastructure/interfaces.go`
5. **(Section 5)** Update `main.go` and `cmd/root.go` import paths
6. **(Section 6)** Update test mock import paths in `test/mocks/`
7. **(Section 7)** Update AGENTS.md documentation files

**BREAKING**: Import paths for infrastructure interfaces change from `twiggit/internal/infrastructure` to `twiggit/internal/application` for service-layer dependencies.

## Capabilities

### New Capabilities
None - this is a refactoring change with no new user-facing functionality.

### Modified Capabilities
None - this refactoring does not change any spec-level behavior or requirements. It only reorganizes where interfaces are defined.

## Impact

- **Service Layer**: `internal/service/*.go` - All 5 services (`worktree_service.go`, `project_service.go`, `context_service.go`, `navigation_service.go`, `shell_service.go`) will update import paths for `GitClient`, `HookRunner`, and `ShellInfrastructure`
- **Infrastructure Layer**: `internal/infrastructure/*.go` - Implementations remain in place but must be imported via new paths; add compile-time interface checks
- **Application Layer**: `internal/application/interfaces.go` - Becomes the single location for all service contracts
- **Domain Layer**: `internal/domain/config.go`, `internal/domain/context.go` - Remove interface definitions
- **Main Entry Point**: `main.go` - Update import paths for moved interfaces
- **Test Mocks**: `test/mocks/*.go` - Update import paths for interface types
- **AGENTS.md Files**: Update interface location documentation in `internal/application/AGENTS.md`, `internal/domain/AGENTS.md`, and `internal/infrastructure/AGENTS.md`
