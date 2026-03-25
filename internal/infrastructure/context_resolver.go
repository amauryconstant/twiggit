package infrastructure

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

// Pure functions extracted from ContextResolver

// validatePathUnder validates that a target path is under a base directory
// Returns an error if validation fails or if path is outside base
func validatePathUnder(base, target, targetType, baseDesc string) error {
	if under, err := IsPathUnder(base, target); err != nil {
		return domain.NewContextDetectionError(target, "path validation failed", err)
	} else if !under {
		return domain.NewContextDetectionError(target,
			fmt.Sprintf("%s path is outside configured %s directory", targetType, baseDesc), nil)
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

// fuzzyMatch performs case-insensitive subsequence matching for fuzzy completion
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

// containsPathTraversal checks if a string contains path traversal sequences
// Handles literal "..", URL-encoded variants (all cases), and double-encoding
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

// buildWorktreePath builds the path to a worktree for a given project and branch
func buildWorktreePath(worktreesDir, project, branch string) string {
	return filepath.Join(worktreesDir, project, branch)
}

// resolveMainIdentifier resolves "main" to the project root path
func (cr *contextResolver) resolveMainIdentifier(ctx *domain.Context) (*domain.ResolutionResult, error) {
	if containsPathTraversal(ctx.ProjectName) {
		return nil, domain.NewResolutionError(
			"main",
			ctx.Path,
			"project name contains path traversal sequences",
			[]string{"Use a valid project name without '..' or path separators"},
			nil,
		)
	}

	projectPath := filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	if err := validatePathUnder(cr.config.ProjectsDirectory, projectPath, "project", "projects"); err != nil {
		return nil, err
	}

	return &domain.ResolutionResult{
		ResolvedPath: projectPath,
		Type:         domain.PathTypeProject,
		ProjectName:  ctx.ProjectName,
		Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
	}, nil
}

// resolveWorktreePath resolves a branch identifier to a worktree path
func (cr *contextResolver) resolveWorktreePath(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	if containsPathTraversal(ctx.ProjectName) || containsPathTraversal(identifier) {
		return nil, domain.NewResolutionError(
			identifier,
			ctx.Path,
			"project or branch name contains path traversal sequences",
			[]string{"Use a valid project or branch name without '..' or path separators"},
			nil,
		)
	}

	worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)
	if err := validatePathUnder(cr.config.WorktreesDirectory, worktreePath, "worktree", "worktrees"); err != nil {
		return nil, err
	}

	return &domain.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         domain.PathTypeWorktree,
		ProjectName:  ctx.ProjectName,
		BranchName:   identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
	}, nil
}

// buildProjectPath builds the path to a project directory
func buildProjectPath(projectsDir, project string) string {
	return filepath.Join(projectsDir, project)
}

// filterSuggestions filters suggestions based on a partial string match
func filterSuggestions(suggestions []string, partial string) []string {
	result := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, partial) {
			result = append(result, suggestion)
		}
	}
	return result
}

type contextResolver struct {
	config     *domain.Config
	gitService application.GitClient
}

// NewContextResolver creates a new context resolver
func NewContextResolver(cfg *domain.Config, gitService application.GitClient) application.ContextResolver {
	return &contextResolver{
		config:     cfg,
		gitService: gitService,
	}
}

func (cr *contextResolver) ResolveIdentifier(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Handle empty identifier
	if identifier == "" {
		return nil, domain.NewResolutionError("", "", "empty identifier", nil, nil)
	}

	switch ctx.Type {
	case domain.ContextProject:
		return cr.resolveFromProjectContext(ctx, identifier)
	case domain.ContextWorktree:
		return cr.resolveFromWorktreeContext(ctx, identifier)
	case domain.ContextOutsideGit:
		return cr.resolveFromOutsideGitContext(ctx, identifier)
	default:
		return &domain.ResolutionResult{
			Type:        domain.PathTypeInvalid,
			Explanation: fmt.Sprintf("Cannot resolve identifier '%s' from unknown context", identifier),
		}, nil
	}
}

func (cr *contextResolver) GetResolutionSuggestions(ctx *domain.Context, partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error) {
	config := &suggestionConfig{}
	for _, opt := range opts {
		opt(config)
	}

	var suggestions []*domain.ResolutionSuggestion

	switch ctx.Type {
	case domain.ContextProject:
		suggestions = append(suggestions, cr.getProjectContextSuggestions(ctx, partial, config)...)
	case domain.ContextWorktree:
		suggestions = append(suggestions, cr.getWorktreeContextSuggestions(ctx, partial, config)...)
	case domain.ContextOutsideGit:
		suggestions = append(suggestions, cr.getOutsideGitContextSuggestions(partial)...)
	}

	return suggestions, nil
}

