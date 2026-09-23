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
		exitCode := signalAwareExitCode(ctx, err)
		if exitCode != 0 {
			os.Exit(int(exitCode))
		}
	}
}

func signalAwareExitCode(ctx context.Context, err error) cmdutil.ExitCode {
	if errors.Is(ctx.Err(), context.Canceled) {
		return cmdutil.ExitError
	}
	return cmdutil.ExitCodeFor(err)
}
