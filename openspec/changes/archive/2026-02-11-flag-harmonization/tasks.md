## 1. Command Layer Changes

- [x] 1.1 Remove zombie `--cd` string flag from cmd/create.go
- [x] 1.2 Add `-C, --cd` boolean flag to cmd/create.go with BoolVarP
- [x] 1.3 Implement path output logic in cmd/create.go executeCreate function
- [x] 1.4 Update cmd/create.go Long description to document `--source` and `-C, --cd` flags
- [x] 1.5 Rename `--change-dir` to `--cd` in cmd/delete.go flag definition
- [x] 1.6 Add `-f` short form for `--force` flag in cmd/delete.go using BoolVarP
- [x] 1.7 Implement getDeleteNavigationTarget helper function in cmd/delete.go
- [x] 1.8 Update cmd/delete.go deleteWorktree function to use navigation logic
- [x] 1.9 Update cmd/delete.go output format (path only when `-C` set)
- [x] 1.10 Update cmd/delete.go Long description to document all flags (`-f, --force`, `--merged-only`, `-C, --cd`)
- [x] 1.11 Add `-f` short form for `--force` in cmd/init.go using BoolVarP
- [x] 1.12 Reorder cmd/init.go flag definitions alphabetically
- [x] 1.13 Update cmd/init.go Long description to list all flags alphabetically

## 2. Shell Wrapper Updates

- [x] 2.1 Update bash wrapper template in internal/infrastructure/shell/service.go
- [x] 2.2 Update zsh wrapper template in internal/infrastructure/shell/service.go
- [x] 2.3 Update fish wrapper template in internal/infrastructure/shell/service.go
- [x] 2.4 Implement Option B pattern (case statement for create/delete commands)

## 3. Documentation Updates

- [x] 3.1 Update cmd/AGENTS.md list command specification
- [x] 3.2 Update cmd/AGENTS.md create command specification with `-C, --cd`
- [x] 3.3 Update cmd/AGENTS.md delete command specification with `-f, --force`, `--merged-only`, `-C, --cd`, and navigation behavior
- [x] 3.4 Update cmd/AGENTS.md init command specification with `-f, --force`, `--dry-run`, `--check`, `--shell`, and alphabetical ordering

## 4. Unit Tests (cmd/create_test.go)

- [x] 4.1 Add test case for create with `-C` flag outputs path only
- [x] 4.2 Add test case for create without `-C` flag outputs success message
- [x] 4.3 Update existing tests to reference `-C, --cd` instead of zombie `--cd` flag

## 5. Unit Tests (cmd/delete_test.go)

- [x] 5.1 Add test case for delete with `-C` flag from worktree context outputs project path
- [x] 5.2 Add test case for delete with `-C` flag from project context outputs nothing
- [x] 5.3 Add test case for delete with `-C` flag from outside git context outputs nothing
- [x] 5.4 Add test case for delete with `-f` short form flag works correctly
- [x] 5.5 Update existing tests to reference `--cd` instead of `--change-dir`

## 6. Unit Tests (cmd/init_test.go)

- [x] 6.1 Add test case for init with `-f` short form flag

## 7. Integration Tests

- [x] 7.1 Add integration test in test/integration/cli_commands_test.go for wrapper with create `-C` flag
- [x] 7.2 Add integration test in test/integration/cli_commands_test.go for wrapper with delete `-C` flag

## 8. E2E Tests (test/e2e/create_test.go)

- [x] 8.1 Add E2E test for create command with `-C` flag outputs path to stdout
- [x] 8.2 Add E2E test for create command without `-C` flag outputs success message

## 9. E2E Tests (test/e2e/delete_test.go)

- [x] 9.1 Add E2E test for delete command with `-C` flag from worktree context
- [x] 9.2 Add E2E test for delete command with `-C` flag from project context
- [x] 9.3 Add E2E test for delete command with `-C` flag from outside git context

## 10. Verification

- [x] 10.1 Run unit tests: `mise run test`
- [x] 10.2 Run integration tests: `mise run test:integration`
- [x] 10.3 Run E2E tests: `mise run test:e2e`
- [x] 10.4 Run lint: `mise run lint:fix`
- [x] 10.5 Run full check: `mise run check`

## 11. Specification Updates (Discovered During Implementation)

- [x] 11.1 Add flag registration pattern requirement to openspec/specs/command-flags/spec.md
- [x] 11.2 Update design decision #6 in openspec/changes/flag-harmonization/design.md to clarify *Var/*VarP vs Get*() pattern
- [x] 11.3 Update task 1.12 in openspec/changes/flag-harmonization/tasks.md to reflect completed flag pattern alignment
