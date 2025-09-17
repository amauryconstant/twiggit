package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiseIntegration_NewMiseIntegration(t *testing.T) {
	integration := NewMiseIntegration()

	assert.NotNil(t, integration)
	assert.Equal(t, "mise", integration.execPath)
	assert.True(t, integration.enabled) // Should be enabled by default if mise is available
}

func TestMiseIntegration_IsAvailable(t *testing.T) {
	integration := NewMiseIntegration()

	// This test depends on system state, but we can test the method exists
	available := integration.IsAvailable()

	// Result depends on whether mise is installed on the system
	// Just ensure the method works without panicking
	assert.IsType(t, true, available)
}

func TestMiseIntegration_SetupWorktree(t *testing.T) {
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
				require.NoError(t, err)

				// Create .mise.local.toml in source
				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte(`
[tools]
node = "20.0.0"
python = "3.11"

[env]
NODE_ENV = "development"
`), 0644)
				require.NoError(t, err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				require.NoError(t, err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{".mise.local.toml"},
		},
		{
			name: "should copy mise/config.local.toml pattern",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				require.NoError(t, err)

				// Create mise/config.local.toml in source
				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				require.NoError(t, err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte(`
[tools]
go = "1.21"
`), 0644)
				require.NoError(t, err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				require.NoError(t, err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{"mise/config.local.toml"},
		},
		{
			name: "should handle missing config files gracefully",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				require.NoError(t, err)
				// No mise config files
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			setupTarget: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-target-*")
				require.NoError(t, err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectError: false,
			expectFiles: []string{}, // No files should be copied
		},
		{
			name: "should return error for non-existent target",
			setupSourceRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-source-*")
				require.NoError(t, err)
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
		t.Run(tt.name, func(t *testing.T) {
			sourceRepo, cleanupSource := tt.setupSourceRepo()
			defer cleanupSource()

			targetPath, cleanupTarget := tt.setupTarget()
			defer cleanupTarget()

			integration := NewMiseIntegration()

			err := integration.SetupWorktree(sourceRepo, targetPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check that expected files were copied
				for _, expectedFile := range tt.expectFiles {
					targetFile := filepath.Join(targetPath, expectedFile)
					_, err := os.Stat(targetFile)
					assert.NoError(t, err, "Expected file %s should exist in target", expectedFile)
				}
			}
		})
	}
}

func TestMiseIntegration_TrustDirectory(t *testing.T) {
	tests := []struct {
		name        string
		setupDir    func() (string, func())
		expectError bool
	}{
		{
			name: "should handle trust operation gracefully",
			setupDir: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-trust-*")
				require.NoError(t, err)
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
		t.Run(tt.name, func(t *testing.T) {
			dirPath, cleanup := tt.setupDir()
			defer cleanup()

			integration := NewMiseIntegration()

			err := integration.TrustDirectory(dirPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Should not error (may be no-op if mise not available)
				assert.NoError(t, err)
			}
		})
	}
}

func TestMiseIntegration_DetectConfigFiles(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func() (string, func())
		expectedFiles []string
	}{
		{
			name: "should detect .mise.local.toml",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				require.NoError(t, err)

				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte("# test config"), 0644)
				require.NoError(t, err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{".mise.local.toml"},
		},
		{
			name: "should detect mise/config.local.toml",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				require.NoError(t, err)

				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				require.NoError(t, err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte("# test config"), 0644)
				require.NoError(t, err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{"mise/config.local.toml"},
		},
		{
			name: "should detect both config patterns",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				require.NoError(t, err)

				// Create .mise.local.toml
				miseFile := filepath.Join(tempDir, ".mise.local.toml")
				err = os.WriteFile(miseFile, []byte("# test config"), 0644)
				require.NoError(t, err)

				// Create mise/config.local.toml
				miseDir := filepath.Join(tempDir, "mise")
				err = os.MkdirAll(miseDir, 0755)
				require.NoError(t, err)

				configFile := filepath.Join(miseDir, "config.local.toml")
				err = os.WriteFile(configFile, []byte("# test config"), 0644)
				require.NoError(t, err)

				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{".mise.local.toml", "mise/config.local.toml"},
		},
		{
			name: "should return empty for no config files",
			setupRepo: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "mise-detect-*")
				require.NoError(t, err)
				return tempDir, func() { _ = os.RemoveAll(tempDir) }
			},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath, cleanup := tt.setupRepo()
			defer cleanup()

			integration := NewMiseIntegration()

			configFiles := integration.DetectConfigFiles(repoPath)

			assert.ElementsMatch(t, tt.expectedFiles, configFiles)
		})
	}
}

func TestMiseIntegration_CopyConfigFiles(t *testing.T) {
	t.Run("should copy files preserving directory structure", func(t *testing.T) {
		// Setup source directory with config files
		sourceDir, err := os.MkdirTemp("", "mise-copy-source-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(sourceDir) }()

		// Create .mise.local.toml
		miseFile := filepath.Join(sourceDir, ".mise.local.toml")
		miseContent := []byte(`[tools]
node = "20.0.0"`)
		err = os.WriteFile(miseFile, miseContent, 0644)
		require.NoError(t, err)

		// Create mise/config.local.toml
		miseDir := filepath.Join(sourceDir, "mise")
		err = os.MkdirAll(miseDir, 0755)
		require.NoError(t, err)

		configFile := filepath.Join(miseDir, "config.local.toml")
		configContent := []byte(`[tools]
go = "1.21"`)
		err = os.WriteFile(configFile, configContent, 0644)
		require.NoError(t, err)

		// Setup target directory
		targetDir, err := os.MkdirTemp("", "mise-copy-target-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(targetDir) }()

		integration := NewMiseIntegration()

		// Copy the files
		configFiles := []string{".mise.local.toml", "mise/config.local.toml"}
		err = integration.CopyConfigFiles(sourceDir, targetDir, configFiles)

		assert.NoError(t, err)

		// Verify files were copied correctly
		targetMiseFile := filepath.Join(targetDir, ".mise.local.toml")
		copiedMiseContent, err := os.ReadFile(targetMiseFile)
		assert.NoError(t, err)
		assert.Equal(t, miseContent, copiedMiseContent)

		targetConfigFile := filepath.Join(targetDir, "mise", "config.local.toml")
		copiedConfigContent, err := os.ReadFile(targetConfigFile)
		assert.NoError(t, err)
		assert.Equal(t, configContent, copiedConfigContent)
	})
}
