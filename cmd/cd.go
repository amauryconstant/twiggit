// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/amaury/twiggit/internal/di"
	"github.com/amaury/twiggit/internal/domain"
	"github.com/spf13/cobra"
)

// NewCdCmd creates the cd command
func NewCdCmd(container *di.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cd <project|project/branch>",
		Short: "Change directory to a project or worktree",
		Long: `Change directory to a project repository or worktree.

Changes to the main project repository or a specific worktree branch.
Supports context-aware directory changes when called from within a project.

Examples:
  twiggit cd myproject              # Change to ~/Projects/myproject
  twiggit cd myproject/feature-branch # Change to ~/Workspaces/myproject/feature-branch
  twiggit cd feature-branch         # When in project context, changes to worktree`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCdCommand(cmd, args, container)
		},
	}

	// Add completion functionality
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeCdCommand(cmd, args, toComplete, container)
	}

	return cmd
}

// runCdCommand implements the cd command functionality
func runCdCommand(_ *cobra.Command, args []string, container *di.Container) error {
	ctx := context.Background()

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return domain.NewWorktreeError(
			domain.ErrInvalidPath,
			"failed to get current directory",
			"",
			err,
		).WithSuggestion("Check current directory permissions")
	}

	// Create context detector and resolver
	config := container.Config()
	contextDetector := domain.NewContextDetector(config.WorkspacesPath, config.ProjectsPath)
	contextResolver := domain.NewContextResolver(config.WorkspacesPath, config.ProjectsPath)

	// Detect current context
	currentContext, err := contextDetector.Detect(currentDir)
	if err != nil {
		return domain.NewWorktreeError(
			domain.ErrValidation,
			"failed to detect current context",
			"",
			err,
		).WithSuggestion("Ensure you are in a valid directory")
	}

	if len(args) == 0 {
		// Show context-aware help when no target specified
		return showContextAwareHelp(currentContext, container)
	}

	target := args[0]

	// Resolve target based on current context
	resolution, err := contextResolver.Resolve(target, currentContext)
	if err != nil {
		return fmt.Errorf("failed to resolve target '%s': %w", target, err)
	}

	// Validate target exists before navigation
	if err := validateTargetExists(ctx, container, resolution); err != nil {
		return err
	}

	// Change to the resolved target path
	return changeDirectory(resolution.TargetPath)
}

// showContextAwareHelp displays context-aware help when no target is specified
func showContextAwareHelp(context *domain.Context, _ *di.Container) error {
	switch context.Type {
	case domain.ContextProject:
		fmt.Printf("Current project: %s\n", context.ProjectName)
		fmt.Printf("Available targets:\n")
		fmt.Printf("  twiggit cd %s          # stay in current project\n", context.ProjectName)
		fmt.Printf("  twiggit cd main        # stay in current project (alias)\n")
		fmt.Printf("  twiggit cd <branch>    # navigate to worktree of current project\n")
		fmt.Printf("  twiggit cd <project>   # navigate to different project\n")
		fmt.Printf("  twiggit cd <project>/<branch> # navigate to cross-project worktree\n")

	case domain.ContextWorktree:
		fmt.Printf("Current worktree: %s/%s\n", context.ProjectName, context.BranchName)
		fmt.Printf("Available targets:\n")
		fmt.Printf("  twiggit cd main        # navigate to main project directory\n")
		fmt.Printf("  twiggit cd %s          # navigate to main project directory\n", context.ProjectName)
		fmt.Printf("  twiggit cd <branch>    # navigate to different worktree of same project\n")
		fmt.Printf("  twiggit cd <project>   # navigate to different project\n")
		fmt.Printf("  twiggit cd <project>/<branch> # navigate to cross-project worktree\n")

	case domain.ContextOutsideGit:
		fmt.Printf("Current context: Outside git repository\n")
		fmt.Printf("Available targets:\n")
		fmt.Printf("  twiggit cd <project>   # navigate to project main directory\n")
		fmt.Printf("  twiggit cd <project>/<branch> # navigate to cross-project worktree\n")

	case domain.ContextUnknown:
		fmt.Printf("Current context: Unknown\n")
		fmt.Printf("Available targets:\n")
		fmt.Printf("  twiggit cd <project>   # navigate to project main directory\n")
		fmt.Printf("  twiggit cd <project>/<branch> # navigate to cross-project worktree\n")
	}

	return nil
}

