# Core Services Layer Implementation Plan

## Overview

This plan implements the core services layer that orchestrates domain logic and provides the main business functionality for twiggit. The services layer coordinates between domain entities, infrastructure components, and CLI commands while maintaining clean separation of concerns.

**Context**: Foundation, configuration, context detection, and hybrid git layers are established. This layer provides the orchestration logic that ties everything together.

**Key Principle**: "Services should be thin orchestration layers" - business logic in domain entities, services coordinate.

## Architecture Overview

### Service Layer Responsibilities

> **From design.md**: "Services layer coordinates between domain, infrastructure, and CLI layers"
> - **WorktreeService**: create/delete/list with safety checks
> - **ProjectService**: project discovery and validation  
> - **NavigationService**: path resolution for cd command
> - **Context integration**: services adapt behavior based on detected context

### Service Design Principles

1. **Thin Orchestration**: Services coordinate, don't contain business logic
2. **Dependency Injection**: All external dependencies injected via interfaces
3. **Context-Aware**: Behavior adapts based on ContextDetector results
4. **Atomic Operations**: Proper error recovery and rollback
5. **Validation Layer**: Input validation and safety checks

## Core Service Interfaces

### 1. WorktreeService Interface

```go
// internal/services/interfaces.go
type WorktreeService interface {
    // CreateWorktree creates a new worktree with safety checks
    CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error)
    
    // DeleteWorktree removes a worktree with safety checks
    DeleteWorktree(ctx context.Context, req *DeleteWorktreeRequest) error
    
    // ListWorktrees lists worktrees based on context
    ListWorktrees(ctx context.Context, req *ListWorktreesRequest) ([]*WorktreeInfo, error)
    
    // GetWorktreeStatus returns status of a specific worktree
    GetWorktreeStatus(ctx context.Context, worktreePath string) (*WorktreeStatus, error)
    
    // ValidateWorktree checks if worktree is safe for operations
    ValidateWorktree(ctx context.Context, worktreePath string) error
}

type CreateWorktreeRequest struct {
    ProjectName    string
    BranchName     string
    SourceBranch   string
    ChangeDir      bool
    Force          bool
    Context        *domain.Context
}

type DeleteWorktreeRequest struct {
    ProjectName    string
    BranchName     string
    WorktreePath   string
    KeepBranch     bool
    Force          bool
    ChangeDir      bool
    Context        *domain.Context
}

type ListWorktreesRequest struct {
    ProjectName    string
    AllProjects    bool
    Context        *domain.Context
}

type WorktreeInfo struct {
    Path           string
    Branch         string
    Commit         string
    CommitMessage  string
    Author         string
    Date           time.Time
    Dirty          bool
    Staged         int
    Modified       int
    Untracked      int
    IsCurrent      bool
}
```

### 2. ProjectService Interface

```go
type ProjectService interface {
    // DiscoverProject finds project based on context or name
    DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*ProjectInfo, error)
    
    // ValidateProject checks if project is valid for operations
    ValidateProject(ctx context.Context, projectPath string) error
    
    // ListProjects returns all discovered projects
    ListProjects(ctx context.Context) ([]*ProjectInfo, error)
    
    // GetProjectInfo returns detailed project information
    GetProjectInfo(ctx context.Context, projectPath string) (*ProjectInfo, error)
}

type ProjectInfo struct {
    Name           string
    Path           string
    GitRepoPath    string
    Worktrees      []*WorktreeInfo
    DefaultBranch  string
    RemoteURL      string
    LastActivity   time.Time
    IsValid        bool
}
```

### 3. NavigationService Interface

```go
type NavigationService interface {
    // ResolvePath resolves target path based on context and input
    ResolvePath(ctx context.Context, req *ResolvePathRequest) (*PathResolution, error)
    
    // ValidatePath checks if path is accessible and valid
    ValidatePath(ctx context.Context, path string) error
    
    // GetNavigationSuggestions provides completion suggestions
    GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*PathSuggestion, error)
}

type ResolvePathRequest struct {
    Target         string // Can be branch, project, or project/branch
    Context        *domain.Context
    CurrentPath    string
}

type PathResolution struct {
    ResolvedPath   string
    Type           PathType
    ProjectName    string
    BranchName     string
    Explanation    string
}

type PathType int

const (
    PathTypeProject PathType = iota
    PathTypeWorktree
    PathTypeInvalid
)

type PathSuggestion struct {
    Text           string
    Description    string
    Type           PathType
    ProjectName    string
    BranchName     string
}
```

