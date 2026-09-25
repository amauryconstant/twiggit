package cmd

import (
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreate_BranchValidationError covers the validation gate at
// the top of runCreate: an invalid branch-name string returns a
// *core.ValidationError instead of attempting any git I/O.
func TestCreate_BranchValidationError(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)

	captured := false
	createCmd := NewCmdCreate(f, func(opts *CreateOptions) error {
		captured = true
		opts.Spec = "invalid@branch"
		opts.Source = "main"
		return runCreate(opts)
	})

	for _, sub := range root.Commands() {
		if sub.Name() == "create" {
			root.RemoveCommand(sub)
			break
		}
	}
	root.AddCommand(createCmd)
	root.SetArgs([]string{"create", "invalid@branch"})

	err := root.Execute()
	require.True(t, captured, "runCreate must be invoked")
	require.Error(t, err)
	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve, "invalid branch name must yield *core.ValidationError")
	assert.Equal(t, "BranchName", ve.Field)
}

// TestNewCmdCreate_RequiresExactlyOneArg pins the cobra args guard.
func TestNewCmdCreate_RequiresExactlyOneArg(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"create"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "missing positional arg must yield *core.UsageError")
}

// TestNewCmdCreate_RejectsTooManyArgs pins the symmetric guard.
func TestNewCmdCreate_RejectsTooManyArgs(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"create", "feature", "extra"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "extra positional arg must yield *core.UsageError")
}
