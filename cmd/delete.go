// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amaury/twiggit/internal/di"
	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/spf13/cobra"
)

// DeleteScope defines the scope for worktree deletion
type DeleteScope struct {
	Project        string // Specific project name
	AllProjects    bool   // All projects in workspace
	ExcludeCurrent bool   // Exclude current worktree
	CurrentPath    string // Path of current worktree (if any)
}

// NewDeleteCmd creates the delete command
func NewDeleteCmd(container *di.Container) *cobra.Command {
	var keepBranch bool
	var mergedOnly bool
	var changeDir bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Git worktrees",
		Long: `Delete Git worktrees from workspace directory.

This command intelligently detects worktrees based on your current location
and will NEVER delete main repositories in ~/Projects/.

Safety features:
- Interactive confirmation before deletion
- Protection of main repositories
- Protection of current worktree
- Prevention of deletion with uncommitted changes

Examples:
  twiggit delete              # Delete worktrees in current scope
  twiggit delete --dry-run    # Preview what would be deleted
  twiggit delete --force      # Skip confirmation and safety checks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(cmd, args, container, keepBranch, mergedOnly, changeDir)
		},
	}

	// Add flags
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")
	cmd.Flags().Bool("force", false, "Skip interactive confirmation and safety checks")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed deletion process")
	cmd.Flags().Bool("keep-branch", false, "Keep branch after removing worktree")
	cmd.Flags().Bool("merged-only", false, "Only delete worktrees for merged branches")
	cmd.Flags().BoolP("change-dir", "C", false, "Change to main project directory after deletion")

	return cmd
}

// runDeleteCommand implements the delete command functionality
func runDeleteCommand(cmd *cobra.Command, _ []string, container *di.Container, keepBranch, mergedOnly, changeDir bool) error {
	ctx := context.Background()

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get services from container
	discoveryService := container.DiscoveryService()

	// Determine scope based on current location (same logic as list command)
	workspacePath, scope, err := determineDeleteScope(ctx, container)
	if err != nil {
		return fmt.Errorf("failed to determine scope: %w", err)
	}

	// Discover worktrees
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to discover worktrees: %w", err)
	}

	// Filter worktrees based on scope and safety rules
	candidates := filterWorktreesForDeletion(worktrees, scope)

	if len(candidates) == 0 {
		fmt.Println("No worktrees found to delete")
		return nil
	}

	// Show candidates and get confirmation
	if !force {
		if err := showDeletionConfirmation(candidates, workspacePath); err != nil {
			return err
		}
	}

	// Perform deletion
	return performDeletion(ctx, container, candidates, dryRun, force, verbose, keepBranch, mergedOnly, changeDir)
}

// determineDeleteScope determines the deletion scope based on current location
func determineDeleteScope(ctx context.Context, container *di.Container) (string, DeleteScope, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", DeleteScope{}, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if we're in a git repository
	repoRoot, err := container.GitClient().GetRepositoryRoot(ctx, currentDir)
	if err != nil {
		// Not in git repo - scope is all projects
		return container.Config().WorkspacesPath, DeleteScope{AllProjects: true}, nil
	}

	// We're in a git repository - use configured workspaces path
	// and determine project scope from current location
	projectName := filepath.Base(repoRoot)

	// Check if current directory is within the workspaces path
	if strings.HasPrefix(currentDir, container.Config().WorkspacesPath) {
		// We're in a worktree - scope is current project
		return container.Config().WorkspacesPath, DeleteScope{Project: projectName}, nil
	}

	// We're in a main repository or other location - scope is all projects
	return container.Config().WorkspacesPath, DeleteScope{AllProjects: true}, nil
}

// filterWorktreesForDeletion applies safety rules and scope filtering
func filterWorktreesForDeletion(worktrees []*domain.Worktree, scope DeleteScope) []*domain.Worktree {
	candidates := make([]*domain.Worktree, 0, len(worktrees))

	for _, wt := range worktrees {
		// Safety Rule 1: Never delete main repositories in ~/Projects/
		if isMainRepositoryPath(wt.Path) {
			continue
		}

		// Safety Rule 2: Never delete current worktree
		if scope.ExcludeCurrent && wt.Path == scope.CurrentPath {
			continue
		}

		// Scope filtering
		if scope.Project != "" {
			// Check if worktree belongs to specified project
			if !strings.Contains(wt.Path, scope.Project) {
				continue
			}
		}

		candidates = append(candidates, wt)
	}

	return candidates
}

// isMainRepositoryPath checks if path is under ~/Projects/
func isMainRepositoryPath(path string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	projectsDir := filepath.Join(homeDir, "Projects")
	cleanPath := filepath.Clean(path)
	return strings.HasPrefix(cleanPath, projectsDir)
}

// showDeletionConfirmation displays candidates and gets user confirmation
func showDeletionConfirmation(candidates []*domain.Worktree, workspacePath string) error {
	fmt.Printf("Found %d worktree(s) to delete:\n", len(candidates))
	fmt.Println()

	// Sort candidates by path for consistent display
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Path < candidates[j].Path
	})

	for i, wt := range candidates {
		relPath := makeRelativePath(wt.Path, workspacePath)
		fmt.Printf("%d. %s\n", i+1, relPath)
		fmt.Printf("   Branch: %s\n", wt.Branch)
		fmt.Printf("   Status: %s\n", wt.Status)
		fmt.Printf("   Last updated: %s\n", formatTimeAgo(wt.LastUpdated))
		fmt.Println()
	}

	fmt.Print("Proceed with deletion? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled")
		return errors.New("deletion cancelled by user")
	}

	return nil
}

// performDeletion executes the worktree deletion
func performDeletion(ctx context.Context, container *di.Container, candidates []*domain.Worktree, dryRun, force, verbose, keepBranch, mergedOnly, changeDir bool) error {
	if dryRun {
		fmt.Println("DRY RUN: No worktrees would be deleted")
		return nil
	}

	var failed []string
	var success int
	worktreeRemover := container.WorktreeRemover()
	gitClient := container.GitClient()

	for _, wt := range candidates {
		if verbose {
			fmt.Printf("Deleting worktree: %s\n", wt.Path)
		}

		// Check merged-only flag
		if mergedOnly {
			shouldSkip, err := shouldSkipDueToMergeStatus(ctx, gitClient, wt, verbose)
			if err != nil {
				failed = append(failed, wt.Path)
				continue
			}
			if shouldSkip {
				continue
			}
		}

		err := worktreeRemover.Remove(ctx, wt.Path, force, keepBranch)
		if err != nil {
			if verbose {
				fmt.Printf("Failed to delete %s: %v\n", wt.Path, err)
			}
			failed = append(failed, wt.Path)
		} else {
			if verbose {
				fmt.Printf("Successfully deleted: %s\n", wt.Path)
			}
			success++
		}
	}

	return finalizeDeletion(ctx, gitClient, candidates, failed, success, changeDir)
}

// shouldSkipDueToMergeStatus checks if a worktree should be skipped due to merged-only flag
func shouldSkipDueToMergeStatus(ctx context.Context, gitClient infrastructure.GitClient, wt *domain.Worktree, verbose bool) (bool, error) {
	repoRoot, err := gitClient.GetRepositoryRoot(ctx, wt.Path)
	if err != nil {
		if verbose {
			fmt.Printf("Failed to get repository root for %s: %v\n", wt.Path, err)
		}
		return true, fmt.Errorf("failed to get repository root: %w", err)
	}

	isMerged, err := gitClient.IsBranchMerged(ctx, repoRoot, wt.Branch)
	if err != nil {
		if verbose {
			fmt.Printf("Failed to check if branch %s is merged: %v\n", wt.Branch, err)
		}
		return true, fmt.Errorf("failed to check merge status: %w", err)
	}

	if !isMerged {
		if verbose {
			fmt.Printf("Skipping %s: branch %s is not merged\n", wt.Path, wt.Branch)
		}
		return true, nil
	}

	return false, nil
}

// finalizeDeletion prints summary and handles change-dir functionality
func finalizeDeletion(ctx context.Context, gitClient infrastructure.GitClient, candidates []*domain.Worktree, failed []string, success int, changeDir bool) error {
	// Summary
	fmt.Printf("Deletion complete: %d succeeded, %d failed\n", success, len(failed))
	if len(failed) > 0 {
		return fmt.Errorf("%d worktrees failed to delete", len(failed))
	}

	// Change to main project directory if requested and deletion was successful
	if changeDir && success > 0 {
		// Get the repository root of the first successfully deleted worktree
		for _, wt := range candidates {
			repoRoot, err := gitClient.GetRepositoryRoot(ctx, wt.Path)
			if err == nil {
				fmt.Printf("%s\n", repoRoot)
				break
			}
		}
	}

	return nil
}

// makeRelativePath makes a path relative to workspace for cleaner display
func makeRelativePath(path, workspacePath string) string {
	relPath := path
	if strings.HasPrefix(path, workspacePath) {
		relPath = strings.TrimPrefix(path, workspacePath)
		relPath = strings.TrimPrefix(relPath, "/")
	}
	return relPath
}
