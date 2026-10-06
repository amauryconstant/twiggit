package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"
)

// rebaseTarget bundles the per-worktree inputs the walk needs. The
// pre-walk fan-out (discoverProject / ListWorktrees) populates the
// fields once; the walk then iterates the slice without re-doing the
// context detection.
type rebaseTarget struct {
	Project      *core.ProjectInfo
	BranchName   string
	WorktreePath string
}

// resolveRebaseTargets produces the slice of worktrees the rebase
// walk should visit. The priority chain mirrors cmd/prune.go's
// resolvePruneProjects: --all fans out across the project's
// worktrees, a positional [project/branch] narrows to one entry, the
// current context picks a single worktree.
func resolveRebaseTargets(ctx context.Context, gitClient *git.Client, cfg *core.Config, opts *RebaseOptions, currentCtx *core.Context) ([]rebaseTarget, error) {
	if opts.IsAll {
		project, err := pickAllProject(ctx, gitClient, cfg, currentCtx)
		if err != nil {
			return nil, err
		}
		return fanOutProject(ctx, gitClient, project)
	}

	if opts.Target != "" {
		parts := strings.SplitN(opts.Target, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, core.NewUsageError("rebase: target must be in <project>/<branch> format", nil)
		}
		projectPath := filepath.Join(cfg.ProjectsDirectory, parts[0])
		project, err := discoverProject(ctx, gitClient, cfg, parts[0], currentCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
		_ = projectPath
		return []rebaseTarget{{
			Project:      project,
			BranchName:   parts[1],
			WorktreePath: filepath.Join(cfg.WorktreesDirectory, parts[0], parts[1]),
		}}, nil
	}

	// Single-target: current worktree if any, else current project
	// (the main repo). The walk calls Rebase against either path.
	if currentCtx != nil {
		switch currentCtx.Type {
		case core.ContextWorktree:
			return []rebaseTarget{{
				BranchName:   currentCtx.BranchName,
				WorktreePath: currentCtx.Path,
			}}, nil
		case core.ContextProject:
			mainRepo := git.FindMainRepoByTraversal(currentCtx.Path)
			if mainRepo == "" {
				return nil, core.NewUsageError("rebase: cannot locate main repository from project context", nil)
			}
			return []rebaseTarget{{
				BranchName:   "main",
				WorktreePath: mainRepo,
			}}, nil
		}
	}
	return nil, core.NewUsageError("rebase: not inside a git context; provide a [project/branch] argument or use --all", nil)
}

// pickAllProject chooses the project whose worktrees --all should
// fan out across. With a current project context we use that; with
// --all outside git we use the only project in the projects dir.
func pickAllProject(ctx context.Context, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.ProjectInfo, error) {
	if currentCtx != nil && (currentCtx.Type == core.ContextProject || currentCtx.Type == core.ContextWorktree) {
		project, err := discoverProject(ctx, gitClient, cfg, "", currentCtx)
		if err == nil {
			return project, nil
		}
	}
	// Outside git: list projects dir, error if more than one.
	dirEntries, err := os.ReadDir(cfg.ProjectsDirectory)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "rebase.all",
			Entity:  cfg.ProjectsDirectory,
			Message: "failed to list projects",
			Cause:   err,
		}
	}
	names := make([]string, 0, len(dirEntries))
	for _, e := range dirEntries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return nil, core.NewUsageError("rebase --all: no projects found", nil)
	}
	if len(names) > 1 {
		return nil, core.NewUsageError("rebase --all: multiple projects; run from inside a project context", nil)
	}
	return discoverProject(ctx, gitClient, cfg, names[0], nil)
}

// fanOutProject lists the worktrees in the project (excluding the
// main repo) and returns one rebaseTarget per worktree.
func fanOutProject(ctx context.Context, gitClient *git.Client, project *core.ProjectInfo) ([]rebaseTarget, error) {
	wts, err := gitClient.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "rebase.list-worktrees",
			Entity:  project.GitRepoPath,
			Message: "failed to list worktrees",
			Cause:   err,
		}
	}
	targets := make([]rebaseTarget, 0, len(wts))
	for _, wt := range wts {
		if wt.Path == project.GitRepoPath {
			continue
		}
		targets = append(targets, rebaseTarget{
			Project:      project,
			BranchName:   wt.Branch,
			WorktreePath: wt.Path,
		})
	}
	return targets, nil
}

