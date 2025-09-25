package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MiseClientTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *MiseClientTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "mise-client-test-*")
	s.Require().NoError(err)
}

func (s *MiseClientTestSuite) TearDownTest() {
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func TestMiseClientSuite(t *testing.T) {
	suite.Run(t, new(MiseClientTestSuite))
}

func (s *MiseClientTestSuite) TestNewMiseIntegration() {
	integration := NewMiseIntegration()

	s.NotNil(integration)
	s.Equal("mise", integration.execPath)
	// enabled status depends on whether mise is available on system
	s.IsType(true, integration.enabled)
}

func (s *MiseClientTestSuite) TestIsAvailable() {
	integration := NewMiseIntegration()

	// Test the method works (result depends on system)
	available := integration.IsAvailable()
	s.IsType(true, available)
}

func (s *MiseClientTestSuite) TestIsAvailable_WithCustomExecPath() {
	integration := &MiseIntegration{
		execPath: "nonexistent-command",
		enabled:  true,
	}

	available := integration.IsAvailable()
	s.False(available)
}

func (s *MiseClientTestSuite) TestIsAvailable_WithValidCommand() {
	integration := &MiseIntegration{
		execPath: "echo", // echo should be available on most systems
		enabled:  true,
	}

	available := integration.IsAvailable()
	s.True(available)
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithNoConfigFiles() {
	integration := NewMiseIntegration()

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.Empty(configFiles)
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithMiseLocalToml() {
	integration := NewMiseIntegration()

	// Create .mise.local.toml file
	miseLocalFile := filepath.Join(s.tempDir, ".mise.local.toml")
	err := os.WriteFile(miseLocalFile, []byte("[tools]\nnode = \"20.0.0\""), 0644)
	s.Require().NoError(err)

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.Equal([]string{".mise.local.toml"}, configFiles)
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithMiseConfigLocalToml() {
	integration := NewMiseIntegration()

	// Create mise/config.local.toml file
	miseDir := filepath.Join(s.tempDir, "mise")
	err := os.MkdirAll(miseDir, 0755)
	s.Require().NoError(err)

	configFile := filepath.Join(miseDir, "config.local.toml")
	err = os.WriteFile(configFile, []byte("[tools]\ngo = \"1.21\""), 0644)
	s.Require().NoError(err)

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.Equal([]string{"mise/config.local.toml"}, configFiles)
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithBothConfigFiles() {
	integration := NewMiseIntegration()

	// Create .mise.local.toml file
	miseLocalFile := filepath.Join(s.tempDir, ".mise.local.toml")
	err := os.WriteFile(miseLocalFile, []byte("[tools]\nnode = \"20.0.0\""), 0644)
	s.Require().NoError(err)

	// Create mise/config.local.toml file
	miseDir := filepath.Join(s.tempDir, "mise")
	err = os.MkdirAll(miseDir, 0755)
	s.Require().NoError(err)

	configFile := filepath.Join(miseDir, "config.local.toml")
	err = os.WriteFile(configFile, []byte("[tools]\ngo = \"1.21\""), 0644)
	s.Require().NoError(err)

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.ElementsMatch([]string{".mise.local.toml", "mise/config.local.toml"}, configFiles)
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithNonexistentDirectory() {
	integration := NewMiseIntegration()

	configFiles := integration.DetectConfigFiles("/nonexistent/directory")
	s.Empty(configFiles)
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithEmptyList() {
	integration := NewMiseIntegration()

	targetDir := filepath.Join(s.tempDir, "target")
	err := os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.CopyConfigFiles(s.tempDir, targetDir, []string{})
	s.NoError(err)
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithSingleFile() {
	integration := NewMiseIntegration()

	// Create source file
	sourceFile := filepath.Join(s.tempDir, ".mise.local.toml")
	content := []byte("[tools]\nnode = \"20.0.0\"")
	err := os.WriteFile(sourceFile, content, 0644)
	s.Require().NoError(err)

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.CopyConfigFiles(s.tempDir, targetDir, []string{".mise.local.toml"})
	s.NoError(err)

	// Verify file was copied
	targetFile := filepath.Join(targetDir, ".mise.local.toml")
	copiedContent, err := os.ReadFile(targetFile)
	s.Require().NoError(err)
	s.Equal(content, copiedContent)
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithNestedFile() {
	integration := NewMiseIntegration()

	// Create source directory structure
	sourceMiseDir := filepath.Join(s.tempDir, "mise")
	err := os.MkdirAll(sourceMiseDir, 0755)
	s.Require().NoError(err)

	// Create source file
	sourceFile := filepath.Join(sourceMiseDir, "config.local.toml")
	content := []byte("[tools]\ngo = \"1.21\"")
	err = os.WriteFile(sourceFile, content, 0644)
	s.Require().NoError(err)

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.CopyConfigFiles(s.tempDir, targetDir, []string{"mise/config.local.toml"})
	s.NoError(err)

	// Verify file was copied with directory structure
	targetFile := filepath.Join(targetDir, "mise", "config.local.toml")
	copiedContent, err := os.ReadFile(targetFile)
	s.Require().NoError(err)
	s.Equal(content, copiedContent)
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithNonexistentSourceFile() {
	integration := NewMiseIntegration()

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err := os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.CopyConfigFiles(s.tempDir, targetDir, []string{"nonexistent.toml"})
	s.Error(err)
	s.Contains(err.Error(), "failed to read source file")
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithUnwritableTarget() {
	integration := NewMiseIntegration()

	// Create source file
	sourceFile := filepath.Join(s.tempDir, ".mise.local.toml")
	content := []byte("[tools]\nnode = \"20.0.0\"")
	err := os.WriteFile(sourceFile, content, 0644)
	s.Require().NoError(err)

	// Create target directory that's not writable
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0444) // read-only
	s.Require().NoError(err)

	err = integration.CopyConfigFiles(s.tempDir, targetDir, []string{".mise.local.toml"})
	s.Error(err)
	s.Contains(err.Error(), "failed to write target file")
}

func (s *MiseClientTestSuite) TestDisable() {
	integration := NewMiseIntegration()
	integration.enabled = true

	integration.Disable()
	s.False(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestEnable_WhenMiseIsAvailable() {
	integration := &MiseIntegration{
		execPath: "echo", // Use echo as it should be available
		enabled:  false,
	}

	integration.Enable()
	s.True(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestEnable_WhenMiseIsNotAvailable() {
	integration := &MiseIntegration{
		execPath: "nonexistent-command",
		enabled:  false,
	}

	integration.Enable()
	s.False(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestIsEnabled() {
	integration := NewMiseIntegration()

	// Test initial state
	enabled := integration.IsEnabled()
	s.IsType(true, enabled)

	// Test after disable
	integration.Disable()
	s.False(integration.IsEnabled())

	// Test after enable
	integration.Enable()
	// Result depends on whether mise is available on system
	s.IsType(true, integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestSetExecPath_WithValidPath() {
	integration := NewMiseIntegration()

	integration.SetExecPath("echo")
	s.Equal("echo", integration.execPath)
	// Should be enabled since echo should be available
	s.True(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestSetExecPath_WithInvalidPath() {
	integration := NewMiseIntegration()

	integration.SetExecPath("nonexistent-command")
	s.Equal("nonexistent-command", integration.execPath)
	s.False(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestSetExecPath_WithEmptyPath() {
	integration := NewMiseIntegration()

	integration.SetExecPath("")
	s.Equal("", integration.execPath)
	s.False(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithNonexistentDirectory() {
	integration := NewMiseIntegration()

	err := integration.TrustDirectory("/nonexistent/directory")
	s.Error(err)
	s.Contains(err.Error(), "directory does not exist")
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithDisabledIntegration() {
	integration := NewMiseIntegration()
	integration.Disable()

	// Create a test directory
	testDir := filepath.Join(s.tempDir, "test")
	err := os.MkdirAll(testDir, 0755)
	s.Require().NoError(err)

	err = integration.TrustDirectory(testDir)
	s.NoError(err) // Should not error when disabled
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithEnabledIntegration() {
	integration := &MiseIntegration{
		execPath: "echo", // Use echo to simulate mise command
		enabled:  true,
	}

	// Create a test directory
	testDir := filepath.Join(s.tempDir, "test")
	err := os.MkdirAll(testDir, 0755)
	s.Require().NoError(err)

	err = integration.TrustDirectory(testDir)
	// Should not error since echo command should succeed
	s.NoError(err)
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithNonexistentTarget() {
	integration := NewMiseIntegration()

	sourceDir := s.tempDir
	targetDir := "/nonexistent/target"

	err := integration.SetupWorktree(sourceDir, targetDir)
	s.Error(err)
	s.Contains(err.Error(), "worktree path does not exist")
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithNoConfigFiles() {
	integration := NewMiseIntegration()

	// Create source directory (no config files)
	sourceDir := filepath.Join(s.tempDir, "source")
	err := os.MkdirAll(sourceDir, 0755)
	s.Require().NoError(err)

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.SetupWorktree(sourceDir, targetDir)
	s.NoError(err) // Should succeed even with no config files
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithConfigFiles() {
	integration := NewMiseIntegration()

	// Create source directory with config file
	sourceDir := filepath.Join(s.tempDir, "source")
	err := os.MkdirAll(sourceDir, 0755)
	s.Require().NoError(err)

	miseFile := filepath.Join(sourceDir, ".mise.local.toml")
	err = os.WriteFile(miseFile, []byte("[tools]\nnode = \"20.0.0\""), 0644)
	s.Require().NoError(err)

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.SetupWorktree(sourceDir, targetDir)
	s.NoError(err)

	// Verify config file was copied
	targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
	_, err = os.Stat(targetMiseFile)
	s.NoError(err)
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithDisabledIntegration() {
	integration := NewMiseIntegration()
	integration.Disable()

	// Create source directory with config file
	sourceDir := filepath.Join(s.tempDir, "source")
	err := os.MkdirAll(sourceDir, 0755)
	s.Require().NoError(err)

	miseFile := filepath.Join(sourceDir, ".mise.local.toml")
	err = os.WriteFile(miseFile, []byte("[tools]\nnode = \"20.0.0\""), 0644)
	s.Require().NoError(err)

	// Create target directory
	targetDir := filepath.Join(s.tempDir, "target")
	err = os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err)

	err = integration.SetupWorktree(sourceDir, targetDir)
	s.NoError(err)

	// Should still copy config files even when disabled
	targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
	_, err = os.Stat(targetMiseFile)
	s.NoError(err)
}
