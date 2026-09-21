# Spec Delta

## ADDED Requirements

### Requirement: Hook env vars injected via cmd.Env

The hook runner SHALL pass the per-hook environment variables
(`TWIGGIT_WORKTREE_PATH`, `TWIGGIT_PROJECT_NAME`,
`TWIGGIT_BRANCH_NAME`, `TWIGGIT_SOURCE_BRANCH`,
`TWIGGIT_MAIN_REPO_PATH`) via `exec.Command.Env`, not via
string concatenation into a `sh -c "<exports>; <cmd>"`
command. The command itself SHALL be invoked directly via
`exec.Command` so paths with embedded single quotes or other
shell metacharacters do not break.

#### Scenario: Hook runs with single-quote in worktree path
- **WHEN** the worktree path contains a single quote (e.g.,
  `/home/user/Worktrees/myproj/feature's-branch/`)
- **THEN** the hook executes successfully and receives
  `TWIGGIT_WORKTREE_PATH` as the exact path (no quoting
  corruption)

#### Scenario: Existing env vars are preserved
- **WHEN** the hook runner builds the command env
- **THEN** the hook receives the parent process env with the
  five `TWIGGIT_*` additions, not a stripped-down env

### Requirement: Hook warnings surface via slog

The hook runner SHALL emit warnings via `slog.Warn` (not direct
`os.Stderr` writes) so that log aggregators capture hook
diagnostics uniformly with the rest of the project. Warnings
SHALL include the hook command, exit code (if any), and the
output truncated to 1 KiB.

#### Scenario: Hook command exits non-zero
- **WHEN** a hook command exits with a non-zero status
- **THEN** the runner emits a `slog.Warn` with the command,
  exit code, and truncated output; the user sees the same
  warning via the cmd layer's formatter
