package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestInitCmd_NewInitCmd_BasicStructure(t *testing.T) {
	shellService := mocks.NewMockShellService()
	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}

	cmd := NewInitCmd(config)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "init [shell]")
	assert.Equal(t, "Generate or install shell wrapper", cmd.Short)

	installFlag := cmd.Flags().Lookup("install")
	assert.NotNil(t, installFlag)
	assert.Equal(t, "i", installFlag.Shorthand)

	configFlag := cmd.Flags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)

	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	checkFlag := cmd.Flags().Lookup("check")
	assert.Nil(t, checkFlag)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.Nil(t, dryRunFlag)

	shellFlag := cmd.Flags().Lookup("shell")
	assert.Nil(t, shellFlag)
}

func TestInitCmd_NewInitCmd_AcceptsOptionalShellArg(t *testing.T) {
	shellService := mocks.NewMockShellService()
	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"bash"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"bash", "extra"})
	require.Error(t, err)
}

func TestInitCmd_FlagValidation_ConfigRequiresInstall(t *testing.T) {
	shellService := mocks.NewMockShellService()
	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)

	err := cmd.Flags().Set("config", "/custom/bashrc")
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "--config requires --install")
}

func TestInitCmd_FlagValidation_ForceRequiresInstall(t *testing.T) {
	shellService := mocks.NewMockShellService()
	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)

	err := cmd.Flags().Set("force", "true")
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err = cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "--force requires --install")
}

func TestInitCmd_StdoutMode_CallsGenerateWrapper(t *testing.T) {
	shellService := mocks.NewMockShellService()
	shellService.On("GenerateWrapper", context.Background(), &domain.GenerateWrapperRequest{
		ShellType: domain.ShellBash,
	}).Return(&domain.GenerateWrapperResult{
		ShellType:      domain.ShellBash,
		WrapperContent: "# Twiggit bash wrapper\ntwiggit() { echo test; }",
		Message:        "Wrapper generated successfully",
	}, nil)

	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "# Twiggit bash wrapper")
	require.Contains(t, output, "twiggit() {")
	assert.NotContains(t, output, "installed")
	assert.NotContains(t, output, "Config file:")

	shellService.AssertExpectations(t)
}

func TestInitCmd_StdoutMode_AutoDetectsShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	shellService := mocks.NewMockShellService()
	shellService.On("GenerateWrapper", context.Background(), &domain.GenerateWrapperRequest{
		ShellType: domain.ShellZsh,
	}).Return(&domain.GenerateWrapperResult{
		ShellType:      domain.ShellZsh,
		WrapperContent: "# Twiggit zsh wrapper\ntwiggit() { echo test; }",
		Message:        "Wrapper generated successfully",
	}, nil)

	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "# Twiggit zsh wrapper")

	shellService.AssertExpectations(t)
}

func TestInitCmd_InstallMode_CallsSetupShell(t *testing.T) {
	shellService := mocks.NewMockShellService()
	shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: false,
		ConfigFile:     "",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Shell wrapper installed for bash")
	require.Contains(t, output, "Config file: /home/user/.bashrc")

	shellService.AssertExpectations(t)
}

func TestInitCmd_InstallMode_WithCustomConfig(t *testing.T) {
	shellService := mocks.NewMockShellService()
	shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: false,
		ConfigFile:     "/custom/bashrc",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/custom/bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install", "--config", "/custom/bashrc"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Shell wrapper installed for bash")
	require.Contains(t, output, "Config file: /custom/bashrc")

	shellService.AssertExpectations(t)
}

func TestInitCmd_InstallMode_WithForce(t *testing.T) {
	shellService := mocks.NewMockShellService()
	shellService.On("SetupShell", context.Background(), &domain.SetupShellRequest{
		ShellType:      domain.ShellBash,
		ForceOverwrite: true,
		ConfigFile:     "",
	}).Return(&domain.SetupShellResult{
		ShellType:  domain.ShellBash,
		Installed:  true,
		ConfigFile: "/home/user/.bashrc",
		Message:    "Shell wrapper installed successfully",
	}, nil)

	config := &CommandConfig{
		Services: &ServiceContainer{
			ShellService: shellService,
		},
	}
	cmd := NewInitCmd(config)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"bash", "--install", "--force"})

	err := cmd.Execute()
	require.NoError(t, err)

	shellService.AssertExpectations(t)
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
