// Path: cmd/list.go — a query command: human output by default, --output for scripts
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/core"
	"github.com/you/myapp/internal/iostreams"
	"github.com/you/myapp/internal/output"
)

type worktreeLister interface {
	ListWorktrees(ctx context.Context) ([]core.Worktree, error)
}

type ListOptions struct {
	IO     *iostreams.IOStreams
	Client func() (worktreeLister, error)
	Ctx    context.Context

	Format string // --output: json | table | plain; empty = human default
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List worktrees",
		Args:  cmdutil.UsageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.IO = f.IOStreams
			opts.Ctx = cmd.Context()
			opts.Client = func() (worktreeLister, error) {
				c, err := f.Client()
				if err != nil {
					return nil, err
				}
				return c, nil
			}
			if runF != nil {
				return runF(opts)
			}
			return runList(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Format, "output", "o", "", "output format: json, table, plain")
	return cmd
}

func runList(opts *ListOptions) error {
	// Resolve the formatter first: a bad --output is a usage error before any work.
	var formatter output.Formatter
	if opts.Format != "" {
		var err error
		if formatter, err = output.NewFormatter(opts.Format); err != nil {
			return err
		}
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}
	items, err := client.ListWorktrees(opts.Ctx)
	if err != nil {
		return err
	}

	if formatter != nil {
		return formatter.Write(opts.IO.Out, worktreeTable(items))
	}
	for _, wt := range items {
		fmt.Fprintf(opts.IO.Out, "%s\t%s\n", opts.IO.Styles().Bold(wt.Name), wt.Path)
	}
	return nil
}

// worktreeTable adapts core values to output.Tabular; presentation stays in cmd/.
type worktreeTable []core.Worktree

func (t worktreeTable) Header() []string { return []string{"NAME", "BRANCH", "PATH"} }

func (t worktreeTable) Rows() [][]string {
	rows := make([][]string, 0, len(t))
	for _, wt := range t {
		rows = append(rows, []string{wt.Name, wt.Branch, wt.Path})
	}
	return rows
}