## Implementation Components

### 1. WorktreeService Implementation

```go
// internal/services/worktree_service.go
type worktreeService struct {
    gitClient      infrastructure.HybridGitClient
    projectService ProjectService
    contextService *domain.ContextService
    config         *domain.Config
    validator      *WorktreeValidator
}

func NewWorktreeService(
    gitClient infrastructure.HybridGitClient,
    projectService ProjectService,
    contextService *domain.ContextService,
    config *domain.Config,
) WorktreeService {
    return &worktreeService{
        gitClient:      gitClient,
        projectService: projectService,
        contextService: contextService,
        config:         config,
        validator:      NewWorktreeValidator(config),
    }
}

func (s *worktreeService) CreateWorktree(ctx context.Context, req *CreateWorktreeRequest) (*WorktreeInfo, error) {
    // 1. Resolve project information
    project, err := s.resolveProjectForCreate(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve project: %w", err)
    }
    
    // 2. Validate request
    if err := s.validator.ValidateCreateRequest(ctx, req, project); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 3. Determine worktree path
    worktreePath := s.calculateWorktreePath(project.Name, req.BranchName)
    
    // 4. Create worktree using hybrid git client
    err = s.gitClient.CreateWorktree(ctx, project.GitRepoPath, req.BranchName, req.SourceBranch, worktreePath)
    if err != nil {
        return nil, fmt.Errorf("failed to create worktree: %w", err)
    }
    
    // 5. Get worktree information
    worktreeInfo, err := s.getWorktreeInfo(ctx, worktreePath)
    if err != nil {
        // Worktree created but info failed - still consider success
        return &WorktreeInfo{
            Path:   worktreePath,
            Branch: req.BranchName,
        }, nil
    }
    
    return worktreeInfo, nil
}

func (s *worktreeService) DeleteWorktree(ctx context.Context, req *DeleteWorktreeRequest) error {
    // 1. Resolve project and worktree
    project, worktreeInfo, err := s.resolveProjectAndWorktree(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to resolve project/worktree: %w", err)
    }
    
    // 2. Safety checks
    if err := s.validator.ValidateDeleteRequest(ctx, req, project, worktreeInfo); err != nil {
        return fmt.Errorf("safety check failed: %w", err)
    }
    
    // 3. Delete worktree using hybrid git client
    err = s.gitClient.DeleteWorktree(ctx, project.GitRepoPath, worktreeInfo.Path, req.KeepBranch)
    if err != nil {
        return fmt.Errorf("failed to delete worktree: %w", err)
    }
    
    return nil
}

func (s *worktreeService) ListWorktrees(ctx context.Context, req *ListWorktreesRequest) ([]*WorktreeInfo, error) {
    if req.AllProjects {
        return s.listAllWorktrees(ctx)
    }
    
    // Resolve project based on context or name
    project, err := s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
    if err != nil {
        return nil, fmt.Errorf("failed to discover project: %w", err)
    }
    
    // List worktrees for specific project
    worktrees, err := s.gitClient.ListWorktrees(ctx, project.GitRepoPath)
    if err != nil {
        return nil, fmt.Errorf("failed to list worktrees: %w", err)
    }
    
    // Enrich with additional information
    return s.enrichWorktreeInfo(ctx, worktrees, project), nil
}

func (s *worktreeService) resolveProjectForCreate(ctx context.Context, req *CreateWorktreeRequest) (*ProjectInfo, error) {
    // Priority order for project resolution:
    // 1. Explicit project name in request
    // 2. Project from context
    // 3. Error if neither available
    
    if req.ProjectName != "" {
        return s.projectService.DiscoverProject(ctx, req.ProjectName, req.Context)
    }
    
    if req.Context != nil && req.Context.Type == domain.ContextProject {
        return s.projectService.DiscoverProject(ctx, req.Context.ProjectName, req.Context)
    }
    
    if req.Context != nil && req.Context.Type == domain.ContextWorktree {
        return s.projectService.DiscoverProject(ctx, req.Context.ProjectName, req.Context)
    }
    
    return nil, fmt.Errorf("project name required when outside git context")
}

func (s *worktreeService) calculateWorktreePath(projectName, branchName string) string {
    return filepath.Join(s.config.WorktreesDirectory, projectName, branchName)
}
```

