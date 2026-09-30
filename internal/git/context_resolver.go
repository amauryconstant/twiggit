package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"twiggit/internal/core"
)

// Pure functions extracted from ContextResolver

// validatePathUnder validates that a target path is under a base directory.
// It returns a *core.OperationError with Op = "context.resolve" on failure
// so callers can dispatch uniformly through output.FormatError.
func validatePathUnder(base, target, targetType, baseDesc string) error {
	if under, err := core.IsPathUnder(base, target); err != nil {
		return &core.OperationError{
			Op:      "context.resolve",
			Entity:  target,
			Message: "path validation failed",
			Cause:   err,
		}
	} else if !under {
		return &core.OperationError{
			Op:      "context.resolve",
			Entity:  target,
			Message: fmt.Sprintf("%s path is outside configured %s directory", targetType, baseDesc),
		}
	}
	return nil
}

// parseCrossProjectReference parses a cross-project reference in the format "project/branch"
func parseCrossProjectReference(identifier string) (project, branch string, valid bool) {
	parts := strings.Split(identifier, "/")
	if len(parts) != 2 {
		return "", "", false
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// fuzzyMatch performs case-insensitive subsequence matching for fuzzy completion.
// pattern "f1" matches "feature-1", "feat-1", "F1", etc.
func fuzzyMatch(pattern, text string) bool {
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)
	pi := 0
	patternRunes := []rune(pattern)
	for _, c := range text {
		if pi < len(patternRunes) && c == patternRunes[pi] {
			pi++
		}
	}
	return pi == len(patternRunes)
}

// matchesExclusionPatterns checks if a name matches any of the given glob patterns
func matchesExclusionPatterns(name string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// containsPathTraversal checks if a string contains path traversal sequences.
// Handles literal "..", URL-encoded variants (all cases), and double-encoding.
func containsPathTraversal(s string) bool {
	if strings.Contains(s, "..") {
		return true
	}

	cleaned := filepath.Clean(s)
	if cleaned != s && strings.Contains(cleaned, "..") {
		return true
	}

	decoded, err := url.QueryUnescape(s)
	if err == nil && decoded != s {
		if strings.Contains(decoded, "..") {
			return true
		}
		doubleDecoded, err := url.QueryUnescape(decoded)
		if err == nil && doubleDecoded != decoded {
			if strings.Contains(doubleDecoded, "..") {
				return true
			}
		}
	}

	return false
}

// resolveMainIdentifier resolves "main" to the project root path
func (cr *contextResolver) resolveMainIdentifier(ctx *core.Context) (*core.ResolutionResult, error) {
	if containsPathTraversal(ctx.ProjectName) {
		return nil, &core.OperationError{
			Op:          "context.resolve",
			Entity:      "main",
			Field:       ctx.Path,
			Message:     "project name contains path traversal sequences",
			Suggestions: []string{"Use a valid project name without '..' or path separators"},
		}
	}

	projectPath := filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	if err := validatePathUnder(cr.config.ProjectsDirectory, projectPath, "project", "projects"); err != nil {
		return nil, err
	}

	return &core.ResolutionResult{
		ResolvedPath: projectPath,
		Type:         core.PathTypeProject,
		ProjectName:  ctx.ProjectName,
		Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
	}, nil
}

// resolveWorktreePath resolves a branch identifier to a worktree path
func (cr *contextResolver) resolveWorktreePath(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	if containsPathTraversal(ctx.ProjectName) || containsPathTraversal(identifier) {
		return nil, &core.OperationError{
			Op:          "context.resolve",
			Entity:      identifier,
			Field:       ctx.Path,
			Message:     "project or branch name contains path traversal sequences",
			Suggestions: []string{"Use a valid project or branch name without '..' or path separators"},
		}
	}

	worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)
	if err := validatePathUnder(cr.config.WorktreesDirectory, worktreePath, "worktree", "worktrees"); err != nil {
		return nil, err
	}

	return &core.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         core.PathTypeWorktree,
		ProjectName:  ctx.ProjectName,
		BranchName:   identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
	}, nil
}

type contextResolver struct {
	config     *core.Config
	goGit      *Client
	cli        *Client
	repoFinder *RepoFinder
}

// ContextResolver is the public type alias for git context resolution.
// cmd/run functions construct it via NewContextResolver.
type ContextResolver = contextResolver

// NewContextResolver creates a new context resolver.
func NewContextResolver(cfg *core.Config, goGit, cli *Client) *ContextResolver {
	return &contextResolver{
		config:     cfg,
		goGit:      goGit,
		cli:        cli,
		repoFinder: NewRepoFinder(goGit),
	}
}

