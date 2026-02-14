package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

// worktreeService implements WorktreeService interface
type worktreeService struct {
	gitService     infrastructure.GitClient
	projectService application.ProjectService
	config         *domain.Config
}

// NewWorktreeService creates a new WorktreeService instance
func NewWorktreeService(
	gitService infrastructure.GitClient,
	projectService application.ProjectService,
	config *domain.Config,
) application.WorktreeService {
	return &worktreeService{
		gitService:     gitService,
		projectService: projectService,
		config:         config,
	}
}

// CreateWorktree creates a new worktree for the specified project and branch
func (s *worktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Resolve project
	project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}

	// Calculate worktree path
	worktreePath := s.calculateWorktreePath(project.Name, req.BranchName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return nil, domain.NewConflictError("worktree", req.BranchName, "CreateWorktree", "worktree already exists at "+worktreePath, nil)
	}

	// Ensure parent directories exist
	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	// Create worktree using CLI client
	err = s.gitService.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, req.BranchName, "CreateWorktree", "failed to create worktree", err)
	}

	// Return worktree info
	return &domain.WorktreeInfo{
		Path:   worktreePath,
		Branch: req.BranchName,
	}, nil
}

// DeleteWorktree deletes an existing worktree
func (s *worktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	// Validate request
	if err := s.validateDeleteRequest(req); err != nil {
		return err
	}

	// Find the project that contains this worktree
	project, err := s.findProjectByWorktree(ctx, req.WorktreePath)
	if err != nil {
		// If worktree is not found in any project, consider it already deleted (idempotent)
		var worktreeErr *domain.WorktreeServiceError
		if errors.As(err, &worktreeErr) && worktreeErr.Message == "worktree not found in any project" {
			return nil
		}
		return domain.NewWorktreeServiceError(req.WorktreePath, "", "DeleteWorktree", "failed to find project for worktree", err)
	}

	// Delete worktree using CLI client
	err = s.gitService.DeleteWorktree(ctx, project.GitRepoPath, req.WorktreePath, req.Force)
	if err != nil {
		return domain.NewWorktreeServiceError(req.WorktreePath, "", "DeleteWorktree", "failed to delete worktree", err)
	}

	return nil
}

// ListWorktrees lists all worktrees for a project
func (s *worktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	var projects []*domain.ProjectInfo
	var err error

	// Handle ListAllProjects case
	if req.ListAllProjects {
		projects, err = s.listAllProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list all projects: %w", err)
		}
	} else if req.Context != nil && (req.Context.Type == domain.ContextProject || req.Context.Type == domain.ContextWorktree) {
		// If we're in a project context, use the current path directly
		project, err := s.projectService.GetProjectInfo(ctx, req.Context.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get project info from context: %w", err)
		}
		projects = []*domain.ProjectInfo{project}
	} else {
		// Resolve project by name
		projectName := req.ProjectName
		if projectName == "" && req.Context != nil {
			projectName = req.Context.ProjectName
		}

		if projectName == "" {
			return nil, domain.NewValidationError("ListWorktreesRequest", "projectName", "", "project name required when not provided in context")
		}

		// Get project info
		project, err := s.projectService.DiscoverProject(ctx, projectName, req.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
		projects = []*domain.ProjectInfo{project}
	}

	// Aggregate worktrees from all projects
	var allWorktrees []*domain.WorktreeInfo
	for _, project := range projects {
		// List worktrees using CLI client
		worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
		if err != nil {
			return nil, domain.NewWorktreeServiceError(project.GitRepoPath, "", "ListWorktrees", "failed to list worktrees", err)
		}

		// Filter out main worktree if not requested
		if !req.IncludeMain {
			var filtered []domain.WorktreeInfo
			for _, wt := range worktrees {
				if wt.Path != project.GitRepoPath {
					filtered = append(filtered, wt)
				}
			}
			worktrees = filtered
		}

		// Convert to pointers and add to result
		for i := range worktrees {
			allWorktrees = append(allWorktrees, &worktrees[i])
		}
	}

	return allWorktrees, nil
}

// listAllProjects retrieves all available projects from the projects directory
func (s *worktreeService) listAllProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	projects, err := s.projectService.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]*domain.ProjectInfo, len(projects))
	for i, project := range projects {
		result[i] = &domain.ProjectInfo{
			Name:        project.Name,
			Path:        project.Path,
			GitRepoPath: project.GitRepoPath,
		}
	}

	return result, nil
}

