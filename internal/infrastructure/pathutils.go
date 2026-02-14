package infrastructure

import (
	"path/filepath"
	"strings"

	"twiggit/internal/domain"
)

// ExtractProjectFromWorktreePath extracts the project name from a worktree path.
// Worktree paths follow the pattern: {worktreesDir}/{projectName}/{branchName}/...
// Returns empty string if the path is not under the worktrees directory.
func ExtractProjectFromWorktreePath(worktreePath, worktreesDir string) (projectName string, err error) {
	cleanedWorktreesDir := filepath.Clean(worktreesDir)
	if !strings.HasPrefix(worktreePath, cleanedWorktreesDir+string(filepath.Separator)) {
		return "", nil
	}

	relPath, err := filepath.Rel(cleanedWorktreesDir, worktreePath)
	if err != nil {
		return "", domain.NewContextDetectionError(worktreePath, "failed to get relative path from worktrees directory", err)
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 1 {
		return "", nil
	}

	return parts[0], nil
}

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
// Resolves symlinks to prevent symlink-based path traversal attacks
func IsPathUnder(base, target string) (bool, error) {
	// Handle edge cases for empty paths
	if base == "" && target == "" {
		return true, nil
	}
	if base == "" {
		return false, domain.NewContextDetectionError(target, "base path cannot be empty", nil)
	}
	if target == "" {
		return false, domain.NewContextDetectionError(base, "target path cannot be empty", nil)
	}

	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		resolvedBase = base
	}

	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		resolvedTarget = target
	}

	absBase, err := filepath.Abs(resolvedBase)
	if err != nil {
		return false, domain.NewContextDetectionError(target, "failed to get absolute path for base", err)
	}

	absTarget, err := filepath.Abs(resolvedTarget)
	if err != nil {
		return false, domain.NewContextDetectionError(target, "failed to get absolute path for target", err)
	}

	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false, domain.NewContextDetectionError(target, "failed to get relative path", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
