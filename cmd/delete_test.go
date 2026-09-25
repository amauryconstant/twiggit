package cmd

import (
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCmdDelete_RequiresExactlyOneArg pins the cobra args guard.
func TestNewCmdDelete_RequiresExactlyOneArg(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"delete"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "missing positional arg must yield *core.UsageError")
}

// TestNewCmdDelete_AliasRm_AcceptsSameArg checks the `rm` alias of
// delete takes the same args shape.
func TestNewCmdDelete_AliasRm_AcceptsSameArg(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"rm"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "missing arg via rm alias must yield *core.UsageError")
}

// TestRunDelete_NoTargetErrors covers the path where runDelete is
// invoked with an empty target. The resolver returns
// "empty identifier" which our code rewraps as an OperationError.
func TestRunDelete_NoTargetErrors(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := newTestFactory(t)

	opts := &DeleteOptions{
		IO:            ios,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        "",
		Force:         true,
	}

	err := runDelete(opts)
	require.Error(t, err)
	assert.NotEmpty(t, err.Error(),
		"runDelete with empty target must surface an error")
}