// suggestionConfig holds configuration for resolution suggestions
type suggestionConfig struct {
	existingOnly bool
}

// WithExistingOnly returns an option that filters suggestions to existing worktrees only
func WithExistingOnly() domain.SuggestionOption {
	return func(c interface{}) {
		if cfg, ok := c.(*suggestionConfig); ok {
			cfg.existingOnly = true
		}
	}
}

func (cr *contextResolver) resolveFromProjectContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	if identifier == "main" {
		return cr.resolveMainIdentifier(ctx)
	}

	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	return cr.resolveWorktreePath(ctx, identifier)
}

func (cr *contextResolver) getProjectContextSuggestions(ctx *domain.Context, partial string, config *suggestionConfig) []*domain.ResolutionSuggestion {
	var suggestions []*domain.ResolutionSuggestion

	// Add main suggestion
	suggestions = cr.addMainSuggestion(suggestions, ctx, partial, config)

	// Add worktree and branch suggestions if git service is available
	if cr.gitService != nil && ctx.Path != "" {
		worktrees, err := cr.gitService.ListWorktrees(context.Background(), ctx.Path)
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
func (cr *contextResolver) addMainSuggestion(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string, config *suggestionConfig) []*domain.ResolutionSuggestion {
	// Skip main suggestion when existingOnly is true (main is not a worktree)
	if config.existingOnly {
		return suggestions
	}

	if strings.HasPrefix("main", partial) {
		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        "main",
			Description: "Project root directory",
			Type:        domain.PathTypeProject,
			ProjectName: ctx.ProjectName,
		})
	}
	return suggestions
}

// addWorktreeSuggestions adds suggestions for existing worktrees
func (cr *contextResolver) addWorktreeSuggestions(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string, worktrees []domain.WorktreeInfo, config *suggestionConfig) []*domain.ResolutionSuggestion {
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

		if config.existingOnly {
			if _, err := os.Stat(worktree.Path); os.IsNotExist(err) {
				continue
			}
		}

		// Check if this is the current worktree
		isCurrent := ctx.Type == domain.ContextWorktree && ctx.BranchName == worktree.Branch

		// Check dirty status for current worktree only (performance optimization)
		var isDirty bool
		if isCurrent && cr.gitService != nil {
			if status, err := cr.gitService.GetRepositoryStatus(context.Background(), worktree.Path); err == nil {
				isDirty = !status.IsClean
			}
		}

		// Build enhanced description with remote tracking info
		description := "Worktree for branch " + worktree.Branch
		if isDirty {
			description = "⚠ " + description
		}

		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        worktree.Branch,
			Description: description,
			Type:        domain.PathTypeWorktree,
			ProjectName: ctx.ProjectName,
			BranchName:  worktree.Branch,
			IsCurrent:   isCurrent,
			IsDirty:     isDirty,
		})
	}
	return suggestions
}

// addBranchSuggestions adds suggestions for branches without worktrees
func (cr *contextResolver) addBranchSuggestions(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string, existingWorktrees []domain.WorktreeInfo, _ *suggestionConfig) []*domain.ResolutionSuggestion {
	// When in worktree context, ListBranches should be called on project path, not worktree path
	var listPath string
	if ctx.Type == domain.ContextWorktree {
		listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	} else {
		listPath = ctx.Path
	}

	branches, err := cr.gitService.ListBranches(context.Background(), listPath)
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

		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        branch.Name,
			Description: description,
			Type:        domain.PathTypeProject,
			ProjectName: ctx.ProjectName,
			BranchName:  branch.Name,
			Remote:      branch.Remote,
		})
	}
	return suggestions
}

// addProjectSuggestions adds suggestions for other projects (for cross-project navigation)
func (cr *contextResolver) addProjectSuggestions(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string, excludeCurrentProject bool) []*domain.ResolutionSuggestion {
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

		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        project.Name,
			Description: "Project directory",
			Type:        domain.PathTypeProject,
			ProjectName: project.Name,
		})
	}
	return suggestions
}

func (cr *contextResolver) resolveFromWorktreeContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	if identifier == "main" {
		return cr.resolveMainIdentifier(ctx)
	}

	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	return cr.resolveWorktreePath(ctx, identifier)
}

