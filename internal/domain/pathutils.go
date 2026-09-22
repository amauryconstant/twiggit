package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

func NormalizePath(path string) (string, error) {
	cleaned := filepath.Clean(path)

	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("normalize path %q: abs: %w", path, err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil
	}

	return resolved, nil
}

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
