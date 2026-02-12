package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestDefaultConfig() {
	config := DefaultConfig()

	s.Require().NotNil(config)
	s.NotEmpty(config.ProjectsDirectory)
	s.NotEmpty(config.WorktreesDirectory)
	s.Equal("main", config.DefaultSourceBranch)
}

func (s *ConfigTestSuite) TestValidate() {
	s.Run("valid configuration", func() {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		s.NoError(err)
	})

	s.Run("invalid projects directory", func() {
		config := &Config{
			ProjectsDirectory:   "relative/path",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		s.Require().Error(err)
		s.Contains(err.Error(), "projects_directory must be absolute path")
	})

	s.Run("invalid worktrees directory", func() {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "relative/path",
			DefaultSourceBranch: "main",
		}

		err := config.Validate()
		s.Require().Error(err)
		s.Contains(err.Error(), "worktrees_directory must be absolute path")
	})

	s.Run("empty default source branch", func() {
		config := &Config{
			ProjectsDirectory:   "/valid/projects",
			WorktreesDirectory:  "/valid/worktrees",
			DefaultSourceBranch: "",
		}

		err := config.Validate()
		s.Require().Error(err)
		s.Contains(err.Error(), "default_source_branch cannot be empty")
	})

	s.Run("multiple validation errors", func() {
		config := &Config{
			ProjectsDirectory:   "relative/path",
			WorktreesDirectory:  "another/relative/path",
			DefaultSourceBranch: "",
		}

		err := config.Validate()
		s.Require().Error(err)
		s.Contains(err.Error(), "config validation failed")
		// Should contain all validation errors
		s.Contains(err.Error(), "projects_directory must be absolute path")
		s.Contains(err.Error(), "worktrees_directory must be absolute path")
		s.Contains(err.Error(), "default_source_branch cannot be empty")
	})
}