### 2. WorktreeValidator Implementation

```go
// internal/services/worktree_validator.go
type WorktreeValidator struct {
    config *domain.Config
}

func NewWorktreeValidator(config *domain.Config) *WorktreeValidator {
    return &WorktreeValidator{config: config}
}

func (v *WorktreeValidator) ValidateCreateRequest(ctx context.Context, req *CreateWorktreeRequest, project *ProjectInfo) error {
    var errors []error
    
    // Validate branch name
    if err := v.validateBranchName(req.BranchName); err != nil {
        errors = append(errors, err)
    }
    
    // Validate source branch exists
    if err := v.validateSourceBranch(ctx, req.SourceBranch, project); err != nil {
        errors = append(errors, err)
    }
    
    // Check if worktree already exists
    worktreePath := filepath.Join(v.config.WorktreesDirectory, project.Name, req.BranchName)
    if _, err := os.Stat(worktreePath); err == nil {
        errors = append(errors, fmt.Errorf("worktree already exists at %s", worktreePath))
    }
    
    // Validate worktree directory exists
    if _, err := os.Stat(v.config.WorktreesDirectory); os.IsNotExist(err) {
        errors = append(errors, fmt.Errorf("worktrees directory does not exist: %s", v.config.WorktreesDirectory))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %v", errors)
    }
    
    return nil
}

func (v *WorktreeValidator) ValidateDeleteRequest(ctx context.Context, req *DeleteWorktreeRequest, project *ProjectInfo, worktree *WorktreeInfo) error {
    // Safety check: uncommitted changes
    if !req.Force && worktree.Dirty {
        return fmt.Errorf("worktree has uncommitted changes. Use --force to override")
    }
    
    // Safety check: current worktree
    currentDir, err := os.Getwd()
    if err == nil && currentDir == worktree.Path {
        return fmt.Errorf("cannot delete currently active worktree")
    }
    
    // Safety check: merged status if --merged-only specified
    // This would require checking if branch is merged
    
    return nil
}

func (v *WorktreeValidator) validateBranchName(branchName string) error {
    if branchName == "" {
        return fmt.Errorf("branch name cannot be empty")
    }
    
    // Git branch name validation rules
    if strings.HasPrefix(branchName, "-") {
        return fmt.Errorf("branch name cannot start with '-'")
    }
    
    if strings.Contains(branchName, "..") {
        return fmt.Errorf("branch name cannot contain '..'")
    }
    
    // Additional validation rules as needed
    return nil
}

func (v *WorktreeValidator) validateSourceBranch(ctx context.Context, sourceBranch string, project *ProjectInfo) error {
    // Check if source branch exists in project
    // This would use the git client to list branches
    return nil
}
```

### 3. ProjectService Implementation

