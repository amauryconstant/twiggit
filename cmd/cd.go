package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// CdOptions captures every input to runCd.
type CdOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (*git.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions

	// Per-command field.
	Target string
}

// NewCmdCd creates a new cd command.
//
// runF is the optional override used by tests; pass nil to install
// the default runCd body.
func NewCmdCd(f *cmdutil.Factory, runF func(*CdOptions) error) *cobra.Command {
	opts := &CdOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "cd <project|project/branch>",
		Short: "Change directory to a worktree",
		Long: `Change directory to the specified worktree.
If no target is provided, changes to the default worktree for the current project.
The command outputs the path to be used by shell integration.

Examples:
  twiggit cd                    # Change to default worktree for current project
  twiggit cd myproject          # Change to main worktree of myproject
  twiggit cd myproject/feature  # Change to feature branch worktree
  twiggit cd feature            # Change to feature branch (relative to current project)`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.MaximumNArgs(1)),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Target = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runCd(opts)
		},
	}

	carapace.Gen(cmd).PositionalCompletion(actionWorktreeTarget(f))

	return cmd
}

// runCd implements the orchestration previously in
// navigationService.ResolvePath and navigationService.ValidatePath.
// It composes the context detector and resolver directly.
func runCd(opts *CdOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	gitClient, err := opts.GitClient()
	if err != nil {
		return fmt.Errorf("git client init failed: %w", err)
	}

	detector, err := git.NewContextDetector(cfg)
	if err != nil {
		return fmt.Errorf("context detector init failed: %w", err)
	}
	resolver := git.NewContextResolver(cfg, gitClient, gitClient)

	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	currentCtx, err := detector.DetectContext(wd)
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	target := opts.Target

	// Context-aware default target.
	if target == "" {
		switch currentCtx.Type {
		case core.ContextWorktree:
			target = currentCtx.BranchName
		case core.ContextProject:
			target = "main"
		default:
			return errors.New("no target specified and no default worktree in context")
		}
	}

	result, err := resolver.ResolveIdentifier(currentCtx, target)
	if err != nil {
		return fmt.Errorf("failed to resolve path for %s: %w", target, err)
	}

	verbosef(opts.IO, "Navigating to worktree")
	verbosef(opts.IO, "target: %s", target)
	verbosef(opts.IO, "worktree path: %s", result.ResolvedPath)

	if validateErr := validatePath(result.ResolvedPath); validateErr != nil {
		if result.Type == core.PathTypeWorktree {
			return &core.OperationError{
				Op:      "cd.worktree",
				Entity:  target,
				Message: fmt.Sprintf("worktree not found for target '%s'", target),
				Cause:   validateErr,
			}
		}
		return &core.OperationError{
			Op:      "cd.worktree",
			Entity:  target,
			Message: fmt.Sprintf("project not found for target '%s'", target),
			Cause:   validateErr,
		}
	}

	if _, err := fmt.Fprintln(opts.IO.Out, result.ResolvedPath); err != nil {
		return fmt.Errorf("failed to output path: %w", err)
	}
	return nil
}

// validatePath checks that path exists and is a directory.
func validatePath(path string) error {
	if path == "" {
		return core.NewOpValidationError("ValidatePath", "path", "", "cannot be empty")
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return &core.OperationError{
			Op:      "validate.path",
			Entity:  path,
			Message: "path does not exist",
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return &core.OperationError{
			Op:      "validate.path",
			Entity:  path,
			Message: "failed to access path",
			Cause:   err,
		}
	}
	if !info.IsDir() {
		return &core.OperationError{
			Op:      "validate.path",
			Entity:  path,
			Message: "path is not a directory",
		}
	}
	return nil
}
