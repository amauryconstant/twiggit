package cmdutil_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
)

// TestExitCodeFor_NilReturnsExitOK confirms the no-error path is the
// success code, not a zero-default surprise.
func TestExitCodeFor_NilReturnsExitOK(t *testing.T) {
	t.Parallel()

	assert.Equal(t, cmdutil.ExitOK, cmdutil.ExitCodeFor(nil))
}

// TestExitCodeFor_UsageErrorReturnsExitUsage confirms the typed usage
// contract is honored, including when wrapped via fmt.Errorf.
func TestExitCodeFor_UsageErrorReturnsExitUsage(t *testing.T) {
	t.Parallel()

	t.Run("direct", func(t *testing.T) {
		t.Parallel()
		err := &core.UsageError{Message: "missing flag"}
		assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(err))
	})

	t.Run("wrapped", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("cmd: %w", &core.UsageError{Message: "bad arg"})
		assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(wrapped))
	})
}

// TestExitCodeFor_OtherErrorReturnsExitError confirms the fallthrough
// is ExitError (1), never ExitUsage, for non-typed errors.
func TestExitCodeFor_OtherErrorReturnsExitError(t *testing.T) {
	t.Parallel()

	assert.Equal(t, cmdutil.ExitError, cmdutil.ExitCodeFor(errors.New("boom")))
}

// TestExitCode_String confirms the human-readable text format used in
// logs and golden tests.
func TestExitCode_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code cmdutil.ExitCode
		want string
	}{
		{cmdutil.ExitOK, "0"},
		{cmdutil.ExitError, "1"},
		{cmdutil.ExitUsage, "2"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.code.String())
		})
	}
}

// TestExitCode_ConstantsAreUnique guards against accidental constant
// collisions during future refactors.
func TestExitCode_ConstantsAreUnique(t *testing.T) {
	t.Parallel()

	seen := map[cmdutil.ExitCode]string{
		cmdutil.ExitOK:    "ExitOK",
		cmdutil.ExitError: "ExitError",
		cmdutil.ExitUsage: "ExitUsage",
	}
	require.Len(t, seen, 3)
}
