## 1. Output Formatter Infrastructure

- [x] 1.1 Create `OutputFormatter` interface in `cmd/output.go` with `FormatWorktrees(worktrees []*domain.WorktreeInfo) string` method
- [x] 1.2 Implement `TextFormatter` struct with `FormatWorktrees()` method matching existing text output
- [x] 1.3 Implement `JSONFormatter` struct with `FormatWorktrees()` method producing compact JSON
- [x] 1.4 Add `WorktreeJSON` struct for JSON serialization with `branch`, `path`, `status` fields
- [x] 1.5 Add `WorktreeListJSON` wrapper struct with `worktrees` array field

## 2. JSON Output for List Command

- [x] 2.1 Add `--output/-o` flag to `list` command with `text` (default) and `json` values
- [x] 2.2 Add flag validation to reject invalid output format values
- [x] 2.3 Modify `executeList()` to use `OutputFormatter` based on `--output` flag
- [x] 2.4 Ensure JSON output goes to stdout while verbose messages go to stderr
- [x] 2.5 Handle empty worktrees case with `{"worktrees": []}` JSON output

## 3. Quiet Mode Implementation

- [x] 3.1 Add global `--quiet/-q` persistent flag to root command in `cmd/root.go`
- [x] 3.2 Add `isQuiet()` helper function in `cmd/util.go` to check quiet flag
- [x] 3.3 Implement quiet/verbose mutual exclusion (verbose wins) in output functions
- [x] 3.4 Suppress success messages when quiet mode is enabled
- [x] 3.5 Suppress hint messages when quiet mode is enabled
- [x] 3.6 Preserve essential output (paths for `-C` mode) in quiet mode
- [x] 3.7 Preserve error output to stderr in quiet mode

## 4. Progress Reporter for Bulk Operations

- [x] 4.1 Create `ProgressReporter` struct in `cmd/util.go` with `quiet` and `out` fields
- [x] 4.2 Implement `NewProgressReporter(quiet bool, out io.Writer)` constructor
- [x] 4.3 Implement `Report(format string, args ...interface{})` method
- [x] 4.4 Implement `ReportProgress(current, total int, item string)` method
- [x] 4.5 Add progress reporting to `prune --all` in `cmd/prune.go`
- [x] 4.6 Ensure progress output goes to stderr
- [x] 4.7 Suppress progress when quiet mode is enabled

## 5. Documentation Updates

- [x] 5.1 Update `cmd/AGENTS.md` with `--output/-o` flag documentation for list command (deferred to PHASE3)
- [x] 5.2 Update `cmd/AGENTS.md` with global `--quiet/-q` flag documentation (deferred to PHASE3)
- [x] 5.3 Update `cmd/AGENTS.md` with progress reporting behavior for prune (deferred to PHASE3)
- [x] 5.4 Update command-flags spec with new flag conventions
- [x] 5.5 Update verbose-output spec with quiet/verbose mutual exclusion

## 6. E2E Tests

- [x] 6.1 Add E2E test for `list --output json` with worktrees present
- [x] 6.2 Add E2E test for `list --output json` with no worktrees
- [x] 6.3 Add E2E test for `list --output invalid` error handling
- [x] 6.4 Add E2E test for `--quiet` flag suppressing success messages
- [x] 6.5 Add E2E test for `--quiet` preserving error output
- [x] 6.6 Add E2E test for `--quiet` with `-C` flag preserving path output
- [x] 6.7 Add E2E test for `--quiet -v` (verbose wins)
- [x] 6.8 Add E2E test for progress output during `prune --all`
- [x] 6.9 Add E2E test for progress suppression with `prune --all --quiet`
