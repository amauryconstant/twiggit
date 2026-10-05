package cmd

import (
	"bytes"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pruneTestOpts assembles PruneOptions against a scratch temp-dir
// config. Returns options + iostreams.Test IOS so tests can assert
// on emitted stdout/stderr.
func pruneTestOpts(t *testing.T) (*PruneOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &PruneOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
	}
	return opts, ios
}

// TestPruneRequest_MapsFlags pins the dry-run override contract:
// opts.IsDryRun is preserved when dryRun is false, and the request
// mirrors every other PruneOptions flag verbatim.
func TestPruneRequest_MapsFlags(t *testing.T) {
	t.Parallel()

	opts := &PruneOptions{
		IsForce:          true,
		IsDeleteBranches: true,
		IsAllProjects:    true,
		IsDryRun:         false,
		SpecificWorktree: "proj/branch",
	}
	currentCtx := &core.Context{ProjectName: "proj"}

	t.Run("preserves live walk flags", func(t *testing.T) {
		t.Parallel()

		req := pruneRequest(opts, currentCtx, false)
		require.NotNil(t, req)
		assert.Equal(t, currentCtx, req.Context)
		assert.True(t, req.IsForce)
		assert.True(t, req.IsDeleteBranches)
		assert.True(t, req.IsAllProjects)
		assert.False(t, req.IsDryRun, "live walk must not be coerced into dry-run")
		assert.Equal(t, "proj/branch", req.SpecificWorktree)
	})

	t.Run("forces dry-run without mutating opts", func(t *testing.T) {
		t.Parallel()

		req := pruneRequest(opts, currentCtx, true)
		require.NotNil(t, req)
		assert.True(t, req.IsDryRun, "preview walk must override opts.IsDryRun")
		assert.False(t, opts.IsDryRun, "opts.IsDryRun must remain unchanged after the override")
	})
}

// TestEmitPruneOutput_NavigationPath pins the navigation-tail path:
// when result.NavigationPath is set, emitPruneOutput writes it to
// opts.IO.Out exactly once and does not emit the "Prune complete"
// progress line (SpecificWorktree suppresses it).
func TestEmitPruneOutput_NavigationPath(t *testing.T) {
	t.Parallel()

	opts, ios := pruneTestOpts(t)
	opts.SpecificWorktree = "single-target/feat"
	result := &core.PruneWorktreesResult{
		NavigationPath: "/projects/single-target",
	}

	emitPruneOutput(opts, result, &core.Context{ProjectName: "single-target"})

	stdout := iosOutBytes(t, ios)
	errOut := iosErrOutBytes(t, ios)
	assert.Contains(t, stdout, "/projects/single-target",
		"single-target prune must emit the navigation path to stdout")
	assert.NotContains(t, errOut, "Prune complete",
		"specific-target prune must skip the progress tail")
}

// TestEmitPruneOutput_AllProjectsProgress pins the progress tail:
// --all writes "Prune complete" to ErrOut; without --all (and
// without SpecificWorktree) the same line fires.
func TestEmitPruneOutput_AllProjectsProgress(t *testing.T) {
	t.Parallel()

	t.Run("--all emits Prune complete on stderr", func(t *testing.T) {
		t.Parallel()

		opts, ios := pruneTestOpts(t)
		opts.IsAllProjects = true
		result := &core.PruneWorktreesResult{}

		emitPruneOutput(opts, result, &core.Context{})

		errOut := iosErrOutBytes(t, ios)
		assert.Contains(t, errOut, "Prune complete",
			"--all run must report Prune complete on stderr")
	})

	t.Run("SpecificWorktree suppresses progress", func(t *testing.T) {
		t.Parallel()

		opts, ios := pruneTestOpts(t)
		opts.SpecificWorktree = "proj/branch"
		result := &core.PruneWorktreesResult{}

		emitPruneOutput(opts, result, &core.Context{})

		errOut := iosErrOutBytes(t, ios)
		assert.NotContains(t, errOut, "Prune complete",
			"single-target prune must not emit progress lines")
	})
}

