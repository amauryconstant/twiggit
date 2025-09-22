// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/amaury/twiggit/internal/di"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates the create command
func NewCreateCmd(container *di.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [branch-name]",
		Short: "Create a new Git worktree",
		Long: `Create a new Git worktree for the specified branch.

If no branch name is provided, an interactive selection will be presented.
The worktree will be created in the configured workspace directory under the project name.

Examples:
  twiggit create feature/new-auth
  twiggit create  # Interactive mode`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCommand(cmd, args, container)
		},
	}

	return cmd
}

// runCreateCommand implements the create command functionality
func runCreateCommand(_ *cobra.Command, args []string, container *di.Container) error {
	ctx := context.Background()

	// Determine branch name
	var branchName string
	if len(args) > 0 {
		branchName = args[0]
	} else {
		// Interactive mode - for now, return error
		return errors.New("interactive mode not yet implemented - please provide a branch name")
	}

	if branchName == "" {
		return errors.New("branch name is required")
	}

	// Try to find repository root from current directory first
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	mainRepoPath, err := container.GitClient().GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		return fmt.Errorf("failed to find repository root from current directory: %w", err)
	}

	// Get services from container
	worktreeCreator := container.WorktreeCreator()

	// Determine target path for worktree using project-aware logic
	targetPath := determineWorktreePath(mainRepoPath, branchName, container.Config().WorkspacesPath)

	// Check if branch exists for logging purposes
	branchExists := container.GitClient().BranchExists(ctx, mainRepoPath, branchName)
	if !branchExists {
		fmt.Printf("Branch '%s' does not exist, it will be created...\n", branchName)
	}

	// Create the worktree using service layer
	fmt.Printf("Creating worktree for branch '%s' at %s...\n", branchName, targetPath)
	err = worktreeCreator.Create(ctx, mainRepoPath, branchName, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Success message
	fmt.Printf("✅ Worktree created successfully!\n")
	if !branchExists {
		fmt.Printf("   Branch: %s (newly created)\n", branchName)
	} else {
		fmt.Printf("   Branch: %s\n", branchName)
	}
	fmt.Printf("   Path:   %s\n", targetPath)
	fmt.Printf("   Navigate: cd %s\n", targetPath)

	return nil
}

// extractProjectNameFromPath extracts the project name from the repository path
func extractProjectNameFromPath(repoPath string) string {
	// Get the base name of the repository directory
	return filepath.Base(repoPath)
}

// determineWorktreePath determines the correct worktree path based on repository location
func determineWorktreePath(repoPath, branchName, workspacesDir string) string {
	projectName := extractProjectNameFromPath(repoPath)
	return filepath.Join(workspacesDir, projectName, branchName)
}
