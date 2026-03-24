## 1. Consolidate Interfaces in application/

- [ ] 1.1 Add ConfigManager interface to internal/application/interfaces.go
- [ ] 1.2 Add ContextDetector interface to internal/application/interfaces.go
- [ ] 1.3 Add ContextResolver interface to internal/application/interfaces.go
- [ ] 1.4 Add GitClient interface to internal/application/interfaces.go
- [ ] 1.5 Add GoGitClient interface to internal/application/interfaces.go
- [ ] 1.6 Add CLIClient interface to internal/application/interfaces.go
- [ ] 1.7 Add HookRunner interface to internal/application/interfaces.go
- [ ] 1.8 Add ShellInfrastructure interface to internal/application/interfaces.go

## 2. Update Domain Layer

- [ ] 2.1 Remove ConfigManager interface from internal/domain/config.go
- [ ] 2.2 Remove ContextDetector interface from internal/domain/context.go
- [ ] 2.3 Remove ContextResolver interface from internal/domain/context.go
- [ ] 2.4 Verify domain package compiles with no infrastructure imports

## 3. Update Service Layer

- [ ] 3.1 Update internal/service/worktree_service.go import paths
- [ ] 3.2 Update internal/service/project_service.go import paths
- [ ] 3.3 Update internal/service/context_service.go import paths
- [ ] 3.4 Update internal/service/navigation_service.go import paths
- [ ] 3.5 Update internal/service/shell_service.go import paths

## 4. Update Infrastructure Layer

- [ ] 4.1 Add compile-time check to internal/infrastructure/gogit_client.go
- [ ] 4.2 Add compile-time check to internal/infrastructure/cli_client.go
- [ ] 4.3 Add compile-time check to internal/infrastructure/hook_runner.go
- [ ] 4.4 Add compile-time check to internal/infrastructure/shell_infra.go
- [ ] 4.5 Remove interface definitions from internal/infrastructure/interfaces.go (keep implementation structs and constructor functions)

## 5. Update Main and Cmd Layer

- [ ] 5.1 Update main.go import paths for moved interfaces
- [ ] 5.2 Update cmd/root.go ServiceContainer if needed
- [ ] 5.3 Verify cmd layer compiles correctly

## 6. Update Test Mocks

- [ ] 6.1 Update test/mocks/ import paths for interface types
- [ ] 6.2 Verify test/mocks/ compiles correctly

## 7. Update Documentation

- [ ] 7.1 Update internal/application/AGENTS.md interface documentation
- [ ] 7.2 Update internal/domain/AGENTS.md to clarify no interfaces defined there
- [ ] 7.3 Update internal/infrastructure/AGENTS.md to remove interface definitions

## 8. Verification

- [ ] 8.1 Run `mise run lint:fix && mise run check`
- [ ] 8.2 Run `mise run test`
- [ ] 8.3 Run `mise run test:e2e`
- [ ] 8.4 Verify `go build ./...` compiles successfully
