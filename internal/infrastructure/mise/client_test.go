package mise

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/test/mocks"
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
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return file exists for .mise.local.toml
	miseLocalFile := filepath.Join(s.tempDir, ".mise.local.toml")
	mockFileInfo := mocks.NewMockFileInfoWithDetails(".mise.local.toml", 25, 0644, time.Now(), false)
	mockFileSystem.On("Stat", miseLocalFile).Return(mockFileInfo, nil)

	// Setup mock to return file not found for mise/config.local.toml
	miseConfigFile := filepath.Join(s.tempDir, "mise", "config.local.toml")
	mockFileSystem.On("Stat", miseConfigFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseConfigFile, Err: os.ErrNotExist})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.Equal([]string{".mise.local.toml"}, configFiles)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithMiseConfigLocalToml() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return file not found for .mise.local.toml
	miseLocalFile := filepath.Join(s.tempDir, ".mise.local.toml")
	mockFileSystem.On("Stat", miseLocalFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseLocalFile, Err: os.ErrNotExist})

	// Setup mock to return file exists for mise/config.local.toml
	miseConfigFile := filepath.Join(s.tempDir, "mise", "config.local.toml")
	mockFileInfo := mocks.NewMockFileInfoWithDetails("config.local.toml", 20, 0644, time.Now(), false)
	mockFileSystem.On("Stat", miseConfigFile).Return(mockFileInfo, nil)

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.Equal([]string{"mise/config.local.toml"}, configFiles)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithBothConfigFiles() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return file exists for .mise.local.toml
	miseLocalFile := filepath.Join(s.tempDir, ".mise.local.toml")
	mockFileInfo1 := mocks.NewMockFileInfoWithDetails(".mise.local.toml", 25, 0644, time.Now(), false)
	mockFileSystem.On("Stat", miseLocalFile).Return(mockFileInfo1, nil)

	// Setup mock to return file exists for mise/config.local.toml
	miseConfigFile := filepath.Join(s.tempDir, "mise", "config.local.toml")
	mockFileInfo2 := mocks.NewMockFileInfoWithDetails("config.local.toml", 20, 0644, time.Now(), false)
	mockFileSystem.On("Stat", miseConfigFile).Return(mockFileInfo2, nil)

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	configFiles := integration.DetectConfigFiles(s.tempDir)
	s.ElementsMatch([]string{".mise.local.toml", "mise/config.local.toml"}, configFiles)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestDetectConfigFiles_WithNonexistentDirectory() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return file not found for both config files
	miseLocalFile := filepath.Join("/nonexistent/directory", ".mise.local.toml")
	mockFileSystem.On("Stat", miseLocalFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseLocalFile, Err: os.ErrNotExist})

	miseConfigFile := filepath.Join("/nonexistent/directory", "mise", "config.local.toml")
	mockFileSystem.On("Stat", miseConfigFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseConfigFile, Err: os.ErrNotExist})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	configFiles := integration.DetectConfigFiles("/nonexistent/directory")
	s.Empty(configFiles)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithEmptyList() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	targetDir := filepath.Join(s.tempDir, "target")

	err := integration.CopyConfigFiles(s.tempDir, targetDir, []string{})
	s.Require().NoError(err)

	// Verify mock expectations (no calls should be made for empty list)
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithSingleFile() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceFile := filepath.Join(s.tempDir, ".mise.local.toml")
	content := []byte("[tools]\nnode = \"20.0.0\"")
	targetDir := filepath.Join(s.tempDir, "target")
	targetFile := filepath.Join(targetDir, ".mise.local.toml")

	// Mock reading source file
	mockFileSystem.On("ReadFile", sourceFile).Return(content, nil)

	// Mock creating target directory (same as target dir for .mise.local.toml)
	mockFileSystem.On("MkdirAll", targetDir, os.FileMode(0755)).Return(nil)

	// Mock writing target file
	mockFileSystem.On("WriteFile", targetFile, content, os.FileMode(0644)).Return(nil)

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.CopyConfigFiles(s.tempDir, targetDir, []string{".mise.local.toml"})
	s.Require().NoError(err)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithNestedFile() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceFile := filepath.Join(s.tempDir, "mise", "config.local.toml")
	content := []byte("[tools]\ngo = \"1.21\"")
	targetDir := filepath.Join(s.tempDir, "target")
	targetMiseDir := filepath.Join(targetDir, "mise")
	targetFile := filepath.Join(targetMiseDir, "config.local.toml")

	// Mock reading source file
	mockFileSystem.On("ReadFile", sourceFile).Return(content, nil)

	// Mock creating target directory structure
	mockFileSystem.On("MkdirAll", targetMiseDir, os.FileMode(0755)).Return(nil)

	// Mock writing target file
	mockFileSystem.On("WriteFile", targetFile, content, os.FileMode(0644)).Return(nil)

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.CopyConfigFiles(s.tempDir, targetDir, []string{"mise/config.local.toml"})
	s.Require().NoError(err)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithNonexistentSourceFile() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceFile := filepath.Join(s.tempDir, "nonexistent.toml")
	targetDir := filepath.Join(s.tempDir, "target")

	// Mock reading source file to return error
	mockFileSystem.On("ReadFile", sourceFile).Return([]byte{}, &os.PathError{Op: "read", Path: sourceFile, Err: os.ErrNotExist})

	// Mock MkdirAll (this will be called even though ReadFile fails)
	mockFileSystem.On("MkdirAll", targetDir, os.FileMode(0755)).Return(nil)

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.CopyConfigFiles(s.tempDir, targetDir, []string{"nonexistent.toml"})
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to read source file")

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestCopyConfigFiles_WithUnwritableTarget() {
	// Create a mock filesystem that simulates permission errors
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to fail on directory creation (simulating permission error)
	targetDir := filepath.Join(s.tempDir, "target")
	targetFileDir := targetDir // For .mise.local.toml, target dir is same as target dir
	mockFileSystem.On("MkdirAll", targetFileDir, os.FileMode(0755)).Return(&os.PathError{Op: "mkdir", Path: targetFileDir, Err: os.ErrPermission})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	// Test copy operation - should fail due to mock filesystem error
	err := integration.CopyConfigFiles(s.tempDir, targetDir, []string{".mise.local.toml"})
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to create directory")

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
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
	s.Empty(integration.execPath)
	s.False(integration.IsEnabled())
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithNonexistentDirectory() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return error for directory stat
	nonexistentDir := "/nonexistent/directory"
	mockFileSystem.On("Stat", nonexistentDir).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: nonexistentDir, Err: os.ErrNotExist})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.TrustDirectory(nonexistentDir)
	s.Require().Error(err)
	s.Contains(err.Error(), "directory does not exist")

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithDisabledIntegration() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return success for directory stat (this will be called even when disabled)
	testDir := filepath.Join(s.tempDir, "test")
	mockFileInfo := mocks.NewMockFileInfoWithDetails("test", 0, 0755, time.Now(), true)
	mockFileSystem.On("Stat", testDir).Return(mockFileInfo, nil)

	// Create integration with mock filesystem and disable it
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))
	integration.Disable()

	err := integration.TrustDirectory(testDir)
	s.Require().NoError(err) // Should not error when disabled

	// Verify mock expectations (no calls should be made when disabled)
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestTrustDirectory_WithEnabledIntegration() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return success for directory stat
	testDir := filepath.Join(s.tempDir, "test")
	mockFileInfo := mocks.NewMockFileInfoWithDetails("test", 0, 0755, time.Now(), true)
	mockFileSystem.On("Stat", testDir).Return(mockFileInfo, nil)

	// Create integration with mock filesystem and custom exec path
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem), WithExecPath("echo"))

	err := integration.TrustDirectory(testDir)
	// Should not error since echo command should succeed
	s.Require().NoError(err)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithNonexistentTarget() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock to return error for target directory stat
	sourceDir := s.tempDir
	targetDir := "/nonexistent/target"
	mockFileSystem.On("Stat", targetDir).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: targetDir, Err: os.ErrNotExist})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.SetupWorktree(sourceDir, targetDir)
	s.Require().Error(err)
	s.Contains(err.Error(), "worktree path does not exist")

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithNoConfigFiles() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceDir := filepath.Join(s.tempDir, "source")
	targetDir := filepath.Join(s.tempDir, "target")

	// Mock target directory exists
	mockFileInfo := mocks.NewMockFileInfoWithDetails("target", 0, 0755, time.Now(), true)
	mockFileSystem.On("Stat", targetDir).Return(mockFileInfo, nil)

	// Mock no config files found in source
	miseLocalFile := filepath.Join(sourceDir, ".mise.local.toml")
	miseConfigFile := filepath.Join(sourceDir, "mise", "config.local.toml")
	mockFileSystem.On("Stat", miseLocalFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseLocalFile, Err: os.ErrNotExist})
	mockFileSystem.On("Stat", miseConfigFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseConfigFile, Err: os.ErrNotExist})

	// Create integration with mock filesystem
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))

	err := integration.SetupWorktree(sourceDir, targetDir)
	s.Require().NoError(err) // Should succeed even with no config files

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithConfigFiles() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceDir := filepath.Join(s.tempDir, "source")
	targetDir := filepath.Join(s.tempDir, "target")
	sourceMiseFile := filepath.Join(sourceDir, ".mise.local.toml")
	targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
	content := []byte("[tools]\nnode = \"20.0.0\"")

	// Mock target directory exists
	targetDirInfo := mocks.NewMockFileInfoWithDetails("target", 0, 0755, time.Now(), true)
	mockFileSystem.On("Stat", targetDir).Return(targetDirInfo, nil)

	// Mock config file exists in source
	sourceMiseFileInfo := mocks.NewMockFileInfoWithDetails(".mise.local.toml", 25, 0644, time.Now(), false)
	mockFileSystem.On("Stat", sourceMiseFile).Return(sourceMiseFileInfo, nil)

	// Mock no other config files
	miseConfigFile := filepath.Join(sourceDir, "mise", "config.local.toml")
	mockFileSystem.On("Stat", miseConfigFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseConfigFile, Err: os.ErrNotExist})

	// Mock reading source file
	mockFileSystem.On("ReadFile", sourceMiseFile).Return(content, nil)

	// Mock creating target directory (MkdirAll is called for target file directory)
	mockFileSystem.On("MkdirAll", targetDir, os.FileMode(0755)).Return(nil)

	// Mock writing target file
	mockFileSystem.On("WriteFile", targetMiseFile, content, os.FileMode(0644)).Return(nil)

	// Create integration with mock filesystem and disable mise to avoid trust command
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))
	integration.Disable()

	err := integration.SetupWorktree(sourceDir, targetDir)
	s.Require().NoError(err)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}

