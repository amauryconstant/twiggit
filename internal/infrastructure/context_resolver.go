package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/domain"
)

// Pure functions extracted from ContextResolver

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

// buildWorktreePath builds the path to a worktree for a given project and branch
func buildWorktreePath(worktreesDir, project, branch string) string {
	return filepath.Join(worktreesDir, project, branch)
}

// buildProjectPath builds the path to a project directory
func buildProjectPath(projectsDir, project string) string {
	return filepath.Join(projectsDir, project)
}

// filterSuggestions filters suggestions based on a partial string match
func filterSuggestions(suggestions []string, partial string) []string {
	result := make([]string, 0)
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, partial) {
			result = append(result, suggestion)
		}
	}
	return result
}

type contextResolver struct {
	config     *domain.Config
	gitService GitClient
}

// NewContextResolver creates a new context resolver
func NewContextResolver(cfg *domain.Config, gitService GitClient) domain.ContextResolver {
	return &contextResolver{
		config:     cfg,
		gitService: gitService,
	}
}

func (cr *contextResolver) ResolveIdentifier(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Handle empty identifier
	if identifier == "" {
		return nil, domain.NewContextDetectionError("", "empty identifier", nil)
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

func (cr *contextResolver) GetResolutionSuggestions(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	var suggestions []*domain.ResolutionSuggestion

	switch ctx.Type {
	case domain.ContextProject:
		suggestions = append(suggestions, cr.getProjectContextSuggestions(ctx, partial)...)
	case domain.ContextWorktree:
		suggestions = append(suggestions, cr.getWorktreeContextSuggestions(ctx, partial)...)
	case domain.ContextOutsideGit:
		suggestions = append(suggestions, cr.getOutsideGitContextSuggestions(partial)...)
	}

	return suggestions, nil
}

func (cr *contextResolver) resolveFromProjectContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Handle special case: "main" resolves to project root
	if identifier == "main" {
		return &domain.ResolutionResult{
			ResolvedPath: ctx.Path,
			Type:         domain.PathTypeProject,
			ProjectName:  ctx.ProjectName,
			Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
		}, nil
	}

	// Check if identifier contains "/" (cross-project reference)
	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	// Resolve as branch name (worktree of current project)
	worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)

	return &domain.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         domain.PathTypeWorktree,
		ProjectName:  ctx.ProjectName,
		BranchName:   identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
	}, nil
}

func (cr *contextResolver) getProjectContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
	var suggestions []*domain.ResolutionSuggestion

	// Add main suggestion
	suggestions = cr.addMainSuggestion(suggestions, ctx, partial)

	// Add worktree and branch suggestions if git service is available
	if cr.gitService != nil && ctx.Path != "" {
		suggestions = cr.addWorktreeSuggestions(suggestions, ctx, partial)
		suggestions = cr.addBranchSuggestions(suggestions, ctx, partial)
	}

	return suggestions
}

// addMainSuggestion adds the "main" project root suggestion
func (cr *contextResolver) addMainSuggestion(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
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
func (cr *contextResolver) addWorktreeSuggestions(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
	worktrees, err := cr.gitService.ListWorktrees(context.Background(), ctx.Path)
	if err != nil {
		return suggestions
	}

	for _, worktree := range worktrees {
		if strings.HasPrefix(worktree.Branch, partial) {
			suggestions = append(suggestions, &domain.ResolutionSuggestion{
				Text:        worktree.Branch,
				Description: "Worktree for branch " + worktree.Branch,
				Type:        domain.PathTypeWorktree,
				ProjectName: ctx.ProjectName,
				BranchName:  worktree.Branch,
			})
		}
	}
	return suggestions
}

// addBranchSuggestions adds suggestions for branches without worktrees
func (cr *contextResolver) addBranchSuggestions(suggestions []*domain.ResolutionSuggestion, ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
	branches, err := cr.gitService.ListBranches(context.Background(), ctx.Path)
	if err != nil {
		return suggestions
	}

	// Get existing worktrees to avoid duplicates
	worktrees, _ := cr.gitService.ListWorktrees(context.Background(), ctx.Path)
	worktreeBranches := make(map[string]bool)
	for _, worktree := range worktrees {
		worktreeBranches[worktree.Branch] = true
	}

	for _, branch := range branches {
		if strings.HasPrefix(branch.Name, partial) && !worktreeBranches[branch.Name] {
			suggestions = append(suggestions, &domain.ResolutionSuggestion{
				Text:        branch.Name,
				Description: fmt.Sprintf("Branch %s (create worktree)", branch.Name),
				Type:        domain.PathTypeProject,
				ProjectName: ctx.ProjectName,
				BranchName:  branch.Name,
			})
		}
	}
	return suggestions
}

func (cr *contextResolver) resolveFromWorktreeContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Handle special case: "main" resolves to project root
	if identifier == "main" {
		projectPath := filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
		return &domain.ResolutionResult{
			ResolvedPath: projectPath,
			Type:         domain.PathTypeProject,
			ProjectName:  ctx.ProjectName,
			Explanation:  fmt.Sprintf("Resolved 'main' to project root '%s'", ctx.ProjectName),
		}, nil
	}

	// Check if identifier contains "/" (cross-project reference)
	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	// Resolve as different worktree of same project
	worktreePath := filepath.Join(cr.config.WorktreesDirectory, ctx.ProjectName, identifier)

	return &domain.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         domain.PathTypeWorktree,
		ProjectName:  ctx.ProjectName,
		BranchName:   identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, ctx.ProjectName),
	}, nil
}

