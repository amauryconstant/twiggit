package infrastructure

import (
	"path/filepath"
	"strings"

	"twiggit/internal/domain"
)

// NormalizePath normalizes a path for cross-platform compatibility
func NormalizePath(path string) (string, error) {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Convert to absolute path
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", domain.NewContextDetectionError(path, "path normalization failed", err)
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
		return false, domain.NewContextDetectionError(target, "failed to get relative path", err)
	}

	// Check if relative path starts with ".."
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
