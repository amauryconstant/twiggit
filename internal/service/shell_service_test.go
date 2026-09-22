package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/core"
	"twiggit/test/mocks"
)

func setupShellService() (application.ShellService, *core.Config) {
	config := core.DefaultConfig()

	bashWrapper, _ := core.ShellWrapper(core.ShellBash)
	zshWrapper, _ := core.ShellWrapper(core.ShellZsh)
	fishWrapper, _ := core.ShellWrapper(core.ShellFish)

	shellInfra := mocks.NewMockShellInfrastructure()
	shellInfra.On("GenerateWrapper", core.ShellBash).Return(bashWrapper, nil)
	shellInfra.On("GenerateWrapper", core.ShellZsh).Return(zshWrapper, nil)
	shellInfra.On("GenerateWrapper", core.ShellFish).Return(fishWrapper, nil)
	shellInfra.On("DetectConfigFile", core.ShellBash).Return("/home/user/.bashrc", nil)
	shellInfra.On("DetectConfigFile", core.ShellZsh).Return("/home/user/.zshrc", nil)
	shellInfra.On("DetectConfigFile", core.ShellFish).Return("/home/user/.config/fish/config.fish", nil)
	shellInfra.On("DetectConfigFile", mock.AnythingOfType("core.ShellType")).Return("", nil).Maybe()
	shellInfra.On("InstallWrapper", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.NewShellWrapperError("mock", "installation", "mock installation failure", nil))
	shellInfra.On("ValidateInstallation", mock.Anything, mock.Anything).Return(core.NewShellNotInstalledError("mock", "mock validation failure", nil))
	service := NewShellService(shellInfra, config)

	return service, config
}

func TestShellService_SetupShell(t *testing.T) {
	tests := []struct {
		name        string
		request     *core.SetupShellRequest
		expectError bool
		validate    func(*testing.T, *core.SetupShellResult)
	}{
		{
			name: "force reinstall setup for bash",
			request: &core.SetupShellRequest{
				ShellType:      core.ShellBash,
				ForceOverwrite: true,
			},
			expectError: true,
		},
		{
			name: "force reinstall setup for zsh",
			request: &core.SetupShellRequest{
				ShellType:      core.ShellZsh,
				ForceOverwrite: true,
			},
			expectError: true,
		},
		{
			name: "force reinstall setup for fish",
			request: &core.SetupShellRequest{
				ShellType:      core.ShellFish,
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
		request      *core.SetupShellRequest
		setEnv       func(*testing.T)
		unsetEnv     func()
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &core.SetupShellRequest{
				ShellType:      core.ShellType("invalid"),
				ForceOverwrite: false,
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type with unsupported SHELL",
			request: &core.SetupShellRequest{
				ShellType:      core.ShellType(""),
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
		request     *core.ValidateInstallationRequest
		expectError bool
	}{
		{
			name: "validate bash installation",
			request: &core.ValidateInstallationRequest{
				ShellType: core.ShellBash,
			},
			expectError: false,
		},
		{
			name: "validate zsh installation",
			request: &core.ValidateInstallationRequest{
				ShellType: core.ShellZsh,
			},
			expectError: false,
		},
		{
			name: "validate fish installation",
			request: &core.ValidateInstallationRequest{
				ShellType: core.ShellFish,
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
				assert.False(t, result.IsInstalled)
				assert.Equal(t, tc.request.ShellType, result.ShellType)
			}
		})
	}
}

func TestShellService_ValidateInstallationValidation(t *testing.T) {
	tests := []struct {
		name         string
		request      *core.ValidateInstallationRequest
		setEnv       func(*testing.T)
		unsetEnv     func()
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &core.ValidateInstallationRequest{
				ShellType: core.ShellType("invalid"),
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type with unsupported SHELL",
			request: &core.ValidateInstallationRequest{
				ShellType: core.ShellType(""),
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
		request     *core.GenerateWrapperRequest
		expectError bool
		validate    func(*testing.T, *core.GenerateWrapperResult)
	}{
		{
			name: "generate bash wrapper",
			request: &core.GenerateWrapperRequest{
				ShellType: core.ShellBash,
			},
			validate: func(t *testing.T, result *core.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, core.ShellBash, result.ShellType)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit bash wrapper")
			},
		},
		{
			name: "generate zsh wrapper",
			request: &core.GenerateWrapperRequest{
				ShellType: core.ShellZsh,
			},
			validate: func(t *testing.T, result *core.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, core.ShellZsh, result.ShellType)
				assert.NotEmpty(t, result.WrapperContent)
				assert.Contains(t, result.WrapperContent, "twiggit() {")
				assert.Contains(t, result.WrapperContent, "# Twiggit zsh wrapper")
			},
		},
		{
			name: "generate fish wrapper",
			request: &core.GenerateWrapperRequest{
				ShellType: core.ShellFish,
			},
			validate: func(t *testing.T, result *core.GenerateWrapperResult) {
				t.Helper()
				assert.Equal(t, core.ShellFish, result.ShellType)
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
		request      *core.GenerateWrapperRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "invalid shell type",
			request: &core.GenerateWrapperRequest{
				ShellType: core.ShellType("invalid"),
			},
			expectError:  true,
			errorMessage: "unsupported shell type",
		},
		{
			name: "empty shell type",
			request: &core.GenerateWrapperRequest{
				ShellType: core.ShellType(""),
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
		request     *core.SetupShellRequest
		setEnv      func(*testing.T)
		unsetEnv    func()
		expectError bool
		validate    func(*testing.T, *core.SetupShellResult)
	}{
		{
			name: "auto-detect bash when no args provided",
			request: &core.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, core.ShellBash, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".bashrc")
			},
		},
		{
			name: "auto-detect zsh when no args provided",
			request: &core.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/zsh")
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, core.ShellZsh, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".zshrc")
			},
		},
		{
			name: "auto-detect fish when no args provided",
			request: &core.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/usr/local/bin/fish")
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, core.ShellFish, result.ShellType)
				assert.Contains(t, result.ConfigFile, "config.fish")
			},
		},
		{
			name: "error when SHELL not set",
			request: &core.SetupShellRequest{
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
			request: &core.SetupShellRequest{
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
			request: &core.SetupShellRequest{
				ShellType:  core.ShellZsh,
				ConfigFile: "",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, core.ShellZsh, result.ShellType)
				assert.Contains(t, result.ConfigFile, ".zshrc")
			},
		},
		{
			name: "explicit config file overrides auto-detection",
			request: &core.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "/custom/zshrc",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, "/custom/zshrc", result.ConfigFile)
			},
		},
		{
			name: "both explicit shell and config file specified",
			request: &core.SetupShellRequest{
				ShellType:  core.ShellBash,
				ConfigFile: "/custom/bashrc",
			},
			setEnv: func(t *testing.T) {
				t.Helper()
			},
			unsetEnv:    func() {},
			expectError: true,
			validate: func(t *testing.T, result *core.SetupShellResult) {
				t.Helper()
				assert.Equal(t, core.ShellBash, result.ShellType)
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
