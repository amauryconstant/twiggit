package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/carapace-sh/carapace"
)

// getCompletionTimeout returns the completion timeout duration from config, defaulting to 500ms
func getCompletionTimeout(f *CommandConfig) time.Duration {
	if f == nil {
		return 500 * time.Millisecond
	}
	cfg, err := f.Config()
	if err != nil || cfg == nil {
		return 500 * time.Millisecond
	}
	if cfg.Completion.Timeout != "" {
		if duration, err := time.ParseDuration(cfg.Completion.Timeout); err == nil {
			return duration
		}
	}
	return 500 * time.Millisecond
}

// actionWorktreeTarget provides completion for worktree targets (project/branch).
// After slice 9 it shells out to the git client directly to enumerate
// candidates. Slice 10 will rebuild the smart-sort + exclusion logic
// that previously lived in ContextService.GetCompletionSuggestions.
func actionWorktreeTarget(f *CommandConfig, _ ...core.SuggestionOption) carapace.Action {
	timeout := getCompletionTimeout(f)
	return carapace.ActionMultiParts("/", func(c carapace.Context) carapace.Action {
		switch len(c.Parts) {
		case 0:
			return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
				return actionProjectsOrBranches(f)
			}).Timeout(timeout, carapace.ActionValues())
		case 1:
			return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
				return actionBranchesForProject(c.Parts[0], f)
			}).Timeout(timeout, carapace.ActionValues())
		default:
			return carapace.ActionValues()
		}
	}).Cache(5 * time.Second)
}

// actionBranches provides completion for branch names (--source flag).
// Returns the cached list of branch names from the current project.
func actionBranches(f *CommandConfig) carapace.Action {
	timeout := getCompletionTimeout(f)
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		ctx := context.Background()
		branches, err := listBranchesFromContext(ctx, f)
		if err != nil || len(branches) == 0 {
			return carapace.ActionValues()
		}
		values := make([]string, len(branches))
		copy(values, branches)
		return carapace.ActionValues(values...).Tag("")
	}).Timeout(timeout, carapace.ActionValues()).Cache(5 * time.Second)
}

// actionProjectsOrBranches suggests projects based on the projects directory.
// The branch-suggestion fallback for projectless contexts is delegated to
// carapace.ActionValues for now; slice 10 will restore the smart-sort path.
func actionProjectsOrBranches(f *CommandConfig) carapace.Action {
	ctx := context.Background()
	projects, err := listProjects(ctx, f)
	if err != nil || len(projects) == 0 {
		return carapace.ActionValues()
	}
	values := make([]string, len(projects))
	for i, p := range projects {
		values[i] = p
	}
	return carapace.ActionValues(values...).Suffix("/").Tag("")
}

// actionBranchesForProject suggests branches for a specific project.
func actionBranchesForProject(projectName string, f *CommandConfig) carapace.Action {
	timeout := getCompletionTimeout(f)
	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		ctx := context.Background()
		branches, err := listBranchesForProject(ctx, f, projectName)
		if err != nil || len(branches) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValues(branches...).Tag("")
	}).Timeout(timeout, carapace.ActionValues()).Cache(5 * time.Second)
}

// listProjects returns the project names discovered under cfg.ProjectsDirectory.
func listProjects(_ context.Context, f *CommandConfig) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	gitClient, err := f.GitClient()
	if err != nil {
		return nil, fmt.Errorf("init git client: %w", err)
	}
	finder := git.NewRepoFinder(gitClient)
	gitDirs, err := finder.FindGitRepositories(cfg.ProjectsDirectory)
	if err != nil {
		return nil, fmt.Errorf("find git repositories: %w", err)
	}
	projects := make([]string, 0, len(gitDirs))
	for _, gd := range gitDirs {
		projects = append(projects, filepath.Base(gd.Path))
	}
	sort.Strings(projects)
	return projects, nil
}

// listBranchesFromContext returns the branches of the project at the
// current working directory, falling back to the first discovered
// project when no project context is available.
func listBranchesFromContext(ctx context.Context, f *CommandConfig) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	gitClient, err := f.GitClient()
	if err != nil {
		return nil, fmt.Errorf("init git client: %w", err)
	}
	projectPath := pickProjectPath(ctx, f, cfg)
	if projectPath == "" {
		return nil, nil
	}
	branches, err := gitClient.ListBranches(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}
	out := make([]string, 0, len(branches))
	for _, b := range branches {
		out = append(out, b.Name)
	}
	sort.Strings(out)
	return out, nil
}

// listBranchesForProject returns the branches of the named project.
func listBranchesForProject(ctx context.Context, f *CommandConfig, projectName string) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	gitClient, err := f.GitClient()
	if err != nil {
		return nil, fmt.Errorf("init git client: %w", err)
	}
	projectPath := filepath.Join(cfg.ProjectsDirectory, filepath.Base(projectName))
	branches, err := gitClient.ListBranches(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}
	out := make([]string, 0, len(branches))
	for _, b := range branches {
		out = append(out, b.Name)
	}
	sort.Strings(out)
	return out, nil
}

// pickProjectPath returns the project path best matching the current
// working directory. Mirrors the legacy resolver priority chain.
func pickProjectPath(_ context.Context, f *CommandConfig, cfg *core.Config) string {
	gitClient, err := f.GitClient()
	if err != nil {
		return ""
	}
	detector, err := git.NewContextDetector(cfg)
	if err != nil {
		return ""
	}
	wd, err := filepath.Abs(".")
	if err != nil {
		return ""
	}
	ctx, err := detector.DetectContext(wd)
	if err != nil || ctx == nil || ctx.ProjectName == "" {
		// Fallback to the first discovered project.
		finder := git.NewRepoFinder(gitClient)
		gitDirs, err := finder.FindGitRepositories(cfg.ProjectsDirectory)
		if err != nil || len(gitDirs) == 0 {
			return ""
		}
		return gitDirs[0].Path
	}
	return filepath.Join(cfg.ProjectsDirectory, ctx.ProjectName)
}
