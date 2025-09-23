package domain

import (
	"errors"
	"fmt"
	"time"
)

// Constants
const (
	// MaxPathLength is the maximum allowed length for worktree paths
	MaxPathLength = 255
)

// Types

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

// WorktreeHealth represents the health status of a worktree
type WorktreeHealth struct {
	Status string
	Issues []string
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

// Constructors

// NewWorktree creates a new Worktree instance with validation
func NewWorktree(path, branch string) (*Worktree, error) {
	return NewWorktreeAt(path, branch, time.Now())
}

// NewWorktreeAt creates a new Worktree instance with deterministic timestamp
func NewWorktreeAt(path, branch string, timestamp time.Time) (*Worktree, error) {
	if err := ValidatePathFormat(path); err != nil {
		return nil, err
	}
	if branch == "" {
		return nil, errors.New("branch name cannot be empty")
	}

	return &Worktree{
		Path:        path,
		Branch:      branch,
		Commit:      "",
		Status:      StatusUnknown,
		LastUpdated: timestamp,
		Metadata:    make(map[string]string),
	}, nil
}

// Methods

// String returns a human-readable representation of the worktree
func (w *Worktree) String() string {
	return fmt.Sprintf("Worktree{path=%s, branch=%s, status=%s}", w.Path, w.Branch, w.Status)
}

// UpdateStatus updates the worktree status and refresh timestamp
func (w *Worktree) UpdateStatus(status WorktreeStatus) error {
	w.UpdateStatusAt(status, time.Now())
	return nil
}

// UpdateStatusAt updates the worktree status with deterministic timestamp
func (w *Worktree) UpdateStatusAt(status WorktreeStatus, timestamp time.Time) {
	w.Status = status
	w.LastUpdated = timestamp
}

// IsClean returns true if the worktree has no uncommitted changes
func (w *Worktree) IsClean() bool {
	return w.Status == StatusClean
}

// IsStatusStale checks if the worktree status is stale (older than 5 minutes)
func (w *Worktree) IsStatusStale() bool {
	return time.Since(w.LastUpdated) > 5*time.Minute
}

// IsStatusStaleWithThreshold checks if the worktree status is stale with custom threshold
func (w *Worktree) IsStatusStaleWithThreshold(threshold time.Duration) bool {
	return time.Since(w.LastUpdated) > threshold
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

// SetMetadata sets a metadata key-value pair
func (w *Worktree) SetMetadata(key, value string) {
	w.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key
func (w *Worktree) GetMetadata(key string) (string, bool) {
	value, exists := w.Metadata[key]
	return value, exists
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

// GetHealth returns the health status of the worktree
func (w *Worktree) GetHealth() *WorktreeHealth {
	health := &WorktreeHealth{
		Status: "unknown",
		Issues: make([]string, 0),
	}

	// Basic validation - path format check
	if err := ValidatePathFormat(w.Path); err != nil {
		health.Issues = append(health.Issues, "path not validated")
	}

	// Validate branch name
	if w.Branch == "" {
		health.Issues = append(health.Issues, "branch name is empty")
	}

	// Determine overall status
	if len(health.Issues) == 0 {
		health.Status = "healthy"
	} else {
		health.Status = "unhealthy"
	}

	return health
}

// Pure Functions

// ValidatePathFormat validates the format of a worktree path (pure business logic)
func ValidatePathFormat(path string) error {
	if path == "" {
		return errors.New("worktree path cannot be empty")
	}
	if len(path) > MaxPathLength {
		return fmt.Errorf("path too long: %d characters (max %d)", len(path), MaxPathLength)
	}

	// Basic path format validation - in a real implementation this would be more sophisticated
	// For now, we accept most paths as long as they're not empty and not too long
	return nil
}

// String returns a human-readable representation of status
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

// WorktreeInfo represents information about a Git worktree (used by infrastructure layer)
type WorktreeInfo struct {
	// Path is the filesystem path to the worktree
	Path string
	// Branch is the branch name checked out in the worktree
	Branch string
	// Commit is the current commit hash
	Commit string
	// Clean indicates if the worktree has no uncommitted changes
	Clean bool
	// CommitTime is the timestamp when the commit was created
	CommitTime time.Time
}

// Validate checks if the WorktreeInfo is valid
func (w *WorktreeInfo) Validate() error {
	if w.Path == "" {
		return errors.New("path cannot be empty")
	}
	if w.Branch == "" {
		return errors.New("branch cannot be empty")
	}
	return nil
}
