package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"twiggit/internal/domain"
)

// OutputFormatter defines the interface for formatting worktree output
type OutputFormatter interface {
	FormatWorktrees(worktrees []*domain.WorktreeInfo) string
}

// TextFormatter implements text-based output formatting
type TextFormatter struct{}

// FormatWorktrees formats worktrees as human-readable text
func (f *TextFormatter) FormatWorktrees(worktrees []*domain.WorktreeInfo) string {
	if len(worktrees) == 0 {
		return "No worktrees found"
	}

	var result strings.Builder
	for _, wt := range worktrees {
		status := ""
		if wt.Modified {
			status += " (modified)"
		}
		if wt.IsDetached {
			status += " (detached)"
		}

		result.WriteString(fmt.Sprintf("%s -> %s%s\n", wt.Branch, wt.Path, status))
	}
	return result.String()
}

// JSONFormatter implements JSON output formatting
type JSONFormatter struct{}

// FormatWorktrees formats worktrees as compact JSON
func (f *JSONFormatter) FormatWorktrees(worktrees []*domain.WorktreeInfo) string {
	// Convert domain types to JSON-serializable types
	worktreeList := WorktreeListJSON{
		Worktrees: make([]WorktreeJSON, len(worktrees)),
	}

	for i, wt := range worktrees {
		worktreeList.Worktrees[i] = WorktreeJSON{
			Branch: wt.Branch,
			Path:   wt.Path,
			Status: getStatus(wt),
		}
	}

	// Marshal to JSON with compact formatting
	data, err := json.Marshal(worktreeList)
	if err != nil {
		// Fallback to error JSON if marshaling fails
		return `{"error": "failed to marshal worktrees to JSON"}`
	}

	return string(data)
}

// WorktreeJSON represents a worktree for JSON serialization
type WorktreeJSON struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

// WorktreeListJSON is the wrapper struct for JSON output
type WorktreeListJSON struct {
	Worktrees []WorktreeJSON `json:"worktrees"`
}

// getStatus converts WorktreeInfo to a status string
func getStatus(wt *domain.WorktreeInfo) string {
	if wt.IsDetached {
		return "detached"
	}
	if wt.Modified {
		return "modified"
	}
	return "clean"
}
