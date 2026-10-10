package git

import (
	"context"
	"strings"
	"testing"
	"time"
	"twiggit/internal/core"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupStatusTestRepo initialises a real go-git repository at
// <tempDir>/repo with one commit on `main` and the caller-supplied
// secondary branch committed at HEAD. The returned path is absolute
// so test assertions can match it against wt paths emitted by
// ListWorktrees (which also absolutise via filepath.Abs).
func setupStatusTestRepo(t *testing.T, secondaryBranch string) string {
	t.Helper()

	tempDir := t.TempDir()
	repoPath := tempDir + "/repo"
	repo, err := git.PlainInit(repoPath, false)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	// Initial commit on the default branch.
	require.NoError(t, worktree.AddGlob("."))
	commitOpts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
		AllowEmptyCommits: true,
	}
	_, err = worktree.Commit("initial commit", commitOpts)
	require.NoError(t, err)

	headRef, err := repo.Head()
	require.NoError(t, err)

	// Rename the default branch to `main` so the tests match the
	// production layout (Init.defaultBranch is `master` on older go-git).
	mainRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), headRef.Hash())
	require.NoError(t, repo.Storer.SetReference(mainRef))
	require.NoError(t, repo.Storer.RemoveReference(headRef.Name()))
	// HEAD must point at the renamed branch; otherwise ListBranches
	// (which reads HEAD) fails to resolve a current branch.
	headMain := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/main"))
	require.NoError(t, repo.Storer.SetReference(headMain))

	if secondaryBranch != "" {
		// Create the secondary branch at HEAD.
		branchRef := plumbing.NewHashReference(
			plumbing.ReferenceName("refs/heads/"+secondaryBranch),
			headRef.Hash(),
		)
		require.NoError(t, repo.Storer.SetReference(branchRef))
	}

	return repoPath
}

// statusTestClient returns a *Client with a MockCommandExecutor
// installed. The mock has NO pre-wired expectations: testify's
// findExpectedCall returns the FIRST matching expectation in
// forward order, so pre-wiring common calls in this helper would
// shadow any per-test override. Each test wires the calls it
// expects via mockRevList / mockBranchMerged / mockTrackedBase /
// mockWorktreeList below.
func statusTestClient(t *testing.T, _, _ string) (*Client, *MockCommandExecutor) {
	t.Helper()
	executor := NewMockCommandExecutor()
	client, err := NewClient(WithExecutor(executor))
	require.NoError(t, err)
	return client, executor
}

// argMatch returns a function that returns true when the args
// slice equals the supplied expected values.
func argMatch(want ...string) func([]string) bool {
	return func(got []string) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range want {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}
}

// mockWorktreeList wires `git worktree list --porcelain` to return
// the supplied repo+branch. All other calls fall through to other
// expectations registered in the test.
func mockWorktreeList(executor *MockCommandExecutor, repoPath, branch string) {
	executor.On("ExecuteWithTimeout",
		mock.Anything, mock.Anything, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("worktree", "list", "--porcelain")),
	).Return(&CommandResult{ExitCode: 0, Stdout: porcelainForBranch(repoPath, branch)}, nil)
}

// mockTrackedBase wires `git config --worktree --get twiggit.tracked-base`
// to return the supplied base (empty string for the absent case).
func mockTrackedBase(executor *MockCommandExecutor, base string) {
	exit := 0
	stdout := base + "\n"
	if base == "" {
		exit = 1
		stdout = ""
	}
	executor.On("ExecuteWithTimeout",
		mock.Anything, mock.Anything, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("config", "--worktree", "--get", "twiggit.tracked-base")),
	).Return(&CommandResult{ExitCode: exit, Stdout: stdout}, nil)
}

// mockBranchMerged wires `git branch --merged` to return the
// supplied stdout (a `\n`-separated branch list).
func mockBranchMerged(executor *MockCommandExecutor, stdout string, exit int) {
	executor.On("ExecuteWithTimeout",
		mock.Anything, mock.Anything, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("branch", "--merged")),
	).Return(&CommandResult{ExitCode: exit, Stdout: stdout}, nil)
}

// mockRevList wires the two rev-list --count expectations for the
// (base, branch) pair against repoPath. Both directions return
// the same canned count.
func mockRevList(executor *MockCommandExecutor, repoPath, base, branch, count string) {
	executor.On("ExecuteWithTimeout",
		mock.Anything, repoPath, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("rev-list", "--count", base+".."+branch)),
	).Return(&CommandResult{ExitCode: 0, Stdout: count + "\n"}, nil)
	executor.On("ExecuteWithTimeout",
		mock.Anything, repoPath, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("rev-list", "--count", branch+".."+base)),
	).Return(&CommandResult{ExitCode: 0, Stdout: count + "\n"}, nil)
}

