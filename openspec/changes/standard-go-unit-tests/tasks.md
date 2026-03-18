## 0. Prerequisites

- [x] 0. Verify test-helpers-cleanup change is archived before starting this conversion

## 5. Domain Layer Conversion

- [x] 5.1 Convert internal/domain/config_test.go to standard Go testing
- [x] 5.2 Convert internal/domain/errors_test.go to standard Go testing
- [x] 5.3 Convert internal/domain/shell_errors_test.go to standard Go testing
- [x] 5.4 Convert internal/domain/shell_test.go to standard Go testing
- [x] 5.5 Convert internal/domain/validation_test.go to standard Go testing
- [x] 5.6 Convert internal/domain/context_test.go to standard Go testing
- [x] 5.7 Convert internal/domain/service_errors_test.go to standard Go testing
- [x] 5.8 Run mise run test:full to verify domain layer conversion

## 6. Infrastructure Layer Conversion

- [x] 6.1 Convert internal/infrastructure/command_executor_test.go to standard Go testing
- [x] 6.2 Convert internal/infrastructure/gogit_client_test.go to standard Go testing
- [x] 6.3 Convert internal/infrastructure/shell_infra_test.go to standard Go testing
- [x] 6.4 Convert internal/infrastructure/context_detector_test.go to standard Go testing
- [x] 6.5 Convert internal/infrastructure/git_client_test.go to standard Go testing
- [x] 6.6 Convert internal/infrastructure/cli_client_test.go to standard Go testing
- [x] 6.7 Convert internal/infrastructure/git_utils_test.go to standard Go testing
- [x] 6.8 Convert internal/infrastructure/pathutils_test.go to standard Go testing
- [x] 6.9 Convert internal/infrastructure/hook_runner_test.go to standard Go testing
- [x] 6.10 Convert internal/infrastructure/context_resolver_test.go to standard Go testing
- [x] 6.11 Convert internal/infrastructure/config_manager_test.go to standard Go testing
- [x] 6.12 Run mise run test:full to verify infrastructure layer conversion

## 7. Service Layer Conversion

- [x] 7.1 Convert internal/service/context_service_test.go to standard Go testing
- [x] 7.2 Convert internal/service/navigation_service_test.go to standard Go testing
- [x] 7.3 Convert internal/service/project_service_test.go to standard Go testing
- [x] 7.4 Convert internal/service/shell_service_test.go to standard Go testing
- [x] 7.5 Convert internal/service/worktree_service_test.go to standard Go testing
- [x] 7.6 Run mise run test:full to verify service layer conversion

## 8. Command Layer Conversion

- [x] 8.1 Convert cmd/completion_test.go to standard Go testing
- [x] 8.2 Convert cmd/error_handler_test.go to standard Go testing
- [x] 8.3 Convert cmd/init_test.go to standard Go testing
- [x] 8.4 Convert cmd/suggestions_test.go to standard Go testing
- [x] 8.5 Run mise run test:full to verify command layer conversion

## 9. Cleanup and Verification

- [x] 9.1 Remove github.com/stretchr/testify/suite import from go.mod
- [x] 9.2 Run mise run test:full to verify all tests pass
- [x] 9.3 Run mise run test:race to verify no race conditions
- [x] 9.4 Run mise run lint to verify code quality
- [x] 9.5 Update test/AGENTS.md with standard Go testing patterns
  - Note: Deferred to PHASE3 (MAINTAIN DOCS) per workflow guidelines