```go
// internal/services/project_service.go
type projectService struct {
    gitClient      infrastructure.HybridGitClient
    contextService *domain.ContextService
    config         *domain.Config
    cache          map[string]*ProjectInfo
    cacheMu        sync.RWMutex
}

func NewProjectService(
    gitClient infrastructure.HybridGitClient,
    contextService *domain.ContextService,
    config *domain.Config,
) ProjectService {
    return &projectService{
        gitClient:      gitClient,
        contextService: contextService,
        config:         config,
        cache:          make(map[string]*ProjectInfo),
    }
}

func (s *projectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*ProjectInfo, error) {
    // Priority order for project discovery:
    // 1. Explicit project name
    // 2. Project from context
    // 3. Search in projects directory
    
    if projectName != "" {
        return s.discoverProjectByName(ctx, projectName)
    }
    
    if context != nil {
        switch context.Type {
        case domain.ContextProject:
            return s.discoverProjectByPath(ctx, context.Path)
        case domain.ContextWorktree:
            return s.discoverProjectByWorktree(ctx, context.ProjectName)
        }
    }
    
    return nil, fmt.Errorf("unable to discover project: no project name or context provided")
}

func (s *projectService) discoverProjectByName(ctx context.Context, projectName string) (*ProjectInfo, error) {
    // Check cache first
    s.cacheMu.RLock()
    if cached, exists := s.cache[projectName]; exists {
        s.cacheMu.RUnlock()
        return cached, nil
    }
    s.cacheMu.RUnlock()
    
    // Look in projects directory
    projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)
    
    projectInfo, err := s.buildProjectInfo(ctx, projectName, projectPath)
    if err != nil {
        return nil, fmt.Errorf("failed to discover project '%s': %w", projectName, err)
    }
    
    // Cache result
    s.cacheMu.Lock()
    s.cache[projectName] = projectInfo
    s.cacheMu.Unlock()
    
    return projectInfo, nil
}

func (s *projectService) buildProjectInfo(ctx context.Context, projectName, projectPath string) (*ProjectInfo, error) {
    // Validate project path exists
    if _, err := os.Stat(projectPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("project directory does not exist: %s", projectPath)
    }
    
    // Validate it's a git repository
    if !s.gitClient.IsRepository(projectPath) {
        return nil, fmt.Errorf("not a git repository: %s", projectPath)
    }
    
    // Get repository information
    repo, err := s.gitClient.OpenRepository(projectPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open repository: %w", err)
    }
    defer repo.Close()
    
    // Get worktrees
    worktrees, err := repo.Worktrees()
    if err != nil {
        return nil, fmt.Errorf("failed to get worktrees: %w", err)
    }
    
    // Get branches to find default branch
    branches, err := repo.Branches()
    if err != nil {
        return nil, fmt.Errorf("failed to get branches: %w", err)
    }
    
    defaultBranch := "main" // Default fallback
    for _, branch := range branches {
        if branch.Name == "main" || branch.Name == "master" {
            defaultBranch = branch.Name
            break
        }
    }
    
    return &ProjectInfo{
        Name:          projectName,
        Path:          projectPath,
        GitRepoPath:   projectPath,
        Worktrees:     s.convertWorktrees(worktrees),
        DefaultBranch: defaultBranch,
        IsValid:       true,
    }, nil
}

func (s *projectService) ValidateProject(ctx context.Context, projectPath string) error {
    if !s.gitClient.IsRepository(projectPath) {
        return fmt.Errorf("not a valid git repository: %s", projectPath)
    }
    
    // Additional validation as needed
    return nil
}

func (s *projectService) ListProjects(ctx context.Context) ([]*ProjectInfo, error) {
    // Scan projects directory for git repositories
    entries, err := os.ReadDir(s.config.ProjectsDirectory)
    if err != nil {
        if os.IsNotExist(err) {
            return []*ProjectInfo{}, nil
        }
        return nil, fmt.Errorf("failed to read projects directory: %w", err)
    }
    
    var projects []*ProjectInfo
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        
        projectName := entry.Name()
        projectPath := filepath.Join(s.config.ProjectsDirectory, projectName)
        
        if s.gitClient.IsRepository(projectPath) {
            projectInfo, err := s.buildProjectInfo(ctx, projectName, projectPath)
            if err == nil {
                projects = append(projects, projectInfo)
            }
        }
    }
    
    return projects, nil
}
```

### 4. NavigationService Implementation

