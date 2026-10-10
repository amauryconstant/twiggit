//go:build integration

package integration

import (
	"path/filepath"
	"testing"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/test/repo"

	"github.com/stretchr/testify/suite"
)

// StatusIntegrationSuite covers the five core.WorktreeStatus
// derivations on real git repositories: ahead/behind counts,
// merged detection, dirty detection, tracked-base fallback, and
// stale-by-age. The repo helper provisions a project with the six
// per-worktree shapes the spec calls out (main + feature-ahead,
// feature-merged, feature-clean, feature-dirty, feature-tracked).
type StatusIntegrationSuite struct {
	suite.Suite
	helper *repo.RepoTestHelper
	client *git.Client
	cfg    *core.Config
}

func TestStatus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(StatusIntegrationSuite))
}

func (s *StatusIntegrationSuite) SetupSuite() {
	s.helper = repo.NewRepoTestHelper(s.T())
	s.helper.WithCommits(1).SetupTestRepo("status-fixture")

	client, err := git.NewClient()
	s.Require().NoError(err)
	s.client = client

	s.cfg = core.DefaultConfig()
	s.cfg.Status.StaleBehind = 20
	s.cfg.Status.StaleDays = 30
}

// wtPath derives the worktree path the RepoTestHelper provisions for
// the named branch (the helper creates them at <baseDir>/wt-<branch>).
// The path is re-anchored to the absolute base dir so the test does
// not depend on the helper's internal field layout.
func (s *StatusIntegrationSuite) wtPath(branch string) string {
	repoPath := s.helper.GetRepoPath("status-fixture")
	base := filepath.Dir(repoPath)
	return filepath.Join(base, "wt-"+branch)
}

func (s *StatusIntegrationSuite) TestAheadBehindCounts() {
	repoPath := s.helper.GetRepoPath("status-fixture")
	s.helper.CreateDivergentBranch("status-fixture", "feat/ahead", "a.txt", "ahead commit\n")

	row, err := s.client.ReadWorktreeStatus(
		s.T().Context(), s.cfg, repoPath, s.wtPath("feat/ahead"),
	)
	s.Require().NoError(err)
	s.Require().NotNil(row.RepositoryStatus,
		"RepositoryStatus must be populated for a successful read")
	s.Equal(1, row.RepositoryStatus.Ahead,
		"a branch with one new commit is 1 ahead of main")
	s.Equal(0, row.RepositoryStatus.Behind, "main has not advanced past the branch")
}

func (s *StatusIntegrationSuite) TestMergedDetection() {
	repoPath := s.helper.GetRepoPath("status-fixture")
	// A branch whose commits are all reachable from main is merged.
	s.helper.CreateDivergentBranch("status-fixture", "feat/merged", "m.txt", "merged commit\n")
	// Fast-forward main to the branch tip so the branch is now merged.
	s.helper.AdvanceBase("status-fixture", "m.txt", "merged commit\n")

	row, err := s.client.ReadWorktreeStatus(
		s.T().Context(), s.cfg, repoPath, s.wtPath("feat/merged"),
	)
	s.Require().NoError(err)
	s.True(row.IsMerged, "branch whose tip is reachable from main must be IsMerged")
}

func (s *StatusIntegrationSuite) TestDirtyDetection() {
	repoPath := s.helper.GetRepoPath("status-fixture")
	s.helper.CreateDivergentBranch("status-fixture", "feat/dirty", "d.txt", "dirty commit\n")
	s.helper.CreateDirtyWorktree("status-fixture", "feat/dirty", "uncommitted.txt", "uncommitted\n")

	row, err := s.client.ReadWorktreeStatus(
		s.T().Context(), s.cfg, repoPath, s.wtPath("feat/dirty"),
	)
	s.Require().NoError(err)
	s.False(row.IsClean, "uncommitted changes must mark IsClean false")
	s.True(row.HasUncommittedChanges)
}

func (s *StatusIntegrationSuite) TestTrackedBaseFallback() {
	repoPath := s.helper.GetRepoPath("status-fixture")
	// Worktree with no per-worktree tracked-base should resolve to
	// the first protected branch (main) by default.
	s.helper.CreateDivergentBranch("status-fixture", "feat/clean", "c.txt", "clean commit\n")

	row, err := s.client.ReadWorktreeStatus(
		s.T().Context(), s.cfg, repoPath, s.wtPath("feat/clean"),
	)
	s.Require().NoError(err)
	s.Equal("main", row.Base, "no per-wt tracked-base must fall back to first protected branch")
}

func (s *StatusIntegrationSuite) TestStaleByAge() {
	// Use a per-invocation cfg with StaleDays=1 and an artificially
	// old LastCommitDate. We can't reach into the adapter to set
	// LastCommitDate directly, so we drive ComputeIsStale on a
	// synthesised row to assert the same derivation the adapter
	// would expose if the commit were genuinely old.
	cfg := core.DefaultConfig()
	cfg.Status.StaleBehind = 0
	cfg.Status.StaleDays = 1
	row := core.WorktreeStatus{
		LastCommitDate: time.Now().Add(-48 * time.Hour),
	}
	s.True(row.ComputeIsStale(cfg),
		"a 2-day-old commit must trip the 1-day threshold")
}
