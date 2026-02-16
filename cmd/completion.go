package cmd

import (
	"time"

	"github.com/carapace-sh/carapace"

	"twiggit/internal/domain"
)

// actionWorktreeTarget provides completion for worktree targets (project/branch)
// Supports progressive completion via ActionMultiParts("/")
func actionWorktreeTarget(config *CommandConfig, opts ...domain.SuggestionOption) carapace.Action {
	return carapace.ActionMultiParts("/", func(c carapace.Context) carapace.Action {
		switch len(c.Parts) {
		case 0:
			return actionProjectsOrBranches(c, config, opts)
		case 1:
			return actionBranchesForProject(c.Parts[0], config)
		default:
			return carapace.ActionValues()
		}
	}).Cache(5 * time.Second)
}

// actionBranches provides completion for branch names (--source flag)
func actionBranches(config *CommandConfig) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ctx, err := config.Services.ContextService.GetCurrentContext()
		if err != nil {
			return carapace.ActionValues()
		}

		suggestions, err := config.Services.ContextService.GetCompletionSuggestionsFromContext(ctx, c.Value)
		if err != nil {
			return carapace.ActionValues()
		}

		return suggestionsToCarapaceAction(suggestions)
	}).Cache(5 * time.Second)
}

// actionProjectsOrBranches suggests projects or branches based on current context
func actionProjectsOrBranches(c carapace.Context, config *CommandConfig, opts []domain.SuggestionOption) carapace.Action {
	ctx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return carapace.ActionValues()
	}

	suggestions, err := config.Services.ContextService.GetCompletionSuggestionsFromContext(ctx, c.Value, opts...)
	if err != nil {
		return carapace.ActionValues()
	}

	return suggestionsToCarapaceAction(suggestions)
}

// actionBranchesForProject suggests branches for a specific project
func actionBranchesForProject(projectName string, config *CommandConfig) carapace.Action {
	ctx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return carapace.ActionValues()
	}

	suggestions, err := config.Services.ContextService.GetCompletionSuggestionsFromContext(ctx, projectName)
	if err != nil {
		return carapace.ActionValues()
	}

	filtered := make([]*domain.ResolutionSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		if s.ProjectName == projectName {
			filtered = append(filtered, s)
		}
	}

	result := make([]string, 0, len(filtered))
	descriptions := make([]string, 0, len(filtered))
	for _, s := range filtered {
		result = append(result, s.Text)
		descriptions = append(descriptions, s.Description)
	}

	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		action := carapace.ActionValues(result...)
		for range result {
			action = action.Tag(descriptions[0])
		}
		return action
	}).Cache(5 * time.Second)
}

// suggestionsToCarapaceAction converts domain.ResolutionSuggestion to carapace.Action
func suggestionsToCarapaceAction(suggestions []*domain.ResolutionSuggestion) carapace.Action {
	if len(suggestions) == 0 {
		return carapace.ActionValues()
	}

	values := make([]string, len(suggestions))
	for _, s := range suggestions {
		values = append(values, s.Text)
	}

	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		action := carapace.ActionValues(values...)
		for _, s := range suggestions {
			action = action.Tag(s.Description)
		}
		return action
	})
}
