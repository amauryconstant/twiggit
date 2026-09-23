package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
)

// TestNewCmdPrune_MaximumOneArg pins the cobra args guard: prune
// takes at most one optional positional argument.
func TestNewCmdPrune_MaximumOneArg(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"prune", "proj/branch", "extra"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "extra positional arg must yield *core.UsageError")
}

// TestNewCmdPrune_NoArgsIsOK pins the empty-args path: `twiggit prune`
// with no positional arg is valid (uses the current context).
func TestNewCmdPrune_NoArgsIsOK(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"prune", "--yes", "--force"})

	// We expect either a successful run or a downstream git failure
	// (since the test runs outside any project) but NOT a UsageError.
	err := root.Execute()
	if err != nil {
		// The test passes when the error is non-usage (i.e. real
		// runtime failure from missing git context).
		assert.NotContains(t, err.Error(), "Usage:")
		assert.NotContains(t, err.Error(), "Error: accepts 0 arg")
	}
}

// TestPruneOptions_DefaultsFalsy ensures the zero-value PruneOptions
// is non-action: every boolean defaults to false. This is the safety
// contract — `twiggit prune --all` requires an explicit flag.
func TestPruneOptions_DefaultsFalsy(t *testing.T) {
	opts := &PruneOptions{}

	assert.False(t, opts.Force)
	assert.False(t, opts.Yes)
	assert.False(t, opts.DeleteBranches)
	assert.False(t, opts.AllProjects)
	assert.False(t, opts.DryRun)
	assert.Empty(t, opts.SpecificWorktree)
}

// TestPrune_SpecificWorktreeValidation covers the parseProjectBranch
// validator: `prune` with a malformed `project/branch` value must
// surface a *core.ValidationError. The test calls the underlying
// project-branch validator directly so it doesn't need a git setup.
func TestPrune_SpecificWorktreeValidation(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)

	called := false
	pruneCmd := NewCmdPrune(f, func(opts *PruneOptions) error {
		called = true
		opts.SpecificWorktree = "/" // malformed
		// Probe the validator directly: parseProjectBranch is package-private
		// to cmd, so we exercise it through runPrune. But runPrune also
		// tries git ops — so we test only the parseProjectBranch call here.
		_, _, err := parseProjectBranch("/", nil)
		return err
	})

	for _, sub := range root.Commands() {
		if sub.Name() == "prune" {
			root.RemoveCommand(sub)
			break
		}
	}
	root.AddCommand(pruneCmd)
	root.SetArgs([]string{"prune", "/"})

	err := root.Execute()
	require.True(t, called, "runPrune override must be invoked")
	require.Error(t, err)
	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve, "malformed project/branch spec must yield *core.ValidationError")
}
