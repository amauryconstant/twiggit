package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
)

func TestShellIntegration_Inference(t *testing.T) {
	testCases := []struct {
		name          string
		configPath    string
		expectedShell domain.ShellType
		expectError   bool
	}{
		{
			name:          "infer bash from .bashrc",
			configPath:    "/home/user/.bashrc",
			expectedShell: domain.ShellBash,
			expectError:   false,
		},
		{
			name:          "infer zsh from .zshrc",
			configPath:    "/home/user/.zshrc",
			expectedShell: domain.ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer fish from config.fish",
			configPath:    "/home/user/.config/fish/config.fish",
			expectedShell: domain.ShellFish,
			expectError:   false,
		},
		{
			name:          "fail inference from unknown config",
			configPath:    "/home/user/config.txt",
			expectedShell: "",
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shellType, err := domain.InferShellTypeFromPath(tc.configPath)

			if tc.expectError {
				require.Error(t, err)
				assert.Empty(t, shellType)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedShell, shellType)
			}
		})
	}
}

func TestShellService_ForceReinstall(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("remove existing wrapper block on force reinstall", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, ".bashrc")

		shellInfra := infrastructure.NewShellInfrastructure()
		shellService := service.NewShellService(shellInfra, &domain.Config{})

		initialContent := `# Bash config
### BEGIN TWIGGIT WRAPPER
# Old wrapper
### END TWIGGIT WRAPPER
`
		require.NoError(t, os.WriteFile(configFile, []byte(initialContent), 0644))

		request := &domain.SetupShellRequest{
			ShellType:      domain.ShellBash,
			ConfigFile:     configFile,
			ForceOverwrite: true,
		}

		result, err := shellService.SetupShell(nil, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Installed)
		assert.False(t, result.Skipped)

		content, err := os.ReadFile(configFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "# Bash config", "original content preserved")
		assert.Contains(t, contentStr, "### BEGIN TWIGGIT WRAPPER", "new wrapper added")
		assert.Contains(t, contentStr, "### END TWIGGIT WRAPPER", "new wrapper added")
		assert.NotContains(t, contentStr, "# Old wrapper", "old wrapper removed")

		beginCount := strings.Count(contentStr, "### BEGIN TWIGGIT WRAPPER")
		endCount := strings.Count(contentStr, "### END TWIGGIT WRAPPER")
		assert.Equal(t, 1, beginCount, "should have exactly one BEGIN delimiter")
		assert.Equal(t, 1, endCount, "should have exactly one END delimiter")
	})

	t.Run("install when wrapper does not exist and force is true", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, ".bashrc")

		shellInfra := infrastructure.NewShellInfrastructure()
		shellService := service.NewShellService(shellInfra, &domain.Config{})

		initialContent := "# Bash config"
		require.NoError(t, os.WriteFile(configFile, []byte(initialContent), 0644))

		request := &domain.SetupShellRequest{
			ShellType:      domain.ShellBash,
			ConfigFile:     configFile,
			ForceOverwrite: true,
		}

		result, err := shellService.SetupShell(nil, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Installed)
		assert.False(t, result.Skipped)

		content, err := os.ReadFile(configFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, initialContent, "original content preserved")
		assert.Contains(t, contentStr, "### BEGIN TWIGGIT WRAPPER", "wrapper added")
		assert.Contains(t, contentStr, "### END TWIGGIT WRAPPER", "wrapper added")
	})
}

func TestShellService_SkipWhenInstalled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("skip installation when wrapper exists", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, ".bashrc")

		shellInfra := infrastructure.NewShellInfrastructure()
		shellService := service.NewShellService(shellInfra, &domain.Config{})

		initialContent := `# Bash config
### BEGIN TWIGGIT WRAPPER
# Existing wrapper
### END TWIGGIT WRAPPER
`
		require.NoError(t, os.WriteFile(configFile, []byte(initialContent), 0644))

		request := &domain.SetupShellRequest{
			ShellType:      domain.ShellBash,
			ConfigFile:     configFile,
			ForceOverwrite: false,
		}

		result, err := shellService.SetupShell(nil, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Installed, "result should indicate installed")
		assert.True(t, result.Skipped, "result should indicate skipped")
		assert.Equal(t, "Shell wrapper already installed", result.Message)

		content, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Equal(t, initialContent, string(content), "content should be unchanged")
	})

	t.Run("install when wrapper does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, ".bashrc")

		shellInfra := infrastructure.NewShellInfrastructure()
		shellService := service.NewShellService(shellInfra, &domain.Config{})

		initialContent := "# Bash config"
		require.NoError(t, os.WriteFile(configFile, []byte(initialContent), 0644))

		request := &domain.SetupShellRequest{
			ShellType:      domain.ShellBash,
			ConfigFile:     configFile,
			ForceOverwrite: false,
		}

		result, err := shellService.SetupShell(nil, request)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Installed)
		assert.False(t, result.Skipped)
		assert.Equal(t, "Shell wrapper installed successfully", result.Message)

		content, err := os.ReadFile(configFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, initialContent, "original content preserved")
		assert.Contains(t, contentStr, "### BEGIN TWIGGIT WRAPPER", "wrapper added")
		assert.Contains(t, contentStr, "### END TWIGGIT WRAPPER", "wrapper added")
	})
}
