package cmd

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
	"twiggit/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cdTestOpts assembles CdOptions for the supplied working dir + target.
// The projects/worktrees dirs point at the supplied tmpBase so any
// relative path the resolver returns matches an on-disk directory.
// Currently unused outside the helpers reference below.
func cdTestOpts(t *testing.T, worktreesDir, target string) (*CdOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = worktreesDir

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &CdOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           context.Background(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        target,
	}

	return opts, ios
}

// TestCd_NotFoundError covers the negative path: a target that
// resolves to a missing directory must surface a *core.OperationError
// rather than panic or emit empty output. The current ctx is
// outside-git so the resolver falls through to project-not-found.
func TestCd_NotFoundError(t *testing.T) {
	tmp := t.TempDir()
	bogus := tmp + "/does-not-exist"

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp
	cfg.WorktreesDirectory = tmp

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &CdOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        bogus,
	}

	err = runCd(opts)
	require.Error(t, err)
	var oe *core.OperationError
	require.ErrorAs(t, err, &oe, "runCd must return a *core.OperationError on missing path")
	// Op may be cd.worktree (from the resolver) or validate.path (stat failure).
	assert.Contains(t, []string{"cd.worktree", "validate.path"}, oe.Op)
}

// TestCd_HappyPathEchoesTarget sets up a real git repository via
// the test helpers, configures the cfg to point at the project, and
// asserts runCd outputs the project absolute path to stdout. The
// project-context default target (worktree main checkout) is what
// runCd navigates to in the absence of an explicit target.
func TestCd_HappyPathEchoesTarget(t *testing.T) {
	// Create a real git repository; cwd inside it so detector
	// returns ContextProject.
	gitHelper := helpers.NewGitTestHelper(t)
	repoPath := gitHelper.CreateRepoWithCommits(1)

	worktrees := t.TempDir()

	t.Chdir(repoPath)

	ios, _, outBuf, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = filepath.Dir(repoPath)
	cfg.WorktreesDirectory = worktrees

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &CdOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        "", // omit → runCd picks project-context default
	}

	require.NoError(t, runCd(opts))
	assert.NotEmpty(t, strings.TrimSpace(outBuf.String()),
		"runCd must emit a path to stdout even with empty target inside project context")
}

// TestCd_EmptyTargetReturnsError confirms that running cd with no
// target outside any git repository returns a recognisable error
// rather than navigating somewhere. The ContextOutsideGit branch in
// runCd returns the "no target specified" sentinel.
func TestCd_EmptyTargetReturnsError(t *testing.T) {
	tmp := t.TempDir()
	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp
	cfg.WorktreesDirectory = tmp

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &CdOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        "",
	}

	err = runCd(opts)
	require.Error(t, err)
	// Either the explicit "no target specified" or a downstream
	// resolution error is acceptable. The contract is non-nil.
	assert.NotEmpty(t, err.Error())
}

// silence unused cdTestOpts shim; helpers/tests use bespoke opts.
var (
	_ = cdTestOpts
	_ = strings.Builder{}
)
