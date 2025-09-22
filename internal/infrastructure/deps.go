// Package infrastructure contains external dependencies and implementations
package infrastructure

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/infrastructure/validation"
)

// Deps contains all essential dependencies for the application
// It provides a centralized dependency injection container
type Deps struct {
	GitClient     domain.GitClient
	Config        *config.Config
	FileSystem    fs.FS
	PathValidator domain.PathValidator
}

// NewDeps creates a new Deps container with initialized dependencies
func NewDeps(cfg *config.Config) *Deps {
	return &Deps{
		GitClient:     git.NewClient(),
		Config:        cfg,
		FileSystem:    os.DirFS("/"),
		PathValidator: validation.NewPathValidator(),
	}
}

// Stat wraps fs.Stat for convenience
func (d *Deps) Stat(name string) (fs.FileInfo, error) {
	info, err := fs.Stat(d.FileSystem, name)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", name, err)
	}
	return info, nil
}

// ReadDir wraps fs.ReadDir for convenience
func (d *Deps) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(d.FileSystem, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", name, err)
	}
	return entries, nil
}