func (cr *contextResolver) ResolveIdentifier(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	// Handle empty identifier
	if identifier == "" {
		return nil, &core.OperationError{
			Op:      "context.resolve",
			Entity:  "",
			Field:   "",
			Message: "empty identifier",
		}
	}

	switch ctx.Type {
	case core.ContextProject:
		return cr.resolveFromProjectContext(ctx, identifier)
	case core.ContextWorktree:
		return cr.resolveFromWorktreeContext(ctx, identifier)
	case core.ContextOutsideGit:
		return cr.resolveFromOutsideGitContext(ctx, identifier)
	default:
		return &core.ResolutionResult{
			Type:        core.PathTypeInvalid,
			Explanation: fmt.Sprintf("Cannot resolve identifier '%s' from unknown context", identifier),
		}, nil
	}
}

func (cr *contextResolver) GetResolutionSuggestions(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
	config := &suggestionConfig{}
	for _, opt := range opts {
		opt(config)
	}

	var suggestions []*core.ResolutionSuggestion

	switch ctx.Type {
	case core.ContextProject:
		suggestions = append(suggestions, cr.getProjectContextSuggestions(ctx, partial, config)...)
	case core.ContextWorktree:
		suggestions = append(suggestions, cr.getWorktreeContextSuggestions(ctx, partial, config)...)
	case core.ContextOutsideGit:
		suggestions = append(suggestions, cr.getOutsideGitContextSuggestions(partial)...)
	}

	return suggestions, nil
}

// suggestionConfig holds configuration for resolution suggestions
type suggestionConfig struct {
	existingOnly bool
}

// WithExistingOnly returns an option that filters suggestions to existing worktrees only.
func WithExistingOnly() core.SuggestionOption {
	return func(c any) {
		if cfg, ok := c.(*suggestionConfig); ok {
			cfg.existingOnly = true
		}
	}
}

// worktreeExists reports whether the given worktree path exists on disk.
// The path enters from git's `worktree list` output (user-influenced), so we
// bind the stat to the parent directory via os.Root to constrain traversal.
func worktreeExists(path string) bool {
	if path == "" {
		return false
	}
	parent := filepath.Dir(path)
	base := filepath.Base(path)
	root, err := os.OpenRoot(parent)
	if err != nil {
		return false
	}
	defer func() { _ = root.Close() }()
	_, err = root.Stat(base)
	return !errors.Is(err, os.ErrNotExist)
}

func (cr *contextResolver) resolveFromProjectContext(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	if identifier == "main" {
		return cr.resolveMainIdentifier(ctx)
	}

	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	return cr.resolveWorktreePath(ctx, identifier)
}

func (cr *contextResolver) getProjectContextSuggestions(ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	var suggestions []*core.ResolutionSuggestion

	// Add main suggestion
	suggestions = cr.addMainSuggestion(suggestions, ctx, partial, config)

	// Add worktree and branch suggestions if git service is available
	if cr.cli != nil && ctx.Path != "" {
		worktrees, err := cr.cli.ListWorktrees(context.Background(), ctx.Path)
		if err == nil {
			suggestions = cr.addWorktreeSuggestions(suggestions, ctx, partial, worktrees, config)
			suggestions = cr.addBranchSuggestions(suggestions, ctx, partial, worktrees, config)
		}
	}

	// Add project suggestions (exclude current project for cross-project navigation)
	suggestions = cr.addProjectSuggestions(suggestions, ctx, partial, true)

	return suggestions
}

// addMainSuggestion adds the "main" project root suggestion
func (cr *contextResolver) addMainSuggestion(suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	// Skip main suggestion when existingOnly is true (main is not a worktree)
	if config.existingOnly {
		return suggestions
	}

	//nolint:gocritic // argOrder: carapace completion wants "main" when the user's partial input is "", "m", "ma", "mai", or "main" — reverse-direction HasPrefix is intentional.
	if strings.HasPrefix("main", partial) {
		suggestions = append(suggestions, &core.ResolutionSuggestion{
			Text:        "main",
			Description: "Project root directory",
			Type:        core.PathTypeProject,
			ProjectName: ctx.ProjectName,
		})
	}
	return suggestions
}

