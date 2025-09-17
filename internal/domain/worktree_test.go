package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// WorktreeTestSuite provides test setup for worktree domain tests
type WorktreeTestSuite struct {
	suite.Suite
}

func TestWorktreeSuite(t *testing.T) {
	suite.Run(t, new(WorktreeTestSuite))
}

func (s *WorktreeTestSuite) TestNewWorktree() {
	tests := []struct {
		name         string
		path         string
		branch       string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid worktree",
			path:        "/home/user/workspace/project/feature-branch",
			branch:      "feature-branch",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			branch:       "main",
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
		{
			name:         "empty branch",
			path:         "/valid/path",
			branch:       "",
			expectError:  true,
			errorMessage: "branch name cannot be empty",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			worktree, err := NewWorktree(tt.path, tt.branch)

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
				s.Nil(worktree)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(worktree)
				s.Equal(tt.path, worktree.Path)
				s.Equal(tt.branch, worktree.Branch)
				s.Equal(StatusUnknown, worktree.Status)
				s.False(worktree.LastUpdated.IsZero())
			}
		})
	}
}

func (s *WorktreeTestSuite) TestUpdateStatus() {
	worktree, err := NewWorktree("/test/path", "main")
	s.Require().NoError(err)

	initialTime := worktree.LastUpdated

	// Wait a bit to ensure timestamp changes
	time.Sleep(time.Millisecond)

	err = worktree.UpdateStatus(StatusClean)
	s.Require().NoError(err)

	s.Equal(StatusClean, worktree.Status)
	s.True(worktree.LastUpdated.After(initialTime))
}

func (s *WorktreeTestSuite) TestIsClean() {
	worktree, err := NewWorktree("/test/path", "main")
	s.Require().NoError(err)

	// Initially unknown status
	s.False(worktree.IsClean())

	// Clean status
	err = worktree.UpdateStatus(StatusClean)
	s.Require().NoError(err)
	s.True(worktree.IsClean())

	// Dirty status
	err = worktree.UpdateStatus(StatusDirty)
	s.Require().NoError(err)
	s.False(worktree.IsClean())
}

func (s *WorktreeTestSuite) TestString() {
	worktree, err := NewWorktree("/home/user/project/feature", "feature-branch")
	s.Require().NoError(err)

	result := worktree.String()
	s.Contains(result, "feature-branch")
	s.Contains(result, "/home/user/project/feature")
	s.Contains(result, "unknown")
}

func (s *WorktreeTestSuite) TestEnhancedFeatures() {
	s.Run("should support commit hash tracking", func() {
		worktree, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		// This should fail initially - we need to add Commit field
		err = worktree.SetCommit("abc123def456")
		s.NoError(err)
		s.Equal("abc123def456", worktree.GetCommit())
	})

	s.Run("should validate path existence", func() {
		// Test with non-existent path
		worktree, err := NewWorktree("/non/existent/path", "main")
		s.Require().NoError(err)

		// This should fail initially - we need to add path validation
		isValid, err := worktree.ValidatePathExists()
		s.Error(err)
		s.False(isValid)
	})

	s.Run("should support status aging", func() {
		worktree, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		// Set initial status
		err = worktree.UpdateStatus(StatusClean)
		s.Require().NoError(err)

		// This should fail initially - we need to add status aging
		isStale := worktree.IsStatusStale()
		s.False(isStale) // Should not be stale immediately

		// This should fail initially - we need to add stale threshold configuration
		isStale = worktree.IsStatusStaleWithThreshold(time.Hour)
		s.False(isStale)
	})

	s.Run("should support equality comparison", func() {
		worktree1, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		worktree2, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		worktree3, err := NewWorktree("/different/path", "main")
		s.Require().NoError(err)

		// This should fail initially - we need to add equality methods
		s.True(worktree1.Equals(worktree2))
		s.False(worktree1.Equals(worktree3))
		s.True(worktree1.SameLocationAs(worktree2))
		s.False(worktree1.SameLocationAs(worktree3))
	})

	s.Run("should support worktree metadata", func() {
		worktree, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		// This should fail initially - we need to add metadata support
		worktree.SetMetadata("last-checked-by", "user1")
		worktree.SetMetadata("priority", "high")

		value, exists := worktree.GetMetadata("last-checked-by")
		s.True(exists)
		s.Equal("user1", value)

		value, exists = worktree.GetMetadata("priority")
		s.True(exists)
		s.Equal("high", value)

		_, exists = worktree.GetMetadata("non-existent")
		s.False(exists)
	})

	s.Run("should support worktree health check", func() {
		worktree, err := NewWorktree("/test/path", "main")
		s.Require().NoError(err)

		// This should fail initially - we need to add health check
		health := worktree.GetHealth()
		s.NotNil(health)
		s.Equal("unhealthy", health.Status)
		s.Contains(health.Issues, "path not validated")
	})
}
