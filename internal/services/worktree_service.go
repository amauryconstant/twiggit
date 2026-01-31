package services

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Resolve project
	project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}

	// Calculate worktree path
	worktreePath := s.calculateWorktreePath(project.Name, req.BranchName)

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
		return fmt.Errorf("validation failed: %w", err)
	}

	// Find the project that contains this worktree
	project, err := s.findProjectByWorktree(ctx, req.WorktreePath)
	if err != nil {
		// If worktree is not found in any project, consider it already deleted (idempotent)
		if strings.Contains(err.Error(), "worktree not found in any project") {
			return nil
		}
		return fmt.Errorf("failed to find project for worktree: %w", err)
	}

	// Delete worktree using CLI client
	err = s.gitService.DeleteWorktree(ctx, project.GitRepoPath, req.WorktreePath, req.KeepBranch)
	if err != nil {
		return domain.NewWorktreeServiceError(req.WorktreePath, "", "DeleteWorktree", "failed to delete worktree", err)
	}

	return nil
}

// ListWorktrees lists all worktrees for a project
func (s *worktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	var project *domain.ProjectInfo
	var err error

	// If we're in a project context, use the current path directly
	if req.Context != nil && (req.Context.Type == domain.ContextProject || req.Context.Type == domain.ContextWorktree) {
		project, err = s.projectService.GetProjectInfo(ctx, req.Context.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get project info from context: %w", err)
		}
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
		project, err = s.projectService.DiscoverProject(ctx, projectName, req.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
	}

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

	// Convert to pointers for return
	result := make([]*domain.WorktreeInfo, len(worktrees))
	for i := range worktrees {
		result[i] = &worktrees[i]
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

	// Get worktree info to determine branch
	worktrees, err := s.gitService.ListWorktrees(ctx, worktreePath)
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

	// Check if it's a worktree (not the main repository)
	worktrees, err := s.gitService.ListWorktrees(ctx, worktreePath)
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
	// Sanitize branch name for filesystem
	safeBranchName := filepath.Base(branchName)
	safeBranchName = filepath.Clean(safeBranchName)

	return filepath.Join(s.config.WorktreesDirectory, projectName, safeBranchName)
}

func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
	// For now, we'll use a simple approach - try to find the parent git repository
	// In a real implementation, this might be more sophisticated

	// Get all projects and check which one contains this worktree
	projects, err := s.projectService.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, project := range projects {
		worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath)
		if err != nil {
			continue // Skip this project if we can't list its worktrees
		}

		for _, wt := range worktrees {
			if wt.Path == worktreePath {
				return project, nil
			}
		}
	}

	return nil, domain.NewWorktreeServiceError(worktreePath, "", "findProjectByWorktree", "worktree not found in any project", nil)
}
