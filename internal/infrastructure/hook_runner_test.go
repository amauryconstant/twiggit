package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type HookRunnerTestSuite struct {
	suite.Suite
	runner     HookRunner
	mockExec   *MockCommandExecutor
	tempDir    string
	configPath string
}

func TestHookRunnerSuite(t *testing.T) {
	suite.Run(t, new(HookRunnerTestSuite))
}

func (s *HookRunnerTestSuite) SetupTest() {
	s.mockExec = NewMockCommandExecutor()
	s.runner = NewHookRunner(s.mockExec)

	s.tempDir = s.T().TempDir()
	s.configPath = filepath.Join(s.tempDir, ".twiggit.toml")
}

func (s *HookRunnerTestSuite) TestRun_NoConfigFile_ReturnsNotExecuted() {
	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   "/tmp/worktree",
		ConfigFilePath: "/nonexistent/.twiggit.toml",
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.False(result.Executed)
	s.True(result.Success)
	s.Nil(result.Failures)
}

func (s *HookRunnerTestSuite) TestRun_EmptyConfigFile_ReturnsNotExecuted() {
	err := os.WriteFile(s.configPath, []byte(""), 0644)
	s.Require().NoError(err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.False(result.Executed)
	s.True(result.Success)
}

func (s *HookRunnerTestSuite) TestRun_ConfigWithCommands_ExecutesCommands() {
	configContent := `
[hooks.post-create]
commands = ["mise trust", "npm install"]
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	s.mockExec.On("ExecuteWithTimeout",
		mock.Anything, s.tempDir, "sh", s.defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Twice()

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ProjectName:    "test-project",
		BranchName:     "feature",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.True(result.Executed)
	s.True(result.Success)
	s.Empty(result.Failures)
	s.mockExec.AssertExpectations(s.T())
}

func (s *HookRunnerTestSuite) TestRun_CommandFailure_ContinuesAndCollectsFailures() {
	configContent := `
[hooks.post-create]
commands = ["mise trust", "npm install", "echo done"]
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	s.mockExec.On("ExecuteWithTimeout",
		mock.Anything, s.tempDir, "sh", s.defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()

	s.mockExec.On("ExecuteWithTimeout",
		mock.Anything, s.tempDir, "sh", s.defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 1, Stdout: "npm error", Stderr: ""}, nil).Once()

	s.mockExec.On("ExecuteWithTimeout",
		mock.Anything, s.tempDir, "sh", s.defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.True(result.Executed)
	s.False(result.Success)
	s.Len(result.Failures, 1)
	s.Equal("npm install", result.Failures[0].Command)
	s.Equal(1, result.Failures[0].ExitCode)
}

func (s *HookRunnerTestSuite) TestRun_MalformedTOML_LogsWarningAndReturnsNotExecuted() {
	configContent := `
[hooks.post-create
commands = ["mise trust"]
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.False(result.Executed)
	s.True(result.Success)
}

func (s *HookRunnerTestSuite) TestRun_MissingCommandsArray_ReturnsNotExecuted() {
	configContent := `
[hooks.post-create]
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.False(result.Executed)
	s.True(result.Success)
}

func (s *HookRunnerTestSuite) TestRun_EmptyCommandsArray_ReturnsNotExecuted() {
	configContent := `
[hooks.post-create]
commands = []
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: s.configPath,
	}

	result, err := s.runner.Run(context.Background(), req)

	s.Require().NoError(err)
	s.False(result.Executed)
	s.True(result.Success)
}

func (s *HookRunnerTestSuite) TestRun_EnvironmentVariablesSet() {
	configContent := `
[hooks.post-create]
commands = ["echo test"]
`
	err := os.WriteFile(s.configPath, []byte(configContent), 0644)
	s.Require().NoError(err)

	var capturedArgs []string
	s.mockExec.On("ExecuteWithTimeout",
		mock.Anything, "/worktree/path", "sh", s.defaultTimeout(), mock.AnythingOfType("[]string"),
	).Run(func(args mock.Arguments) {
		capturedArgs = args.Get(4).([]string)
	}).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   "/worktree/path",
		ProjectName:    "my-project",
		BranchName:     "feature-branch",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: s.configPath,
	}

	_, err = s.runner.Run(context.Background(), req)
	s.Require().NoError(err)

	s.Require().Len(capturedArgs, 2)
	fullCmd := capturedArgs[1]

	s.Contains(fullCmd, "TWIGGIT_WORKTREE_PATH")
	s.Contains(fullCmd, "/worktree/path")
	s.Contains(fullCmd, "TWIGGIT_PROJECT_NAME")
	s.Contains(fullCmd, "my-project")
	s.Contains(fullCmd, "TWIGGIT_BRANCH_NAME")
	s.Contains(fullCmd, "feature-branch")
	s.Contains(fullCmd, "TWIGGIT_SOURCE_BRANCH")
	s.Contains(fullCmd, "main")
	s.Contains(fullCmd, "TWIGGIT_MAIN_REPO_PATH")
	s.Contains(fullCmd, "/repo/main")
}

func (s *HookRunnerTestSuite) defaultTimeout() time.Duration {
	return 30 * time.Second
}
