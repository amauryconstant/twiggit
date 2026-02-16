package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"twiggit/internal/domain"
)

// NewCreateCommand creates a new create command
func NewCreateCommand(config *CommandConfig) *cobra.Command {
	var source string
	var cdFlag bool

	cmd := &cobra.Command{
		Use:   "create <project>/<branch> | <branch>",
		Short: "Create a new worktree",
		Long: `Create a new worktree for the specified project and branch.
If only a branch name is provided, the project is inferred from the current context.

Flags:
  --source <branch>  Source branch to create from (default: main)
  -C, --cd          Output worktree path to stdout (for shell wrapper)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCreate(cmd, config, args[0], source, cdFlag)
		},
	}

	cmd.Flags().StringVar(&source, "source", "main", "Source branch to create from")
	cmd.Flags().BoolVarP(&cdFlag, "cd", "C", false, "Output worktree path to stdout (for shell wrapper)")

	// Silence usage to prevent double error printing
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(config),
	)

	carapace.Gen(cmd).FlagCompletion(map[string]carapace.Action{
		"source": actionBranches(config),
	})

	return cmd
}

// executeCreate executes the create command with the given configuration
func executeCreate(cmd *cobra.Command, config *CommandConfig, spec, source string, cdFlag bool) error {
	ctx := context.Background()

	// Extract branch name for validation first (before any context detection)
	branchName := extractBranchNameForValidation(spec)

	// Validate branch name first (before any context detection or project discovery)
	branchValidation := domain.ValidateBranchName(branchName)
	if branchValidation.IsError() {
		return branchValidation.Error
	}

	// Now detect current context (after branch validation passes)
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	// Parse project and branch from spec with proper context
	projectName, branchName, err := parseProjectBranch(spec, currentCtx)
	if err != nil {
		return err
	}

	// Discover project after branch validation
	project, err := config.Services.ProjectService.DiscoverProject(ctx, projectName, currentCtx)
	if err != nil {
		return fmt.Errorf("failed to discover project %s: %w", projectName, err)
	}

	// Validate source branch exists before creating worktree
	sourceBranchExists, err := config.Services.WorktreeService.BranchExists(ctx, project.Path, source)
	if err != nil {
		return fmt.Errorf("failed to check if source branch '%s' exists: %w", source, err)
	}
	if !sourceBranchExists {
		return fmt.Errorf("source branch '%s' does not exist", source)
	}

	// Create worktree request
	req := &domain.CreateWorktreeRequest{
		ProjectName:  project.Name,
		BranchName:   branchName,
		SourceBranch: source,
		Context:      currentCtx,
		Force:        false,
	}

	logv(cmd, 1, "Creating worktree for %s/%s", project.Name, branchName)
	logv(cmd, 2, "  from branch: %s", source)
	logv(cmd, 2, "  to path: %s", project.Name+"/"+branchName)
	logv(cmd, 2, "  in repo: %s", project.GitRepoPath)
	logv(cmd, 2, "  creating parent dir: %s", project.Path+"/"+branchName)

	// Create worktree
	worktree, err := config.Services.WorktreeService.CreateWorktree(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	logv(cmd, 2, "  created worktree at: %s", worktree.Path)

	// Display output based on cdFlag
	if cdFlag {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), worktree.Path)
	} else {
		if err := displayCreateSuccess(cmd.OutOrStdout(), worktree); err != nil {
			return err
		}
	}

	return nil
}

// parseProjectBranch parses the project/branch specification
func parseProjectBranch(spec string, ctx *domain.Context) (string, string, error) {
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", errors.New("invalid format: expected <project>/<branch>")
		}
		projectName := parts[0]

		if validation := domain.ValidateProjectName(projectName); validation.IsError() {
			return "", "", validation.Error
		}

		return projectName, parts[1], nil
	}

	if ctx.ProjectName != "" {
		return ctx.ProjectName, spec, nil
	}

	return "", "", errors.New("cannot infer project: not in a project context and no project specified")
}

// extractBranchNameForValidation extracts branch name from spec for validation
func extractBranchNameForValidation(spec string) string {
	// Check if spec contains a slash (project/branch format)
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		if len(parts) == 2 && parts[1] != "" {
			return parts[1]
		}
	}
	// If no slash, spec itself is the branch name
	return spec
}

// displayCreateSuccess displays the success message for worktree creation
func displayCreateSuccess(out io.Writer, worktree *domain.WorktreeInfo) error {
	_, err := fmt.Fprintf(out, "Created worktree: %s -> %s\n", worktree.Branch, worktree.Path)
	if err != nil {
		return fmt.Errorf("failed to display success message: %w", err)
	}
	return nil
}
