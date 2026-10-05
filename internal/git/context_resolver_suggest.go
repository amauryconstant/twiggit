package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"twiggit/internal/core"
)

// Suggestion builders for ContextResolver. Split from
// context_resolver.go so the resolver chain (ResolveIdentifier and
// the four core.Result helpers) reads as one continuous concern and
// the completion path reads as another. Both files share the
// contextResolver struct; only the file boundary changes.

// suggestionConfig holds configuration for resolution suggestions
type suggestionConfig struct {
	isExistingOnly bool
}

// WithExistingOnly returns an option that filters suggestions to existing worktrees only.
func WithExistingOnly() core.SuggestionOption {
	return func(c any) {
		if cfg, ok := c.(*suggestionConfig); ok {
			cfg.isExistingOnly = true
		}
	}
}

func (cr *contextResolver) getProjectContextSuggestions(ctxStd context.Context, ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	return cr.collectContextSuggestions(ctxStd, ctx, partial, config, ctx.Path)
}

// addMainSuggestion adds the "main" project root suggestion
func (cr *contextResolver) addMainSuggestion(ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	// Skip main suggestion when isExistingOnly is true (main is not a worktree)
	if config.isExistingOnly {
		return nil
	}

	//nolint:gocritic // argOrder: carapace completion wants "main" when the user's partial input is "", "m", "ma", "mai", or "main" — reverse-direction HasPrefix is intentional.
	if strings.HasPrefix("main", partial) {
		return []*core.ResolutionSuggestion{{
			Text:        "main",
			Description: "Project root directory",
			Type:        core.PathTypeProject,
			ProjectName: ctx.ProjectName,
		}}
	}
	return nil
}

// addWorktreeSuggestions adds suggestions for existing worktrees
func (cr *contextResolver) addWorktreeSuggestions(ctxStd context.Context, suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, worktrees []core.Worktree, config *suggestionConfig) []*core.ResolutionSuggestion {
	for _, worktree := range worktrees {
		// Apply fuzzy matching if enabled
		if cr.config.Navigation.IsFuzzyMatching {
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

		if config.isExistingOnly && !worktreeExists(worktree.Path) {
			continue
		}

		// Check if this is the current worktree
		isCurrent := ctx.Type == core.ContextWorktree && ctx.BranchName == worktree.Branch

		// Check dirty status for current worktree only (performance optimization)
		var isDirty bool
		if isCurrent && cr.goGit != nil {
			if status, err := cr.goGit.RepositoryStatus(ctxStd, worktree.Path); err == nil {
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

// addBranchSuggestions adds suggestions for branches without worktrees.
// The suggestionConfig parameter was dropped; existingOnly is the only
// filter and is applied in addWorktreeSuggestions against the same
// worktree list.
func (cr *contextResolver) addBranchSuggestions(ctxStd context.Context, suggestions []*core.ResolutionSuggestion, ctx *core.Context, partial string, existingWorktrees []core.Worktree) []*core.ResolutionSuggestion {
	// When in worktree context, ListBranches should be called on project path, not worktree path
	var listPath string
	if ctx.Type == core.ContextWorktree {
		listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	} else {
		listPath = ctx.Path
	}

	branches, err := cr.goGit.ListBranches(ctxStd, listPath)
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
		if cr.config.Navigation.IsFuzzyMatching {
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
		if cr.config.Navigation.IsFuzzyMatching {
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

func (cr *contextResolver) getWorktreeContextSuggestions(ctxStd context.Context, ctx *core.Context, partial string, config *suggestionConfig) []*core.ResolutionSuggestion {
	listPath := ctx.Path
	if ctx.Type == core.ContextWorktree {
		listPath = filepath.Join(cr.config.ProjectsDirectory, ctx.ProjectName)
	}
	return cr.collectContextSuggestions(ctxStd, ctx, partial, config, listPath)
}

// collectContextSuggestions is the merged helper that backs both
// getProjectContextSuggestions and getWorktreeContextSuggestions.
// The only per-context difference was the path passed to
// ListWorktrees (the project context uses ctx.Path; the worktree
// context re-anchors to the project root via cr.config.ProjectsDirectory +
// ctx.ProjectName). Centralising the path choice eliminates the
// duplicated walk over worktrees / branches / projects and keeps the
// two public entry points as thin dispatchers.
func (cr *contextResolver) collectContextSuggestions(ctxStd context.Context, ctx *core.Context, partial string, config *suggestionConfig, listPath string) []*core.ResolutionSuggestion {
	suggestions := cr.addMainSuggestion(ctx, partial, config)

	if cr.cli != nil && listPath != "" {
		if worktrees, err := cr.cli.ListWorktrees(ctxStd, listPath); err == nil {
			suggestions = cr.addWorktreeSuggestions(ctxStd, suggestions, ctx, partial, worktrees, config)
			suggestions = cr.addBranchSuggestions(ctxStd, suggestions, ctx, partial, worktrees)
		}
	}

	suggestions = cr.addProjectSuggestions(suggestions, ctx, partial, true)

	return suggestions
}

func (cr *contextResolver) getOutsideGitContextSuggestions(_ context.Context, partial string) []*core.ResolutionSuggestion {
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
		if cr.config.Navigation.IsFuzzyMatching {
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
