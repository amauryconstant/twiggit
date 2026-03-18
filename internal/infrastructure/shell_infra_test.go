package infrastructure

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestShellInfrastructure_GenerateWrapper(t *testing.T) {
	tests := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
		validate    func(t *testing.T, wrapper string)
	}{
		{
			name:      "generate bash wrapper",
			shellType: domain.ShellBash,
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
			shellType: domain.ShellZsh,
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
			shellType: domain.ShellFish,
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
			service := NewShellInfrastructure()
			wrapper, err := service.GenerateWrapper(tc.shellType)

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

func TestShellInfrastructure_GenerateWrapper_InvalidShellType(t *testing.T) {
	service := NewShellInfrastructure()
	wrapper, err := service.GenerateWrapper(domain.ShellType("invalid"))

	require.Error(t, err)
	assert.Empty(t, wrapper)
	assert.Contains(t, err.Error(), "unsupported shell type")
}

func TestShellInfrastructure_GenerateWrapper_SyntaxValidation(t *testing.T) {
	tests := []struct {
		name      string
		shellType domain.ShellType
	}{
		{name: "bash wrapper syntax", shellType: domain.ShellBash},
		{name: "zsh wrapper syntax", shellType: domain.ShellZsh},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellInfrastructure()
			wrapper, err := service.GenerateWrapper(tc.shellType)
			require.NoError(t, err)

			assert.NotContains(t, wrapper, "]] ]]", "wrapper should not contain double closing brackets")
			assert.Contains(t, wrapper, "if [[", "wrapper should use if [[ for conditionals")
			assert.Contains(t, wrapper, "]] || [[", "wrapper should use ]]] || [[ for OR conditionals")
			assert.Contains(t, wrapper, "]]; then", "wrapper should use ]]; then for conditional end")

			syntaxCheckCmd := "bash"
			if tc.shellType == domain.ShellZsh {
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

func TestShellInfrastructure_GenerateWrapper_FishSyntaxValidation(t *testing.T) {
	service := NewShellInfrastructure()
	wrapper, err := service.GenerateWrapper(domain.ShellFish)
	require.NoError(t, err)

	assert.NotContains(t, wrapper, "]] ]]", "fish wrapper should not contain bash-style double brackets")
	assert.NotContains(t, wrapper, "if [[", "fish wrapper should not use bash-style if [[")
	assert.Contains(t, wrapper, "if", "fish wrapper should use fish if syntax")
	assert.Contains(t, wrapper, "or", "fish wrapper should use 'or' for OR conditionals")
	assert.Contains(t, wrapper, "end", "fish wrapper should use 'end' for block closure")

	indentCount := strings.Count(wrapper, "    if")
	assert.Positive(t, indentCount, "fish wrapper should contain properly indented if statements")
}

func TestShellInfrastructure_DetectConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
	}{
		{
			name:      "detect bash config file",
			shellType: domain.ShellBash,
		},
		{
			name:      "detect zsh config file",
			shellType: domain.ShellZsh,
		},
		{
			name:      "detect fish config file",
			shellType: domain.ShellFish,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellInfrastructure()
			configFile, err := service.DetectConfigFile(tc.shellType)

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

func TestShellInfrastructure_ValidateInstallation(t *testing.T) {
	originalHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
	}{
		{
			name:        "validate bash installation",
			shellType:   domain.ShellBash,
			expectError: true,
		},
		{
			name:        "validate zsh installation",
			shellType:   domain.ShellZsh,
			expectError: true,
		},
		{
			name:        "validate fish installation",
			shellType:   domain.ShellFish,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewShellInfrastructure()
			configFile := tempHome + "/.bashrc"
			err := service.ValidateInstallation(tc.shellType, configFile)

			if tc.expectError {
				require.Error(t, err)
				var shellErr *domain.ShellError
				require.ErrorAs(t, err, &shellErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShellInfrastructure_ValidateInstallation_InvalidShellType(t *testing.T) {
	service := NewShellInfrastructure()
	err := service.ValidateInstallation(domain.ShellType("invalid"), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config file path is empty")
}

func TestShellInfrastructure_HasWrapperBlock(t *testing.T) {
	service := NewShellInfrastructure()
	shellInfraImpl := service.(*shellInfrastructure)

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
			result := shellInfraImpl.hasWrapperBlock(tc.content)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestShellInfrastructure_RemoveWrapperBlock(t *testing.T) {
	service := NewShellInfrastructure()
	shellInfraImpl := service.(*shellInfrastructure)

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
			result := shellInfraImpl.removeWrapperBlock(tc.content)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
