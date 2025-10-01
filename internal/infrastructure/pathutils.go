package infrastructure

import (
	"fmt"
	"path/filepath"
	"strings"
)

// NormalizePath normalizes a path for cross-platform compatibility
func NormalizePath(path string) (string, error) {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Convert to absolute path
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("failed to normalize path: %w", err)
	}

	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// If symlink resolution fails, use absolute path
		return abs, nil
	}

	return resolved, nil
}

// IsPathUnder checks if target is under base directory
func IsPathUnder(base, target string) (bool, error) {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false, fmt.Errorf("failed to get relative path from %s to %s: %w", base, target, err)
	}

	// Check if relative path starts with ".."
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
