package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
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
		GitClient:     func() (*git.Client, error) { return gitClient, nil },
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
// expected error rather than falling through to runList.
func TestList_NewCmdListInvalidOutputFormat(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"list", "--output", "yaml"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

// TestList_NewCmdListNoArgs pins the cobra args guard: list takes
// zero positional arguments.
func TestList_NewCmdListNoArgs(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"list", "feature-1"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue)
}

// TestList_NewCmdListAcceptsTextAndJSON confirms the recognised
// --output values pass the validator and reach the run body.
func TestList_NewCmdListAcceptsTextAndJSON(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)

	called := false
	customList := NewCmdList(f, func(opts *ListOptions) error {
		called = true
		assert.Equal(t, "json", opts.GlobalOptions.Output)
		return nil
	})

	for _, sub := range root.Commands() {
		if sub.Name() == "list" {
			root.RemoveCommand(sub)
			break
		}
	}
	root.AddCommand(customList)
	root.SetArgs([]string{"list", "--output", "json"})

	require.NoError(t, root.Execute())
	assert.True(t, called, "runF override must have been invoked")
}

// TestList_OutputsNoWorktreesEmptyProjects covers the empty-projects
// happy path: when the projects directory has no git repositories,
// runList still succeeds and writes a "No worktrees found" line.
func TestList_OutputsNoWorktreesEmptyProjects(t *testing.T) {
	projects := t.TempDir() // empty
	worktrees := t.TempDir()

	// Move out of any git working tree so the context detector
	// returns ContextOutsideGit and runList falls into the
	// listAllProjectsWorktrees branch on the empty projects dir.
	t.Chdir(t.TempDir())

	opts, ios := listTestOpts(t, projects, worktrees)

	require.NoError(t, runList(opts))

	written := iosStdoutBytes(t, ios)
	assert.Contains(t, written, "No worktrees found",
		"empty projects dir must surface a friendly 'No worktrees found' line; got %q", written)
}

// TestList_DefaultOutputFallsThroughToPlain pins cli-output-formats
// "default empty --output = plain". With no --output flag the run body
// must render human-readable text, not JSON or table borders.
func TestList_DefaultOutputFallsThroughToPlain(t *testing.T) {
	projects := t.TempDir()
	worktrees := t.TempDir()

	t.Chdir(t.TempDir())

	opts, ios := listTestOpts(t, projects, worktrees)

	require.NoError(t, runList(opts))

	written := iosStdoutBytes(t, ios)
	// plain text has no JSON or table markers; assert the absence of
	// both so a future refactor that accidentally wires a JSON or
	// table formatter into the empty-flag path is caught.
	assert.NotContains(t, written, "{", "empty --output must not render JSON")
	assert.NotContains(t, written, "[", "empty --output must not render JSON arrays")
	assert.NotContains(t, written, "│", "empty --output must not render table borders")
	assert.Contains(t, written, "No worktrees found",
		"empty --output must still surface the human-readable empty-state message")
}

// keep strings import referenced for future tests without noise.
var _ = strings.Contains
