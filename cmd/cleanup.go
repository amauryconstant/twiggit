// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/services"
	"github.com/spf13/cobra"
)

// NewCleanupCmd creates the cleanup command
func NewCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove unused or stale worktrees",
		Long: `Clean up worktrees that are no longer needed.

This command helps identify and remove:
- Worktrees with merged branches
- Stale worktrees that haven't been used recently
- Worktrees pointing to deleted branches

Safety features:
- Interactive confirmation before deletion
- Backup option for uncommitted changes
- Dry-run mode to preview actions

Examples:
  twiggit cleanup
  twiggit cleanup --dry-run
  twiggit cleanup --force --merged-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanupCommand(cmd, args)
		},
	}

	// Add flags specific to cleanup
	cmd.Flags().Bool("dry-run", false, "Show what would be cleaned up without actually doing it")
	cmd.Flags().Bool("force", false, "Skip interactive confirmation")
	cmd.Flags().Bool("merged-only", false, "Only clean up worktrees with merged branches")
	cmd.Flags().Duration("older-than", 0, "Only clean up worktrees older than specified duration (e.g., 30d, 1w)")

	return cmd
}

// runCleanupCommand implements the cleanup command functionality
func runCleanupCommand(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	mergedOnly, _ := cmd.Flags().GetBool("merged-only")
	olderThan, _ := cmd.Flags().GetDuration("older-than")

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

	// Filter worktrees based on criteria
	cutoffTime := time.Now().Add(-olderThan)

	// Pre-allocate slice with reasonable capacity
	candidates := make([]*domain.Worktree, 0, len(worktrees)/2) // Estimate half will be candidates

	for _, wt := range worktrees {
		// Skip if worktree is too recent (if older-than is specified)
		if olderThan > 0 && wt.LastUpdated.After(cutoffTime) {
			continue
		}

		// Skip if not merged-only and worktree is clean (simplified logic)
		if mergedOnly && wt.Status == domain.StatusClean {
			continue
		}

		// For now, consider all worktrees as candidates (simplified implementation)
		// In a full implementation, we would check:
		// - If branches are merged
		// - If branches still exist remotely
		// - If worktrees are actually stale
		candidates = append(candidates, wt)
	}

	if len(candidates) == 0 {
		fmt.Println("No worktrees found matching cleanup criteria")
		return nil
	}

	// Sort candidates by path
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Path < candidates[j].Path
	})

	// Display candidates
	fmt.Printf("Found %d worktree(s) to clean up:\n", len(candidates))
	fmt.Println()

	for i, wt := range candidates {
		// Make path relative to workspace for cleaner display
		relPath := wt.Path
		if strings.HasPrefix(wt.Path, workspacePath) {
			relPath = strings.TrimPrefix(wt.Path, workspacePath)
			relPath = strings.TrimPrefix(relPath, "/")
		}

		fmt.Printf("%d. %s\n", i+1, relPath)
		fmt.Printf("   Branch: %s\n", wt.Branch)
		fmt.Printf("   Status: %s\n", wt.Status)
		fmt.Printf("   Last updated: %s\n", formatCleanupTimeAgo(wt.LastUpdated))
		fmt.Println()
	}

	// Confirm before proceeding
	if !force {
		fmt.Print("Proceed with cleanup? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cleanup cancelled")
			return nil
		}
	}

	// Perform cleanup
	if dryRun {
		fmt.Println("DRY RUN: No worktrees would be removed")
	} else {
		fmt.Println("Cleanup functionality would remove worktrees here")
		fmt.Println("(Full implementation would use git worktree remove)")
	}

	return nil
}

// formatCleanupTimeAgo formats a time as a human-readable "time ago" string
func formatCleanupTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}
