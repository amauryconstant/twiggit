## 1. Domain Layer Cleanup

- [ ] 1.1 Remove `DryRun` field from `SetupShellRequest` in `internal/domain/shell_requests.go`
- [ ] 1.2 Remove `DryRun` field from `SetupShellResult` in `internal/domain/shell_results.go`

## 2. Service Layer Updates

- [ ] 2.1 Remove `DryRun` branch from `SetupShell()` in `internal/service/shell_service.go`
- [ ] 2.2 Update `SetupShell()` to no longer check/handle `req.DryRun`

## 3. Command Layer Rewrite

- [ ] 3.1 Update `NewInitCmd()` to use `[shell]` positional argument instead of `[config-file]`
- [ ] 3.2 Add `-i, --install` flag for file installation mode
- [ ] 3.3 Add `-c, --config` flag for custom config file path
- [ ] 3.4 Remove `--dry-run` flag
- [ ] 3.5 Remove `--shell` flag
- [ ] 3.6 Remove `--check` flag and `runInitCheck()` function
- [ ] 3.7 Implement flag validation: `--config` requires `--install`
- [ ] 3.8 Implement flag validation: `--force` requires `--install`
- [ ] 3.9 Implement stdout mode: call `GenerateWrapper()` and print to stdout (default behavior)
- [ ] 3.10 Implement install mode: call `SetupShell()` when `--install` flag present
- [ ] 3.11 Update `displayInitResults()` to only handle install mode (stdout mode has no metadata)
- [ ] 3.12 Update command `Use` string to `init [shell]`
- [ ] 3.13 Update command `Short` and `Long` descriptions for new behavior
- [ ] 3.14 Update `Args` validation to accept 0-1 positional args

## 4. Update Unit Tests

- [ ] 4.1 Update `cmd/init_test.go` for new flag structure
- [ ] 4.2 Remove tests for `--dry-run`, `--shell`, `--check` flags
- [ ] 4.3 Add tests for `-i, --install` flag
- [ ] 4.4 Add tests for `-c, --config` flag validation
- [ ] 4.5 Add tests for `--force` requiring `--install`
- [ ] 4.6 Add tests for stdout output mode (default)
- [ ] 4.7 Add tests for positional `[shell]` argument
- [ ] 4.8 Update `internal/service/shell_service_test.go` to remove DryRun test cases

## 5. Update E2E Tests

- [ ] 5.1 Update `test/e2e/init_test.go` for new command behavior
- [ ] 5.2 Add E2E test: `twiggit init` outputs to stdout
- [ ] 5.3 Add E2E test: `twiggit init bash` outputs bash wrapper to stdout
- [ ] 5.4 Add E2E test: `twiggit init --install` writes to auto-detected config
- [ ] 5.5 Add E2E test: `twiggit init bash --install --config <path>` writes to custom config
- [ ] 5.6 Add E2E test: `twiggit init --config <path>` errors without `--install`
- [ ] 5.7 Add E2E test: `twiggit init --force` errors without `--install`
- [ ] 5.8 Remove E2E tests for removed flags (`--dry-run`, `--shell`, `--check`)

## 6. Update Shell Completion

- [ ] 6.1 Update shell completion if needed for new positional argument

## 7. Verification

- [ ] 7.1 Run `mise run check` to ensure all linting and tests pass
- [ ] 7.2 Verify `eval "$(twiggit init bash)"` works correctly
- [ ] 7.3 Verify `twiggit init --install` writes to config file
- [ ] 7.4 Verify error messages guide users when flags are misused