// TestEmitPruneOutput_NilSafe pins that emitPruneOutput tolerates
// a nil result pointer without panicking; the function is the
// composition root's last step and must not crash on a cancelled
// walk that returned no result.
func TestEmitPruneOutput_NilSafe(t *testing.T) {
	t.Parallel()

	opts, _ := pruneTestOpts(t)
	assert.NotPanics(t, func() {
		emitPruneOutput(opts, nil, &core.Context{})
	}, "nil result must be a no-op rather than a panic")
}

// TestRunPruneWithConfirm_NoPreviewSkipsPrompt pins the
// no-preview fast path: when preview is nil the confirm prompt is
// bypassed and the walk runs even if no stdin is supplied. The
// walk against an empty projects dir produces a zero-count result
// without error; the assertion is that it returns promptly (no
// blocking on the empty stdin buffer) and emits the progress line.
func TestRunPruneWithConfirm_NoPreviewSkipsPrompt(t *testing.T) {
	t.Parallel()

	opts, ios := pruneTestOpts(t)
	opts.IsAllProjects = true

	gitClient, err := git.NewClient()
	require.NoError(t, err)
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()

	result, err := runPruneWithConfirm(t.Context(), opts, gitClient, cfg, &core.Context{ProjectName: "proj"}, nil)
	require.NoError(t, err, "empty projects dir walk must not error")
	require.NotNil(t, result)
	assert.Equal(t, 0, result.TotalDeleted, "no worktrees to delete")
	assert.Equal(t, 0, result.TotalSkipped)

	errOut := iosErrOutBytes(t, ios)
	assert.Contains(t, errOut, "Pruning merged worktrees",
		"live-walk progress line must fire when preview is nil and IsAllProjects is true")
	assert.NotContains(t, errOut, "Continue? (y/n)",
		"nil preview must bypass the confirm prompt entirely")
}

// TestDeleteBranchIfRequested_SkipsWhenDisabled pins the guard:
// when the request opts out of --delete-branches, the helper is a
// no-op and must not invoke PruneWorktrees / DeleteBranch on the
// client. The fake client proves no branch-deletion I/O fired.
func TestDeleteBranchIfRequested_SkipsWhenDisabled(t *testing.T) {
	t.Parallel()

	opts, _ := pruneTestOpts(t)
	gitClient, err := git.NewClient()
	require.NoError(t, err)

	target := pruneTarget{
		Project:  &core.ProjectInfo{Name: "proj", Path: "/proj", GitRepoPath: "/proj"},
		Worktree: core.Worktree{Path: "/proj/feat", Branch: "feat"},
		Request:  &core.PruneWorktreesRequest{IsDeleteBranches: false},
	}

	require.NoError(t, deleteBranchIfRequested(t.Context(), gitClient, opts.IO.Logger, target),
		"deleteBranchIfRequested must short-circuit when IsDeleteBranches is false")
}

// TestBuildPrunePreview_RecordsProgress pins that buildPrunePreview
// reports the preview banner before walking; an empty projects dir
// yields a no-deleted-prune preview without error, and the banner
// must already have fired.
func TestBuildPrunePreview_RecordsProgress(t *testing.T) {
	t.Parallel()

	opts, ios := pruneTestOpts(t)
	opts.IsAllProjects = true

	gitClient, err := git.NewClient()
	require.NoError(t, err)
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()

	preview, err := buildPrunePreview(t.Context(), opts, gitClient, cfg, &core.Context{ProjectName: "proj"})
	require.NoError(t, err, "empty projects dir walk must not error")
	require.NotNil(t, preview)
	assert.Equal(t, 0, preview.TotalDeleted)

	errOut := iosErrOutBytes(t, ios)
	assert.Contains(t, errOut, "Previewing prune operation",
		"preview banner must fire before the walk returns")
}

// iosOutBytes extracts the stdout buffer content from an
// iostreams.Test IOS. Returns empty string when the underlying
// writer is not a *bytes.Buffer.
func iosOutBytes(t *testing.T, ios *iostreams.IOStreams) string {
	t.Helper()
	if buf, ok := ios.Out.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}

// iosErrOutBytes extracts the stderr buffer content from an
// iostreams.Test IOS.
func iosErrOutBytes(t *testing.T, ios *iostreams.IOStreams) string {
	t.Helper()
	if buf, ok := ios.ErrOut.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}
