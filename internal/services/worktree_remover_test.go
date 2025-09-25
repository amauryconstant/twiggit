package services

import (
	"context"
	"errors"
	"testing"

	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/suite"
)

// WorktreeRemoverTestSuite provides suite setup for worktree remover tests
type WorktreeRemoverTestSuite struct {
	suite.Suite
	ctx           context.Context
	mockGitClient *mocks.GitClientMock
	remover       *WorktreeRemover
}

// SetupTest initializes test dependencies for each test
func (s *WorktreeRemoverTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockGitClient = new(mocks.GitClientMock)
	s.remover = NewWorktreeRemover(s.mockGitClient)
}

// TestWorktreeRemover_Remove tests worktree removal with table-driven approach
func (s *WorktreeRemoverTestSuite) TestWorktreeRemover_Remove() {
	testCases := []struct {
		name                  string
		worktreePath          string
		force                 bool
		setupMockExpectations func()
		expectError           bool
		errorMessage          string
	}{
		{
			name:         "should remove worktree successfully without force",
			worktreePath: "/worktree/path",
			force:        false,
			setupMockExpectations: func() {
				s.mockGitClient.On("GetRepositoryRoot", s.ctx, "/worktree/path").
					Return("/project/path", nil)
				s.mockGitClient.On("HasUncommittedChanges", s.ctx, "/worktree/path").
					Return(false)
				s.mockGitClient.On("RemoveWorktree", s.ctx, "/project/path", "/worktree/path", false).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:         "should remove worktree successfully with force",
			worktreePath: "/worktree/path",
			force:        true,
			setupMockExpectations: func() {
				s.mockGitClient.On("GetRepositoryRoot", s.ctx, "/worktree/path").
					Return("/project/path", nil)
				s.mockGitClient.On("RemoveWorktree", s.ctx, "/project/path", "/worktree/path", true).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:         "should return error when worktree path is empty",
			worktreePath: "",
			force:        false,
			setupMockExpectations: func() {
				// No mock expectations - should fail before any calls
			},
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
		{
			name:         "should return error when getting repository root fails",
			worktreePath: "/worktree/path",
			force:        false,
			setupMockExpectations: func() {
				s.mockGitClient.On("GetRepositoryRoot", s.ctx, "/worktree/path").
					Return("", errors.New("not a git repository"))
			},
			expectError:  true,
			errorMessage: "failed to get repository root",
		},
		{
			name:         "should return error when worktree has uncommitted changes and force is false",
			worktreePath: "/worktree/path",
			force:        false,
			setupMockExpectations: func() {
				s.mockGitClient.On("GetRepositoryRoot", s.ctx, "/worktree/path").
					Return("/project/path", nil)
				s.mockGitClient.On("HasUncommittedChanges", s.ctx, "/worktree/path").
					Return(true)
			},
			expectError:  true,
			errorMessage: "uncommitted changes detected",
		},
		{
			name:         "should return error when remove worktree fails",
			worktreePath: "/worktree/path",
			force:        false,
			setupMockExpectations: func() {
				s.mockGitClient.On("GetRepositoryRoot", s.ctx, "/worktree/path").
					Return("/project/path", nil)
				s.mockGitClient.On("HasUncommittedChanges", s.ctx, "/worktree/path").
					Return(false)
				s.mockGitClient.On("RemoveWorktree", s.ctx, "/project/path", "/worktree/path", false).
					Return(errors.New("git worktree remove failed"))
			},
			expectError:  true,
			errorMessage: "failed to remove worktree",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Reset mocks for each test case
			s.mockGitClient = new(mocks.GitClientMock)
			s.remover = NewWorktreeRemover(s.mockGitClient)

			// Setup mock expectations
			tt.setupMockExpectations()

			// Execute
			err := s.remover.Remove(s.ctx, tt.worktreePath, tt.force)

			// Verify
			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
			} else {
				s.Require().NoError(err)
			}

			// Verify mock expectations
			s.mockGitClient.AssertExpectations(s.T())
		})
	}
}

// TestWorktreeRemover_NewWorktreeRemover tests constructor
func (s *WorktreeRemoverTestSuite) TestWorktreeRemover_NewWorktreeRemover() {
	// Create fresh mocks for this test
	mockGitClient := new(mocks.GitClientMock)
	remover := NewWorktreeRemover(mockGitClient)

	s.Require().NotNil(remover, "WorktreeRemover should not be nil")
	s.Equal(mockGitClient, remover.gitClient, "gitClient should be set correctly")
}

// Test suite entry point
func TestWorktreeRemoverSuite(t *testing.T) {
	suite.Run(t, new(WorktreeRemoverTestSuite))
}
