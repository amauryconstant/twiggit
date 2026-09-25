package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExtractProjectFromWorktreePath returns the project segment of
// worktreePath when it lives under worktreesDir (e.g. <worktreesDir>/<project>/<branch>).
// Returns an empty string and nil error when the path is not under
// worktreesDir, or a wrapped error when filepath.Rel fails.
func ExtractProjectFromWorktreePath(worktreePath, worktreesDir string) (string, error) {
	cleanedWorktreesDir := filepath.Clean(worktreesDir)
	if !strings.HasPrefix(worktreePath, cleanedWorktreesDir+string(filepath.Separator)) {
		return "", nil
	}

	relPath, err := filepath.Rel(cleanedWorktreesDir, worktreePath)
	if err != nil {
		return "", fmt.Errorf("extract project from worktree path %q under %q: %w", worktreePath, cleanedWorktreesDir, err)
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 1 {
		return "", nil
	}

	return parts[0], nil
}

// NormalizePath cleans path, resolves it to an absolute path, and
// follows symlinks when possible. When symlink resolution fails it
// falls back to the absolute form and returns nil error so callers
// see a usable path even on broken-symlink scenarios.
func NormalizePath(path string) (string, error) {
	cleaned := filepath.Clean(path)

	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("normalize path %q: abs: %w", path, err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		//nolint:nilerr // EvalSymlinks can fail on broken symlinks; callers get a usable absolute path instead of an error.
		return abs, nil
	}

	return resolved, nil
}

// IsPathUnder reports whether target is the same path as, or sits
// beneath, base after symlink and absolute-path normalisation. Empty
// base or target yields an error so callers can distinguish "outside
// the tree" from "misconfigured input".
func IsPathUnder(base, target string) (bool, error) {
	if base == "" && target == "" {
		return true, nil
	}
	if base == "" {
		return false, fmt.Errorf("is path under: base path cannot be empty (target=%q)", target)
	}
	if target == "" {
		return false, fmt.Errorf("is path under: target path cannot be empty (base=%q)", base)
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
		return false, fmt.Errorf("is path under: abs base %q: %w", resolvedBase, err)
	}

	absTarget, err := filepath.Abs(resolvedTarget)
	if err != nil {
		return false, fmt.Errorf("is path under: abs target %q: %w", resolvedTarget, err)
	}

	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false, fmt.Errorf("is path under: relative path from %q to %q: %w", absBase, absTarget, err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
