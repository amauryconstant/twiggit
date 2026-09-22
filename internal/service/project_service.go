package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

var _ application.ProjectService = (*projectService)(nil)

type projectService struct {
	goGit       application.GoGitClient
	cli         application.CLIClient
	repoLocator application.RepoLocator
	config      *domain.Config
}

func NewProjectService(
	goGit application.GoGitClient,
	cli application.CLIClient,
	repoLocator application.RepoLocator,
	config *domain.Config,
) application.ProjectService {
	return &projectService{
		goGit:       goGit,
		cli:         cli,
		repoLocator: repoLocator,
		config:      config,
	}
}

func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	if projectName != "" {
		return s.discoverProjectByName(ctx, projectName, context)
	}

	if context != nil {
		return s.discoverProjectFromContext(ctx, context)
	}

	return nil, domain.NewValidationError("DiscoverProject", "projectName", "", "project name required when outside git context")
}

func (s *projectService) ValidateProject(_ context.Context, projectPath string) error {
	if projectPath == "" {
		return domain.NewValidationError("ValidateProject", "projectPath", "", "project path cannot be empty")
	}

	err := s.goGit.ValidateRepository(projectPath)
	if err != nil {
		return domain.NewProjectServiceError("", projectPath, "ValidateProject", "invalid git repository", err)
	}

	return nil
}

func (s *projectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	projectsDir := s.config.ProjectsDirectory

	gitDirs, err := s.repoLocator.FindGitRepositories(projectsDir)
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

func (s *projectService) ListProjectSummaries(ctx context.Context) ([]*domain.ProjectSummary, error) {
	projectsDir := s.config.ProjectsDirectory

	gitDirs, err := s.repoLocator.FindGitRepositories(projectsDir)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectsDir, "ListProjectSummaries", "failed to scan for git repositories", err)
	}

	summaries := make([]*domain.ProjectSummary, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		if err := s.ValidateProject(ctx, gitDir.Path); err != nil {
			continue
		}

		mainRepoPath := s.findMainRepoFromWorktree(gitDir.Path)
		projectName := filepath.Base(mainRepoPath)

		summaries = append(summaries, &domain.ProjectSummary{
			Name:        projectName,
			Path:        gitDir.Path,
			GitRepoPath: mainRepoPath,
		})
	}

	return summaries, nil
}

func (s *projectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	if projectPath == "" {
		return nil, domain.NewValidationError("GetProjectInfo", "projectPath", "", "project path cannot be empty")
	}

	if err := s.ValidateProject(ctx, projectPath); err != nil {
		return nil, err
	}

	mainRepoPath := s.findMainRepoFromWorktree(projectPath)

	repoInfo, err := s.goGit.GetRepositoryInfo(ctx, mainRepoPath)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to get repository info", err)
	}

	worktrees, err := s.cli.ListWorktrees(ctx, mainRepoPath)
	if err != nil {
		return nil, domain.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to list worktrees", err)
	}

	worktreePtrs := make([]*domain.WorktreeInfo, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	branchPtrs := make([]*domain.BranchInfo, len(repoInfo.Branches))
	for i := range repoInfo.Branches {
		branchPtrs[i] = &repoInfo.Branches[i]
	}

	remotePtrs := make([]*domain.RemoteInfo, len(repoInfo.Remotes))
	for i := range repoInfo.Remotes {
		remotePtrs[i] = &repoInfo.Remotes[i]
	}

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

func (s *projectService) discoverProjectByName(ctx context.Context, projectName string, currentContext *domain.Context) (*domain.ProjectInfo, error) {
	if currentContext != nil && currentContext.Type == domain.ContextProject {
		if currentContext.ProjectName == projectName {
			return s.GetProjectInfo(ctx, currentContext.Path)
		}
	}

	projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	if err := s.ValidateProject(ctx, projectPath); err != nil {
		return s.searchProjectByName(ctx, projectName)
	}

	return s.GetProjectInfo(ctx, projectPath)
}

func (s *projectService) discoverProjectFromContext(ctx context.Context, context *domain.Context) (*domain.ProjectInfo, error) {
	switch context.Type {
	case domain.ContextProject, domain.ContextWorktree:
		projectPath := context.Path
		if context.Type == domain.ContextWorktree {
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
	projectsDir := s.config.ProjectsDirectory

	if _, err := os.Stat(projectsDir); errors.Is(err, os.ErrNotExist) {
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

	if mainRepo := domain.FindMainRepoByTraversal(worktreePath); mainRepo != "" {
		return mainRepo
	}

	return worktreePath
}

func (s *projectService) findMainRepoFromConfig(worktreePath string) string {
	if s.config == nil || s.config.WorktreesDirectory == "" || s.config.ProjectsDirectory == "" {
		return ""
	}

	projectName, err := domain.ExtractProjectFromWorktreePath(worktreePath, s.config.WorktreesDirectory)
	if err != nil || projectName == "" {
		return ""
	}

	mainRepoPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	if domain.IsMainRepo(mainRepoPath) {
		return mainRepoPath
	}

	return ""
}