// addWorktreeSuggestions adds suggestions for existing worktrees
func (cr *contextResolver) addWorktreeSuggestions(suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, worktrees []core.WorktreeInfo, config *suggestionConfig) []*core.ResolutionSuggestion {
	for _, worktree := range worktrees {
		// Apply fuzzy matching if enabled
		if cr.config.Navigation.FuzzyMatching {
			if !fuzzyMatch(partial, worktree.Branch) {
				continue
			}
		} else {
			if !strings.HasPrefix(worktree.Branch, partial) {
				continue
			}
		}

		// Apply exclusion patterns
		if matchesExclusionPatterns(worktree.Branch, cr.config.Completion.ExcludeBranches) {
			continue
		}

		if config.existingOnly && !worktreeExists(worktree.Path) {
			continue
		}

		// Check if this is the current worktree
		isCurrent := ctx.Type == core.ContextWorktree && ctx.BranchName == worktree.Branch

		// Check dirty status for current worktree only (performance optimization)
		var isDirty bool
		if isCurrent && cr.goGit != nil {
			if status, err := cr.goGit.GetRepositoryStatus(context.Background(), worktree.Path); err == nil {
				isDirty = !status.IsClean
			}
		}

		// Build enhanced description with remote tracking info
		description := "Worktree for branch " + worktree.Branch
		if isDirty {
			description = "⚠ " + description
		}

		suggestions = append(suggestions, &core.ResolutionSuggestion{
			Text:        worktree.Branch,
			Description: description,
			Type:        core.PathTypeWorktree,
			ProjectName: ctx.ProjectName,
			BranchName:  worktree.Branch,
			IsCurrent:   isCurrent,
			IsDirty:     isDirty,
		})
	}
	return suggestions
}

// addBranchSuggestions adds suggestions for branches without worktrees
func (cr *contextResolver) addBranchSuggestions(suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, existingWorktrees []core.WorktreeInfo, _ *suggestionConfig) []*core.ResolutionSuggestion {
	// When in worktree context, ListBranches should be called on project path, not worktree path
	var listPath string
	if ctx.Type == core.ContextWorktree {
		listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	} else {
		listPath = ctx.Path
	}

	branches, err := cr.goGit.ListBranches(context.Background(), listPath)
	if err != nil {
		// Silent degradation is acceptable for suggestions - errors shouldn't prevent
		// operation from proceeding, just reduce in helpfulness of completions
		return suggestions
	}

	// Build map of existing worktree branches from passed list
	worktreeBranches := make(map[string]bool)
	for _, worktree := range existingWorktrees {
		worktreeBranches[worktree.Branch] = true
	}

	for _, branch := range branches {
		// Skip if already has worktree
		if worktreeBranches[branch.Name] {
			continue
		}

		// Apply fuzzy matching if enabled
		if cr.config.Navigation.FuzzyMatching {
			if !fuzzyMatch(partial, branch.Name) {
				continue
			}
		} else {
			if !strings.HasPrefix(branch.Name, partial) {
				continue
			}
		}

		// Apply exclusion patterns
		if matchesExclusionPatterns(branch.Name, cr.config.Completion.ExcludeBranches) {
			continue
		}

		// Build enhanced description with remote info
		description := fmt.Sprintf("Branch %s (create worktree)", branch.Name)
		if branch.Remote != "" {
			description = fmt.Sprintf("Branch • %s (create worktree)", branch.Remote)
		}

		suggestions = append(suggestions, &core.ResolutionSuggestion{
			Text:        branch.Name,
			Description: description,
			Type:        core.PathTypeProject,
			ProjectName: ctx.ProjectName,
			BranchName:  branch.Name,
		})
	}
	return suggestions
}

// addProjectSuggestions adds suggestions for other projects (for cross-project navigation)
func (cr *contextResolver) addProjectSuggestions(suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, excludeCurrentProject bool) []*core.ResolutionSuggestion {
	projects, err := cr.discoverProjects()
	if err != nil {
		// Graceful degradation - return existing suggestions on error
		return suggestions
	}

	for _, project := range projects {
		// Exclude current project if requested
		if excludeCurrentProject && project.Name == ctx.ProjectName {
			continue
		}

		// Apply fuzzy matching if enabled
		if cr.config.Navigation.FuzzyMatching {
			if !fuzzyMatch(partial, project.Name) {
				continue
			}
		} else {
			if !strings.HasPrefix(project.Name, partial) {
				continue
			}
		}

		// Apply exclusion patterns
		if matchesExclusionPatterns(project.Name, cr.config.Completion.ExcludeProjects) {
			continue
		}

		suggestions = append(suggestions, &core.ResolutionSuggestion{
			Text:        project.Name,
			Description: "Project directory",
			Type:        core.PathTypeProject,
			ProjectName: project.Name,
		})
	}
	return suggestions
}

func (cr *contextResolver) resolveFromWorktreeContext(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	if identifier == "main" {
		return cr.resolveMainIdentifier(ctx)
	}

	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	return cr.resolveWorktreePath(ctx, identifier)
}

