## Context

**Current State**: Interfaces are defined in three locations:

| Interface | Current Location | Issue |
|-----------|-----------------|-------|
| `ContextService`, `WorktreeService`, `ProjectService`, `NavigationService`, `ShellService` | `application/interfaces.go` | ✅ Correct |
| `ConfigManager` | `domain/config.go` | ❌ Domain should not define infrastructure contracts |
| `ContextDetector`, `ContextResolver` | `domain/context.go` | ❌ Domain should not define detection/resolution contracts |
| `GitClient`, `GoGitClient`, `CLIClient` | `infrastructure/interfaces.go` | ❌ Infrastructure should not define its own contracts |
| `HookRunner`, `ShellInfrastructure` | `infrastructure/interfaces.go` | ❌ Same issue |

**Architecture Rule Being Violated**: The target architecture states that `application/` shall contain all interface contracts. Currently `domain/` and `infrastructure/` each define their own interfaces, which violates the principle that interfaces represent contracts between layers, not implementations within layers.

**Constraints**:
- Service layer (`service/`) MUST be able to depend on interfaces without importing infrastructure
- Infrastructure implementations MUST implement interfaces defined in application layer
- Domain layer MUST have zero external dependencies
- No changes to public API or user-facing behavior

## Goals / Non-Goals

**Goals**:
1. Centralize all interface definitions in `internal/application/interfaces.go`
2. Enable `service/` to depend only on `application/` for interface contracts
3. Add compile-time interface checks in infrastructure implementations
4. Maintain backward compatibility for `cmd/` layer (services still injected via constructors)

**Non-Goals**:
1. Changing any domain types, entities, or validation logic
2. Modifying infrastructure implementations (only their import paths change)
3. Adding new functionality or changing behavior
4. Moving domain value objects (e.g., `Context`, `ResolutionResult`) - they belong in domain
5. Changing how services are wired in `main.go` (constructor signatures stay the same)

## Decisions

### Decision 1: Consolidate All Infrastructure-Facing Interfaces in `application/`

**Choice**: Move `ConfigManager`, `ContextDetector`, `ContextResolver`, `GitClient`, `GoGitClient`, `CLIClient`, `HookRunner`, and `ShellInfrastructure` to `application/interfaces.go`.

**Rationale**: These interfaces represent contracts between the application/service layer and infrastructure implementations. The service layer should depend on these contracts without knowing about infrastructure package existence.

**Alternatives Considered**:
- **Keep interfaces where implemented**: Rejected - this causes service layer to depend on infrastructure
- **Create separate interface packages per domain**: Over-engineering - single `application/interfaces.go` provides sufficient organization
- **Keep ConfigManager in domain**: Rejected - config loading is infrastructure concern; domain should only hold the `Config` struct (value object)

### Decision 2: Remove Interfaces from `domain/` Altogether

**Choice**: Move `ConfigManager` from `domain/config.go` and `ContextDetector`/`ContextResolver` from `domain/context.go` to `application/interfaces.go`. Keep the domain types (`Config`, `Context`, `ResolutionResult`, etc.) in domain.

**Rationale**: Domain types represent business concepts. `ContextDetector` is an implementation detail for detecting git context - it has no business meaning. `ConfigManager` loads configuration from files - pure infrastructure.

**Alternatives Considered**:
- **Move to infrastructure instead**: Rejected - service layer should not depend on infrastructure interfaces directly
- **Create new `internal/interfaces/` package**: Rejected - `application/` already serves this purpose

### Decision 3: Add Compile-Time Interface Checks

**Choice**: Add `var _ Interface = (*Implementation)(nil)` in all infrastructure implementation files.

**Rationale**: Catches missing interface implementations at build time rather than runtime. Documents interface compliance explicitly.

**Format**:
```go
// In infrastructure/gogit_client.go
var _ GoGitClient = (*GoGitClientImpl)(nil)

type GoGitClientImpl struct { ... }
```

### Decision 4: Maintain Service Constructor Signatures

**Choice**: Service constructors accept interfaces as parameters but import from `application/` not `infrastructure/`.

**Rationale**: `main.go` wiring changes from:
```go
gitClient := infrastructure.NewCompositeGitClient(...)
service.NewWorktreeService(gitClient, ...)
```
to:
```go
gitClient := infrastructure.NewCompositeGitClient(...)  // Still returns infrastructure.GitClient
service.NewWorktreeService(gitClient, ...)  // But now accepts application.GitClient
```

**Alternatives Considered**:
- **Create adapter wrappers**: Over-engineering - Go's structural typing allows this direct assignment since interfaces are compatible

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Breaking change for external consumers** | If anyone imports `twiggit/internal/infrastructure` for interface types, their code will break | Document breaking change; migration path is import path update only |
| **Large merge conflict potential** | All interface definitions in one file | Coordinate with other active changes; do this refactor before other major work |
| **Test mocks need updates** | `test/mocks/` files reference infrastructure interfaces | Update import paths in all mock files as part of this change |
| **Compile-time check boilerplate** | Adds 7 `var _` lines | Low cost; provides valuable documentation and early error detection |

### Trade-off: Single Large File vs Multiple Files

**Choice**: Keep all interfaces in single `application/interfaces.go`.

**Rationale**: 
- Current `interfaces.go` already has 100 lines with 5 interfaces
- Adding 7 more interfaces brings total to ~12 interfaces
- Single file makes it easy to see all contracts at once
- Alternative (separate files per interface) adds complexity without benefit

## Migration Plan

### Phase 1: Create consolidated interface file

1. Add all interfaces to `internal/application/interfaces.go`
2. Do NOT remove from old locations yet

### Phase 2: Update service layer

3. Update `internal/service/*.go` to import from `application/` instead of `infrastructure/` or `domain/`
4. Verify `mise run check` passes

### Phase 3: Update infrastructure layer

5. Add `var _ Interface = (*Implementation)(nil)` compile-time checks to infrastructure implementations
6. Verify `mise run check` passes
7. Remove interface definitions from `internal/infrastructure/interfaces.go` (keep structs and constructors)

### Phase 4: Update main.go and cmd layer

8. Update `main.go` import paths
9. Update `cmd/root.go` if it references moved interfaces
10. Verify `mise run check` passes

### Phase 5: Clean up domain layer

11. Remove interface definitions from `domain/config.go` and `domain/context.go`
12. Verify `mise run check` passes

### Phase 6: Update test mocks

13. Update `test/mocks/*.go` import paths
14. Run `mise run test` to verify

### Rollback

If issues arise:
1. **Trigger**: Any compilation failure in service, cmd, or test packages
2. **Procedure**:
   - Run `git stash` to save current changes
   - Verify build passes: `go build ./...`
   - If build passes, restore with `git stash pop`
3. **Revert**: `git revert HEAD` if simpler than fixing import issues

## Open Questions

1. **Should `domain/` AGENTS.md be updated to explicitly state "no interfaces defined here"?** Yes - document the rule.

2. **Closed**: Depguard enforcement is out of scope for this change.

3. **Do we need to update any exported constants in domain that reference interfaces?** Currently domain only has value objects and types; no references to the moved interfaces should exist in domain.
