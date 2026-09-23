package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"twiggit/cmd"
	"twiggit/internal/cmdutil"
)

func main() {
	slogLevel := slog.LevelInfo
	if os.Getenv("TWIGGIT_DEBUG") != "" {
		slogLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slogLevel})))

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
		os.Exit(int(cmd.HandleCLIError(err)))
	}

	rootCmd := cmd.NewRootCommand(factory)

	if err := rootCmd.Execute(); err != nil {
		exitCode := cmd.HandleCLIErrorWithCommand(rootCmd, err)
		os.Exit(int(exitCode))
	}
}
