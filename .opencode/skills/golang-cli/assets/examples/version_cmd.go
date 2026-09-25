// Path: cmd/version.go
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/output"
	"github.com/you/myapp/internal/version"
)

func NewCmdVersion(f *cmdutil.Factory) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cmdutil.UsageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := version.Get()
			if format != "" {
				formatter, err := output.NewFormatter(format)
				if err != nil {
					return err
				}
				return formatter.Write(f.IOStreams.Out, info) // json; Info is not Tabular
			}
			details := []string{info.Go}
			if info.Commit != "" {
				details = append(details, "commit "+info.Commit)
			}
			if info.Date != "" {
				details = append(details, "built "+info.Date)
			}
			fmt.Fprintf(f.IOStreams.Out, "myapp %s (%s)\n", info.Version, strings.Join(details, ", "))
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "output", "o", "", "output format: json")
	return cmd
}
