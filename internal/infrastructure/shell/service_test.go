package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestShellService_GenerateWrapper_Success(t *testing.T) {
	testCases := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
		validate    func(t *testing.T, wrapper string)
	}{
		{
			name:      "generate bash wrapper",
			shellType: domain.ShellBash,
			validate: func(t *testing.T, wrapper string) {
				t.Helper()
				assert.Contains(t, wrapper, "twiggit() {")
				assert.Contains(t, wrapper, "builtin cd")
				assert.Contains(t, wrapper, "command twiggit")
				assert.Contains(t, wrapper, "# Twiggit bash wrapper")
			},
		},
		{
			name:      "generate zsh wrapper",
			shellType: domain.ShellZsh,
			validate: func(t *testing.T, wrapper string) {
				t.Helper()
				assert.Contains(t, wrapper, "twiggit() {")
				assert.Contains(t, wrapper, "builtin cd")
				assert.Contains(t, wrapper, "command twiggit")
				assert.Contains(t, wrapper, "# Twiggit zsh wrapper")
			},
		},
		{
			name:      "generate fish wrapper",
			shellType: domain.ShellFish,
			validate: func(t *testing.T, wrapper string) {
				t.Helper()
				assert.Contains(t, wrapper, "function twiggit")
				assert.Contains(t, wrapper, "builtin cd")
				assert.Contains(t, wrapper, "command twiggit")
				assert.Contains(t, wrapper, "# Twiggit fish wrapper")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellService()
			wrapper, err := service.GenerateWrapper(tc.shellType)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, wrapper)
				tc.validate(t, wrapper)
			}
		})
	}
}

func TestShellService_GenerateWrapper_InvalidShellType(t *testing.T) {
	service := NewShellService()
	wrapper, err := service.GenerateWrapper(domain.ShellType("invalid"))

	require.Error(t, err)
	assert.Empty(t, wrapper)
	assert.Contains(t, err.Error(), "unsupported shell type")
}

func TestShellService_DetectConfigFile_Success(t *testing.T) {
	testCases := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
	}{
		{
			name:      "detect bash config file",
			shellType: domain.ShellBash,
		},
		{
			name:      "detect zsh config file",
			shellType: domain.ShellZsh,
		},
		{
			name:      "detect fish config file",
			shellType: domain.ShellFish,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellService()
			configFile, err := service.DetectConfigFile(tc.shellType)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, configFile)
				// Should contain home directory and a valid config file name
				assert.Contains(t, configFile, "/")
			}
		})
	}
}

func TestShellService_ValidateInstallation_Success(t *testing.T) {
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	testCases := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
	}{
		{
			name:        "validate bash installation",
			shellType:   domain.ShellBash,
			expectError: true, // Will fail since wrapper not installed
		},
		{
			name:        "validate zsh installation",
			shellType:   domain.ShellZsh,
			expectError: true, // Will fail since wrapper not installed
		},
		{
			name:        "validate fish installation",
			shellType:   domain.ShellFish,
			expectError: true, // Will fail since wrapper not installed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellService()
			err := service.ValidateInstallation(tc.shellType)

			if tc.expectError {
				require.Error(t, err)
				// Should contain shell error information
				assert.Contains(t, err.Error(), "shell error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShellService_ValidateInstallation_InvalidShellType(t *testing.T) {
	service := NewShellService()
	err := service.ValidateInstallation(domain.ShellType("invalid"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detect config file")
}
