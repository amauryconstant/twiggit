package output

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
)

func TestGenerateWrapper(t *testing.T) {
	tests := []struct {
		name        string
		shellType   core.ShellType
		expectError bool
		validate    func(t *testing.T, wrapper string)
	}{
		{
			name:      "generate bash wrapper",
			shellType: core.ShellBash,
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
			shellType: core.ShellZsh,
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
			shellType: core.ShellFish,
			validate: func(t *testing.T, wrapper string) {
				t.Helper()
				assert.Contains(t, wrapper, "function twiggit")
				assert.Contains(t, wrapper, "builtin cd")
				assert.Contains(t, wrapper, "command twiggit")
				assert.Contains(t, wrapper, "# Twiggit fish wrapper")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapper, err := core.ShellWrapper(tc.shellType)

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

func TestDetectConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		shellType   core.ShellType
		expectError bool
	}{
		{name: "detect bash config file", shellType: core.ShellBash},
		{name: "detect zsh config file", shellType: core.ShellZsh},
		{name: "detect fish config file", shellType: core.ShellFish},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configFile, err := DetectConfigFile(tc.shellType)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, configFile)
				assert.Contains(t, configFile, "/")
			}
		})
	}
}

func TestValidateInstallation(t *testing.T) {
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	tests := []struct {
		name        string
		shellType   core.ShellType
		expectError bool
	}{
		{name: "validate bash installation", shellType: core.ShellBash, expectError: true},
		{name: "validate zsh installation", shellType: core.ShellZsh, expectError: true},
		{name: "validate fish installation", shellType: core.ShellFish, expectError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configFile := tempHome + "/.bashrc"
			err := ValidateInstallation(tc.shellType, configFile)

			if tc.expectError {
				require.Error(t, err)
				var oe *core.OperationError
				require.ErrorAs(t, err, &oe)
				require.Equal(t, "shell.not_installed", oe.Op)
			} else {
				require.NoError(t, err)
			}
		})
	}

	_ = originalHome // keep import live in case future tests need it
}

func TestValidateInstallation_InvalidShellType(t *testing.T) {
	err := ValidateInstallation(core.ShellType("invalid"), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config file path is empty")
}

func TestHasWrapperBlock(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedResult bool
	}{
		{
			name:           "has both delimiters",
			content:        "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n### END TWIGGIT WRAPPER\n# More config",
			expectedResult: true,
		},
		{
			name:           "only begin delimiter",
			content:        "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n",
			expectedResult: false,
		},
		{
			name:           "only end delimiter",
			content:        "# Some config\ntwiggit() { echo test; }\n### END TWIGGIT WRAPPER\n# More config",
			expectedResult: false,
		},
		{
			name:           "no delimiters",
			content:        "# Some config\ntwiggit() { echo test; }\n# More config",
			expectedResult: false,
		},
		{
			name:           "empty content",
			content:        "",
			expectedResult: false,
		},
		{
			name:           "wrapper with whitespace",
			content:        "# Some config\n  ### BEGIN TWIGGIT WRAPPER  \n twiggit() { echo test; }\n  ### END TWIGGIT WRAPPER  \n",
			expectedResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := hasWrapperBlock(tc.content)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestRemoveWrapperBlock(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedResult string
	}{
		{
			name:           "remove complete wrapper block",
			content:        "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n### END TWIGGIT WRAPPER\n# More config",
			expectedResult: "# Some config\n# More config",
		},
		{
			name:           "no delimiters returns original",
			content:        "# Some config\ntwiggit() { echo test; }\n# More config",
			expectedResult: "# Some config\ntwiggit() { echo test; }\n# More config",
		},
		{
			name:           "only begin delimiter removes to end",
			content:        "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n# More config",
			expectedResult: "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n# More config",
		},
		{
			name:           "empty content returns empty",
			content:        "",
			expectedResult: "",
		},
		{
			name:           "wrapper at start",
			content:        "### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n### END TWIGGIT WRAPPER\n# More config",
			expectedResult: "# More config",
		},
		{
			name:           "wrapper at end",
			content:        "# Some config\n### BEGIN TWIGGIT WRAPPER\ntwiggit() { echo test; }\n### END TWIGGIT WRAPPER",
			expectedResult: "# Some config\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := removeWrapperBlock(tc.content)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// TestShellWrapperSyntaxValidation exercises the wrapper templates via
// the bash/zsh interpreters when present on the system. It mirrors the
// pre-refactor behavior (which lived on the infrastructure layer's
// GenerateWrapper) and gates the run on shell availability so CI hosts
// without bash/zsh still pass.
func TestShellWrapperSyntaxValidation(t *testing.T) {
	tests := []struct {
		name      string
		shellType core.ShellType
	}{
		{name: "bash wrapper syntax", shellType: core.ShellBash},
		{name: "zsh wrapper syntax", shellType: core.ShellZsh},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapper, err := core.ShellWrapper(tc.shellType)
			require.NoError(t, err)

			assert.NotContains(t, wrapper, "]] ]]", "wrapper should not contain double closing brackets")
			assert.Contains(t, wrapper, "if [[", "wrapper should use if [[ for conditionals")
			assert.Contains(t, wrapper, "]] || [[", "wrapper should use ]]] || [[ for OR conditionals")
			assert.Contains(t, wrapper, "]]; then", "wrapper should use ]]; then for conditional end")

			syntaxCheckCmd := "bash"
			if tc.shellType == core.ShellZsh {
				syntaxCheckCmd = "zsh"
			}

			if _, err := exec.LookPath(syntaxCheckCmd); err == nil {
				tmpFile, err := os.CreateTemp("", "wrapper_test_*.sh")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(wrapper)
				require.NoError(t, err)
				tmpFile.Close()

				cmd := exec.Command(syntaxCheckCmd, "-n", tmpFile.Name())
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "wrapper should have valid %s syntax: %s", syntaxCheckCmd, string(output))
			}
		})
	}
}

func TestShellWrapperFishSyntaxValidation(t *testing.T) {
	wrapper, err := core.ShellWrapper(core.ShellFish)
	require.NoError(t, err)

	assert.NotContains(t, wrapper, "]] ]]", "fish wrapper should not contain bash-style double brackets")
	assert.NotContains(t, wrapper, "if [[", "fish wrapper should not use bash-style if [[")
	assert.Contains(t, wrapper, "if", "fish wrapper should use fish if syntax")
	assert.Contains(t, wrapper, "or", "fish wrapper should use 'or' for OR conditionals")
	assert.Contains(t, wrapper, "end", "fish wrapper should use 'end' for block closure")

	indentCount := strings.Count(wrapper, "    if")
	assert.Positive(t, indentCount, "fish wrapper should contain properly indented if statements")
}
