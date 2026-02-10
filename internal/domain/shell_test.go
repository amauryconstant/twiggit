package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellDomain_ContractCompliance(t *testing.T) {
	testCases := []struct {
		name        string
		shellType   ShellType
		expectValid bool
	}{
		{
			name:        "bash shell type is valid",
			shellType:   ShellBash,
			expectValid: true,
		},
		{
			name:        "zsh shell type is valid",
			shellType:   ShellZsh,
			expectValid: true,
		},
		{
			name:        "fish shell type is valid",
			shellType:   ShellFish,
			expectValid: true,
		},
		{
			name:        "unknown shell type is invalid",
			shellType:   ShellType("unknown"),
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shell, err := NewShell(tc.shellType, "/bin/test", "1.0")

			if tc.expectValid {
				require.NoError(t, err)
				require.NotNil(t, shell)
				assert.Equal(t, tc.shellType, shell.Type())
				assert.Equal(t, "/bin/test", shell.Path())
				assert.Equal(t, "1.0", shell.Version())
				assert.NotEmpty(t, shell.ConfigFiles())
				assert.NotEmpty(t, shell.WrapperTemplate())
			} else {
				require.Error(t, err)
				assert.Nil(t, shell)
				assert.Contains(t, err.Error(), "unsupported shell type")
			}
		})
	}
}

func TestShellDomain_ConfigFiles(t *testing.T) {
	testCases := []struct {
		name          string
		shellType     ShellType
		expectedFiles []string
	}{
		{
			name:      "bash config files",
			shellType: ShellBash,
			expectedFiles: []string{
				".bashrc", ".bash_profile", ".profile",
			},
		},
		{
			name:      "zsh config files",
			shellType: ShellZsh,
			expectedFiles: []string{
				".zshrc", ".zprofile", ".profile",
			},
		},
		{
			name:      "fish config files",
			shellType: ShellFish,
			expectedFiles: []string{
				"config.fish", ".fishrc",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shell, err := NewShell(tc.shellType, "/bin/test", "1.0")
			require.NoError(t, err)

			configFiles := shell.ConfigFiles()
			assert.NotEmpty(t, configFiles)

			// Verify expected files are present
			for _, expectedFile := range tc.expectedFiles {
				assert.Contains(t, configFiles, expectedFile)
			}
		})
	}
}

func TestShellDomain_WrapperTemplate(t *testing.T) {
	testCases := []struct {
		name             string
		shellType        ShellType
		expectedContains []string
	}{
		{
			name:      "bash wrapper template",
			shellType: ShellBash,
			expectedContains: []string{
				"twiggit() {", "builtin cd", "command twiggit", "# Twiggit bash wrapper",
			},
		},
		{
			name:      "zsh wrapper template",
			shellType: ShellZsh,
			expectedContains: []string{
				"twiggit() {", "builtin cd", "command twiggit", "# Twiggit zsh wrapper",
			},
		},
		{
			name:      "fish wrapper template",
			shellType: ShellFish,
			expectedContains: []string{
				"function twiggit", "builtin cd", "command twiggit", "# Twiggit fish wrapper",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shell, err := NewShell(tc.shellType, "/bin/test", "1.0")
			require.NoError(t, err)

			template := shell.WrapperTemplate()
			assert.NotEmpty(t, template)

			// Verify expected content is present
			for _, expectedContent := range tc.expectedContains {
				assert.Contains(t, template, expectedContent)
			}
		})
	}
}

func TestShellDomain_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		shellType   ShellType
		path        string
		version     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid shell with all parameters",
			shellType:   ShellBash,
			path:        "/bin/bash",
			version:     "5.0.0",
			expectError: false,
		},
		{
			name:        "empty path should still work",
			shellType:   ShellZsh,
			path:        "",
			version:     "5.8",
			expectError: false,
		},
		{
			name:        "empty version should still work",
			shellType:   ShellFish,
			path:        "/usr/bin/fish",
			version:     "",
			expectError: false,
		},
		{
			name:        "invalid shell type",
			shellType:   ShellType("invalid"),
			path:        "/bin/invalid",
			version:     "1.0",
			expectError: true,
			errorMsg:    "unsupported shell type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shell, err := NewShell(tc.shellType, tc.path, tc.version)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, shell)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, shell)
				assert.Equal(t, tc.shellType, shell.Type())
				assert.Equal(t, tc.path, shell.Path())
				assert.Equal(t, tc.version, shell.Version())
			}
		})
	}
}

func TestInferShellTypeFromPath(t *testing.T) {
	testCases := []struct {
		name          string
		configPath    string
		expectedShell ShellType
		expectError   bool
		errorContains string
	}{
		{
			name:          "infer bash from .bashrc",
			configPath:    "/home/user/.bashrc",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from .bash_profile",
			configPath:    "/home/user/.bash_profile",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from .profile",
			configPath:    "/home/user/.profile",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from custom.bash",
			configPath:    "/home/user/custom.bash",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from my-bash-config",
			configPath:    "/home/user/my-bash-config",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer zsh from .zshrc",
			configPath:    "/home/user/.zshrc",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer zsh from .zprofile",
			configPath:    "/home/user/.zprofile",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer zsh from custom.zsh",
			configPath:    "/home/user/custom.zsh",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer fish from config.fish",
			configPath:    "/home/user/.config/fish/config.fish",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "infer fish from .fishrc",
			configPath:    "/home/user/.fishrc",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "infer fish from path containing fish",
			configPath:    "/home/user/fish/config",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "return error for unknown config file",
			configPath:    "/home/user/config.txt",
			expectedShell: "",
			expectError:   true,
			errorContains: "cannot infer shell type from path",
		},
		{
			name:          "return error for path without shell indicator",
			configPath:    "/home/user/myconfig",
			expectedShell: "",
			expectError:   true,
			errorContains: "cannot infer shell type from path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shellType, err := InferShellTypeFromPath(tc.configPath)

			if tc.expectError {
				require.Error(t, err)
				assert.Equal(t, ShellType(""), shellType)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedShell, shellType)
			}
		})
	}
}
