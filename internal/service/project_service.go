package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

// projectService implements ProjectService interface
type projectService struct {
	gitService     infrastructure.GitClient
	contextService application.ContextService
	config         *domain.Config
}

// NewProjectService creates a new ProjectService instance
func NewProjectService(
	gitService infrastructure.GitClient,
	contextService application.ContextService,
	config *domain.Config,
) application.ProjectService {
	return &projectService{
		gitService:     gitService,
		contextService: contextService,
		config:         config,
	}
}

// DiscoverProject discovers a project by name or from context
func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	if projectName != "" {
		return s.discoverProjectByName(ctx, projectName, context)
	}

	if context != nil {
		return s.discoverProjectFromContext(ctx, context)
	}

	return nil, domain.NewValidationError("DiscoverProject", "projectName", "", "project name required when outside git context")
}

// ValidateProject validates that a project is properly configured
func (s *projectService) ValidateProject(_ context.Context, projectPath string) error {
	if projectPath == "" {
		return domain.NewValidationError("ValidateProject", "projectPath", "", "project path cannot be empty")
	}

	// Validate that path contains a git repository
	err := s.gitService.ValidateRepository(projectPath)
	if err != nil {
		return domain.NewProjectServiceError("", projectPath, "ValidateProject", "invalid git repository", err)
	}

	return nil
}

// ListProjects lists all available projects
func (s *projectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	projectsDir := s.config.ProjectsDirectory

	gitDirs, err := infrastructure.FindGitRepositories(projectsDir, s.gitService)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectsDir, "ListProjects", "failed to scan for git repositories", err)
	}

	projects := make([]*domain.ProjectInfo, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		projectInfo, err := s.GetProjectInfo(ctx, gitDir.Path)
		if err != nil {
			continue
		}

		projects = append(projects, projectInfo)
	}

	return projects, nil
}

// GetProjectInfo retrieves detailed information about a project
func (s *projectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	if projectPath == "" {
		return nil, domain.NewValidationError("GetProjectInfo", "projectPath", "", "project path cannot be empty")
	}

	// Validate project first
	if err := s.ValidateProject(ctx, projectPath); err != nil {
		return nil, err
	}

	// If projectPath is a worktree, get the main repository path
	mainRepoPath := s.findMainRepoFromWorktree(projectPath)

	// Get repository info
	repoInfo, err := s.gitService.GetRepositoryInfo(ctx, mainRepoPath)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to get repository info", err)
	}

	// Get worktrees
	worktrees, err := s.gitService.ListWorktrees(ctx, mainRepoPath)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to list worktrees", err)
	}

	// Convert worktrees to pointers
	worktreePtrs := make([]*domain.WorktreeInfo, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	// Convert branches to pointers
	branchPtrs := make([]*domain.BranchInfo, len(repoInfo.Branches))
	for i := range repoInfo.Branches {
		branchPtrs[i] = &repoInfo.Branches[i]
	}

	// Convert remotes to pointers
	remotePtrs := make([]*domain.RemoteInfo, len(repoInfo.Remotes))
	for i := range repoInfo.Remotes {
		remotePtrs[i] = &repoInfo.Remotes[i]
	}

	// Extract project name from main repository path
	projectName := filepath.Base(mainRepoPath)

	return &domain.ProjectInfo{
		Name:          projectName,
		Path:          projectPath,
		GitRepoPath:   mainRepoPath,
		Worktrees:     worktreePtrs,
		Branches:      branchPtrs,
		Remotes:       remotePtrs,
		DefaultBranch: repoInfo.DefaultBranch,
		IsBare:        repoInfo.IsBare,
	}, nil
}

// Private helper methods

func (s *projectService) discoverProjectByName(ctx context.Context, projectName string, currentContext *domain.Context) (*domain.ProjectInfo, error) {
	// If in project context and project name matches, use context path
	if currentContext != nil && currentContext.Type == domain.ContextProject {
		if currentContext.ProjectName == projectName {
			return s.GetProjectInfo(ctx, currentContext.Path)
		}
	}

	// Look for project in projects directory
	projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	// Validate project
	if err := s.ValidateProject(ctx, projectPath); err != nil {
		// Try to find it in other locations
		return s.searchProjectByName(ctx, projectName)
	}

	// Get project info
	return s.GetProjectInfo(ctx, projectPath)
}

func (s *projectService) discoverProjectFromContext(ctx context.Context, context *domain.Context) (*domain.ProjectInfo, error) {
	switch context.Type {
	case domain.ContextProject, domain.ContextWorktree:
		// Use the path from context
		projectPath := context.Path
		if context.Type == domain.ContextWorktree {
			// For worktree context, we need to find the main project repository
			projectPath = s.findMainRepoFromWorktree(context.Path)
		}

		return s.GetProjectInfo(ctx, projectPath)

	case domain.ContextOutsideGit:
		return nil, domain.NewValidationError("DiscoverProject", "context", context.Type.String(), "project name required when outside git context")

	default:
		return nil, domain.NewProjectServiceError("", "", "DiscoverProject", "unsupported context type", nil)
	}
}

func (s *projectService) searchProjectByName(ctx context.Context, projectName string) (*domain.ProjectInfo, error) {
	// Search in projects directory
	projectsDir := s.config.ProjectsDirectory

	// Check if projects directory exists
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		// Projects directory doesn't exist, return project not found
		return nil, domain.NewProjectServiceError(projectName, "", "searchProjectByName", "project not found", nil)
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, domain.NewProjectServiceError(projectName, "", "searchProjectByName", "failed to search projects", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if strings.EqualFold(entry.Name(), projectName) {
			projectPath := filepath.Join(projectsDir, entry.Name())
			return s.GetProjectInfo(ctx, projectPath)
		}
	}

	return nil, domain.NewProjectServiceError(projectName, "", "searchProjectByName", "project not found", nil)
}

func (s *projectService) findMainRepoFromWorktree(worktreePath string) string {
	if mainRepo := s.findMainRepoFromConfig(worktreePath); mainRepo != "" {
		return mainRepo
	}

	if mainRepo := infrastructure.FindMainRepoByTraversal(worktreePath); mainRepo != "" {
		return mainRepo
	}

	return worktreePath
}

func (s *projectService) findMainRepoFromConfig(worktreePath string) string {
	if s.config == nil || s.config.WorktreesDirectory == "" || s.config.ProjectsDirectory == "" {
		return ""
	}

	worktreeDir := filepath.Clean(s.config.WorktreesDirectory)
	if !strings.HasPrefix(worktreePath, worktreeDir+string(filepath.Separator)) {
		return ""
	}

	relPath, err := filepath.Rel(worktreeDir, worktreePath)
	if err != nil {
		return ""
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 1 {
		return ""
	}

	projectName := parts[0]
	mainRepoPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	if infrastructure.IsMainRepo(mainRepoPath) {
		return mainRepoPath
	}

	return ""
}
