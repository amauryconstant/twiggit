// Package mise provides integration with the mise tool for twiggit
package mise

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/amaury/twiggit/internal/infrastructure"
)

// MiseIntegration handles integration with Mise development environment tool
type MiseIntegration struct {
	execPath   string
	enabled    bool
	fileSystem infrastructure.FileSystem
}

// Option is a functional option for configuring MiseIntegration
type Option func(*MiseIntegration)

// WithFileSystem sets a custom filesystem for MiseIntegration
func WithFileSystem(fs infrastructure.FileSystem) Option {
	return func(mi *MiseIntegration) {
		mi.fileSystem = fs
	}
}

// WithExecPath sets a custom exec path for MiseIntegration
func WithExecPath(path string) Option {
	return func(mi *MiseIntegration) {
		mi.execPath = path
	}
}

// NewMiseIntegration creates a new MiseIntegration instance with optional configuration
func NewMiseIntegration(opts ...Option) *MiseIntegration {
	integration := &MiseIntegration{
		execPath:   "mise",
		enabled:    true,
		fileSystem: infrastructure.NewOSFileSystem(),
	}

	// Apply options
	for _, opt := range opts {
		opt(integration)
	}

	// Check if mise is available on system
	if !integration.IsAvailable() {
		integration.enabled = false
	}

	return integration
}

// IsAvailable checks if mise is available on system
func (mi *MiseIntegration) IsAvailable() bool {
	_, err := exec.LookPath(mi.execPath)
	return err == nil
}

// SetupWorktree sets up mise configuration for a new worktree
func (mi *MiseIntegration) SetupWorktree(sourceRepoPath, worktreePath string) error {
	// Validate target directory exists
	if _, err := mi.fileSystem.Stat(worktreePath); err != nil {
		return fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}

	// Detect configuration files in source repository
	configFiles := mi.DetectConfigFiles(sourceRepoPath)

	if len(configFiles) == 0 {
		// No mise config files found, nothing to set up
		return nil
	}

	// Copy configuration files to worktree
	if err := mi.CopyConfigFiles(sourceRepoPath, worktreePath, configFiles); err != nil {
		return fmt.Errorf("failed to copy mise config files: %w", err)
	}

	// Trust new worktree directory if mise is available
	if mi.enabled {
		if err := mi.TrustDirectory(worktreePath); err != nil {
			// Don't fail the entire operation if trust fails
			_ = err
		}
	}

	return nil
}

// DetectConfigFiles finds mise configuration files in the given repository path
func (mi *MiseIntegration) DetectConfigFiles(repoPath string) []string {
	var configFiles []string

	// Check for .mise.local.toml
	miseLocalFile := filepath.Join(repoPath, ".mise.local.toml")
	if _, err := mi.fileSystem.Stat(miseLocalFile); err == nil {
		configFiles = append(configFiles, ".mise.local.toml")
	}

	// Check for mise/config.local.toml
	miseConfigFile := filepath.Join(repoPath, "mise", "config.local.toml")
	if _, err := mi.fileSystem.Stat(miseConfigFile); err == nil {
		configFiles = append(configFiles, "mise/config.local.toml")
	}

	return configFiles
}

// CopyConfigFiles copies mise configuration files from source to target
func (mi *MiseIntegration) CopyConfigFiles(sourceDir, targetDir string, configFiles []string) error {
	for _, configFile := range configFiles {
		sourceFile := filepath.Join(sourceDir, configFile)
		targetFile := filepath.Join(targetDir, configFile)

		// Create target directory if needed
		targetFileDir := filepath.Dir(targetFile)
		if err := mi.fileSystem.MkdirAll(targetFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetFileDir, err)
		}

		// Read source file
		content, err := mi.fileSystem.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %w", sourceFile, err)
		}

		// Write to target file
		if err := mi.fileSystem.WriteFile(targetFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write target file %s: %w", targetFile, err)
		}
	}

	return nil
}

// TrustDirectory runs 'mise trust' on the specified directory if mise is available
func (mi *MiseIntegration) TrustDirectory(dirPath string) error {
	// Validate directory exists
	if _, err := mi.fileSystem.Stat(dirPath); err != nil {
		return fmt.Errorf("directory does not exist: %s", dirPath)
	}

	// Skip if mise is not available
	if !mi.enabled {
		return nil
	}

	// Run mise trust command
	cmd := exec.Command(mi.execPath, "trust", dirPath)

	// Capture output for potential debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mise trust failed for %s: %w (output: %s)", dirPath, err, string(output))
	}

	return nil
}

// Disable disables mise integration
func (mi *MiseIntegration) Disable() {
	mi.enabled = false
}

// Enable enables mise integration if mise is available
func (mi *MiseIntegration) Enable() {
	if mi.IsAvailable() {
		mi.enabled = true
	}
}

// IsEnabled returns whether mise integration is currently enabled
func (mi *MiseIntegration) IsEnabled() bool {
	return mi.enabled
}

// SetExecPath allows customizing the mise executable path (useful for testing)
func (mi *MiseIntegration) SetExecPath(path string) {
	mi.execPath = path
	// Re-check availability with new path
	mi.enabled = mi.IsAvailable()
}