```go
// internal/services/navigation_service.go
type navigationService struct {
    projectService ProjectService
    contextService *domain.ContextService
    config         *domain.Config
}

func NewNavigationService(
    projectService ProjectService,
    contextService *domain.ContextService,
    config *domain.Config,
) NavigationService {
    return &navigationService{
        projectService: projectService,
        contextService: contextService,
        config:         config,
    }
}

func (s *navigationService) ResolvePath(ctx context.Context, req *ResolvePathRequest) (*PathResolution, error) {
    // Parse target to determine type
    targetInfo := s.parseTarget(req.Target)
    
    switch req.Context.Type {
    case domain.ContextProject:
        return s.resolveFromProjectContext(ctx, req, targetInfo)
    case domain.ContextWorktree:
        return s.resolveFromWorktreeContext(ctx, req, targetInfo)
    case domain.ContextOutsideGit:
        return s.resolveFromOutsideGitContext(ctx, req, targetInfo)
    default:
        return nil, fmt.Errorf("unknown context type: %v", req.Context.Type)
    }
}

func (s *navigationService) resolveFromProjectContext(ctx context.Context, req *ResolvePathRequest, target *TargetInfo) (*PathResolution, error) {
    currentProject := req.Context.ProjectName
    
    switch target.Type {
    case TargetTypeBranch:
        // Navigate to worktree of current project
        worktreePath := filepath.Join(s.config.WorktreesDirectory, currentProject, target.Branch)
        return &PathResolution{
            ResolvedPath: worktreePath,
            Type:         PathTypeWorktree,
            ProjectName:  currentProject,
            BranchName:   target.Branch,
            Explanation:  fmt.Sprintf("Navigate to worktree '%s' of project '%s'", target.Branch, currentProject),
        }, nil
        
    case TargetTypeProject:
        // Navigate to different project
        project, err := s.projectService.DiscoverProject(ctx, target.Project, req.Context)
        if err != nil {
            return nil, fmt.Errorf("project not found: %s", target.Project)
        }
        
        return &PathResolution{
            ResolvedPath: project.Path,
            Type:         PathTypeProject,
            ProjectName:  target.Project,
            Explanation:  fmt.Sprintf("Navigate to project '%s'", target.Project),
        }, nil
        
    case TargetTypeProjectBranch:
        // Navigate to cross-project worktree
        worktreePath := filepath.Join(s.config.WorktreesDirectory, target.Project, target.Branch)
        return &PathResolution{
            ResolvedPath: worktreePath,
            Type:         PathTypeWorktree,
            ProjectName:  target.Project,
            BranchName:   target.Branch,
            Explanation:  fmt.Sprintf("Navigate to worktree '%s' of project '%s'", target.Branch, target.Project),
        }, nil
    }
    
    return nil, fmt.Errorf("invalid target for project context: %s", req.Target)
}

func (s *navigationService) resolveFromWorktreeContext(ctx context.Context, req *ResolvePathRequest, target *TargetInfo) (*PathResolution, error) {
    currentProject := req.Context.ProjectName
    currentBranch := req.Context.BranchName
    
    switch target.Type {
    case TargetTypeBranch:
        if target.Branch == "main" {
            // Special case: navigate to main project
            project, err := s.projectService.DiscoverProject(ctx, currentProject, req.Context)
            if err != nil {
                return nil, fmt.Errorf("project not found: %s", currentProject)
            }
            
            return &PathResolution{
                ResolvedPath: project.Path,
                Type:         PathTypeProject,
                ProjectName:  currentProject,
                Explanation:  fmt.Sprintf("Navigate to main project '%s'", currentProject),
            }, nil
        } else {
            // Navigate to different worktree of same project
            worktreePath := filepath.Join(s.config.WorktreesDirectory, currentProject, target.Branch)
            return &PathResolution{
                ResolvedPath: worktreePath,
                Type:         PathTypeWorktree,
                ProjectName:  currentProject,
                BranchName:   target.Branch,
                Explanation:  fmt.Sprintf("Navigate to worktree '%s' of project '%s'", target.Branch, currentProject),
            }, nil
        }
        
    case TargetTypeProject:
        // Navigate to different project
        project, err := s.projectService.DiscoverProject(ctx, target.Project, req.Context)
        if err != nil {
            return nil, fmt.Errorf("project not found: %s", target.Project)
        }
        
        return &PathResolution{
            ResolvedPath: project.Path,
            Type:         PathTypeProject,
            ProjectName:  target.Project,
            Explanation:  fmt.Sprintf("Navigate to project '%s'", target.Project),
        }, nil
        
    case TargetTypeProjectBranch:
        // Navigate to cross-project worktree
        worktreePath := filepath.Join(s.config.WorktreesDirectory, target.Project, target.Branch)
        return &PathResolution{
            ResolvedPath: worktreePath,
            Type:         PathTypeWorktree,
            ProjectName:  target.Project,
            BranchName:   target.Branch,
            Explanation:  fmt.Sprintf("Navigate to worktree '%s' of project '%s'", target.Branch, target.Project),
        }, nil
    }
    
    return nil, fmt.Errorf("invalid target for worktree context: %s", req.Target)
}

func (s *navigationService) resolveFromOutsideGitContext(ctx context.Context, req *ResolvePathRequest, target *TargetInfo) (*PathResolution, error) {
    switch target.Type {
    case TargetTypeProject:
        // Navigate to project main directory
        project, err := s.projectService.DiscoverProject(ctx, target.Project, req.Context)
        if err != nil {
            return nil, fmt.Errorf("project not found: %s", target.Project)
        }
        
        return &PathResolution{
            ResolvedPath: project.Path,
            Type:         PathTypeProject,
            ProjectName:  target.Project,
            Explanation:  fmt.Sprintf("Navigate to project '%s'", target.Project),
        }, nil
        
    case TargetTypeProjectBranch:
        // Navigate to cross-project worktree
        worktreePath := filepath.Join(s.config.WorktreesDirectory, target.Project, target.Branch)
        return &PathResolution{
            ResolvedPath: worktreePath,
            Type:         PathTypeWorktree,
            ProjectName:  target.Project,
            BranchName:   target.Branch,
            Explanation:  fmt.Sprintf("Navigate to worktree '%s' of project '%s'", target.Branch, target.Project),
        }, nil
        
    case TargetTypeBranch:
        // Invalid - need project context
        return nil, fmt.Errorf("branch name '%s' requires project context when outside git", target.Branch)
    }
    
    return nil, fmt.Errorf("invalid target for outside git context: %s", req.Target)
}

type TargetType int

const (
    TargetTypeBranch TargetType = iota
    TargetTypeProject
    TargetTypeProjectBranch
)

type TargetInfo struct {
    Type     TargetType
    Project  string
    Branch   string
}

func (s *navigationService) parseTarget(target string) *TargetInfo {
    parts := strings.Split(target, "/")
    
    switch len(parts) {
    case 1:
        // Could be branch or project
        if strings.Contains(target, "/") {
            return &TargetInfo{
                Type:    TargetTypeProjectBranch,
                Project: parts[0],
                Branch:  parts[1],
            }
        } else {
            // Ambiguous - will be resolved based on context
            return &TargetInfo{
                Type:   TargetTypeBranch,
                Branch: target,
            }
        }
    case 2:
        // project/branch format
        return &TargetInfo{
            Type:    TargetTypeProjectBranch,
            Project: parts[0],
            Branch:  parts[1],
        }
    default:
        return &TargetInfo{
            Type: TargetTypeBranch,
            Branch: target,
        }
    }
}
```

