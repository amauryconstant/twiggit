package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

// NewVersionCommand creates and returns the version command
func NewVersionCommand(_ *CommandConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of twiggit",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("twiggit %s (%s) %s\n", version, commit, date)
		},
	}
}
