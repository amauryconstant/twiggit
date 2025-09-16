package domain

import (
	"fmt"
	"time"
)

// WorktreeStatus represents the current state of a worktree
type WorktreeStatus int

const (
	// StatusUnknown indicates the worktree status hasn't been determined
	StatusUnknown WorktreeStatus = iota
	// StatusClean indicates the worktree has no uncommitted changes
	StatusClean
	// StatusDirty indicates the worktree has uncommitted changes
	StatusDirty
)

// String returns a human-readable representation of the status
func (s WorktreeStatus) String() string {
	switch s {
	case StatusUnknown:
		return "unknown"
	case StatusClean:
		return "clean"
	case StatusDirty:
		return "dirty"
	default:
		return "invalid"
	}
}

// Worktree represents a Git worktree entity
type Worktree struct {
	// Path is the filesystem path to the worktree
	Path string
	// Branch is the Git branch currently checked out
	Branch string
	// Status represents the current working tree status
	Status WorktreeStatus
	// LastUpdated is when the status was last refreshed
	LastUpdated time.Time
}

// NewWorktree creates a new Worktree instance with validation
func NewWorktree(path, branch string) (*Worktree, error) {
	if path == "" {
		return nil, fmt.Errorf("worktree path cannot be empty")
	}
	if branch == "" {
		return nil, fmt.Errorf("branch name cannot be empty")
	}

	return &Worktree{
		Path:        path,
		Branch:      branch,
		Status:      StatusUnknown,
		LastUpdated: time.Now(),
	}, nil
}

// UpdateStatus updates the worktree status and refresh timestamp
func (w *Worktree) UpdateStatus(status WorktreeStatus) error {
	w.Status = status
	w.LastUpdated = time.Now()
	return nil
}

// IsClean returns true if the worktree has no uncommitted changes
func (w *Worktree) IsClean() bool {
	return w.Status == StatusClean
}

// String returns a human-readable representation of the worktree
func (w *Worktree) String() string {
	return fmt.Sprintf("Worktree{path=%s, branch=%s, status=%s}", w.Path, w.Branch, w.Status)
}
