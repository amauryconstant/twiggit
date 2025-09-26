// Package domain contains core business entities and interfaces for twiggit
package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ContextType represents the type of context detected
type ContextType int

const (
	// ContextUnknown indicates the context could not be determined
	ContextUnknown ContextType = iota
	// ContextProject indicates the user is in a main project directory
	ContextProject
	// ContextWorktree indicates the user is in a worktree directory
	ContextWorktree
	// ContextOutsideGit indicates the user is outside any git repository
	ContextOutsideGit
)

// Context represents the detected user context
type Context struct {
	// Type is the type of context detected
	Type ContextType
	// ProjectName is the name of the project (if applicable)
	ProjectName string
	// BranchName is the name of the branch (if in worktree context)
	BranchName string
	// ProjectPath is the path to the main project directory
	ProjectPath string
	// WorktreePath is the path to the worktree directory (if applicable)
	WorktreePath string
	// CurrentPath is the current working directory
	CurrentPath string
}

// String returns a human-readable representation of the context
func (c *Context) String() string {
	switch c.Type {
	case ContextProject:
		return fmt.Sprintf("Project(project=%s, path=%s)", c.ProjectName, c.ProjectPath)
	case ContextWorktree:
		return fmt.Sprintf("Worktree(project=%s, branch=%s, path=%s)", c.ProjectName, c.BranchName, c.WorktreePath)
	case ContextOutsideGit:
		return "OutsideGit"
	case ContextUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}

// IsInGitContext returns true if the user is in any git context (project or worktree)
func (c *Context) IsInGitContext() bool {
	return c.Type == ContextProject || c.Type == ContextWorktree
}

// IsInProjectContext returns true if the user is in a main project directory
func (c *Context) IsInProjectContext() bool {
	return c.Type == ContextProject
}

// IsInWorktreeContext returns true if the user is in a worktree directory
func (c *Context) IsInWorktreeContext() bool {
	return c.Type == ContextWorktree
}

// FileSystemChecker defines an interface for checking filesystem operations
type FileSystemChecker interface {
	// Exists checks if a path exists on the filesystem
	Exists(path string) bool
}

// ContextDetector handles detection of user context
type ContextDetector struct {
	// workspaceBasePath is the base path for workspaces (e.g., "$HOME/Workspaces")
	workspaceBasePath string
	// projectsBasePath is the base path for projects (e.g., "$HOME/Projects")
	projectsBasePath string
	// fsChecker is the filesystem checker (can be mocked for testing)
	fsChecker FileSystemChecker
}

// NewContextDetector creates a new ContextDetector instance
func NewContextDetector(workspaceBasePath, projectsBasePath string, fsChecker FileSystemChecker) *ContextDetector {
	return &ContextDetector{
		workspaceBasePath: workspaceBasePath,
		projectsBasePath:  projectsBasePath,
		fsChecker:         fsChecker,
	}
}

// Detect detects the current user context
func (cd *ContextDetector) Detect(currentDir string) (*Context, error) {
	if currentDir == "" {
		return nil, NewWorktreeError(
			ErrValidation,
			"current directory path cannot be empty",
			"",
		).WithSuggestion("Provide a valid current directory path")
	}

	// Normalize the current directory path
	currentDir = filepath.Clean(currentDir)

	// First, check if we're in a workspace pattern (worktree context)
	if context := cd.detectWorktreeContext(currentDir); context != nil {
		return context, nil
	}

	// Then, check if we're in a project directory (project context)
	if context := cd.detectProjectContext(currentDir); context != nil {
		return context, nil
	}

	// If neither, we're outside git context
	return &Context{
		Type:        ContextOutsideGit,
		CurrentPath: currentDir,
	}, nil
}

// detectWorktreeContext detects if the current directory is in a worktree
func (cd *ContextDetector) detectWorktreeContext(currentDir string) *Context {
	// Check if the current directory matches the workspace pattern
	// Expected pattern: $HOME/Workspaces/<project>/<branch>/

	// Normalize paths for comparison
	normalizedCurrent := filepath.Clean(currentDir)
	normalizedWorkspace := filepath.Clean(cd.workspaceBasePath)

	// Check if current directory is under the workspace base path
	if !strings.HasPrefix(normalizedCurrent, normalizedWorkspace+string(filepath.Separator)) {
		return nil
	}

	// Extract the relative path from workspace base
	relativePath, err := filepath.Rel(normalizedWorkspace, normalizedCurrent)
	if err != nil {
		return nil
	}

	// Split the relative path into components
	components := strings.Split(relativePath, string(filepath.Separator))
	if len(components) < 2 {
		return nil // Need at least project/branch
	}

	projectName := components[0]
	branchName := components[1]

	// Validate project name and branch name
	if projectName == "" || branchName == "" {
		return nil
	}

	// Construct the expected project path
	expectedProjectPath := filepath.Join(cd.projectsBasePath, projectName)

	// Check if the project directory exists
	if !cd.fsChecker.Exists(expectedProjectPath) {
		return nil
	}

	return &Context{
		Type:         ContextWorktree,
		ProjectName:  projectName,
		BranchName:   branchName,
		ProjectPath:  expectedProjectPath,
		WorktreePath: normalizedCurrent,
		CurrentPath:  normalizedCurrent,
	}
}

// detectProjectContext detects if the current directory is in a main project
func (cd *ContextDetector) detectProjectContext(currentDir string) *Context {
	// Normalize paths for comparison
	normalizedCurrent := filepath.Clean(currentDir)
	normalizedProjects := filepath.Clean(cd.projectsBasePath)

	// Check if current directory is under the projects base path
	if !strings.HasPrefix(normalizedCurrent, normalizedProjects+string(filepath.Separator)) {
		return nil
	}

	// Extract the relative path from projects base
	relativePath, err := filepath.Rel(normalizedProjects, normalizedCurrent)
	if err != nil {
		return nil
	}

	// Split the relative path into components
	components := strings.Split(relativePath, string(filepath.Separator))
	if len(components) < 1 {
		return nil
	}

	projectName := components[0]

	// Validate project name
	if projectName == "" {
		return nil
	}

	// Check if this is exactly the project directory (not a subdirectory)
	expectedProjectPath := filepath.Join(normalizedProjects, projectName)
	if normalizedCurrent != expectedProjectPath {
		return nil
	}

	return &Context{
		Type:        ContextProject,
		ProjectName: projectName,
		ProjectPath: normalizedCurrent,
		CurrentPath: normalizedCurrent,
	}
}

// ContextResolution represents the result of resolving a target identifier
type ContextResolution struct {
	// TargetType is the type of target being resolved
	TargetType string
	// ProjectName is the resolved project name
	ProjectName string
	// BranchName is the resolved branch name (if applicable)
	BranchName string
	// TargetPath is the resolved target path
	TargetPath string
	// ResolutionMethod describes how the target was resolved
	ResolutionMethod string
}

// ContextResolver handles resolution of target identifiers based on context
type ContextResolver struct {
	workspaceBasePath string
	projectsBasePath  string
}

// NewContextResolver creates a new ContextResolver instance
func NewContextResolver(workspaceBasePath, projectsBasePath string) *ContextResolver {
	return &ContextResolver{
		workspaceBasePath: workspaceBasePath,
		projectsBasePath:  projectsBasePath,
	}
}

// Resolve resolves a target identifier based on the current context
func (cr *ContextResolver) Resolve(target string, context *Context) (*ContextResolution, error) {
	if target == "" {
		return nil, NewWorktreeError(
			ErrValidation,
			"target identifier cannot be empty",
			"",
		).WithSuggestion("Provide a target in the format 'project', 'branch', or 'project/branch'")
	}

	switch context.Type {
	case ContextProject:
		return cr.resolveFromProjectContext(target, context)
	case ContextWorktree:
		return cr.resolveFromWorktreeContext(target, context)
	case ContextOutsideGit:
		return cr.resolveFromOutsideGitContext(target, context)
	default:
		return nil, NewWorktreeError(
			ErrValidation,
			"unknown context type",
			"",
		).WithSuggestion("Ensure you are in a valid directory context")
	}
}

// resolveFromProjectContext resolves a target from a project directory context
func (cr *ContextResolver) resolveFromProjectContext(target string, context *Context) (*ContextResolution, error) {
	if strings.Contains(target, "/") {
		// Format: <project>/<branch> - cross-project worktree
		return cr.resolveCrossProjectWorktree(target)
	}

	// Check if target is the same as current project name
	if target == context.ProjectName {
		// User wants to stay in current project
		return &ContextResolution{
			TargetType:       "project",
			ProjectName:      context.ProjectName,
			TargetPath:       context.ProjectPath,
			ResolutionMethod: "current-project",
		}, nil
	}

	// Check if target is "main" - treat as current project
	if target == "main" {
		return &ContextResolution{
			TargetType:       "project",
			ProjectName:      context.ProjectName,
			TargetPath:       context.ProjectPath,
			ResolutionMethod: "main-alias",
		}, nil
	}

	// Assume target is a branch name for current project
	targetPath := filepath.Join(cr.workspaceBasePath, context.ProjectName, target)
	return &ContextResolution{
		TargetType:       "worktree",
		ProjectName:      context.ProjectName,
		BranchName:       target,
		TargetPath:       targetPath,
		ResolutionMethod: "project-branch",
	}, nil
}

// resolveFromWorktreeContext resolves a target from a worktree directory context
func (cr *ContextResolver) resolveFromWorktreeContext(target string, context *Context) (*ContextResolution, error) {
	if strings.Contains(target, "/") {
		// Format: <project>/<branch> - cross-project worktree
		return cr.resolveCrossProjectWorktree(target)
	}

	// Special case: "main" means go to main project directory
	if target == "main" {
		return &ContextResolution{
			TargetType:       "project",
			ProjectName:      context.ProjectName,
			TargetPath:       context.ProjectPath,
			ResolutionMethod: "main-special-case",
		}, nil
	}

	// Check if target is the same as current project name
	if target == context.ProjectName {
		// User wants to go to main project directory
		return &ContextResolution{
			TargetType:       "project",
			ProjectName:      context.ProjectName,
			TargetPath:       context.ProjectPath,
			ResolutionMethod: "project-name",
		}, nil
	}

	// Assume target is a different branch name for current project
	targetPath := filepath.Join(cr.workspaceBasePath, context.ProjectName, target)
	return &ContextResolution{
		TargetType:       "worktree",
		ProjectName:      context.ProjectName,
		BranchName:       target,
		TargetPath:       targetPath,
		ResolutionMethod: "worktree-branch",
	}, nil
}

// resolveFromOutsideGitContext resolves a target from outside git context
func (cr *ContextResolver) resolveFromOutsideGitContext(target string, _ *Context) (*ContextResolution, error) {
	if strings.Contains(target, "/") {
		// Format: <project>/<branch> - cross-project worktree
		return cr.resolveCrossProjectWorktree(target)
	}

	// Only project names are allowed from outside git context
	targetPath := filepath.Join(cr.projectsBasePath, target)
	return &ContextResolution{
		TargetType:       "project",
		ProjectName:      target,
		TargetPath:       targetPath,
		ResolutionMethod: "outside-project",
	}, nil
}

// resolveCrossProjectWorktree resolves a cross-project worktree target
func (cr *ContextResolver) resolveCrossProjectWorktree(target string) (*ContextResolution, error) {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return nil, NewWorktreeError(
			ErrValidation,
			"invalid cross-project worktree format",
			target,
		).WithSuggestion("Use the format 'project/branch' for cross-project worktrees")
	}

	projectName, branchName := parts[0], parts[1]
	targetPath := filepath.Join(cr.workspaceBasePath, projectName, branchName)

	return &ContextResolution{
		TargetType:       "worktree",
		ProjectName:      projectName,
		BranchName:       branchName,
		TargetPath:       targetPath,
		ResolutionMethod: "cross-project",
	}, nil
}