// GetWorktreeStatus retrieves the status of a specific worktree
func (s *worktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	// Validate input
	if worktreePath == "" {
		return nil, domain.NewValidationError("GetWorktreeStatus", "worktreePath", "", "worktree path cannot be empty")
	}

	// Validate that worktree exists
	err := s.ValidateWorktree(ctx, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("worktree validation failed: %w", err)
	}

	// Get repository status
	repoStatus, err := s.gitService.GetRepositoryStatus(ctx, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "failed to get repository status", err)
	}

	// Find the project that contains this worktree to list worktrees from main repo
	project, err := s.findProjectByWorktree(ctx, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeServiceError(worktreePath, "", "GetWorktreeStatus", "failed to find parent project", err)
	}

	// Get worktree info to determine branch
	worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
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

	// Determine branch status
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

// ValidateWorktree validates that a worktree is properly configured
func (s *worktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	if worktreePath == "" {
		return domain.NewValidationError("ValidateWorktree", "worktreePath", "", "worktree path cannot be empty")
	}

	// Validate that path contains a git repository
	err := s.gitService.ValidateRepository(worktreePath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "invalid git repository", err)
	}

	// Find the project that contains this worktree to list worktrees from main repo
	project, err := s.findProjectByWorktree(ctx, worktreePath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "failed to find parent project", err)
	}

	// List worktrees from main repository
	worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return domain.NewWorktreeServiceError(worktreePath, "", "ValidateWorktree", "failed to list worktrees", err)
	}

	// Find this worktree in the list
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

// Private helper methods

func (s *worktreeService) validateCreateRequest(req *domain.CreateWorktreeRequest) error {
	// Use functional validation pipeline for branch name
	branchValidation := domain.ValidateBranchName(req.BranchName)
	if branchValidation.IsError() {
		return branchValidation.Error
	}

	// Validate project context
	if req.ProjectName == "" && req.Context != nil && req.Context.Type != domain.ContextProject {
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

	worktreeDir := filepath.Clean(s.config.WorktreesDirectory)
	if !strings.HasPrefix(worktreePath, worktreeDir+string(filepath.Separator)) {
		return nil, nil
	}

	relPath, err := filepath.Rel(worktreeDir, worktreePath)
	if err != nil {
		return nil, nil
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 1 {
		return nil, nil
	}

	projectName := parts[0]
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
		if s.isWorktreeInProject(ctx, worktreePath, project) {
			return project, nil
		}
	}

	return nil, domain.NewWorktreeServiceError(worktreePath, "", "findProjectByWorktree", "worktree not found in any project", nil)
}

func (s *worktreeService) isWorktreeInProject(ctx context.Context, worktreePath string, project *domain.ProjectInfo) bool {
	worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return false
	}

	for _, wt := range worktrees {
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
	var err error

	if req.AllProjects {
		projects, err = s.projectService.ListProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
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
	worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
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

	isMerged, err := s.gitService.IsBranchMerged(ctx, project.GitRepoPath, wt.Branch)
	if err != nil {
		return &worktreeSkipResult{reason: "failed to check merge status", err: err, category: "skipped"}
	}

	if !isMerged {
		return &worktreeSkipResult{reason: "branch not merged", category: "unmerged"}
	}

	if !req.Force && !req.DryRun {
		status, err := s.gitService.GetRepositoryStatus(ctx, wt.Path)
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
	err := s.gitService.DeleteWorktree(ctx, project.GitRepoPath, wt.Path, req.Force)
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
		_ = s.gitService.PruneWorktrees(ctx, project.GitRepoPath)
		err = s.gitService.DeleteBranch(ctx, project.GitRepoPath, wt.Branch)
		if err != nil {
			pruneResult.Error = fmt.Errorf("worktree deleted but branch deletion failed: %w", err)
		} else {
			pruneResult.BranchDeleted = true
			result.TotalBranchesDeleted++
		}
	}
}

func (s *worktreeService) isProtectedBranch(branchName string) bool {
	protectedBranches := s.config.Validation.ProtectedBranches
	for _, protected := range protectedBranches {
		if branchName == protected {
			return true
		}
	}
	return false
}
