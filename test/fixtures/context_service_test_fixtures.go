package fixtures

import "twiggit/internal/domain"

// NewProjectContext creates a project context fixture
func NewProjectContext() *domain.Context {
	return &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "test-project",
		Path:        "/home/user/Projects/test-project",
		Explanation: "Project context detected",
	}
}

// NewWorktreeContext creates a worktree context fixture
func NewWorktreeContext() *domain.Context {
	return &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  "feature-branch",
		Path:        "/home/user/Worktrees/test-project/feature-branch",
		Explanation: "Worktree context detected",
	}
}

// NewOutsideGitContext creates an outside git context fixture
func NewOutsideGitContext() *domain.Context {
	return &domain.Context{
		Type:        domain.ContextOutsideGit,
		Path:        "/home/user",
		Explanation: "Outside git context detected",
	}
}

// NewProjectResolutionResult creates a project resolution result fixture
func NewProjectResolutionResult() *domain.ResolutionResult {
	return &domain.ResolutionResult{
		ResolvedPath: "/home/user/Projects/test-project",
		Type:         domain.PathTypeProject,
		ProjectName:  "test-project",
		Explanation:  "Resolved to project path",
	}
}

// NewWorktreeResolutionResult creates a worktree resolution result fixture
func NewWorktreeResolutionResult() *domain.ResolutionResult {
	return &domain.ResolutionResult{
		ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
		Type:         domain.PathTypeWorktree,
		ProjectName:  "test-project",
		BranchName:   "feature-branch",
		Explanation:  "Resolved to worktree path",
	}
}

// NewMainSuggestion creates a main branch suggestion fixture
func NewMainSuggestion() *domain.ResolutionSuggestion {
	return &domain.ResolutionSuggestion{
		Text:        "main",
		Description: "Navigate to main branch",
		Type:        domain.PathTypeProject,
		ProjectName: "test-project",
	}
}

// NewFeatureSuggestions creates feature branch suggestions fixture
func NewFeatureSuggestions() []*domain.ResolutionSuggestion {
	return []*domain.ResolutionSuggestion{
		{
			Text:        "feature-branch",
			Description: "Navigate to feature branch",
			Type:        domain.PathTypeWorktree,
			ProjectName: "test-project",
			BranchName:  "feature-branch",
		},
	}
}

// NewTestConfig creates a test configuration fixture
func NewTestConfig() *domain.Config {
	return &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}
}
