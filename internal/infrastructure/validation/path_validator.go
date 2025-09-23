// Package validation provides infrastructure-level validation services for path operations.
package validation

import (
	"path/filepath"
	"strings"

	"github.com/amaury/twiggit/internal/infrastructure"
)

// PathValidatorImpl implements PathValidator interface
type PathValidatorImpl struct {
	// Could inject filesystem abstraction here if needed for testing
}

// NewPathValidator creates a new PathValidator instance
func NewPathValidator() infrastructure.PathValidator {
	return &PathValidatorImpl{}
}

// IsValidGitRepoPath validates if a path is suitable for a git repository
func (pv *PathValidatorImpl) IsValidGitRepoPath(path string) bool {
	if path == "" {
		return false
	}

	// Basic path format validation
	if len(path) > 255 { // MaxPathLength from domain
		return false
	}

	// Check if it looks like an absolute path (starts with /)
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// Check for invalid characters
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return false
	}

	// Additional filesystem-specific validation
	return pv.isValidFilesystemPath(path)
}

// IsValidWorkspacePath validates if a path is suitable for a workspace
func (pv *PathValidatorImpl) IsValidWorkspacePath(path string) bool {
	if path == "" {
		return false
	}

	if len(path) > 255 { // MaxPathLength from domain
		return false
	}

	return pv.isValidFilesystemPath(path)
}

// isValidFilesystemPath performs basic filesystem path validation
func (pv *PathValidatorImpl) isValidFilesystemPath(path string) bool {
	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Check if the cleaned path is different (indicates path traversal attempts)
	if cleanPath != path {
		return false
	}

	// Check for empty path components
	if strings.Contains(cleanPath, "//") {
		return false
	}

	return true
}
