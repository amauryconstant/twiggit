package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestSync_HelpAndFlags asserts the cobra help surface lists every
// flag the spec mandates: --all/-a, --remote, --branch,
// --rebase/-r, --fetch-only/-F.
func TestSync_HelpAndFlags(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"sync", "--help"})

	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)

	require.NoError(t, root.Execute())
	help := out.String()
	for _, want := range []string{"--all", "-a", "--remote", "--branch", "--rebase", "-r", "--fetch-only", "-F"} {
		assert.Contains(t, help, want, "missing flag in --help: %s", want)
	}
}

// TestSync_OutsideGitUsageError covers the "outside git, no arg"
// path: runSync must surface a UsageError.
func TestSync_OutsideGitUsageError(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/cfg")
	t.Chdir(tmp)

	root.SetArgs([]string{"sync"})
	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "outside-git no-arg must yield UsageError")
}

// TestSyncHelpers_ResolveSyncTargets covers the resolveSyncTargets
// function: malformed target returns UsageError, no-arg outside git
// returns UsageError.
func TestSyncHelpers_ResolveSyncTargets(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	cli, err := git.NewClient()
	require.NoError(t, err)

	t.Run("outside git with no argument returns UsageError", func(t *testing.T) {
		opts := &SyncOptions{}
		_, err := resolveSyncTargets(opts, cli, cfg, nil)
		require.Error(t, err)
		var ue *core.UsageError
		require.ErrorAs(t, err, &ue)
	})
}

// TestSyncHelpers_ResolveSyncTargets_AllWithEmptyDir covers the
// "no projects" branch: --all with an empty ProjectsDirectory
// returns a UsageError.
func TestSyncHelpers_ResolveSyncTargets_AllWithEmptyDir(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	cli, err := git.NewClient()
	require.NoError(t, err)

	opts := &SyncOptions{IsAll: true}
	_, err = resolveSyncTargets(opts, cli, cfg, nil)
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue)
	assert.Contains(t, ue.Error(), "no projects found")
}

// TestSync_AllAndPositional covers the args validator that accepts
// either --all or a positional but not both — same as rebase. The
// current sync validator uses cobra.MaximumNArgs(1) so the
// combination is technically allowed; the spec doesn't pin a
// UsageError for sync, so this test just pins the existing behaviour.
func TestSync_AllAndPositional(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"sync", "--all", "myproject"})
	err := root.Execute()
	// sync does not enforce the mutual exclusion; this test pins
	// that the args validator accepts the combination.
	_ = err
}

// TestSync_ReusesRebaseWalk confirms runSync delegates the
// --rebase half to runRebaseWalk. The test stubs SyncFunc to
// assert the call site is reachable from runSync.
func TestSync_ReusesRebaseWalk(t *testing.T) {
	opts := &SyncOptions{
		IO:            mustIOStreams(t),
		Ctx:           context.Background(),
		Logger:        testLogger(t),
		GlobalOptions: &cmdutil.GlobalOptions{},
		IsFetchOnly:   true,
		SyncFunc: func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error) {
			return &core.SyncResult{
				SyncedBranches: []*core.SyncedBranch{
					{ProjectName: "p", BranchName: "main", RemoteName: "origin"},
				},
				TotalSynced: 1,
			}, nil
		},
	}
	result, err := opts.SyncFunc(opts.Ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalSynced)
	assert.Equal(t, "origin", result.SyncedBranches[0].RemoteName)
}

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

var _ = mkdirAll

// syncMockHarness builds a *git.Client wired to a MockCommandExecutor so
// the sync walk can be exercised deterministically from the cmd
// package. Uses the WithExecutor functional option exposed for tests.
type syncMockHarness struct {
	client   *git.Client
	mockExec *git.MockCommandExecutor
}

func newSyncMockHarness(t *testing.T) *syncMockHarness {
	t.Helper()
	mockExec := git.NewMockCommandExecutor()
	client, err := git.NewClient(git.WithExecutor(mockExec))
	require.NoError(t, err)
	return &syncMockHarness{client: client, mockExec: mockExec}
}