// validateTargetExists validates that the resolved target actually exists
func validateTargetExists(_ context.Context, _ *di.Container, resolution *domain.ContextResolution) error {
	// Check if the target path exists
	if _, err := os.Stat(resolution.TargetPath); os.IsNotExist(err) {
		switch resolution.TargetType {
		case "project":
			return domain.NewWorkspaceError(
				domain.ErrWorkspaceProjectNotFound,
				fmt.Sprintf("project '%s' not found at %s", resolution.ProjectName, resolution.TargetPath),
				err,
			).WithSuggestion("Check project name spelling").
				WithSuggestion("Verify project exists in projects directory").
				WithSuggestion("Use 'twiggit list' to see available projects")

		case "worktree":
			return domain.NewWorkspaceError(
				domain.ErrWorkspaceWorktreeNotFound,
				fmt.Sprintf("worktree '%s/%s' not found at %s", resolution.ProjectName, resolution.BranchName, resolution.TargetPath),
				err,
			).WithSuggestion("Check worktree name spelling").
				WithSuggestion("Verify worktree exists for this project").
				WithSuggestion("Use 'twiggit list' to see available worktrees")

		default:
			return domain.NewWorkspaceError(
				domain.ErrWorkspaceDiscoveryFailed,
				"target not found at "+resolution.TargetPath,
				err,
			).WithSuggestion("Check target name spelling").
				WithSuggestion("Verify target exists")
		}
	}

	return nil
}

// changeDirectory changes the current working directory
func changeDirectory(path string) error {
	// Print only the target path for shell wrapper consumption
	fmt.Printf("%s\n", path)
	return nil
}

// completeCdCommand provides completion for the cd command
func completeCdCommand(_ *cobra.Command, args []string, toComplete string, container *di.Container) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()

	// If we already have one argument, we're completing a branch name for a project
	if len(args) == 1 {
		return completeBranches(ctx, args[0], toComplete, container)
	}

	// If no arguments yet, complete project names and project/branch combinations
	return completeProjectsAndWorktrees(ctx, toComplete, container)
}

// completeProjectsAndWorktrees completes project names and project/branch combinations
func completeProjectsAndWorktrees(ctx context.Context, toComplete string, container *di.Container) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// Get available projects
	discoveryService := container.DiscoveryService()
	projects, err := discoveryService.DiscoverProjects(ctx, container.Config().ProjectsPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get available worktrees
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, container.Config().WorkspacesPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Filter projects based on toComplete
	for _, project := range projects {
		if strings.HasPrefix(project.Name, toComplete) {
			completions = append(completions, project.Name)
		}
	}

	// Filter worktrees based on toComplete (format: project/branch)
	for _, worktree := range worktrees {
		// Extract project name from worktree path
		pathParts := strings.Split(worktree.Path, "/")
		if len(pathParts) >= 2 {
			projectName := pathParts[len(pathParts)-2]
			worktreeName := fmt.Sprintf("%s/%s", projectName, worktree.Branch)
			if strings.HasPrefix(worktreeName, toComplete) {
				completions = append(completions, worktreeName)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeBranches completes branch names for a given project
func completeBranches(ctx context.Context, project string, toComplete string, container *di.Container) ([]string, cobra.ShellCompDirective) {
	var completions []string

	// Get available worktrees
	discoveryService := container.DiscoveryService()
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, container.Config().WorkspacesPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Filter worktrees for the specified project and branch
	for _, worktree := range worktrees {
		// Extract project name from worktree path
		pathParts := strings.Split(worktree.Path, "/")
		if len(pathParts) >= 2 {
			projectName := pathParts[len(pathParts)-2]
			if projectName == project && strings.HasPrefix(worktree.Branch, toComplete) {
				completions = append(completions, worktree.Branch)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
