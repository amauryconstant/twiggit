// Package domain contains core business entities and interfaces for twiggit
package domain

import (
	"strings"
)

// Types

// WorkspaceStatistics represents statistics about the workspace
type WorkspaceStatistics struct {
	ProjectCount         int
	TotalWorktreeCount   int
	UnknownWorktreeCount int
	CleanWorktreeCount   int
	DirtyWorktreeCount   int
	AllBranches          []string
}

// WorkspaceHealth represents health status of the workspace
type WorkspaceHealth struct {
	Status        string
	Issues        []string
	ProjectCount  int
	WorktreeCount int
}

// Structs

// Workspace represents the bounded context containing all projects and their worktrees
type Workspace struct {
	// Path is the filesystem path to the workspace root
	Path string
	// Projects contains all projects in this workspace
	Projects []*Project
	// Metadata stores additional workspace information
	Metadata map[string]string
	// Config stores workspace configuration
	Config map[string]interface{}
}

// Constructors

// NewWorkspace creates a new Workspace instance with validation
func NewWorkspace(path string) (*Workspace, error) {
	// Validate workspace path using the new validation system
	validationResult := ValidateWorkspaceCreation(path)
	if !validationResult.IsValid() {
		return nil, validationResult.ToError()
	}

	return &Workspace{
		Path:     path,
		Projects: make([]*Project, 0),
		Metadata: make(map[string]string),
		Config:   make(map[string]interface{}),
	}, nil
}

// Methods

// AddProject adds a new project to the workspace
func (w *Workspace) AddProject(project *Project) error {
	// Validate project name using the new validation system
	validationResult := ValidateWorkspaceProjectName(project.Name, w)
	if !validationResult.IsValid() {
		return validationResult.ToError()
	}

	w.Projects = append(w.Projects, project)
	return nil
}

// RemoveProject removes a project by name
func (w *Workspace) RemoveProject(name string) error {
	// Validate that project exists using the new validation system
	validationResult := ValidateWorkspaceProjectExists(name, w)
	if !validationResult.IsValid() {
		return validationResult.ToError()
	}

	for i, project := range w.Projects {
		if project.Name == name {
			// Remove element by swapping with last and truncating
			w.Projects[i] = w.Projects[len(w.Projects)-1]
			w.Projects = w.Projects[:len(w.Projects)-1]
			return nil
		}
	}

	// This should not happen since we validated above
	return nil
}

// GetProject retrieves a project by name
func (w *Workspace) GetProject(name string) (*Project, error) {
	// Validate that project exists using the new validation system
	validationResult := ValidateWorkspaceProjectExists(name, w)
	if !validationResult.IsValid() {
		return nil, validationResult.ToError()
	}

	for _, project := range w.Projects {
		if project.Name == name {
			return project, nil
		}
	}

	// This should not happen since we validated above
	return nil, nil
}

// ListAllWorktrees returns all worktrees from all projects in the workspace
func (w *Workspace) ListAllWorktrees() []*Worktree {
	var allWorktrees []*Worktree
	for _, project := range w.Projects {
		allWorktrees = append(allWorktrees, project.Worktrees...)
	}
	return allWorktrees
}

// GetWorktreeByPath finds a worktree by its path across all projects
func (w *Workspace) GetWorktreeByPath(path string) (*Worktree, error) {
	for _, project := range w.Projects {
		for _, worktree := range project.Worktrees {
			if worktree.Path == path {
				return worktree, nil
			}
		}
	}
	return nil, NewWorkspaceError(ErrWorkspaceWorktreeNotFound, "worktree not found at path: "+path)
}

