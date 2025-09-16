package domain

import "fmt"

// Project represents a Git project aggregate containing multiple worktrees
type Project struct {
	// Name is the project identifier
	Name string
	// GitRepo is the path to the main Git repository
	GitRepo string
	// Worktrees contains all worktrees belonging to this project
	Worktrees []*Worktree
}

// NewProject creates a new Project instance with validation
func NewProject(name, gitRepo string) (*Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	if gitRepo == "" {
		return nil, fmt.Errorf("git repository path cannot be empty")
	}

	return &Project{
		Name:      name,
		GitRepo:   gitRepo,
		Worktrees: make([]*Worktree, 0),
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
