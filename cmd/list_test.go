package cmd

import (
	"bytes"
	"strings"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRootForTest builds a root cobra.Command wired with the same
// persistent flags + subcommands as NewRootCommand, but without
// touching cmd/root.go directly so tests can supply overrides via
// the optional mutate hook.
func newRootForTest(f *cmdutil.Factory, mutate ...func(*cobra.Command)) *cobra.Command {
	root := NewRootCommand(f)
	if len(mutate) > 0 {
		mutate[0](root)
	}
	return root
}

// listTestOpts assembles ListOptions for the supplied cfg override.
// Returns options + iostreams.Test IOS so tests can assert on stdout.
func listTestOpts(t *testing.T, projectsDir, worktreesDir string) (*ListOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = projectsDir
	cfg.WorktreesDirectory = worktreesDir

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &ListOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
	}

	return opts, ios
}

// iosStdoutBytes is a helper returning the ios Stdout buffer's
// string content when ios.Stdout is a *bytes.Buffer.
func iosStdoutBytes(t *testing.T, ios *iostreams.IOStreams) string {
	t.Helper()
	if buf, ok := ios.Out.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}

// TestList_NewCmdListInvalidOutputFormat checks the NewCmdList
// flag-parse path: an invalid `--output` value fails RunE with the
// expected error rather than falling through to runList, and the
// dispatched exit code is 2 (UsageError) per cli-output spec
// `Unknown --output value is rejected`.
//
// Cannot use t.Parallel: NewRootCommand triggers
// carapace.Gen → cobra.OnInitialize which mutates cobra's
// package-level initializer slice. Parallel tests that build a
// root tree race on that global state.
func TestList_NewCmdListInvalidOutputFormat(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"list", "--output", "yaml"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(err),
		"unknown --output must dispatch to UsageError / exit 2")
}

// TestList_NewCmdList_JSONLReturnsExitUsage pins the dropped-format
// scenario from cli-output spec: `--output jsonl` is rejected as a
// usage error with exit code 2. The constructor returns
// (nil, *core.UsageError); the wrap preserves the typed chain so
// ExitCodeFor still classifies it.
//
// Cannot use t.Parallel: see TestList_NewCmdListInvalidOutputFormat.
func TestList_NewCmdList_JSONLReturnsExitUsage(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"list", "--output", "jsonl"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(err),
		"--output jsonl must exit 2 (UsageError)")
}

// TestList_NewCmdListNoArgs pins the cobra args guard: list takes
// zero positional arguments.
//
// Cannot use t.Parallel: see TestList_NewCmdListInvalidOutputFormat.
func TestList_NewCmdListNoArgs(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"list", "feature-1"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue)
}

// TestList_OutputFlagCompletion drives cobra's `__complete` verb
// against the list command and asserts the three accepted values
// (json, table, plain) are offered while the legacy `text` value
// and the dropped `jsonl` value are NOT. This covers
// cli-output/Shell-completable S1 and cli-output-formats duplicate.
// Cobra's `__complete` request emits candidates on stdout
// separated by newlines; we join and match against the captured
// buffer.
//
// Cannot use t.Parallel: NewRootCommand triggers
// carapace.Gen → cobra.OnInitialize which mutates cobra's
// package-level initializer slice. Parallel tests that build a
// root tree race on that global state.
func TestList_OutputFlagCompletion(t *testing.T) {
	cases := []struct {
		name     string
		toType   string
		wantSubs []string
		denySubs []string
	}{
		{
			name:     "empty prefix offers the three accepted values",
			toType:   "",
			wantSubs: []string{"json", "table", "plain"},
			denySubs: []string{"text", "jsonl", "yaml", "xml"},
		},
		{
			name:     "json prefix still offers all accepted values",
			toType:   "j",
			wantSubs: []string{"json", "table", "plain"},
			denySubs: []string{"text", "jsonl", "yaml", "xml"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newTestFactory(t)
			root := newRootForTest(f)
			stdout := new(bytes.Buffer)
			root.SetOut(stdout)
			root.SetErr(stdout)
			root.SetArgs([]string{"__complete", "list", "--output", tc.toType})

			require.NoError(t, root.Execute(),
				"cobra __complete must not error for --output %q", tc.toType)

			got := stdout.String()
			for _, w := range tc.wantSubs {
				assert.Contains(t, got, w,
					"completion candidates must include %q; got %q", w, got)
			}
			for _, d := range tc.denySubs {
				assert.NotContains(t, got, d,
					"completion must not offer legacy/dropped format %q; got %q", d, got)
			}
		})
	}
}

