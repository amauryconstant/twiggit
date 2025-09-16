package domain

import "fmt"

// Workspace represents the bounded context containing all projects and their worktrees
type Workspace struct {
	// Path is the filesystem path to the workspace root
	Path string
	// Projects contains all projects in this workspace
	Projects []*Project
}

// NewWorkspace creates a new Workspace instance with validation
func NewWorkspace(path string) (*Workspace, error) {
	if path == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	return &Workspace{
		Path:     path,
		Projects: make([]*Project, 0),
	}, nil
}

// AddProject adds a new project to the workspace
func (w *Workspace) AddProject(project *Project) error {
	// Check for duplicate name
	for _, existing := range w.Projects {
		if existing.Name == project.Name {
			return fmt.Errorf("project already exists: %s", project.Name)
		}
	}

	w.Projects = append(w.Projects, project)
	return nil
}

// RemoveProject removes a project by name
func (w *Workspace) RemoveProject(name string) error {
	for i, project := range w.Projects {
		if project.Name == name {
			// Remove element by swapping with last and truncating
			w.Projects[i] = w.Projects[len(w.Projects)-1]
			w.Projects = w.Projects[:len(w.Projects)-1]
			return nil
		}
	}
	return fmt.Errorf("project not found: %s", name)
}

// GetProject retrieves a project by name
func (w *Workspace) GetProject(name string) (*Project, error) {
	for _, project := range w.Projects {
		if project.Name == name {
			return project, nil
		}
	}
	return nil, fmt.Errorf("project not found: %s", name)
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
	return nil, fmt.Errorf("worktree not found at path: %s", path)
}
