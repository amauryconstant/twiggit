package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
)

// VersionOptions captures every input to the version command.
type VersionOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (*git.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	AppVersion    string
}

// NewCmdVersion creates the version command.
//
// runF is the optional override used by tests; pass nil to install
// the default runVersion body.
func NewCmdVersion(f *cmdutil.Factory, runF func(*VersionOptions) error) *cobra.Command {
	opts := &VersionOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
		AppVersion:    f.AppVersion,
	}

	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Show version of twiggit",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if runF != nil {
				return runF(opts)
			}
			return runVersion(opts)
		},
	}

	return cmd
}

// runVersion writes the build-time version string to opts.IO.Out.
func runVersion(opts *VersionOptions) error {
	if opts == nil || opts.IO == nil {
		return nil
	}
	version := opts.AppVersion
	if version == "" {
		version = "dev"
	}
	_, err := fmt.Fprintf(opts.IO.Out, "twiggit %s\n", version)
	return err
}