// TestList_EmptyProjectEmitsNoLines covers cli-list/Status
// indicators/Empty project renders no lines: an empty project must
// emit no stdout lines when no `--output` flag is supplied. After
// the renderWorktrees production fix, the bespoke empty path emits
// zero lines (the for-loop runs zero iterations).
//
// Cannot use t.Parallel: t.Chdir requires a non-parallel test
// (Go testing rule — t.Chdir mutates process-global state).
func TestList_EmptyProjectEmitsNoLines(t *testing.T) {
	projects := t.TempDir()
	worktrees := t.TempDir()

	// Move out of any git working tree so the context detector
	// returns ContextOutsideGit and runList falls into the
	// listAllProjectsWorktrees branch on the empty projects dir.
	t.Chdir(t.TempDir())

	opts, ios := listTestOpts(t, projects, worktrees)

	require.NoError(t, runList(opts))

	written := iosStdoutBytes(t, ios)
	assert.Empty(t, written,
		"empty project must emit no lines per cli-list spec; got %q", written)
}

// TestRenderWorktrees_BespokeShape pins the cli-list/Status
// indicators contract for the empty `--output` path: each worktree
// renders as `branch -> path` with a single conditional suffix
// (`(modified)`, `(detached)`, or no suffix for clean attached).
// Exercises renderWorktrees directly with synthetic worktrees so
// the test does not depend on a real git repository fixture.
func TestRenderWorktrees_BespokeShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		wt        *core.WorktreeInfo
		wantLines []string
		denyLines []string
	}{
		{
			name: "dirty worktree carries (modified) suffix",
			wt: &core.WorktreeInfo{
				Branch:     "feat/foo",
				Path:       "/tmp/feat/foo",
				IsModified: true,
			},
			wantLines: []string{"feat/foo -> /tmp/feat/foo (modified)"},
			denyLines: []string{"(detached)", "[", "\t"},
		},
		{
			name: "detached HEAD carries (detached) suffix",
			wt: &core.WorktreeInfo{
				Branch:     "feat/bar",
				Path:       "/tmp/feat/bar",
				IsDetached: true,
			},
			wantLines: []string{"feat/bar -> /tmp/feat/bar (detached)"},
			denyLines: []string{"(modified)", "[", "\t"},
		},
		{
			name: "clean attached worktree has no suffix",
			wt: &core.WorktreeInfo{
				Branch: "feat/baz",
				Path:   "/tmp/feat/baz",
			},
			wantLines: []string{"feat/baz -> /tmp/feat/baz"},
			denyLines: []string{"(modified)", "(detached)", "[", "\t"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, renderWorktrees(&buf, []*core.WorktreeInfo{tc.wt}, nil))

			got := buf.String()
			lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
			require.Len(t, lines, len(tc.wantLines),
				"bespoke shape must emit exactly one line per worktree; got %q", got)
			for _, w := range tc.wantLines {
				assert.Contains(t, got, w)
			}
			for _, d := range tc.denyLines {
				assert.NotContains(t, got, d,
					"bespoke shape must not render %q markers; got %q", d, got)
			}
		})
	}
}

// keep strings import referenced for future tests without noise.
var _ = strings.Contains
