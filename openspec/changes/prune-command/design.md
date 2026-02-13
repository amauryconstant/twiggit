## Context

Current twiggit provides `delete` command for single worktree deletion but lacks bulk cleanup capability for post-merge workflows. Users must manually identify and delete each merged worktree across multiple projects, which is time-consuming. Infrastructure already supports:
- `IsBranchMerged()` via CLI to check merge status
- `ListWorktrees()` and `DeleteWorktree()` via CLI for worktree operations
- Context detection and resolution for project/worktree/outside git states
- Navigation service for path resolution (used by `delete -C`)

Architecture constraints:
- Domain-driven design with clear layer separation
- GoGitClient for branch operations (deterministic routing guide)
- CLIClient for worktree operations (go-git lacks worktree support)
- Context-aware commands adapt behavior based on detection
- No code comments unless explicitly requested
- Constructor injection with interface contracts
- Structured error handling with clear error types for all operations

## Goals / Non-Goals

**Goals:**
- Provide context-aware `prune` command that deletes merged worktrees across projects
- Support optional branch deletion via new `DeleteBranch()` infrastructure
- Protect reasonable default branches (main, master, develop, staging, production)
- Navigate to project directory after single-worktree prune (matches `delete -C` behavior)
- Require confirmation for bulk operations (`--all`) to prevent accidental multi-project deletion
- Support dry-run mode for preview before execution

**Non-Goals:**
- Deleting unmerged branches (users can use `delete --force` for this)
- Automatic branch deletion by default (opt-in via `--delete-branches`)
- Git remote branch management (local only)
- Merging branches (prune is cleanup-only, not a merge tool)
- Modifying existing `delete` command behavior

## Decisions

### 1. Main Worktree Detection: Path Comparison

**Choice:** Detect main worktree by comparing worktree path to project GitRepoPath

**Rationale:**
- Existing `ListWorktrees()` already uses this pattern (`wt.Path != project.GitRepoPath`)
- Matches the canonical location: project directory contains `.git/` directory, worktree contains `.git` file
- Simple, deterministic, no branch name assumptions

**Alternatives considered:**
- Detect by branch name (check if `main` or `master`): Would miss custom default branches
- Check for `.git` directory vs `.git` file: Less explicit than path comparison
- Use `git worktree list` "detached" flag: Adds complexity, path comparison is sufficient

### 2. Protected Branches: Configurable List

**Choice:** Add `ProtectedBranches []string` to ValidationConfig with sensible defaults

**Rationale:**
- Users have different branch naming conventions (main vs master vs develop)
- Configurable defaults cover common patterns while allowing customization
- Aligns with existing validation configuration patterns
- Prevents accidental deletion of long-lived integration/production branches

**Alternatives considered:**
- Hardcode only `main` and `master`: Too restrictive for teams using git flow
- No protection at all: Risky for production deployments
- Check if branch is default branch of remote: Adds network dependency, overkill

### 3. Branch Deletion: GoGitClient via go-git

**Choice:** Implement `DeleteBranch()` in GoGitClient using go-git's `repo.Storer.DeleteReference()`

**Rationale:**
- Aligns with infrastructure guide: GoGitClient for branch operations, CLIClient for worktree operations
- Deterministic and portable (no CLI fallback ambiguity)
- Already using go-git for `ListBranches()`, `BranchExists()`, `GetRepositoryStatus()`
- Faster than spawning CLI process for each branch deletion
- Consistent with structured error handling patterns

**Alternatives considered:**
- Use CLI (`git branch -d <branch>`): Requires additional CLI routing logic
- Use CLI for consistency with worktree ops: Mixing responsibilities, breaks deterministic routing guide
- Don't delete branches at all: Requires manual cleanup after pruning worktrees

### 4. Confirmation Prompt: Bulk Mode Only

**Choice:** Interactive confirmation required for `--all` flag only

**Rationale:**
- Single-project prune is low-risk (limited to one project)
- Bulk mode (`--all`) affects all projects across user's workspace
- Balances safety with convenience for single-project use case
- Confirmation goes to stderr to avoid polluting stdout (which is used for path output)

