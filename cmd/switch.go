// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/services"
	"github.com/spf13/cobra"
)

// NewSwitchCmd creates the switch command
func NewSwitchCmd(deps *infrastructure.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <project|project/branch>",
		Short: "Switch to a project or worktree",
		Long: `Switch to a project repository or worktree.

Switches to the main project repository or a specific worktree branch.
Supports context-aware switching when called from within a project.

Examples:
  twiggit switch myproject              # Switch to ~/Projects/myproject
  twiggit switch myproject/feature-branch # Switch to ~/Workspaces/myproject/feature-branch
  twiggit switch feature-branch         # When in project context, switches to worktree`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwitchCommand(cmd, args, deps)
		},
	}

	return cmd
}

// runSwitchCommand implements the switch command functionality
func runSwitchCommand(_ *cobra.Command, args []string, deps *infrastructure.Deps) error {
	ctx := context.Background()

	if len(args) == 0 {
		// Try to use current context for intelligent switching
		project, err := detectCurrentContext(ctx, deps)
		if err != nil {
			return errors.New("specify a target: <project> or <project/branch>")
		}

		if project != "" {
			fmt.Printf("Current project: %s\n", project)
			fmt.Printf("Available targets:\n")
			fmt.Printf("  twiggit switch %s          # main repository\n", project)
			fmt.Printf("  twiggit switch %s/<branch> # worktree\n", project)
			return nil
		}

		return errors.New("specify a target: <project> or <project/branch>")
	}

	target := args[0]

	// Handle relative branch names when in project context
	if !strings.Contains(target, "/") {
		project, err := detectCurrentContext(ctx, deps)
		if err == nil && project != "" && target != project {
			// User said "switch feature-branch" while in project context
			// But only if target is not the same as project name
			return switchToWorktree(ctx, deps, fmt.Sprintf("%s/%s", project, target))
		}
	}

	// Original logic
	if strings.Contains(target, "/") {
		return switchToWorktree(ctx, deps, target)
	}

	return switchToProject(ctx, deps, target)
}

// switchToProject switches to a main project repository
func switchToProject(ctx context.Context, deps *infrastructure.Deps, project string) error {
	// Use discovery service to find projects
	discoveryService := services.NewDiscoveryService(deps)
	projects, err := discoveryService.DiscoverProjects(ctx, deps.Config.ProjectsPath)
	if err != nil {
		return fmt.Errorf("failed to discover projects: %w", err)
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
		return fmt.Errorf("project not found: %s", project)
	}

	return changeDirectory(targetPath)
}

// switchToWorktree switches to a specific worktree
func switchToWorktree(ctx context.Context, deps *infrastructure.Deps, target string) error {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return errors.New("invalid worktree format: use <project>/<branch>")
	}

	project, branch := parts[0], parts[1]

	// Use discovery service to find worktrees
	discoveryService := services.NewDiscoveryService(deps)
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, deps.Config.WorkspacesPath)
	if err != nil {
		return fmt.Errorf("failed to discover worktrees: %w", err)
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
		return fmt.Errorf("worktree not found: %s/%s", project, branch)
	}

	return changeDirectory(targetPath)
}

// detectCurrentContext detects the current project context from the working directory
func detectCurrentContext(ctx context.Context, deps *infrastructure.Deps) (project string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if we're in a git repository
	repoRoot, err := deps.GitClient.GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		return "", nil // Not in git repo, no context
	}

	// Extract project name from repository root
	projectName := filepath.Base(repoRoot)
	return projectName, nil
}

// changeDirectory changes the current working directory
func changeDirectory(path string) error {
	// For now, print the cd command since Go can't change the parent shell's directory
	// Users can use: eval "$(twiggit switch target)"
	fmt.Printf("cd %s\n", path)
	return nil
}