// porcelainForBranch returns the `git worktree list --porcelain`
// output describing a single worktree at repoPath with the supplied
// branch. Used to drive ListWorktrees through the mock executor.
func porcelainForBranch(repoPath, branch string) string {
	return "worktree " + repoPath + "\nHEAD 0000000000000000000000000000000000000000\nbranch refs/heads/" + branch + "\n"
}

func TestReadWorktreeStatus_EmptyInputsReturnError(t *testing.T) {
	t.Parallel()
	client, _ := statusTestClient(t, "/repo", "main")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		_, err := client.ReadWorktreeStatus(ctx, nil, "/repo", "/wt")
		require.Error(t, err)
		var oe *core.OperationError
		require.ErrorAs(t, err, &oe)
		assert.Equal(t, statusOp, oe.Op)
	})

	t.Run("empty repo path", func(t *testing.T) {
		t.Parallel()
		_, err := client.ReadWorktreeStatus(ctx, cfg, "", "/wt")
		require.Error(t, err)
		var oe *core.OperationError
		require.ErrorAs(t, err, &oe)
		assert.Equal(t, statusOp, oe.Op)
	})

	t.Run("empty worktree path", func(t *testing.T) {
		t.Parallel()
		_, err := client.ReadWorktreeStatus(ctx, cfg, "/repo", "")
		require.Error(t, err)
		var oe *core.OperationError
		require.ErrorAs(t, err, &oe)
		assert.Equal(t, statusOp, oe.Op)
	})
}

func TestReadWorktreeStatus_PopulatesDirtyState(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)

	require.NotNil(t, row.RepositoryStatus,
		"RepositoryStatus must be populated for a successful read")
	assert.True(t, row.IsClean,
		"freshly-init'd repo with no uncommitted changes must be clean")
	assert.False(t, row.HasUncommittedChanges)
	assert.Equal(t, "main", row.Base,
		"with no per-worktree tracked-base the protected-branch fallback must supply main")
	assert.Equal(t, "feat/x", row.Worktree.Branch,
		"Worktree.Branch must mirror the porcelain worktree list")
	assert.Equal(t, repoPath, row.Worktree.Path)
	assert.False(t, row.IsSkipped)
	assert.Empty(t, row.SkipReason)
}

func TestReadWorktreeStatus_AheadBehindParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		stdout  string
		want    int
		wantErr bool
	}{
		{"five ahead", "5\n", 5, false},
		{"zero", "0\n", 0, false},
		{"twelve behind", "12\n", 12, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repoPath := setupStatusTestRepo(t, "feat/x")
			client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
			mockWorktreeList(mockExecutor, repoPath, "feat/x")
			mockTrackedBase(mockExecutor, "")
			mockBranchMerged(mockExecutor, "", 0)
			count := strings.TrimSpace(tc.stdout)
			mockRevList(mockExecutor, repoPath, "main", "feat/x", count)
			ctx := context.Background()
			cfg := core.DefaultConfig()

			row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
			require.NoError(t, err)
			require.NotNil(t, row.RepositoryStatus)
			assert.Equal(t, tc.want, row.RepositoryStatus.Ahead)
			assert.Equal(t, tc.want, row.RepositoryStatus.Behind)
		})
	}
}

func TestReadWorktreeStatus_AheadBehindFailureMarksSkip(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	// rev-list main..feat/x fails: the populateDirtyAndAheadBehind
	// short-circuits on the first failure and never invokes the
	// second direction. Only the failing call is registered.
	mockExecutor.On("ExecuteWithTimeout",
		mock.Anything, repoPath, CmdGit, mock.AnythingOfType("time.Duration"),
		mock.MatchedBy(argMatch("rev-list", "--count", "main..feat/x")),
	).Return(&CommandResult{ExitCode: 128, Stderr: "fatal: bad ref"}, nil)
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err,
		"per-wt failure must surface as a skip row, not a returned error")

	assert.True(t, row.IsSkipped,
		"non-zero rev-list exit code must set IsSkipped")
	assert.NotEmpty(t, row.SkipReason)
	assert.Contains(t, row.SkipReason, "ahead/behind")
}

