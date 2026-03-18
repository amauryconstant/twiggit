package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/test/mocks"
)

func setupShellService() (application.ShellService, *domain.Config) {
	config := domain.DefaultConfig()

	realShellInfra := infrastructure.NewShellInfrastructure()

	bashWrapper, _ := realShellInfra.GenerateWrapper(domain.ShellBash)
	zshWrapper, _ := realShellInfra.GenerateWrapper(domain.ShellZsh)
	fishWrapper, _ := realShellInfra.GenerateWrapper(domain.ShellFish)

	shellInfra := mocks.NewMockShellInfrastructure()
	shellInfra.On("GenerateWrapper", domain.ShellBash).Return(bashWrapper, nil)
	shellInfra.On("GenerateWrapper", domain.ShellZsh).Return(zshWrapper, nil)
	shellInfra.On("GenerateWrapper", domain.ShellFish).Return(fishWrapper, nil)
	shellInfra.On("DetectConfigFile", domain.ShellBash).Return("/home/user/.bashrc", nil)
	shellInfra.On("DetectConfigFile", domain.ShellZsh).Return("/home/user/.zshrc", nil)
	shellInfra.On("DetectConfigFile", domain.ShellFish).Return("/home/user/.config/fish/config.fish", nil)
	shellInfra.On("DetectConfigFile", mock.AnythingOfType("domain.ShellType")).Return("", nil).Maybe()
	shellInfra.On("InstallWrapper", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(domain.NewShellError(domain.ErrWrapperInstallation, "mock", "mock installation failure"))
	shellInfra.On("ValidateInstallation", mock.Anything, mock.Anything).Return(domain.NewShellError(domain.ErrShellNotInstalled, "mock", "mock validation failure"))
	service := NewShellService(shellInfra, config)

	return service, config
}

func TestShellService_SetupShell(t *testing.T) {
	tests := []struct {
		name        string
		request     *domain.SetupShellRequest
		expectError bool
		validate    func(*testing.T, *domain.SetupShellResult)
	}{
		{
			name: "force reinstall setup for bash",
			request: &domain.SetupShellRequest{
				ShellType:      domain.ShellBash,
				ForceOverwrite: true,
			},
			expectError: true,
		},
		{
			name: "force reinstall setup for zsh",
			request: &domain.SetupShellRequest{
				ShellType:      domain.ShellZsh,
				ForceOverwrite: true,
			},
			expectError: true,
		},
		{
			name: "force reinstall setup for fish",
			request: &domain.SetupShellRequest{
				ShellType:      domain.ShellFish,
				ForceOverwrite: true,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			result, err := service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.ShellType, result.ShellType)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}

func TestShellService_SetupShellValidation(t *testing.T) {
	tests := []struct {
		name         string
		request      *domain.SetupShellRequest
		setEnv       func(*testing.T)
		unsetEnv     func()
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &domain.SetupShellRequest{
				ShellType:      domain.ShellType("invalid"),
				ForceOverwrite: false,
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type with unsupported SHELL",
			request: &domain.SetupShellRequest{
				ShellType:      domain.ShellType(""),
				ForceOverwrite: false,
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			if tc.setEnv != nil {
				tc.setEnv(t)
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

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

func TestShellService_ValidateInstallation(t *testing.T) {
	tests := []struct {
		name        string
		request     *domain.ValidateInstallationRequest
		expectError bool
	}{
		{
			name: "validate bash installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellBash,
			},
			expectError: false,
		},
		{
			name: "validate zsh installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellZsh,
			},
			expectError: false,
		},
		{
			name: "validate fish installation",
			request: &domain.ValidateInstallationRequest{
				ShellType: domain.ShellFish,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			result, err := service.ValidateInstallation(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.Installed)
				assert.Equal(t, tc.request.ShellType, result.ShellType)
			}
		})
	}
}

func TestShellService_ValidateInstallationValidation(t *testing.T) {
	tests := []struct {
		name         string
		request      *domain.ValidateInstallationRequest
		setEnv       func(*testing.T)
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
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			if tc.setEnv != nil {
				tc.setEnv(t)
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

			result, err := service.ValidateInstallation(context.Background(), tc.request)

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

func TestShellService_GenerateWrapper(t *testing.T) {
	tests := []struct {
		name        string
		request     *domain.GenerateWrapperRequest
		expectError bool
		validate    func(*testing.T, *domain.GenerateWrapperResult)
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			result, err := service.GenerateWrapper(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				tc.validate(t, result)
			}
		})
	}
}

func TestShellService_GenerateWrapperValidation(t *testing.T) {
	tests := []struct {
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

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

func TestShellService_SetupShellAutoDetection(t *testing.T) {
	tests := []struct {
		name        string
		request     *domain.SetupShellRequest
		setEnv      func(*testing.T)
		unsetEnv    func()
		expectError bool
		validate    func(*testing.T, *domain.SetupShellResult)
	}{
		{
			name: "auto-detect bash when no args provided",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/zsh")
			},
			unsetEnv:    func() {},
			expectError: true,
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/usr/local/bin/fish")
			},
			unsetEnv:    func() {},
			expectError: true,
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:    func() {},
			expectError: true,
		},
		{
			name: "explicit shell overrides auto-detection",
			request: &domain.SetupShellRequest{
				ShellType:  domain.ShellZsh,
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
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
			},
			setEnv: func(t *testing.T) {
				t.Helper()
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *domain.SetupShellResult) {
				t.Helper()
				assert.Equal(t, domain.ShellBash, result.ShellType)
				assert.Equal(t, "/custom/bashrc", result.ConfigFile)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := setupShellService()

			if tc.setEnv != nil {
				tc.setEnv(t)
				defer tc.unsetEnv()
			}

			result, err := service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tc.validate != nil {
					tc.validate(t, result)
				}
			}
		})
	}
}
