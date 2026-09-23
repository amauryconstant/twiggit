// Command twiggit is the worktree-management CLI entry point.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"twiggit/cmd"
	"twiggit/internal/cmdutil"
	"twiggit/internal/output"
)

func main() {
	slogLevel := slog.LevelInfo
	if os.Getenv("TWIGGIT_DEBUG") != "" {
		slogLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Internal error: %v\n", r)
			if os.Getenv("TWIGGIT_DEBUG") != "" {
				fmt.Fprintln(os.Stderr, "\nStack trace:")
				debug.PrintStack()
			}
			os.Exit(1)
		}
	}()

	factory := cmdutil.NewFactory()
	if err := factory.Init(); err != nil {
		output.FormatError(os.Stderr, err, nil)
		os.Exit(int(cmdutil.ExitCodeFor(err)))
	}
	factory.Context = ctx

	rootCmd := cmd.NewRootCommand(factory)
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		// Signal cancellation (SIGINT/SIGTERM) bypasses the error
		// formatter and ExitCodeFor per cli-exit-codes and
		// cli-error-formatting specs: the binary exits 130 (SIGINT)
		// or 143 (SIGTERM) directly. The formatter MUST NOT run for
		// signal-cancelled invocations.
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				os.Exit(130)
			}
			os.Exit(143)
		}

		ios := factory.IOStreams
		exitCode := cmdutil.ExitCodeFor(err)
		if exitCode == cmdutil.ExitUsage {
			_, _ = fmt.Fprintf(ios.ErrOut, "Error: %s\n", err.Error())
		} else {
			output.FormatError(ios.ErrOut, err, ios)
		}
		if exitCode != 0 {
			os.Exit(int(exitCode))
		}
	}
}
