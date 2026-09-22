package fixture

import "twiggit/internal/core"

// NewProjectContext creates a project context fixture
func NewProjectContext() *core.Context {
	return &core.Context{
		Type:        core.ContextProject,
		ProjectName: "test-project",
		Path:        "/home/user/Projects/test-project",
		Explanation: "Project context detected",
	}
}

// NewWorktreeContext creates a worktree context fixture
func NewWorktreeContext() *core.Context {
	return &core.Context{
		Type:        core.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  "feature-branch",
		Path:        "/home/user/Worktrees/test-project/feature-branch",
		Explanation: "Worktree context detected",
	}
}

// NewOutsideGitContext creates an outside git context fixture
func NewOutsideGitContext() *core.Context {
	return &core.Context{
		Type:        core.ContextOutsideGit,
		Path:        "/home/user",
		Explanation: "Outside git context detected",
	}
}

// NewProjectResolutionResult creates a project resolution result fixture
func NewProjectResolutionResult() *core.ResolutionResult {
	return &core.ResolutionResult{
		ResolvedPath: "/home/user/Projects/test-project",
		Type:         core.PathTypeProject,
		ProjectName:  "test-project",
		Explanation:  "Resolved to project path",
	}
}

// NewWorktreeResolutionResult creates a worktree resolution result fixture
func NewWorktreeResolutionResult() *core.ResolutionResult {
	return &core.ResolutionResult{
		ResolvedPath: "/home/user/Worktrees/test-project/feature-branch",
		Type:         core.PathTypeWorktree,
		ProjectName:  "test-project",
		BranchName:   "feature-branch",
		Explanation:  "Resolved to worktree path",
	}
}

// NewMainSuggestion creates a main branch suggestion fixture
func NewMainSuggestion() *core.ResolutionSuggestion {
	return &core.ResolutionSuggestion{
		Text:        "main",
		Description: "Navigate to main branch",
		Type:        core.PathTypeProject,
		ProjectName: "test-project",
	}
}

// NewFeatureSuggestions creates feature branch suggestions fixture
func NewFeatureSuggestions() []*core.ResolutionSuggestion {
	return []*core.ResolutionSuggestion{
		{
			Text:        "feature-branch",
			Description: "Navigate to feature branch",
			Type:        core.PathTypeWorktree,
			ProjectName: "test-project",
			BranchName:  "feature-branch",
		},
	}
}

// NewTestConfig creates a test configuration fixture
func NewTestConfig() *core.Config {
	return &core.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}
}
