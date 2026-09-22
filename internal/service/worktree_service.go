package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

var _ application.WorktreeService = (*worktreeService)(nil)

type worktreeService struct {
	goGit          application.GoGitClient
	cli            application.CLIClient
	projectService application.ProjectService
	config         *domain.Config
	hookRunner     application.HookRunner
}

func NewWorktreeService(
	goGit application.GoGitClient,
	cli application.CLIClient,
	projectService application.ProjectService,
	config *domain.Config,
	hookRunner application.HookRunner,
) application.WorktreeService {
	return &worktreeService{
		goGit:          goGit,
		cli:            cli,
		projectService: projectService,
		config:         config,
		hookRunner:     hookRunner,
	}
}

func (s *worktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.CreateWorktreeResult, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}

	worktreePath := s.calculateWorktreePath(project.Name, req.BranchName)

	if _, err := os.Stat(worktreePath); err == nil {
		return nil, domain.NewConflictError("worktree", req.BranchName, "CreateWorktree", "worktree already exists at "+worktreePath, nil)
	}

	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil { // #nosec G301 -- standard directory perms (rwxr-xr-x)
		return nil, fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	err = s.cli.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, req.BranchName, "CreateWorktree", "failed to create worktree", err)
	}

	worktreeInfo := &domain.WorktreeInfo{
		Path:   worktreePath,
		Branch: req.BranchName,
	}

	var hookResult *domain.HookResult
	if s.hookRunner != nil {
		hookReq := &application.HookRunRequest{
			HookType:       domain.HookPostCreate,
			WorktreePath:   worktreePath,
			ProjectName:    project.Name,
			BranchName:     req.BranchName,
			SourceBranch:   req.SourceBranch,
			MainRepoPath:   project.GitRepoPath,
			ConfigFilePath: filepath.Join(project.GitRepoPath, ".twiggit.toml"),
		}
		hookResult, err = s.hookRunner.Run(ctx, hookReq)
		if err != nil {
			slog.Error("hook execution failed",
				"error", err,
				"worktree_path", hookReq.WorktreePath,
				"hook_type", string(hookReq.HookType),
			)
		}
	}

	return &domain.CreateWorktreeResult{
		Worktree:   worktreeInfo,
		HookResult: hookResult,
	}, nil
}

func (s *worktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	if err := s.validateDeleteRequest(req); err != nil {
		return err
	}

	project, err := s.findProjectByWorktree(ctx, req.WorktreePath)
	if err != nil {
		var worktreeErr *domain.WorktreeServiceError
		if errors.As(err, &worktreeErr) && worktreeErr.Message == "worktree not found in any project" {
			return nil
		}
		return domain.NewWorktreeServiceError(req.WorktreePath, "", "DeleteWorktree", "failed to find project for worktree", err)
	}

	err = s.cli.DeleteWorktree(ctx, project.GitRepoPath, req.WorktreePath, req.Force)
	if err != nil {
		return domain.NewWorktreeServiceError(req.WorktreePath, "", "DeleteWorktree", "failed to delete worktree", err)
	}

	return nil
}

func (s *worktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	var projects []*domain.ProjectInfo
	var err error

	if req.ListAllProjects {
		projects, err = s.listAllProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list all projects: %w", err)
		}
	} else if req.Context != nil && (req.Context.Type == domain.ContextProject || req.Context.Type == domain.ContextWorktree) {
		project, err := s.projectService.GetProjectInfo(ctx, req.Context.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get project info from context: %w", err)
		}
		projects = []*domain.ProjectInfo{project}
	} else {
		projectName := req.ProjectName
		if projectName == "" && req.Context != nil {
			projectName = req.Context.ProjectName
		}

		if projectName == "" {
			return nil, domain.NewValidationError("ListWorktreesRequest", "projectName", "", "project name required when not provided in context")
		}

		project, err := s.projectService.DiscoverProject(ctx, projectName, req.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
		projects = []*domain.ProjectInfo{project}
	}

	var allWorktrees []*domain.WorktreeInfo
	for _, project := range projects {
		worktrees, err := s.cli.ListWorktrees(ctx, project.GitRepoPath)
		if err != nil {
			return nil, domain.NewWorktreeServiceError(project.GitRepoPath, "", "ListWorktrees", "failed to list worktrees", err)
		}

		if !req.IncludeMain {
			var filtered []domain.WorktreeInfo
			for _, wt := range worktrees {
				if wt.Path != project.GitRepoPath {
					filtered = append(filtered, wt)
				}
			}
			worktrees = filtered
		}

		for i := range worktrees {
			allWorktrees = append(allWorktrees, &worktrees[i])
		}
	}

	return allWorktrees, nil
}

