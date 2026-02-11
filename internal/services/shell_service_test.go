package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure/shell"
)

func TestShellService_SetupShell_Success(t *testing.T) {
	testCases := []struct {
		name        string
		request     *domain.SetupShellRequest
		expectError bool
		validate    func(t *testing.T, result *domain.SetupShellResult)
	}{
		{
			name: "dry run setup for bash",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellBash,
				Force:     false,
				DryRun:    true,
			},
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.True(t, result.DryRun)
				assert.False(t, result.Installed)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit bash wrapper")
			},
		},
		{
			name: "dry run setup for zsh",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellZsh,
				Force:     false,
				DryRun:    true,
			},
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.True(t, result.DryRun)
				assert.False(t, result.Installed)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit zsh wrapper")
			},
		},
		{
			name: "dry run setup for fish",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellFish,
				Force:     false,
				DryRun:    true,
			},
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.True(t, result.DryRun)
				assert.False(t, result.Installed)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "function twiggit")
				assert.Contains(t, result.WrapperContent, "# Twiggit fish wrapper")
			},
		},
		{
			name: "force reinstall setup",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellBash,
				Force:     true,
				DryRun:    false,
			},
			expectError: true, // Will fail since we can't actually install in tests
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestShellService()
			result, err := service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.request.ShellType, result.ShellType)
				tc.validate(t, result)
			}
		})
	}
}

func TestShellService_SetupShell_Validation(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.SetupShellRequest
		setEnv       func()
		unsetEnv     func()
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellType("invalid"),
				Force:     false,
				DryRun:    true,
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type with unsupported SHELL",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellType(""),
				Force:     false,
				DryRun:    true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv != nil {
				tc.setEnv()
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

			service := setupTestShellService()
			result, err := service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestShellService_ValidateInstallation_Success(t *testing.T) {
	testCases := []struct {
		name        string
		request     *domain.ValidateInstallationRequest
		expectError bool
	}{
		{
			name: "validate bash installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellBash,
			},
			expectError: false, // Should succeed with result showing not installed
		},
		{
			name: "validate zsh installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellZsh,
			},
			expectError: false, // Should succeed with result showing not installed
		},
		{
			name: "validate fish installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellFish,
			},
			expectError: false, // Should succeed with result showing not installed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestShellService()
			result, err := service.ValidateInstallation(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, result.Installed) // Should show not installed
				assert.Equal(t, tc.request.ShellType, result.ShellType)
			}
		})
	}
}

func TestShellService_ValidateInstallation_Validation(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.ValidateInstallationRequest
		setEnv       func()
		unsetEnv     func()
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellType("invalid"),
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type with unsupported SHELL",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellType(""),
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv != nil {
				tc.setEnv()
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

			service := setupTestShellService()
			result, err := service.ValidateInstallation(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestShellService_GenerateWrapper_Success(t *testing.T) {
	testCases := []struct {
		name        string
		request     *domain.GenerateWrapperRequest
		expectError bool
		validate    func(t *testing.T, result *domain.GenerateWrapperResult)
	}{
		{
			name: "generate bash wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellBash,
			},
			validate: func(t *testing.T, result *domain.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, domain.ShellBash, result.ShellType)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit bash wrapper")
			},
		},
		{
			name: "generate zsh wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellZsh,
			},
			validate: func(t *testing.T, result *domain.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, domain.ShellZsh, result.ShellType)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit zsh wrapper")
			},
		},
		{
			name: "generate fish wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellFish,
			},
			validate: func(t *testing.T, result *domain.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, domain.ShellFish, result.ShellType)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "function twiggit")
				assert.Contains(t, result.WrapperContent, "# Twiggit fish wrapper")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestShellService()
			result, err := service.GenerateWrapper(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tc.validate(t, result)
			}
		})
	}
}

