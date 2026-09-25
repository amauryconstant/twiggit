// Path: main.go (module root — the composition root)
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/you/myapp/cmd"
	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/output"
)

func main() {
	os.Exit(run())
}

// run owns the process lifecycle and returns the exit code, so every deferred
// cleanup runs before os.Exit (which skips defers).
func run() (code int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", r)
			if os.Getenv("MYAPP_DEBUG") != "" {
				fmt.Fprintf(os.Stderr, "\n%s\n", debug.Stack())
			}
			code = int(cmdutil.ExitError)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	f := cmdutil.NewFactory()
	err := cmd.NewRootCommand(f).ExecuteContext(ctx)
	switch {
	case err == nil:
		return int(cmdutil.ExitOK)
	case ctx.Err() != nil:
		// ponytail: SIGTERM also exits 130; split on the signal if callers need 143.
		return int(cmdutil.ExitInterrupt)
	default:
		output.FormatError(f.IOStreams, err)
		return int(cmdutil.ExitCodeFor(err))
	}
}