func (s *worktreeService) listAllProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	summaries, err := s.projectService.ListProjectSummaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]*domain.ProjectInfo, len(summaries))
	for i, summary := range summaries {
		result[i] = &domain.ProjectInfo{
			Name:        summary.Name,
			Path:        summary.Path,
			GitRepoPath: summary.GitRepoPath,
		}
	}

	return result, nil
}

func (s *worktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	if worktreePath == "" {
		return nil, domain.NewValidationError("GetWorktreeStatus", "worktreePath", "", "worktree path cannot be empty")
	}

	err := s.ValidateWorktree(ctx, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("worktree validation failed: %w", err)
	}

	repoStatus, err := s.goGit.GetRepositoryStatus(ctx, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "failed to get repository status", err)
	}

	project, err := s.findProjectByWorktree(ctx, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "failed to find parent project", err)
	}

	worktrees, err := s.cli.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "failed to get worktree info", err)
	}

	var worktreeInfo *domain.WorktreeInfo
	for i := range worktrees {
		if worktrees[i].Path == worktreePath {
			worktreeInfo = &worktrees[i]
			break
		}
	}

	if worktreeInfo == nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "worktree not found in list", nil)
	}

	branchStatus := "up-to-date"
	if repoStatus.Ahead > 0 && repoStatus.Behind > 0 {
		branchStatus = "diverged"
	} else if repoStatus.Ahead > 0 {
		branchStatus = "ahead"
	} else if repoStatus.Behind > 0 {
		branchStatus = "behind"
	}

	return &domain.WorktreeStatus{
		WorktreeInfo:          worktreeInfo,
		RepositoryStatus:      &repoStatus,
		LastChecked:           time.Now(),
		IsClean:               repoStatus.IsClean,
		HasUncommittedChanges: !repoStatus.IsClean,
		BranchStatus:          branchStatus,
	}, nil
}

func (s *worktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	if worktreePath == "" {
		return domain.NewValidationError("ValidateWorktree", "worktreePath", "", "worktree path cannot be empty")
	}

	err := s.goGit.ValidateRepository(worktreePath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "invalid git repository", err)
	}

	project, err := s.findProjectByWorktree(ctx, worktreePath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "failed to find parent project", err)
	}

	worktrees, err := s.cli.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "failed to list worktrees", err)
	}

	isWorktree := false
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			isWorktree = true
			break
		}
	}

	if !isWorktree {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "path is not a valid worktree", nil)
	}

	return nil
}

func (s *worktreeService) validateCreateRequest(req *domain.CreateWorktreeRequest) error {
	branchValidation := domain.ValidateBranchName(req.BranchName)
	if branchValidation.IsError() {
		return branchValidation.Error
	}

	if req.Context == nil {
		return domain.NewValidationError("CreateWorktreeRequest", "Context", "", "context is required").
			WithSuggestions([]string{"Run from within a project or worktree directory"})
	}
	if req.ProjectName == "" && req.Context.Type != domain.ContextProject {
		return domain.NewValidationError("CreateWorktreeRequest", "ProjectName", "", "project name required when not in project context").
			WithSuggestions([]string{"Specify a project name (e.g., my-project/feature-branch)", "Run from within a project directory"})
	}

	return nil
}

func (s *worktreeService) validateDeleteRequest(req *domain.DeleteWorktreeRequest) error {
	if req.WorktreePath == "" {
		return domain.NewValidationError("DeleteWorktreeRequest", "WorktreePath", "", "worktree path cannot be empty")
	}

	return nil
}

func (s *worktreeService) calculateWorktreePath(projectName, branchName string) string {
	safeProjectName := filepath.Base(projectName)
	safeProjectName = filepath.Clean(safeProjectName)

	safeBranchName := filepath.Base(branchName)
	safeBranchName = filepath.Clean(safeBranchName)

	return filepath.Join(s.config.WorktreesDirectory, safeProjectName, safeBranchName)
}

func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
	if info, err := s.findProjectFromConfig(ctx, worktreePath); err != nil {
		return nil, err
	} else if info != nil {
		return info, nil
	}

	return s.findProjectByListing(ctx, worktreePath)
}

func (s *worktreeService) findProjectFromConfig(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
	if s.config == nil || s.config.WorktreesDirectory == "" || s.config.ProjectsDirectory == "" {
		return nil, nil
	}

	projectName, err := domain.ExtractProjectFromWorktreePath(worktreePath, s.config.WorktreesDirectory)
	if err != nil || projectName == "" {
		return nil, nil
	}

	projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)

	if _, statErr := os.Stat(projectPath); statErr != nil {
		return nil, nil
	}

	info, err := s.projectService.GetProjectInfo(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get project info for %s: %w", projectPath, err)
	}
	return info, nil
}

