package cmd

import (
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteTestOpts assembles DeleteOptions against a scratch temp-dir
// config. Mirrors the newTestFactory pattern but allows per-test
// override of fields like Target / IsForce.
func deleteTestOpts(t *testing.T) (*DeleteOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()
	f := newTestFactory(t)

	opts := &DeleteOptions{
		IO:            ios,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        "demo/feat",
		IsForce:       true,
	}
	return opts, ios
}

// TestResolveDeleteTarget_EmptyIdentifier pins the empty-identifier
// guard: the resolver surfaces "empty identifier" and runDelete must
// wrap it. The shape matches the rest of the error-dispatch table
// (errors.Is walks to *core.OperationError via the Op-prefix walk).
func TestResolveDeleteTarget_EmptyIdentifier(t *testing.T) {
	t.Parallel()

	opts, _ := deleteTestOpts(t)
	opts.Target = ""

	cfg, err := opts.Config()
	require.NoError(t, err)

	_, err = resolveDeleteTarget(t.Context(), &core.Context{Type: core.ContextProject, ProjectName: "demo"}, cfg, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve target",
		"resolve failure must wrap the resolver error for cmd-boundary dispatch")
}

// TestGetDeleteNavigationTarget pins the navigation-target rule:
// only ContextWorktree callers get a project-root path; other
// contexts stay in place (empty string).
func TestGetDeleteNavigationTarget(t *testing.T) {
	t.Parallel()

	t.Run("worktree context returns project root", func(t *testing.T) {
		t.Parallel()
		ctx := &core.Context{Type: core.ContextWorktree, ProjectName: "demo"}
		got := getDeleteNavigationTarget(ctx, "/projects", "/worktrees/demo/feat")
		assert.Equal(t, "/projects/demo", got)
	})

	t.Run("project context returns empty", func(t *testing.T) {
		t.Parallel()
		ctx := &core.Context{Type: core.ContextProject, ProjectName: "demo"}
		got := getDeleteNavigationTarget(ctx, "/projects", "/projects/demo/feat")
		assert.Empty(t, got, "project-context callers stay in place; nav must be empty")
	})

	t.Run("outside-git context returns empty", func(t *testing.T) {
		t.Parallel()
		ctx := &core.Context{Type: core.ContextOutsideGit}
		got := getDeleteNavigationTarget(ctx, "/projects", "/worktrees/demo/feat")
		assert.Empty(t, got, "outside-git callers asked for explicit deletion; nav must be empty")
	})
}

// TestEmitDeleteResult_CdFlag pins the post-delete -C contract:
// the navigation target must land on stdout when opts.IsChangeDir
// is set, regardless of quiet mode.
func TestEmitDeleteResult_CdFlag(t *testing.T) {
	t.Parallel()

	opts, _ := deleteTestOpts(t)
	opts.IsChangeDir = true
	ctx := &core.Context{Type: core.ContextWorktree, ProjectName: "demo"}

	require.NoError(t, emitDeleteResult(opts, ctx, "/projects", "/worktrees/demo/feat"))

	stdout := iosOutBytes(t, opts.IO)
	assert.Contains(t, stdout, "/projects/demo",
		"worktree-context delete must emit project root for cd navigation")
}

// TestEmitDeleteResult_Quiet pins the non-cd quiet contract:
// quiet callers receive no stdout output (the deleted-ack line is
// suppressed).
func TestEmitDeleteResult_Quiet(t *testing.T) {
	t.Parallel()

	opts, _ := deleteTestOpts(t)
	opts.IsChangeDir = false
	opts.IO.IsQuiet = true

	require.NoError(t, emitDeleteResult(opts, &core.Context{}, "/projects", "/worktrees/demo/feat"))

	stdout := iosOutBytes(t, opts.IO)
	assert.Empty(t, stdout, "quiet mode must suppress the deleted-ack line")
}

// TestEmitDeleteResult_NonQuiet pins the default contract: non-quiet
// non-cd callers receive the human-readable deleted-ack line.
func TestEmitDeleteResult_NonQuiet(t *testing.T) {
	t.Parallel()

	opts, _ := deleteTestOpts(t)
	opts.IsChangeDir = false
	opts.IO.IsQuiet = false

	require.NoError(t, emitDeleteResult(opts, &core.Context{}, "/projects", "/worktrees/demo/feat"))

	stdout := iosOutBytes(t, opts.IO)
	assert.Contains(t, stdout, "Deleted worktree",
		"non-quiet non-cd delete must emit the deleted-ack line")
}

// TestDeleteOptions_GitClientAccessor pins the typed-narrowing seam:
// opts.gitClient() unwraps the cmdutil.Client (any) into the
// concrete *git.Client, returning nil on failure rather than
// panicking.
func TestDeleteOptions_GitClientAccessor(t *testing.T) {
	t.Parallel()

	opts, _ := deleteTestOpts(t)
	require.NotNil(t, opts.gitClient(),
		"factory-wired GitClient must unwrap to *git.Client without panic")

	opts.GitClient = nil
	assert.Nil(t, opts.gitClient(), "nil GitClient closure must return nil without panic")
}