## Context-Aware Service Behavior

### Context Integration Pattern

```go
// internal/services/context_aware_service.go
type ContextAwareService struct {
    contextService *domain.ContextService
    worktreeService WorktreeService
    projectService  ProjectService
    navigationService NavigationService
}

func (s *ContextAwareService) ExecuteWithContext(ctx context.Context, operation func(*domain.Context) error) error {
    // Detect current context
    currentContext, err := s.contextService.GetCurrentContext()
    if err != nil {
        return fmt.Errorf("failed to detect context: %w", err)
    }
    
    // Execute operation with context
    return operation(currentContext)
}

func (s *ContextAwareService) AdaptRequestBasedOnContext(req interface{}, context *domain.Context) interface{} {
    // Adapt request based on context type
    switch context.Type {
    case domain.ContextProject:
        return s.adaptForProjectContext(req, context)
    case domain.ContextWorktree:
        return s.adaptForWorktreeContext(req, context)
    case domain.ContextOutsideGit:
        return s.adaptForOutsideGitContext(req, context)
    default:
        return req
    }
}
```

## Error Handling and Validation

### Service-Level Error Types

```go
// internal/domain/service_errors.go
type ServiceError struct {
    Service   string
    Operation string
    Cause     error
    Context   map[string]interface{}
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("%s service error in %s: %v", e.Service, e.Operation, e.Cause)
}

func (e *ServiceError) Unwrap() error {
    return e.Cause
}

// Specific service errors
type WorktreeExistsError struct {
    Path string
}

func (e *WorktreeExistsError) Error() string {
    return fmt.Sprintf("worktree already exists at %s", e.Path)
}

type ProjectNotFoundError struct {
    Name string
}

func (e *ProjectNotFoundError) Error() string {
    return fmt.Sprintf("project not found: %s", e.Name)
}

type UnsafeOperationError struct {
    Operation string
    Reason    string
}

func (e *UnsafeOperationError) Error() string {
    return fmt.Sprintf("unsafe operation '%s': %s", e.Operation, e.Reason)
}
```

## Testing Strategy

### Unit Tests (Testify)