// resolveTrackedBase picks the base branch a worktree should rebase
// onto. The priority chain: explicit --set-base value (already
// written by runRebaseSetBase; here we just read) > per-worktree
// twiggit.tracked-base > first entry of
// Config.Validation.ProtectedBranches > ErrBaseNotSet. An empty
// fallback list with no tracked base is the only condition that
// surfaces ErrBaseNotSet.
func resolveTrackedBase(ctx context.Context, gitClient *git.Client, cfg *core.Config, wtPath string) (string, error) {
	base, err := gitClient.GetTrackedBase(ctx, wtPath)
	if err != nil {
		return "", &core.OperationError{
			Op:      "rebase.tracked-base",
			Entity:  wtPath,
			Message: "failed to read tracked base",
			Cause:   err,
		}
	}
	if base != "" {
		return base, nil
	}
	if len(cfg.Validation.ProtectedBranches) > 0 {
		return cfg.Validation.ProtectedBranches[0], nil
	}
	return "", &core.OperationError{
		Op:      "rebase.tracked-base",
		Entity:  wtPath,
		Message: "no tracked base and no fallback",
		Cause:   core.ErrBaseNotSet,
	}
}

// dispatchRebaseHooks runs the pre-rebase hooks for a target. Returns
// the *core.HookResult and a bool indicating whether a PreRebase
// failure must abort the rebase (always true for PreRebase since the
// runner already short-circuits on failure).
func dispatchRebaseHooks(ctx context.Context, logger *slog.Logger, cfg *core.Config, target rebaseTarget, base string) (*core.HookResult, bool) {
	runner := git.NewHookRunner(git.NewCommandExecutor(time.Duration(cfg.Shell.HookTimeout)*time.Second), cfg.Shell.HookTimeout)
	req := &core.HookRunRequest{
		HookType:     core.HookTypePreRebase,
		WorktreePath: target.WorktreePath,
		BranchName:   target.BranchName,
		MainRepoPath: projectGitRepo(target),
		RebaseBase:   base,
		ConfigFilePath: filepath.Join(
			projectGitRepo(target), ".twiggit.toml",
		),
	}
	result, err := runner.Run(ctx, req)
	if err != nil {
		logger.Warn("pre-rebase hook run failed", "err", err, "worktree_path", target.WorktreePath)
		return nil, true
	}
	if result != nil && !result.IsSuccessful {
		return result, true
	}
	return result, false
}

// dispatchPostRebaseHooks runs the post-rebase hooks for a clean
// rebase. Failures are warnings; the rebase outcome is preserved.
func dispatchPostRebaseHooks(ctx context.Context, logger *slog.Logger, cfg *core.Config, target rebaseTarget, base, oldTip, newTip, outcome string) {
	runner := git.NewHookRunner(git.NewCommandExecutor(time.Duration(cfg.Shell.HookTimeout)*time.Second), cfg.Shell.HookTimeout)
	req := &core.HookRunRequest{
		HookType:       core.HookTypePostRebase,
		WorktreePath:   target.WorktreePath,
		BranchName:     target.BranchName,
		MainRepoPath:   projectGitRepo(target),
		RebaseBase:     base,
		RebaseOldTip:   oldTip,
		RebaseNewTip:   newTip,
		RebaseResult:   outcome,
		ConfigFilePath: filepath.Join(projectGitRepo(target), ".twiggit.toml"),
	}
	result, err := runner.Run(ctx, req)
	if err != nil {
		logger.Warn("post-rebase hook run failed", "err", err, "worktree_path", target.WorktreePath)
		return
	}
	if result != nil && !result.IsSuccessful {
		logger.Warn("post-rebase hook reported failure", "worktree_path", target.WorktreePath)
	}
}

// projectGitRepo returns the main repo path for a rebaseTarget or
// the worktree path itself when the target has no project binding
// (current-context rebase). The empty string is acceptable: the
// hook runner uses ConfigFilePath emptiness to short-circuit.
func projectGitRepo(target rebaseTarget) string {
	if target.Project != nil {
		return target.Project.GitRepoPath
	}
	return target.WorktreePath
}

