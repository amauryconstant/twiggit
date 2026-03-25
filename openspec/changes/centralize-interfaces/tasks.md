## 1. Consolidate Interfaces in application/

- [x] 1.1 Add ConfigManager interface to internal/application/interfaces.go
- [x] 1.2 Add ContextDetector interface to internal/application/interfaces.go
- [x] 1.3 Add ContextResolver interface to internal/application/interfaces.go
- [x] 1.4 Add GitClient interface to internal/application/interfaces.go
- [x] 1.5 Add GoGitClient interface to internal/application/interfaces.go
- [x] 1.6 Add CLIClient interface to internal/application/interfaces.go
- [x] 1.7 Add HookRunner interface to internal/application/interfaces.go
- [x] 1.8 Add ShellInfrastructure interface to internal/application/interfaces.go

## 2. Update Service Layer

**IMPORTANT**: Service layer MUST be updated before removing interfaces from domain (Section 5) to maintain compilability.

- [x] 2.1 Update internal/service/worktree_service.go import paths from `infrastructure/` to `application/`
- [x] 2.2 Update internal/service/project_service.go import paths from `infrastructure/` to `application/`
- [x] 2.3 Update internal/service/context_service.go import paths from `domain/` and `infrastructure/` to `application/`
- [x] 2.4 Update internal/service/navigation_service.go import paths from `infrastructure/` to `application/`
- [x] 2.5 Update internal/service/shell_service.go import paths from `infrastructure/` to `application/`

## 3. Update Infrastructure Layer

- [x] 3.1 Add compile-time check `var _ GoGitClient = (*GoGitClientImpl)(nil)` to internal/infrastructure/gogit_client.go
- [x] 3.2 Add compile-time check `var _ CLIClient = (*CLIClientImpl)(nil)` to internal/infrastructure/cli_client.go
- [x] 3.3 Add compile-time check `var _ HookRunner = (*HookRunnerImpl)(nil)` to internal/infrastructure/hook_runner.go
- [x] 3.4 Add compile-time check `var _ ShellInfrastructure = (*ShellInfrastructureImpl)(nil)` to internal/infrastructure/shell_infra.go
- [x] 3.5 Remove interface definitions from internal/infrastructure/interfaces.go (keep implementation structs and constructor functions)

## 4. Update Main and Cmd Layer

- [x] 4.1 Update main.go import paths for moved interfaces
- [x] 4.2 Update cmd/root.go ServiceContainer if needed
- [x] 4.3 Verify cmd layer compiles correctly

## 5. Update Domain Layer

**NOTE**: This section is safe to execute AFTER Section 2 (service layer) because services now import from `application/` instead of `domain/`.

- [x] 5.1 Remove ConfigManager interface from internal/domain/config.go
- [x] 5.2 Remove ContextDetector interface from internal/domain/context.go
- [x] 5.3 Remove ContextResolver interface from internal/domain/context.go
- [x] 5.4 Verify domain package compiles with no infrastructure imports

## 6. Update Test Mocks

- [x] 6.1 Update test/mocks/git_service_mock.go import paths for GoGitClient, CLIClient, GitClient types
- [x] 6.2 Update test/mocks/shell_infrastructure_mock.go import paths for ShellInfrastructure type
- [x] 6.3 Verify test/mocks/ compiles correctly

## 7. Update Documentation

- [ ] 7.1 Update internal/application/AGENTS.md interface documentation
- [ ] 7.2 Update internal/domain/AGENTS.md to clarify no interfaces defined there
- [ ] 7.3 Update internal/infrastructure/AGENTS.md to remove interface definitions

## 8. Verification

- [x] 8.1 Run `mise run lint:fix && mise run check`
- [x] 8.2 Run `mise run test`
- [x] 8.3 Run `mise run test:e2e`
- [x] 8.4 Verify `go build ./...` compiles successfully