**Alternatives considered:**
- Always require confirmation: Too intrusive for common single-project usage
- Never require confirmation: Risky for bulk operations
- Configurable confirmation flag: Adds complexity, bulk-only is clear boundary

### 5. Navigation Output: Context-Aware Only

**Choice:** Output navigation path to stdout only when pruning single worktree from worktree context

**Rationale:**
- Matches `delete -C` behavior for consistency
- Multiple-worktree or bulk modes have no single navigation target
- Shell wrapper consumes stdout for `cd` command
- Single-worktree context means user was working in that directory and needs to go somewhere

**Alternatives considered:**
- Always output project path: Confusing when pruning from project context
- Never output navigation: Breaks user flow when self-deleting from worktree
- Ask user for target: Adds friction, predictable behavior is better

### 6. Error Handling: Non-Blocking Branch Deletion Failures

**Choice:** Continue worktree deletion even if branch deletion fails; report as warning

**Rationale:**
- Worktree directory is primary cleanup target; branch deletion is optional optimization
- Branch ref might not exist (orphaned state) - worktree deletion should still succeed
- Allows users to manually clean up failed branch deletions
- Prevents "partial failure" rollback complexity
- Consistent with structured error handling and error type checking

**Alternatives considered:**
- Fail entire operation on branch deletion error: Overly strict, blocks worktree cleanup
- Retry branch deletion: Could loop indefinitely if branch truly doesn't exist
- Require `--force` to ignore branch errors: Adds flag complexity for edge case

### 7. Testing Strategy: Layer-Appropriate Coverage

**Choice:** Distribute tests across layers based on what's being tested

**Rationale:**
- E2E tests cover CLI-layer concerns (stdin interaction, confirmation prompts, process execution)
- Unit tests with mocks cover git behavior (worktree removal failures, branch deletion failures)
- Integration tests use Testify suites per AGENTS.md guidelines for real git operations
- Separation ensures tests are focused and maintainable

**Test distribution:**
- Bulk confirmation prompts: E2E only (requires stdin mocking with `gexec`)
- Worktree removal failures: Unit tests with mocks (testing git library behavior)
- Branch deletion failures: Unit tests with mocks (testing git library behavior)
- Merged worktree detection: Integration tests (real git operations)
- Protected branch filtering: Integration tests (real git operations)

**Alternatives considered:**
- Test everything with integration tests: Slow, harder to isolate failures
- Test everything with mocks: Doesn't catch real git behavior changes
- No layer separation: Tests become unfocused and brittle

## Risks / Trade-offs

### [Risk] User accidentally deletes protected branch via config bypass

**Mitigation:**
- Protected branches are enforced in code, not just validation
- Clear error message explains why branch is protected
- Config validation on startup prevents invalid protected branch patterns

### [Risk] Dry-run preview doesn't match actual execution state

**Mitigation:**
- Dry-run uses same logic as actual execution (no separate code path)
- Document that dry-run is best-effort preview
- User confirmation for bulk mode catches large discrepancies

### [Risk] Branch deletion conflicts with active worktrees

**Mitigation:**
- Git's `DeleteReference()` returns error if branch is checked out
- Worktree deletion happens first, so branch should be safe to delete
- Document that user may need to manually clean up orphaned refs

### [Risk] Navigation output conflicts with error messages

**Mitigation:**
- Path output to stdout, errors to stderr (consistent with `delete -C`)
- Only output navigation path when worktree deletion succeeds
- Shell wrapper only reads stdout, errors don't affect navigation

### [Trade-off] Protected branches hardcoded in default config

**Trade-off:** Default list (main, master, develop, staging, production) may not match all workflows
**Mitigation:** Configurable via `protected_branches` in config file; documentation explains customization

### [Trade-off] No automatic branch deletion

**Trade-off:** Users must opt-in to branch deletion, leaving orphaned refs after prune
**Mitigation:** Clear messaging in dry-run shows branches that would be deleted; `--delete-branches` flag is discoverable

## Open Questions

None. Design decisions resolve all open questions from implementation planning phase.
