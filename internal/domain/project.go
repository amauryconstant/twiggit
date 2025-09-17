// Package domain contains core business entities and interfaces for twiggit
package domain

import (
	"errors"
	"fmt"
	"os"
)

// Project represents a Git project aggregate containing multiple worktrees
type Project struct {
	// Name is the project identifier
	Name string
	// GitRepo is the path to the main Git repository
	GitRepo string
	// Worktrees contains all worktrees belonging to this project
	Worktrees []*Worktree
	// Metadata stores additional project information
	Metadata map[string]string
	// Config stores project configuration
	Config map[string]interface{}
}

// NewProject creates a new Project instance with validation
func NewProject(name, gitRepo string) (*Project, error) {
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}
	if gitRepo == "" {
		return nil, errors.New("git repository path cannot be empty")
	}

	return &Project{
		Name:      name,
		GitRepo:   gitRepo,
		Worktrees: make([]*Worktree, 0),
		Metadata:  make(map[string]string),
		Config:    make(map[string]interface{}),
	}, nil
}

// AddWorktree adds a new worktree to the project
func (p *Project) AddWorktree(worktree *Worktree) error {
	// Check for duplicate path
	for _, existing := range p.Worktrees {
		if existing.Path == worktree.Path {
			return fmt.Errorf("worktree already exists at path: %s", worktree.Path)
		}
	}

	p.Worktrees = append(p.Worktrees, worktree)
	return nil
}

// RemoveWorktree removes a worktree by path
func (p *Project) RemoveWorktree(path string) error {
	for i, worktree := range p.Worktrees {
		if worktree.Path == path {
			// Remove element by swapping with last and truncating
			p.Worktrees[i] = p.Worktrees[len(p.Worktrees)-1]
			p.Worktrees = p.Worktrees[:len(p.Worktrees)-1]
			return nil
		}
	}
	return fmt.Errorf("worktree not found at path: %s", path)
}

// GetWorktree retrieves a worktree by path
func (p *Project) GetWorktree(path string) (*Worktree, error) {
	for _, worktree := range p.Worktrees {
		if worktree.Path == path {
			return worktree, nil
		}
	}
	return nil, fmt.Errorf("worktree not found at path: %s", path)
}

// ListBranches returns a unique list of all branches in the project's worktrees
func (p *Project) ListBranches() []string {
	branchSet := make(map[string]struct{})
	for _, worktree := range p.Worktrees {
		branchSet[worktree.Branch] = struct{}{}
	}

	branches := make([]string, 0, len(branchSet))
	for branch := range branchSet {
		branches = append(branches, branch)
	}
	return branches
}

// SetMetadata sets a metadata key-value pair
func (p *Project) SetMetadata(key, value string) {
	p.Metadata[key] = value
}

// GetMetadata retrieves a metadata value by key
func (p *Project) GetMetadata(key string) (string, bool) {
	value, exists := p.Metadata[key]
	return value, exists
}

// ValidateGitRepoExists checks if the git repository path exists
func (p *Project) ValidateGitRepoExists() (bool, error) {
	if _, err := os.Stat(p.GitRepo); os.IsNotExist(err) {
		return false, fmt.Errorf("git repository path does not exist: %s", p.GitRepo)
	}
	return true, nil
}

// WorktreeStatistics represents statistics about project worktrees
type WorktreeStatistics struct {
	TotalCount   int
	UnknownCount int
	CleanCount   int
	DirtyCount   int
	Branches     []string
}

// GetWorktreeStatistics returns statistics about project worktrees
func (p *Project) GetWorktreeStatistics() *WorktreeStatistics {
	stats := &WorktreeStatistics{
		Branches: p.ListBranches(),
	}

	for _, worktree := range p.Worktrees {
		stats.TotalCount++
		switch worktree.Status {
		case StatusUnknown:
			stats.UnknownCount++
		case StatusClean:
			stats.CleanCount++
		case StatusDirty:
			stats.DirtyCount++
		}
	}

	return stats
}

// ProjectHealth represents health status of a project
type ProjectHealth struct {
	Status        string
	Issues        []string
	WorktreeCount int
}

// GetHealth returns health status of the project
func (p *Project) GetHealth() *ProjectHealth {
	health := &ProjectHealth{
		Status:        "unknown",
		Issues:        make([]string, 0),
		WorktreeCount: len(p.Worktrees),
	}

	// Check if git repo exists
	if exists, err := p.ValidateGitRepoExists(); err != nil || !exists {
		health.Issues = append(health.Issues, "git repository not validated")
	}

	// Determine overall status
	if len(health.Issues) == 0 {
		health.Status = "healthy"
	} else {
		health.Status = "unhealthy"
	}

	return health
}

// SetConfig sets a configuration key-value pair
func (p *Project) SetConfig(key string, value interface{}) {
	p.Config[key] = value
}

// GetConfig retrieves a configuration value by key
func (p *Project) GetConfig(key string) (interface{}, bool) {
	value, exists := p.Config[key]
	return value, exists
}

// GetWorktreesByBranch returns all worktrees for a specific branch
func (p *Project) GetWorktreesByBranch(branch string) []*Worktree {
	var result []*Worktree
	for _, worktree := range p.Worktrees {
		if worktree.Branch == branch {
			result = append(result, worktree)
		}
	}
	return result
}

// GetWorktreesByStatus returns all worktrees with a specific status
func (p *Project) GetWorktreesByStatus(status WorktreeStatus) []*Worktree {
	var result []*Worktree
	for _, worktree := range p.Worktrees {
		if worktree.Status == status {
			result = append(result, worktree)
		}
	}
	return result
}
