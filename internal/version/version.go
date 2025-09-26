// Package version provides version information for the twiggit CLI
package version

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// FileSystem defines the interface for filesystem operations
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
}

// DefaultFileSystem is the default implementation using os package
type DefaultFileSystem struct{}

// ReadFile reads a file from the filesystem
func (fs *DefaultFileSystem) ReadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return data, nil
}

// Option is a functional option for configuring version loading
type Option func(*versionLoader)

// WithFileSystem sets a custom filesystem for version loading
func WithFileSystem(fs FileSystem) Option {
	return func(vl *versionLoader) {
		vl.fileSystem = fs
	}
}

// versionLoader handles version loading with configurable dependencies
type versionLoader struct {
	fileSystem FileSystem
}

// Version returns the current version from go.mod with optional configuration
func Version(opts ...Option) string {
	loader := &versionLoader{
		fileSystem: &DefaultFileSystem{},
	}

	// Apply options
	for _, opt := range opts {
		opt(loader)
	}

	return loader.load()
}

// load handles the actual version loading logic
func (vl *versionLoader) load() string {
	// Read go.mod file
	data, err := vl.fileSystem.ReadFile("go.mod")
	if err != nil {
		return "dev"
	}

	// Extract version from comment using regex
	re := regexp.MustCompile(`// Version: ([^\s]+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return "dev"
	}

	return strings.TrimSpace(matches[1])
}

//revive:enable stutter