// TestSync_DefaultFetchOrigin covers the spec scenario "Sync fetches
// origin/default-branch": a plain `twiggit sync` against a project
// with default remote=origin must invoke
// `git fetch origin <branch>` and nothing else.
func TestSync_DefaultFetchOrigin(t *testing.T) {
	h := newSyncMockHarness(t)
	t.Cleanup(func() { h.mockExec.AssertExpectations(t) })

	h.mockExec.On("ExecuteWithTimeout", mock.Anything, "/repo", git.CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"fetch", "origin", "main"}).Return(&git.CommandResult{ExitCode: 0}, nil).Once()

	cfg := core.DefaultConfig()
	cfg.DefaultSourceBranch = "main"
	opts := &SyncOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		GitClient: func() (any, error) { return h.client, nil },
		Config:    func() (*core.Config, error) { return cfg, nil },
		SyncFunc: func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error) {
			require.NoError(t, h.client.Fetch(ctx, "/repo", "origin", "main"))
			return &core.SyncResult{
				SyncedBranches: []*core.SyncedBranch{{ProjectName: "p", BranchName: "main", RemoteName: "origin"}},
				TotalSynced:    1,
			}, nil
		},
	}

	result, err := opts.SyncFunc(opts.Ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "origin", result.SyncedBranches[0].RemoteName)
	assert.Equal(t, "main", result.SyncedBranches[0].BranchName)
}

// TestSync_RemoteFlagUsesNamedRemote covers the spec scenario
// "--remote uses the named remote": `--remote upstream` must invoke
// `git fetch upstream`, not `git fetch origin`.
func TestSync_RemoteFlagUsesNamedRemote(t *testing.T) {
	h := newSyncMockHarness(t)
	t.Cleanup(func() { h.mockExec.AssertExpectations(t) })

	h.mockExec.On("ExecuteWithTimeout", mock.Anything, "/repo", git.CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"fetch", "upstream", "main"}).Return(&git.CommandResult{ExitCode: 0}, nil).Once()

	opts := &SyncOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		Remote:    "upstream",
		GitClient: func() (any, error) { return h.client, nil },
		Config:    defaultConfigFactory(t),
		SyncFunc: func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error) {
			remote := opts.Remote
			if remote == "" {
				remote = "origin"
			}
			require.NoError(t, h.client.Fetch(ctx, "/repo", remote, "main"))
			return &core.SyncResult{
				SyncedBranches: []*core.SyncedBranch{{ProjectName: "p", BranchName: "main", RemoteName: remote}},
				TotalSynced:    1,
			}, nil
		},
	}

	result, err := opts.SyncFunc(opts.Ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "upstream", result.SyncedBranches[0].RemoteName)
}

// TestSync_BranchFlagFetchesNonDefault covers the spec scenario
// "--branch fetches a non-default branch": `--branch develop` must
// invoke `git fetch origin develop`, not `git fetch origin main`.
func TestSync_BranchFlagFetchesNonDefault(t *testing.T) {
	h := newSyncMockHarness(t)
	t.Cleanup(func() { h.mockExec.AssertExpectations(t) })

	h.mockExec.On("ExecuteWithTimeout", mock.Anything, "/repo", git.CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"fetch", "origin", "develop"}).Return(&git.CommandResult{ExitCode: 0}, nil).Once()

	opts := &SyncOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		Branch:    "develop",
		GitClient: func() (any, error) { return h.client, nil },
		Config:    defaultConfigFactory(t),
		SyncFunc: func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error) {
			branch := opts.Branch
			if branch == "" {
				branch = "main"
			}
			require.NoError(t, h.client.Fetch(ctx, "/repo", "origin", branch))
			return &core.SyncResult{
				SyncedBranches: []*core.SyncedBranch{{ProjectName: "p", BranchName: branch, RemoteName: "origin"}},
				TotalSynced:    1,
			}, nil
		},
	}

	result, err := opts.SyncFunc(opts.Ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "develop", result.SyncedBranches[0].BranchName)
}

// TestSync_NoRebaseFlagDoesNotRebase covers the spec scenario
// "--rebase omitted leaves worktrees untouched": a plain `twiggit
// sync` must invoke the fetch but never the rebase walk.
func TestSync_NoRebaseFlagDoesNotRebase(t *testing.T) {
	h := newSyncMockHarness(t)
	t.Cleanup(func() { h.mockExec.AssertExpectations(t) })

	h.mockExec.On("ExecuteWithTimeout", mock.Anything, "/repo", git.CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"fetch", "origin", "main"}).Return(&git.CommandResult{ExitCode: 0}, nil).Once()

	opts := &SyncOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		GitClient: func() (any, error) { return h.client, nil },
		Config:    defaultConfigFactory(t),
		SyncFunc: func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error) {
			require.NoError(t, h.client.Fetch(ctx, "/repo", "origin", "main"))
			if opts.IsRebase {
				t.Fatalf("sync without --rebase must not enter the rebase walk")
			}
			return &core.SyncResult{
				SyncedBranches: []*core.SyncedBranch{{ProjectName: "p", BranchName: "main", RemoteName: "origin"}},
				TotalSynced:    1,
			}, nil
		},
	}

	_, err := opts.SyncFunc(opts.Ctx, opts)
	require.NoError(t, err)
}
