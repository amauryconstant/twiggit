// Path: cmd/create.go — one command = Options struct + constructor + run function
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/core"
	"github.com/you/myapp/internal/iostreams"
)

// worktreeCreator is defined here, where it is consumed, so tests pass a fake.
type worktreeCreator interface {
	CreateWorktree(ctx context.Context, name core.BranchName, source string) (*core.Worktree, error)
}

// CreateOptions holds everything runCreate needs: I/O, lazy dependencies, and
// parsed input. Tests build it directly.
type CreateOptions struct {
	IO     *iostreams.IOStreams
	Client func() (worktreeCreator, error)
	Ctx    context.Context

	Name   string
	Source string
}

// NewCmdCreate wires flags and RunE. runF is the test seam: when non-nil it
// receives the parsed options instead of runCreate.
func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{}
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a worktree",
		Args:  cmdutil.UsageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = cmd.Context()
			opts.Name = args[0]
			opts.Client = func() (worktreeCreator, error) {
				c, err := f.Client()
				if err != nil {
					return nil, err
				}
				return c, nil
			}
			if runF != nil {
				return runF(opts)
			}
			return runCreate(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Source, "source", "s", "main", "branch to start from")
	return cmd
}

// runCreate is Parse → Execute → Respond for one command.
func runCreate(opts *CreateOptions) error {
	name, err := core.NewBranchName(opts.Name) // validation lives in the core
	if err != nil {
		return err
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}
	opts.IO.Verbosef("creating %s from %s", name, opts.Source)
	wt, err := client.CreateWorktree(opts.Ctx, name, opts.Source)
	if err != nil {
		return err
	}

	if opts.IO.IsStdoutTTY() {
		s := opts.IO.Styles()
		fmt.Fprintf(opts.IO.Out, "Created %s at %s\n", s.Success(wt.Name), s.Dim(wt.Path))
		return nil
	}
	fmt.Fprintln(opts.IO.Out, wt.Path) // piped: bare data for the next program
	return nil
}
