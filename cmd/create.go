// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/amaury/twiggit/internal/domain"
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
The worktree will be created in the configured workspace directory.

Examples:
  twiggit create feature/new-auth
  twiggit create --from=main hotfix/critical-bug
  twiggit create  # Interactive mode`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCommand(cmd, args)
		},
	}

	// Add flags specific to create
	cmd.Flags().String("from", "", "Create worktree from specific branch or commit")
	cmd.Flags().String("template", "", "Use project template for worktree setup")
	cmd.Flags().Bool("open", false, "Open worktree in default IDE after creation")

	return cmd
}

// runCreateCommand implements the create command functionality
func runCreateCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override workspace if specified as flag
	if workspaceFlag, err := cmd.Flags().GetString("workspace"); err == nil && workspaceFlag != "" {
		cfg.Workspace = workspaceFlag
	}

	// Get flags
	fromBranch, _ := cmd.Flags().GetString("from")
	template, _ := cmd.Flags().GetString("template")
	openInIDE, _ := cmd.Flags().GetBool("open")

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

	// Validate source branch exists
	sourceBranch := fromBranch
	if sourceBranch == "" {
		sourceBranch = branchName
	}

	if !gitClient.BranchExists(ctx, mainRepoPath, sourceBranch) {
		return fmt.Errorf("branch '%s' does not exist in repository", sourceBranch)
	}

	// Determine target path for worktree
	targetPath := filepath.Join(cfg.Workspace, branchName)

	// Check if target path already exists
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Create the worktree
	fmt.Printf("Creating worktree for branch '%s' at %s...\n", branchName, targetPath)

	// For now, use native git command for worktree creation
	// TODO: Fix go-git implementation or use exec.Command for git operations
	err = createWorktreeNative(mainRepoPath, sourceBranch, targetPath)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Success message
	fmt.Printf("✅ Worktree created successfully!\n")
	fmt.Printf("   Branch: %s\n", branchName)
	fmt.Printf("   Path:   %s\n", targetPath)
	fmt.Printf("   Navigate: cd %s\n", targetPath)

	// Apply template if specified
	if template != "" {
		fmt.Printf("📋 Applying template '%s'...\n", template)
		// Template application would go here
		fmt.Printf("   Template application not yet implemented\n")
	}

	// Open in IDE if requested
	if openInIDE {
		fmt.Printf("🔧 Opening in default IDE...\n")
		// IDE opening would go here
		fmt.Printf("   IDE opening not yet implemented\n")
	}

	return nil
}

// findMainRepository finds the main git repository in the workspace
func findMainRepository(ctx context.Context, gitClient domain.GitClient, workspacePath string) (string, error) {
	// Check if workspace path exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace path does not exist: %s", workspacePath)
	}

	// Read workspace directory
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to read workspace directory: %w", err)
	}

	// Look for git repositories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := filepath.Join(workspacePath, entry.Name())

		// Check if it's a main repository (not a worktree)
		isMainRepo, err := gitClient.IsMainRepository(ctx, candidatePath)
		if err != nil {
			continue // Skip on error
		}

		if isMainRepo {
			return candidatePath, nil
		}
	}

	return "", fmt.Errorf("no main git repository found in workspace: %s", workspacePath)
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