```go
// internal/services/worktree_service_test.go
func TestWorktreeService_CreateWorktree_Success(t *testing.T) {
    mockGitClient := &MockHybridGitClient{}
    mockProjectService := &MockProjectService{}
    mockContextService := &MockContextService{}
    config := &domain.Config{
        WorktreesDirectory: "/test/worktrees",
    }
    
    service := NewWorktreeService(mockGitClient, mockProjectService, mockContextService, config)
    
    req := &CreateWorktreeRequest{
        ProjectName:  "test-project",
        BranchName:   "feature-branch",
        SourceBranch: "main",
        Context: &domain.Context{
            Type:       domain.ContextProject,
            ProjectName: "test-project",
        },
    }
    
    project := &ProjectInfo{
        Name:        "test-project",
        GitRepoPath: "/test/projects/test-project",
    }
    
    mockProjectService.On("DiscoverProject", mock.Anything, "test-project", req.Context).Return(project, nil)
    mockGitClient.On("CreateWorktree", mock.Anything, project.GitRepoPath, req.BranchName, req.SourceBranch, mock.AnythingOfType("string")).Return(nil)
    
    result, err := service.CreateWorktree(context.Background(), req)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockProjectService.AssertExpectations(t)
    mockGitClient.AssertExpectations(t)
}

func TestWorktreeService_CreateWorktree_ValidationError(t *testing.T) {
    mockGitClient := &MockHybridGitClient{}
    mockProjectService := &MockProjectService{}
    mockContextService := &MockContextService{}
    config := &domain.Config{
        WorktreesDirectory: "/test/worktrees",
    }
    
    service := NewWorktreeService(mockGitClient, mockProjectService, mockContextService, config)
    
    req := &CreateWorktreeRequest{
        ProjectName:  "test-project",
        BranchName:   "", // Invalid: empty branch name
        SourceBranch: "main",
        Context: &domain.Context{
            Type:       domain.ContextProject,
            ProjectName: "test-project",
        },
    }
    
    _, err := service.CreateWorktree(context.Background(), req)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
}
```

### Integration Tests (Ginkgo/Gomega)

```go
// internal/services/integration_test.go
var _ = Describe("Services Integration", func() {
    var (
        tempDir          string
        worktreeService  WorktreeService
        projectService   ProjectService
        navigationService NavigationService
        gitClient        infrastructure.HybridGitClient
        contextService   *domain.ContextService
    )
    
    BeforeEach(func() {
        var err error
        tempDir, err = os.MkdirTemp("", "twiggit-services-test")
        Expect(err).NotTo(HaveOccurred())
        
        // Setup real infrastructure components
        config := &domain.Config{
            ProjectsDirectory:  filepath.Join(tempDir, "Projects"),
            WorktreesDirectory: filepath.Join(tempDir, "Worktrees"),
        }
        
        // Create directories
        err = os.MkdirAll(config.ProjectsDirectory, 0755)
        Expect(err).NotTo(HaveOccurred())
        err = os.MkdirAll(config.WorktreesDirectory, 0755)
        Expect(err).NotTo(HaveOccurred())
        
        // Initialize test repository
        repoPath := filepath.Join(config.ProjectsDirectory, "test-repo")
        err = exec.Command("git", "init", repoPath).Run()
        Expect(err).NotTo(HaveOccurred())
        
        // Setup services
        gitClient = setupHybridGitClient()
        contextService = domain.NewContextService(setupContextDetector(config), config)
        projectService = NewProjectService(gitClient, contextService, config)
        worktreeService = NewWorktreeService(gitClient, projectService, contextService, config)
        navigationService = NewNavigationService(projectService, contextService, config)
    })
    
    AfterEach(func() {
        os.RemoveAll(tempDir)
    })
    
    Context("when creating worktrees", func() {
        It("should create worktree with proper validation", func() {
            req := &CreateWorktreeRequest{
                ProjectName:  "test-repo",
                BranchName:   "feature-branch",
                SourceBranch: "main",
                Context: &domain.Context{
                    Type:       domain.ContextOutsideGit,
                    ProjectName: "",
                },
            }
            
            worktreeInfo, err := worktreeService.CreateWorktree(context.Background(), req)
            Expect(err).NotTo(HaveOccurred())
            Expect(worktreeInfo.Branch).To(Equal("feature-branch"))
            Expect(worktreeInfo.Path).To(ContainSubstring("feature-branch"))
        })
        
        It("should fail when worktree already exists", func() {
            // Create first worktree
            req := &CreateWorktreeRequest{
                ProjectName:  "test-repo",
                BranchName:   "feature-branch",
                SourceBranch: "main",
                Context: &domain.Context{
                    Type:       domain.ContextOutsideGit,
                    ProjectName: "",
                },
            }
            
            _, err := worktreeService.CreateWorktree(context.Background(), req)
            Expect(err).NotTo(HaveOccurred())
            
            // Try to create again
            _, err = worktreeService.CreateWorktree(context.Background(), req)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(ContainSubstring("already exists"))
        })
    })
    
    Context("when navigating paths", func() {
        It("should resolve paths correctly from project context", func() {
            req := &ResolvePathRequest{
                Target: "feature-branch",
                Context: &domain.Context{
                    Type:       domain.ContextProject,
                    ProjectName: "test-repo",
                    Path:       filepath.Join(tempDir, "Projects", "test-repo"),
                },
            }
            
            resolution, err := navigationService.ResolvePath(context.Background(), req)
            Expect(err).NotTo(HaveOccurred())
            Expect(resolution.Type).To(Equal(PathTypeWorktree))
            Expect(resolution.ProjectName).To(Equal("test-repo"))
            Expect(resolution.BranchName).To(Equal("feature-branch"))
        })
    })
})
```