func (cr *contextResolver) getWorktreeContextSuggestions(ctx *domain.Context, partial string, config *suggestionConfig) []*domain.ResolutionSuggestion {
	suggestions := cr.addMainSuggestion(nil, ctx, partial, config)

	if cr.gitService != nil && ctx.Path != "" {
		// When in worktree context, ListWorktrees should be called on project path, not worktree path
		// Construct project path from project name and projects directory
		var listPath string
		if ctx.Type == domain.ContextWorktree {
			listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
		} else {
			listPath = ctx.Path
		}

		if worktrees, err := cr.gitService.ListWorktrees(context.Background(), listPath); err == nil {
			suggestions = cr.addWorktreeSuggestions(suggestions, ctx, partial, worktrees, config)
			suggestions = cr.addBranchSuggestions(suggestions, ctx, partial, worktrees, config)
		}
	}

	// Add project suggestions (exclude current project for cross-project navigation)
	suggestions = cr.addProjectSuggestions(suggestions, ctx, partial, true)

	return suggestions
}

func (cr *contextResolver) resolveFromOutsideGitContext(_ *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Check if identifier contains "/" (project/branch format)
	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	// Validate project name doesn't contain path traversal sequences
	if containsPathTraversal(identifier) {
		return nil, domain.NewResolutionError(
			identifier,
			"",
			"project name contains path traversal sequences",
			[]string{"Use a valid project name without '..' or path separators"},
			nil,
		)
	}

	// Resolve as project name
	projectPath := filepath.Join(cr.config.ProjectsDirectory, identifier)

	// Validate the project path is under the projects directory to prevent path traversal
	if err := validatePathUnder(cr.config.ProjectsDirectory, projectPath, "project", "projects"); err != nil {
		return nil, err
	}

	return &domain.ResolutionResult{
		ResolvedPath: projectPath,
		Type:         domain.PathTypeProject,
		ProjectName:  identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to project directory", identifier),
	}, nil
}

func (cr *contextResolver) getOutsideGitContextSuggestions(partial string) []*domain.ResolutionSuggestion {
	// Check if projects directory is configured and accessible
	if cr.config.ProjectsDirectory == "" {
		return []*domain.ResolutionSuggestion{}
	}

	// Discover projects in the configured directory
	projects, err := cr.discoverProjects()
	if err != nil {
		// Graceful degradation - return empty suggestions on error
		return []*domain.ResolutionSuggestion{}
	}

	// Filter projects by partial match and create suggestions
	var suggestions []*domain.ResolutionSuggestion
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

		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        project.Name,
			Description: "Project directory",
			Type:        domain.PathTypeProject,
			ProjectName: project.Name,
		})
	}

	return suggestions
}

// ProjectRef represents a lightweight project reference for internal use
// This is distinct from domain.ProjectInfo which contains comprehensive project details
type ProjectRef struct {
	Name string
	Path string
}

// discoverProjects scans the projects directory for git repositories
// Returns lightweight project references for suggestion generation
func (cr *contextResolver) discoverProjects() ([]ProjectRef, error) {
	projectsDir := cr.config.ProjectsDirectory

	gitDirs, err := FindGitRepositories(projectsDir, cr.gitService)
	if err != nil {
		return nil, domain.NewContextDetectionError(projectsDir, "failed to scan for git repositories", err)
	}

	projects := make([]ProjectRef, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		projects = append(projects, ProjectRef{
			Name: gitDir.Name,
			Path: gitDir.Path,
		})
	}

	return projects, nil
}

func (cr *contextResolver) resolveCrossProjectReference(identifier string) (*domain.ResolutionResult, error) {
	// Check for path traversal before parsing
	if containsPathTraversal(identifier) {
		return nil, domain.NewResolutionError(
			identifier,
			"",
			"identifier contains path traversal sequences",
			[]string{"Use format 'project/branch' with valid names"},
			nil,
		)
	}

	projectName, branchName, valid := parseCrossProjectReference(identifier)
	if !valid {
		return &domain.ResolutionResult{
			Type:        domain.PathTypeInvalid,
			Explanation: fmt.Sprintf("Invalid cross-project reference format: '%s'. Expected: project/branch", identifier),
		}, nil
	}

	// Resolve to worktree of specified project
	worktreePath := filepath.Join(cr.config.WorktreesDirectory, projectName, branchName)

	// Validate the worktree path is under the worktrees directory to prevent path traversal
	if err := validatePathUnder(cr.config.WorktreesDirectory, worktreePath, "worktree", "worktrees"); err != nil {
		return nil, err
	}

	return &domain.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         domain.PathTypeWorktree,
		ProjectName:  projectName,
		BranchName:   branchName,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, projectName),
	}, nil
}
