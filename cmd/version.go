package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"twiggit/internal/version"
)

// NewVersionCommand creates and returns the version command
func NewVersionCommand(_ *CommandConfig) *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Show version of twiggit",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("twiggit %s\n", version.String())
			return nil
		},
	}
}