func (s *MiseClientTestSuite) TestSetupWorktree_WithDisabledIntegration() {
	// Create mock filesystem
	mockFileSystem := mocks.NewFileSystemMock()

	// Setup mock expectations
	sourceDir := filepath.Join(s.tempDir, "source")
	targetDir := filepath.Join(s.tempDir, "target")
	sourceMiseFile := filepath.Join(sourceDir, ".mise.local.toml")
	targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
	content := []byte("[tools]\nnode = \"20.0.0\"")

	// Mock target directory exists
	targetDirInfo := mocks.NewMockFileInfoWithDetails("target", 0, 0755, time.Now(), true)
	mockFileSystem.On("Stat", targetDir).Return(targetDirInfo, nil)

	// Mock config file exists in source
	sourceMiseFileInfo := mocks.NewMockFileInfoWithDetails(".mise.local.toml", 25, 0644, time.Now(), false)
	mockFileSystem.On("Stat", sourceMiseFile).Return(sourceMiseFileInfo, nil)

	// Mock no other config files
	miseConfigFile := filepath.Join(sourceDir, "mise", "config.local.toml")
	mockFileSystem.On("Stat", miseConfigFile).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: miseConfigFile, Err: os.ErrNotExist})

	// Mock reading source file
	mockFileSystem.On("ReadFile", sourceMiseFile).Return(content, nil)

	// Mock creating target directory (MkdirAll is called for target file directory)
	mockFileSystem.On("MkdirAll", targetDir, os.FileMode(0755)).Return(nil)

	// Mock writing target file
	mockFileSystem.On("WriteFile", targetMiseFile, content, os.FileMode(0644)).Return(nil)

	// Create integration with mock filesystem and disable it
	integration := NewMiseIntegration(WithFileSystem(mockFileSystem))
	integration.Disable()

	err := integration.SetupWorktree(sourceDir, targetDir)
	s.Require().NoError(err)

	// Verify mock expectations
	mockFileSystem.AssertExpectations(s.T())
}