## Implementation Steps

### Phase 1: Core Service Interfaces and Structure
1. **Define service interfaces** in `internal/services/interfaces.go`
2. **Implement service error types** in `internal/domain/service_errors.go`
3. **Create service base structures** with dependency injection
4. **Write unit tests** for interface contracts

### Phase 2: WorktreeService Implementation
1. **Implement core worktree operations** (create, delete, list)
2. **Implement WorktreeValidator** with safety checks
3. **Add context-aware project resolution**
4. **Write comprehensive unit tests** with mocks

### Phase 3: ProjectService Implementation
1. **Implement project discovery** logic
2. **Add project validation** and caching
3. **Implement project listing** functionality
4. **Write unit and integration tests**

### Phase 4: NavigationService Implementation
1. **Implement path resolution** logic for all contexts
2. **Add target parsing** and validation
3. **Implement navigation suggestions** for completion
4. **Write tests for all navigation scenarios**

### Phase 5: Context Integration and Orchestration
1. **Implement context-aware service behavior**
2. **Add service coordination** patterns
3. **Implement proper error handling** and recovery
4. **Write integration tests** for service coordination

### Phase 6: Performance and Optimization
1. **Add caching layers** where appropriate
2. **Implement concurrent operations** where safe
3. **Add performance monitoring** and metrics
4. **Optimize for large project sets**

## Configuration Integration

### Service Configuration

```toml
# config.toml additions for services
[services]
cache_enabled = true
cache_ttl = "5m"
concurrent_operations = true
max_concurrent = 4

[services.validation]
strict_branch_names = true
require_clean_worktree = true
allow_force_delete = false

[services.navigation]
enable_suggestions = true
max_suggestions = 10
fuzzy_matching = false
```

## Success Criteria

1. **Functional Requirements**
   - All service operations work correctly with context awareness
   - Proper validation and safety checks implemented
   - Error handling provides actionable feedback

2. **Quality Requirements**
   - >80% test coverage for all service logic
   - All integration tests pass with real git repositories
   - Code follows established patterns and conventions

3. **Performance Requirements**
   - Service operations complete within specified time limits
   - Memory usage remains within bounds
   - Context detection and resolution is efficient

4. **Integration Requirements**
   - Services integrate seamlessly with existing layers
   - CLI commands can use services effectively
   - Context-aware behavior works as specified

## Dependencies

- **Foundation Layer**: Domain entities, interfaces, dependency injection
- **Configuration Layer**: Koanf-based configuration management
- **Context Detection**: ContextDetector and ContextService
- **Hybrid Git**: HybridGitClient for git operations
- **Testing**: Testify for unit tests, Ginkgo/Gomega for integration tests

## Timeline

- **Phase 1-2**: Core interfaces and WorktreeService (3-4 days)
- **Phase 3**: ProjectService implementation (2-3 days)
- **Phase 4**: NavigationService implementation (2-3 days)
- **Phase 5**: Context integration and orchestration (2-3 days)
- **Phase 6**: Performance and optimization (1-2 days)

**Total Estimated Time**: 10-15 days

This implementation plan provides a comprehensive approach to building the core services layer while maintaining consistency with the established architecture and requirements from the design and implementation documents.