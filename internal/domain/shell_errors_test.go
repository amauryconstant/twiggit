package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShellErrorsTestSuite struct {
	suite.Suite
}

func TestShellErrors(t *testing.T) {
	suite.Run(t, new(ShellErrorsTestSuite))
}

func (s *ShellErrorsTestSuite) TestShellError_Error_WithContextAndCause() {
	cause := errors.New("file not found")
	err := NewShellErrorWithCause(ErrConfigFileNotFound, "bash", "wrapper installation", cause)
	msg := err.Error()

	s.Contains(msg, "wrapper installation")
	s.Contains(msg, "file not found")
}

func (s *ShellErrorsTestSuite) TestShellError_Error_WithContextAndShellType() {
	err := NewShellError(ErrInvalidShellType, "invalid", "shell detection")
	msg := err.Error()

	s.Contains(msg, "shell detection")
	s.Contains(msg, "invalid")
}

func (s *ShellErrorsTestSuite) TestShellError_Error_OnlyContext() {
	err := NewShellError(ErrInferenceFailed, "", "could not infer shell type")
	msg := err.Error()

	s.Equal("could not infer shell type", msg)
}

func (s *ShellErrorsTestSuite) TestShellError_Error_OnlyCause() {
	cause := errors.New("permission denied")
	err := NewShellErrorWithCause(ErrConfigFileNotWritable, "", "", cause)
	msg := err.Error()

	s.Equal("permission denied", msg)
}

func (s *ShellErrorsTestSuite) TestShellError_Error_OnlyShellType() {
	err := NewShellError(ErrInvalidShellType, "fish", "")
	msg := err.Error()

	s.Contains(msg, ErrInvalidShellType)
	s.Contains(msg, "fish")
	s.Equal(ErrInvalidShellType+": fish", msg)
}

func (s *ShellErrorsTestSuite) TestShellError_Error_OnlyCode() {
	err := NewShellError(ErrWrapperGeneration, "", "")
	msg := err.Error()

	s.Contains(msg, ErrWrapperGeneration)
	s.Equal(fmt.Sprintf("shell error [%s]", ErrWrapperGeneration), msg)
}

func (s *ShellErrorsTestSuite) TestShellError_Unwrap() {
	s.Run("with cause returns cause", func() {
		cause := errors.New("underlying error")
		err := NewShellErrorWithCause(ErrConfigFileNotFound, "bash", "context", cause)
		s.Equal(cause, err.Unwrap())
	})

	s.Run("without cause returns nil", func() {
		err := NewShellError(ErrInvalidShellType, "bash", "context")
		s.NoError(err.Unwrap())
	})
}

func (s *ShellErrorsTestSuite) TestShellError_ConstantValues() {
	s.Equal("INVALID_SHELL_TYPE", ErrInvalidShellType)
	s.Equal("SHELL_NOT_INSTALLED", ErrShellNotInstalled)
	s.Equal("SHELL_ALREADY_INSTALLED", ErrShellAlreadyInstalled)
	s.Equal("CONFIG_FILE_NOT_FOUND", ErrConfigFileNotFound)
	s.Equal("CONFIG_FILE_NOT_WRITABLE", ErrConfigFileNotWritable)
	s.Equal("WRAPPER_GENERATION_FAILED", ErrWrapperGeneration)
	s.Equal("WRAPPER_INSTALLATION_FAILED", ErrWrapperInstallation)
	s.Equal("INFERENCE_FAILED", ErrInferenceFailed)
	s.Equal("SHELL_DETECTION_FAILED", ErrShellDetectionFailed)
}

func (s *ShellErrorsTestSuite) TestShellError_ComplexScenarios() {
	s.Run("full error with all fields", func() {
		cause := errors.New("disk full")
		err := NewShellErrorWithCause(ErrWrapperInstallation, "zsh", "installing wrapper", cause)
		msg := err.Error()

		s.Contains(msg, "installing wrapper")
		s.Contains(msg, "disk full")
		s.Equal("zsh", err.ShellType)
		s.Equal(ErrWrapperInstallation, err.Code)
		s.Equal("installing wrapper", err.Context)
	})

	s.Run("error with wrapped ShellError", func() {
		innerCause := errors.New("permission denied")
		innerErr := NewShellErrorWithCause(ErrConfigFileNotWritable, "bash", "writing config", innerCause)
		outerCause := innerErr
		outerErr := NewShellErrorWithCause(ErrWrapperInstallation, "bash", "wrapper installation", outerCause)
		msg := outerErr.Error()

		s.Contains(msg, "wrapper installation")
		s.Contains(msg, "writing config")
	})
}

func (s *ShellErrorsTestSuite) TestShellError_NewShellError_CreatesWithoutCause() {
	err := NewShellError(ErrInvalidShellType, "bash", "shell detection")

	s.Equal(ErrInvalidShellType, err.Code)
	s.Equal("bash", err.ShellType)
	s.Equal("shell detection", err.Context)
	s.NoError(err.Cause)
}

func (s *ShellErrorsTestSuite) TestShellError_NewShellErrorWithCause_CreatesWithCause() {
	cause := errors.New("test error")
	err := NewShellErrorWithCause(ErrConfigFileNotFound, "zsh", "file lookup", cause)

	s.Equal(ErrConfigFileNotFound, err.Code)
	s.Equal("zsh", err.ShellType)
	s.Equal("file lookup", err.Context)
	s.Equal(cause, err.Cause)
}

func (s *ShellErrorsTestSuite) TestShellError_Error_AllBranches() {
	testCases := []struct {
		name        string
		err         *ShellError
		contains    []string
		notContains []string
		expected    string
	}{
		{
			name:     "context + cause",
			err:      NewShellErrorWithCause("CODE", "bash", "context", errors.New("cause msg")),
			contains: []string{"context", "cause msg"},
			expected: "context: cause msg",
		},
		{
			name:     "context + shellType (no cause)",
			err:      NewShellError("CODE", "bash", "context"),
			contains: []string{"context", "bash"},
			expected: "context: bash",
		},
		{
			name:     "context only",
			err:      NewShellError("CODE", "", "context"),
			contains: []string{"context"},
			expected: "context",
		},
		{
			name:     "cause only (no context)",
			err:      NewShellErrorWithCause("CODE", "", "", errors.New("cause msg")),
			contains: []string{"cause msg"},
			expected: "cause msg",
		},
		{
			name:     "shellType only (no context)",
			err:      NewShellError("CODE", "bash", ""),
			contains: []string{"CODE", "bash"},
			expected: "CODE: bash",
		},
		{
			name:     "code only",
			err:      NewShellError("CODE", "", ""),
			contains: []string{"CODE"},
			expected: "shell error [CODE]",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			msg := tc.err.Error()
			s.Equal(tc.expected, msg)

			for _, contain := range tc.contains {
				s.Contains(msg, contain)
			}
			for _, notContain := range tc.notContains {
				s.NotContains(msg, notContain)
			}
		})
	}
}
