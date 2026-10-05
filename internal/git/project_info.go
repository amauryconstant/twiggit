package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"twiggit/internal/core"
)

// ProjectInfoFromGitDir returns the *core.ProjectInfo for the supplied
// gitDir. The path is validated as a git repo, then walked upward to
// locate the main-repo root (the directory whose .git is a directory,
// not a worktree pointer file). The returned ProjectInfo carries
// defensive copies of all slice fields (Worktrees, Branches, Remotes)
// so callers cannot mutate the underlying repo state by writing
// through the returned slice headers.
//
// Accepts the composite *Client (which embeds *reader and *cliClient)
// directly per the cmdutil.Factory depguard boundary; future readers
// can narrow via role-typed locals without changing this signature.
//
// Behaviour:
//
//   - Empty gitDir: returns *core.OperationError{Op: "project.info"}.
//   - Invalid git repository at gitDir: returns *core.OperationError
//     with Cause=core.ErrGitRepoNotFound.
//   - Repository() failure: falls back to a zero-value Repository
//     anchored at the resolved mainRepoPath so the slice fields stay
//     populated from the partial result.
//   - ListWorktrees() failure: returns an empty Worktrees slice
//     rather than failing the whole projection; callers branch on
//     ProjectInfo.Worktrees == nil to detect the degraded case.
func ProjectInfoFromGitDir(ctx context.Context, client *Client, gitDir string) (*core.ProjectInfo, error) {
	if gitDir == "" {
		return nil, &core.OperationError{
			Op:      "project.info",
			Entity:  gitDir,
			Message: "gitDir cannot be empty",
		}
	}
	if client == nil {
		return nil, errors.New("git: ProjectInfoFromGitDir: nil client")
	}

	if err := client.ValidateRepository(gitDir); err != nil {
		return nil, &core.OperationError{
			Op:      "project.info",
			Entity:  gitDir,
			Message: "invalid git repository",
			Cause:   err,
		}
	}

	mainRepoPath := gitDir
	if resolved := FindMainRepoByTraversal(gitDir); resolved != "" {
		mainRepoPath = resolved
	}

	repoInfo, err := client.Repository(ctx, mainRepoPath)
	if err != nil {
		// Partial: the repo exists (ValidateRepository passed) but
		// metadata fetch failed. Surface a zero-anchored Repository
		// so callers still get a populated ProjectInfo.
		repoInfo = &core.Repository{Path: mainRepoPath}
	}

	worktrees, _ := client.ListWorktrees(ctx, mainRepoPath)
	worktreePtrs := make([]*core.Worktree, len(worktrees))
	for i := range worktrees {
		wt := worktrees[i]
		worktreePtrs[i] = &wt
	}

	branchPtrs := make([]*core.Branch, len(repoInfo.Branches))
	for i := range repoInfo.Branches {
		br := repoInfo.Branches[i]
		branchPtrs[i] = &br
	}

	remotePtrs := make([]*core.Remote, len(repoInfo.Remotes))
	for i := range repoInfo.Remotes {
		rm := repoInfo.Remotes[i]
		remotePtrs[i] = &rm
	}

	return &core.ProjectInfo{
		Name:          filepath.Base(mainRepoPath),
		Path:          gitDir,
		GitRepoPath:   mainRepoPath,
		Worktrees:     worktreePtrs,
		Branches:      branchPtrs,
		Remotes:       remotePtrs,
		DefaultBranch: repoInfo.DefaultBranch,
		IsBare:        repoInfo.IsBare,
	}, nil
}

// MustProjectInfo is the panic-on-error wrapper for callers in
// initialisation paths where a missing repo is a programmer bug.
// Production callers should use ProjectInfoFromGitDir and handle
// the error explicitly.
func MustProjectInfo(info *core.ProjectInfo, err error) *core.ProjectInfo {
	if err != nil {
		panic(fmt.Sprintf("git: MustProjectInfo: %v", err))
	}
	return info
}
