package cmd

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/carapace-sh/carapace"

	"twiggit/internal/domain"
)

// getCompletionTimeout returns the completion timeout duration from config, defaulting to 500ms
func getCompletionTimeout(config *domain.Config) time.Duration {
	if config == nil {
		return 500 * time.Millisecond
	}
	if config.Completion.Timeout != "" {
		if duration, err := time.ParseDuration(config.Completion.Timeout); err == nil {
			return duration
		}
	}
	return 500 * time.Millisecond
}

// actionWorktreeTarget provides completion for worktree targets (project/branch)
// Supports progressive completion via ActionMultiParts("/")
func actionWorktreeTarget(config *CommandConfig, opts ...domain.SuggestionOption) carapace.Action {
	return carapace.ActionMultiParts("/", func(c carapace.Context) carapace.Action {
		timeout := getCompletionTimeout(config.Config)

		switch len(c.Parts) {
		case 0:
			return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
				return actionProjectsOrBranches(c, config, opts)
			}).Timeout(timeout, carapace.ActionValues())
		case 1:
			return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
				return actionBranchesForProject(c.Parts[0], config)
			}).Timeout(timeout, carapace.ActionValues())
		default:
			return carapace.ActionValues()
		}
	}).Cache(5 * time.Second)
}

// actionBranches provides completion for branch names (--source flag)
func actionBranches(config *CommandConfig) carapace.Action {
	timeout := getCompletionTimeout(config.Config)

	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ctx, err := config.Services.ContextService.GetCurrentContext()
		if err != nil {
			return carapace.ActionValues()
		}

		suggestions, err := config.Services.ContextService.GetCompletionSuggestionsFromContext(ctx, c.Value)
		if err != nil {
			return carapace.ActionValues()
		}

		return suggestionsToCarapaceAction(suggestions, config.Config.DefaultSourceBranch)
	}).Timeout(timeout, carapace.ActionValues()).Cache(5 * time.Second)
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

	return suggestionsToCarapaceAction(suggestions, config.Config.DefaultSourceBranch)
}

// actionBranchesForProject suggests branches for a specific project
// Creates a synthetic context for the target project to get correct branch suggestions
func actionBranchesForProject(projectName string, config *CommandConfig) carapace.Action {
	timeout := getCompletionTimeout(config.Config)

	return carapace.ActionCallback(func(_ carapace.Context) carapace.Action {
		// Create synthetic context for the target project
		targetCtx := &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: projectName,
			Path:        filepath.Join(config.Config.ProjectsDirectory, projectName),
		}

		// Get suggestions for the target project
		suggestions, err := config.Services.ContextService.GetCompletionSuggestionsFromContext(targetCtx, "")
		if err != nil {
			return carapace.ActionValues()
		}

		// Filter to only include worktrees and branches for this project
		// (exclude other project suggestions)
		filtered := make([]*domain.ResolutionSuggestion, 0, len(suggestions))
		for _, s := range suggestions {
			if s.ProjectName == projectName && s.Type != domain.PathTypeProject {
				filtered = append(filtered, s)
			}
		}

		return suggestionsToCarapaceAction(filtered, config.Config.DefaultSourceBranch)
	}).Timeout(timeout, carapace.ActionValues()).Cache(5 * time.Second)
}

// sortSuggestions implements smart sorting:
// 1. Current worktree first
// 2. Default branch second
// 3. Other branches alphabetically
func sortSuggestions(suggestions []*domain.ResolutionSuggestion, defaultBranch string) {
	sort.SliceStable(suggestions, func(i, j int) bool {
		si, sj := suggestions[i], suggestions[j]

		// Current worktree always first (but handle multiple current worktrees)
		if si.IsCurrent && !sj.IsCurrent {
			return true
		}
		if !si.IsCurrent && sj.IsCurrent {
			return false
		}

		// If both are current (edge case), use alphabetical
		if si.IsCurrent && sj.IsCurrent {
			return si.Text < sj.Text
		}

		// Default branch second (after current worktrees)
		if si.Text == defaultBranch && sj.Text != defaultBranch {
			return true
		}
		if sj.Text == defaultBranch && si.Text != defaultBranch {
			return false
		}

		// Alphabetical for rest
		return si.Text < sj.Text
	})
}

// suggestionsToCarapaceAction converts domain.ResolutionSuggestion to carapace.Action
// Uses Batch to apply "/" suffix to project suggestions only
func suggestionsToCarapaceAction(suggestions []*domain.ResolutionSuggestion, defaultBranch string) carapace.Action {
	if len(suggestions) == 0 {
		return carapace.ActionValues()
	}

	// Apply smart sorting
	sortSuggestions(suggestions, defaultBranch)

	// Separate projects and branches/worktrees
	var projectValues, branchValues []string
	var projectDescs, branchDescs []string

	for _, s := range suggestions {
		// Projects get "/" suffix for progressive completion
		if s.Type == domain.PathTypeProject && s.BranchName == "" {
			projectValues = append(projectValues, s.Text)
			projectDescs = append(projectDescs, s.Description)
		} else {
			// Branches and worktrees don't get suffix
			branchValues = append(branchValues, s.Text)
			branchDescs = append(branchDescs, s.Description)
		}
	}

	// Use Batch to combine project suggestions (with suffix) and branch suggestions (without suffix)
	var actions []carapace.Action

	// Add project suggestions with "/" suffix
	if len(projectValues) > 0 {
		projectAction := carapace.ActionValues(projectValues...).Suffix("/")
		for _, desc := range projectDescs {
			projectAction = projectAction.Tag(desc)
		}
		actions = append(actions, projectAction)
	}

	// Add branch/worktree suggestions without suffix
	if len(branchValues) > 0 {
		branchAction := carapace.ActionValues(branchValues...)
		for _, desc := range branchDescs {
			branchAction = branchAction.Tag(desc)
		}
		actions = append(actions, branchAction)
	}

	return carapace.Batch(actions...).ToA()
}
