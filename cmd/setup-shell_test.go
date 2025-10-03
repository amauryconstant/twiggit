package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

// MockShellService implements ShellService for testing
type MockShellService struct {
	setupResult      *domain.SetupShellResult
	setupError       error
	validateResult   *domain.ValidateInstallationResult
	validateError    error
	generateResult   *domain.GenerateWrapperResult
	generateError    error
	calledSetup      bool
	calledValidate   bool
	calledGenerate   bool
	lastSetupRequest *domain.SetupShellRequest
	lastValidateReq  *domain.ValidateInstallationRequest
	lastGenerateReq  *domain.GenerateWrapperRequest
}

func (m *MockShellService) SetupShell(ctx context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error) {
	m.calledSetup = true
	m.lastSetupRequest = req
	return m.setupResult, m.setupError
}

func (m *MockShellService) ValidateInstallation(ctx context.Context, req *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error) {
	m.calledValidate = true
	m.lastValidateReq = req
	return m.validateResult, m.validateError
}

func (m *MockShellService) GenerateWrapper(ctx context.Context, req *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error) {
	m.calledGenerate = true
	m.lastGenerateReq = req
	return m.generateResult, m.generateError
}

func setupTestCommandConfig() *CommandConfig {
	mockShellService := &MockShellService{
		setupResult: &domain.SetupShellResult{
			ShellType:      domain.ShellBash,
			Installed:      true,
			DryRun:         false,
			WrapperContent: "# Twiggit bash wrapper\ntwiggit() { ... }",
			Message:        "Shell wrapper installed successfully",
		},
	}

	return &CommandConfig{
		Services: &ServiceContainer{
			ShellService: mockShellService,
		},
		Config: &domain.Config{},
	}
}

func TestSetupShellCommand_Success(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		validate    func(t *testing.T, output string, mock *MockShellService)
	}{
		{
			name: "setup bash shell dry run",
			args: []string{"--shell=bash", "--dry-run"},
			validate: func(t *testing.T, output string, mock *MockShellService) {
				t.Helper()
				assert.Contains(t, output, "Would install wrapper for bash:")
				assert.Contains(t, output, "twiggit() { builtin cd \"$target_dir\"; }")
				assert.True(t, mock.calledSetup)
				assert.NotNil(t, mock.lastSetupRequest)
				assert.Equal(t, domain.ShellBash, mock.lastSetupRequest.ShellType)
				assert.True(t, mock.lastSetupRequest.DryRun)
				assert.False(t, mock.lastSetupRequest.Force)
			},
		},
		{
			name: "setup zsh with force flag",
			args: []string{"--shell=zsh", "--force"},
			validate: func(t *testing.T, output string, mock *MockShellService) {
				t.Helper()
				assert.Contains(t, output, "Shell wrapper installed successfully")
				assert.True(t, mock.calledSetup)
				assert.NotNil(t, mock.lastSetupRequest)
				assert.Equal(t, domain.ShellZsh, mock.lastSetupRequest.ShellType)
				assert.False(t, mock.lastSetupRequest.DryRun)
				assert.True(t, mock.lastSetupRequest.Force)
			},
		},
		{
			name: "setup fish shell",
			args: []string{"--shell=fish"},
			validate: func(t *testing.T, output string, mock *MockShellService) {
				t.Helper()
				assert.Contains(t, output, "Shell wrapper installed successfully")
				assert.True(t, mock.calledSetup)
				assert.NotNil(t, mock.lastSetupRequest)
				assert.Equal(t, domain.ShellFish, mock.lastSetupRequest.ShellType)
				assert.False(t, mock.lastSetupRequest.DryRun)
				assert.False(t, mock.lastSetupRequest.Force)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			config := setupTestCommandConfig()
			mockService := config.Services.ShellService.(*MockShellService)

			// Configure mock for this test
			if tc.name == "setup bash shell dry run" {
				mockService.setupResult = &domain.SetupShellResult{
					ShellType:      domain.ShellBash,
					Installed:      false,
					DryRun:         true,
					WrapperContent: "# Twiggit bash wrapper\ntwiggit() { builtin cd \"$target_dir\"; }",
					Message:        "Dry run completed",
				}
			} else if tc.name == "setup zsh with force flag" {
				mockService.setupResult = &domain.SetupShellResult{
					ShellType: domain.ShellZsh,
					Installed: true,
					Message:   "Shell wrapper installed successfully",
				}
			} else if tc.name == "setup fish shell" {
				mockService.setupResult = &domain.SetupShellResult{
					ShellType: domain.ShellFish,
					Installed: true,
					Message:   "Shell wrapper installed successfully",
				}
			}

			cmd := NewSetupShellCmd(config)

			// Execute command
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, buf.String(), mockService)
				}
			}
		})
	}
}

func TestSetupShellCommand_Errors(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		expectError  bool
		errorMessage string
		setupMock    func(*MockShellService)
	}{
		{
			name:         "missing shell flag fails",
			args:         []string{"setup-shell"},
			expectError:  true,
			errorMessage: "required flag(s) \"shell\" not set",
		},
		{
			name:         "invalid shell type fails",
			args:         []string{"setup-shell", "--shell=invalid"},
			expectError:  true,
			errorMessage: "unsupported shell type: invalid",
		},
		{
			name:         "service error fails",
			args:         []string{"setup-shell", "--shell=bash"},
			expectError:  true,
			errorMessage: "shell error [WRAPPER_INSTALLATION_FAILED] for bash: service error",
			setupMock: func(m *MockShellService) {
				m.setupError = domain.NewShellError(domain.ErrWrapperInstallation, "bash", "service error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment
			config := setupTestCommandConfig()
			mockService := config.Services.ShellService.(*MockShellService)

			if tc.setupMock != nil {
				tc.setupMock(mockService)
			}

			cmd := NewSetupShellCmd(config)

			// Execute command
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectError {
				require.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetupShellCommand_SkippedInstallation(t *testing.T) {
	t.Run("already installed skips with force false", func(t *testing.T) {
		// Set up test environment
		config := setupTestCommandConfig()
		mockService := config.Services.ShellService.(*MockShellService)

		// Configure mock to return "already installed" result
		mockService.setupResult = &domain.SetupShellResult{
			ShellType: domain.ShellBash,
			Installed: true,
			Skipped:   true,
			Message:   "Shell wrapper already installed",
		}

		cmd := NewSetupShellCmd(config)

		// Execute command
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"setup-shell", "--shell=bash"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Shell wrapper already installed for bash")
		assert.Contains(t, output, "Use --force to reinstall")
		assert.True(t, mockService.calledSetup)
		assert.NotNil(t, mockService.lastSetupRequest)
		assert.False(t, mockService.lastSetupRequest.Force)
	})
}

func TestSetupShellCommand_Help(t *testing.T) {
	t.Run("command shows help", func(t *testing.T) {
		config := setupTestCommandConfig()
		cmd := NewSetupShellCmd(config)

		// Execute help
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()

		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Install shell wrapper functions that intercept 'twiggit cd' calls")
		assert.Contains(t, output, "--shell")
		assert.Contains(t, output, "--force")
		assert.Contains(t, output, "--dry-run")
		assert.Contains(t, output, "bash|zsh|fish")
	})
}
