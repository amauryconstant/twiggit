// Command twiggit is the worktree-management CLI entry point.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"twiggit/cmd"
	"twiggit/internal/cmdutil"
	"twiggit/internal/config"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/output"
	"twiggit/internal/version"
)

// loadConfig closes over config.NewManager so cmdutil stays free of
// internal/config imports. main.go is the sole composition root for
// the config dependency.
func loadConfig() (*core.Config, error) {
	//nolint:wrapcheck // composition-root closure: cmdutil.WithConfigLoader expects the raw (T, error) signature.
	return config.NewManager().Load()
}

func main() {
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

	factory := cmdutil.NewFactory(
		cmdutil.WithVersion(version.Version),
		cmdutil.WithConfigLoader(loadConfig),
		cmdutil.WithGitClientFactory(func() (cmdutil.Client, error) {
			return git.NewClient()
		}),
	)
	slog.SetDefault(factory.Logger())
	if err := factory.Init(); err != nil {
		output.FormatError(os.Stderr, err, nil)
		//nolint:gocritic // exitAfterDefer: factory init failure aborts the process; pending defers (slog, signal) are non-essential.
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
		if code, ok := cmdutil.SignalExitCode(ctx); ok {
			os.Exit(code)
		}

		ios := factory.IOStreams
		exitCode := cmdutil.ExitCodeFor(err)
		output.FormatError(ios.ErrOut, err, ios)
		if exitCode != 0 {
			os.Exit(int(exitCode))
		}
	}
}