func (s *worktreeService) findProjectByListing(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
	projects, err := s.projectService.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		if s.isWorktreeInProject(worktreePath, project) {
			return project, nil
		}
	}

	return nil, domain.NewWorktreeServiceError(worktreePath, "", "findProjectByWorktree", "worktree not found in any project", nil)
}

func (s *worktreeService) isWorktreeInProject(worktreePath string, project *domain.ProjectInfo) bool {
	for _, wt := range project.Worktrees {
		if wt.Path == worktreePath {
			return true
		}
	}
	return false
}

func (s *worktreeService) PruneMergedWorktrees(ctx context.Context, req *domain.PruneWorktreesRequest) (*domain.PruneWorktreesResult, error) {
	if err := s.validatePruneRequest(req); err != nil {
		return nil, err
	}

	var projects []*domain.ProjectInfo

	if req.AllProjects {
		summaries, err := s.projectService.ListProjectSummaries(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}
		projects = make([]*domain.ProjectInfo, len(summaries))
		for i, summary := range summaries {
			projects[i] = &domain.ProjectInfo{
				Name:        summary.Name,
				Path:        summary.Path,
				GitRepoPath: summary.GitRepoPath,
			}
		}
	} else if req.SpecificWorktree != "" {
		parts := strings.Split(req.SpecificWorktree, "/")
		if len(parts) != 2 {
			return nil, domain.NewValidationError("PruneWorktreesRequest", "SpecificWorktree", req.SpecificWorktree, "must be in format project/branch")
		}
		project, err := s.projectService.DiscoverProject(ctx, parts[0], req.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
		projects = []*domain.ProjectInfo{project}
	} else {
		projectName := req.ProjectName
		if projectName == "" && req.Context != nil {
			projectName = req.Context.ProjectName
		}
		if projectName == "" {
			project, err := s.projectService.GetProjectInfo(ctx, req.Context.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to get project info from context: %w", err)
			}
			projects = []*domain.ProjectInfo{project}
		} else {
			project, err := s.projectService.DiscoverProject(ctx, projectName, req.Context)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve project: %w", err)
			}
			projects = []*domain.ProjectInfo{project}
		}
	}

	result := &domain.PruneWorktreesResult{
		DeletedWorktrees:     []*domain.PruneWorktreeResult{},
		SkippedWorktrees:     []*domain.PruneWorktreeResult{},
		ProtectedSkipped:     []*domain.PruneWorktreeResult{},
		UnmergedSkipped:      []*domain.PruneWorktreeResult{},
		TotalDeleted:         0,
		TotalSkipped:         0,
		TotalBranchesDeleted: 0,
	}

	singleWorktreeTarget := ""
	if req.SpecificWorktree != "" {
		parts := strings.Split(req.SpecificWorktree, "/")
		singleWorktreeTarget = parts[1]
	}

	for _, project := range projects {
		s.pruneProjectWorktrees(ctx, req, project, result, singleWorktreeTarget)
	}

	if len(result.DeletedWorktrees) == 1 && req.SpecificWorktree != "" {
		projectName := strings.Split(req.SpecificWorktree, "/")[0]
		projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)
		if _, statErr := os.Stat(projectPath); statErr == nil {
			result.NavigationPath = projectPath
		}
	}

	return result, nil
}

func (s *worktreeService) validatePruneRequest(req *domain.PruneWorktreesRequest) error {
	if req.SpecificWorktree != "" && req.AllProjects {
		return domain.NewValidationError("PruneWorktreesRequest", "AllProjects", "true", "cannot use --all with specific worktree")
	}
	return nil
}

func (s *worktreeService) pruneProjectWorktrees(ctx context.Context, req *domain.PruneWorktreesRequest, project *domain.ProjectInfo, result *domain.PruneWorktreesResult, singleWorktreeTarget string) {
	worktrees, err := s.cli.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return
	}

	cwd, _ := os.Getwd()

	for _, wt := range worktrees {
		if wt.Path == project.GitRepoPath {
			continue
		}

		if singleWorktreeTarget != "" && wt.Branch != singleWorktreeTarget {
			continue
		}

		pruneResult := &domain.PruneWorktreeResult{
			ProjectName:   project.Name,
			WorktreePath:  wt.Path,
			BranchName:    wt.Branch,
			Deleted:       false,
			BranchDeleted: false,
		}

		if skip := s.checkWorktreeSkip(ctx, wt, project, cwd, req); skip != nil {
			s.addSkippedResult(result, pruneResult, skip)
			continue
		}

		s.deleteWorktreeAndBranch(ctx, project, wt, req, pruneResult, result)
	}
}

type worktreeSkipResult struct {
	reason   string
	err      error
	category string
}

