package infrastructure

import (
	"context"
	"os"
	"path/filepath"
)

// InfrastructureServiceImpl implements the InfrastructureService interface
// It provides filesystem and git repository operations for domain entities
type InfrastructureServiceImpl struct {
	deps *Deps
}

// NewInfrastructureService creates a new InfrastructureService instance
func NewInfrastructureService(deps *Deps) *InfrastructureServiceImpl {
	return &InfrastructureServiceImpl{
		deps: deps,
	}
}

// PathExists checks if a path exists on the filesystem
func (s *InfrastructureServiceImpl) PathExists(path string) bool {
	// Convert absolute path to relative path for fs.FS interface
	relPath := path
	if len(path) > 0 && path[0] == '/' {
		relPath = path[1:]
	}
	_, err := s.deps.Stat(relPath)
	return err == nil
}

// PathWritable checks if a path is writable
func (s *InfrastructureServiceImpl) PathWritable(path string) bool {
	// Check if path already exists
	if s.PathExists(path) {
		return false
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(path)

	// Convert absolute path to relative path for fs.FS interface
	relParentDir := parentDir
	if len(parentDir) > 0 && parentDir[0] == '/' {
		relParentDir = parentDir[1:]
	}

	parentInfo, err := s.deps.Stat(relParentDir)
	if err != nil {
		return false
	}

	if !parentInfo.IsDir() {
		return false
	}

	// Test writability by attempting to create a temporary file
	tempFile := filepath.Join(parentDir, ".twiggit-write-test")
	file, err := os.Create(tempFile)
	if err != nil {
		return false
	}

	// Clean up the temporary file
	_ = file.Close()
	_ = os.Remove(tempFile)

	return true
}

// IsGitRepository checks if a path is a valid git repository
func (s *InfrastructureServiceImpl) IsGitRepository(path string) bool {
	isRepo, err := s.deps.GitClient.IsGitRepository(context.TODO(), path)
	if err != nil {
		return false
	}
	return isRepo
}
