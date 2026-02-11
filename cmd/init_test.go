package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestNewInitCmd(t *testing.T) {
	config := &CommandConfig{}
	cmd := NewInitCmd(config)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "init")
	assert.Equal(t, "Install shell wrapper", cmd.Short)
	assert.NotNil(t, cmd.Flags().Lookup("shell"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("check"))
}

func TestNewInitCmd_AcceptsOptionalConfigFile(t *testing.T) {
	config := &CommandConfig{}
	cmd := NewInitCmd(config)

	// Test with no args (should be valid now)
	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)

	// Test with one arg (should be valid)
	err = cmd.Args(cmd, []string{"/home/user/.bashrc"})
	require.NoError(t, err)

	// Test with two args (should be invalid)
	err = cmd.Args(cmd, []string{"/home/user/.bashrc", "extra"})
	require.Error(t, err)
}

func TestNewInitCmd_ShellFlagPrecedence(t *testing.T) {
	config := &CommandConfig{}
	cmd := NewInitCmd(config)

	// Verify --shell flag exists
	flag := cmd.Flags().Lookup("shell")
	require.NotNil(t, flag)
	assert.Equal(t, "shell", flag.Name)
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

	err := displayInitResults(out, result, false)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Shell wrapper already installed for bash")
	assert.Contains(t, output, "Config file: /home/user/.bashrc")
	assert.Contains(t, output, "Use --force to reinstall")
}

func TestDisplayInitResults_DryRun(t *testing.T) {
	out := &bytes.Buffer{}
	result := &domain.SetupShellResult{
		ShellType:      domain.ShellZsh,
		Installed:      false,
		DryRun:         true,
		ConfigFile:     "/home/user/.zshrc",
		WrapperContent: "# Twiggit wrapper\nfunction twiggit { echo test; }",
		Message:        "Dry run completed",
	}

	err := displayInitResults(out, result, true)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Would install wrapper for zsh")
	assert.Contains(t, output, "Config file: /home/user/.zshrc")
	assert.Contains(t, output, "Wrapper function:")
	assert.Contains(t, output, "# Twiggit wrapper")
}

func TestDisplayInitResults_Installed(t *testing.T) {
	out := &bytes.Buffer{}
	result := &domain.SetupShellResult{
		ShellType:  domain.ShellFish,
		Installed:  true,
		ConfigFile: "/home/user/.config/fish/config.fish",
		Message:    "Shell wrapper installed successfully",
	}

	err := displayInitResults(out, result, false)
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

	err := displayInitResults(out, result, false)
	require.NoError(t, err)

	output := out.String()
	assert.NotContains(t, output, "Shell wrapper installed")
	assert.NotContains(t, output, "To activate the wrapper")
}
