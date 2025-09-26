// Package infrastructure provides concrete implementations of external dependencies
// including Git clients, configuration management, and validation services.
package infrastructure

import (
	"fmt"
	"io/fs"
	"os"
)

// FileSystem extends the standard fs.FS interface with additional filesystem operations
// needed by the application services.
type FileSystem interface {
	fs.FS

	// Stat returns a FileInfo describing the named file.
	Stat(name string) (fs.FileInfo, error)

	// ReadDir reads the named directory, returning all its directory entries sorted by filename.
	ReadDir(name string) ([]fs.DirEntry, error)

	// MkdirAll creates a directory named path, along with any necessary parents,
	// and returns nil, or else returns an error.
	MkdirAll(path string, perm os.FileMode) error

	// ReadFile reads the named file and returns its contents.
	ReadFile(filename string) ([]byte, error)

	// WriteFile writes data to the named file, creating it if necessary.
	WriteFile(filename string, data []byte, perm os.FileMode) error

	// Remove removes the named file or directory.
	Remove(name string) error
}

// RealFileSystem is a concrete implementation of FileSystem that uses the actual OS filesystem
type RealFileSystem struct{}

// NewRealFileSystem creates a new RealFileSystem instance
func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}

// Open implements fs.FS
func (fsys *RealFileSystem) Open(name string) (fs.File, error) {
	file, err := os.DirFS("/").Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", name, err)
	}
	return file, nil
}

// Stat implements FileSystem
func (fsys *RealFileSystem) Stat(name string) (fs.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", name, err)
	}
	return info, nil
}

// ReadDir implements FileSystem
func (fsys *RealFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", name, err)
	}
	return entries, nil
}

// MkdirAll implements FileSystem
func (fsys *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// ReadFile implements FileSystem
func (fsys *RealFileSystem) ReadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return data, nil
}

// WriteFile implements FileSystem
func (fsys *RealFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(filename, data, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	return nil
}

// Remove implements FileSystem
// Remove removes the named file or directory.
func (fsys *RealFileSystem) Remove(name string) error {
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", name, err)
	}
	return nil
}
