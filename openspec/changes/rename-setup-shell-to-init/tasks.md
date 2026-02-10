## 1. Domain Layer Changes

- [ ] 1.1 Add inference error constant to shell_errors.go
- [ ] 1.2 Add InferShellTypeFromPath function to shell.go with pattern matching
- [ ] 1.3 Update wrapper templates in shell.go to include block delimiters
- [ ] 1.4 Create unit tests for InferShellTypeFromPath

## 2. Infrastructure Layer Changes

- [ ] 2.1 Update InstallWrapper signature to accept explicit config file and force flag
- [ ] 2.2 Implement hasWrapperBlock helper for checking delimiters
- [ ] 2.3 Implement removeWrapperBlock helper for removing old wrappers
- [ ] 2.4 Implement appendWrapper helper for appending to config files
- [ ] 2.5 Update InstallWrapper to use block delimiter logic
- [ ] 2.6 Update ValidateInstallation to use block delimiters
- [ ] 2.7 Create unit tests for hasWrapperBlock
- [ ] 2.8 Create unit tests for removeWrapperBlock
- [ ] 2.9 Update integration tests for InstallWrapper with force flag

## 3. Service Layer Changes

- [ ] 3.1 Update SetupShell to infer shell type when not specified
- [ ] 3.2 Add isValidShellType helper to shell_service.go
- [ ] 3.3 Update SetupShell to handle inference errors gracefully
- [ ] 3.4 Update SetupShell to use explicit config file from request
- [ ] 3.5 Update integration tests for inference scenarios

## 4. Command Layer Changes

- [ ] 4.1 Create cmd/init.go with NewInitCommand function
- [ ] 4.2 Implement runInit function with positional argument and flags
- [ ] 4.3 Implement displayInitResults function for output
- [ ] 4.4 Create cmd/init_test.go with unit tests
- [ ] 4.5 Update cmd/root.go to use NewInitCommand
- [ ] 4.6 Delete cmd/setup-shell.go
- [ ] 4.7 Delete cmd/setup-shell_test.go

## 5. Install Script Updates

- [ ] 5.1 Add detect_config_file helper function to install.sh
- [ ] 5.2 Update shell wrapper installation section to orchestrate init command
- [ ] 5.3 Add confirmation prompts for existing wrappers
- [ ] 5.4 Add confirmation prompts for new installations
- [ ] 5.5 Update error handling to show warnings (not fail)

## 6. Test Updates - E2E

- [ ] 6.1 Delete test/e2e/setup_shell_test.go
- [ ] 6.2 Create test/e2e/init_test.go with inference scenarios
- [ ] 6.3 Add test for install to existing config file
- [ ] 6.4 Add test for install to missing config file
- [ ] 6.5 Add test for shell type inference from filename
- [ ] 6.6 Add test for force reinstall with block replacement
- [ ] 6.7 Add test for skip when wrapper exists
- [ ] 6.8 Add test for dry-run output
- [ ] 6.9 Add test for inference failure with custom path
- [ ] 6.10 Add test for explicit --shell override

## 7. Test Updates - Integration

- [ ] 7.1 Update test/integration/cli_commands_test.go expected commands
- [ ] 7.2 Add integration tests for shell type inference
- [ ] 7.3 Add integration tests for force reinstall scenarios

## 8. Validation

- [ ] 8.1 Run mise run lint:fix to check code quality
- [ ] 8.2 Run mise run test:unit to verify unit tests pass
- [ ] 8.3 Run mise run test:integration to verify integration tests pass
- [ ] 8.4 Run mise run test:e2e to verify E2E tests pass
- [ ] 8.5 Run mise run test to run full test suite
- [ ] 8.6 Run mise run build to verify binary builds
- [ ] 8.7 Run mise run check to run all validation
