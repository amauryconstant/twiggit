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

	checkFlag := cmd.Flags().Lookup("check")
	assert.NotNil(t, checkFlag)

	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)

	shellFlag := cmd.Flags().Lookup("shell")
	assert.NotNil(t, shellFlag)
}

func TestNewInitCmd_FShortForm(t *testing.T) {
	config := &CommandConfig{}
	cmd := NewInitCmd(config)

	flag := cmd.Flags().Lookup("force")
	assert.NotNil(t, flag)
	assert.Equal(t, "f", flag.Shorthand)

	err := cmd.Flags().Set("force", "true")
	require.NoError(t, err)
	require.True(t, flag.Changed)

	force, err := cmd.Flags().GetBool("force")
	require.NoError(t, err)
	require.True(t, force)
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
	require.Contains(t, output, "Shell wrapper already installed for bash")
	require.Contains(t, output, "Config file: /home/user/.bashrc")
	require.Contains(t, output, "Use --force to reinstall")
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
	require.Contains(t, output, "Would install wrapper for zsh")
	require.Contains(t, output, "Config file: /home/user/.zshrc")
	require.Contains(t, output, "Wrapper function:")
	require.Contains(t, output, "# Twiggit wrapper")
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
