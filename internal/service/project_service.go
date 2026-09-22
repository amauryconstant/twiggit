package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

var _ application.ProjectService = (*projectService)(nil)

type projectService struct {
	goGit       application.GoGitClient
	cli         application.CLIClient
	repoLocator application.RepoLocator
	config      *core.Config
}

func NewProjectService(
	goGit application.GoGitClient,
	cli application.CLIClient,
	repoLocator application.RepoLocator,
	config *core.Config,
) application.ProjectService {
	return &projectService{
		goGit:       goGit,
		cli:         cli,
		repoLocator: repoLocator,
		config:      config,
	}
}

func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *core.Context) (*core.ProjectInfo, error) {
	if projectName != "" {
		return s.discoverProjectByName(ctx, projectName, context)
	}

	if context != nil {
		return s.discoverProjectFromContext(ctx, context)
	}

	return nil, core.NewOpValidationError("DiscoverProject", "projectName", "", "project name required when outside git context")
}

func (s *projectService) ValidateProject(_ context.Context, projectPath string) error {
	if projectPath == "" {
		return core.NewOpValidationError("ValidateProject", "projectPath", "", "project path cannot be empty")
	}

	err := s.goGit.ValidateRepository(projectPath)
	if err != nil {
		return core.NewProjectServiceError("", projectPath, "ValidateProject", "invalid git repository", err)
	}

	return nil
}

func (s *projectService) ListProjects(ctx context.Context) ([]*core.ProjectInfo, error) {
	projectsDir := s.config.ProjectsDirectory

	gitDirs, err := s.repoLocator.FindGitRepositories(projectsDir)
	if err != nil {
		return nil, core.NewProjectServiceError("", projectsDir, "ListProjects", "failed to scan for git repositories", err)
	}

	projects := make([]*core.ProjectInfo, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		projectInfo, err := s.GetProjectInfo(ctx, gitDir.Path)
		if err != nil {
			continue
		}

		projects = append(projects, projectInfo)
	}

	return projects, nil
}

func (s *projectService) ListProjectSummaries(ctx context.Context) ([]*core.ProjectSummary, error) {
	projectsDir := s.config.ProjectsDirectory

	gitDirs, err := s.repoLocator.FindGitRepositories(projectsDir)
	if err != nil {
		return nil, core.NewProjectServiceError("", projectsDir, "ListProjectSummaries", "failed to scan for git repositories", err)
	}

	summaries := make([]*core.ProjectSummary, 0, len(gitDirs))
	for _, gitDir := range gitDirs {
		if err := s.ValidateProject(ctx, gitDir.Path); err != nil {
			continue
		}

		mainRepoPath := s.findMainRepoFromWorktree(gitDir.Path)
		projectName := filepath.Base(mainRepoPath)

		summaries = append(summaries, &core.ProjectSummary{
			Name:        projectName,
			Path:        gitDir.Path,
			GitRepoPath: mainRepoPath,
		})
	}

	return summaries, nil
}

func (s *projectService) GetProjectInfo(ctx context.Context, projectPath string) (*core.ProjectInfo, error) {
	if projectPath == "" {
		return nil, core.NewOpValidationError("GetProjectInfo", "projectPath", "", "project path cannot be empty")
	}

	if err := s.ValidateProject(ctx, projectPath); err != nil {
		return nil, err
	}

	mainRepoPath := s.findMainRepoFromWorktree(projectPath)

	repoInfo, err := s.goGit.GetRepositoryInfo(ctx, mainRepoPath)
	if err != nil {
		return nil, core.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to get repository info", err)
	}

	worktrees, err := s.cli.ListWorktrees(ctx, mainRepoPath)
	if err != nil {
		return nil, core.NewProjectServiceError("", projectPath, "GetProjectInfo", "failed to list worktrees", err)
	}

	worktreePtrs := make([]*core.WorktreeInfo, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	branchPtrs := make([]*core.BranchInfo, len(repoInfo.Branches))
	for i := range repoInfo.Branches {
		branchPtrs[i] = &repoInfo.Branches[i]
	}

	remotePtrs := make([]*core.RemoteInfo, len(repoInfo.Remotes))
	for i := range repoInfo.Remotes {
		remotePtrs[i] = &repoInfo.Remotes[i]
	}

	projectName := filepath.Base(mainRepoPath)

	return &core.ProjectInfo{
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

func (s *projectService) discoverProjectByName(ctx context.Context, projectName string, currentContext *core.Context) (*core.ProjectInfo, error) {
	if currentContext != nil && currentContext.Type == core.ContextProject {
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

func (s *projectService) discoverProjectFromContext(ctx context.Context, context *core.Context) (*core.ProjectInfo, error) {
	switch context.Type {
	case core.ContextProject, core.ContextWorktree:
		projectPath := context.Path
		if context.Type == core.ContextWorktree {
			projectPath = s.findMainRepoFromWorktree(context.Path)
		}

		return s.GetProjectInfo(ctx, projectPath)

	case core.ContextOutsideGit:
		return nil, core.NewOpValidationError("DiscoverProject", "context", context.Type.String(), "project name required when outside git context")

	default:
		return nil, core.NewProjectServiceError("", "", "DiscoverProject", "unsupported context type", nil)
	}
}

func (s *projectService) searchProjectByName(ctx context.Context, projectName string) (*core.ProjectInfo, error) {
	projectsDir := s.config.ProjectsDirectory

	if _, err := os.Stat(projectsDir); errors.Is(err, os.ErrNotExist) {
		return nil, core.NewProjectServiceError(projectName, "", "searchProjectByName", "project not found", nil)
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, core.NewProjectServiceError(projectName, "", "searchProjectByName", "failed to search projects", err)
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

	return nil, core.NewProjectServiceError(projectName, "", "searchProjectByName", "project not found", nil)
}

func (s *projectService) findMainRepoFromWorktree(worktreePath string) string {
	if mainRepo := s.findMainRepoFromConfig(worktreePath); mainRepo != "" {
		return mainRepo
	}

	if mainRepo := core.FindMainRepoByTraversal(worktreePath); mainRepo != "" {
		return mainRepo
	}

	return worktreePath
}

func (s *projectService) findMainRepoFromConfig(worktreePath string) string {
	if s.config == nil || s.config.WorktreesDirectory == "" || s.config.ProjectsDirectory == "" {
		return ""
	}

	projectName, err := core.ExtractProjectFromWorktreePath(worktreePath, s.config.WorktreesDirectory)
	if err != nil || projectName == "" {
		return ""
	}

	mainRepoPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	if core.IsMainRepo(mainRepoPath) {
		return mainRepoPath
	}

	return ""
}
