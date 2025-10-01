package infrastructure

import (
	"fmt"
	"path/filepath"
	"strings"

	"twiggit/internal/domain"
)

type contextResolver struct {
	config *domain.Config
}

// NewContextResolver creates a new context resolver
func NewContextResolver(cfg *domain.Config) domain.ContextResolver {
	return &contextResolver{config: cfg}
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
		suggestions = append(suggestions, cr.getOutsideGitContextSuggestions(ctx, partial)...)
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

	// Always suggest "main" for project context
	if strings.HasPrefix("main", partial) {
		suggestions = append(suggestions, &domain.ResolutionSuggestion{
			Text:        "main",
			Description: "Project root directory",
			Type:        domain.PathTypeProject,
			ProjectName: ctx.ProjectName,
		})
	}

	// TODO: Add actual worktree discovery when git operations are available
	// For now, provide basic branch name suggestions

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

	// TODO: Add actual worktree discovery when git operations are available

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

func (cr *contextResolver) getOutsideGitContextSuggestions(_ *domain.Context, _ string) []*domain.ResolutionSuggestion {
	var suggestions []*domain.ResolutionSuggestion

	// TODO: Add actual project discovery when git operations are available
	// For now, provide basic suggestions

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
