package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/core"
	"twiggit/internal/git"
)

// NewCDCommand creates a new cd command
func NewCDCommand(f *CommandConfig) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			return executeCD(cmd, f, target)
		},
	}

	carapace.Gen(cmd).PositionalCompletion(actionWorktreeTarget(f))

	return cmd
}

// executeCD implements the orchestration previously in navigationService.ResolvePath
// and navigationService.ValidatePath. It composes the context
// detector and resolver directly.
func executeCD(cmd *cobra.Command, f *CommandConfig, target string) error {
	_ = context.Background()

	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	gitClient, err := f.GitClient()
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

	logv(cmd, 1, "Navigating to worktree")
	logv(cmd, 2, "  target: %s", target)
	logv(cmd, 2, "  worktree path: %s", result.ResolvedPath)

	if validateErr := validatePath(result.ResolvedPath); validateErr != nil {
		if result.Type == core.PathTypeWorktree {
			return &core.OperationError{
				Op:      "cd.worktree",
				Entity:  target,
				Message: "worktree not found",
				Cause:   validateErr,
			}
		}
		return &core.OperationError{
			Op:      "cd.worktree",
			Entity:  target,
			Message: "project not found",
			Cause:   validateErr,
		}
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), result.ResolvedPath); err != nil {
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
