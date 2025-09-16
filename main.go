package main

import (
	"fmt"
	"os"

	"github.com/amaury/twiggit/cmd"
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
	return cmd.NewRootCmd()
}
