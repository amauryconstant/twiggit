package main

import (
	"fmt"
	"os"

	"github.com/amaury/twiggit/cmd"
	"github.com/amaury/twiggit/internal/di"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := getRootCommand()
	carapace.Gen(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getRootCommand() *cobra.Command {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create dependency container
	container := di.NewContainer(cfg)

	return cmd.NewRootCmd(container)
}
