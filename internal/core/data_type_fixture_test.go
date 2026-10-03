package core

import "testing"

// TestDataTypeFixture exercises the zero value of every data type exposed
// by the core-types spec § Requirement: Compile-time fixture exercises every
// renamed data type. A rename that drops one of these types without
// fixture update surfaces as a compile error local to the failing subtest.
func TestDataTypeFixture(t *testing.T) {
	t.Run("Repository", func(t *testing.T) { _ = Repository{} })
	t.Run("Branch", func(t *testing.T) { _ = Branch{} })
	t.Run("Worktree", func(t *testing.T) { _ = Worktree{} })
	t.Run("Commit", func(t *testing.T) { _ = Commit{} })
	t.Run("Remote", func(t *testing.T) { _ = Remote{} })
	t.Run("RepoDir", func(t *testing.T) { _ = RepoDir{} })
}
