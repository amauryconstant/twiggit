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

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available worktrees",
		Long: `List all Git worktrees in the current project or workspace.

Shows detailed information about each worktree including:
- Path and branch name
- Status (clean/dirty)
- Last commit information
- Creation date

Examples:
  twiggit list
  twiggit list --all-projects
  twiggit list --format=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, args)
		},
	}

	// Add flags specific to list
	cmd.Flags().Bool("all-projects", false, "List worktrees from all projects in workspace")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml, simple)")
	cmd.Flags().String("sort", "name", "Sort order (name, date, branch, status)")

	return cmd
}

// runListCommand implements the list command functionality
func runListCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get flags
	allProjects, _ := cmd.Flags().GetBool("all-projects")
	format, _ := cmd.Flags().GetString("format")
	sortBy, _ := cmd.Flags().GetString("sort")

	// Create git client
	gitClient := git.NewClient()

	// Create discovery service
	discoveryService := services.NewDiscoveryService(gitClient)

	// Determine workspace path
	workspacePath := cfg.Workspace
	if !allProjects {
		// If not listing all projects, try to detect current project
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

	// Sort worktrees
	sortWorktrees(worktrees, sortBy)

	// Format output
	switch format {
	case "json":
		return outputJSON(worktrees)
	case "yaml":
		return outputYAML(worktrees)
	case "simple":
		return outputSimple(worktrees)
	default:
		return outputTable(worktrees, workspacePath)
	}
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

// outputTable displays worktrees in a table format
func outputTable(worktrees []*domain.Worktree, workspacePath string) error {
	if len(worktrees) == 0 {
		fmt.Printf("No worktrees found in %s\n", workspacePath)
		return nil
	}

	fmt.Printf("Worktrees in %s:\n", workspacePath)

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "Path\tBranch\tStatus\tLast Updated")
	fmt.Fprintln(w, "----\t------\t------\t------------")

	// Print worktree rows
	for _, wt := range worktrees {
		// Make path relative to workspace for cleaner display
		relPath := wt.Path
		if strings.HasPrefix(wt.Path, workspacePath) {
			relPath = strings.TrimPrefix(wt.Path, workspacePath)
			if strings.HasPrefix(relPath, "/") {
				relPath = relPath[1:]
			}
		}

		// Format last updated time
		timeAgo := formatTimeAgo(wt.LastUpdated)

		// Format status with color indicators
		status := wt.Status.String()
		switch wt.Status {
		case domain.StatusDirty:
			status = "dirty"
		case domain.StatusClean:
			status = "clean"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", relPath, wt.Branch, status, timeAgo)
	}

	return nil
}

// outputSimple displays worktrees in a simple format
func outputSimple(worktrees []*domain.Worktree) error {
	for _, wt := range worktrees {
		fmt.Printf("%s (%s) - %s\n", wt.Path, wt.Branch, wt.Status.String())
	}
	return nil
}

// outputJSON displays worktrees in JSON format
func outputJSON(worktrees []*domain.Worktree) error {
	fmt.Println("JSON output not yet implemented")
	return nil
}

// outputYAML displays worktrees in YAML format
func outputYAML(worktrees []*domain.Worktree) error {
	fmt.Println("YAML output not yet implemented")
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
