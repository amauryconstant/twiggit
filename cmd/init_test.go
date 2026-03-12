package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

type InitCmdTestSuite struct {
	suite.Suite
	config       *CommandConfig
	shellService *mocks.MockShellService
}

func (s *InitCmdTestSuite) SetupTest() {
	s.shellService = mocks.NewMockShellService()
	s.config = &CommandConfig{
		Services: &ServiceContainer{
			ShellService: s.shellService,
		},
	}
}

func TestInitCmd(t *testing.T) {
	suite.Run(t, new(InitCmdTestSuite))
}

func (s *InitCmdTestSuite) TestNewInitCmd_BasicStructure() {
	cmd := NewInitCmd(s.config)

	s.NotNil(cmd)
	s.Contains(cmd.Use, "init [shell]")
	s.Equal("Generate or install shell wrapper", cmd.Short)

	// Check new flags exist
	installFlag := cmd.Flags().Lookup("install")
	s.NotNil(installFlag)
	s.Equal("i", installFlag.Shorthand)

	configFlag := cmd.Flags().Lookup("config")
	s.NotNil(configFlag)
	s.Equal("c", configFlag.Shorthand)

	forceFlag := cmd.Flags().Lookup("force")
	s.NotNil(forceFlag)
	s.Equal("f", forceFlag.Shorthand)

	// Check old flags are removed
	checkFlag := cmd.Flags().Lookup("check")
	s.Nil(checkFlag)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	s.Nil(dryRunFlag)

	shellFlag := cmd.Flags().Lookup("shell")
	s.Nil(shellFlag)
}

func (s *InitCmdTestSuite) TestNewInitCmd_AcceptsOptionalShellArg() {
	cmd := NewInitCmd(s.config)

	// Test with no args (should be valid)
	err := cmd.Args(cmd, []string{})
	s.Require().NoError(err)

	// Test with one arg (should be valid)
	err = cmd.Args(cmd, []string{"bash"})
	s.Require().NoError(err)

	// Test with two args (should be invalid)
	err = cmd.Args(cmd, []string{"bash", "extra"})
	s.Require().Error(err)
}

func (s *InitCmdTestSuite) TestFlagValidation_ConfigRequiresInstall() {
	cmd := NewInitCmd(s.config)

	// Set --config without --install should error
	err := cmd.Flags().Set("config", "/custom/bashrc")
	s.Require().NoError(err)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "--config requires --install")
}

func (s *InitCmdTestSuite) TestFlagValidation_ForceRequiresInstall() {
	cmd := NewInitCmd(s.config)

	// Set --force without --install should error
	err := cmd.Flags().Set("force", "true")
	s.Require().NoError(err)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "--force requires --install")
}

func (s *InitCmdTestSuite) TestStdoutMode_CallsGenerateWrapper() {
	s.shellService.On("GenerateWrapper", context.Background(), &domain.GenerateWrapperRequest{
		ShellType: domain.ShellBash,
	}).Return(&domain.GenerateWrapperResult{
		ShellType:      domain.ShellBash,
		WrapperContent: "# Twiggit bash wrapper\ntwiggit() { echo test; }",
		Message:        "Wrapper generated successfully",
	}, nil)

	cmd := NewInitCmd(s.config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash"})

	err := cmd.Execute()
	s.Require().NoError(err)

	// Should output wrapper content directly (no metadata)
	output := buf.String()
	s.Require().Contains(output, "# Twiggit bash wrapper")
	s.Require().Contains(output, "twiggit() {")
	// Should NOT contain installation messages
	s.NotContains(output, "installed")
	s.NotContains(output, "Config file:")

	s.shellService.AssertExpectations(s.T())
}

func (s *InitCmdTestSuite) TestStdoutMode_AutoDetectsShell() {
	s.T().Setenv("SHELL", "/bin/zsh")
	s.shellService.On("GenerateWrapper", context.Background(), &domain.GenerateWrapperRequest{
		ShellType: domain.ShellZsh,
	}).Return(&domain.GenerateWrapperResult{
		ShellType:      domain.ShellZsh,
		WrapperContent: "# Twiggit zsh wrapper\ntwiggit() { echo test; }",
		Message:        "Wrapper generated successfully",
	}, nil)

	cmd := NewInitCmd(s.config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{}) // No shell arg

	err := cmd.Execute()
	s.Require().NoError(err)

	output := buf.String()
	s.Require().Contains(output, "# Twiggit zsh wrapper")

	s.shellService.AssertExpectations(s.T())
}

func (s *InitCmdTestSuite) TestInstallMode_CallsSetupShell() {
	s.shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: false,
		ConfigFile:     "",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	cmd := NewInitCmd(s.config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install"})

	err := cmd.Execute()
	s.Require().NoError(err)

	output := buf.String()
	s.Require().Contains(output, "Shell wrapper installed for bash")
	s.Require().Contains(output, "Config file: /home/user/.bashrc")

	s.shellService.AssertExpectations(s.T())
}

func (s *InitCmdTestSuite) TestInstallMode_WithCustomConfig() {
	s.shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: false,
		ConfigFile:     "/custom/bashrc",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/custom/bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	cmd := NewInitCmd(s.config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install", "--config", "/custom/bashrc"})

	err := cmd.Execute()
	s.Require().NoError(err)

	output := buf.String()
	s.Require().Contains(output, "Shell wrapper installed for bash")
	s.Require().Contains(output, "Config file: /custom/bashrc")

	s.shellService.AssertExpectations(s.T())
}

func (s *InitCmdTestSuite) TestInstallMode_WithForce() {
	s.shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: true,
		ConfigFile:     "",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	cmd := NewInitCmd(s.config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install", "--force"})

	err := cmd.Execute()
	s.Require().NoError(err)

	s.shellService.AssertExpectations(s.T())
}

func TestDisplayInitResults_Skipped(t *testing.T) {
	out := &bytes.Buffer{}
	result := &domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		Skipped:    true,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Shell wrapper already installed",
	}

	err := displayInitResults(out, result)
	require.NoError(t, err)

	output := out.String()
	require.Contains(t, output, "Shell wrapper already installed for bash")
	require.Contains(t, output, "Config file: /home/user/.bashrc")
	require.Contains(t, output, "Use --force to reinstall")
}

func TestDisplayInitResults_Installed(t *testing.T) {
	out := &bytes.Buffer{}
	result := &domain.SetupShellResult{
		ShellType:  domain.ShellFish,
		Installed:  true,
		ConfigFile: "/home/user/.config/fish/config.fish",
		Message:    "Shell wrapper installed successfully",
	}

	err := displayInitResults(out, result)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Shell wrapper installed for fish")
	assert.Contains(t, output, "Config file: /home/user/.config/fish/config.fish")
	assert.Contains(t, output, "twiggit cd <branch>")
	assert.Contains(t, output, "builtin cd <path>")
}

func TestDisplayInitResults_NotInstalled(t *testing.T) {
	out := &bytes.Buffer{}
	result := &domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  false,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Installation failed",
	}

	err := displayInitResults(out, result)
	require.NoError(t, err)

	output := out.String()
	assert.NotContains(t, output, "Shell wrapper installed")
	assert.NotContains(t, output, "To activate the wrapper")
}
