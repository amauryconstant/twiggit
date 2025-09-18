// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/services"
	"github.com/spf13/cobra"
)

// NewSwitchCmd creates the switch command
func NewSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [worktree-path]",
		Short: "Switch to an existing worktree",
		Long: `Switch to an existing Git worktree.

If no path is provided, an interactive selection will be presented.
This command changes your shell working directory to the selected worktree.

Examples:
  twiggit switch /path/to/worktree
  twiggit switch  # Interactive mode`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSwitchCommand(cmd, args)
		},
	}

	return cmd
}

// runSwitchCommand implements the switch command functionality
func runSwitchCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create git client
	gitClient := git.NewClient()

	// Create discovery service
	discoveryService := services.NewDiscoveryService(gitClient)

	// Determine workspace path
	workspacePath := cfg.Workspace
	if currentDir, err := os.Getwd(); err == nil {
		if repoRoot, err := gitClient.GetRepositoryRoot(ctx, currentDir); err == nil {
			// We're in a git repository, use its parent as workspace
			workspacePath = filepath.Dir(repoRoot)
		}
	}

	// Discover worktrees
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to discover worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Printf("No worktrees found in %s\n", workspacePath)
		return nil
	}

	// Get current directory to highlight current worktree
	currentDir, _ := os.Getwd()

	// If a specific path was provided, try to find and switch to it
	if len(args) > 0 {
		targetPath := args[0]
		for _, wt := range worktrees {
			if strings.HasSuffix(wt.Path, targetPath) || wt.Path == targetPath {
				fmt.Printf("Switch to worktree: cd %s\n", wt.Path)
				fmt.Printf("Branch: %s\n", wt.Branch)
				fmt.Printf("Status: %s\n", wt.Status)
				return nil
			}
		}
		return fmt.Errorf("worktree not found: %s", targetPath)
	}

	// Interactive mode: list all worktrees with navigation hints
	fmt.Printf("Available worktrees in %s:\n", workspacePath)
	fmt.Println()

	for i, wt := range worktrees {
		// Make path relative to workspace for cleaner display
		relPath := wt.Path
		if strings.HasPrefix(wt.Path, workspacePath) {
			relPath = strings.TrimPrefix(wt.Path, workspacePath)
			if strings.HasPrefix(relPath, "/") {
				relPath = relPath[1:]
			}
		}

		// Check if this is the current worktree
		isCurrent := strings.HasPrefix(currentDir, wt.Path)
		currentMarker := ""
		if isCurrent {
			currentMarker = " (current)"
		}

		fmt.Printf("%d. %s%s\n", i+1, relPath, currentMarker)
		fmt.Printf("   Branch: %s\n", wt.Branch)
		fmt.Printf("   Status: %s\n", wt.Status)
		fmt.Printf("   Switch: cd %s\n", wt.Path)
		fmt.Println()
	}

	fmt.Println("Use: twiggit switch <worktree-path> to switch directly")
	fmt.Println("Or: cd <path> to navigate to a worktree")

	return nil
}
