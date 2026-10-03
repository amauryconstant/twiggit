package git

import (
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

// ContextResolver is the public type alias for git context resolution.
// cmd/run functions construct it via NewContextResolver.
// Declared BEFORE the private struct so `go doc -all` can surface the
// role-method signatures that *git.Client callers see through the
// alias (the unexported contextResolver struct hides them otherwise).
type ContextResolver = contextResolver

type contextResolver struct {
	config     *core.Config
	goGit      *Client
	cli        *Client
	repoFinder *RepoFinder
}

// NewContextResolver creates a new context resolver.
func NewContextResolver(cfg *core.Config, goGit, cli *Client) *ContextResolver {
	return &contextResolver{
		config:     cfg,
		goGit:      goGit,
		cli:        cli,
		repoFinder: NewRepoFinder(goGit),
	}
}

// ResolveIdentifier routes the identifier through the resolver chain
// matching the supplied Context type: project, worktree, or outside-git.
// An empty identifier returns a "context.resolve" OperationError; an
// unrecognised Context type returns a PathTypeInvalid result. The
// returned ResolutionResult is non-nil on success.
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
		// resolveFromProjectContext handles both Project and Worktree
		// contexts identically (the previous identical-twin
		// resolveFromWorktreeContext was deleted in §10 of
		// naming-refactor-modernize).
		return cr.resolveFromProjectContext(ctx, identifier)
	case core.ContextOutsideGit:
		return cr.resolveFromOutsideGitContext(ctx, identifier)
	default:
		return &core.ResolutionResult{
			Type:        core.PathTypeInvalid,
			Explanation: fmt.Sprintf("Cannot resolve identifier '%s' from unknown context", identifier),
		}, nil
	}
}

// ResolutionSuggestions returns completion suggestions matching the
// partial identifier for the supplied Context type. Options
// (e.g. WithExistingOnly) filter the result set; callers can pass
// multiple options. The returned slice is non-nil on success and may
// be empty.
func (cr *contextResolver) ResolutionSuggestions(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
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
