package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShellTestSuite struct {
	suite.Suite
}

func TestShellSuite(t *testing.T) {
	suite.Run(t, new(ShellTestSuite))
}

func (s *ShellTestSuite) TestContractCompliance() {
	testCases := []struct {
		name        string
		shellType   ShellType
		expectValid bool
	}{
		{
			name:        "bash shell type is valid",
			shellType:   ShellBash,
			expectValid: true,
		},
		{
			name:        "zsh shell type is valid",
			shellType:   ShellZsh,
			expectValid: true,
		},
		{
			name:        "fish shell type is valid",
			shellType:   ShellFish,
			expectValid: true,
		},
		{
			name:        "unknown shell type is invalid",
			shellType:   ShellType("unknown"),
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			shell, err := NewShell(tc.shellType, "/bin/test", "1.0")

			if tc.expectValid {
				s.Require().NoError(err)
				s.Require().NotNil(shell)
				s.Equal(tc.shellType, shell.Type())
				s.Equal("/bin/test", shell.Path())
				s.Equal("1.0", shell.Version())
			} else {
				s.Require().Error(err)
				s.Nil(shell)
				s.Contains(err.Error(), "unsupported shell type")
			}
		})
	}
}

func (s *ShellTestSuite) TestValidation() {
	testCases := []struct {
		name        string
		shellType   ShellType
		path        string
		version     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid shell with all parameters",
			shellType:   ShellBash,
			path:        "/bin/bash",
			version:     "5.0.0",
			expectError: false,
		},
		{
			name:        "empty path should still work",
			shellType:   ShellZsh,
			path:        "",
			version:     "5.8",
			expectError: false,
		},
		{
			name:        "empty version should still work",
			shellType:   ShellFish,
			path:        "/usr/bin/fish",
			version:     "",
			expectError: false,
		},
		{
			name:        "invalid shell type",
			shellType:   ShellType("invalid"),
			path:        "/bin/invalid",
			version:     "1.0",
			expectError: true,
			errorMsg:    "unsupported shell type",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			shell, err := NewShell(tc.shellType, tc.path, tc.version)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(shell)
				if tc.errorMsg != "" {
					s.Contains(err.Error(), tc.errorMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(shell)
				s.Equal(tc.shellType, shell.Type())
				s.Equal(tc.path, shell.Path())
				s.Equal(tc.version, shell.Version())
			}
		})
	}
}

func (s *ShellTestSuite) TestInferShellTypeFromPath() {
	testCases := []struct {
		name          string
		configPath    string
		expectedShell ShellType
		expectError   bool
		errorContains string
	}{
		{
			name:          "infer bash from .bashrc",
			configPath:    "/home/user/.bashrc",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from .bash_profile",
			configPath:    "/home/user/.bash_profile",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from .profile",
			configPath:    "/home/user/.profile",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from custom.bash",
			configPath:    "/home/user/custom.bash",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer bash from my-bash-config",
			configPath:    "/home/user/my-bash-config",
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name:          "infer zsh from .zshrc",
			configPath:    "/home/user/.zshrc",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer zsh from .zprofile",
			configPath:    "/home/user/.zprofile",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer zsh from custom.zsh",
			configPath:    "/home/user/custom.zsh",
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name:          "infer fish from config.fish",
			configPath:    "/home/user/.config/fish/config.fish",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "infer fish from .fishrc",
			configPath:    "/home/user/.fishrc",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "infer fish from path containing fish",
			configPath:    "/home/user/fish/config",
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name:          "return error for unknown config file",
			configPath:    "/home/user/config.txt",
			expectedShell: "",
			expectError:   true,
			errorContains: "cannot infer shell type from path",
		},
		{
			name:          "return error for path without shell indicator",
			configPath:    "/home/user/myconfig",
			expectedShell: "",
			expectError:   true,
			errorContains: "cannot infer shell type from path",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			shellType, err := InferShellTypeFromPath(tc.configPath)

			if tc.expectError {
				s.Require().Error(err)
				s.Equal(ShellType(""), shellType)
				if tc.errorContains != "" {
					s.Contains(err.Error(), tc.errorContains)
				}
			} else {
				s.Require().NoError(err)
				s.Equal(tc.expectedShell, shellType)
			}
		})
	}
}

func (s *ShellTestSuite) TestDetectShellFromEnv() {
	testCases := []struct {
		name          string
		setEnv        func()
		unsetEnv      func()
		expectedShell ShellType
		expectError   bool
		errorMsg      string
	}{
		{
			name: "detect bash from /bin/bash",
			setEnv: func() {
				s.T().Setenv("SHELL", "/bin/bash")
			},
			unsetEnv:      func() {},
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name: "detect bash from /usr/local/bin/bash",
			setEnv: func() {
				s.T().Setenv("SHELL", "/usr/local/bin/bash")
			},
			unsetEnv:      func() {},
			expectedShell: ShellBash,
			expectError:   false,
		},
		{
			name: "detect zsh from /bin/zsh",
			setEnv: func() {
				s.T().Setenv("SHELL", "/bin/zsh")
			},
			unsetEnv:      func() {},
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name: "detect zsh from /usr/bin/zsh",
			setEnv: func() {
				s.T().Setenv("SHELL", "/usr/bin/zsh")
			},
			unsetEnv:      func() {},
			expectedShell: ShellZsh,
			expectError:   false,
		},
		{
			name: "detect fish from /usr/local/bin/fish",
			setEnv: func() {
				s.T().Setenv("SHELL", "/usr/local/bin/fish")
			},
			unsetEnv:      func() {},
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name: "detect fish from /bin/fish",
			setEnv: func() {
				s.T().Setenv("SHELL", "/bin/fish")
			},
			unsetEnv:      func() {},
			expectedShell: ShellFish,
			expectError:   false,
		},
		{
			name: "fail when SHELL not set",
			setEnv: func() {
				s.T().Setenv("SHELL", "")
			},
			unsetEnv:      func() {},
			expectedShell: "",
			expectError:   true,
			errorMsg:      "SHELL environment variable not set",
		},
		{
			name: "fail with unknown shell /bin/sh",
			setEnv: func() {
				s.T().Setenv("SHELL", "/bin/sh")
			},
			unsetEnv:      func() {},
			expectedShell: "",
			expectError:   true,
			errorMsg:      "unsupported shell detected",
		},
		{
			name: "fail with unknown shell /usr/bin/tcsh",
			setEnv: func() {
				s.T().Setenv("SHELL", "/usr/bin/tcsh")
			},
			unsetEnv:      func() {},
			expectedShell: "",
			expectError:   true,
			errorMsg:      "unsupported shell detected",
		},
		{
			name: "case insensitivity for bash path",
			setEnv: func() {
				s.T().Setenv("SHELL", "/usr/local/BASH/Bin/bash")
			},
			unsetEnv:      func() {},
			expectedShell: ShellBash,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setEnv()
			defer tc.unsetEnv()

			shellType, err := DetectShellFromEnv()

			if tc.expectError {
				s.Require().Error(err)
				s.Equal(ShellType(""), shellType)
				if tc.errorMsg != "" {
					s.Contains(err.Error(), tc.errorMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Equal(tc.expectedShell, shellType)
			}
		})
	}
}
