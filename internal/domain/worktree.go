// Package domain contains core entities for git worktree management.
package domain

import ()

// Worktree represents a git worktree with basic validation
type Worktree struct {
	path   string
	branch string
}

// NewWorktree creates a new worktree with validation
func NewWorktree(path, branch string) (*Worktree, error) {
	if path == "" {
		return nil, NewValidationError("NewWorktree", "path", "", "cannot be empty")
	}
	if branch == "" {
		return nil, NewValidationError("NewWorktree", "branch", "", "cannot be empty")
	}

	return &Worktree{
		path:   path,
		branch: branch,
	}, nil
}

// Path returns the worktree filesystem path
func (w *Worktree) Path() string {
	return w.path
}

// Branch returns the worktree branch name
func (w *Worktree) Branch() string {
	return w.branch
}