func (s *worktreeService) checkWorktreeSkip(ctx context.Context, wt domain.WorktreeInfo, project *domain.ProjectInfo, cwd string, req *domain.PruneWorktreesRequest) *worktreeSkipResult {
	if cwd != "" && (strings.HasPrefix(cwd, wt.Path+string(filepath.Separator)) || cwd == wt.Path) {
		return &worktreeSkipResult{reason: "cannot prune current worktree", category: "current"}
	}

	if s.isProtectedBranch(wt.Branch) {
		return &worktreeSkipResult{reason: "protected branch", category: "protected"}
	}

	isMerged, err := s.cli.IsBranchMerged(ctx, project.GitRepoPath, wt.Branch)
	if err != nil {
		return &worktreeSkipResult{reason: "failed to check merge status", err: err, category: "skipped"}
	}

	if !isMerged {
		return &worktreeSkipResult{reason: "branch not merged", category: "unmerged"}
	}

	if !req.Force && !req.DryRun {
		status, err := s.goGit.GetRepositoryStatus(ctx, wt.Path)
		if err == nil && !status.IsClean {
			return &worktreeSkipResult{reason: "uncommitted changes (use --force to override)", category: "skipped"}
		}
	}

	if req.DryRun {
		return &worktreeSkipResult{reason: "dry run", category: "skipped"}
	}

	return nil
}

func (s *worktreeService) addSkippedResult(result *domain.PruneWorktreesResult, pruneResult *domain.PruneWorktreeResult, skip *worktreeSkipResult) {
	pruneResult.SkipReason = skip.reason
	pruneResult.Error = skip.err

	switch skip.category {
	case "current":
		result.CurrentWorktreeSkipped = append(result.CurrentWorktreeSkipped, pruneResult)
	case "protected":
		result.ProtectedSkipped = append(result.ProtectedSkipped, pruneResult)
	case "unmerged":
		result.UnmergedSkipped = append(result.UnmergedSkipped, pruneResult)
	default:
		result.SkippedWorktrees = append(result.SkippedWorktrees, pruneResult)
	}
	result.TotalSkipped++
}

func (s *worktreeService) deleteWorktreeAndBranch(ctx context.Context, project *domain.ProjectInfo, wt domain.WorktreeInfo, req *domain.PruneWorktreesRequest, pruneResult *domain.PruneWorktreeResult, result *domain.PruneWorktreesResult) {
	err := s.cli.DeleteWorktree(ctx, project.GitRepoPath, wt.Path, req.Force)
	if err != nil {
		pruneResult.Error = err
		result.SkippedWorktrees = append(result.SkippedWorktrees, pruneResult)
		result.TotalSkipped++
		return
	}

	pruneResult.Deleted = true
	result.DeletedWorktrees = append(result.DeletedWorktrees, pruneResult)
	result.TotalDeleted++

	if req.DeleteBranches {
		if err := s.cli.PruneWorktrees(ctx, project.GitRepoPath); err != nil {
			slog.Error("prune worktrees failed",
				"error", err,
				"repo_path", project.GitRepoPath,
			)
		}
		err = s.cli.DeleteBranch(ctx, project.GitRepoPath, wt.Branch)
		if err != nil {
			pruneResult.Error = fmt.Errorf("worktree deleted but branch deletion failed: %w", err)
		} else {
			pruneResult.BranchDeleted = true
			result.TotalBranchesDeleted++
		}
	}
}

func (s *worktreeService) isProtectedBranch(branchName string) bool {
	return slices.Contains(s.config.Validation.ProtectedBranches, branchName)
}

func (s *worktreeService) BranchExists(ctx context.Context, projectPath string, branchName string) (bool, error) {
	exists, err := s.goGit.BranchExists(ctx, projectPath, branchName)
	if err != nil {
		return false, domain.NewWorktreeServiceError(projectPath, branchName, "BranchExists", "failed to check branch existence", err)
	}
	return exists, nil
}

func (s *worktreeService) IsBranchMerged(ctx context.Context, worktreePath string, branchName string) (bool, error) {
	merged, err := s.cli.IsBranchMerged(ctx, worktreePath, branchName)
	if err != nil {
		return false, domain.NewWorktreeServiceError(worktreePath, branchName, "IsBranchMerged", "failed to check merge status", err)
	}
	return merged, nil
}

func (s *worktreeService) GetWorktreeByPath(ctx context.Context, projectPath, worktreePath string) (*domain.WorktreeInfo, error) {
	worktrees, err := s.cli.ListWorktrees(ctx, projectPath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeByPath", "failed to list worktrees", err)
	}

	for i := range worktrees {
		if worktrees[i].Path == worktreePath {
			return &worktrees[i], nil
		}
	}

	return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeByPath", "worktree not found", nil)
}
