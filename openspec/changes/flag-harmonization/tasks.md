## 1. Command Layer Changes

- [ ] 1.1 Remove zombie `--cd` string flag from cmd/create.go
- [ ] 1.2 Add `-C, --cd` boolean flag to cmd/create.go with BoolVarP
- [ ] 1.3 Implement path output logic in cmd/create.go executeCreate function
- [ ] 1.4 Update cmd/create.go Long description to document `--source` and `-C, --cd` flags
- [ ] 1.5 Rename `--change-dir` to `--cd` in cmd/delete.go flag definition
- [ ] 1.6 Add `-f` short form for `--force` flag in cmd/delete.go using BoolVarP
- [ ] 1.7 Implement getDeleteNavigationTarget helper function in cmd/delete.go
- [ ] 1.8 Update cmd/delete.go deleteWorktree function to use navigation logic
- [ ] 1.9 Update cmd/delete.go output format (path only when `-C` set)
- [ ] 1.10 Update cmd/delete.go Long description to document all flags (`-f, --force`, `--merged-only`, `-C, --cd`)
- [ ] 1.11 Add `-f` short form for `--force` in cmd/init.go using BoolVarP
- [ ] 1.12 Reorder cmd/init.go flag definitions alphabetically
- [ ] 1.13 Update cmd/init.go Long description to list all flags alphabetically

## 2. Shell Wrapper Updates

- [ ] 2.1 Update bash wrapper template in internal/infrastructure/shell/service.go
- [ ] 2.2 Update zsh wrapper template in internal/infrastructure/shell/service.go
- [ ] 2.3 Update fish wrapper template in internal/infrastructure/shell/service.go
- [ ] 2.4 Implement Option B pattern (case statement for create/delete commands)

## 3. Documentation Updates

- [ ] 3.1 Update cmd/AGENTS.md list command specification
- [ ] 3.2 Update cmd/AGENTS.md create command specification with `-C, --cd`
- [ ] 3.3 Update cmd/AGENTS.md delete command specification with `-f, --force`, `--merged-only`, `-C, --cd`, and navigation behavior
- [ ] 3.4 Update cmd/AGENTS.md init command specification with `-f, --force`, `--dry-run`, `--check`, `--shell`, and alphabetical ordering

## 4. Unit Tests (cmd/create_test.go)

- [ ] 4.1 Add test case for create with `-C` flag outputs path only
- [ ] 4.2 Add test case for create without `-C` flag outputs success message
- [ ] 4.3 Update existing tests to reference `-C, --cd` instead of zombie `--cd` flag

## 5. Unit Tests (cmd/delete_test.go)

- [ ] 5.1 Add test case for delete with `-C` flag from worktree context outputs project path
- [ ] 5.2 Add test case for delete with `-C` flag from project context outputs nothing
- [ ] 5.3 Add test case for delete with `-C` flag from outside git context outputs nothing
- [ ] 5.4 Add test case for delete with `-f` short form flag works correctly
- [ ] 5.5 Update existing tests to reference `--cd` instead of `--change-dir`

## 6. Unit Tests (cmd/init_test.go)

- [ ] 6.1 Add test case for init with `-f` short form flag

## 7. Integration Tests

- [ ] 7.1 Add integration test in test/integration/cli_commands_test.go for wrapper with create `-C` flag
- [ ] 7.2 Add integration test in test/integration/cli_commands_test.go for wrapper with delete `-C` flag

## 8. E2E Tests (test/e2e/create_test.go)

- [ ] 8.1 Add E2E test for create command with `-C` flag outputs path to stdout
- [ ] 8.2 Add E2E test for create command without `-C` flag outputs success message

## 9. E2E Tests (test/e2e/delete_test.go)

- [ ] 9.1 Add E2E test for delete command with `-C` flag from worktree context
- [ ] 9.2 Add E2E test for delete command with `-C` flag from project context
- [ ] 9.3 Add E2E test for delete command with `-C` flag from outside git context

## 10. Verification

- [ ] 10.1 Run unit tests: `mise run test`
- [ ] 10.2 Run integration tests: `mise run test:integration`
- [ ] 10.3 Run E2E tests: `mise run test:e2e`
- [ ] 10.4 Run lint: `mise run lint:fix`
- [ ] 10.5 Run full check: `mise run check`