func (cr *contextResolver) getWorktreeContextSuggestions(ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	suggestions := cr.addMainSuggestion(nil, ctx, partial, config)

	if cr.cli != nil && ctx.Path != "" {
		// When in worktree context, ListWorktrees should be called on project path, not worktree path
		// Construct project path from project name and projects directory
		var listPath string
		if ctx.Type == core.ContextWorktree {
			listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
		} else {
			listPath = ctx.Path
		}

		if worktrees, err := cr.cli.ListWorktrees(context.Background(), listPath); err == nil {
			suggestions = cr.addWorktreeSuggestions(suggestions, ctx, partial, worktrees, config)
			suggestions = cr.addBranchSuggestions(suggestions, ctx, partial, worktrees, config)
		}
	}

	// Add project suggestions (exclude current project for cross-project navigation)
	suggestions = cr.addProjectSuggestions(suggestions, ctx, partial, true)

	return suggestions
}

func (cr *contextResolver) resolveFromOutsideGitContext(_ *core.Context, identifier string) (*core.ResolutionResult, error) {
	// Check if identifier contains "/" (project/branch format)
	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	// Validate project name doesn't contain path traversal sequences
	if containsPathTraversal(identifier) {
		return nil, &core.OperationError{
			Op:          "context.resolve",
			Entity:      identifier,
			Message:     "project name contains path traversal sequences",
			Suggestions: []string{"Use a valid project name without '..' or path separators"},
		}
	}

	// Resolve as project name
	projectPath := filepath.Join(cr.config.ProjectsDirectory, identifier)

	// Validate the project path is under the projects directory to prevent path traversal
	if err := validatePathUnder(cr.config.ProjectsDirectory, projectPath, "project", "projects"); err != nil {
		return nil, err
	}

	return &core.ResolutionResult{
		ResolvedPath: projectPath,
		Type:         core.PathTypeProject,
		ProjectName:  identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to project directory", identifier),
	}, nil
}

func (cr *contextResolver) getOutsideGitContextSuggestions(partial string) []*core.ResolutionSuggestion {
	// Check if projects directory is configured and accessible
	if cr.config.ProjectsDirectory == "" {
		return []*core.ResolutionSuggestion{}
	}

	// Discover projects in the configured directory
	projects, err := cr.discoverProjects()
	if err != nil {
		// Graceful degradation - return empty suggestions on error
		return []*core.ResolutionSuggestion{}
	}

	// Filter projects by partial match and create suggestions
	var suggestions []*core.ResolutionSuggestion
	for _, project := range projects {
		// Apply fuzzy matching if enabled
		if cr.config.Navigation.FuzzyMatching {
			if !fuzzyMatch(partial, project.Name) {
				continue
			}
		} else {
			if !strings.HasPrefix(project.Name, partial) {
				continue
			}
		}

		// Apply exclusion patterns
		if matchesExclusionPatterns(project.Name, cr.config.Completion.ExcludeProjects) {
			continue
		}

		suggestions = append(suggestions, &core.ResolutionSuggestion{
			Text:        project.Name,
			Description: "Project directory",
			Type:        core.PathTypeProject,
			ProjectName: project.Name,
		})
	}

	return suggestions
}

// discoverProjects scans the projects directory for git repositories.
// Returns lightweight project summaries for suggestion generation.
// Failures are wrapped as *core.OperationError with Op = "context.resolve".
func (cr *contextResolver) discoverProjects() ([]core.ProjectSummary, error) {
	projectsDir := cr.config.ProjectsDirectory

	gitDirs, err := cr.repoFinder.FindGitRepositories(projectsDir)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "context.resolve",
			Entity:  projectsDir,
			Message: "failed to scan for git repositories",
			Cause:   err,
		}
	}

	projects := make([]core.ProjectSummary, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		projects = append(projects, core.ProjectSummary{
			Name:        gitDir.Name,
			Path:        gitDir.Path,
			GitRepoPath: gitDir.Path,
		})
	}

	return projects, nil
}

func (cr *contextResolver) resolveCrossProjectReference(identifier string) (*core.ResolutionResult, error) {
	// Check for path traversal before parsing
	if containsPathTraversal(identifier) {
		return nil, &core.OperationError{
			Op:          "context.resolve",
			Entity:      identifier,
			Message:     "identifier contains path traversal sequences",
			Suggestions: []string{"Use format 'project/branch' with valid names"},
		}
	}

	projectName, branchName, valid := parseCrossProjectReference(identifier)
	if !valid {
		return &core.ResolutionResult{
			Type:        core.PathTypeInvalid,
			Explanation: fmt.Sprintf("Invalid cross-project reference format: '%s'. Expected: project/branch", identifier),
		}, nil
	}

	// Resolve to worktree of specified project
	worktreePath := filepath.Join(cr.config.WorktreesDirectory, projectName, branchName)

	// Validate the worktree path is under the worktrees directory to prevent path traversal
	if err := validatePathUnder(cr.config.WorktreesDirectory, worktreePath, "worktree", "worktrees"); err != nil {
		return nil, err
	}

	return &core.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         core.PathTypeWorktree,
		ProjectName:  projectName,
		BranchName:   branchName,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, projectName),
	}, nil
}
