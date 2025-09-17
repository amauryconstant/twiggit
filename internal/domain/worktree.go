package domain

import (
	"fmt"
	"os"
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
	// Commit is the current commit hash
	Commit string
	// Status represents the current working tree status
	Status WorktreeStatus
	// LastUpdated is when the status was last refreshed
	LastUpdated time.Time
	// Metadata stores additional worktree information
	Metadata map[string]string
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
		Commit:      "",
		Status:      StatusUnknown,
		LastUpdated: time.Now(),
		Metadata:    make(map[string]string),
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

// SetCommit sets the commit hash for the worktree
func (w *Worktree) SetCommit(commit string) error {
	w.Commit = commit
	return nil
}

// GetCommit returns the commit hash for the worktree
func (w *Worktree) GetCommit() string {
	return w.Commit
}

// ValidatePathExists checks if the worktree path exists on the filesystem
func (w *Worktree) ValidatePathExists() (bool, error) {
	if _, err := os.Stat(w.Path); os.IsNotExist(err) {
		return false, fmt.Errorf("worktree path does not exist: %s", w.Path)
	}
	return true, nil
}

// IsStatusStale checks if the worktree status is stale (older than 5 minutes)
func (w *Worktree) IsStatusStale() bool {
	return time.Since(w.LastUpdated) > 5*time.Minute
}

// IsStatusStaleWithThreshold checks if the worktree status is stale with custom threshold
func (w *Worktree) IsStatusStaleWithThreshold(threshold time.Duration) bool {
	return time.Since(w.LastUpdated) > threshold
}

// Equals checks if two worktrees are equal (all fields match)
func (w *Worktree) Equals(other *Worktree) bool {
	if other == nil {
		return false
	}
	return w.Path == other.Path &&
		w.Branch == other.Branch &&
		w.Commit == other.Commit &&
		w.Status == other.Status
}

// SameLocationAs checks if two worktrees are at the same location (path only)
func (w *Worktree) SameLocationAs(other *Worktree) bool {
	if other == nil {
		return false
	}
	return w.Path == other.Path
}

// SetMetadata sets a metadata key-value pair
func (w *Worktree) SetMetadata(key, value string) {
	w.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key
func (w *Worktree) GetMetadata(key string) (string, bool) {
	value, exists := w.Metadata[key]
	return value, exists
}

// WorktreeHealth represents the health status of a worktree
type WorktreeHealth struct {
	Status string
	Issues []string
}

// GetHealth returns the health status of the worktree
func (w *Worktree) GetHealth() *WorktreeHealth {
	health := &WorktreeHealth{
		Status: "unknown",
		Issues: make([]string, 0),
	}

	// Check if path exists
	if exists, err := w.ValidatePathExists(); err != nil || !exists {
		health.Issues = append(health.Issues, "path not validated")
	}

	// Determine overall status
	if len(health.Issues) == 0 {
		health.Status = "healthy"
	} else {
		health.Status = "unhealthy"
	}

	return health
}
