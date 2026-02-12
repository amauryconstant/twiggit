package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/test/mocks"
)

type ShellServiceTestSuite struct {
	suite.Suite
	service          application.ShellService
	shellIntegration *mocks.MockShellInfrastructure
	config           *domain.Config
}

func (s *ShellServiceTestSuite) SetupTest() {
	s.config = domain.DefaultConfig()
	s.shellIntegration = mocks.NewMockShellInfrastructure()
	s.configureShellMock()
	s.service = NewShellService(s.shellIntegration, s.config)
}

func (s *ShellServiceTestSuite) configureShellMock() {
	s.shellIntegration.GenerateWrapperFunc = func(shellType domain.ShellType) (string, error) {
		realService := infrastructure.NewShellInfrastructure()
		return realService.GenerateWrapper(shellType)
	}

	s.shellIntegration.DetectConfigFileFunc = func(shellType domain.ShellType) (string, error) {
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

	s.shellIntegration.InstallWrapperFunc = func(shellType domain.ShellType, wrapper, configFile string, force bool) error {
		return domain.NewShellError(domain.ErrWrapperInstallation, string(shellType), "mock installation failure")
	}

	s.shellIntegration.ValidateInstallationFunc = func(shellType domain.ShellType, configFile string) error {
		return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "mock validation failure")
	}
}

func TestShellService(t *testing.T) {
	suite.Run(t, new(ShellServiceTestSuite))
}

func (s *ShellServiceTestSuite) TestSetupShell() {
	tests := []struct {
		name        string
		request     *domain.SetupShellRequest
		expectError bool
		validate    func(*domain.SetupShellResult)
	}{
		{
			name: "dry run setup for bash",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellBash,
				Force:     false,
				DryRun:    true,
			},
			validate: func(result *domain.SetupShellResult) {
				s.True(result.DryRun)
				s.False(result.Installed)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "twiggit() {")
				s.Contains(result.WrapperContent, "# Twiggit bash wrapper")
			},
		},
		{
			name: "dry run setup for zsh",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellZsh,
				Force:     false,
				DryRun:    true,
			},
			validate: func(result *domain.SetupShellResult) {
				s.True(result.DryRun)
				s.False(result.Installed)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "twiggit() {")
				s.Contains(result.WrapperContent, "# Twiggit zsh wrapper")
			},
		},
		{
			name: "dry run setup for fish",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellFish,
				Force:     false,
				DryRun:    true,
			},
			validate: func(result *domain.SetupShellResult) {
				s.True(result.DryRun)
				s.False(result.Installed)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "function twiggit")
				s.Contains(result.WrapperContent, "# Twiggit fish wrapper")
			},
		},
		{
			name: "force reinstall setup",
			request: &domain.SetupShellRequest{
				ShellType: domain.ShellBash,
				Force:     true,
				DryRun:    false,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				s.Equal(tc.request.ShellType, result.ShellType)
				tc.validate(result)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestSetupShellValidation() {
	tests := []struct {
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
				s.T().Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setEnv != nil {
				tc.setEnv()
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

			result, err := s.service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
				if tc.errorMessage != "" {
					s.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestValidateInstallation() {
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
		s.Run(tc.name, func() {
			result, err := s.service.ValidateInstallation(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				s.False(result.Installed)
				s.Equal(tc.request.ShellType, result.ShellType)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestValidateInstallationValidation() {
	tests := []struct {
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
				s.T().Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:     func() {},
			expectError:  true,
			errorMessage: "shell auto-detection failed",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setEnv != nil {
				tc.setEnv()
				defer func() {
					if tc.unsetEnv != nil {
						tc.unsetEnv()
					}
				}()
			}

			result, err := s.service.ValidateInstallation(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
				if tc.errorMessage != "" {
					s.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestGenerateWrapper() {
	tests := []struct {
		name        string
		request     *domain.GenerateWrapperRequest
		expectError bool
		validate    func(*domain.GenerateWrapperResult)
	}{
		{
			name: "generate bash wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellBash,
			},
			validate: func(result *domain.GenerateWrapperResult) {
				s.Equal(domain.ShellBash, result.ShellType)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "twiggit() {")
				s.Contains(result.WrapperContent, "# Twiggit bash wrapper")
			},
		},
		{
			name: "generate zsh wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellZsh,
			},
			validate: func(result *domain.GenerateWrapperResult) {
				s.Equal(domain.ShellZsh, result.ShellType)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "twiggit() {")
				s.Contains(result.WrapperContent, "# Twiggit zsh wrapper")
			},
		},
		{
			name: "generate fish wrapper",
			request: &domain.GenerateWrapperRequest{
				ShellType: domain.ShellFish,
			},
			validate: func(result *domain.GenerateWrapperResult) {
				s.Equal(domain.ShellFish, result.ShellType)
				s.NotEmpty(result.WrapperContent)
				s.Contains(result.WrapperContent, "function twiggit")
				s.Contains(result.WrapperContent, "# Twiggit fish wrapper")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.GenerateWrapper(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				tc.validate(result)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestGenerateWrapperValidation() {
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
		s.Run(tc.name, func() {
			result, err := s.service.GenerateWrapper(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
				if tc.errorMessage != "" {
					s.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *ShellServiceTestSuite) TestComposeWrapper() {
	tests := []struct {
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			config := domain.DefaultConfig()
			service := &shellService{
				config: config,
			}

			result := service.composeWrapper(tc.template, tc.shellType)

			s.Equal(tc.expected, result)
		})
	}
}

func (s *ShellServiceTestSuite) TestSetupShellAutoDetection() {
	tests := []struct {
		name        string
		request     *domain.SetupShellRequest
		setEnv      func()
		unsetEnv    func()
		expectError bool
		validate    func(*domain.SetupShellResult)
	}{
		{
			name: "auto-detect bash when no args provided",
			request: &domain.SetupShellRequest{
				ShellType:  "",
				ConfigFile: "",
				DryRun:     true,
			},
			setEnv: func() {
				s.T().Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(result *domain.SetupShellResult) {
				s.Equal(domain.ShellBash, result.ShellType)
				s.Contains(result.ConfigFile, ".bashrc")
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
				s.T().Setenv("SHELL", "/bin/zsh")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(result *domain.SetupShellResult) {
				s.Equal(domain.ShellZsh, result.ShellType)
				s.Contains(result.ConfigFile, ".zshrc")
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
				s.T().Setenv("SHELL", "/usr/local/bin/fish")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(result *domain.SetupShellResult) {
				s.Equal(domain.ShellFish, result.ShellType)
				s.Contains(result.ConfigFile, "config.fish")
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
				s.T().Setenv("SHELL", "")
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
				s.T().Setenv("SHELL", "/bin/sh")
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
				s.T().Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(result *domain.SetupShellResult) {
				s.Equal(domain.ShellZsh, result.ShellType)
				s.Contains(result.ConfigFile, ".zshrc")
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
				s.T().Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:    func() {},
			expectError: false,
			validate: func(result *domain.SetupShellResult) {
				s.Equal("/custom/zshrc", result.ConfigFile)
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
			validate: func(result *domain.SetupShellResult) {
				s.Equal(domain.ShellBash, result.ShellType)
				s.Equal("/custom/bashrc", result.ConfigFile)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setEnv != nil {
				tc.setEnv()
				defer tc.unsetEnv()
			}

			result, err := s.service.SetupShell(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				if tc.validate != nil {
					tc.validate(result)
				}
			}
		})
	}
}