// runRebaseWalk is the shared inner machinery. It is invoked by
// `twiggit rebase` and by `twiggit sync --rebase` so both verbs see
// the same conflict / hook / output behaviour.
func runRebaseWalk(opts *RebaseOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.RebaseResult, error) {
	targets, err := resolveRebaseTargets(opts.Ctx, gitClient, cfg, opts, currentCtx)
	if err != nil {
		return nil, err
	}

	result := &core.RebaseResult{
		RebasedWorktrees: []*core.RebasedWorktree{},
		SkippedWorktrees: []*core.RebasedWorktree{},
	}

	stopOnConflict := opts.IsAll

	for _, target := range targets {
		entry := &core.RebasedWorktree{
			BranchName:   target.BranchName,
			WorktreePath: target.WorktreePath,
		}
		if target.Project != nil {
			entry.ProjectName = target.Project.Name
		}

		base, err := resolveTrackedBase(opts.Ctx, gitClient, cfg, target.WorktreePath)
		if err != nil {
			entry.Error = err
			entry.SkipReason = "no tracked base"
			result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
			result.TotalSkipped++
			if errors.Is(err, core.ErrBaseNotSet) {
				// surface to caller as a usage error
				return nil, err
			}
			continue
		}
		entry.TrackedBase = base

		// Optional fetch.
		if opts.IsFetch {
			repoPath := projectGitRepo(target)
			if err := gitClient.Fetch(opts.Ctx, repoPath, "origin", base); err != nil {
				entry.Error = err
				entry.SkipReason = "fetch failed"
				result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
				result.TotalSkipped++
				continue
			}
		}

		// Pre-rebase hook gate.
		if _, abort := dispatchRebaseHooks(opts.Ctx, opts.Logger, cfg, target, base); abort {
			entry.SkipReason = "pre-rebase hook failed"
			result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
			result.TotalSkipped++
			continue
		}

		outcome, err := gitClient.Rebase(opts.Ctx, target.WorktreePath, base)
		entry.Outcome = outcome
		entry.Error = err
		switch outcome {
		case core.RebaseOutcomeClean:
			entry.Outcome = outcome
			result.RebasedWorktrees = append(result.RebasedWorktrees, entry)
			result.TotalRebased++
			dispatchPostRebaseHooks(opts.Ctx, opts.Logger, cfg, target, base, "", "", "clean")
		case core.RebaseOutcomeNothingToDo:
			entry.Outcome = outcome
			result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
			result.TotalSkipped++
		case core.RebaseOutcomeConflicted:
			entry.Outcome = outcome
			result.RebasedWorktrees = append(result.RebasedWorktrees, entry)
			result.TotalConflicts++
			if stopOnConflict {
				return result, &core.OperationError{
					Op:      "rebase.worktree",
					Entity:  target.WorktreePath,
					Message: "rebase hit a conflict",
					Cause:   core.ErrRebaseConflict,
				}
			}
		default:
			entry.SkipReason = "rebase failed"
			result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
			result.TotalSkipped++
		}
	}

	// Single-target successful rebase: emit the worktree path for
	// the shell wrapper.
	if len(result.RebasedWorktrees) == 1 &&
		result.RebasedWorktrees[0].Outcome == core.RebaseOutcomeClean &&
		!opts.IsAll {
		result.NavigationPath = result.RebasedWorktrees[0].WorktreePath
	}

	return result, nil
}

// emitRebaseOutput prints the per-worktree summary + totals.
func emitRebaseOutput(opts *RebaseOptions, result *core.RebaseResult, _ *core.Context) {
	if result == nil {
		return
	}
	errOut := writeOrIgnore(opts.IO.ErrOut)
	if len(result.RebasedWorktrees) > 0 {
		_, _ = fmt.Fprintln(errOut, "Rebased:")
		for _, wt := range result.RebasedWorktrees {
			line := fmt.Sprintf("  %s (%s) onto %s", wt.WorktreePath, wt.BranchName, wt.TrackedBase)
			if wt.Outcome == core.RebaseOutcomeConflicted {
				line += " [CONFLICTED]"
			}
			_, _ = fmt.Fprintln(errOut, line)
		}
	}
	if len(result.SkippedWorktrees) > 0 {
		_, _ = fmt.Fprintln(errOut, "Skipped:")
		for _, wt := range result.SkippedWorktrees {
			_, _ = fmt.Fprintf(errOut, "  %s (%s): %s\n", wt.WorktreePath, wt.BranchName, wt.SkipReason)
		}
	}
	_, _ = fmt.Fprintf(errOut, "Summary: %d rebased, %d skipped", result.TotalRebased, result.TotalSkipped)
	if result.TotalConflicts > 0 {
		_, _ = fmt.Fprintf(errOut, ", %d conflicts", result.TotalConflicts)
	}
	_, _ = fmt.Fprintln(errOut)

	if result.NavigationPath != "" {
		_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), result.NavigationPath)
	}
}
