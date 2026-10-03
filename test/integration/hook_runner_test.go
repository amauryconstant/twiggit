//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/suite"
)

type HookRunnerIntegrationSuite struct {
	suite.Suite
	runner    *git.HookRunner
	tempDir   string
	configDir string
}

func TestHookRunnerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(HookRunnerIntegrationSuite))
}

func (s *HookRunnerIntegrationSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
	s.configDir = s.T().TempDir()

	executor := git.NewCommandExecutor(30 * time.Second)
	s.runner = git.NewHookRunner(executor)
}

func (s *HookRunnerIntegrationSuite) TestRun_RealConfigFile_ExecutesEchoCommand() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	configContent := `
[hooks.post-create]
commands = ["echo hello"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.True(result.HasExecuted)
	s.True(result.IsSuccessful)
	s.Empty(result.Failures)
}

func (s *HookRunnerIntegrationSuite) TestRun_RealConfigFile_MultipleCommandsSucceed() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	configContent := `
[hooks.post-create]
commands = [
    "echo first",
    "echo second",
    "echo third",
]
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.True(result.HasExecuted)
	s.True(result.IsSuccessful)
	s.Empty(result.Failures)
}

func (s *HookRunnerIntegrationSuite) TestRun_RealConfigFile_CommandFailure_ContinuesExecution() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	configContent := `
[hooks.post-create]
commands = [
    "echo first",
    "sh -c 'exit 1'",
    "echo third",
]
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.True(result.HasExecuted)
	s.False(result.IsSuccessful)
	s.Len(result.Failures, 1)
	s.Contains(result.Failures[0].Command, "exit 1")
	s.Equal(1, result.Failures[0].ExitCode)
}

func (s *HookRunnerIntegrationSuite) TestRun_RealConfigFile_EnvironmentVariablesAvailable() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	configContent := `
[hooks.post-create]
commands = ["printenv TWIGGIT_PROJECT_NAME"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ProjectName:    "test-project",
		BranchName:     "feature-branch",
		SourceBranch:   "main",
		MainRepoPath:   "/path/to/main/repo",
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.True(result.HasExecuted)
	s.True(result.IsSuccessful, "Command should succeed")
}

func (s *HookRunnerIntegrationSuite) TestRun_RealConfigFile_CommandExecutesInWorktreeDirectory() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	testFile := filepath.Join(s.tempDir, "marker.txt")
	configContent := `
[hooks.post-create]
commands = ["touch marker.txt"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.True(result.HasExecuted, "Hooks should have executed")
	s.True(result.IsSuccessful, "Command should succeed, failures: %v", result.Failures)

	_, err = os.Stat(testFile)
	s.Require().NoError(err, "File should have been created in worktree directory")
}

func (s *HookRunnerIntegrationSuite) TestRun_NoConfigFile_ReturnsNotExecuted() {
	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: filepath.Join(s.configDir, "nonexistent.toml"),
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.False(result.HasExecuted)
	s.True(result.IsSuccessful)
}

func (s *HookRunnerIntegrationSuite) TestRun_EmptyHooksSection_ReturnsNotExecuted() {
	configPath := filepath.Join(s.configDir, ".twiggit.toml")
	configContent := `
[other-section]
key = "value"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	s.Require().NoError(err)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   s.tempDir,
		ConfigFilePath: configPath,
	}

	result, err := s.runner.Run(s.T().Context(), req)

	s.Require().NoError(err)
	s.False(result.HasExecuted)
	s.True(result.IsSuccessful)
}