func TestReadWorktreeStatus_IsMergedSuccess(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/merged")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/merged")
	mockWorktreeList(mockExecutor, repoPath, "feat/merged")
	mockTrackedBase(mockExecutor, "")
	mockRevList(mockExecutor, repoPath, "main", "feat/merged", "0")
	// IsBranchMerged shells out via the executor; the branch appears
	// in the merged-branches output, so IsMerged = true.
	mockBranchMerged(mockExecutor, "  feat/merged\n  main\n", 0)
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.True(t, row.IsMerged,
		"worktree branch present in `git branch --merged` must yield IsMerged = true")
}

func TestReadWorktreeStatus_IsMergedFailureMarksSkip(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	mockBranchMerged(mockExecutor, "", 128)
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err,
		"IsBranchMerged failure must mark the row skipped, not return an error")
	assert.True(t, row.IsSkipped)
	assert.NotEmpty(t, row.SkipReason)
}

func TestReadWorktreeStatus_TrackedBasePresent(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	// GetTrackedBase returns "develop" — overrides the protected-branch
	// fallback to main.
	mockTrackedBase(mockExecutor, "develop")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "develop", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.Equal(t, "develop", row.Base,
		"per-worktree tracked-base must override the protected-branch fallback")
}

func TestReadWorktreeStatus_NoTrackedBaseFallsBackToProtected(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.Equal(t, "main", row.Base,
		"with no tracked-base the protected-branch fallback must supply the first entry")
	assert.False(t, row.IsSkipped)
}

func TestReadWorktreeStatus_NoTrackedBaseNoProtectedSkips(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	ctx := context.Background()
	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = nil

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err,
		"absent base must mark the row skipped, not abort the walk")
	assert.Empty(t, row.Base)
	assert.True(t, row.IsSkipped)
	assert.Contains(t, row.SkipReason, "no tracked base")
}

func TestReadWorktreeStatus_LastCommitDatePopulated(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.False(t, row.LastCommitDate.IsZero(),
		"LastCommitDate must be populated from go-git ListBranches when the branch exists")
}

func TestReadWorktreeStatus_LastCommitDateUnsetWhenBranchAbsent(t *testing.T) {
	t.Parallel()

	// Repo with no secondary branch — populateLastCommitDate will
	// not find the branch in ListBranches and must leave the zero
	// time without marking the row skipped. The porcelain still
	// names "feat/x" so the worktree's branch is set; the rev-list
	// mocks let populateDirtyAndAheadBehind complete.
	repoPath := setupStatusTestRepo(t, "")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.True(t, row.LastCommitDate.IsZero(),
		"absent branch must leave LastCommitDate at the zero time")
	assert.False(t, row.IsSkipped,
		"absent last-commit date must not, by itself, mark the row skipped")
}

func TestReadWorktreeStatus_IsStaleAdapterLeavesZero(t *testing.T) {
	t.Parallel()

	// Adapter contract: IsStale is set to false (zero) here; the cmd
	// layer fills it per row via ComputeIsStale(cfg). Any populated
	// row must carry IsStale = false on return.
	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "9999")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err)
	assert.False(t, row.IsStale,
		"adapter must leave IsStale at zero even when the counts would trip the heuristic")
}

func TestReadWorktreeStatus_LastCheckedRecent(t *testing.T) {
	t.Parallel()

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	before := time.Now()
	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	after := time.Now()
	require.NoError(t, err)
	mockExecutor.AssertExpectations(t)

	assert.False(t, row.LastChecked.IsZero())
	assert.True(t, !row.LastChecked.Before(before) && !row.LastChecked.After(after),
		"LastChecked must be within the test's wall-clock window")
}

// TestClient_WorktreeStatusReader is the sentinel-passthrough test:
// *Client satisfies core.WorktreeStatusReader at compile time and a
// real call returns a populated value. The compile-time guard lives
// in internal/git/client.go.
func TestClient_WorktreeStatusReader(t *testing.T) {
	t.Parallel()

	var _ core.WorktreeStatusReader = (*Client)(nil)

	repoPath := setupStatusTestRepo(t, "feat/x")
	client, mockExecutor := statusTestClient(t, repoPath, "feat/x")
	mockWorktreeList(mockExecutor, repoPath, "feat/x")
	mockTrackedBase(mockExecutor, "")
	mockBranchMerged(mockExecutor, "", 0)
	mockRevList(mockExecutor, repoPath, "main", "feat/x", "0")
	ctx := context.Background()
	cfg := core.DefaultConfig()

	row, err := client.ReadWorktreeStatus(ctx, cfg, repoPath, repoPath)
	require.NoError(t, err, "runtime call to ReadWorktreeStatus must succeed for a fresh repo")
	assert.Equal(t, "main", row.Base, "runtime call must populate the Base field")
}
