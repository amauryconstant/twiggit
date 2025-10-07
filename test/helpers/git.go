package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitTestHelper provides functional git repository testing utilities
type GitTestHelper struct {
	t           *testing.T
	baseDir     string
	commitCount int
	branch      string
}

// NewGitTestHelper creates a new GitTestHelper instance
func NewGitTestHelper(t *testing.T) *GitTestHelper {
	t.Helper()
	return &GitTestHelper{
		t:       t,
		baseDir: t.TempDir(),
	}
}

// WithCommits sets the commit count for functional composition
func (h *GitTestHelper) WithCommits(count int) *GitTestHelper {
	h.commitCount = count
	return h
}

// WithBranch sets the branch name for functional composition
func (h *GitTestHelper) WithBranch(branch string) *GitTestHelper {
	h.branch = branch
	return h
}

// CreateRepoWithCommits creates a git repository with the specified number of commits
func (h *GitTestHelper) CreateRepoWithCommits(commitCount int) string {
	if commitCount < 0 {
		panic("commit count cannot be negative")
	}

	repoPath := filepath.Join(h.baseDir, "repo")
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		h.t.Fatalf("Failed to create repo: %v", err)
	}

	// Pure function to create commits
	createCommits := func(repo *git.Repository, count int) error {
		wt, err := repo.Worktree()
		if err != nil {
			return err
		}

		for i := 0; i < count; i++ {
			filename := filepath.Join(repoPath, "file.txt")
			content := []byte(fmt.Sprintf("Content %d\n", i))

			if err := os.WriteFile(filename, content, 0644); err != nil {
				return err
			}

			_, err = wt.Add("file.txt")
			if err != nil {
				return err
			}

			commit := &object.Commit{
				Message: fmt.Sprintf("Commit %d", i),
				Author: object.Signature{
					Name:  "Test User",
					Email: "test@example.com",
				},
			}

			_, err = wt.Commit(commit.Message, &git.CommitOptions{
				Author: &commit.Author,
			})
			if err != nil {
				return err
			}
		}

		return nil
	}

	if err := createCommits(repo, commitCount); err != nil {
		h.t.Fatalf("Failed to create commits: %v", err)
	}

	// Create branch if specified
	if h.branch != "" && h.branch != "main" {
		if err := h.CreateBranch(repoPath, h.branch); err != nil {
			h.t.Fatalf("Failed to create branch: %v", err)
		}
	}

	return repoPath
}

// PlainOpen opens a git repository at the given path
func (h *GitTestHelper) PlainOpen(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

// CreateBranch creates a new branch in the repository
func (h *GitTestHelper) CreateBranch(repoPath, branchName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Create branch reference
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	branchRef := plumbing.NewHashReference(branchRefName, headRef.Hash())

	if err := repo.Storer.SetReference(branchRef); err != nil {
		return fmt.Errorf("failed to create branch reference: %w", err)
	}

	return nil
}

// ListBranches returns a list of all branch names in the repository
func (h *GitTestHelper) ListBranches(repoPath string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branchNames []string
	if err := branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchNames = append(branchNames, ref.Name().Short())
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branchNames, nil
}

// CreateShallowClone creates a shallow clone of a repository
func (h *GitTestHelper) CreateShallowClone(sourcePath, destPath string, depth int) error {
	repo, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL:          sourcePath,
		Depth:        depth,
		SingleBranch: true,
		NoCheckout:   false,
	})
	if err != nil {
		return fmt.Errorf("failed to create shallow clone: %w", err)
	}
	_ = repo
	return nil
}

// CreateDetachedHEAD creates a detached HEAD state in the repository
func (h *GitTestHelper) CreateDetachedHEAD(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current HEAD
	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Create worktree in detached HEAD state
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout to the commit hash (detached HEAD)
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash:   headRef.Hash(),
		Branch: plumbing.ReferenceName(""), // Empty branch name creates detached HEAD
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("failed to create detached HEAD: %w", err)
	}

	return nil
}