func (cr *contextResolver) getWorktreeContextSuggestions(ctx *domain.Context, partial string) []*domain.ResolutionSuggestion {
	var suggestions []*domain.ResolutionSuggestion

	// Always suggest "main" for worktree context
	if strings.HasPrefix("main", partial) {
		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        "main",
			Description: "Project root directory",
			Type:        domain.PathTypeProject,
			ProjectName: ctx.ProjectName,
		})
	}

	// Add actual worktree discovery using git operations
	if cr.gitService != nil && ctx.Path != "" {
		if worktrees, err := cr.gitService.ListWorktrees(context.Background(), ctx.Path); err == nil {
			for _, worktree := range worktrees {
				if strings.HasPrefix(worktree.Branch, partial) {
					suggestions = append(suggestions, &domain.ResolutionSuggestion{
						Text:        worktree.Branch,
						Description: "Worktree for branch " + worktree.Branch,
						Type:        domain.PathTypeWorktree,
						ProjectName: ctx.ProjectName,
						BranchName:  worktree.Branch,
					})
				}
			}
		}
	}

	return suggestions
}

func (cr *contextResolver) resolveFromOutsideGitContext(_ *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	// Check if identifier contains "/" (project/branch format)
	if strings.Contains(identifier, "/") {
		return cr.resolveCrossProjectReference(identifier)
	}

	// Resolve as project name
	projectPath := filepath.Join(cr.config.ProjectsDirectory, identifier)

	return &domain.ResolutionResult{
		ResolvedPath: projectPath,
		Type:         domain.PathTypeProject,
		ProjectName:  identifier,
		Explanation:  fmt.Sprintf("Resolved '%s' to project directory", identifier),
	}, nil
}

// ProjectInfo represents information about a discovered project
type ProjectInfo struct {
	Name string
	Path string
}

// discoverProjects scans the projects directory for git repositories
func (cr *contextResolver) discoverProjects() ([]ProjectInfo, error) {
	projectsDir := cr.config.ProjectsDirectory

	// Check if directory exists
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("projects directory does not exist: %s", projectsDir)
	}

	// Read directory contents
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	projects := make([]ProjectInfo, 0, 10) // Pre-allocate with reasonable capacity
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())

		// Validate it's a git repository if git service is available
		if cr.gitService != nil {
			if err := cr.gitService.ValidateRepository(projectPath); err != nil {
				continue // Skip non-git directories
			}
		}

		projects = append(projects, ProjectInfo{
			Name: entry.Name(),
			Path: projectPath,
		})
	}

	return projects, nil
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
		if strings.HasPrefix(project.Name, partial) {
			suggestions = append(suggestions, &domain.ResolutionSuggestion{
				Text:        project.Name,
				Description: "Project directory",
				Type:        domain.PathTypeProject,
				ProjectName: project.Name,
			})
		}
	}

	return suggestions
}

func (cr *contextResolver) resolveCrossProjectReference(identifier string) (*domain.ResolutionResult, error) {
	parts := strings.Split(identifier, "/")
	if len(parts) != 2 {
		return &domain.ResolutionResult{
			Type:        domain.PathTypeInvalid,
			Explanation: fmt.Sprintf("Invalid cross-project reference format: '%s'. Expected: project/branch", identifier),
		}, nil
	}

	projectName := parts[0]
	branchName := parts[1]

	// Resolve to worktree of specified project
	worktreePath := filepath.Join(cr.config.WorktreesDirectory, projectName, branchName)

	return &domain.ResolutionResult{
		ResolvedPath: worktreePath,
		Type:         domain.PathTypeWorktree,
		ProjectName:  projectName,
		BranchName:   branchName,
		Explanation:  fmt.Sprintf("Resolved '%s' to worktree of project '%s'", identifier, projectName),
	}, nil
}
