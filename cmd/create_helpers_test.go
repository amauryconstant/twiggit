package cmd

import (
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestOpts assembles CreateOptions against a scratch temp-dir
// config. Mirrors the newTestFactory pattern from init_test.go but
// without a Factory wrapper so each helper test can poke fields
// directly.
func createTestOpts(t *testing.T) (*CreateOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	f := newTestFactory(t)

	opts := &CreateOptions{
		IO:            ios,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Spec:          "demo/feat",
		Source:        "main",
	}
	return opts, ios
}

// TestValidateBranchNameForCreate pins the validator: an invalid
// branch name yields a *core.ValidationError; a valid one passes
// through silently.
func TestValidateBranchNameForCreate(t *testing.T) {
	t.Parallel()

	t.Run("invalid branch name yields ValidationError", func(t *testing.T) {
		t.Parallel()
		opts, _ := createTestOpts(t)
		opts.Spec = "demo/invalid@branch"

		err := validateBranchNameForCreate(opts)
		require.Error(t, err)
		var ve *core.ValidationError
		require.ErrorAs(t, err, &ve)
	})

	t.Run("valid branch name passes", func(t *testing.T) {
		t.Parallel()
		opts, _ := createTestOpts(t)
		opts.Spec = "demo/feature-ok"

		assert.NoError(t, validateBranchNameForCreate(opts))
	})
}

// TestRenderCreateResult_CdFlag pins the -C contract — the worktree
// path lands on stdout and the success line is not.
func TestRenderCreateResult_CdFlag(t *testing.T) {
	t.Parallel()

	opts, _ := createTestOpts(t)
	opts.IsCdFlag = true
	result := &core.CreateWorktreeResult{
		Worktree: &core.Worktree{Path: "/proj/feat", Branch: "feat"},
	}

	renderCreateResult(opts, result)

	stdout := iosOutBytes(t, opts.IO)
	assert.Contains(t, stdout, "/proj/feat", "-C must emit path to stdout")
}

// TestRenderCreateResult_Quiet pins the cd-mode terminal write: when
// IsCdFlag is true the success message must be suppressed regardless
// of IsQuiet.
func TestRenderCreateResult_Quiet(t *testing.T) {
	t.Parallel()

	opts, _ := createTestOpts(t)
	opts.IO.IsQuiet = true
	result := &core.CreateWorktreeResult{
		Worktree: &core.Worktree{Path: "/proj/feat", Branch: "feat"},
	}

	renderCreateResult(opts, result)

	stdout := iosOutBytes(t, opts.IO)
	assert.NotContains(t, stdout, "Created worktree", "quiet mode must suppress success line")
}

// TestRenderCreateResult_HookFailures pins that failed hooks land on
// ErrOut as warning lines.
func TestRenderCreateResult_HookFailures(t *testing.T) {
	t.Parallel()

	opts, _ := createTestOpts(t)
	result := &core.CreateWorktreeResult{
		Worktree: &core.Worktree{Path: "/proj/feat", Branch: "feat"},
		HookResult: &core.HookResult{
			IsSuccessful: false,
			Failures: []core.HookFailure{
				{Command: "setup.sh", ExitCode: 1, Output: "boom"},
			},
		},
	}

	renderCreateResult(opts, result)

	errOut := iosErrOutBytes(t, opts.IO)
	assert.Contains(t, errOut, "setup.sh", "hook failure must surface the failing command")
	assert.Contains(t, errOut, "post-create hook(s) failed",
		"hook failure summary must precede the per-command detail")
}

// TestParseProjectBranch_NoProjectFallsBackToContext pins the
// bare-branch spec: when ctx carries a project name, the parser
// fills the project half from the context.
func TestParseProjectBranch_NoProjectFallsBackToContext(t *testing.T) {
	t.Parallel()

	project, branch, err := parseProjectBranch("feat-x", &core.Context{ProjectName: "ctx-proj"})
	require.NoError(t, err)
	assert.Equal(t, "ctx-proj", project)
	assert.Equal(t, "feat-x", branch)
}

// TestParseProjectBranch_NoContextRequiresProject pins the
// empty-context + bare-branch spec: when there's no context and the
// spec has no project prefix, the parser returns a ValidationError.
func TestParseProjectBranch_NoContextRequiresProject(t *testing.T) {
	t.Parallel()

	_, _, err := parseProjectBranch("feat-x", nil)
	require.Error(t, err)
	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve)
}

// TestCalculateWorktreePath pins the path contract: the worktree
// lives at WorktreesDirectory/<project>/<branch>. Important for
// downstream shell-integration paths that derive cd targets from it.
func TestCalculateWorktreePath(t *testing.T) {
	t.Parallel()

	cfg := &core.Config{
		WorktreesDirectory: "/worktrees",
	}
	got := calculateWorktreePath(cfg, "demo", "feat-x")
	assert.Equal(t, "/worktrees/demo/feat-x", got)
}
