// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/amaury/twiggit/internal/di"
	"github.com/amaury/twiggit/internal/domain"
	"github.com/spf13/cobra"
)

// NewListCmd creates the unified list command
func NewListCmd(container *di.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available worktrees",
		Long: `List Git worktrees with intelligent auto-detection.

When inside a Git repository, shows worktrees for that project only.
When outside a Git repository, shows worktrees for all projects.
Use --all to override and show all projects regardless of location.

Shows information about:
- Path and branch name
- Status (clean/dirty)
- Last commit information
- Last activity time
- Summary statistics

Examples:
  twiggit list              # Auto-detect scope based on current location
  twiggit list --all        # Show all projects' worktrees
  twiggit list --sort=date  # Sort by last updated time`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, args, container)
		},
	}

	// Add flags
	cmd.Flags().BoolP("all", "a", false, "Show worktrees from all projects")
	cmd.Flags().String("sort", "name", "Sort order (name, date, branch, status)")

	return cmd
}

// runListCommand implements the unified list command functionality
func runListCommand(cmd *cobra.Command, _ []string, container *di.Container) error {
	ctx := context.Background()

	// Get flags
	allFlag, _ := cmd.Flags().GetBool("all")
	sortBy, _ := cmd.Flags().GetString("sort")

	// Get services from container
	discoveryService := container.DiscoveryService()

	// Determine workspace path and scope
	workspacePath := container.Config().WorkspacesPath
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Auto-detect if we're in a git repository
	repoRoot, err := container.GitClient().GetRepositoryRoot(ctx, currentDir)
	inGitRepo := err == nil

	// Determine scope based on location and --all flag
	if !allFlag && inGitRepo {
		// Inside git repo without --all: show current project only
		workspacePath = filepath.Dir(repoRoot)
	}
	// Otherwise: show all projects (either --all flag or not in git repo)

	// Discover worktrees
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to discover worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		if inGitRepo && !allFlag {
			fmt.Printf("No worktrees found in current project\n")
		} else {
			fmt.Printf("No worktrees found in %s\n", workspacePath)
		}
		return nil
	}

	// Sort worktrees
	sortWorktrees(worktrees, sortBy)

	// Output table with all information
	return outputTable(worktrees, workspacePath)
}

// sortWorktrees sorts worktrees based on the specified criteria
func sortWorktrees(worktrees []*domain.Worktree, sortBy string) {
	sort.Slice(worktrees, func(i, j int) bool {
		switch sortBy {
		case "date":
			return worktrees[i].LastUpdated.After(worktrees[j].LastUpdated)
		case "branch":
			return worktrees[i].Branch < worktrees[j].Branch
		case "status":
			return worktrees[i].Status.String() < worktrees[j].Status.String()
		default: // name
			return worktrees[i].Path < worktrees[j].Path
		}
	})
}

// outputTable displays worktrees in a comprehensive table format
func outputTable(worktrees []*domain.Worktree, workspacePath string) error {
	fmt.Printf("Worktrees in %s:\n", workspacePath)
	fmt.Println()

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to flush output: %v\n", err)
		}
	}()

	// Print header
	if _, err := fmt.Fprintln(w, "Path\tBranch\tStatus\tLast Commit\tLast Updated"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "----\t------\t------\t-----------\t------------"); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Print worktree rows
	for _, wt := range worktrees {
		// Make path relative to workspace for cleaner display
		relPath := wt.Path
		if strings.HasPrefix(wt.Path, workspacePath) {
			relPath = strings.TrimPrefix(wt.Path, workspacePath)
			relPath = strings.TrimPrefix(relPath, "/")
		}

		// Format last commit (truncate hash for display)
		lastCommit := wt.Commit
		if len(lastCommit) > 7 {
			lastCommit = lastCommit[:7]
		}
		if lastCommit == "" {
			lastCommit = "unknown"
		}

		// Format last updated time
		timeAgo := formatTimeAgo(wt.LastUpdated)

		// Format status
		status := wt.Status.String()
		switch wt.Status {
		case domain.StatusDirty:
			status = "dirty"
		case domain.StatusClean:
			status = "clean"
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", relPath, wt.Branch, status, lastCommit, timeAgo); err != nil {
			return fmt.Errorf("failed to write worktree row: %w", err)
		}
	}

	// Add summary section
	fmt.Println()
	var cleanCount, dirtyCount int
	for _, wt := range worktrees {
		switch wt.Status {
		case domain.StatusClean:
			cleanCount++
		case domain.StatusDirty:
			dirtyCount++
		}
	}

	fmt.Printf("Summary: %d total, %d clean, %d dirty\n", len(worktrees), cleanCount, dirtyCount)

	return nil
}

// formatTimeAgo formats a time as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
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
		return t.Format("Jan 2")
	}
}