func TestShellService_GenerateWrapper_Validation(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.GenerateWrapperRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellType("invalid"),
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellType(""),
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestShellService()
			result, err := service.GenerateWrapper(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// setupTestShellService creates a test instance of ShellService
func setupTestShellService() application.ShellService {
	// Create a mock shell integration service
	shellIntegration := &mockShellIntegration{}
	config := domain.DefaultConfig()

	return NewShellService(shellIntegration, config)
}

// mockShellIntegration is a mock implementation of shell integration
type mockShellIntegration struct{}

func (m *mockShellIntegration) GenerateWrapper(shellType domain.ShellType) (string, error) {
	// Use the real shell infrastructure service for wrapper generation
	realService := shell.NewShellService()
	return realService.GenerateWrapper(shellType)
}

func (m *mockShellIntegration) DetectConfigFile(shellType domain.ShellType) (string, error) {
	// Return a mock config file path based on shell type
	switch shellType {
	case domain.ShellBash:
		return "/home/user/.bashrc", nil
	case domain.ShellZsh:
		return "/home/user/.zshrc", nil
	case domain.ShellFish:
		return "/home/user/.config/fish/config.fish", nil
	default:
		return "/home/user/.bashrc", nil
	}
}

func (m *mockShellIntegration) InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error {
	// Always fail installation in tests to avoid actual file system changes
	return domain.NewShellError(domain.ErrWrapperInstallation, string(shellType), "mock installation failure")
}

func (m *mockShellIntegration) ValidateInstallation(shellType domain.ShellType, configFile string) error {
	// Always fail validation in tests since wrapper not installed
	return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "mock validation failure")
}

func TestShellService_composeWrapper(t *testing.T) {
	testCases := []struct {
		name      string
		template  string
		shellType domain.ShellType
		expected  string
	}{
		{
			name:      "template with %s format specifier",
			template:  "# Shell: %s",
			shellType: domain.ShellBash,
			expected:  "# Shell: {{SHELL_TYPE}}%!(EXTRA string=bash)",
		},
		{
			name:      "multiple %s placeholders",
			template:  "%s wrapper for %s",
			shellType: domain.ShellZsh,
			expected:  "{{SHELL_TYPE}} wrapper for zsh",
		},
		{
			name:      "custom template with %s",
			template:  "function twiggit_%s { echo 'hello'; }",
			shellType: domain.ShellFish,
			expected:  "function twiggit_{{SHELL_TYPE}} { echo 'hello'; }%!(EXTRA string=fish)",
		},
		{
			name:      "empty template",
			template:  "",
			shellType: domain.ShellBash,
			expected:  "%!(EXTRA string={{SHELL_TYPE}}, string=bash)",
		},
		{
			name:      "bash shell type",
			template:  "SHELL=%s",
			shellType: domain.ShellBash,
			expected:  "SHELL={{SHELL_TYPE}}%!(EXTRA string=bash)",
		},
		{
			name:      "zsh shell type",
			template:  "SHELL=%s",
			shellType: domain.ShellZsh,
			expected:  "SHELL={{SHELL_TYPE}}%!(EXTRA string=zsh)",
		},
		{
			name:      "fish shell type",
			template:  "SHELL=%s",
			shellType: domain.ShellFish,
			expected:  "SHELL={{SHELL_TYPE}}%!(EXTRA string=fish)",
		},
		{
			name:      "template without %s placeholder",
			template:  "No placeholders here",
			shellType: domain.ShellBash,
			expected:  "No placeholders here%!(EXTRA string={{SHELL_TYPE}}, string=bash)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := domain.DefaultConfig()
			service := &shellService{
				config: config,
			}

			result := service.composeWrapper(tc.template, tc.shellType)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestShellService_SetupShell_AutoDetection(t *testing.T) {
	testCases := []struct {
		name        string
		request     *domain.SetupShellRequest
		setEnv      func()
		unsetEnv    func()
		expectError bool
		validate    func(t *testing.T, result *domain.SetupShellResult)
	}{
		{
			name: "auto-detect bash when no args provided",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellBash, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".bashrc")
			},
		},
		{
			name: "auto-detect zsh when no args provided",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/zsh")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellZsh, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".zshrc")
			},
		},
		{
			name: "auto-detect fish when no args provided",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/usr/local/bin/fish")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellFish, result.ShellType)
				assert.Contains(t, result.ConfigFile, "config.fish")
			},
		},
		{
			name: "error when SHELL not set",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "")
			},
			unsetEnv:    func() {},
			expectError: true,
		},
		{
			name: "error when SHELL is unsupported",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:    func() {},
			expectError: true,
		},
		{
			name: "explicit --shell flag overrides auto-detection",
			request: &domain.SetupShellRequest{
				ShellType:  domain.ShellZsh,
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellZsh, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".zshrc")
			},
		},
		{
			name: "explicit config file overrides auto-detection",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "/custom/zshrc",
				DryRun:     true,
			},
			setEnv: func() {
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, "/custom/zshrc", result.ConfigFile)
			},
		},
		{
			name: "both explicit shell and config file specified",
			request: &domain.SetupShellRequest{
				ShellType:  domain.ShellBash,
				ConfigFile: "/custom/bashrc",
				DryRun:     true,
			},
			setEnv:      func() {},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellBash, result.ShellType)
				assert.Equal(t, "/custom/bashrc", result.ConfigFile)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setEnv()
			defer tc.unsetEnv()

			service := setupTestShellService()
			result, err := service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}
