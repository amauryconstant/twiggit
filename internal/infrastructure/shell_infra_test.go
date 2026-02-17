package infrastructure

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type ShellInfrastructureTestSuite struct {
	suite.Suite
}

func TestShellInfrastructure(t *testing.T) {
	suite.Run(t, new(ShellInfrastructureTestSuite))
}

func (s *ShellInfrastructureTestSuite) TestGenerateWrapper() {
	tests := []struct {
		name        string
		shellType   domain.ShellType
		expectError bool
		validate    func(wrapper string)
	}{
		{
			name:      "generate bash wrapper",
			shellType: domain.ShellBash,
			validate: func(wrapper string) {
				s.Contains(wrapper, "twiggit() {")
				s.Contains(wrapper, "builtin cd")
				s.Contains(wrapper, "command twiggit")
				s.Contains(wrapper, "# Twiggit bash wrapper")
			},
		},
		{
			name:      "generate zsh wrapper",
			shellType: domain.ShellZsh,
			validate: func(wrapper string) {
				s.Contains(wrapper, "twiggit() {")
				s.Contains(wrapper, "builtin cd")
				s.Contains(wrapper, "command twiggit")
				s.Contains(wrapper, "# Twiggit zsh wrapper")
			},
		},
		{
			name:      "generate fish wrapper",
			shellType: domain.ShellFish,
			validate: func(wrapper string) {
				s.Contains(wrapper, "function twiggit")
				s.Contains(wrapper, "builtin cd")
				s.Contains(wrapper, "command twiggit")
				s.Contains(wrapper, "# Twiggit fish wrapper")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			service := NewShellInfrastructure()
			wrapper, err := service.GenerateWrapper(tc.shellType)

			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.NotEmpty(wrapper)
				tc.validate(wrapper)
			}
		})
	}
}

func (s *ShellInfrastructureTestSuite) TestGenerateWrapper_InvalidShellType() {
	service := NewShellInfrastructure()
	wrapper, err := service.GenerateWrapper(domain.ShellType("invalid"))

	s.Require().Error(err)
	s.Empty(wrapper)
	s.Contains(err.Error(), "unsupported shell type")
}

func (s *ShellInfrastructureTestSuite) TestGenerateWrapper_SyntaxValidation() {
	tests := []struct {
		name      string
		shellType domain.ShellType
	}{
		{name: "bash wrapper syntax", shellType: domain.ShellBash},
		{name: "zsh wrapper syntax", shellType: domain.ShellZsh},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			service := NewShellInfrastructure()
			wrapper, err := service.GenerateWrapper(tc.shellType)
			s.Require().NoError(err)

			s.NotContains(wrapper, "]] ]]", "wrapper should not contain double closing brackets")
			s.Contains(wrapper, "if [[", "wrapper should use if [[ for conditionals")
			s.Contains(wrapper, "]] || [[", "wrapper should use ]]] || [[ for OR conditionals")
			s.Contains(wrapper, "]]; then", "wrapper should use ]]; then for conditional end")

			syntaxCheckCmd := "bash"
			if tc.shellType == domain.ShellZsh {
				syntaxCheckCmd = "zsh"
			}

			if _, err := exec.LookPath(syntaxCheckCmd); err == nil {
				tmpFile, err := os.CreateTemp("", "wrapper_test_*.sh")
				s.Require().NoError(err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString(wrapper)
				s.Require().NoError(err)
				tmpFile.Close()

				cmd := exec.Command(syntaxCheckCmd, "-n", tmpFile.Name())
				output, err := cmd.CombinedOutput()
				s.Require().NoError(err, "wrapper should have valid %s syntax: %s", syntaxCheckCmd, string(output))
			}
		})
	}
}

func (s *ShellInfrastructureTestSuite) TestGenerateWrapper_FishSyntaxValidation() {
	service := NewShellInfrastructure()
	wrapper, err := service.GenerateWrapper(domain.ShellFish)
	s.Require().NoError(err)

	s.NotContains(wrapper, "]] ]]", "fish wrapper should not contain bash-style double brackets")
	s.NotContains(wrapper, "if [[", "fish wrapper should not use bash-style if [[")
	s.Contains(wrapper, "if", "fish wrapper should use fish if syntax")
	s.Contains(wrapper, "or", "fish wrapper should use 'or' for OR conditionals")
	s.Contains(wrapper, "end", "fish wrapper should use 'end' for block closure")

	indentCount := strings.Count(wrapper, "    if")
	s.Positive(indentCount, "fish wrapper should contain properly indented if statements")
}

func (s *ShellInfrastructureTestSuite) TestDetectConfigFile() {
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
		s.Run(tc.name, func() {
			service := NewShellInfrastructure()
			configFile, err := service.DetectConfigFile(tc.shellType)

			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.NotEmpty(configFile)
				s.Contains(configFile, "/")
			}
		})
	}
}

func (s *ShellInfrastructureTestSuite) TestValidateInstallation() {
	originalHome := os.Getenv("HOME")
	tempHome := s.T().TempDir()
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
		s.Run(tc.name, func() {
			service := NewShellInfrastructure()
			configFile := tempHome + "/.bashrc"
			err := service.ValidateInstallation(tc.shellType, configFile)

			if tc.expectError {
				s.Require().Error(err)
				var shellErr *domain.ShellError
				s.Require().ErrorAs(err, &shellErr)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ShellInfrastructureTestSuite) TestValidateInstallation_InvalidShellType() {
	service := NewShellInfrastructure()
	err := service.ValidateInstallation(domain.ShellType("invalid"), "")

	s.Require().Error(err)
	s.Contains(err.Error(), "config file path is empty")
}

func (s *ShellInfrastructureTestSuite) TestHasWrapperBlock() {
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
		s.Run(tc.name, func() {
			result := shellInfraImpl.hasWrapperBlock(tc.content)
			s.Equal(tc.expectedResult, result)
		})
	}
}

func (s *ShellInfrastructureTestSuite) TestRemoveWrapperBlock() {
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
		s.Run(tc.name, func() {
			result := shellInfraImpl.removeWrapperBlock(tc.content)
			s.Equal(tc.expectedResult, result)
		})
	}
}
