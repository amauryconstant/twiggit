package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/test/helpers"
	"github.com/stretchr/testify/suite"
)

type MiseIntegrationTestSuite struct {
	suite.Suite
}

func (s *MiseIntegrationTestSuite) SetupTest() {
	// Setup code if needed
}

func (s *MiseIntegrationTestSuite) TearDownTest() {
	// Cleanup code if needed
}

func TestMiseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MiseIntegrationTestSuite))
}

func (s *MiseIntegrationTestSuite) TestNewMiseIntegration() {
	integration := helpers.NewMiseIntegration()

	s.NotNil(integration)
	s.True(integration.IsAvailable()) // Should check availability instead of accessing execPath directly
	s.True(integration.IsEnabled())   // Should use public method instead of accessing enabled directly
}

func (s *MiseIntegrationTestSuite) TestIsAvailable() {
	integration := helpers.NewMiseIntegration()

	// This test depends on system state, but we can test the method exists
	available := integration.IsAvailable()

	// Result depends on whether mise is installed on system
	// Just ensure that method works without panicking
	s.IsType(true, available)
}

func (s *MiseIntegrationTestSuite) TestSetupWorktree() {
	tests := []struct {
		name            string
		setupSourceRepo func() (string, func())
		setupTarget     func() (string, func())
		expectError     bool
		expectFiles     []string
	}{
		{
			name: "should copy .mise.local.toml from source to target",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				s.Require().NoError(err)

				// Create .mise.local.toml in source
				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte(`
[tools]
node = "20.0.0"
python = "3.11"

[env]
NODE_ENV = "development"
`), 0644)
				s.Require().NoError(err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{".mise.local.toml"},
		},
		{
			name: "should copy mise/config.local.toml pattern",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				s.Require().NoError(err)

				// Create mise/config.local.toml in source
				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				s.Require().NoError(err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte(`
[tools]
go = "1.21"
`), 0644)
				s.Require().NoError(err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{"mise/config.local.toml"},
		},
		{
			name: "should handle missing config files gracefully",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				s.Require().NoError(err)
				// No mise config files
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{}, // No files should be copied
		},
		{
			name: "should return error for non-existent target",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				return "/non/existent/path", func() {}
			},
			expectError: true,
			expectFiles: []string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceRepo, cleanupSource := tt.setupSourceRepo()
			defer cleanupSource()

			targetPath, cleanupTarget := tt.setupTarget()
			defer cleanupTarget()

			integration := helpers.NewMiseIntegration()

			err := integration.SetupWorktree(sourceRepo, targetPath)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)

				// Check that expected files were copied
				for _, expectedFile := range tt.expectFiles {
					targetFile := filepath.Join(targetPath, expectedFile)
					_, err := os.Stat(targetFile)
					s.NoError(err, "Expected file %s should exist in target", expectedFile)
				}
			}
		})
	}
}

func (s *MiseIntegrationTestSuite) TestTrustDirectory() {
	tests := []struct {
		name        string
		setupDir    func() (string, func())
		expectError bool
	}{
		{
			name: "should handle trust operation gracefully",
			setupDir: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-trust-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false, // Should not error even if mise is not available
		},
		{
			name: "should handle non-existent directory",
			setupDir: func() (string, func()) {
				return "/non/existent/path", func() {}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			dirPath, cleanup := tt.setupDir()
			defer cleanup()

			integration := helpers.NewMiseIntegration()

			err := integration.TrustDirectory(dirPath)

			if tt.expectError {
				s.Error(err)
			} else {
				// Should not error (may be no-op if mise not available)
				s.NoError(err)
			}
		})
	}
}

func (s *MiseIntegrationTestSuite) TestDetectConfigFiles() {
	tests := []struct {
		name          string
		setupRepo     func() (string, func())
		expectedFiles []string
	}{
		{
			name: "should detect .mise.local.toml",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				s.Require().NoError(err)

				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte("# test config"), 0644)
				s.Require().NoError(err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{".mise.local.toml"},
		},
		{
			name: "should detect mise/config.local.toml",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				s.Require().NoError(err)

				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				s.Require().NoError(err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte("# test config"), 0644)
				s.Require().NoError(err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{"mise/config.local.toml"},
		},
		{
			name: "should detect both config patterns",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				s.Require().NoError(err)

				// Create .mise.local.toml
				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte("# test config"), 0644)
				s.Require().NoError(err)

				// Create mise/config.local.toml
				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				s.Require().NoError(err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte("# test config"), 0644)
				s.Require().NoError(err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{".mise.local.toml", "mise/config.local.toml"},
		},
		{
			name: "should return empty for no config files",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				s.Require().NoError(err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			repoPath, cleanup := tt.setupRepo()
			defer cleanup()

			integration := helpers.NewMiseIntegration()

			configFiles := integration.DetectConfigFiles(repoPath)

			s.ElementsMatch(tt.expectedFiles, configFiles)
		})
	}
}

func (s *MiseIntegrationTestSuite) TestCopyConfigFiles() {
	s.Run("should copy files preserving directory structure", func() {
		// Setup source directory with config files
		sourceDir, err := os.MkdirTemp("", "mise-copy-source-*")
		s.Require().NoError(err)
		defer func() { _ = os.RemoveAll(sourceDir) }()

		// Create .mise.local.toml
		miseFile := filepath.Join(sourceDir, ".mise.local.toml")
		miseContent := []byte(`[tools]
node = "20.0.0"`)
		err = os.WriteFile(miseFile, miseContent, 0644)
		s.Require().NoError(err)

		// Create mise/config.local.toml
		miseDir := filepath.Join(sourceDir, "mise")
		err = os.MkdirAll(miseDir, 0755)
		s.Require().NoError(err)

		configFile := filepath.Join(miseDir, "config.local.toml")
		configContent := []byte(`[tools]
go = "1.21"`)
		err = os.WriteFile(configFile, configContent, 0644)
		s.Require().NoError(err)

		// Setup target directory
		targetDir, err := os.MkdirTemp("", "mise-copy-target-*")
		s.Require().NoError(err)
		defer func() { _ = os.RemoveAll(targetDir) }()

		integration := helpers.NewMiseIntegration()

		// Copy files
		configFiles := []string{".mise.local.toml", "mise/config.local.toml"}
		err = integration.CopyConfigFiles(sourceDir, targetDir, configFiles)

		s.NoError(err)

		// Verify files were copied correctly
		targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
		copiedMiseContent, err := os.ReadFile(targetMiseFile)
		s.NoError(err)
		s.Equal(miseContent, copiedMiseContent)

		targetConfigFile := filepath.Join(targetDir, "mise", "config.local.toml")
		copiedConfigContent, err := os.ReadFile(targetConfigFile)
		s.NoError(err)
		s.Equal(configContent, copiedConfigContent)
	})
}
