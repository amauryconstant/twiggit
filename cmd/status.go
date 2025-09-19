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

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/services"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of all worktrees",
		Long: `Display a comprehensive status overview of all worktrees.

Shows information about:
- Branch name and status (clean/dirty)
- Last commit information
- Uncommitted changes
- Creation date and last activity

Examples:
  twiggit status
  twiggit status --global
  twiggit status --project myproject`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusCommand(cmd, args)
		},
	}

	// Add flags specific to status
	cmd.Flags().Bool("global", false, "Show status across all projects in workspace")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")
	cmd.Flags().Bool("dirty-only", false, "Show only worktrees with uncommitted changes")

	return cmd
}

// runStatusCommand implements the status command functionality
func runStatusCommand(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get flags
	global, _ := cmd.Flags().GetBool("global")
	format, _ := cmd.Flags().GetString("format")
	dirtyOnly, _ := cmd.Flags().GetBool("dirty-only")

	// Create git client
	gitClient := git.NewClient()

	// Create discovery service
	discoveryService := services.NewDiscoveryService(gitClient)

	// Determine workspace path
	workspacePath := cfg.Workspace
	if !global {
		// If not global, try to detect current project
		if currentDir, err := os.Getwd(); err == nil {
			if repoRoot, err := gitClient.GetRepositoryRoot(ctx, currentDir); err == nil {
				// We're in a git repository, use its parent as workspace
				workspacePath = filepath.Dir(repoRoot)
			}
		}
	}

	// Discover worktrees
	worktrees, err := discoveryService.DiscoverWorktrees(ctx, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to discover worktrees: %w", err)
	}

	// Filter worktrees if dirty-only flag is set
	if dirtyOnly {
		var filteredWorktrees []*domain.Worktree
		for _, wt := range worktrees {
			if wt.Status == domain.StatusDirty {
				filteredWorktrees = append(filteredWorktrees, wt)
			}
		}
		worktrees = filteredWorktrees
	}

	if len(worktrees) == 0 {
		if dirtyOnly {
			fmt.Println("No dirty worktrees found")
		} else {
			fmt.Printf("No worktrees found in %s\n", workspacePath)
		}
		return nil
	}

	// Sort worktrees by path
	sort.Slice(worktrees, func(i, j int) bool {
		return worktrees[i].Path < worktrees[j].Path
	})

	// Format output
	switch format {
	case "json":
		return outputStatusJSON(worktrees)
	case "yaml":
		return outputStatusYAML(worktrees)
	default:
		return outputStatusTable(worktrees, workspacePath)
	}
}

// outputStatusTable displays worktree status in a detailed table format
func outputStatusTable(worktrees []*domain.Worktree, workspacePath string) error {
	fmt.Printf("Worktree Status in %s:\n", workspacePath)
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
		timeAgo := formatStatusTimeAgo(wt.LastUpdated)

		// Format status with color indicators
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

// outputStatusJSON displays worktree status in JSON format
func outputStatusJSON(_ []*domain.Worktree) error {
	fmt.Println("JSON output not yet implemented")
	return nil
}

// outputStatusYAML displays worktree status in YAML format
func outputStatusYAML(_ []*domain.Worktree) error {
	fmt.Println("YAML output not yet implemented")
	return nil
}

// formatStatusTimeAgo formats a time as a human-readable "time ago" string
func formatStatusTimeAgo(t time.Time) string {
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
