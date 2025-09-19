// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewSwitchCmd creates the switch command
func NewSwitchCmd() *cobra.Command {
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
			return runSwitchCommand(cmd, args)
		},
	}

	return cmd
}

// runSwitchCommand implements the switch command functionality
func runSwitchCommand(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Try to use current context for intelligent switching
		project, err := detectCurrentContext()
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
		project, err := detectCurrentContext()
		if err == nil && project != "" && target != project {
			// User said "switch feature-branch" while in project context
			// But only if target is not the same as project name
			return switchToWorktree(fmt.Sprintf("%s/%s", project, target))
		}
	}

	// Original logic
	if strings.Contains(target, "/") {
		return switchToWorktree(target)
	}

	return switchToProject(target)
}

// switchToProject switches to a main project repository
func switchToProject(project string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetPath := filepath.Join(homeDir, "Projects", project)

	if err := validatePathExists(targetPath); err != nil {
		return err
	}

	return changeDirectory(targetPath)
}

// switchToWorktree switches to a specific worktree
func switchToWorktree(target string) error {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return errors.New("invalid worktree format: use <project>/<branch>")
	}

	project, branch := parts[0], parts[1]

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetPath := filepath.Join(homeDir, "Workspaces", project, branch)

	if err := validatePathExists(targetPath); err != nil {
		return err
	}

	return changeDirectory(targetPath)
}

// detectCurrentContext detects the current project context from the working directory
func detectCurrentContext() (project string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check if we're in a worktree: ~/Workspaces/project/branch/
	workspacesDir := filepath.Join(homeDir, "Workspaces")
	if strings.HasPrefix(currentDir, workspacesDir) {
		relPath := strings.TrimPrefix(currentDir, workspacesDir)
		parts := strings.Split(strings.TrimPrefix(relPath, string(filepath.Separator)), string(filepath.Separator))
		if len(parts) >= 2 && parts[0] != "" {
			return parts[0], nil
		}
	}

	// Check if we're in a project: ~/Projects/project/
	projectsDir := filepath.Join(homeDir, "Projects")
	if strings.HasPrefix(currentDir, projectsDir) {
		relPath := strings.TrimPrefix(currentDir, projectsDir)
		parts := strings.Split(strings.TrimPrefix(relPath, string(filepath.Separator)), string(filepath.Separator))
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0], nil
		}
	}

	return "", nil
}

// validatePathExists validates that a path exists and returns appropriate errors
func validatePathExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if strings.Contains(path, "Workspaces") {
			return fmt.Errorf("worktree not found: %s", path)
		}
		return fmt.Errorf("project not found: %s", path)
	}
	return nil
}

// changeDirectory changes the current working directory
func changeDirectory(path string) error {
	// For now, print the cd command since Go can't change the parent shell's directory
	// Users can use: eval "$(twiggit switch target)"
	fmt.Printf("cd %s\n", path)
	return nil
}
