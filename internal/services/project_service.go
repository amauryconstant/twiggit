package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/domain"
	"twiggit/internal/service"
)

// ContextServiceInterface defines the context service operations we need
type ContextServiceInterface interface {
	GetCurrentContext() (*domain.Context, error)
	DetectContextFromPath(path string) (*domain.Context, error)
	ResolveIdentifier(identifier string) (*domain.ResolutionResult, error)
	ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error)
	GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error)
}

// projectService implements ProjectService interface
type projectService struct {
	gitService     service.GitService
	contextService ContextServiceInterface
	config         *domain.Config
}

// NewProjectService creates a new ProjectService instance
func NewProjectService(
	gitService service.GitService,
	contextService ContextServiceInterface,
	config *domain.Config,
) ProjectService {
	return &projectService{
		gitService:     gitService,
		contextService: contextService,
		config:         config,
	}
}

// DiscoverProject discovers a project by name or from context
func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	if projectName != "" {
		return s.discoverProjectByName(ctx, projectName)
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

	// Check if projects directory exists
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return []*domain.ProjectInfo{}, nil
	}

	// Read projects directory
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectsDir, "ListProjects", "failed to read projects directory", err)
	}

	projects := make([]*domain.ProjectInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectName := entry.Name()
		projectPath := filepath.Join(projectsDir, projectName)

		// Validate that it's a git repository
		if err := s.gitService.ValidateRepository(projectPath); err != nil {
			continue // Skip non-git directories
		}

		// Get project info
		projectInfo, err := s.GetProjectInfo(ctx, projectPath)
		if err != nil {
			continue // Skip projects we can't get info for
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

	// Get repository info
	repoInfo, err := s.gitService.GetRepositoryInfo(ctx, projectPath)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to get repository info", err)
	}

	// Get worktrees
	worktrees, err := s.gitService.ListWorktrees(ctx, projectPath)
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

	// Extract project name from path
	projectName := filepath.Base(projectPath)

	return &domain.ProjectInfo{
		Name:          projectName,
		Path:          projectPath,
		GitRepoPath:   projectPath,
		Worktrees:     worktreePtrs,
		Branches:      branchPtrs,
		Remotes:       remotePtrs,
		DefaultBranch: repoInfo.DefaultBranch,
		IsBare:        repoInfo.IsBare,
	}, nil
}

// Private helper methods

func (s *projectService) discoverProjectByName(ctx context.Context, projectName string) (*domain.ProjectInfo, error) {
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
	// Simple implementation: go up directories until we find a .git directory
	// that doesn't contain a gitdir file (which indicates a worktree)
	currentPath := worktreePath

	for {
		gitPath := filepath.Join(currentPath, ".git")

		// Check if .git is a directory (main repo) not a file (worktree)
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			// Check if it's a worktree by looking for gitdir file
			gitdirPath := filepath.Join(gitPath, "gitdir")
			if _, err := os.Stat(gitdirPath); os.IsNotExist(err) {
				// This is the main repository
				return currentPath
			}
		}

		// Go up one directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached root
			break
		}
		currentPath = parent
	}

	// Fallback to worktree path if we can't find main repo
	return worktreePath
}