// GetStatistics returns statistics about the workspace
func (w *Workspace) GetStatistics() *WorkspaceStatistics {
	stats := &WorkspaceStatistics{
		AllBranches: make([]string, 0),
	}

	branchSet := make(map[string]struct{})

	for _, project := range w.Projects {
		stats.ProjectCount++
		projectStats := project.GetWorktreeStatistics()
		stats.TotalWorktreeCount += projectStats.TotalCount
		stats.UnknownWorktreeCount += projectStats.UnknownCount
		stats.CleanWorktreeCount += projectStats.CleanCount
		stats.DirtyWorktreeCount += projectStats.DirtyCount

		for _, branch := range projectStats.Branches {
			branchSet[branch] = struct{}{}
		}
	}

	for branch := range branchSet {
		stats.AllBranches = append(stats.AllBranches, branch)
	}

	return stats
}

// SetConfig sets a configuration key-value pair
func (w *Workspace) SetConfig(key string, value interface{}) {
	w.Config[key] = value
}

// GetConfig retrieves a configuration value by key
func (w *Workspace) GetConfig(key string) (interface{}, bool) {
	value, exists := w.Config[key]
	return value, exists
}

// GetHealth returns health status of the workspace
func (w *Workspace) GetHealth() *WorkspaceHealth {
	health := &WorkspaceHealth{
		Status:        "unknown",
		Issues:        make([]string, 0),
		ProjectCount:  len(w.Projects),
		WorktreeCount: len(w.ListAllWorktrees()),
	}

	// Validate workspace health using the new validation system
	validationResult := ValidateWorkspaceHealth(w)
	if !validationResult.IsValid() {
		for _, err := range validationResult.Errors {
			health.Issues = append(health.Issues, err.Message)
		}
	}

	// Determine overall status
	if len(health.Issues) == 0 {
		health.Status = "healthy"
	} else {
		health.Status = "unhealthy"
	}

	return health
}

// DiscoverProjects discovers projects in the workspace directory
func (w *Workspace) DiscoverProjects() ([]*Project, error) {
	// For now, return empty slice - this will be enhanced later
	return []*Project{}, nil
}

// SetMetadata sets a metadata key-value pair
func (w *Workspace) SetMetadata(key, value string) {
	w.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key
func (w *Workspace) GetMetadata(key string) (string, bool) {
	value, exists := w.Metadata[key]
	return value, exists
}

// FindWorktreesByBranch finds all worktrees with a specific branch across all projects
func (w *Workspace) FindWorktreesByBranch(branch string) []*Worktree {
	var result []*Worktree
	for _, project := range w.Projects {
		result = append(result, project.GetWorktreesByBranch(branch)...)
	}
	return result
}

// FindWorktreesByBranchPattern finds all worktrees matching a branch pattern across all projects
func (w *Workspace) FindWorktreesByBranchPattern(pattern string) []*Worktree {
	var result []*Worktree
	for _, project := range w.Projects {
		for _, worktree := range project.Worktrees {
			if w.matchesPattern(worktree.Branch, pattern) {
				result = append(result, worktree)
			}
		}
	}
	return result
}

// FindWorktreesByProject finds all worktrees for a specific project
func (w *Workspace) FindWorktreesByProject(projectName string) []*Worktree {
	project, err := w.GetProject(projectName)
	if err != nil {
		return []*Worktree{}
	}
	return project.Worktrees
}

// FindWorktreesByStatus finds all worktrees with a specific status across all projects
func (w *Workspace) FindWorktreesByStatus(status WorktreeStatus) []*Worktree {
	var result []*Worktree
	for _, project := range w.Projects {
		result = append(result, project.GetWorktreesByStatus(status)...)
	}
	return result
}

// Pure Functions

// matchesPattern checks if a branch name matches a pattern with wildcards
func (w *Workspace) matchesPattern(branch, pattern string) bool {
	// Handle simple wildcard patterns
	if pattern == "*" {
		return true
	}

	// Handle prefix patterns like "feature-*"
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(branch, prefix)
	}

	// Handle suffix patterns like "*-fix"
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(branch, suffix)
	}

	// Handle exact match
	return branch == pattern
}
