// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates the create command
func NewCreateCmd() *cobra.Command {
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
			return runCreateCommand(cmd, args)
		},
	}

	return cmd
}

// runCreateCommand implements the create command functionality
func runCreateCommand(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

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

	// Create git client
	gitClient := git.NewClient()

	// Try to find repository root from current directory first
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	mainRepoPath, err := gitClient.GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		return fmt.Errorf("failed to find repository root from current directory: %w", err)
	}

	// Check if branch exists, create it if it doesn't
	if !gitClient.BranchExists(ctx, mainRepoPath, branchName) {
		fmt.Printf("Branch '%s' does not exist, creating it...\n", branchName)
		err := createBranchNative(mainRepoPath, branchName)
		if err != nil {
			return fmt.Errorf("failed to create branch '%s': %w", branchName, err)
		}
		fmt.Printf("✅ Branch '%s' created successfully\n", branchName)
	}

	// Determine target path for worktree using project-aware logic
	targetPath := determineWorktreePath(mainRepoPath, branchName, cfg.WorkspacesPath)

	// Check if target path already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Create the worktree
	fmt.Printf("Creating worktree for branch '%s' at %s...\n", branchName, targetPath)

	// For now, use native git command for worktree creation
	// TODO: Fix go-git implementation or use exec.Command for git operations
	err = createWorktreeNative(mainRepoPath, branchName, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Success message
	fmt.Printf("✅ Worktree created successfully!\n")
	fmt.Printf("   Branch: %s\n", branchName)
	fmt.Printf("   Path:   %s\n", targetPath)
	fmt.Printf("   Navigate: cd %s\n", targetPath)

	return nil
}

// createWorktreeNative creates a worktree using the native git command
func createWorktreeNative(repoPath, branch, targetPath string) error {
	// For now, use os/exec to run git worktree add command
	// This is a temporary solution until the go-git implementation is fixed
	cmd := exec.Command("git", "worktree", "add", targetPath, branch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
	}

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

// createBranchNative creates a new branch using the native git command without switching to it
func createBranchNative(repoPath, branchName string) error {
	// Create the new branch from the current branch without switching to it
	cmd := exec.Command("git", "branch", branchName)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
