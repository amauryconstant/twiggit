package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorktreeTestSuite struct {
	suite.Suite
}

func TestWorktreeSuite(t *testing.T) {
	suite.Run(t, new(WorktreeTestSuite))
}

func (s *WorktreeTestSuite) TestNewWorktree() {
	testCases := []struct {
		name         string
		path         string
		branch       string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid worktree",
			path:        "/home/user/Worktrees/project/feature",
			branch:      "feature",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			branch:       "feature",
			expectError:  true,
			errorMessage: "new worktree: path cannot be empty",
		},
		{
			name:         "empty branch",
			path:         "/home/user/Worktrees/project/feature",
			branch:       "",
			expectError:  true,
			errorMessage: "new worktree: branch cannot be empty",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			worktree, err := NewWorktree(tc.path, tc.branch)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(worktree)
			} else {
				s.Require().NoError(err)
				s.NotNil(worktree)
				s.Equal(tc.path, worktree.Path())
				s.Equal(tc.branch, worktree.Branch())
			}
		})
	}
}
