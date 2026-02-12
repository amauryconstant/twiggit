// Package domain contains core entities for git worktree management.
package domain

import (
	"errors"
)

// Project represents a git project with basic validation
type Project struct {
	name string
	path string
}

// NewProject creates a new project with validation
func NewProject(name, path string) (*Project, error) {
	if name == "" {
		return nil, errors.New("new project: name cannot be empty")
	}
	if path == "" {
		return nil, errors.New("new project: path cannot be empty")
	}

	return &Project{
		name: name,
		path: path,
	}, nil
}

// Name returns the project name
func (p *Project) Name() string {
	return p.name
}

// Path returns the project filesystem path
func (p *Project) Path() string {
	return p.path
}
