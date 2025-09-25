// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

	if len(args) == 0 {
		// Try to use current context for intelligent switching
		project, err := detectCurrentContext(ctx, container)
		if err != nil {
			return domain.NewWorktreeError(
				domain.ErrValidation,
				"no target specified",
				"",
			).WithSuggestion("Provide a target in the format 'project' or 'project/branch'")
		}

		if project != "" {
			fmt.Printf("Current project: %s\n", project)
			fmt.Printf("Available targets:\n")
			fmt.Printf("  twiggit cd %s          # main repository\n", project)
			fmt.Printf("  twiggit cd %s/<branch> # worktree\n", project)
			return nil
		}

		return domain.NewWorktreeError(
			domain.ErrValidation,
			"no target specified",
			"",
		).WithSuggestion("Provide a target in the format 'project' or 'project/branch'")
	}

	target := args[0]

	// Handle relative branch names when in project context
	if !strings.Contains(target, "/") {
		project, err := detectCurrentContext(ctx, container)
		if err == nil && project != "" && target != project {
			// User said "cd feature-branch" while in project context
			// But only if target is not the same as project name
			return cdToWorktree(ctx, container, fmt.Sprintf("%s/%s", project, target))
		}
	}

	// Original logic
	if strings.Contains(target, "/") {
		return cdToWorktree(ctx, container, target)
	}

	return cdToProject(ctx, container, target)
}

// cdToProject changes directory to a main project repository
func cdToProject(ctx context.Context, container *di.Container, project string) error {
	// Get services from container
	discoveryService := container.DiscoveryService()
	projects, err := discoveryService.DiscoverProjects(ctx, container.Config().ProjectsPath)
	if err != nil {
		return domain.NewWorkspaceError(
			domain.ErrWorkspaceProjectNotFound,
			fmt.Sprintf("project '%s' not found", project),
			err,
		).WithSuggestion("Check project name spelling").
			WithSuggestion("Verify project exists in projects directory").
			WithSuggestion("Use 'twiggit list' to see available projects")
	}

	// Find the target project
	var targetPath string
	for _, p := range projects {
		if p.Name == project {
			targetPath = p.GitRepo
			break
		}
	}

	if targetPath == "" {
		return domain.NewWorkspaceError(
			domain.ErrWorkspaceProjectNotFound,
			fmt.Sprintf("project '%s' not found", project),
		).WithSuggestion("Check project name spelling").
			WithSuggestion("Verify project exists in projects directory").
			WithSuggestion("Use 'twiggit list' to see available projects")
	}

	return changeDirectory(targetPath)
}

// cdToWorktree changes directory to a specific worktree
func cdToWorktree(ctx context.Context, container *di.Container, target string) error {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return domain.NewWorktreeError(
			domain.ErrValidation,
			"invalid worktree format",
			"",
		).WithSuggestion("Use the format 'project/branch' to specify a worktree")
	}

	project, branch := parts[0], parts[1]

	// Get services from container
	discoveryService := container.DiscoveryService()
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, container.Config().WorkspacesPath)
	if err != nil {
		return domain.NewWorkspaceError(
			domain.ErrWorkspaceWorktreeNotFound,
			fmt.Sprintf("worktree '%s' not found", target),
			err,
		).WithSuggestion("Check worktree name spelling").
			WithSuggestion("Verify worktree exists for this project").
			WithSuggestion("Use 'twiggit list' to see available worktrees")
	}

	// Find the target worktree
	var targetPath string
	for _, wt := range worktrees {
		// Check if worktree matches project and branch
		if strings.Contains(wt.Path, project) && wt.Branch == branch {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		return domain.NewWorkspaceError(
			domain.ErrWorkspaceWorktreeNotFound,
			fmt.Sprintf("worktree '%s' not found", target),
		).WithSuggestion("Check worktree name spelling").
			WithSuggestion("Verify worktree exists for this project").
			WithSuggestion("Use 'twiggit list' to see available worktrees")
	}

	return changeDirectory(targetPath)
}

// detectCurrentContext detects the current project context from the working directory
func detectCurrentContext(ctx context.Context, container *di.Container) (project string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", domain.NewWorktreeError(
			domain.ErrInvalidPath,
			"failed to get current directory",
			"",
			err,
		).WithSuggestion("Check current directory permissions")
	}

	// Check if we're in a git repository
	repoRoot, err := container.GitClient().GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		return "", nil // Not in git repo, no context
	}

	// Extract project name from repository root
	projectName := filepath.Base(repoRoot)
	return projectName, nil
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
