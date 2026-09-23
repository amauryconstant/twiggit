package cmdutil_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
)

func TestSignalExitCode_NilCtx(t *testing.T) {
	is := assert.New(t)

	code, ok := cmdutil.SignalExitCode(nil)

	is.False(ok)
	is.Zero(code)
}

func TestSignalExitCode_BackgroundCtx(t *testing.T) {
	is := assert.New(t)

	code, ok := cmdutil.SignalExitCode(context.Background())

	is.False(ok)
	is.Zero(code)
}

func TestSignalExitCode_CanceledCtx(t *testing.T) {
	is := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	code, ok := cmdutil.SignalExitCode(ctx)

	is.True(ok)
	is.Equal(130, code, "SIGINT → 130 (128 + SIGINT 2)")
}

func TestSignalExitCode_DeadlineExceededCtx(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		is := assert.New(t)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		time.Sleep(time.Second - time.Nanosecond)
		synctest.Wait()
		require.NoError(t, ctx.Err())

		time.Sleep(time.Nanosecond)
		synctest.Wait()
		require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)

		code, ok := cmdutil.SignalExitCode(ctx)

		is.True(ok)
		is.Equal(143, code, "deadline / SIGTERM-like ctx.Err() → 143 (128 + SIGTERM 15)")
	})
}

func TestSignalExitCode_ParentCanceledPropagatesToChild(t *testing.T) {
	is := assert.New(t)

	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithCancel(parent)
	defer cancelChild()

	cancelParent()

	is.ErrorIs(parent.Err(), context.Canceled)
	is.ErrorIs(child.Err(), context.Canceled)

	code, ok := cmdutil.SignalExitCode(child)

	is.True(ok)
	is.Equal(130, code)
}
